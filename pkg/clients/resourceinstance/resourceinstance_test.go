package resourceinstance

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-openapi/strfmt"
	"github.com/google/go-cmp/cmp"

	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/reference"

	"github.com/IBM/go-sdk-core/core"
	gcat "github.com/IBM/platform-services-go-sdk/globalcatalogv1"
	gtagv1 "github.com/IBM/platform-services-go-sdk/globaltaggingv1"
	rcv2 "github.com/IBM/platform-services-go-sdk/resourcecontrollerv2"
	rmgrv2 "github.com/IBM/platform-services-go-sdk/resourcemanagerv2"

	"github.com/crossplane-contrib/provider-ibm-cloud/apis/resourcecontrollerv2/v1alpha1"
	ibmc "github.com/crossplane-contrib/provider-ibm-cloud/pkg/clients"
)

const (
	bearerTok = "mock-token"
)

var (
	accountID         = "fake-account-id"
	resourcePlanName  = "standard"
	resourceGroupName = "default"
	allowCleanup      = false
	crnRes            = "crn:v1:bluemix:public:cloud-object-storage:global:a/0b5a00334eaf9eb9339d2ab48f20d7f5:78d88b2b-bbbb-aaaa-8888-5c26e8b6a555::"
	guid              = "78d88b2b-bbbb-aaaa-8888-5c26e8b6a555"
	id                = "crn:v1:bluemix:public:cloud-object-storage:global:a/0b5a00334eaf9eb9339d2ab48f20d7f5:78d88b2b-bbbb-aaaa-8888-5c26e8b6a555::"
	locked            = false
	instName          = "cos-wow"
	createdAt, _      = strfmt.ParseDateTime("2020-10-31T02:33:06Z")
	resourceGroupID   = "mock-resource-group-id"
	resourceID        = "aaaaaaaa-bbbb-cccc-b470-411c3edbe49c"
	resourcePlanID    = "744bfc56-d12c-4866-88d5-dac9139e0e5d"
	state             = StateActive
	serviceName       = "cloud-object-storage"
	tags              = []string{"dev"}
	target            = "global"
	entityLock        = "false"
	dashboardURL      = "https://cloud.ibm.com/objectstorage/crn%3Av1%3Abluemix%3Apublic%3Acloud-object-storage%3Aglobal%3Aa%2F0b5a00334eaf9eb9339d2a0008f20d7f5%3A614500000-7ae6-4755-a5ae-83a8dd806ee4%3A%3A"
	parameters        = map[string]interface{}{
		"par1": "value1",
		"par2": "value2",
	}
	lastOperation = map[string]interface{}{
		"type":        "create",
		"state":       "succeeded",
		"async":       false,
		"description": "Completed create instance operation",
	}
	startDate, _        = strfmt.ParseDateTime("2020-10-27T14:53:07.001933907Z")
	resourceAliasesURL  = "/v2/resource_instances/614566d9-7ae6-4755-a5ae-83a8dd806ee4/resource_aliases"
	resourceBindingsURL = "/v2/resource_instances/614566d9-7ae6-4755-a5ae-83a8dd806ee4/resource_bindings"
	resourceGroupCrn    = "crn:v1:bluemix:public:resource-controller::a/0b5a00334eaf9eb9339d2ab48f20d7f5::resource-group:80bd19ee87314085bb8ac243e6e010d9"
	resourceKeysURL     = "/v2/resource_instances/614566d9-7ae6-4755-a5ae-83a8dd806ee4/resource_keys"
	targetCrn           = "crn:v1:bluemix:public:globalcatalog::::deployment:744bfc56-d12c-4866-88d5-dac9139e0e5d%3Aglobal"
	rcType              = "service_instance"
	url                 = "/v2/resource_instances/614566d9-7ae6-4755-a5ae-83a8dd806ee4"
	createdBy           = "user0001"
)

func params(m ...func(*v1alpha1.ResourceInstanceParameters)) *v1alpha1.ResourceInstanceParameters {
	p := &v1alpha1.ResourceInstanceParameters{
		Name:              instName,
		EntityLock:        reference.ToPtrValue(entityLock),
		AllowCleanup:      ibmc.BoolPtr(allowCleanup),
		Parameters:        ibmc.MapToRawExtension(parameters),
		ResourceGroupName: resourceGroupName,
		ResourcePlanName:  resourcePlanName,
		ServiceName:       serviceName,
		Tags:              tags,
		Target:            target,
	}

	for _, f := range m {
		f(p)
	}
	return p
}

func observation(m ...func(*v1alpha1.ResourceInstanceObservation)) *v1alpha1.ResourceInstanceObservation {
	o := &v1alpha1.ResourceInstanceObservation{
		AccountID:     accountID,
		CreatedAt:     ibmc.DateTimeToMetaV1Time(&createdAt),
		Crn:           crnRes,
		DashboardURL:  dashboardURL,
		GUID:          guid,
		ID:            id,
		LastOperation: ibmc.MapToRawExtension(lastOperation),
		DeletedAt:     nil,
		PlanHistory: []v1alpha1.PlanHistoryItem{
			{
				ResourcePlanID: resourcePlanID,
				StartDate:      ibmc.DateTimeToMetaV1Time(&startDate),
			},
		},
		ResourceAliasesURL:  resourceAliasesURL,
		ResourceBindingsURL: resourceBindingsURL,
		ResourceGroupCrn:    resourceGroupCrn,
		ResourceGroupID:     resourceGroupID,
		ResourceID:          resourceID,
		ResourceKeysURL:     resourceKeysURL,
		ResourcePlanID:      resourcePlanID,
		State:               state,
		SubType:             "",
		TargetCrn:           targetCrn,
		Type:                rcType,
		URL:                 url,
		UpdatedAt:           ibmc.DateTimeToMetaV1Time(&createdAt),
		CreatedBy:           createdBy,
		UpdatedBy:           createdBy,
	}

	for _, f := range m {
		f(o)
	}
	return o
}

func instance(m ...func(*rcv2.ResourceInstance)) *rcv2.ResourceInstance {
	i := &rcv2.ResourceInstance{
		AccountID:     &accountID,
		AllowCleanup:  &allowCleanup,
		CreatedAt:     &createdAt,
		CreatedBy:     &createdBy,
		Crn:           &crnRes,
		DashboardURL:  &dashboardURL,
		DeletedAt:     nil,
		DeletedBy:     nil,
		Guid:          &guid,
		ID:            &id,
		LastOperation: lastOperation,
		Locked:        &locked,
		Name:          &instName,
		Parameters:    parameters,
		PlanHistory: []rcv2.PlanHistoryItem{
			{
				ResourcePlanID: &resourcePlanID,
				StartDate:      &startDate,
			},
		},
		ResourceAliasesURL:  &resourceAliasesURL,
		ResourceBindingsURL: &resourceBindingsURL,
		ResourceGroupCrn:    &resourceGroupCrn,
		ResourceGroupID:     &resourceGroupID,
		ResourceID:          &resourceID,
		ResourceKeysURL:     &resourceKeysURL,
		ResourcePlanID:      &resourcePlanID,
		RestoredAt:          nil,
		RestoredBy:          nil,
		ScheduledReclaimAt:  nil,
		ScheduledReclaimBy:  nil,
		State:               &state,
		SubType:             nil,
		TargetCrn:           &targetCrn,
		Type:                &rcType,
		URL:                 &url,
		UpdatedAt:           &createdAt,
		UpdatedBy:           &createdBy,
	}
	for _, f := range m {
		f(i)
	}
	return i
}

func instanceOpts(m ...func(*rcv2.CreateResourceInstanceOptions)) *rcv2.CreateResourceInstanceOptions {
	i := &rcv2.CreateResourceInstanceOptions{
		EntityLock:     &entityLock,
		AllowCleanup:   &allowCleanup,
		Name:           &instName,
		Parameters:     parameters,
		ResourceGroup:  &resourceGroupID,
		ResourcePlanID: &resourcePlanID,
		Tags:           tags,
		Target:         &target,
	}
	for _, f := range m {
		f(i)
	}
	return i
}

func instanceUpdOpts(m ...func(*rcv2.UpdateResourceInstanceOptions)) *rcv2.UpdateResourceInstanceOptions {
	i := &rcv2.UpdateResourceInstanceOptions{
		AllowCleanup:   &allowCleanup,
		ID:             &id,
		Name:           &instName,
		Parameters:     parameters,
		ResourcePlanID: &resourcePlanID,
	}
	for _, f := range m {
		f(i)
	}
	return i
}

// handler to mock client SDK call to global tags API
var tagsHandler = func(w http.ResponseWriter, r *http.Request) {
	_ = r.Body.Close()
	w.Header().Set("Content-Type", "application/json")
	tags := gtagv1.TagList{
		Items: []gtagv1.Tag{
			{
				Name: reference.ToPtrValue("dev"),
			},
		},
	}
	_ = json.NewEncoder(w).Encode(tags)
}

// handler to mock client SDK call to resource manager API
var rgHandler = func(w http.ResponseWriter, r *http.Request) {
	_ = r.Body.Close()
	w.Header().Set("Content-Type", "application/json")
	rgl := rmgrv2.ResourceGroupList{
		Resources: []rmgrv2.ResourceGroup{
			{
				ID:   reference.ToPtrValue(resourceGroupID),
				Name: reference.ToPtrValue(resourceGroupName),
			},
		},
	}
	_ = json.NewEncoder(w).Encode(rgl)
}

// handler to mock client SDK call to global catalog API for plans
var pcatHandler = func(w http.ResponseWriter, r *http.Request) {
	_ = r.Body.Close()
	w.Header().Set("Content-Type", "application/json")
	planEntry := gcat.EntrySearchResult{
		Resources: []gcat.CatalogEntry{
			{
				ID:   reference.ToPtrValue(resourcePlanID),
				Name: reference.ToPtrValue(resourcePlanName),
			},
		},
	}
	_ = json.NewEncoder(w).Encode(planEntry)
}

// handler to mock client SDK call to global catalog API for services
var svcatHandler = func(w http.ResponseWriter, r *http.Request) {
	_ = r.Body.Close()
	w.Header().Set("Content-Type", "application/json")
	catEntry := gcat.EntrySearchResult{
		Resources: []gcat.CatalogEntry{
			{
				Metadata: &gcat.CatalogEntryMetadata{
					Ui: &gcat.UIMetaData{
						PrimaryOfferingID: reference.ToPtrValue(serviceName),
					},
				},
			},
		},
	}
	_ = json.NewEncoder(w).Encode(catEntry)
}

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
			want: want{instance: instanceOpts()},
		},
		"MissingFields": {
			args: args{
				params: *params(func(p *v1alpha1.ResourceInstanceParameters) {
					p.AllowCleanup = nil
					p.EntityLock = nil
					p.Tags = nil
				})},
			want: want{instance: instanceOpts(func(ri *rcv2.CreateResourceInstanceOptions) {
				ri.EntityLock = nil
				ri.AllowCleanup = nil
				ri.Tags = nil
			})},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			mux := http.NewServeMux()
			mux.HandleFunc("/resource_groups/", rgHandler)
			mux.HandleFunc("/", svcatHandler)
			mux.HandleFunc("/"+serviceName+"/", pcatHandler)
			server := httptest.NewServer(mux)
			defer server.Close()

			opts := ibmc.ClientOptions{URL: server.URL, Authenticator: &core.BearerTokenAuthenticator{
				BearerToken: bearerTok,
			}}
			mClient, _ := ibmc.NewClient(opts)

			r := &rcv2.CreateResourceInstanceOptions{}
			GenerateCreateResourceInstanceOptions(mClient, tc.args.params, r)
			if diff := cmp.Diff(tc.want.instance, r); diff != "" {
				t.Errorf("GenerateCreateResourceInstanceOptions(...): -want, +got:\n%s", diff)
			}
		})
	}
}

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
			want: want{instance: instanceUpdOpts()},
		},
		"MissingFields": {
			args: args{
				params: *params(func(p *v1alpha1.ResourceInstanceParameters) {
					p.AllowCleanup = nil
				})},
			want: want{instance: instanceUpdOpts(func(ri *rcv2.UpdateResourceInstanceOptions) {
				ri.AllowCleanup = nil
			})},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			mux := http.NewServeMux()
			mux.HandleFunc("/resource_groups/", rgHandler)
			mux.HandleFunc("/", svcatHandler)
			mux.HandleFunc("/"+serviceName+"/", pcatHandler)
			server := httptest.NewServer(mux)
			defer server.Close()

			opts := ibmc.ClientOptions{URL: server.URL, Authenticator: &core.BearerTokenAuthenticator{
				BearerToken: bearerTok,
			}}
			mClient, _ := ibmc.NewClient(opts)

			r := &rcv2.UpdateResourceInstanceOptions{}
			GenerateUpdateResourceInstanceOptions(mClient, id, tc.args.params, r)
			if diff := cmp.Diff(tc.want.instance, r); diff != "" {
				t.Errorf("GenerateUpdateResourceInstanceOptions(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestLateInitializeSpecs(t *testing.T) {
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
					p.EntityLock = nil
					p.AllowCleanup = nil
					p.Tags = nil
				}),
				instance: instance(func(i *rcv2.ResourceInstance) {
					i.Locked = &locked
					i.AllowCleanup = &allowCleanup
					i.ResourceGroupID = &resourceGroupID
				}),
			},
			want: want{
				params: params(func(p *v1alpha1.ResourceInstanceParameters) {
					p.EntityLock = &entityLock
					p.AllowCleanup = &allowCleanup
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
			mux.HandleFunc("/resource_groups/", rgHandler)
			mux.HandleFunc("/", svcatHandler)
			mux.HandleFunc("/v3/tags/", tagsHandler)
			mux.HandleFunc("/"+serviceName+"/", pcatHandler)
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

func TestGenerateObservation(t *testing.T) {
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
			mux.HandleFunc("/v3/tags/", tagsHandler)
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

func TestIsUpToDate(t *testing.T) {
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
					i.Parameters = map[string]interface{}{
						"par1": "old-value",
					}
				}),
			},
			want: want{upToDate: false, isErr: false},
		},
		"NeedsUpdateOnName": {
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
			mux.HandleFunc("/resource_groups/", rgHandler)
			mux.HandleFunc("/", svcatHandler)
			mux.HandleFunc("/v3/tags/", tagsHandler)
			mux.HandleFunc("/"+serviceName+"/", pcatHandler)
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
