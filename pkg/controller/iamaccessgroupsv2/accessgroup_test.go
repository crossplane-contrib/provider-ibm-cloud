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

package iamaccessgroupsv2

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-openapi/strfmt"
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
	iamagv2 "github.com/IBM/platform-services-go-sdk/iamaccessgroupsv2"

	"github.com/crossplane-contrib/provider-ibm-cloud/apis/iamaccessgroupsv2/v1alpha1"
	ibmc "github.com/crossplane-contrib/provider-ibm-cloud/pkg/clients"
	ibmcag "github.com/crossplane-contrib/provider-ibm-cloud/pkg/clients/accessgroup"
)

const (
	bearerTok       = "mock-token"
	errAgBadRequest = "error getting access group: Bad Request"
	errAgForbidden  = "error getting access group: Forbidden"
)

var (
	accountID         = "aa5a00334eaf9eb9339d2ab48f20d7ff"
	agName            = "myAccessGroup"
	agDescription     = "myAccessGroup Description"
	agDescription2    = "myAccessGroup Description 2"
	agID              = "12345678-abcd-1a2b-a1b2-1234567890ab"
	createdAt, _      = strfmt.ParseDateTime("2020-10-31T02:33:06Z")
	lastModifiedAt, _ = strfmt.ParseDateTime("2020-10-31T03:33:06Z")
	agHRef            = "https://iam.cloud.ibm.com/v2/accessgroups/12345678-abcd-1a2b-a1b2-1234567890ab"
	eTag              = "1-eb832c7ff8c8016a542974b9f880b55e"
	transactionID     = "12345-abcd-ef000-abac"
	createdByID       = "IBM-User-0001"
	isFederated       = false
)

var _ managed.ExternalConnecter = &agConnector{}
var _ managed.ExternalClient = &agExternal{}

type handler struct {
	path        string
	handlerFunc func(w http.ResponseWriter, r *http.Request)
}

type agModifier func(*v1alpha1.AccessGroup)

func ag(im ...agModifier) *v1alpha1.AccessGroup {
	i := &v1alpha1.AccessGroup{
		ObjectMeta: metav1.ObjectMeta{
			Name:       agName,
			Finalizers: []string{},
			Annotations: map[string]string{
				meta.AnnotationKeyExternalName: agID,
			},
		},
		Spec: v1alpha1.AccessGroupSpec{
			ForProvider: v1alpha1.AccessGroupParameters{},
		},
	}
	for _, m := range im {
		m(i)
	}
	return i
}

func agWithExternalNameAnnotation(externalName string) agModifier {
	return func(i *v1alpha1.AccessGroup) {
		if i.ObjectMeta.Annotations == nil {
			i.ObjectMeta.Annotations = make(map[string]string)
		}
		i.ObjectMeta.Annotations[meta.AnnotationKeyExternalName] = externalName
	}
}

func agWithEtagAnnotation(eTag string) agModifier {
	return func(i *v1alpha1.AccessGroup) {
		if i.ObjectMeta.Annotations == nil {
			i.ObjectMeta.Annotations = make(map[string]string)
		}
		i.ObjectMeta.Annotations[ibmc.ETagAnnotation] = eTag
	}
}

func agWithSpec(p v1alpha1.AccessGroupParameters) agModifier {
	return func(r *v1alpha1.AccessGroup) { r.Spec.ForProvider = p }
}

func agWithConditions(c ...cpv1alpha1.Condition) agModifier {
	return func(i *v1alpha1.AccessGroup) { i.Status.SetConditions(c...) }
}

func agWithStatus(p v1alpha1.AccessGroupObservation) agModifier {
	return func(r *v1alpha1.AccessGroup) { r.Status.AtProvider = p }
}

func agParams(m ...func(*v1alpha1.AccessGroupParameters)) *v1alpha1.AccessGroupParameters {
	p := &v1alpha1.AccessGroupParameters{
		Name:          agName,
		AccountID:     accountID,
		Description:   &agDescription,
		TransactionID: &transactionID,
	}
	for _, f := range m {
		f(p)
	}
	return p
}

func crObservation(m ...func(*v1alpha1.AccessGroupObservation)) *v1alpha1.AccessGroupObservation {
	o := &v1alpha1.AccessGroupObservation{
		ID:               agID,
		CreatedAt:        ibmc.DateTimeToMetaV1Time(&createdAt),
		LastModifiedAt:   ibmc.DateTimeToMetaV1Time(&lastModifiedAt),
		CreatedByID:      createdByID,
		LastModifiedByID: createdByID,
		Href:             agHRef,
		IsFederated:      isFederated,
	}

	for _, f := range m {
		f(o)
	}
	return o
}

func crInstance(m ...func(*iamagv2.Group)) *iamagv2.Group {
	i := &iamagv2.Group{
		ID:               &agID,
		Name:             &agName,
		AccountID:        &accountID,
		CreatedAt:        &createdAt,
		CreatedByID:      &createdByID,
		LastModifiedAt:   &lastModifiedAt,
		LastModifiedByID: &createdByID,
		Description:      &agDescription,
		Href:             &agHRef,
		IsFederated:      &isFederated,
	}

	for _, f := range m {
		f(i)
	}
	return i
}

func TestAccessGroupObserve(t *testing.T) {
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
						_ = json.NewEncoder(w).Encode(&iamagv2.Group{})
					},
				},
			},
			args: args{
				mg: ag(),
			},
			want: want{
				mg:  ag(),
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
						_ = json.NewEncoder(w).Encode(&iamagv2.Group{})
					},
				},
			},
			args: args{
				mg: ag(),
			},
			want: want{
				mg:  ag(),
				err: errors.New(errAgBadRequest),
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
						_ = json.NewEncoder(w).Encode(&iamagv2.Group{})
					},
				},
			},
			args: args{
				mg: ag(),
			},
			want: want{
				mg:  ag(),
				err: errors.New(errAgForbidden),
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
				mg: ag(
					agWithExternalNameAnnotation(agID),
					agWithSpec(*agParams()),
				),
			},
			want: want{
				mg: ag(agWithSpec(*agParams()),
					agWithConditions(cpv1alpha1.Available()),
					agWithStatus(*crObservation(func(cro *v1alpha1.AccessGroupObservation) {
						cro.State = ibmcag.StateActive
					})),
					agWithEtagAnnotation(eTag)),
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
						cr := crInstance(func(p *iamagv2.Group) {
							p.Description = &agDescription2
						})
						_ = json.NewEncoder(w).Encode(cr)
					},
				},
			},
			kube: &test.MockClient{
				MockUpdate: test.NewMockUpdateFn(nil),
			},
			args: args{
				mg: ag(
					agWithExternalNameAnnotation(agID),
					agWithSpec(*agParams()),
				),
			},
			want: want{
				mg: ag(agWithSpec(*agParams()),
					agWithEtagAnnotation(eTag),
					agWithConditions(cpv1alpha1.Available()),
					agWithStatus(*crObservation(func(cro *v1alpha1.AccessGroupObservation) {
						cro.State = ibmcag.StateActive
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
			e := agExternal{
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

func TestAccessGroupCreate(t *testing.T) {
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
				mg: ag(agWithSpec(*agParams())),
			},
			want: want{
				mg: ag(agWithSpec(*agParams()),
					agWithConditions(cpv1alpha1.Creating()),
					agWithExternalNameAnnotation(agID)),
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
				mg: ag(agWithSpec(*agParams())),
			},
			want: want{
				mg: ag(agWithSpec(*agParams()),
					agWithConditions(cpv1alpha1.Creating())),
				cre: managed.ExternalCreation{ExternalNameAssigned: false},
				err: errors.Wrap(errors.New(http.StatusText(http.StatusBadRequest)), errCreateAccessGroup),
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
				mg: ag(agWithSpec(*agParams())),
			},
			want: want{
				mg: ag(agWithSpec(*agParams()),
					agWithConditions(cpv1alpha1.Creating())),
				cre: managed.ExternalCreation{ExternalNameAssigned: false},
				err: errors.Wrap(errors.New(http.StatusText(http.StatusConflict)), errCreateAccessGroup),
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
				mg: ag(agWithSpec(*agParams())),
			},
			want: want{
				mg: ag(agWithSpec(*agParams()),
					agWithConditions(cpv1alpha1.Creating())),
				cre: managed.ExternalCreation{ExternalNameAssigned: false},
				err: errors.Wrap(errors.New(http.StatusText(http.StatusForbidden)), errCreateAccessGroup),
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
			e := agExternal{
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

func TestAccessGroupDelete(t *testing.T) {
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
				mg: ag(agWithStatus(*crObservation())),
			},
			want: want{
				mg:  ag(agWithStatus(*crObservation()), agWithConditions(cpv1alpha1.Deleting())),
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
				mg: ag(agWithStatus(*crObservation())),
			},
			want: want{
				mg:  ag(agWithStatus(*crObservation()), agWithConditions(cpv1alpha1.Deleting())),
				err: errors.Wrap(errors.New(http.StatusText(http.StatusBadRequest)), errDeleteAccessGroup),
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
				mg: ag(agWithStatus(*crObservation())),
			},
			want: want{
				mg:  ag(agWithStatus(*crObservation()), agWithConditions(cpv1alpha1.Deleting())),
				err: errors.Wrap(errors.New(http.StatusText(http.StatusUnauthorized)), errDeleteAccessGroup),
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
				mg: ag(agWithStatus(*crObservation())),
			},
			want: want{
				mg:  ag(agWithStatus(*crObservation()), agWithConditions(cpv1alpha1.Deleting())),
				err: errors.Wrap(errors.New(http.StatusText(http.StatusForbidden)), errDeleteAccessGroup),
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
			e := agExternal{
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

func TestAccessGroupUpdate(t *testing.T) {
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
						if diff := cmp.Diff(http.MethodPatch, r.Method); diff != "" {
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
				mg: ag(agWithSpec(*agParams()), agWithStatus(*crObservation()), agWithEtagAnnotation(eTag)),
			},
			want: want{
				mg:  ag(agWithSpec(*agParams()), agWithStatus(*crObservation()), agWithEtagAnnotation(eTag)),
				upd: managed.ExternalUpdate{},
				err: nil,
			},
		},
		"BadRequest": {
			handlers: []handler{
				{
					path: "/",
					handlerFunc: func(w http.ResponseWriter, r *http.Request) {
						if diff := cmp.Diff(http.MethodPatch, r.Method); diff != "" {
							t.Errorf("r: -want, +got:\n%s", diff)
						}
						w.Header().Set("Content-Type", "application/json")
						w.WriteHeader(http.StatusBadRequest)
						_ = r.Body.Close()
					},
				},
			},
			args: args{
				mg: ag(agWithSpec(*agParams()), agWithStatus(*crObservation())),
			},
			want: want{
				mg:  ag(agWithSpec(*agParams()), agWithStatus(*crObservation())),
				err: errors.Wrap(errors.New(http.StatusText(http.StatusBadRequest)), errUpdAccessGroup),
			},
		},
		"NotFound": {
			handlers: []handler{
				{
					path: "/",
					handlerFunc: func(w http.ResponseWriter, r *http.Request) {
						if diff := cmp.Diff(http.MethodPatch, r.Method); diff != "" {
							t.Errorf("r: -want, +got:\n%s", diff)
						}
						w.Header().Set("Content-Type", "application/json")
						w.WriteHeader(http.StatusNotFound)
						_ = r.Body.Close()
					},
				},
			},
			args: args{
				mg: ag(agWithSpec(*agParams()), agWithStatus(*crObservation())),
			},
			want: want{
				mg:  ag(agWithSpec(*agParams()), agWithStatus(*crObservation())),
				err: errors.Wrap(errors.New(http.StatusText(http.StatusNotFound)), errUpdAccessGroup),
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
			e := agExternal{
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
