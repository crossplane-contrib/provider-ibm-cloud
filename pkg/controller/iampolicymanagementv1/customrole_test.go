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

	cpv1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/crossplane/crossplane-runtime/pkg/test"

	"github.com/IBM/go-sdk-core/core"
	iampmv1 "github.com/IBM/platform-services-go-sdk/iampolicymanagementv1"

	"github.com/crossplane-contrib/provider-ibm-cloud/apis/iampolicymanagementv1/v1alpha1"
	ibmc "github.com/crossplane-contrib/provider-ibm-cloud/pkg/clients"
	ibmccr "github.com/crossplane-contrib/provider-ibm-cloud/pkg/clients/customrole"
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

func TestCustomRoleObserve(t *testing.T) {
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
						_ = json.NewEncoder(w).Encode(&iampmv1.CustomRole{})
					},
				},
			},
			args: args{
				mg: cr(),
			},
			want: want{
				mg:  cr(),
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
						_ = json.NewEncoder(w).Encode(&iampmv1.CustomRole{})
					},
				},
			},
			args: args{
				mg: cr(),
			},
			want: want{
				mg:  cr(),
				err: errors.New(errCrBadRequest),
			},
		},

		"GetForbidden": {
			handlers: []handler{
				{
					path: "/",
					handlerFunc: func(w http.ResponseWriter, r *http.Request) {
						_ = r.Body.Close()
						if diff := cmp.Diff(http.MethodGet, r.Method); diff != "" {
							t.Errorf("r: -want, +got:\n%s", diff)
						}
						w.Header().Set("Content-Type", "application/json")
						w.WriteHeader(http.StatusForbidden)
						_ = json.NewEncoder(w).Encode(&iampmv1.CustomRole{})
					},
				},
			},
			args: args{
				mg: cr(),
			},
			want: want{
				mg:  cr(),
				err: errors.New(errCrForbidden),
			},
		},
		"UpToDate": {
			handlers: []handler{
				{
					path: "/",
					handlerFunc: func(w http.ResponseWriter, r *http.Request) {
						_ = r.Body.Close()
						if diff := cmp.Diff(http.MethodGet, r.Method); diff != "" {
							t.Errorf("r: -want, +got:\n%s", diff)
						}
						w.Header().Set("Content-Type", "application/json")
						w.Header().Set("ETag", eTag)
						cr := crInstance()
						_ = json.NewEncoder(w).Encode(cr)
					},
				},
			},
			kube: &test.MockClient{
				MockUpdate: test.NewMockUpdateFn(nil),
			},
			args: args{
				mg: cr(
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
			handlers: []handler{
				{
					path: "/",
					handlerFunc: func(w http.ResponseWriter, r *http.Request) {
						_ = r.Body.Close()
						if diff := cmp.Diff(http.MethodGet, r.Method); diff != "" {
							t.Errorf("r: -want, +got:\n%s", diff)
						}
						w.Header().Set("Content-Type", "application/json")
						w.Header().Set("ETag", eTag)
						cr := crInstance(func(p *iampmv1.CustomRole) {
							p.Actions = []string{action1}
						})
						_ = json.NewEncoder(w).Encode(cr)
					},
				},
			},
			kube: &test.MockClient{
				MockUpdate: test.NewMockUpdateFn(nil),
			},
			args: args{
				mg: cr(
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
			e := crExternal{
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

func TestCustomRoleCreate(t *testing.T) {
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
					path: "/",
					handlerFunc: func(w http.ResponseWriter, r *http.Request) {
						if diff := cmp.Diff(http.MethodPost, r.Method); diff != "" {
							t.Errorf("r: -want, +got:\n%s", diff)
						}
						w.Header().Set("Content-Type", "application/json")
						w.WriteHeader(http.StatusCreated)
						_ = r.Body.Close()
						cr := crInstance()
						_ = json.NewEncoder(w).Encode(cr)
					},
				},
			},
			args: args{
				mg: cr(crWithSpec(*crParams())),
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
			handlers: []handler{
				{
					path: "/",
					handlerFunc: func(w http.ResponseWriter, r *http.Request) {
						if diff := cmp.Diff(http.MethodPost, r.Method); diff != "" {
							t.Errorf("r: -want, +got:\n%s", diff)
						}
						w.Header().Set("Content-Type", "application/json")
						w.WriteHeader(http.StatusBadRequest)
						_ = r.Body.Close()
						cr := crInstance()
						_ = json.NewEncoder(w).Encode(cr)
					},
				},
			},
			args: args{
				mg: cr(crWithSpec(*crParams())),
			},
			want: want{
				mg: cr(crWithSpec(*crParams()),
					crWithConditions(cpv1alpha1.Creating())),
				cre: managed.ExternalCreation{ExternalNameAssigned: false},
				err: errors.Wrap(errors.New(http.StatusText(http.StatusBadRequest)), errCreateCustomRole),
			},
		},
		"Conflict": {
			handlers: []handler{
				{
					path: "/",
					handlerFunc: func(w http.ResponseWriter, r *http.Request) {
						if diff := cmp.Diff(http.MethodPost, r.Method); diff != "" {
							t.Errorf("r: -want, +got:\n%s", diff)
						}
						w.Header().Set("Content-Type", "application/json")
						w.WriteHeader(http.StatusConflict)
						_ = r.Body.Close()
						cr := crInstance()
						_ = json.NewEncoder(w).Encode(cr)
					},
				},
			},
			args: args{
				mg: cr(crWithSpec(*crParams())),
			},
			want: want{
				mg: cr(crWithSpec(*crParams()),
					crWithConditions(cpv1alpha1.Creating())),
				cre: managed.ExternalCreation{ExternalNameAssigned: false},
				err: errors.Wrap(errors.New(http.StatusText(http.StatusConflict)), errCreateCustomRole),
			},
		},
		"Forbidden": {
			handlers: []handler{
				{
					path: "/",
					handlerFunc: func(w http.ResponseWriter, r *http.Request) {
						if diff := cmp.Diff(http.MethodPost, r.Method); diff != "" {
							t.Errorf("r: -want, +got:\n%s", diff)
						}
						w.Header().Set("Content-Type", "application/json")
						w.WriteHeader(http.StatusForbidden)
						_ = r.Body.Close()
						cr := crInstance()
						_ = json.NewEncoder(w).Encode(cr)
					},
				},
			},
			args: args{
				mg: cr(crWithSpec(*crParams())),
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
			e := crExternal{
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

func TestCustomRoleDelete(t *testing.T) {
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
					path: "/",
					handlerFunc: func(w http.ResponseWriter, r *http.Request) {
						if diff := cmp.Diff(http.MethodDelete, r.Method); diff != "" {
							t.Errorf("r: -want, +got:\n%s", diff)
						}
						w.Header().Set("Content-Type", "application/json")
						w.WriteHeader(http.StatusNoContent)
						_ = r.Body.Close()
					},
				},
			},
			args: args{
				mg: cr(crWithStatus(*crObservation())),
			},
			want: want{
				mg:  cr(crWithStatus(*crObservation()), crWithConditions(cpv1alpha1.Deleting())),
				err: nil,
			},
		},
		"BadRequest": {
			handlers: []handler{
				{
					path: "/",
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
				mg: cr(crWithStatus(*crObservation())),
			},
			want: want{
				mg:  cr(crWithStatus(*crObservation()), crWithConditions(cpv1alpha1.Deleting())),
				err: errors.Wrap(errors.New(http.StatusText(http.StatusBadRequest)), errDeleteCustomRole),
			},
		},
		"InvalidToken": {
			handlers: []handler{
				{
					path: "/",
					handlerFunc: func(w http.ResponseWriter, r *http.Request) {
						if diff := cmp.Diff(http.MethodDelete, r.Method); diff != "" {
							t.Errorf("r: -want, +got:\n%s", diff)
						}
						w.Header().Set("Content-Type", "application/json")
						w.WriteHeader(http.StatusUnauthorized)
						_ = r.Body.Close()
					},
				},
			},
			args: args{
				mg: cr(crWithStatus(*crObservation())),
			},
			want: want{
				mg:  cr(crWithStatus(*crObservation()), crWithConditions(cpv1alpha1.Deleting())),
				err: errors.Wrap(errors.New(http.StatusText(http.StatusUnauthorized)), errDeleteCustomRole),
			},
		},
		"Forbidden": {
			handlers: []handler{
				{
					path: "/",
					handlerFunc: func(w http.ResponseWriter, r *http.Request) {
						if diff := cmp.Diff(http.MethodDelete, r.Method); diff != "" {
							t.Errorf("r: -want, +got:\n%s", diff)
						}
						w.Header().Set("Content-Type", "application/json")
						w.WriteHeader(http.StatusForbidden)
						_ = r.Body.Close()
					},
				},
			},
			args: args{
				mg: cr(crWithStatus(*crObservation())),
			},
			want: want{
				mg:  cr(crWithStatus(*crObservation()), crWithConditions(cpv1alpha1.Deleting())),
				err: errors.Wrap(errors.New(http.StatusText(http.StatusForbidden)), errDeleteCustomRole),
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
			e := crExternal{
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

func TestCustomRoleUpdate(t *testing.T) {
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
					path: "/",
					handlerFunc: func(w http.ResponseWriter, r *http.Request) {
						if diff := cmp.Diff(http.MethodPut, r.Method); diff != "" {
							t.Errorf("r: -want, +got:\n%s", diff)
						}
						w.Header().Set("Content-Type", "application/json")
						w.WriteHeader(http.StatusOK)
						_ = r.Body.Close()
						cr := crInstance()
						_ = json.NewEncoder(w).Encode(cr)
					},
				},
			},
			args: args{
				mg: cr(crWithSpec(*crParams()), crWithStatus(*crObservation()), crWithEtagAnnotation(eTag)),
			},
			want: want{
				mg:  cr(crWithSpec(*crParams()), crWithStatus(*crObservation()), crWithEtagAnnotation(eTag)),
				upd: managed.ExternalUpdate{},
				err: nil,
			},
		},
		"BadRequest": {
			handlers: []handler{
				{
					path: "/",
					handlerFunc: func(w http.ResponseWriter, r *http.Request) {
						if diff := cmp.Diff(http.MethodPut, r.Method); diff != "" {
							t.Errorf("r: -want, +got:\n%s", diff)
						}
						w.Header().Set("Content-Type", "application/json")
						w.WriteHeader(http.StatusBadRequest)
						_ = r.Body.Close()
					},
				},
			},
			args: args{
				mg: cr(crWithSpec(*crParams()), crWithStatus(*crObservation())),
			},
			want: want{
				mg:  cr(crWithSpec(*crParams()), crWithStatus(*crObservation())),
				err: errors.Wrap(errors.New(http.StatusText(http.StatusBadRequest)), errUpdCustomRole),
			},
		},
		"NotFound": {
			handlers: []handler{
				{
					path: "/",
					handlerFunc: func(w http.ResponseWriter, r *http.Request) {
						if diff := cmp.Diff(http.MethodPut, r.Method); diff != "" {
							t.Errorf("r: -want, +got:\n%s", diff)
						}
						w.Header().Set("Content-Type", "application/json")
						w.WriteHeader(http.StatusNotFound)
						_ = r.Body.Close()
					},
				},
			},
			args: args{
				mg: cr(crWithSpec(*crParams()), crWithStatus(*crObservation())),
			},
			want: want{
				mg:  cr(crWithSpec(*crParams()), crWithStatus(*crObservation())),
				err: errors.Wrap(errors.New(http.StatusText(http.StatusNotFound)), errUpdCustomRole),
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
			e := crExternal{
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
