/*
Copyright 2020 The Crossplane Authors.

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

package iampolicymanagementv1

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"

	cpv1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/crossplane/crossplane-runtime/pkg/test"

	iampmv1 "github.com/IBM/platform-services-go-sdk/iampolicymanagementv1"

	"github.com/crossplane-contrib/provider-ibm-cloud/apis/iampolicymanagementv1/v1alpha1"
	ibmc "github.com/crossplane-contrib/provider-ibm-cloud/pkg/clients"
	ibmccr "github.com/crossplane-contrib/provider-ibm-cloud/pkg/clients/customrole"
	"github.com/crossplane-contrib/provider-ibm-cloud/pkg/controller/tstutil"
)

const (
	errCrBadRequest = "error getting role: Bad Request"
	errCrForbidden  = "error getting role: Forbidden"
)

var (
	roleName         = "myCustomRole"
	croleDescription = "role for my service"
	accountID        = "aa5a00334eaf9eb9339d2ab48f20d7ff"
	croleDisplayName = "MyCustomRole"
	serviceName      = "mypostgres"
	action1          = "iam.policy.create"
	action2          = "iam.policy.update"
	cRoleID          = "12345678-abcd-1a2b-a1b2-1234567890ab"
	crHRef           = "https://iam.cloud.ibm.com/v1/roles/12345678-abcd-1a2b-a1b2-1234567890ab"
	crCrn            = "crn:v1:bluemix:public:iam::::role:" + roleName
)

var _ managed.ExternalConnecter = &pConnector{}
var _ managed.ExternalClient = &pExternal{}

type crModifier func(*v1alpha1.CustomRole)

func cr(im ...crModifier) *v1alpha1.CustomRole {
	i := &v1alpha1.CustomRole{
		ObjectMeta: metav1.ObjectMeta{
			Name:       roleName,
			Finalizers: []string{},
			Annotations: map[string]string{
				meta.AnnotationKeyExternalName: cRoleID,
			},
		},
		Spec: v1alpha1.CustomRoleSpec{
			ForProvider: v1alpha1.CustomRoleParameters{},
		},
	}
	for _, m := range im {
		m(i)
	}
	return i
}

func crWithExternalNameAnnotation(externalName string) crModifier {
	return func(i *v1alpha1.CustomRole) {
		if i.ObjectMeta.Annotations == nil {
			i.ObjectMeta.Annotations = make(map[string]string)
		}
		i.ObjectMeta.Annotations[meta.AnnotationKeyExternalName] = externalName
	}
}

func crWithEtagAnnotation(eTag string) crModifier {
	return func(i *v1alpha1.CustomRole) {
		if i.ObjectMeta.Annotations == nil {
			i.ObjectMeta.Annotations = make(map[string]string)
		}
		i.ObjectMeta.Annotations[ibmc.ETagAnnotation] = eTag
	}
}

func crWithSpec(p v1alpha1.CustomRoleParameters) crModifier {
	return func(r *v1alpha1.CustomRole) { r.Spec.ForProvider = p }
}

func crWithConditions(c ...cpv1alpha1.Condition) crModifier {
	return func(i *v1alpha1.CustomRole) { i.Status.SetConditions(c...) }
}

func crWithStatus(p v1alpha1.CustomRoleObservation) crModifier {
	return func(r *v1alpha1.CustomRole) { r.Status.AtProvider = p }
}

func crParams(m ...func(*v1alpha1.CustomRoleParameters)) *v1alpha1.CustomRoleParameters {
	p := &v1alpha1.CustomRoleParameters{
		DisplayName: croleDisplayName,
		Actions:     []string{action1, action2},
		Name:        roleName,
		AccountID:   accountID,
		ServiceName: serviceName,
		Description: &croleDescription,
	}
	for _, f := range m {
		f(p)
	}
	return p
}

func crObservation(m ...func(*v1alpha1.CustomRoleObservation)) *v1alpha1.CustomRoleObservation {
	o := &v1alpha1.CustomRoleObservation{
		ID:               cRoleID,
		CRN:              crCrn,
		CreatedAt:        ibmc.DateTimeToMetaV1Time(&createdAt),
		LastModifiedAt:   ibmc.DateTimeToMetaV1Time(&lastModifiedAt),
		CreatedByID:      createdByID,
		LastModifiedByID: createdByID,
		Href:             crHRef,
	}

	for _, f := range m {
		f(o)
	}
	return o
}

func crInstance(m ...func(*iampmv1.CustomRole)) *iampmv1.CustomRole {
	i := &iampmv1.CustomRole{
		ID:               &cRoleID,
		Name:             &roleName,
		AccountID:        &accountID,
		ServiceName:      &serviceName,
		CreatedAt:        &createdAt,
		CreatedByID:      &createdByID,
		LastModifiedAt:   &lastModifiedAt,
		LastModifiedByID: &createdByID,
		DisplayName:      &croleDisplayName,
		Description:      &croleDescription,
		Actions:          []string{action1, action2},
		CRN:              &crCrn,
		Href:             &crHRef,
	}

	for _, f := range m {
		f(i)
	}
	return i
}

// Sets up a unit test http server, and creates an external client structure appropriate for unit test.
//
// Params
//
//	testingObj - the test object
//	handlers - the handlers that create the responses
//	client - the controller runtime client
//
// Returns
//   - the external bucket config, ready for unit test
//   - the test http server, on which the caller should call 'defer ....Close()' (reason for this is we need to keep it around to prevent
//     garbage collection)
//   - an error (iff...)
func setupServerAndGetUnitTestExternalCR(testingObj *testing.T, handlers *[]tstutil.Handler, kube *client.Client) (*crExternal, *httptest.Server, error) {
	mClient, tstServer, err := tstutil.SetupTestServerClient(testingObj, handlers)
	if err != nil {
		return nil, nil, err
	}

	return &crExternal{
			kube:   *kube,
			client: *mClient,
			logger: logging.NewNopLogger(),
		},
		tstServer,
		nil
}

func TestCustomRoleObserve(t *testing.T) {
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
						err := json.NewEncoder(w).Encode(&iampmv1.CustomRole{})
						if err != nil {
							klog.Errorf("%s", err)
						}
					},
				},
			},
			args: tstutil.Args{
				Managed: cr(),
			},
			want: want{
				mg:  cr(),
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
						err := json.NewEncoder(w).Encode(&iampmv1.CustomRole{})
						if err != nil {
							klog.Errorf("%s", err)
						}
					},
				},
			},
			args: tstutil.Args{
				Managed: cr(),
			},
			want: want{
				mg:  cr(),
				err: errors.New(errCrBadRequest),
			},
		},

		"GetForbidden": {
			handlers: []tstutil.Handler{
				{
					Path: "/",
					HandlerFunc: func(w http.ResponseWriter, r *http.Request) {
						_ = r.Body.Close()
						if diff := cmp.Diff(http.MethodGet, r.Method); diff != "" {
							t.Errorf("r: -want, +got:\n%s", diff)
						}
						w.Header().Set("Content-Type", "application/json")
						w.WriteHeader(http.StatusForbidden)
						err := json.NewEncoder(w).Encode(&iampmv1.CustomRole{})
						if err != nil {
							klog.Errorf("%s", err)
						}
					},
				},
			},
			args: tstutil.Args{
				Managed: cr(),
			},
			want: want{
				mg:  cr(),
				err: errors.New(errCrForbidden),
			},
		},
		"UpToDate": {
			handlers: []tstutil.Handler{
				{
					Path: "/",
					HandlerFunc: func(w http.ResponseWriter, r *http.Request) {
						_ = r.Body.Close()
						if diff := cmp.Diff(http.MethodGet, r.Method); diff != "" {
							t.Errorf("r: -want, +got:\n%s", diff)
						}
						w.Header().Set("Content-Type", "application/json")
						w.Header().Set("ETag", eTag)
						cr := crInstance()
						err := json.NewEncoder(w).Encode(cr)
						if err != nil {
							klog.Errorf("%s", err)
						}
					},
				},
			},
			kube: &test.MockClient{
				MockUpdate: test.NewMockUpdateFn(nil),
			},
			args: tstutil.Args{
				Managed: cr(
					crWithExternalNameAnnotation(policyID),
					crWithSpec(*crParams()),
				),
			},
			want: want{
				mg: cr(crWithSpec(*crParams()),
					crWithConditions(cpv1alpha1.Available()),
					crWithStatus(*crObservation(func(cro *v1alpha1.CustomRoleObservation) {
						cro.State = ibmccr.StateActive
					})),
					crWithEtagAnnotation(eTag)),
				obs: managed.ExternalObservation{
					ResourceExists:    true,
					ResourceUpToDate:  true,
					ConnectionDetails: nil,
				},
			},
		},
		"NotUpToDate": {
			handlers: []tstutil.Handler{
				{
					Path: "/",
					HandlerFunc: func(w http.ResponseWriter, r *http.Request) {
						_ = r.Body.Close()
						if diff := cmp.Diff(http.MethodGet, r.Method); diff != "" {
							t.Errorf("r: -want, +got:\n%s", diff)
						}
						w.Header().Set("Content-Type", "application/json")
						w.Header().Set("ETag", eTag)
						cr := crInstance(func(p *iampmv1.CustomRole) {
							p.Actions = []string{action1}
						})
						err := json.NewEncoder(w).Encode(cr)
						if err != nil {
							klog.Errorf("%s", err)
						}
					},
				},
			},
			kube: &test.MockClient{
				MockUpdate: test.NewMockUpdateFn(nil),
			},
			args: tstutil.Args{
				Managed: cr(
					crWithExternalNameAnnotation(policyID),
					crWithSpec(*crParams()),
				),
			},
			want: want{
				mg: cr(crWithSpec(*crParams()),
					crWithEtagAnnotation(eTag),
					crWithConditions(cpv1alpha1.Available()),
					crWithStatus(*crObservation(func(cro *v1alpha1.CustomRoleObservation) {
						cro.State = ibmccr.StateActive
					}))),
				obs: managed.ExternalObservation{
					ResourceExists:    true,
					ResourceUpToDate:  false,
					ConnectionDetails: nil,
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e, server, errCr := setupServerAndGetUnitTestExternalCR(t, &tc.handlers, &tc.kube)
			if errCr != nil {
				t.Errorf("Delete(...): problem setting up the test server %s", errCr)
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

func TestCustomRoleCreate(t *testing.T) {
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
						if diff := cmp.Diff(http.MethodPost, r.Method); diff != "" {
							t.Errorf("r: -want, +got:\n%s", diff)
						}
						w.Header().Set("Content-Type", "application/json")
						w.WriteHeader(http.StatusCreated)
						_ = r.Body.Close()
						cr := crInstance()
						err := json.NewEncoder(w).Encode(cr)
						if err != nil {
							klog.Errorf("%s", err)
						}
					},
				},
			},
			args: tstutil.Args{
				Managed: cr(crWithSpec(*crParams())),
			},
			want: want{
				mg: cr(crWithSpec(*crParams()),
					crWithConditions(cpv1alpha1.Creating()),
					crWithExternalNameAnnotation(policyID)),
				cre: managed.ExternalCreation{ExternalNameAssigned: true},
				err: nil,
			},
		},
		"BadRequest": {
			handlers: []tstutil.Handler{
				{
					Path: "/",
					HandlerFunc: func(w http.ResponseWriter, r *http.Request) {
						if diff := cmp.Diff(http.MethodPost, r.Method); diff != "" {
							t.Errorf("r: -want, +got:\n%s", diff)
						}
						w.Header().Set("Content-Type", "application/json")
						w.WriteHeader(http.StatusBadRequest)
						_ = r.Body.Close()
						cr := crInstance()
						err := json.NewEncoder(w).Encode(cr)
						if err != nil {
							klog.Errorf("%s", err)
						}
					},
				},
			},
			args: tstutil.Args{
				Managed: cr(crWithSpec(*crParams())),
			},
			want: want{
				mg: cr(crWithSpec(*crParams()),
					crWithConditions(cpv1alpha1.Creating())),
				cre: managed.ExternalCreation{ExternalNameAssigned: false},
				err: errors.Wrap(errors.New(http.StatusText(http.StatusBadRequest)), errCreateCustomRole),
			},
		},
		"Conflict": {
			handlers: []tstutil.Handler{
				{
					Path: "/",
					HandlerFunc: func(w http.ResponseWriter, r *http.Request) {
						if diff := cmp.Diff(http.MethodPost, r.Method); diff != "" {
							t.Errorf("r: -want, +got:\n%s", diff)
						}
						w.Header().Set("Content-Type", "application/json")
						w.WriteHeader(http.StatusConflict)
						_ = r.Body.Close()
						cr := crInstance()
						err := json.NewEncoder(w).Encode(cr)
						if err != nil {
							klog.Errorf("%s", err)
						}
					},
				},
			},
			args: tstutil.Args{
				Managed: cr(crWithSpec(*crParams())),
			},
			want: want{
				mg: cr(crWithSpec(*crParams()),
					crWithConditions(cpv1alpha1.Creating())),
				cre: managed.ExternalCreation{ExternalNameAssigned: false},
				err: errors.Wrap(errors.New(http.StatusText(http.StatusConflict)), errCreateCustomRole),
			},
		},
		"Forbidden": {
			handlers: []tstutil.Handler{
				{
					Path: "/",
					HandlerFunc: func(w http.ResponseWriter, r *http.Request) {
						if diff := cmp.Diff(http.MethodPost, r.Method); diff != "" {
							t.Errorf("r: -want, +got:\n%s", diff)
						}
						w.Header().Set("Content-Type", "application/json")
						w.WriteHeader(http.StatusForbidden)
						_ = r.Body.Close()
						cr := crInstance()
						err := json.NewEncoder(w).Encode(cr)
						if err != nil {
							klog.Errorf("%s", err)
						}
					},
				},
			},
			args: tstutil.Args{
				Managed: cr(crWithSpec(*crParams())),
			},
			want: want{
				mg: cr(crWithSpec(*crParams()),
					crWithConditions(cpv1alpha1.Creating())),
				cre: managed.ExternalCreation{ExternalNameAssigned: false},
				err: errors.Wrap(errors.New(http.StatusText(http.StatusForbidden)), errCreateCustomRole),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e, server, errCr := setupServerAndGetUnitTestExternalCR(t, &tc.handlers, &tc.kube)
			if errCr != nil {
				t.Errorf("Delete(...): problem setting up the test server %s", errCr)
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

func TestCustomRoleDelete(t *testing.T) {
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
						w.WriteHeader(http.StatusNoContent)
						_ = r.Body.Close()
					},
				},
			},
			args: tstutil.Args{
				Managed: cr(crWithStatus(*crObservation())),
			},
			want: want{
				mg:  cr(crWithStatus(*crObservation()), crWithConditions(cpv1alpha1.Deleting())),
				err: nil,
			},
		},
		"BadRequest": {
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
				Managed: cr(crWithStatus(*crObservation())),
			},
			want: want{
				mg:  cr(crWithStatus(*crObservation()), crWithConditions(cpv1alpha1.Deleting())),
				err: errors.Wrap(errors.New(http.StatusText(http.StatusBadRequest)), errDeleteCustomRole),
			},
		},
		"InvalidToken": {
			handlers: []tstutil.Handler{
				{
					Path: "/",
					HandlerFunc: func(w http.ResponseWriter, r *http.Request) {
						if diff := cmp.Diff(http.MethodDelete, r.Method); diff != "" {
							t.Errorf("r: -want, +got:\n%s", diff)
						}
						w.Header().Set("Content-Type", "application/json")
						w.WriteHeader(http.StatusUnauthorized)
						_ = r.Body.Close()
					},
				},
			},
			args: tstutil.Args{
				Managed: cr(crWithStatus(*crObservation())),
			},
			want: want{
				mg:  cr(crWithStatus(*crObservation()), crWithConditions(cpv1alpha1.Deleting())),
				err: errors.Wrap(errors.New(http.StatusText(http.StatusUnauthorized)), errDeleteCustomRole),
			},
		},
		"Forbidden": {
			handlers: []tstutil.Handler{
				{
					Path: "/",
					HandlerFunc: func(w http.ResponseWriter, r *http.Request) {
						if diff := cmp.Diff(http.MethodDelete, r.Method); diff != "" {
							t.Errorf("r: -want, +got:\n%s", diff)
						}
						w.Header().Set("Content-Type", "application/json")
						w.WriteHeader(http.StatusForbidden)
						_ = r.Body.Close()
					},
				},
			},
			args: tstutil.Args{
				Managed: cr(crWithStatus(*crObservation())),
			},
			want: want{
				mg:  cr(crWithStatus(*crObservation()), crWithConditions(cpv1alpha1.Deleting())),
				err: errors.Wrap(errors.New(http.StatusText(http.StatusForbidden)), errDeleteCustomRole),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e, server, errCr := setupServerAndGetUnitTestExternalCR(t, &tc.handlers, &tc.kube)
			if errCr != nil {
				t.Errorf("Delete(...): problem setting up the test server %s", errCr)
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

func TestCustomRoleUpdate(t *testing.T) {
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
						if diff := cmp.Diff(http.MethodPut, r.Method); diff != "" {
							t.Errorf("r: -want, +got:\n%s", diff)
						}
						w.Header().Set("Content-Type", "application/json")
						w.WriteHeader(http.StatusOK)
						_ = r.Body.Close()
						cr := crInstance()
						err := json.NewEncoder(w).Encode(cr)
						if err != nil {
							klog.Errorf("%s", err)
						}
					},
				},
			},
			args: tstutil.Args{
				Managed: cr(crWithSpec(*crParams()), crWithStatus(*crObservation()), crWithEtagAnnotation(eTag)),
			},
			want: want{
				mg:  cr(crWithSpec(*crParams()), crWithStatus(*crObservation()), crWithEtagAnnotation(eTag)),
				upd: managed.ExternalUpdate{},
				err: nil,
			},
		},
		"BadRequest": {
			handlers: []tstutil.Handler{
				{
					Path: "/",
					HandlerFunc: func(w http.ResponseWriter, r *http.Request) {
						if diff := cmp.Diff(http.MethodPut, r.Method); diff != "" {
							t.Errorf("r: -want, +got:\n%s", diff)
						}
						w.Header().Set("Content-Type", "application/json")
						w.WriteHeader(http.StatusBadRequest)
						_ = r.Body.Close()
					},
				},
			},
			args: tstutil.Args{
				Managed: cr(crWithSpec(*crParams()), crWithStatus(*crObservation())),
			},
			want: want{
				mg:  cr(crWithSpec(*crParams()), crWithStatus(*crObservation())),
				err: errors.Wrap(errors.New(http.StatusText(http.StatusBadRequest)), errUpdCustomRole),
			},
		},
		"NotFound": {
			handlers: []tstutil.Handler{
				{
					Path: "/",
					HandlerFunc: func(w http.ResponseWriter, r *http.Request) {
						if diff := cmp.Diff(http.MethodPut, r.Method); diff != "" {
							t.Errorf("r: -want, +got:\n%s", diff)
						}
						w.Header().Set("Content-Type", "application/json")
						w.WriteHeader(http.StatusNotFound)
						_ = r.Body.Close()
					},
				},
			},
			args: tstutil.Args{
				Managed: cr(crWithSpec(*crParams()), crWithStatus(*crObservation())),
			},
			want: want{
				mg:  cr(crWithSpec(*crParams()), crWithStatus(*crObservation())),
				err: errors.Wrap(errors.New(http.StatusText(http.StatusNotFound)), errUpdCustomRole),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e, server, errCr := setupServerAndGetUnitTestExternalCR(t, &tc.handlers, &tc.kube)
			if errCr != nil {
				t.Errorf("Delete(...): problem setting up the test server %s", errCr)
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
