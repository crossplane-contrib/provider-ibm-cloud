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

package resourcekey

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

func params(m ...func(*v1alpha1.ResourceKeyParameters)) *v1alpha1.ResourceKeyParameters {
	p := &v1alpha1.ResourceKeyParameters{
		Name:   "my-instance-key-1",
		Source: reference.ToPtrValue("25eba2a9-beef-450b-82cf-f5ad5e36c6dd"),
		Parameters: &v1alpha1.ResourceKeyPostParameters{
			ServiceidCRN: "crn:v1:bluemix:public:iam-identity::a/9fceaa56d1ab84893af6b9eec5ab81bb::serviceid:ServiceId-fe4c29b5-db13-410a-bacc-b5779a03d393",
		},
		Role: reference.ToPtrValue("crn:v1:bluemix:public:iam::::serviceRole:Writer"),
	}

	for _, f := range m {
		f(p)
	}
	return p
}

func observation(m ...func(*v1alpha1.ResourceKeyObservation)) *v1alpha1.ResourceKeyObservation {
	o := &v1alpha1.ResourceKeyObservation{
		ID:                  "crn:v1:bluemix:public:cloud-object-storage:global:a/4329073d16d2f3663f74bfa955259139:8d7af921-b136-4078-9666-081bd8470d94:resource-key:23693f48-aaa2-4079-b0c7-334846eff8d0",
		GUID:                "23693f48-aaa2-4079-b0c7-334846eff8d0",
		CRN:                 "crn:v1:bluemix:public:cloud-object-storage:global:a/4329073d16d2f3663f74bfa955259139:8d7af921-b136-4078-9666-081bd8470d94:resource-key:23693f48-aaa2-4079-b0c7-334846eff8d0",
		URL:                 "/v2/resource_keys/23693f48-aaa2-4079-b0c7-334846eff8d0",
		AccountID:           "4329073d16d2f3663f74bfa955259139",
		ResourceGroupID:     "0be5ad401ae913d8ff665d92680664ed",
		SourceCRN:           "crn:v1:bluemix:public:cloud-object-storage:global:a/4329073d16d2f3663f74bfa955259139:8d7af921-b136-4078-9666-081bd8470d94::",
		State:               "active",
		IamCompatible:       true,
		ResourceInstanceURL: "/v2/resource_instances/8d7af921-b136-4078-9666-081bd8470d94",
		CreatedAt:           ibmc.ParseMetaV1Time("2020-10-31T02:33:06Z"),
		UpdatedAt:           ibmc.ParseMetaV1Time("2020-10-31T02:33:06Z"),
		DeletedAt:           ibmc.ParseMetaV1Time("2020-10-31T02:33:06Z"),
		CreatedBy:           "IBMid-5500093BHN",
		UpdatedBy:           "IBMid-5500093BHN",
		DeletedBy:           "testString",
	}

	for _, f := range m {
		f(o)
	}
	return o
}

func instance(m ...func(*rcv2.ResourceKey)) *rcv2.ResourceKey {
	i := &rcv2.ResourceKey{
		ID:                  reference.ToPtrValue("crn:v1:bluemix:public:cloud-object-storage:global:a/4329073d16d2f3663f74bfa955259139:8d7af921-b136-4078-9666-081bd8470d94:resource-key:23693f48-aaa2-4079-b0c7-334846eff8d0"),
		GUID:                reference.ToPtrValue("23693f48-aaa2-4079-b0c7-334846eff8d0"),
		CRN:                 reference.ToPtrValue("crn:v1:bluemix:public:cloud-object-storage:global:a/4329073d16d2f3663f74bfa955259139:8d7af921-b136-4078-9666-081bd8470d94:resource-key:23693f48-aaa2-4079-b0c7-334846eff8d0"),
		URL:                 reference.ToPtrValue("/v2/resource_keys/23693f48-aaa2-4079-b0c7-334846eff8d0"),
		Name:                reference.ToPtrValue("my-instance-key-1"),
		AccountID:           reference.ToPtrValue("4329073d16d2f3663f74bfa955259139"),
		ResourceGroupID:     reference.ToPtrValue("0be5ad401ae913d8ff665d92680664ed"),
		SourceCRN:           reference.ToPtrValue("crn:v1:bluemix:public:cloud-object-storage:global:a/4329073d16d2f3663f74bfa955259139:8d7af921-b136-4078-9666-081bd8470d94::"),
		Role:                reference.ToPtrValue("crn:v1:bluemix:public:iam::::serviceRole:Writer"),
		State:               reference.ToPtrValue("active"),
		IamCompatible:       ibmc.BoolPtr(true),
		ResourceInstanceURL: reference.ToPtrValue("/v2/resource_instances/8d7af921-b136-4078-9666-081bd8470d94"),
		CreatedAt:           ibmc.ParseDateTimePtr("2020-10-31T02:33:06Z"),
		UpdatedAt:           ibmc.ParseDateTimePtr("2020-10-31T02:33:06Z"),
		DeletedAt:           ibmc.ParseDateTimePtr("2020-10-31T02:33:06Z"),
		CreatedBy:           reference.ToPtrValue("IBMid-5500093BHN"),
		UpdatedBy:           reference.ToPtrValue("IBMid-5500093BHN"),
		DeletedBy:           reference.ToPtrValue("testString"),
	}

	for _, f := range m {
		f(i)
	}
	return i
}

func instanceCreateOpts(m ...func(*rcv2.CreateResourceKeyOptions)) *rcv2.CreateResourceKeyOptions {
	i := &rcv2.CreateResourceKeyOptions{
		Name:   reference.ToPtrValue("my-instance-key-1"),
		Source: reference.ToPtrValue("25eba2a9-beef-450b-82cf-f5ad5e36c6dd"),
		Parameters: &rcv2.ResourceKeyPostParameters{
			ServiceidCRN: reference.ToPtrValue("crn:v1:bluemix:public:iam-identity::a/9fceaa56d1ab84893af6b9eec5ab81bb::serviceid:ServiceId-fe4c29b5-db13-410a-bacc-b5779a03d393"),
		},
		Role: reference.ToPtrValue("crn:v1:bluemix:public:iam::::serviceRole:Writer"),
	}
	for _, f := range m {
		f(i)
	}
	return i
}

func instanceUpdateOpts(m ...func(*rcv2.UpdateResourceKeyOptions)) *rcv2.UpdateResourceKeyOptions {
	i := &rcv2.UpdateResourceKeyOptions{
		ID:   reference.ToPtrValue("crn:v1:bluemix:public:cloud-object-storage:global:a/4329073d16d2f3663f74bfa955259139:8d7af921-b136-4078-9666-081bd8470d94:resource-key:23693f48-aaa2-4079-b0c7-334846eff8d0"),
		Name: reference.ToPtrValue("my-instance-key-1"),
	}
	for _, f := range m {
		f(i)
	}
	return i
}

// Test GenerateCreateResourceKeyOptions method
func TestGenerateCreateResourceKeyOptions(t *testing.T) {
	type args struct {
		params v1alpha1.ResourceKeyParameters
	}
	type want struct {
		instance *rcv2.CreateResourceKeyOptions
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
				params: *params(func(p *v1alpha1.ResourceKeyParameters) {
					p.Parameters = nil
					p.Role = nil
				})},
			want: want{
				instance: instanceCreateOpts(func(i *rcv2.CreateResourceKeyOptions) {
					i.Parameters = nil
					i.Role = nil
				})},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			mux := http.NewServeMux()
			server := httptest.NewServer(mux)
			defer server.Close()

			opts := ibmc.ClientOptions{URL: server.URL, Authenticator: &core.BearerTokenAuthenticator{
				BearerToken: ibmc.FakeBearerToken,
			}}
			mClient, _ := ibmc.NewClient(opts)

			r := &rcv2.CreateResourceKeyOptions{}
			GenerateCreateResourceKeyOptions(mClient, tc.args.params, r)
			if diff := cmp.Diff(tc.want.instance, r,
				// temporary hack
				cmpopts.IgnoreUnexported((rcv2.ResourceKeyPostParameters{})),
				cmpopts.IgnoreFields(rcv2.CreateResourceKeyOptions{})); diff != "" {
				t.Errorf("GenerateCreateResourceKeyOptions(...): -want, +got:\n%s", diff)
			}
		})
	}
}

// Test GenerateUpdateResourceKeyOptions method
func TestGenerateUpdateResourceKeyOptions(t *testing.T) {
	type args struct {
		params v1alpha1.ResourceKeyParameters
	}
	type want struct {
		instance *rcv2.UpdateResourceKeyOptions
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
				params: *params(func(p *v1alpha1.ResourceKeyParameters) {
				})},
			want: want{
				instance: instanceUpdateOpts(func(i *rcv2.UpdateResourceKeyOptions) {
				})},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			mux := http.NewServeMux()
			server := httptest.NewServer(mux)
			defer server.Close()

			opts := ibmc.ClientOptions{URL: server.URL, Authenticator: &core.BearerTokenAuthenticator{
				BearerToken: ibmc.FakeBearerToken,
			}}
			mClient, _ := ibmc.NewClient(opts)

			r := &rcv2.UpdateResourceKeyOptions{}
			GenerateUpdateResourceKeyOptions(mClient, observation().ID, tc.args.params, r)
			if diff := cmp.Diff(tc.want.instance, r,
				cmpopts.IgnoreFields(rcv2.UpdateResourceKeyOptions{})); diff != "" {
				t.Errorf("GenerateUpdateResourceKeyOptions(...): -want, +got:\n%s", diff)
			}
		})
	}
}

// Test LateInitializeSpecs method
func TestResourceKeyLateInitializeSpecs(t *testing.T) {
	type args struct {
		instance *rcv2.ResourceKey
		params   *v1alpha1.ResourceKeyParameters
	}
	type want struct {
		params *v1alpha1.ResourceKeyParameters
	}
	cases := map[string]struct {
		args args
		want want
	}{
		"SomeFields": {
			args: args{
				params: params(func(p *v1alpha1.ResourceKeyParameters) {
					p.Role = nil
				}),
				instance: instance(func(i *rcv2.ResourceKey) {
				}),
			},
			want: want{
				params: params(func(p *v1alpha1.ResourceKeyParameters) {
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
			server := httptest.NewServer(mux)
			defer server.Close()

			opts := ibmc.ClientOptions{URL: server.URL, Authenticator: &core.BearerTokenAuthenticator{
				BearerToken: ibmc.FakeBearerToken,
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
func TestResourceKeyGenerateObservation(t *testing.T) {
	type args struct {
		instance *rcv2.ResourceKey
	}
	type want struct {
		obs v1alpha1.ResourceKeyObservation
	}
	cases := map[string]struct {
		args args
		want want
	}{
		"FullConversion": {
			args: args{
				instance: instance(func(p *rcv2.ResourceKey) {
				}),
			},
			want: want{*observation()},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			mux := http.NewServeMux()
			server := httptest.NewServer(mux)
			defer server.Close()

			opts := ibmc.ClientOptions{URL: server.URL, Authenticator: &core.BearerTokenAuthenticator{
				BearerToken: ibmc.FakeBearerToken,
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
func TestResourceKeyIsUpToDate(t *testing.T) {
	type args struct {
		params   *v1alpha1.ResourceKeyParameters
		instance *rcv2.ResourceKey
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
				instance: instance(func(i *rcv2.ResourceKey) {
					i.Name = reference.ToPtrValue("different-name")
				}),
			},
			want: want{upToDate: false, isErr: false},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			mux := http.NewServeMux()
			server := httptest.NewServer(mux)
			defer server.Close()

			opts := ibmc.ClientOptions{URL: server.URL, Authenticator: &core.BearerTokenAuthenticator{
				BearerToken: ibmc.FakeBearerToken,
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
