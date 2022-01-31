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

	iamagv2 "github.com/IBM/platform-services-go-sdk/iamaccessgroupsv2"

	"github.com/crossplane-contrib/provider-ibm-cloud/apis/iamaccessgroupsv2/v1alpha1"
	ibmc "github.com/crossplane-contrib/provider-ibm-cloud/pkg/clients"
	ibmcagr "github.com/crossplane-contrib/provider-ibm-cloud/pkg/clients/accessgrouprule"
	"github.com/crossplane-contrib/provider-ibm-cloud/pkg/controller/tstutil"
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

// Sets up a unit test http server, and creates an external access-group-rule structure appropriate for unit test.
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
func setupServerAndGetUnitTestExternalAGR(testingObj *testing.T, handlers *[]tstutil.Handler, kube *client.Client) (*agrExternal, *httptest.Server, error) {
	mClient, tstServer, err := tstutil.SetupTestServerClient(testingObj, handlers)
	if err != nil || mClient == nil || tstServer == nil {
		return nil, nil, err
	}

	return &agrExternal{
			kube:   *kube,
			client: *mClient,
			logger: logging.NewNopLogger(),
		},
		tstServer,
		nil
}
func TestAccessGroupRuleObserve(t *testing.T) {
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
						_ = json.NewEncoder(w).Encode(&iamagv2.Rule{})
					},
				},
			},
			args: tstutil.Args{
				Managed: agr(agrWithExternalNameAnnotation(ruleID), agrWithSpec(*agrParams())),
			},
			want: want{
				mg:  agr(agrWithSpec(*agrParams())),
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
						_ = json.NewEncoder(w).Encode(&iamagv2.Group{})
					},
				},
			},
			args: tstutil.Args{
				Managed: agr(agrWithExternalNameAnnotation(ruleID), agrWithSpec(*agrParams())),
			},
			want: want{
				mg:  agr(agrWithSpec(*agrParams())),
				err: errors.New(errAgrBadRequest),
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
						_ = json.NewEncoder(w).Encode(&iamagv2.Group{})
					},
				},
			},
			args: tstutil.Args{
				Managed: agr(agrWithExternalNameAnnotation(ruleID), agrWithSpec(*agrParams())),
			},
			want: want{
				mg:  agr(agrWithSpec(*agrParams())),
				err: errors.New(errAgrForbidden),
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
						cr := agrInstance()
						_ = json.NewEncoder(w).Encode(cr)
					},
				},
			},
			kube: &test.MockClient{
				MockUpdate: test.NewMockUpdateFn(nil),
			},
			args: tstutil.Args{
				Managed: agr(
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
			args: tstutil.Args{
				Managed: agr(
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
			e, server, err := setupServerAndGetUnitTestExternalAGR(t, &tc.handlers, &tc.kube)
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

func TestAccessGroupRuleCreate(t *testing.T) {
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
						cr := agrInstance()
						_ = json.NewEncoder(w).Encode(cr)
					},
				},
			},
			args: tstutil.Args{
				Managed: agr(agrWithSpec(*agrParams())),
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
						cr := agrInstance()
						_ = json.NewEncoder(w).Encode(cr)
					},
				},
			},
			args: tstutil.Args{
				Managed: agr(agrWithSpec(*agrParams())),
			},
			want: want{
				mg: agr(agrWithSpec(*agrParams()),
					agrWithConditions(cpv1alpha1.Creating())),
				cre: managed.ExternalCreation{ExternalNameAssigned: false},
				err: errors.Wrap(errors.New(http.StatusText(http.StatusBadRequest)), errCreateAccessGroupRule),
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
						cr := agrInstance()
						_ = json.NewEncoder(w).Encode(cr)
					},
				},
			},
			args: tstutil.Args{
				Managed: agr(agrWithSpec(*agrParams())),
			},
			want: want{
				mg: agr(agrWithSpec(*agrParams()),
					agrWithConditions(cpv1alpha1.Creating())),
				cre: managed.ExternalCreation{ExternalNameAssigned: false},
				err: errors.Wrap(errors.New(http.StatusText(http.StatusConflict)), errCreateAccessGroupRule),
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
						cr := agrInstance()
						_ = json.NewEncoder(w).Encode(cr)
					},
				},
			},
			args: tstutil.Args{
				Managed: agr(agrWithSpec(*agrParams())),
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
			e, server, errSetup := setupServerAndGetUnitTestExternalAGR(t, &tc.handlers, &tc.kube)
			if errSetup != nil {
				t.Errorf("Create(...): problem setting up the test server %s", errSetup)
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

func TestAccessGroupRuleDelete(t *testing.T) {
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
				Managed: agr(agrWithStatus(*agrObservation()), agrWithSpec(*agrParams())),
			},
			want: want{
				mg:  agr(agrWithStatus(*agrObservation()), agrWithSpec(*agrParams()), agrWithConditions(cpv1alpha1.Deleting())),
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
				Managed: agr(agrWithStatus(*agrObservation()), agrWithSpec(*agrParams())),
			},
			want: want{
				mg:  agr(agrWithStatus(*agrObservation()), agrWithSpec(*agrParams()), agrWithConditions(cpv1alpha1.Deleting())),
				err: errors.Wrap(errors.New(http.StatusText(http.StatusBadRequest)), errDeleteAccessGroupRule),
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
				Managed: agr(agrWithStatus(*agrObservation()), agrWithSpec(*agrParams())),
			},
			want: want{
				mg:  agr(agrWithStatus(*agrObservation()), agrWithSpec(*agrParams()), agrWithConditions(cpv1alpha1.Deleting())),
				err: errors.Wrap(errors.New(http.StatusText(http.StatusUnauthorized)), errDeleteAccessGroupRule),
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
				Managed: agr(agrWithStatus(*agrObservation()), agrWithSpec(*agrParams())),
			},
			want: want{
				mg:  agr(agrWithStatus(*agrObservation()), agrWithSpec(*agrParams()), agrWithConditions(cpv1alpha1.Deleting())),
				err: errors.Wrap(errors.New(http.StatusText(http.StatusForbidden)), errDeleteAccessGroupRule),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e, server, errServer := setupServerAndGetUnitTestExternalAGR(t, &tc.handlers, &tc.kube)
			if errServer != nil {
				t.Errorf("Create(...): problem setting up the test server %s", errServer)
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

func TestAccessGroupRuleUpdate(t *testing.T) {
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
						cr := agrInstance()
						_ = json.NewEncoder(w).Encode(cr)
					},
				},
			},
			args: tstutil.Args{
				Managed: agr(agrWithSpec(*agrParams()), agrWithStatus(*agrObservation()), agrWithEtagAnnotation(eTag)),
			},
			want: want{
				mg:  agr(agrWithSpec(*agrParams()), agrWithStatus(*agrObservation()), agrWithEtagAnnotation(eTag)),
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
				Managed: agr(agrWithSpec(*agrParams()), agrWithStatus(*agrObservation()), agrWithEtagAnnotation(eTag)),
			},
			want: want{
				mg:  agr(agrWithSpec(*agrParams()), agrWithStatus(*agrObservation()), agrWithEtagAnnotation(eTag)),
				err: errors.Wrap(errors.New(http.StatusText(http.StatusBadRequest)), errUpdAccessGroupRule),
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
				Managed: agr(agrWithSpec(*agrParams()), agrWithStatus(*agrObservation()), agrWithEtagAnnotation(eTag)),
			},
			want: want{
				mg:  agr(agrWithSpec(*agrParams()), agrWithStatus(*agrObservation()), agrWithEtagAnnotation(eTag)),
				err: errors.Wrap(errors.New(http.StatusText(http.StatusNotFound)), errUpdAccessGroupRule),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e, server, err := setupServerAndGetUnitTestExternalAGR(t, &tc.handlers, &tc.kube)
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
