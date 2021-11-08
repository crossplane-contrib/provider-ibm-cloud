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

	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	cpv1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/reference"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/crossplane/crossplane-runtime/pkg/test"

	"github.com/IBM/go-sdk-core/core"
	iamagv2 "github.com/IBM/platform-services-go-sdk/iamaccessgroupsv2"

	"github.com/crossplane-contrib/provider-ibm-cloud/apis/iamaccessgroupsv2/v1alpha1"
	ibmc "github.com/crossplane-contrib/provider-ibm-cloud/pkg/clients"
	ibmcagr "github.com/crossplane-contrib/provider-ibm-cloud/pkg/clients/accessgrouprule"
)

const (
	errAgrBadRequest = "error getting access group rule: Bad Request"
	errAgrForbidden  = "error getting access group rule: Forbidden"
)

var (
	ruleID     = "abcd-12345689-1a2b-a1b2-123456789000"
	ruleName2  = "Manager group rule2"
	realmName  = "https://idp.example.org/SAML2"
	expiration = 24
	claim1     = "isManager"
	opEquals   = "EQUALS"
)

var _ managed.ExternalConnecter = &agrConnector{}
var _ managed.ExternalClient = &agrExternal{}

type agrModifier func(*v1alpha1.AccessGroupRule)

func agr(im ...agrModifier) *v1alpha1.AccessGroupRule {
	i := &v1alpha1.AccessGroupRule{
		ObjectMeta: metav1.ObjectMeta{
			Name:       agName,
			Finalizers: []string{},
			Annotations: map[string]string{
				meta.AnnotationKeyExternalName: ruleID,
			},
		},
		Spec: v1alpha1.AccessGroupRuleSpec{
			ForProvider: v1alpha1.AccessGroupRuleParameters{},
		},
	}
	for _, m := range im {
		m(i)
	}
	return i
}

func agrWithExternalNameAnnotation(externalName string) agrModifier {
	return func(i *v1alpha1.AccessGroupRule) {
		if i.ObjectMeta.Annotations == nil {
			i.ObjectMeta.Annotations = make(map[string]string)
		}
		i.ObjectMeta.Annotations[meta.AnnotationKeyExternalName] = externalName
	}
}

func agrWithEtagAnnotation(eTag string) agrModifier {
	return func(i *v1alpha1.AccessGroupRule) {
		if i.ObjectMeta.Annotations == nil {
			i.ObjectMeta.Annotations = make(map[string]string)
		}
		i.ObjectMeta.Annotations[ibmc.ETagAnnotation] = eTag
	}
}

func agrWithSpec(p v1alpha1.AccessGroupRuleParameters) agrModifier {
	return func(r *v1alpha1.AccessGroupRule) { r.Spec.ForProvider = p }
}

func agrWithConditions(c ...cpv1alpha1.Condition) agrModifier {
	return func(i *v1alpha1.AccessGroupRule) { i.Status.SetConditions(c...) }
}

func agrWithStatus(p v1alpha1.AccessGroupRuleObservation) agrModifier {
	return func(r *v1alpha1.AccessGroupRule) { r.Status.AtProvider = p }
}

func agrParams(m ...func(*v1alpha1.AccessGroupRuleParameters)) *v1alpha1.AccessGroupRuleParameters {
	p := &v1alpha1.AccessGroupRuleParameters{
		Name:          agName,
		AccessGroupID: &accessGroupID,
		Expiration:    int64(expiration),
		RealmName:     realmName,
		Conditions: []v1alpha1.RuleCondition{
			{
				Claim:    claim1,
				Operator: opEquals,
				Value:    "true",
			},
		},
		TransactionID: &transactionID,
	}
	for _, f := range m {
		f(p)
	}
	return p
}

func agrObservation(m ...func(*v1alpha1.AccessGroupRuleObservation)) *v1alpha1.AccessGroupRuleObservation {
	o := &v1alpha1.AccessGroupRuleObservation{
		ID:               ruleID,
		CreatedAt:        ibmc.DateTimeToMetaV1Time(&createdAt),
		LastModifiedAt:   ibmc.DateTimeToMetaV1Time(&lastModifiedAt),
		CreatedByID:      createdByID,
		LastModifiedByID: createdByID,
		AccountID:        accountID,
	}

	for _, f := range m {
		f(o)
	}
	return o
}

func agrInstance(m ...func(*iamagv2.Rule)) *iamagv2.Rule {
	i := &iamagv2.Rule{
		ID:               &ruleID,
		Name:             &agName,
		AccountID:        &accountID,
		CreatedAt:        &createdAt,
		CreatedByID:      &createdByID,
		LastModifiedAt:   &lastModifiedAt,
		LastModifiedByID: &createdByID,
		Expiration:       ibmc.Int64Ptr(int64(expiration)),
		RealmName:        &realmName,
		AccessGroupID:    &accessGroupID,
		Conditions: []iamagv2.RuleConditions{
			{
				Claim:    &claim1,
				Operator: &opEquals,
				Value:    reference.ToPtrValue("true"),
			},
		},
	}

	for _, f := range m {
		f(i)
	}
	return i
}

func TestAccessGroupRuleObserve(t *testing.T) {
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
						_ = json.NewEncoder(w).Encode(&iamagv2.Rule{})
					},
				},
			},
			args: args{
				mg: agr(agrWithExternalNameAnnotation(ruleID), agrWithSpec(*agrParams())),
			},
			want: want{
				mg:  agr(agrWithSpec(*agrParams())),
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
				mg: agr(agrWithExternalNameAnnotation(ruleID), agrWithSpec(*agrParams())),
			},
			want: want{
				mg:  agr(agrWithSpec(*agrParams())),
				err: errors.New(errAgrBadRequest),
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
				mg: agr(agrWithExternalNameAnnotation(ruleID), agrWithSpec(*agrParams())),
			},
			want: want{
				mg:  agr(agrWithSpec(*agrParams())),
				err: errors.New(errAgrForbidden),
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
						cr := agrInstance()
						_ = json.NewEncoder(w).Encode(cr)
					},
				},
			},
			kube: &test.MockClient{
				MockUpdate: test.NewMockUpdateFn(nil),
			},
			args: args{
				mg: agr(
					agrWithExternalNameAnnotation(ruleID),
					agrWithSpec(*agrParams()),
				),
			},
			want: want{
				mg: agr(agrWithSpec(*agrParams()),
					agrWithConditions(cpv1alpha1.Available()),
					agrWithStatus(*agrObservation(func(cro *v1alpha1.AccessGroupRuleObservation) {
						cro.State = ibmcagr.StateActive
					})),
					agrWithEtagAnnotation(eTag)),
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
						cr := agrInstance(func(p *iamagv2.Rule) {
							p.Name = &ruleName2
						})
						_ = json.NewEncoder(w).Encode(cr)
					},
				},
			},
			kube: &test.MockClient{
				MockUpdate: test.NewMockUpdateFn(nil),
			},
			args: args{
				mg: agr(
					agrWithExternalNameAnnotation(ruleID),
					agrWithSpec(*agrParams()),
				),
			},
			want: want{
				mg: agr(agrWithSpec(*agrParams()),
					agrWithEtagAnnotation(eTag),
					agrWithConditions(cpv1alpha1.Available()),
					agrWithStatus(*agrObservation(func(cro *v1alpha1.AccessGroupRuleObservation) {
						cro.State = ibmcagr.StateActive
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
				BearerToken: ibmc.FakeBearerToken,
			}}
			mClient, _ := ibmc.NewClient(opts)
			e := agrExternal{
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

func TestAccessGroupRuleCreate(t *testing.T) {
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
						cr := agrInstance()
						_ = json.NewEncoder(w).Encode(cr)
					},
				},
			},
			args: args{
				mg: agr(agrWithSpec(*agrParams())),
			},
			want: want{
				mg: agr(agrWithSpec(*agrParams()),
					agrWithConditions(cpv1alpha1.Creating()),
					agrWithExternalNameAnnotation(ruleID)),
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
						cr := agrInstance()
						_ = json.NewEncoder(w).Encode(cr)
					},
				},
			},
			args: args{
				mg: agr(agrWithSpec(*agrParams())),
			},
			want: want{
				mg: agr(agrWithSpec(*agrParams()),
					agrWithConditions(cpv1alpha1.Creating())),
				cre: managed.ExternalCreation{ExternalNameAssigned: false},
				err: errors.Wrap(errors.New(http.StatusText(http.StatusBadRequest)), errCreateAccessGroupRule),
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
						cr := agrInstance()
						_ = json.NewEncoder(w).Encode(cr)
					},
				},
			},
			args: args{
				mg: agr(agrWithSpec(*agrParams())),
			},
			want: want{
				mg: agr(agrWithSpec(*agrParams()),
					agrWithConditions(cpv1alpha1.Creating())),
				cre: managed.ExternalCreation{ExternalNameAssigned: false},
				err: errors.Wrap(errors.New(http.StatusText(http.StatusConflict)), errCreateAccessGroupRule),
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
						cr := agrInstance()
						_ = json.NewEncoder(w).Encode(cr)
					},
				},
			},
			args: args{
				mg: agr(agrWithSpec(*agrParams())),
			},
			want: want{
				mg: agr(agrWithSpec(*agrParams()),
					agrWithConditions(cpv1alpha1.Creating())),
				cre: managed.ExternalCreation{ExternalNameAssigned: false},
				err: errors.Wrap(errors.New(http.StatusText(http.StatusForbidden)), errCreateAccessGroupRule),
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
				BearerToken: ibmc.FakeBearerToken,
			}}
			mClient, _ := ibmc.NewClient(opts)
			e := agrExternal{
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

func TestAccessGroupRuleDelete(t *testing.T) {
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
				mg: agr(agrWithStatus(*agrObservation()), agrWithSpec(*agrParams())),
			},
			want: want{
				mg:  agr(agrWithStatus(*agrObservation()), agrWithSpec(*agrParams()), agrWithConditions(cpv1alpha1.Deleting())),
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
				mg: agr(agrWithStatus(*agrObservation()), agrWithSpec(*agrParams())),
			},
			want: want{
				mg:  agr(agrWithStatus(*agrObservation()), agrWithSpec(*agrParams()), agrWithConditions(cpv1alpha1.Deleting())),
				err: errors.Wrap(errors.New(http.StatusText(http.StatusBadRequest)), errDeleteAccessGroupRule),
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
				mg: agr(agrWithStatus(*agrObservation()), agrWithSpec(*agrParams())),
			},
			want: want{
				mg:  agr(agrWithStatus(*agrObservation()), agrWithSpec(*agrParams()), agrWithConditions(cpv1alpha1.Deleting())),
				err: errors.Wrap(errors.New(http.StatusText(http.StatusUnauthorized)), errDeleteAccessGroupRule),
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
				mg: agr(agrWithStatus(*agrObservation()), agrWithSpec(*agrParams())),
			},
			want: want{
				mg:  agr(agrWithStatus(*agrObservation()), agrWithSpec(*agrParams()), agrWithConditions(cpv1alpha1.Deleting())),
				err: errors.Wrap(errors.New(http.StatusText(http.StatusForbidden)), errDeleteAccessGroupRule),
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
				BearerToken: ibmc.FakeBearerToken,
			}}
			mClient, _ := ibmc.NewClient(opts)
			e := agrExternal{
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

func TestAccessGroupRuleUpdate(t *testing.T) {
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
						cr := agrInstance()
						_ = json.NewEncoder(w).Encode(cr)
					},
				},
			},
			args: args{
				mg: agr(agrWithSpec(*agrParams()), agrWithStatus(*agrObservation()), agrWithEtagAnnotation(eTag)),
			},
			want: want{
				mg:  agr(agrWithSpec(*agrParams()), agrWithStatus(*agrObservation()), agrWithEtagAnnotation(eTag)),
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
				mg: agr(agrWithSpec(*agrParams()), agrWithStatus(*agrObservation()), agrWithEtagAnnotation(eTag)),
			},
			want: want{
				mg:  agr(agrWithSpec(*agrParams()), agrWithStatus(*agrObservation()), agrWithEtagAnnotation(eTag)),
				err: errors.Wrap(errors.New(http.StatusText(http.StatusBadRequest)), errUpdAccessGroupRule),
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
				mg: agr(agrWithSpec(*agrParams()), agrWithStatus(*agrObservation()), agrWithEtagAnnotation(eTag)),
			},
			want: want{
				mg:  agr(agrWithSpec(*agrParams()), agrWithStatus(*agrObservation()), agrWithEtagAnnotation(eTag)),
				err: errors.Wrap(errors.New(http.StatusText(http.StatusNotFound)), errUpdAccessGroupRule),
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
				BearerToken: ibmc.FakeBearerToken,
			}}
			mClient, _ := ibmc.NewClient(opts)
			e := agrExternal{
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
