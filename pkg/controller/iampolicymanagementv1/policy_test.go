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

	"github.com/go-openapi/strfmt"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

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
)

const (
	bearerTok     = "mock-token"
	errBadRequest = "error getting policy: Bad Request"
	errForbidden  = "error getting policy: Forbidden"
)

var (
	pName                = "myPolicy"
	policyTypeAccess     = "access"
	policyTypeAuth       = "authorization"
	policyAttributeName  = "iam_id"
	policyAttributeValue = "IBMid-123453user"
	createdByID          = "IBMid-123453user"
	roleID               = "crn:v1:bluemix:public:iam::::role:Editor"
	displayName          = "editor"
	roleDescription      = "role for editor"
	resAttr1Name         = "accountId"
	resAttr1Value        = "my-account-id"
	resAttr2Name         = "serviceName"
	resAttr2Value        = "cos"
	resAttr3Name         = "resource"
	resAttr3Value        = "mycos"
	resAttr3Operator     = "stringEquals"
	policyDescription    = "this is my policy 1"
	policyID             = "12345678-abcd-1a2b-a1b2-1234567890ab"
	createdAt, _         = strfmt.ParseDateTime("2020-10-31T02:33:06Z")
	lastModifiedAt, _    = strfmt.ParseDateTime("2020-10-31T03:33:06Z")
	hRef                 = "https://iam.cloud.ibm.com/v1/policies/12345678-abcd-1a2b-a1b2-1234567890ab"
	eTag                 = "1-eb832c7ff8c8016a542974b9f880b55e"
)

var _ managed.ExternalConnecter = &pConnector{}
var _ managed.ExternalClient = &pExternal{}

type pModifier func(*v1alpha1.Policy)

type handler struct {
	path        string
	handlerFunc func(w http.ResponseWriter, r *http.Request)
}

func p(im ...pModifier) *v1alpha1.Policy {
	i := &v1alpha1.Policy{
		ObjectMeta: metav1.ObjectMeta{
			Name:       pName,
			Finalizers: []string{},
			Annotations: map[string]string{
				meta.AnnotationKeyExternalName: policyID,
			},
		},
		Spec: v1alpha1.PolicySpec{
			ForProvider: v1alpha1.PolicyParameters{},
		},
	}
	for _, m := range im {
		m(i)
	}
	return i
}

func pWithExternalNameAnnotation(externalName string) pModifier {
	return func(i *v1alpha1.Policy) {
		if i.ObjectMeta.Annotations == nil {
			i.ObjectMeta.Annotations = make(map[string]string)
		}
		i.ObjectMeta.Annotations[meta.AnnotationKeyExternalName] = externalName
	}
}

func pWithEtagAnnotation(eTag string) pModifier {
	return func(i *v1alpha1.Policy) {
		if i.ObjectMeta.Annotations == nil {
			i.ObjectMeta.Annotations = make(map[string]string)
		}
		i.ObjectMeta.Annotations[ibmc.ETagAnnotation] = eTag
	}
}

func pWithSpec(p v1alpha1.PolicyParameters) pModifier {
	return func(r *v1alpha1.Policy) { r.Spec.ForProvider = p }
}

func pWithConditions(c ...cpv1alpha1.Condition) pModifier {
	return func(i *v1alpha1.Policy) { i.Status.SetConditions(c...) }
}

func pWithStatus(p v1alpha1.PolicyObservation) pModifier {
	return func(r *v1alpha1.Policy) { r.Status.AtProvider = p }
}

func params(m ...func(*v1alpha1.PolicyParameters)) *v1alpha1.PolicyParameters {
	p := &v1alpha1.PolicyParameters{
		Type: policyTypeAccess,
		Subjects: []v1alpha1.PolicySubject{
			{
				Attributes: []v1alpha1.SubjectAttribute{
					{
						Name:  &policyAttributeName,
						Value: &policyAttributeValue,
					},
				},
			},
		},
		Roles: []v1alpha1.PolicyRole{
			{
				RoleID: roleID,
			},
		},
		Resources: []v1alpha1.PolicyResource{
			{
				Attributes: []v1alpha1.ResourceAttribute{
					{
						Name:  &resAttr1Name,
						Value: &resAttr1Value,
					},
					{
						Name:  &resAttr2Name,
						Value: &resAttr2Value,
					},
					{
						Name:     &resAttr3Name,
						Value:    &resAttr3Value,
						Operator: &resAttr3Operator,
					},
				},
			},
		},
		Description: &policyDescription,
	}
	for _, f := range m {
		f(p)
	}
	return p
}

func observation(m ...func(*v1alpha1.PolicyObservation)) *v1alpha1.PolicyObservation {
	o := &v1alpha1.PolicyObservation{
		ID:               policyID,
		CreatedAt:        ibmc.DateTimeToMetaV1Time(&createdAt),
		LastModifiedAt:   ibmc.DateTimeToMetaV1Time(&lastModifiedAt),
		CreatedByID:      policyAttributeValue,
		LastModifiedByID: policyAttributeValue,
		Href:             hRef,
	}

	for _, f := range m {
		f(o)
	}
	return o
}

func instance(m ...func(*iampmv1.Policy)) *iampmv1.Policy {
	i := &iampmv1.Policy{
		ID:          &policyID,
		Type:        &policyTypeAccess,
		Description: &policyDescription,
		Subjects: []iampmv1.PolicySubject{
			{
				Attributes: []iampmv1.SubjectAttribute{
					{
						Name:  &policyAttributeName,
						Value: &policyAttributeValue,
					},
				},
			},
		},
		Roles: []iampmv1.PolicyRole{
			{
				RoleID:      &roleID,
				DisplayName: &displayName,
				Description: &roleDescription,
			},
		},
		Resources: []iampmv1.PolicyResource{
			{
				Attributes: []iampmv1.ResourceAttribute{
					{
						Name:  &resAttr1Name,
						Value: &resAttr1Value,
					},
					{
						Name:  &resAttr2Name,
						Value: &resAttr2Value,
					},
					{
						Name:     &resAttr3Name,
						Value:    &resAttr3Value,
						Operator: &resAttr3Operator,
					},
				},
			},
		},
		Href:             &hRef,
		CreatedAt:        &createdAt,
		CreatedByID:      &createdByID,
		LastModifiedAt:   &lastModifiedAt,
		LastModifiedByID: &createdByID,
	}

	for _, f := range m {
		f(i)
	}
	return i
}

func TestPolicyObserve(t *testing.T) {
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
						_ = json.NewEncoder(w).Encode(&iampmv1.Policy{})
					},
				},
			},
			args: args{
				mg: p(),
			},
			want: want{
				mg:  p(),
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
						_ = json.NewEncoder(w).Encode(&iampmv1.Policy{})
					},
				},
			},
			args: args{
				mg: p(),
			},
			want: want{
				mg:  p(),
				err: errors.New(errBadRequest),
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
						_ = json.NewEncoder(w).Encode(&iampmv1.Policy{})
					},
				},
			},
			args: args{
				mg: p(),
			},
			want: want{
				mg:  p(),
				err: errors.New(errForbidden),
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
						p := instance()
						_ = json.NewEncoder(w).Encode(p)
					},
				},
			},
			kube: &test.MockClient{
				MockUpdate: test.NewMockUpdateFn(nil),
			},
			args: args{
				mg: p(
					pWithExternalNameAnnotation(policyID),
					pWithSpec(*params()),
				),
			},
			want: want{
				mg: p(pWithSpec(*params()),
					pWithConditions(cpv1alpha1.Available()),
					pWithStatus(*observation()),
					pWithEtagAnnotation(eTag)),
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
						p := instance(func(p *iampmv1.Policy) {
							p.Type = &policyTypeAuth
						})
						_ = json.NewEncoder(w).Encode(p)
					},
				},
			},
			kube: &test.MockClient{
				MockUpdate: test.NewMockUpdateFn(nil),
			},
			args: args{
				mg: p(
					pWithExternalNameAnnotation(policyID),
					pWithSpec(*params()),
				),
			},
			want: want{
				mg: p(pWithSpec(*params()),
					pWithEtagAnnotation(eTag),
					pWithConditions(cpv1alpha1.Available()),
					pWithStatus(*observation(func(p *v1alpha1.PolicyObservation) {
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
			e := pExternal{
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

func TestPolicyCreate(t *testing.T) {
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
						p := instance()
						_ = json.NewEncoder(w).Encode(p)
					},
				},
			},
			args: args{
				mg: p(pWithSpec(*params())),
			},
			want: want{
				mg: p(pWithSpec(*params()),
					pWithConditions(cpv1alpha1.Creating()),
					pWithExternalNameAnnotation(policyID)),
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
						p := instance()
						_ = json.NewEncoder(w).Encode(p)
					},
				},
			},
			args: args{
				mg: p(pWithSpec(*params())),
			},
			want: want{
				mg: p(pWithSpec(*params()),
					pWithConditions(cpv1alpha1.Creating())),
				cre: managed.ExternalCreation{ExternalNameAssigned: false},
				err: errors.Wrap(errors.New(http.StatusText(http.StatusBadRequest)), errCreatePolicy),
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
						p := instance()
						_ = json.NewEncoder(w).Encode(p)
					},
				},
			},
			args: args{
				mg: p(pWithSpec(*params())),
			},
			want: want{
				mg: p(pWithSpec(*params()),
					pWithConditions(cpv1alpha1.Creating())),
				cre: managed.ExternalCreation{ExternalNameAssigned: false},
				err: errors.Wrap(errors.New(http.StatusText(http.StatusConflict)), errCreatePolicy),
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
						p := instance()
						_ = json.NewEncoder(w).Encode(p)
					},
				},
			},
			args: args{
				mg: p(pWithSpec(*params())),
			},
			want: want{
				mg: p(pWithSpec(*params()),
					pWithConditions(cpv1alpha1.Creating())),
				cre: managed.ExternalCreation{ExternalNameAssigned: false},
				err: errors.Wrap(errors.New(http.StatusText(http.StatusForbidden)), errCreatePolicy),
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
			e := pExternal{
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

func TestPolicyDelete(t *testing.T) {
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
				mg: p(pWithStatus(*observation())),
			},
			want: want{
				mg:  p(pWithStatus(*observation()), pWithConditions(cpv1alpha1.Deleting())),
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
				mg: p(pWithStatus(*observation())),
			},
			want: want{
				mg:  p(pWithStatus(*observation()), pWithConditions(cpv1alpha1.Deleting())),
				err: errors.Wrap(errors.New(http.StatusText(http.StatusBadRequest)), errDeletePolicy),
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
				mg: p(pWithStatus(*observation())),
			},
			want: want{
				mg:  p(pWithStatus(*observation()), pWithConditions(cpv1alpha1.Deleting())),
				err: errors.Wrap(errors.New(http.StatusText(http.StatusUnauthorized)), errDeletePolicy),
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
				mg: p(pWithStatus(*observation())),
			},
			want: want{
				mg:  p(pWithStatus(*observation()), pWithConditions(cpv1alpha1.Deleting())),
				err: errors.Wrap(errors.New(http.StatusText(http.StatusForbidden)), errDeletePolicy),
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
			e := pExternal{
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

func TestPolicyUpdate(t *testing.T) {
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
						p := instance()
						_ = json.NewEncoder(w).Encode(p)
					},
				},
			},
			args: args{
				mg: p(pWithSpec(*params()), pWithStatus(*observation()), pWithEtagAnnotation(eTag)),
			},
			want: want{
				mg:  p(pWithSpec(*params()), pWithStatus(*observation()), pWithEtagAnnotation(eTag)),
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
				mg: p(pWithSpec(*params()), pWithStatus(*observation())),
			},
			want: want{
				mg:  p(pWithSpec(*params()), pWithStatus(*observation())),
				err: errors.Wrap(errors.New(http.StatusText(http.StatusBadRequest)), errUpdPolicy),
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
				mg: p(pWithSpec(*params()), pWithStatus(*observation())),
			},
			want: want{
				mg:  p(pWithSpec(*params()), pWithStatus(*observation())),
				err: errors.Wrap(errors.New(http.StatusText(http.StatusNotFound)), errUpdPolicy),
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
			e := pExternal{
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
