/*
Copyright 2021 The Crossplane Authors.

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

package containerv2

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	ibmContainerV2 "github.com/IBM-Cloud/bluemix-go/api/container/containerv2"
	cpv1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	crossplaneApi "github.com/crossplane-contrib/provider-ibm-cloud/apis/container/containerv2/v1alpha1"
	crossplaneClient "github.com/crossplane-contrib/provider-ibm-cloud/pkg/clients/container/containerv2"
	"github.com/crossplane-contrib/provider-ibm-cloud/pkg/controller/tstutil"
)

// Interface to a function that takes as argument a cluster create request, and modifies it
type clusterModifier func(*crossplaneApi.Cluster)

// Creates a cluster, by creating a generic one + applying a list of modifiers to the Spec part
func createCrossplaneCluster(modifiers ...clusterModifier) *crossplaneApi.Cluster {
	result := &crossplaneApi.Cluster{
		Spec: crossplaneApi.ClusterSpec{
			ForProvider: *crossplaneClient.GetClusterCreateCrossplaneRequest(),
		},
		Status: crossplaneApi.ClusterStatus{
			AtProvider: crossplaneClient.GetContainerClusterInfo(),
		},
	}

	for _, modifier := range modifiers {
		modifier(result)
	}

	return result
}

// Returns a string
func aStr() string {
	return "foobar"
}

// Returns a function that sets the cluster conditions
func withConditions(c ...cpv1alpha1.Condition) clusterModifier {
	return func(i *crossplaneApi.Cluster) {
		i.Status.SetConditions(c...)
	}
}

// Converts a crossplane cluster to an IBM-cloud one
func crossplaneToIBMCloud(c *crossplaneApi.Cluster) *ibmContainerV2.ClusterInfo {
	return &ibmContainerV2.ClusterInfo{
		Name: c.Spec.ProviderConfigReference.Name,
		// CreatedDate: &c.Status.AtProvider.DeepCopy().CreatedDate.Time,
	}
}

// Sets up a unit test http server, and creates an external bucket appropriate for unit test.
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
func setupServerAndGetUnitTestExternal(testingObj *testing.T, handlers *[]tstutil.Handler, kube *client.Client) (*clusterExternal, *httptest.Server, error) {
	mClient, tstServer, err := tstutil.SetupTestServerClient(testingObj, handlers)
	if err != nil || mClient == nil || tstServer == nil {
		return nil, nil, err
	}

	return &clusterExternal{
			kube:   *kube,
			client: *mClient,
			logger: logging.NewNopLogger(),
		},
		tstServer,
		nil
}

// Tests the cluster "Create" method
func TestCreate(t *testing.T) {
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
						_ = r.Body.Close()

						if diff := cmp.Diff(http.MethodPost, r.Method); diff != "" {
							t.Errorf("r: -want, +got:\n%s", diff)
						}
						w.Header().Set("Content-Type", "application/json")
						w.WriteHeader(http.StatusCreated)

						_ = json.NewEncoder(w).Encode(ibmContainerV2.ClusterCreateResponse{
							ID: aStr(),
						})
					},
				},
			},
			args: tstutil.Args{
				Managed: createCrossplaneCluster(),
			},
			want: want{
				mg:  createCrossplaneCluster(withConditions(cpv1alpha1.Creating())),
				cre: managed.ExternalCreation{ExternalNameAssigned: true},
				err: nil,
			},
		},
		"Failed": {
			handlers: []tstutil.Handler{
				{
					Path: "/",
					HandlerFunc: func(w http.ResponseWriter, r *http.Request) {
						_ = r.Body.Close()

						if diff := cmp.Diff(http.MethodPost, r.Method); diff != "" {
							t.Errorf("r: -want, +got:\n%s", diff)
						}
						w.Header().Set("Content-Type", "application/json")
						w.WriteHeader(http.StatusBadRequest)
					},
				},
			},
			args: tstutil.Args{
				Managed: createCrossplaneCluster(),
			},
			want: want{
				mg:  createCrossplaneCluster(withConditions(cpv1alpha1.Creating())),
				cre: managed.ExternalCreation{ExternalNameAssigned: false},
				err: errors.Wrap(errors.New(http.StatusText(http.StatusBadRequest)), errCreateCluster),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e, server, err := setupServerAndGetUnitTestExternal(t, &tc.handlers, &tc.kube)
			if err != nil {
				t.Errorf("Create(...): problem setting up the test server %s", err)
			}

			defer server.Close()

			cre, err := e.Create(context.Background(), tc.args.Managed)
			if tc.want.err != nil && err != nil {
				wantedPrefix := strings.Split(tc.want.err.Error(), ":")[0]
				actualPrefix := strings.Split(err.Error(), ":")[0]
				if diff := cmp.Diff(wantedPrefix, actualPrefix); diff != "" {
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
		})
	}
}

// Tests the cluster "Delete" method
func TestDelete(t *testing.T) {
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
						_ = r.Body.Close()

						if diff := cmp.Diff(http.MethodDelete, r.Method); diff != "" {
							t.Errorf("r: -want, +got:\n%s", diff)
						}

						w.Header().Set("Content-Type", "application/json")
						w.WriteHeader(http.StatusAccepted)
					},
				},
			},
			args: tstutil.Args{
				Managed: createCrossplaneCluster(),
			},
			want: want{
				mg:  createCrossplaneCluster(withConditions(cpv1alpha1.Deleting())),
				err: nil,
			},
		},
		"AlreadyGone": {
			handlers: []tstutil.Handler{
				{
					Path: "/",
					HandlerFunc: func(w http.ResponseWriter, r *http.Request) {
						_ = r.Body.Close()

						if diff := cmp.Diff(http.MethodDelete, r.Method); diff != "" {
							t.Errorf("r: -want, +got:\n%s", diff)
						}

						w.Header().Set("Content-Type", "application/json")
						w.WriteHeader(http.StatusNotFound)
					},
				},
			},
			args: tstutil.Args{
				Managed: createCrossplaneCluster(),
			},
			want: want{
				mg:  createCrossplaneCluster(withConditions(cpv1alpha1.Deleting())),
				err: errors.Wrap(errors.New(http.StatusText(http.StatusNotFound)), errDeleteCluster),
			},
		},
		"Failed": {
			handlers: []tstutil.Handler{
				{
					Path: "/",
					HandlerFunc: func(w http.ResponseWriter, r *http.Request) {
						_ = r.Body.Close()

						if diff := cmp.Diff(http.MethodDelete, r.Method); diff != "" {
							t.Errorf("r: -want, +got:\n%s", diff)
						}

						w.Header().Set("Content-Type", "application/json")
						w.WriteHeader(http.StatusBadRequest)
					},
				},
			},
			args: tstutil.Args{
				Managed: createCrossplaneCluster(),
			},
			want: want{
				mg:  createCrossplaneCluster(withConditions(cpv1alpha1.Deleting())),
				err: errors.Wrap(errors.New(http.StatusText(http.StatusBadRequest)), errDeleteCluster),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e, server, setupErr := setupServerAndGetUnitTestExternal(t, &tc.handlers, &tc.kube)
			if setupErr != nil {
				t.Errorf("Create(...): problem setting up the test server %s", setupErr)
			}

			defer server.Close()

			err := e.Delete(context.Background(), tc.args.Managed)
			if tc.want.err != nil && err != nil {
				// the case where our mock server returns error, is tricky, as the returned error string is long/spans multiple lines
				wantedPrefix := strings.Split(tc.want.err.Error(), ":")[0]
				actualPrefix := strings.Split(err.Error(), ":")[0]
				if diff := cmp.Diff(wantedPrefix, actualPrefix); diff != "" {
					t.Errorf("Create(...): -want, +got:\n%s", diff)
				}
			} else if diff := cmp.Diff(tc.want.err, err); diff != "" {
				t.Errorf("Delete(...): -want, +got:\n%s", diff)
			}

			if diff := cmp.Diff(tc.want.mg, tc.args.Managed); diff != "" {
				t.Errorf("Delete(...): -want, +got:\n%s", diff)
			}
		})
	}
}

/*
// Tests the cluster "Observe" method
func TestBucketObserve(t *testing.T) {
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
					},
				},
			},
			args: tstutil.Args{
				Managed: createCrossplaneCluster(),
			},
			want: want{
				mg:  createCrossplaneCluster(withConditions()),
				obs: managed.ExternalObservation{ResourceExists: false},
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
					},
				},
			},
			args: tstutil.Args{
				Managed: createCrossplaneCluster(),
			},
			want: want{
				mg:  createCrossplaneCluster(withConditions()),
				obs: managed.ExternalObservation{},
				err: errors.Wrap(errors.New(http.StatusText(http.StatusBadRequest)), errGetClusterFailed),
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
					},
				},
			},
			args: tstutil.Args{
				Managed: createCrossplaneCluster(),
			},
			want: want{
				mg:  createCrossplaneCluster(withConditions()),
				obs: managed.ExternalObservation{},
				err: errors.Wrap(errors.New(http.StatusText(http.StatusBadRequest)), errGetClusterFailed),
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
						w.WriteHeader(http.StatusOK)

					},
				},
			},
			args: tstutil.Args{
				Managed: createCrossplaneBucket(withBucketAtProvider(*bucketObservation())),
			},
			want: want{
				mg: createCrossplaneBucket(withBucketAtProvider(*bucketObservation())),
				obs: managed.ExternalObservation{
					ResourceExists:    true,
					ResourceUpToDate:  true,
					ConnectionDetails: nil,
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e, server, errCr := setupServerAndGetUnitTestExternal(t, &tc.handlers, &tc.kube)
			if errCr != nil {
				t.Errorf("Delete(...): problem setting up the test server %s", errCr)
			}

			defer server.Close()

			obs, err := e.Observe(context.Background(), tc.args.Managed)
			if tc.want.err != nil && err != nil {
				// the case where our mock server returns error, is tricky, as the returned error string is long/spans multiple lines
				expectedNoSpace := strings.ReplaceAll(tc.want.err.Error(), " ", "")
				returnedNoSpace := strings.ReplaceAll(err.Error(), " ", "")
				if strings.HasPrefix(returnedNoSpace, expectedNoSpace) == false {
					diff := cmp.Diff(tc.want.err.Error(), err.Error())
					t.Errorf("Observe(...): -want, +got:\n%s", diff)
				}
			} else if diff := cmp.Diff(tc.want.err, err); diff != "" {
				t.Errorf("Observe(...): want error != got error:\n%s", diff)
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
*/
