/*
Copyright 2021 The Crossplane Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package resourceinstance

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/reference"

	"github.com/IBM/go-sdk-core/core"
	rcv2 "github.com/IBM/platform-services-go-sdk/resourcecontrollerv2"

	"github.com/crossplane-contrib/provider-ibm-cloud/apis/resourcecontrollerv2/v1alpha1"
	ibmc "github.com/crossplane-contrib/provider-ibm-cloud/pkg/clients"
)

const (
	bearerTok = "mock-token"
)

func params(m ...func(*v1alpha1.ResourceInstanceParameters)) *v1alpha1.ResourceInstanceParameters {
	p := &v1alpha1.ResourceInstanceParameters{
		Name:              "my-instance",
		Target:            "global",
		ResourceGroupName: reference.ToPtrValue("default"),
		ServiceName:       "cloud-object-storage",
		ResourcePlanName:  "standard",
		Tags:              []string{"testString"},
		AllowCleanup:      ibmc.BoolPtr(false),
		Parameters:        ibmc.MapToRawExtension(make(map[string]interface{})),
	}

	for _, f := range m {
		f(p)
	}
	return p
}

func observation(m ...func(*v1alpha1.ResourceInstanceObservation)) *v1alpha1.ResourceInstanceObservation {
	o := &v1alpha1.ResourceInstanceObservation{
		ID:                  "crn:v1:bluemix:public:cloud-object-storage:global:a/4329073d16d2f3663f74bfa955259139:8d7af921-b136-4078-9666-081bd8470d94::",
		GUID:                "8d7af921-b136-4078-9666-081bd8470d94",
		CRN:                 "crn:v1:bluemix:public:cloud-object-storage:global:a/4329073d16d2f3663f74bfa955259139:8d7af921-b136-4078-9666-081bd8470d94::",
		URL:                 "/v2/resource_instances/8d7af921-b136-4078-9666-081bd8470d94",
		AccountID:           "4329073d16d2f3663f74bfa955259139",
		ResourceGroupID:     "0be5ad401ae913d8ff665d92680664ed",
		ResourceGroupCRN:    "crn:v1:bluemix:public:resource-controller::a/4329073d16d2f3663f74bfa955259139::resource-group:0be5ad401ae913d8ff665d92680664ed",
		ResourceID:          "dff97f5c-bc5e-4455-b470-411c3edbe49c",
		ResourcePlanID:      "2fdf0c08-2d32-4f46-84b5-32e0c92fffd8",
		TargetCRN:           "crn:v1:bluemix:public:resource-catalog::a/9e16d1fed8aa7e1bd73e7a9d23434a5a::deployment:2fdf0c08-2d32-4f46-84b5-32e0c92fffd8%3Aglobal",
		State:               "active",
		Type:                "service_instance",
		SubType:             "testString",
		Locked:              true,
		LastOperation:       ibmc.MapToRawExtension(make(map[string]interface{})),
		DashboardURL:        "/objectstorage/crn%3Av1%3Abluemix%3Apublic%3Acloud-object-storage%3Aglobal%3Aa%2F4329073d16d2f3663f74bfa955259139%3A8d7af921-b136-4078-9666-081bd8470d94%3A%3A",
		PlanHistory:         []v1alpha1.PlanHistoryItem{},
		Extensions:          ibmc.MapToRawExtension(make(map[string]interface{})),
		ResourceAliasesURL:  "/v2/resource_instances/8d7af921-b136-4078-9666-081bd8470d94/resource_aliases",
		ResourceBindingsURL: "/v2/resource_instances/8d7af921-b136-4078-9666-081bd8470d94/resource_bindings",
		ResourceKeysURL:     "/v2/resource_instances/8d7af921-b136-4078-9666-081bd8470d94/resource_keys",
		CreatedAt:           ibmc.ParseMetaV1Time("2020-10-31T02:33:06Z"),
		CreatedBy:           "IBMid-5500093BHN",
		UpdatedAt:           ibmc.ParseMetaV1Time("2020-10-31T02:33:06Z"),
		UpdatedBy:           "IBMid-5500093BHN",
		DeletedAt:           ibmc.ParseMetaV1Time("2020-10-31T02:33:06Z"),
		DeletedBy:           "testString",
		ScheduledReclaimAt:  ibmc.ParseMetaV1Time("2020-10-31T02:33:06Z"),
		ScheduledReclaimBy:  "testString",
		RestoredAt:          ibmc.ParseMetaV1Time("2020-10-31T02:33:06Z"),
		RestoredBy:          "testString",
	}

	for _, f := range m {
		f(o)
	}
	return o
}

func instance(m ...func(*rcv2.ResourceInstance)) *rcv2.ResourceInstance {
	i := &rcv2.ResourceInstance{
		ID:                  reference.ToPtrValue("crn:v1:bluemix:public:cloud-object-storage:global:a/4329073d16d2f3663f74bfa955259139:8d7af921-b136-4078-9666-081bd8470d94::"),
		GUID:                reference.ToPtrValue("8d7af921-b136-4078-9666-081bd8470d94"),
		CRN:                 reference.ToPtrValue("crn:v1:bluemix:public:cloud-object-storage:global:a/4329073d16d2f3663f74bfa955259139:8d7af921-b136-4078-9666-081bd8470d94::"),
		URL:                 reference.ToPtrValue("/v2/resource_instances/8d7af921-b136-4078-9666-081bd8470d94"),
		Name:                reference.ToPtrValue("my-instance"),
		AccountID:           reference.ToPtrValue("4329073d16d2f3663f74bfa955259139"),
		ResourceGroupID:     reference.ToPtrValue("0be5ad401ae913d8ff665d92680664ed"),
		ResourceGroupCRN:    reference.ToPtrValue("crn:v1:bluemix:public:resource-controller::a/4329073d16d2f3663f74bfa955259139::resource-group:0be5ad401ae913d8ff665d92680664ed"),
		ResourceID:          reference.ToPtrValue("dff97f5c-bc5e-4455-b470-411c3edbe49c"),
		ResourcePlanID:      reference.ToPtrValue("2fdf0c08-2d32-4f46-84b5-32e0c92fffd8"),
		TargetCRN:           reference.ToPtrValue("crn:v1:bluemix:public:resource-catalog::a/9e16d1fed8aa7e1bd73e7a9d23434a5a::deployment:2fdf0c08-2d32-4f46-84b5-32e0c92fffd8%3Aglobal"),
		Parameters:          make(map[string]interface{}),
		State:               reference.ToPtrValue("active"),
		Type:                reference.ToPtrValue("service_instance"),
		SubType:             reference.ToPtrValue("testString"),
		AllowCleanup:        ibmc.BoolPtr(false),
		Locked:              ibmc.BoolPtr(true),
		LastOperation:       make(map[string]interface{}),
		DashboardURL:        reference.ToPtrValue("/objectstorage/crn%3Av1%3Abluemix%3Apublic%3Acloud-object-storage%3Aglobal%3Aa%2F4329073d16d2f3663f74bfa955259139%3A8d7af921-b136-4078-9666-081bd8470d94%3A%3A"),
		PlanHistory:         []rcv2.PlanHistoryItem{},
		Extensions:          make(map[string]interface{}),
		ResourceAliasesURL:  reference.ToPtrValue("/v2/resource_instances/8d7af921-b136-4078-9666-081bd8470d94/resource_aliases"),
		ResourceBindingsURL: reference.ToPtrValue("/v2/resource_instances/8d7af921-b136-4078-9666-081bd8470d94/resource_bindings"),
		ResourceKeysURL:     reference.ToPtrValue("/v2/resource_instances/8d7af921-b136-4078-9666-081bd8470d94/resource_keys"),
		CreatedAt:           ibmc.ParseDateTimePtr("2020-10-31T02:33:06Z"),
		CreatedBy:           reference.ToPtrValue("IBMid-5500093BHN"),
		UpdatedAt:           ibmc.ParseDateTimePtr("2020-10-31T02:33:06Z"),
		UpdatedBy:           reference.ToPtrValue("IBMid-5500093BHN"),
		DeletedAt:           ibmc.ParseDateTimePtr("2020-10-31T02:33:06Z"),
		DeletedBy:           reference.ToPtrValue("testString"),
		ScheduledReclaimAt:  ibmc.ParseDateTimePtr("2020-10-31T02:33:06Z"),
		ScheduledReclaimBy:  reference.ToPtrValue("testString"),
		RestoredAt:          ibmc.ParseDateTimePtr("2020-10-31T02:33:06Z"),
		RestoredBy:          reference.ToPtrValue("testString"),
	}

	for _, f := range m {
		f(i)
	}
	return i
}

func instanceCreateOpts(m ...func(*rcv2.CreateResourceInstanceOptions)) *rcv2.CreateResourceInstanceOptions {
	i := &rcv2.CreateResourceInstanceOptions{
		Name:           reference.ToPtrValue("my-instance"),
		Target:         reference.ToPtrValue("global"),
		ResourceGroup:  reference.ToPtrValue("0be5ad401ae913d8ff665d92680664ed"),
		ResourcePlanID: reference.ToPtrValue("2fdf0c08-2d32-4f46-84b5-32e0c92fffd8"),
		Tags:           []string{"testString"},
		AllowCleanup:   ibmc.BoolPtr(false),
		Parameters:     make(map[string]interface{}),
	}
	for _, f := range m {
		f(i)
	}
	return i
}

func instanceUpdateOpts(m ...func(*rcv2.UpdateResourceInstanceOptions)) *rcv2.UpdateResourceInstanceOptions {
	i := &rcv2.UpdateResourceInstanceOptions{
		ID:             reference.ToPtrValue("crn:v1:bluemix:public:cloud-object-storage:global:a/4329073d16d2f3663f74bfa955259139:8d7af921-b136-4078-9666-081bd8470d94::"),
		Name:           reference.ToPtrValue("my-instance"),
		Parameters:     make(map[string]interface{}),
		ResourcePlanID: reference.ToPtrValue("2fdf0c08-2d32-4f46-84b5-32e0c92fffd8"),
		AllowCleanup:   ibmc.BoolPtr(false),
	}
	for _, f := range m {
		f(i)
	}
	return i
}

// Test GenerateCreateResourceInstanceOptions method
func TestGenerateCreateResourceInstanceOptions(t *testing.T) {
	type args struct {
		params v1alpha1.ResourceInstanceParameters
	}
	type want struct {
		instance *rcv2.CreateResourceInstanceOptions
	}
	cases := map[string]struct {
		args args
		want want
	}{
		"FullConversion": {
			args: args{params: *params()},
			want: want{instance: instanceCreateOpts()},
		},
		"MissingFields": {
			args: args{
				params: *params(func(p *v1alpha1.ResourceInstanceParameters) {
					p.Tags = nil
					p.AllowCleanup = nil
					p.Parameters = nil
				})},
			want: want{
				instance: instanceCreateOpts(func(i *rcv2.CreateResourceInstanceOptions) {
					i.Tags = nil
					i.AllowCleanup = nil
					i.Parameters = nil
				})},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			mux := http.NewServeMux()
			mux.HandleFunc("/resource_groups/", ibmc.RgTestHandler)
			mux.HandleFunc("/", ibmc.SvcatTestHandler)
			mux.HandleFunc("/"+ibmc.ServiceNameMockVal+"/", ibmc.PcatTestHandler)
			server := httptest.NewServer(mux)
			defer server.Close()

			opts := ibmc.ClientOptions{URL: server.URL, Authenticator: &core.BearerTokenAuthenticator{
				BearerToken: bearerTok,
			}}
			mClient, _ := ibmc.NewClient(opts)

			r := &rcv2.CreateResourceInstanceOptions{}
			GenerateCreateResourceInstanceOptions(mClient, tc.args.params, r)
			if diff := cmp.Diff(tc.want.instance, r,
				// temporary hack
				cmpopts.IgnoreUnexported((rcv2.ResourceKeyPostParameters{})),
				cmpopts.IgnoreFields(rcv2.CreateResourceInstanceOptions{})); diff != "" {
				t.Errorf("GenerateCreateResourceInstanceOptions(...): -want, +got:\n%s", diff)
			}
		})
	}
}

// Test GenerateUpdateResourceInstanceOptions method
func TestGenerateUpdateResourceInstanceOptions(t *testing.T) {
	type args struct {
		params v1alpha1.ResourceInstanceParameters
	}
	type want struct {
		instance *rcv2.UpdateResourceInstanceOptions
	}
	cases := map[string]struct {
		args args
		want want
	}{
		"FullConversion": {
			args: args{params: *params()},
			want: want{instance: instanceUpdateOpts()},
		},
		"MissingFields": {
			args: args{
				params: *params(func(p *v1alpha1.ResourceInstanceParameters) {
					p.Parameters = nil
					p.AllowCleanup = nil
				})},
			want: want{
				instance: instanceUpdateOpts(func(i *rcv2.UpdateResourceInstanceOptions) {
					i.Parameters = nil
					i.AllowCleanup = nil
				})},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			mux := http.NewServeMux()
			mux.HandleFunc("/resource_groups/", ibmc.RgTestHandler)
			mux.HandleFunc("/", ibmc.SvcatTestHandler)
			mux.HandleFunc("/"+ibmc.ServiceNameMockVal+"/", ibmc.PcatTestHandler)
			server := httptest.NewServer(mux)
			defer server.Close()

			opts := ibmc.ClientOptions{URL: server.URL, Authenticator: &core.BearerTokenAuthenticator{
				BearerToken: bearerTok,
			}}
			mClient, _ := ibmc.NewClient(opts)

			r := &rcv2.UpdateResourceInstanceOptions{}
			GenerateUpdateResourceInstanceOptions(mClient, observation().ID, tc.args.params, r)
			if diff := cmp.Diff(tc.want.instance, r,
				cmpopts.IgnoreFields(rcv2.UpdateResourceInstanceOptions{})); diff != "" {
				t.Errorf("GenerateUpdateResourceInstanceOptions(...): -want, +got:\n%s", diff)
			}
		})
	}
}

// Test LateInitializeSpecs method
func TestResourceInstanceLateInitializeSpecs(t *testing.T) {
	type args struct {
		instance *rcv2.ResourceInstance
		params   *v1alpha1.ResourceInstanceParameters
	}
	type want struct {
		params *v1alpha1.ResourceInstanceParameters
	}
	cases := map[string]struct {
		args args
		want want
	}{
		"SomeFields": {
			args: args{
				params: params(func(p *v1alpha1.ResourceInstanceParameters) {
					p.AllowCleanup = nil
					p.Parameters = nil
				}),
				instance: instance(func(i *rcv2.ResourceInstance) {
				}),
			},
			want: want{
				params: params(func(p *v1alpha1.ResourceInstanceParameters) {
				})},
		},
		"AllFilledAlready": {
			args: args{
				params:   params(),
				instance: instance(),
			},
			want: want{
				params: params()},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			mux := http.NewServeMux()
			mux.HandleFunc("/resource_groups/", ibmc.RgTestHandler)
			mux.HandleFunc("/", ibmc.SvcatTestHandler)
			mux.HandleFunc("/"+ibmc.ServiceNameMockVal+"/", ibmc.PcatTestHandler)
			mux.HandleFunc("/v3/tags/", ibmc.TagsTestHandler)
			server := httptest.NewServer(mux)
			defer server.Close()

			opts := ibmc.ClientOptions{URL: server.URL, Authenticator: &core.BearerTokenAuthenticator{
				BearerToken: bearerTok,
			}}
			mClient, _ := ibmc.NewClient(opts)

			LateInitializeSpec(mClient, tc.args.params, tc.args.instance)
			if diff := cmp.Diff(tc.want.params, tc.args.params); diff != "" {
				t.Errorf("LateInitializeSpec(...): -want, +got:\n%s", diff)
			}
		})
	}
}

// Test GenerateObservation method
func TestResourceInstanceGenerateObservation(t *testing.T) {
	type args struct {
		instance *rcv2.ResourceInstance
	}
	type want struct {
		obs v1alpha1.ResourceInstanceObservation
	}
	cases := map[string]struct {
		args args
		want want
	}{
		"FullConversion": {
			args: args{
				instance: instance(func(p *rcv2.ResourceInstance) {
				}),
			},
			want: want{*observation()},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			mux := http.NewServeMux()
			mux.HandleFunc("/v3/tags/", ibmc.TagsTestHandler)
			server := httptest.NewServer(mux)
			defer server.Close()

			opts := ibmc.ClientOptions{URL: server.URL, Authenticator: &core.BearerTokenAuthenticator{
				BearerToken: bearerTok,
			}}
			mClient, _ := ibmc.NewClient(opts)

			o, err := GenerateObservation(mClient, tc.args.instance)
			if diff := cmp.Diff(nil, err); diff != "" {
				t.Errorf("GenerateObservation(...): want error != got error:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.obs, o); diff != "" {
				t.Errorf("GenerateObservation(...): -want, +got:\n%s", diff)
			}
		})
	}
}

// Test IsUpToDate method
func TestResourceInstanceIsUpToDate(t *testing.T) {
	type args struct {
		params   *v1alpha1.ResourceInstanceParameters
		instance *rcv2.ResourceInstance
	}
	type want struct {
		upToDate bool
		isErr    bool
	}
	cases := map[string]struct {
		args args
		want want
	}{
		"IsUpToDate": {
			args: args{
				params:   params(),
				instance: instance(),
			},
			want: want{upToDate: true, isErr: false},
		},
		"NeedsUpdate": {
			args: args{
				params: params(),
				instance: instance(func(i *rcv2.ResourceInstance) {
					i.Name = reference.ToPtrValue("different-name")
				}),
			},
			want: want{upToDate: false, isErr: false},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			mux := http.NewServeMux()
			mux.HandleFunc("/resource_groups/", ibmc.RgTestHandler)
			mux.HandleFunc("/", ibmc.SvcatTestHandler)
			mux.HandleFunc("/"+ibmc.ServiceNameMockVal+"/", ibmc.PcatTestHandler)
			mux.HandleFunc("/v3/tags/", ibmc.TagsTestHandler)
			server := httptest.NewServer(mux)
			defer server.Close()

			opts := ibmc.ClientOptions{URL: server.URL, Authenticator: &core.BearerTokenAuthenticator{
				BearerToken: bearerTok,
			}}
			mClient, _ := ibmc.NewClient(opts)

			r, err := IsUpToDate(mClient, tc.args.params, tc.args.instance, logging.NewNopLogger())
			if err != nil && !tc.want.isErr {
				t.Error("IsUpToDate(...) unexpected error")
			}
			if diff := cmp.Diff(tc.want.upToDate, r); diff != "" {
				t.Errorf("IsUpToDate(...): -want, +got:\n%s", diff)
			}
		})
	}
}
