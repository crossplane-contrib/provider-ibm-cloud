/*
Copyright 2019 The Crossplane Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance rkWith the License.
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
	"fmt"
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

	rcv2 "github.com/IBM/platform-services-go-sdk/resourcecontrollerv2"

	"github.com/crossplane-contrib/provider-ibm-cloud/apis/resourcecontrollerv2/v1alpha1"
	ibmc "github.com/crossplane-contrib/provider-ibm-cloud/pkg/clients"
	"github.com/crossplane-contrib/provider-ibm-cloud/pkg/controller/tstutil"
)

var (
	role          = "Manager"
	role2         = "Viewer"
	rkName        = "cos-creds"
	rkID          = "crn:v1:bluemix:public:cloud-object-storage:global:a/0b5a00334eaf9eb9339d2ab48f20d7f5:f931e669-6c11-4d4d-b720-8b2f844a6d9e:resource-key:bbeca5fe-283f-443c-9aca-cd3f72c6f493"
	createdBy     = "user00001"
	iamCompatible = true
	resInstURL    = "/v2/resource_keys/614566d9-7ae6-4755-a5ae-83a8dd806ee4"
	sourceCrn     = "crn:v1:bluemix:public:cloud-object-storage:global:a/0b5a00334eaf9eb9339d2ab48f20d7f5:78d88b2b-bbbb-aaaa-8888-5c26e8b6a555::"
	rkCrn         = "crn:v1:bluemix:public:key:global:a/0b5a00334eaf9eb9339d2ab48f20d7f5:78d88b2b-bbbb-aaaa-8888-5c26e8b6a555::"
	accountID     = "fake-account-id"
	url           = "/v2/resource_keys/614566d9-7ae6-4755-a5ae-83a8dd806ee4"
	wrongGUID     = "wrong-guid"
	errWrongGUID  = fmt.Sprintf("Failed to retrieve an alias with guid: %s", wrongGUID)
)

var _ managed.ExternalConnecter = &resourcekeyConnector{}
var _ managed.ExternalClient = &resourcekeyExternal{}

type keyModifier func(*v1alpha1.ResourceKey)

func rkWithConditions(c ...cpv1alpha1.Condition) keyModifier {
	return func(i *v1alpha1.ResourceKey) { i.Status.SetConditions(c...) }
}

func rkWithState(s string) keyModifier {
	return func(i *v1alpha1.ResourceKey) { i.Status.AtProvider.State = s }
}

func rkWithID(s string) keyModifier {
	return func(i *v1alpha1.ResourceKey) { i.Status.AtProvider.ID = s }
}

func rkWithGUID(s string) keyModifier {
	return func(i *v1alpha1.ResourceKey) { i.Status.AtProvider.GUID = s }
}

func rkWithCRN(s string) keyModifier {
	return func(i *v1alpha1.ResourceKey) { i.Status.AtProvider.CRN = s }
}

func rkWithURL(s string) keyModifier {
	return func(i *v1alpha1.ResourceKey) { i.Status.AtProvider.URL = s }
}

func rkWithAccountID(s string) keyModifier {
	return func(i *v1alpha1.ResourceKey) { i.Status.AtProvider.AccountID = s }
}

func rkWihIAMCompatible(b bool) keyModifier {
	return func(i *v1alpha1.ResourceKey) { i.Status.AtProvider.IamCompatible = b }
}

func rkWithCreatedBy(s string) keyModifier {
	return func(i *v1alpha1.ResourceKey) { i.Status.AtProvider.CreatedBy = s }
}

func rkWithResourceGroupID(s string) keyModifier {
	return func(i *v1alpha1.ResourceKey) { i.Status.AtProvider.ResourceGroupID = s }
}

func rkWithSourceCRN(s string) keyModifier {
	return func(i *v1alpha1.ResourceKey) { i.Status.AtProvider.SourceCRN = s }
}

func rkWithResourceInstanceURL(s string) keyModifier {
	return func(i *v1alpha1.ResourceKey) { i.Status.AtProvider.ResourceInstanceURL = s }
}

func rkWithCreatedAt(t strfmt.DateTime) keyModifier {
	return func(i *v1alpha1.ResourceKey) {
		i.Status.AtProvider.CreatedAt = ibmc.DateTimeToMetaV1Time(&t)
	}
}

func rkWithExternalNameAnnotation(externalName string) keyModifier {
	return func(i *v1alpha1.ResourceKey) {
		if i.ObjectMeta.Annotations == nil {
			i.ObjectMeta.Annotations = make(map[string]string)
		}
		i.ObjectMeta.Annotations[wtfConst] = externalName
	}
}

func rkWithSpec(p v1alpha1.ResourceKeyParameters) keyModifier {
	return func(r *v1alpha1.ResourceKey) { r.Spec.ForProvider = p }
}

func key(im ...keyModifier) *v1alpha1.ResourceKey {
	i := &v1alpha1.ResourceKey{
		ObjectMeta: metav1.ObjectMeta{
			Name:       rkName,
			Finalizers: []string{},
			Annotations: map[string]string{
				meta.AnnotationKeyExternalName: rkName,
			},
		},
		Spec: v1alpha1.ResourceKeySpec{
			ForProvider: v1alpha1.ResourceKeyParameters{},
		},
	}
	for _, m := range im {
		m(i)
	}
	return i
}

func resourceKeySpec() v1alpha1.ResourceKeyParameters {
	o := v1alpha1.ResourceKeyParameters{
		Name:   rkName,
		Role:   &role,
		Source: &crn,
	}
	return o
}

func genTestSDKResourceKey() *rcv2.ResourceKey {
	i := &rcv2.ResourceKey{
		CreatedAt:           &createdAt,
		CRN:                 &rkCrn,
		GUID:                &guid,
		ID:                  &rkID,
		Name:                &rkName,
		ResourceGroupID:     &resourceGroupID,
		State:               &state,
		AccountID:           &accountID,
		CreatedBy:           &createdBy,
		IamCompatible:       &iamCompatible,
		Role:                &role,
		ResourceInstanceURL: &resInstURL,
		SourceCRN:           &sourceCrn,
		URL:                 &url,
	}
	return i
}

func genTestCRResourceKey(im ...keyModifier) *v1alpha1.ResourceKey {
	i := key(
		rkWithAccountID(accountID),
		rkWithCreatedAt(createdAt),
		rkWithCRN(rkCrn),
		rkWithGUID(guid),
		rkWithID(rkID),
		rkWithResourceGroupID(resourceGroupID),
		rkWithState(state),
		rkWithURL(url),
		rkWihIAMCompatible(iamCompatible),
		rkWithResourceInstanceURL(resInstURL),
		rkWithCreatedBy(createdBy),
		rkWithConditions(cpv1alpha1.Available()),
		rkWithSpec(resourceKeySpec()),
	)
	for _, m := range im {
		m(i)
	}
	return i
}

func listResourceKeysNoItems(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = r.Body.Close()
	list := &rcv2.ResourceKeysList{
		RowsCount: ibmc.Int64Ptr(0),
		Resources: []rcv2.ResourceKey{},
	}
	_ = json.NewEncoder(w).Encode(list)
}

// Sets up a unit test http server, and creates an external resource key structure appropriate for unit test.
//
// Params
//	   testingObj - the test object
//	   handlers - the handlers that create the responses
//	   client - the controller runtime client
//
// Returns
//		- the external object, ready for unit test
//		- the test http server, on which the caller should call 'defer ....Close()' (reason for this is we need to keep it around to prevent
//		  garbage collection)
//      -- an error (if...)
func setupServerAndGetUnitTestExternalRK(testingObj *testing.T, handlers *[]tstutil.Handler, kube *client.Client) (*resourcekeyExternal, *httptest.Server, error) {
	mClient, tstServer, err := tstutil.SetupTestServerClient(testingObj, handlers)
	if err != nil || mClient == nil || tstServer == nil {
		return nil, nil, err
	}

	return &resourcekeyExternal{
			kube:   *kube,
			client: *mClient,
			logger: logging.NewNopLogger(),
		},
		tstServer,
		nil
}

func TestResourceKeyObserve(t *testing.T) {
	type want struct {
		mg  resource.Managed
		obs managed.ExternalObservation
		err error
	}
	cases := map[string]struct {
		handlers []tstutil.Handler
		kube     client.Client
		args     tstutil.Args
		want     want
	}{
		"NotFound": {
			handlers: []tstutil.Handler{
				{
					Path: "/",
					HandlerFunc: func(w http.ResponseWriter, r *http.Request) {
						_ = r.Body.Close()
						if diff := cmp.Diff(http.MethodGet, r.Method); diff != "" {
							t.Errorf("r: -want, +got:\n%s", diff)
						}
						// content type should always set before writeHeader()
						w.Header().Set("Content-Type", "application/json")
						w.WriteHeader(http.StatusNotFound)
						_ = json.NewEncoder(w).Encode(&rcv2.ResourceKey{})
					},
				},
			},
			args: tstutil.Args{
				Managed: key(),
			},
			want: want{
				mg:  key(),
				err: nil,
			},
		},
		"GetFailed": {
			handlers: []tstutil.Handler{
				{
					Path: "/",
					HandlerFunc: func(w http.ResponseWriter, r *http.Request) {
						_ = r.Body.Close()
						if diff := cmp.Diff(http.MethodGet, r.Method); diff != "" {
							t.Errorf("r: -want, +got:\n%s", diff)
						}
						w.Header().Set("Content-Type", "application/json")
						w.WriteHeader(http.StatusBadRequest)
						_ = json.NewEncoder(w).Encode(&rcv2.ResourceKey{})
					},
				},
			},
			args: tstutil.Args{
				Managed: key(),
			},
			want: want{
				mg:  key(),
				err: errors.New(errGetResourceKeyFailed + ": Bad Request"),
			},
		},
		"ObservedResourceKeyUpToDate": {
			handlers: []tstutil.Handler{
				{
					Path: "/",
					HandlerFunc: func(w http.ResponseWriter, r *http.Request) {
						_ = r.Body.Close()
						if diff := cmp.Diff(http.MethodGet, r.Method); diff != "" {
							t.Errorf("r: -want, +got:\n%s", diff)
						}
						w.Header().Set("Content-Type", "application/json")
						rk := genTestSDKResourceKey()
						_ = json.NewEncoder(w).Encode(rk)
					},
				},
			},
			kube: &test.MockClient{
				MockUpdate: test.NewMockUpdateFn(nil),
			},
			args: tstutil.Args{
				Managed: key(
					rkWithExternalNameAnnotation(rkName),
					rkWithID(rkID),
					rkWithSpec(resourceKeySpec()),
				),
			},
			want: want{
				mg: genTestCRResourceKey(rkWithSpec(resourceKeySpec()), rkWithSourceCRN(sourceCrn)),
				obs: managed.ExternalObservation{
					ResourceExists:    true,
					ResourceUpToDate:  true,
					ConnectionDetails: managed.ConnectionDetails{},
				},
			},
		},
		"ObservedResourceKeyNotUpToDate": {
			handlers: []tstutil.Handler{
				{
					Path: "/",
					HandlerFunc: func(w http.ResponseWriter, r *http.Request) {
						_ = r.Body.Close()
						if diff := cmp.Diff(http.MethodGet, r.Method); diff != "" {
							t.Errorf("r: -want, +got:\n%s", diff)
						}
						w.Header().Set("Content-Type", "application/json")
						rk := genTestSDKResourceKey()
						rk.Role = &role2
						_ = json.NewEncoder(w).Encode(rk)
					},
				},
			},
			kube: &test.MockClient{
				MockUpdate: test.NewMockUpdateFn(nil),
			},
			args: tstutil.Args{
				Managed: key(
					rkWithExternalNameAnnotation(rkID),
					rkWithID(rkID),
					rkWithSpec(resourceKeySpec()),
				),
			},
			want: want{
				mg: genTestCRResourceKey(rkWithSpec(resourceKeySpec()),
					rkWithExternalNameAnnotation(rkID), rkWithSourceCRN(sourceCrn)),
				obs: managed.ExternalObservation{
					ResourceExists:    true,
					ResourceUpToDate:  false,
					ConnectionDetails: managed.ConnectionDetails{},
				},
			},
		},
		"ObservedResourceKeyRemoved": {
			handlers: []tstutil.Handler{
				{
					Path: "/",
					HandlerFunc: func(w http.ResponseWriter, r *http.Request) {
						_ = r.Body.Close()
						if diff := cmp.Diff(http.MethodGet, r.Method); diff != "" {
							t.Errorf("r: -want, +got:\n%s", diff)
						}
						w.Header().Set("Content-Type", "application/json")
						rk := genTestSDKResourceKey()
						rk.State = reference.ToPtrValue("removed")
						_ = json.NewEncoder(w).Encode(rk)
					},
				},
			},
			args: tstutil.Args{
				Managed: key(),
			},
			want: want{
				mg: key(),
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
			e, server, err := setupServerAndGetUnitTestExternalRK(t, &tc.handlers, &tc.kube)
			if err != nil {
				t.Errorf("Create(...): problem setting up the test server %s", err)
			}

			defer server.Close()

			obs, err := e.Observe(context.Background(), tc.args.Managed)
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
			if diff := cmp.Diff(tc.want.mg, tc.args.Managed); diff != "" {
				t.Errorf("Observe(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestResourceKeyCreate(t *testing.T) {
	type want struct {
		mg  resource.Managed
		cre managed.ExternalCreation
		err error
	}
	cases := map[string]struct {
		handlers []tstutil.Handler
		kube     client.Client
		args     tstutil.Args
		want     want
	}{
		"Successful": {
			handlers: []tstutil.Handler{
				{
					Path: "/",
					HandlerFunc: func(w http.ResponseWriter, r *http.Request) {
						if r.Method == http.MethodGet {
							listResourceKeysNoItems(w, r)
							return
						}
						if diff := cmp.Diff(http.MethodPost, r.Method); diff != "" {
							t.Errorf("r: -want, +got:\n%s", diff)
						}
						w.Header().Set("Content-Type", "application/json")
						w.WriteHeader(http.StatusCreated)
						_ = r.Body.Close()
						ri := genTestSDKResourceKey()
						_ = json.NewEncoder(w).Encode(ri)
					},
				},
			},
			args: tstutil.Args{
				Managed: key(rkWithSpec(resourceKeySpec())),
			},
			want: want{
				mg: key(rkWithSpec(resourceKeySpec()),
					rkWithConditions(cpv1alpha1.Creating()),
					rkWithExternalNameAnnotation(rkID)),
				cre: managed.ExternalCreation{ExternalNameAssigned: true},
				err: nil,
			},
		},
		"Failed": {
			handlers: []tstutil.Handler{
				{
					Path: "/",
					HandlerFunc: func(w http.ResponseWriter, r *http.Request) {
						if r.Method == http.MethodGet {
							listResourceKeysNoItems(w, r)
							return
						}

						if diff := cmp.Diff(http.MethodPost, r.Method); diff != "" {
							t.Errorf("r: -want, +got:\n%s", diff)
						}
						w.Header().Set("Content-Type", "application/json")
						w.WriteHeader(http.StatusBadRequest)
						_ = r.Body.Close()

						b := map[string]interface{}{
							"message":     errWrongGUID,
							"status_code": 400,
						}
						_ = json.NewEncoder(w).Encode(&b)
					},
				},
			},
			args: tstutil.Args{
				Managed: key(rkWithSpec(resourceKeySpec())),
			},
			want: want{
				mg:  key(rkWithSpec(resourceKeySpec()), rkWithConditions(cpv1alpha1.Creating())),
				err: errors.Wrap(errors.New(errWrongGUID), errCreateResourceKey),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e, server, err := setupServerAndGetUnitTestExternalRK(t, &tc.handlers, &tc.kube)
			if err != nil {
				t.Errorf("Create(...): problem setting up the test server %s", err)
			}

			defer server.Close()

			cre, err := e.Create(context.Background(), tc.args.Managed)
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
			if diff := cmp.Diff(tc.want.mg, tc.args.Managed); diff != "" {
				t.Errorf("Create(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestResourceKeyDelete(t *testing.T) {
	type want struct {
		mg  resource.Managed
		err error
	}
	cases := map[string]struct {
		handlers []tstutil.Handler
		kube     client.Client
		args     tstutil.Args
		want     want
	}{
		"Successful": {
			handlers: []tstutil.Handler{
				{
					Path: "/",
					HandlerFunc: func(w http.ResponseWriter, r *http.Request) {
						if diff := cmp.Diff(http.MethodDelete, r.Method); diff != "" {
							t.Errorf("r: -want, +got:\n%s", diff)
						}
						w.Header().Set("Content-Type", "application/json")
						w.WriteHeader(http.StatusAccepted)
						_ = r.Body.Close()
					},
				},
			},
			args: tstutil.Args{
				Managed: key(rkWithID(id)),
			},
			want: want{
				mg:  key(rkWithID(id), rkWithConditions(cpv1alpha1.Deleting())),
				err: nil,
			},
		},
		"AlreadyGone": {
			handlers: []tstutil.Handler{
				{
					Path: "/",
					HandlerFunc: func(w http.ResponseWriter, r *http.Request) {
						if diff := cmp.Diff(http.MethodDelete, r.Method); diff != "" {
							t.Errorf("r: -want, +got:\n%s", diff)
						}
						w.Header().Set("Content-Type", "application/json")
						w.WriteHeader(http.StatusNotFound)
						_ = r.Body.Close()
					},
				},
			},
			args: tstutil.Args{
				Managed: key(rkWithID(id)),
			},
			want: want{
				mg:  key(rkWithID(id), rkWithConditions(cpv1alpha1.Deleting())),
				err: nil,
			},
		},
		"Failed": {
			handlers: []tstutil.Handler{
				{
					Path: "/",
					HandlerFunc: func(w http.ResponseWriter, r *http.Request) {
						if diff := cmp.Diff(http.MethodDelete, r.Method); diff != "" {
							t.Errorf("r: -want, +got:\n%s", diff)
						}
						w.Header().Set("Content-Type", "application/json")
						w.WriteHeader(http.StatusBadRequest)
						_ = r.Body.Close()
					},
				},
			},
			args: tstutil.Args{
				Managed: key(rkWithID(id)),
			},
			want: want{
				mg:  key(rkWithID(id), rkWithConditions(cpv1alpha1.Deleting())),
				err: errors.Wrap(errors.New(http.StatusText(http.StatusBadRequest)), errDeleteResourceKey),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e, server, errSetup := setupServerAndGetUnitTestExternalRK(t, &tc.handlers, &tc.kube)
			if errSetup != nil {
				t.Errorf("Create(...): problem setting up the test server %s", errSetup)
			}

			defer server.Close()

			err := e.Delete(context.Background(), tc.args.Managed)
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
			if diff := cmp.Diff(tc.want.mg, tc.args.Managed); diff != "" {
				t.Errorf("Delete(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestResourceKeyUpdate(t *testing.T) {
	type want struct {
		mg  resource.Managed
		upd managed.ExternalUpdate
		err error
	}
	cases := map[string]struct {
		handlers []tstutil.Handler
		kube     client.Client
		args     tstutil.Args
		want     want
	}{
		"Successful": {
			handlers: []tstutil.Handler{
				{
					Path: "/",
					HandlerFunc: func(w http.ResponseWriter, r *http.Request) {
						if diff := cmp.Diff(http.MethodPatch, r.Method); diff != "" {
							t.Errorf("r: -want, +got:\n%s", diff)
						}
						w.Header().Set("Content-Type", "application/json")
						w.WriteHeader(http.StatusOK)
						_ = r.Body.Close()
						ri := genTestSDKResourceKey()
						_ = json.NewEncoder(w).Encode(ri)
					},
				},
			},
			args: tstutil.Args{
				Managed: genTestCRResourceKey(rkWithSpec(resourceKeySpec())),
			},
			want: want{
				mg:  genTestCRResourceKey(),
				upd: managed.ExternalUpdate{},
				err: nil,
			},
		},
		"PatchFails": {
			handlers: []tstutil.Handler{
				{
					Path: "/",
					HandlerFunc: func(w http.ResponseWriter, r *http.Request) {
						if diff := cmp.Diff(http.MethodPatch, r.Method); diff != "" {
							t.Errorf("r: -want, +got:\n%s", diff)
						}
						w.Header().Set("Content-Type", "application/json")
						w.WriteHeader(http.StatusBadRequest)
						_ = r.Body.Close()
					},
				},
			},
			args: tstutil.Args{
				Managed: genTestCRResourceKey(rkWithSpec(resourceKeySpec())),
			},
			want: want{
				mg:  genTestCRResourceKey(),
				err: errors.Wrap(errors.New(http.StatusText(http.StatusBadRequest)), errUpdResourceKey),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e, server, err := setupServerAndGetUnitTestExternalRK(t, &tc.handlers, &tc.kube)
			if err != nil {
				t.Errorf("Create(...): problem setting up the test server %s", err)
			}

			defer server.Close()

			upd, err := e.Update(context.Background(), tc.args.Managed)
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
				if diff := cmp.Diff(tc.want.mg, tc.args.Managed); diff != "" {
					t.Errorf("Update(...): -want, +got:\n%s", diff)
				}
				if diff := cmp.Diff(tc.want.upd, upd); diff != "" {
					t.Errorf("Update(...): -want, +got:\n%s", diff)
				}
			}
		})
	}
}
