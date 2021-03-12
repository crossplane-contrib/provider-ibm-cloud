/*
Copyright 2019 The Crossplane Authors.

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

package resourcecontrollerv2

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-openapi/strfmt"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	cpv1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/reference"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/crossplane/crossplane-runtime/pkg/test"

	"github.com/IBM/go-sdk-core/core"
	gcat "github.com/IBM/platform-services-go-sdk/globalcatalogv1"
	gtagv1 "github.com/IBM/platform-services-go-sdk/globaltaggingv1"
	rcv2 "github.com/IBM/platform-services-go-sdk/resourcecontrollerv2"
	rmgrv2 "github.com/IBM/platform-services-go-sdk/resourcemanagerv2"

	"github.com/crossplane-contrib/provider-ibm-cloud/apis/resourcecontrollerv2/v1alpha1"
	ibmc "github.com/crossplane-contrib/provider-ibm-cloud/pkg/clients"
	ibmcc "github.com/crossplane-contrib/provider-ibm-cloud/pkg/clients/resourceinstance"
)

const (
	bearerTok  = "mock-token"
	wtfConst   = "crossplane.io/external-name"
	errNoRCDep = "No RC deployments for plan: standard with target wrong-target"
)

var (
	resourcePlanName  = "standard"
	resourceGroupName = "default"
	allowCleanup      = false
	crn               = "crn:v1:bluemix:public:cloud-object-storage:global:a/0b5a00334eaf9eb9339d2ab48f20d7f5:78d88b2b-bbbb-aaaa-8888-5c26e8b6a555::"
	guid              = "78d88b2b-bbbb-aaaa-8888-5c26e8b6a555"
	id                = "crn:v1:bluemix:public:cloud-object-storage:global:a/0b5a00334eaf9eb9339d2ab48f20d7f5:78d88b2b-bbbb-aaaa-8888-5c26e8b6a555::"
	locked            = false
	name              = "cos-wow"
	createdAt, _      = strfmt.ParseDateTime("2020-10-31T02:33:06Z")
	resourceGroupID   = "mock-resource-group-id"
	resourceID        = "aaaaaaaa-bbbb-cccc-b470-411c3edbe49c"
	resourcePlanID    = "744bfc56-d12c-4866-88d5-dac9139e0e5d"
	state             = ibmcc.StateActive
	serviceName       = "cloud-object-storage"
	tags              = []string{"dev"}
	target            = "global"
	entityLock        = false
	parameters        = map[string]interface{}{}
)

var _ managed.ExternalConnecter = &riConnector{}
var _ managed.ExternalClient = &riExternal{}

type handler struct {
	path        string
	handlerFunc func(w http.ResponseWriter, r *http.Request)
}

type instanceModifier func(*v1alpha1.ResourceInstance)

func withConditions(c ...cpv1alpha1.Condition) instanceModifier {
	return func(i *v1alpha1.ResourceInstance) { i.Status.SetConditions(c...) }
}

func withState(s string) instanceModifier {
	return func(i *v1alpha1.ResourceInstance) { i.Status.AtProvider.State = s }
}

func withID(s string) instanceModifier {
	return func(i *v1alpha1.ResourceInstance) { i.Status.AtProvider.ID = s }
}

func withGUID(s string) instanceModifier {
	return func(i *v1alpha1.ResourceInstance) { i.Status.AtProvider.GUID = s }
}

func withCRN(s string) instanceModifier {
	return func(i *v1alpha1.ResourceInstance) { i.Status.AtProvider.CRN = s }
}

func withResourceID(s string) instanceModifier {
	return func(i *v1alpha1.ResourceInstance) { i.Status.AtProvider.ResourceID = s }
}

func withResourcePlanID(s string) instanceModifier {
	return func(i *v1alpha1.ResourceInstance) { i.Status.AtProvider.ResourcePlanID = s }
}

func withResourceGroupID(s string) instanceModifier {
	return func(i *v1alpha1.ResourceInstance) { i.Status.AtProvider.ResourceGroupID = s }
}

func withCreatedAt(t strfmt.DateTime) instanceModifier {
	return func(i *v1alpha1.ResourceInstance) {
		i.Status.AtProvider.CreatedAt = ibmc.DateTimeToMetaV1Time(&t)
	}
}

func withExternalNameAnnotation(externalName string) instanceModifier {
	return func(i *v1alpha1.ResourceInstance) {
		if i.ObjectMeta.Annotations == nil {
			i.ObjectMeta.Annotations = make(map[string]string)
		}
		i.ObjectMeta.Annotations[wtfConst] = externalName
	}
}

func withSpec(p v1alpha1.ResourceInstanceParameters) instanceModifier {
	return func(r *v1alpha1.ResourceInstance) { r.Spec.ForProvider = p }
}

func instance(im ...instanceModifier) *v1alpha1.ResourceInstance {
	i := &v1alpha1.ResourceInstance{
		ObjectMeta: metav1.ObjectMeta{
			Name:       name,
			Finalizers: []string{},
			Annotations: map[string]string{
				meta.AnnotationKeyExternalName: name,
			},
		},
		Spec: v1alpha1.ResourceInstanceSpec{
			ForProvider: v1alpha1.ResourceInstanceParameters{},
		},
	}
	for _, m := range im {
		m(i)
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

// handler to mock client SDK call to global catalog API for services
var svcatHandler = func(w http.ResponseWriter, r *http.Request) {
	_ = r.Body.Close()
	w.Header().Set("Content-Type", "application/json")
	catEntry := gcat.EntrySearchResult{
		Resources: []gcat.CatalogEntry{
			{
				Metadata: &gcat.CatalogEntryMetadata{
					UI: &gcat.UIMetaData{
						PrimaryOfferingID: reference.ToPtrValue(serviceName),
					},
				},
			},
		},
	}
	_ = json.NewEncoder(w).Encode(catEntry)
}

func listResourceInstancesNoItems(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = r.Body.Close()
	_ = json.NewEncoder(w).Encode(&rcv2.ResourceInstancesList{RowsCount: ibmc.Int64Ptr(0)})
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

func resourceInstanceSpec() v1alpha1.ResourceInstanceParameters {
	o := v1alpha1.ResourceInstanceParameters{
		Name:              name,
		EntityLock:        &entityLock,
		AllowCleanup:      ibmc.BoolPtr(false),
		Parameters:        ibmc.MapToRawExtension(parameters),
		ResourceGroupName: resourceGroupName,
		ResourcePlanName:  resourcePlanName,
		ServiceName:       serviceName,
		Tags:              tags,
		Target:            target,
	}
	return o
}

func genTestSDKResourceInstance() *rcv2.ResourceInstance {
	i := &rcv2.ResourceInstance{
		AllowCleanup:    &allowCleanup,
		CreatedAt:       &createdAt,
		CRN:             &crn,
		GUID:            &guid,
		ID:              &id,
		Locked:          &locked,
		Name:            &name,
		ResourceGroupID: &resourceGroupID,
		ResourceID:      &resourceID,
		ResourcePlanID:  &resourcePlanID,
		State:           &state,
		Parameters:      parameters,
	}
	return i
}

func genTestCRResourceInstance(im ...instanceModifier) *v1alpha1.ResourceInstance {
	i := instance(
		withExternalNameAnnotation(id),
		withSpec(resourceInstanceSpec()),
		withID(id),
		withCRN(crn),
		withGUID(guid),
		withState(state),
		withResourcePlanID(resourcePlanID),
		withResourceGroupID(resourceGroupID),
		withResourceID(resourceID),
		withCreatedAt(createdAt),
		withConditions(cpv1alpha1.Available()),
	)
	for _, m := range im {
		m(i)
	}
	return i
}

func TestObserve(t *testing.T) {
	type args struct {
		mg resource.Managed
	}
	type want struct {
		mg  resource.Managed
		obs managed.ExternalObservation
		err error
	}
	cases := map[string]struct {
		handlers []handler
		kube     client.Client
		args     args
		want     want
	}{
		"NotFound": {
			handlers: []handler{
				{
					path: "/",
					handlerFunc: func(w http.ResponseWriter, r *http.Request) {
						_ = r.Body.Close()
						if diff := cmp.Diff(http.MethodGet, r.Method); diff != "" {
							t.Errorf("r: -want, +got:\n%s", diff)
						}
						// content type should always set before writeHeader()
						w.Header().Set("Content-Type", "application/json")
						w.WriteHeader(http.StatusNotFound)
						_ = json.NewEncoder(w).Encode(&rcv2.ResourceInstance{})
					},
				},
			},
			args: args{
				mg: instance(),
			},
			want: want{
				mg:  instance(),
				err: nil,
			},
		},
		"GetFailed": {
			handlers: []handler{
				{
					path: "/",
					handlerFunc: func(w http.ResponseWriter, r *http.Request) {
						_ = r.Body.Close()
						if diff := cmp.Diff(http.MethodGet, r.Method); diff != "" {
							t.Errorf("r: -want, +got:\n%s", diff)
						}
						w.Header().Set("Content-Type", "application/json")
						w.WriteHeader(http.StatusBadRequest)
						_ = json.NewEncoder(w).Encode(&rcv2.ResourceInstance{})
					},
				},
			},
			args: args{
				mg: instance(),
			},
			want: want{
				mg:  instance(),
				err: errors.New(errBadRequest),
			},
		},
		"ObservedResourceInstanceUpToDate": {
			handlers: []handler{
				{
					path: "/v2/resource_instances/",
					handlerFunc: func(w http.ResponseWriter, r *http.Request) {
						_ = r.Body.Close()
						if diff := cmp.Diff(http.MethodGet, r.Method); diff != "" {
							t.Errorf("r: -want, +got:\n%s", diff)
						}
						w.Header().Set("Content-Type", "application/json")
						ri := genTestSDKResourceInstance()
						_ = json.NewEncoder(w).Encode(ri)
					},
				},
				{
					path:        "/v3/tags/",
					handlerFunc: tagsHandler,
				},
				{
					path:        "/resource_groups/",
					handlerFunc: rgHandler,
				},
				{
					path:        "/",
					handlerFunc: svcatHandler,
				},
				{
					path:        "/" + serviceName + "/",
					handlerFunc: pcatHandler,
				},
			},
			kube: &test.MockClient{
				MockUpdate: test.NewMockUpdateFn(nil),
			},
			args: args{
				mg: instance(
					withExternalNameAnnotation(id),
					withID(id),
					withSpec(resourceInstanceSpec()),
				),
			},
			want: want{
				mg: genTestCRResourceInstance(),
				obs: managed.ExternalObservation{
					ResourceExists:    true,
					ResourceUpToDate:  true,
					ConnectionDetails: nil,
				},
			},
		},
		"ObservedResourceInstanceNotUpToDate": {
			handlers: []handler{
				{
					path: "/v2/resource_instances/",
					handlerFunc: func(w http.ResponseWriter, r *http.Request) {
						_ = r.Body.Close()
						if diff := cmp.Diff(http.MethodGet, r.Method); diff != "" {
							t.Errorf("r: -want, +got:\n%s", diff)
						}
						w.Header().Set("Content-Type", "application/json")
						ri := genTestSDKResourceInstance()
						ri.Locked = ibmc.BoolPtr(true)
						_ = json.NewEncoder(w).Encode(ri)
					},
				},
				{
					path:        "/v3/tags/",
					handlerFunc: tagsHandler,
				},
				{
					path:        "/resource_groups/",
					handlerFunc: rgHandler,
				},
				{
					path:        "/",
					handlerFunc: svcatHandler,
				},
				{
					path:        "/" + serviceName + "/",
					handlerFunc: pcatHandler,
				},
			},
			kube: &test.MockClient{
				MockUpdate: test.NewMockUpdateFn(nil),
			},
			args: args{
				mg: instance(
					withExternalNameAnnotation(id),
					withID(id),
					withSpec(resourceInstanceSpec()),
				),
			},
			want: want{
				mg: genTestCRResourceInstance(),
				obs: managed.ExternalObservation{
					ResourceExists:    true,
					ResourceUpToDate:  false,
					ConnectionDetails: nil,
				},
			},
		},
		"ObservedResourceInstancePendingReclamation": {
			handlers: []handler{
				{
					path: "/v2/resource_instances/",
					handlerFunc: func(w http.ResponseWriter, r *http.Request) {
						_ = r.Body.Close()
						if diff := cmp.Diff(http.MethodGet, r.Method); diff != "" {
							t.Errorf("r: -want, +got:\n%s", diff)
						}
						w.Header().Set("Content-Type", "application/json")
						ri := genTestSDKResourceInstance()
						ri.State = reference.ToPtrValue(ibmcc.StatePendingReclamation)
						_ = json.NewEncoder(w).Encode(ri)
					},
				},
			},
			args: args{
				mg: instance(),
			},
			want: want{
				mg: instance(),
				obs: managed.ExternalObservation{
					ResourceExists:    false,
					ResourceUpToDate:  false,
					ConnectionDetails: nil,
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			mux := http.NewServeMux()
			for _, h := range tc.handlers {
				mux.HandleFunc(h.path, h.handlerFunc)
			}
			server := httptest.NewServer(mux)
			defer server.Close()

			opts := ibmc.ClientOptions{URL: server.URL, Authenticator: &core.BearerTokenAuthenticator{
				BearerToken: bearerTok,
			}}
			mClient, _ := ibmc.NewClient(opts)
			e := riExternal{
				kube:   tc.kube,
				client: mClient,
				logger: logging.NewNopLogger(),
			}
			obs, err := e.Observe(context.Background(), tc.args.mg)
			if tc.want.err != nil && err != nil {
				// the case where our mock server returns error.
				if diff := cmp.Diff(tc.want.err.Error(), err.Error()); diff != "" {
					t.Errorf("Observe(...): want error string != got error string:\n%s", diff)
				}
			} else {
				if diff := cmp.Diff(tc.want.err, err); diff != "" {
					t.Errorf("Observe(...): want error != got error:\n%s", diff)
				}
			}
			if diff := cmp.Diff(tc.want.obs, obs); diff != "" {
				t.Errorf("Observe(...): -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.mg, tc.args.mg); diff != "" {
				t.Errorf("Observe(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestCreate(t *testing.T) {
	type args struct {
		mg resource.Managed
	}
	type want struct {
		mg  resource.Managed
		cre managed.ExternalCreation
		err error
	}
	cases := map[string]struct {
		handlers []handler
		kube     client.Client
		args     args
		want     want
	}{
		"Successful": {
			handlers: []handler{
				{
					path: "/v2/resource_instances",
					handlerFunc: func(w http.ResponseWriter, r *http.Request) {
						if r.Method == http.MethodGet {
							listResourceInstancesNoItems(w, r)
							return
						}
						if diff := cmp.Diff(http.MethodPost, r.Method); diff != "" {
							t.Errorf("r: -want, +got:\n%s", diff)
						}
						w.Header().Set("Content-Type", "application/json")
						w.WriteHeader(http.StatusCreated)
						_ = r.Body.Close()
						ri := genTestSDKResourceInstance()
						_ = json.NewEncoder(w).Encode(ri)
					},
				},
				{
					path:        "/resource_groups/",
					handlerFunc: rgHandler,
				},
				{
					path:        "/v3/tags/",
					handlerFunc: tagsHandler,
				},
				{
					path:        "/",
					handlerFunc: svcatHandler,
				},
				{
					path:        "/" + serviceName + "/",
					handlerFunc: pcatHandler,
				},
			},
			args: args{
				mg: instance(withSpec(resourceInstanceSpec())),
			},
			want: want{
				mg: instance(withSpec(resourceInstanceSpec()),
					withConditions(cpv1alpha1.Creating()),
					withExternalNameAnnotation(id)),
				cre: managed.ExternalCreation{ExternalNameAssigned: true},
				err: nil,
			},
		},
		"Failed": {
			handlers: []handler{
				{
					path: "/v2/resource_instances",
					handlerFunc: func(w http.ResponseWriter, r *http.Request) {
						if r.Method == http.MethodGet {
							listResourceInstancesNoItems(w, r)
							return
						}

						if diff := cmp.Diff(http.MethodPost, r.Method); diff != "" {
							t.Errorf("r: -want, +got:\n%s", diff)
						}
						w.Header().Set("Content-Type", "application/json")
						w.WriteHeader(http.StatusBadRequest)
						_ = r.Body.Close()

						b := map[string]interface{}{
							"message":     errNoRCDep,
							"status_code": 400,
						}
						_ = json.NewEncoder(w).Encode(&b)
					},
				},
				{
					path:        "/resource_groups/",
					handlerFunc: rgHandler,
				},
				{
					path:        "/",
					handlerFunc: svcatHandler,
				},
				{
					path:        "/" + serviceName + "/",
					handlerFunc: pcatHandler,
				},
			},
			args: args{
				mg: instance(withSpec(resourceInstanceSpec())),
			},
			want: want{
				mg:  instance(withSpec(resourceInstanceSpec()), withConditions(cpv1alpha1.Creating())),
				err: errors.Wrap(errors.New(errNoRCDep), errCreateRes),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			mux := http.NewServeMux()
			for _, h := range tc.handlers {
				mux.HandleFunc(h.path, h.handlerFunc)
			}
			server := httptest.NewServer(mux)
			defer server.Close()

			opts := ibmc.ClientOptions{URL: server.URL, Authenticator: &core.BearerTokenAuthenticator{
				BearerToken: bearerTok,
			}}
			mClient, _ := ibmc.NewClient(opts)
			e := riExternal{
				kube:   tc.kube,
				client: mClient,
				logger: logging.NewNopLogger(),
			}
			cre, err := e.Create(context.Background(), tc.args.mg)
			if tc.want.err != nil && err != nil {
				// the case where our mock server returns error.
				if diff := cmp.Diff(tc.want.err.Error(), err.Error()); diff != "" {
					t.Errorf("Create(...): -want, +got:\n%s", diff)
				}
			} else {
				if diff := cmp.Diff(tc.want.err, err); diff != "" {
					t.Errorf("Create(...): -want, +got:\n%s", diff)
				}
			}
			if diff := cmp.Diff(tc.want.cre, cre); diff != "" {
				t.Errorf("Create(...): -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.mg, tc.args.mg); diff != "" {
				t.Errorf("Create(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestDelete(t *testing.T) {
	type args struct {
		mg resource.Managed
	}
	type want struct {
		mg  resource.Managed
		err error
	}
	cases := map[string]struct {
		handlers []handler
		kube     client.Client
		args     args
		want     want
	}{
		"Successful": {
			handlers: []handler{
				{
					path: "/v2/resource_instances/",
					handlerFunc: func(w http.ResponseWriter, r *http.Request) {
						if diff := cmp.Diff(http.MethodDelete, r.Method); diff != "" {
							t.Errorf("r: -want, +got:\n%s", diff)
						}
						w.Header().Set("Content-Type", "application/json")
						w.WriteHeader(http.StatusAccepted)
						_ = r.Body.Close()
					},
				},
			},
			args: args{
				mg: instance(withID(id)),
			},
			want: want{
				mg:  instance(withID(id), withConditions(cpv1alpha1.Deleting())),
				err: nil,
			},
		},
		"AlreadyGone": {
			handlers: []handler{
				{
					path: "/v2/resource_instances/",
					handlerFunc: func(w http.ResponseWriter, r *http.Request) {
						if diff := cmp.Diff(http.MethodDelete, r.Method); diff != "" {
							t.Errorf("r: -want, +got:\n%s", diff)
						}
						w.Header().Set("Content-Type", "application/json")
						w.WriteHeader(http.StatusNotFound)
						_ = r.Body.Close()
					},
				},
			},
			args: args{
				mg: instance(withID(id)),
			},
			want: want{
				mg:  instance(withID(id), withConditions(cpv1alpha1.Deleting())),
				err: nil,
			},
		},
		"Failed": {
			handlers: []handler{
				{
					path: "/v2/resource_instances/",
					handlerFunc: func(w http.ResponseWriter, r *http.Request) {
						if diff := cmp.Diff(http.MethodDelete, r.Method); diff != "" {
							t.Errorf("r: -want, +got:\n%s", diff)
						}
						w.Header().Set("Content-Type", "application/json")
						w.WriteHeader(http.StatusBadRequest)
						_ = r.Body.Close()
					},
				},
			},
			args: args{
				mg: instance(withID(id)),
			},
			want: want{
				mg:  instance(withID(id), withConditions(cpv1alpha1.Deleting())),
				err: errors.Wrap(errors.New(http.StatusText(http.StatusBadRequest)), errDeleteRes),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			mux := http.NewServeMux()
			for _, h := range tc.handlers {
				mux.HandleFunc(h.path, h.handlerFunc)
			}
			server := httptest.NewServer(mux)
			defer server.Close()

			opts := ibmc.ClientOptions{URL: server.URL, Authenticator: &core.BearerTokenAuthenticator{
				BearerToken: bearerTok,
			}}
			mClient, _ := ibmc.NewClient(opts)
			e := riExternal{
				kube:   tc.kube,
				client: mClient,
				logger: logging.NewNopLogger(),
			}
			err := e.Delete(context.Background(), tc.args.mg)
			if tc.want.err != nil && err != nil {
				// the case where our mock server returns error.
				if diff := cmp.Diff(tc.want.err.Error(), err.Error()); diff != "" {
					t.Errorf("Delete(...): -want, +got:\n%s", diff)
				}
			} else {
				if diff := cmp.Diff(tc.want.err, err); diff != "" {
					t.Errorf("Delete(...): -want, +got:\n%s", diff)
				}
			}
			if diff := cmp.Diff(tc.want.mg, tc.args.mg); diff != "" {
				t.Errorf("Delete(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestUpdate(t *testing.T) {
	type args struct {
		mg resource.Managed
	}
	type want struct {
		mg  resource.Managed
		upd managed.ExternalUpdate
		err error
	}
	cases := map[string]struct {
		handlers []handler
		kube     client.Client
		args     args
		want     want
	}{
		"Successful": {
			handlers: []handler{
				{
					path: "/v2/resource_instances/",
					handlerFunc: func(w http.ResponseWriter, r *http.Request) {
						if diff := cmp.Diff(http.MethodPatch, r.Method); diff != "" {
							t.Errorf("r: -want, +got:\n%s", diff)
						}
						w.Header().Set("Content-Type", "application/json")
						w.WriteHeader(http.StatusOK)
						_ = r.Body.Close()
						ri := genTestSDKResourceInstance()
						_ = json.NewEncoder(w).Encode(ri)
					},
				},
				{
					path:        "/",
					handlerFunc: svcatHandler,
				},
				{
					path:        "/" + serviceName + "/",
					handlerFunc: pcatHandler,
				},
			},
			args: args{
				mg: genTestCRResourceInstance(withSpec(resourceInstanceSpec())),
			},
			want: want{
				mg:  genTestCRResourceInstance(),
				upd: managed.ExternalUpdate{ConnectionDetails: nil},
				err: nil,
			},
		},
		"PatchFails": {
			handlers: []handler{
				{
					path: "/v2/resource_instances/",
					handlerFunc: func(w http.ResponseWriter, r *http.Request) {
						if diff := cmp.Diff(http.MethodPatch, r.Method); diff != "" {
							t.Errorf("r: -want, +got:\n%s", diff)
						}
						w.Header().Set("Content-Type", "application/json")
						w.WriteHeader(http.StatusBadRequest)
						_ = r.Body.Close()
					},
				},
				{
					path:        "/",
					handlerFunc: svcatHandler,
				},
				{
					path:        "/" + serviceName + "/",
					handlerFunc: pcatHandler,
				},
			},

			args: args{
				mg: genTestCRResourceInstance(withSpec(resourceInstanceSpec())),
			},
			want: want{
				mg:  genTestCRResourceInstance(),
				err: errors.Wrap(errors.New(http.StatusText(http.StatusBadRequest)), errUpdRes),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			mux := http.NewServeMux()
			for _, h := range tc.handlers {
				mux.HandleFunc(h.path, h.handlerFunc)
			}
			server := httptest.NewServer(mux)
			defer server.Close()

			opts := ibmc.ClientOptions{URL: server.URL, Authenticator: &core.BearerTokenAuthenticator{
				BearerToken: bearerTok,
			}}
			mClient, _ := ibmc.NewClient(opts)
			e := riExternal{
				kube:   tc.kube,
				client: mClient,
				logger: logging.NewNopLogger(),
			}
			upd, err := e.Update(context.Background(), tc.args.mg)
			if tc.want.err != nil && err != nil {
				// the case where our mock server returns error.
				if diff := cmp.Diff(tc.want.err.Error(), err.Error()); diff != "" {
					t.Errorf("Update(...): -want, +got:\n%s", diff)
				}
			} else {
				if diff := cmp.Diff(tc.want.err, err); diff != "" {
					t.Errorf("Update(...): -want, +got:\n%s", diff)
				}
			}
			if tc.want.err == nil {
				if diff := cmp.Diff(tc.want.mg, tc.args.mg); diff != "" {
					t.Errorf("Update(...): -want, +got:\n%s", diff)
				}
				if diff := cmp.Diff(tc.want.upd, upd); diff != "" {
					t.Errorf("Update(...): -want, +got:\n%s", diff)
				}
			}
		})
	}
}
