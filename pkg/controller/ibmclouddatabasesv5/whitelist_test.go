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

package ibmclouddatabasesv5

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

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

	icdv5 "github.com/IBM/experimental-go-sdk/ibmclouddatabasesv5"

	"github.com/crossplane-contrib/provider-ibm-cloud/apis/ibmclouddatabasesv5/v1alpha1"
	"github.com/crossplane-contrib/provider-ibm-cloud/pkg/controller/tstutil"
)

var (
	ip1  = "195.212.0.0/16"
	ip1d = "Dev IP space 1"
	ip2  = "195.0.0.0/8"
	ip2d = "Dev IP space 2"
	ip3  = "46.5.0.0/16"
)

var _ managed.ExternalConnecter = &sgConnector{}
var _ managed.ExternalClient = &sgExternal{}

type wlModifier func(*v1alpha1.Whitelist)

func wl(im ...wlModifier) *v1alpha1.Whitelist {
	i := &v1alpha1.Whitelist{
		ObjectMeta: metav1.ObjectMeta{
			Name:       sgName,
			Finalizers: []string{},
			Annotations: map[string]string{
				meta.AnnotationKeyExternalName: id,
			},
		},
		Spec: v1alpha1.WhitelistSpec{
			ForProvider: v1alpha1.WhitelistParameters{},
		},
	}
	for _, m := range im {
		m(i)
	}
	return i
}

func wlWithExternalNameAnnotation(externalName string) wlModifier {
	return func(i *v1alpha1.Whitelist) {
		if i.ObjectMeta.Annotations == nil {
			i.ObjectMeta.Annotations = make(map[string]string)
		}
		i.ObjectMeta.Annotations[wtfConst] = externalName
	}
}

func wlWithSpec(p v1alpha1.WhitelistParameters) wlModifier {
	return func(r *v1alpha1.Whitelist) { r.Spec.ForProvider = p }
}

func wlWithConditions(c ...cpv1alpha1.Condition) wlModifier {
	return func(i *v1alpha1.Whitelist) { i.Status.SetConditions(c...) }
}

func wlWithStatus(p v1alpha1.WhitelistObservation) wlModifier {
	return func(r *v1alpha1.Whitelist) { r.Status.AtProvider = p }
}

func wlParams(m ...func(*v1alpha1.WhitelistParameters)) *v1alpha1.WhitelistParameters {
	p := &v1alpha1.WhitelistParameters{
		ID: &id,
		IPAddresses: []v1alpha1.WhitelistEntry{
			{
				Address:     ip1,
				Description: &ip1d,
			},
			{
				Address:     ip2,
				Description: &ip2d,
			},
		},
	}
	for _, f := range m {
		f(p)
	}
	return p
}

func wlObservation(m ...func(*v1alpha1.WhitelistObservation)) *v1alpha1.WhitelistObservation {
	o := &v1alpha1.WhitelistObservation{
		State: string(cpv1alpha1.Available().Reason),
	}

	for _, f := range m {
		f(o)
	}
	return o
}

func wlInstance(m ...func(*icdv5.Whitelist)) *icdv5.Whitelist {
	i := &icdv5.Whitelist{
		IpAddresses: []icdv5.WhitelistEntry{
			{
				Address:     &ip1,
				Description: &ip1d,
			},
			{
				Address:     &ip2,
				Description: &ip2d,
			},
		},
	}

	for _, f := range m {
		f(i)
	}
	return i
}

// Sets up a unit test http server, and creates an external white list structure appropriate for unit test.
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
func setupServerAndGetUnitTestExternalWL(testingObj *testing.T, handlers *[]tstutil.Handler, kube *client.Client) (*wlExternal, *httptest.Server, error) {
	mClient, tstServer, err := tstutil.SetupTestServerClient(testingObj, handlers)
	if err != nil || mClient == nil || tstServer == nil {
		return nil, nil, err
	}

	return &wlExternal{
			kube:   *kube,
			client: *mClient,
			logger: logging.NewNopLogger(),
		},
		tstServer,
		nil
}

func TestWhitelistObserve(t *testing.T) {
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
						_ = json.NewEncoder(w).Encode(&icdv5.Whitelist{})
					},
				},
			},
			args: tstutil.Args{
				Managed: wl(),
			},
			want: want{
				mg:  wl(),
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
						_ = json.NewEncoder(w).Encode(&icdv5.Whitelist{})
					},
				},
			},
			args: tstutil.Args{
				Managed: wl(),
			},
			want: want{
				mg:  wl(),
				err: errors.New(errBadRequest),
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
						_ = json.NewEncoder(w).Encode(&icdv5.Whitelist{})
					},
				},
			},
			args: tstutil.Args{
				Managed: wl(),
			},
			want: want{
				mg:  wl(),
				err: errors.New(errForbidden),
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
						_ = json.NewEncoder(w).Encode(wlInstance())
					},
				},
			},
			kube: &test.MockClient{
				MockUpdate: test.NewMockUpdateFn(nil),
			},
			args: tstutil.Args{
				Managed: wl(
					wlWithExternalNameAnnotation(id),
					wlWithSpec(*wlParams()),
				),
			},
			want: want{
				mg: wl(wlWithSpec(*wlParams()),
					wlWithConditions(cpv1alpha1.Available()),
					wlWithStatus(*wlObservation())),
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
						sg := wlInstance(func(p *icdv5.Whitelist) {
							p.IpAddresses[0].Address = &ip3
						})
						_ = json.NewEncoder(w).Encode(sg)
					},
				},
			},
			kube: &test.MockClient{
				MockUpdate: test.NewMockUpdateFn(nil),
			},
			args: tstutil.Args{
				Managed: wl(
					wlWithExternalNameAnnotation(id),
					wlWithSpec(*wlParams()),
				),
			},
			want: want{
				mg: wl(wlWithSpec(*wlParams()),
					wlWithConditions(cpv1alpha1.Available()),
					wlWithStatus(*wlObservation(func(p *v1alpha1.WhitelistObservation) {
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
			e, server, setupErr := setupServerAndGetUnitTestExternalWL(t, &tc.handlers, &tc.kube)
			if setupErr != nil {
				t.Errorf("Create(...): problem setting up the test server %s", setupErr)
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

func TestWhitelistCreate(t *testing.T) {
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
						if diff := cmp.Diff(http.MethodPut, r.Method); diff != "" {
							t.Errorf("r: -want, +got:\n%s", diff)
						}
						w.Header().Set("Content-Type", "application/json")
						w.WriteHeader(http.StatusCreated)
						_ = r.Body.Close()
						_ = json.NewEncoder(w).Encode(wlInstance())
					},
				},
			},
			args: tstutil.Args{
				Managed: wl(wlWithSpec(*wlParams())),
			},
			want: want{
				mg: wl(wlWithSpec(*wlParams()),
					wlWithConditions(cpv1alpha1.Creating()),
					wlWithExternalNameAnnotation(id)),
				cre: managed.ExternalCreation{ExternalNameAssigned: true},
				err: nil,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e, server, setupErr := setupServerAndGetUnitTestExternalWL(t, &tc.handlers, &tc.kube)
			if setupErr != nil {
				t.Errorf("Create(...): problem setting up the test server %s", setupErr)
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

func TestWhitelistDelete(t *testing.T) {
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
						if diff := cmp.Diff(http.MethodPut, r.Method); diff != "" {
							t.Errorf("r: -want, +got:\n%s", diff)
						}
						w.Header().Set("Content-Type", "application/json")
						w.WriteHeader(http.StatusAccepted)
						_ = r.Body.Close()
						_ = json.NewEncoder(w).Encode(wlInstance())
					},
				},
			},
			args: tstutil.Args{
				Managed: wl(wlWithStatus(*wlObservation())),
			},
			want: want{
				mg:  wl(wlWithStatus(*wlObservation()), wlWithConditions(cpv1alpha1.Deleting())),
				err: nil,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e, server, setupErr := setupServerAndGetUnitTestExternalWL(t, &tc.handlers, &tc.kube)
			if setupErr != nil {
				t.Errorf("Create(...): problem setting up the test server %s", setupErr)
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

func TestWhitelistUpdate(t *testing.T) {
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
						_ = json.NewEncoder(w).Encode(wlInstance())
					},
				},
			},
			args: tstutil.Args{
				Managed: wl(wlWithSpec(*wlParams()), wlWithStatus(*wlObservation())),
			},
			want: want{
				mg:  wl(wlWithSpec(*wlParams()), wlWithStatus(*wlObservation())),
				upd: managed.ExternalUpdate{},
				err: nil,
			},
		},
		"PatchFails": {
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
				Managed: wl(wlWithSpec(*wlParams()), wlWithStatus(*wlObservation())),
			},
			want: want{
				mg:  wl(wlWithSpec(*wlParams()), wlWithStatus(*wlObservation())),
				err: errors.New(http.StatusText(http.StatusBadRequest)),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e, server, setupErr := setupServerAndGetUnitTestExternalWL(t, &tc.handlers, &tc.kube)
			if setupErr != nil {
				t.Errorf("Create(...): problem setting up the test server %s", setupErr)
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
