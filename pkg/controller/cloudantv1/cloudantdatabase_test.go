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

package cloudantv1

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
	"github.com/crossplane/crossplane-runtime/pkg/reference"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/crossplane/crossplane-runtime/pkg/test"

	cv1 "github.com/IBM/cloudant-go-sdk/cloudantv1"

	"github.com/crossplane-contrib/provider-ibm-cloud/apis/cloudantv1/v1alpha1"
	ibmc "github.com/crossplane-contrib/provider-ibm-cloud/pkg/clients"
	"github.com/crossplane-contrib/provider-ibm-cloud/pkg/controller/tstutil"
)

var (
	cdbName = "mycloudantdatabase"
)

var _ managed.ExternalConnecter = &cloudantdatabaseConnector{}
var _ managed.ExternalClient = &cloudantdatabaseExternal{}

type cdbModifier func(*v1alpha1.CloudantDatabase)

func cloudantdatabase(im ...cdbModifier) *v1alpha1.CloudantDatabase {
	i := &v1alpha1.CloudantDatabase{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "mycloudantdatabase",
			Finalizers: []string{},
			Annotations: map[string]string{
				meta.AnnotationKeyExternalName: "mycloudantdatabase",
			},
		},
		Spec: v1alpha1.CloudantDatabaseSpec{
			ForProvider: v1alpha1.CloudantDatabaseParameters{},
		},
	}
	for _, m := range im {
		m(i)
	}
	return i
}

func cdbWithExternalNameAnnotation(externalName string) cdbModifier {
	return func(i *v1alpha1.CloudantDatabase) {
		if i.ObjectMeta.Annotations == nil {
			i.ObjectMeta.Annotations = make(map[string]string)
		}
		i.ObjectMeta.Annotations[meta.AnnotationKeyExternalName] = externalName
	}
}

func cdbWithSpec(p v1alpha1.CloudantDatabaseParameters) cdbModifier {
	return func(r *v1alpha1.CloudantDatabase) { r.Spec.ForProvider = p }
}

func cdbWithConditions(c ...cpv1alpha1.Condition) cdbModifier {
	return func(i *v1alpha1.CloudantDatabase) { i.Status.SetConditions(c...) }
}

func cdbWithStatus(p v1alpha1.CloudantDatabaseObservation) cdbModifier {
	return func(r *v1alpha1.CloudantDatabase) { r.Status.AtProvider = p }
}

func cdbParams(m ...func(*v1alpha1.CloudantDatabaseParameters)) *v1alpha1.CloudantDatabaseParameters {
	p := &v1alpha1.CloudantDatabaseParameters{
		Db:          "mycloudantdatabase",
		Partitioned: ibmc.BoolPtr(false),
		Q:           ibmc.Int64Ptr(int64(2)),
	}
	for _, f := range m {
		f(p)
	}
	return p
}

func cdbEmptyObservation(m ...func(*v1alpha1.CloudantDatabaseObservation)) *v1alpha1.CloudantDatabaseObservation {
	o := &v1alpha1.CloudantDatabaseObservation{
		Cluster:            nil,
		CommittedUpdateSeq: "",
		CompactRunning:     false,
		CompactedSeq:       "",
		DiskFormatVersion:  0,
		DocCount:           0,
		DocDelCount:        0,
		Engine:             "",
		Sizes:              nil,
		UpdateSeq:          "",
		UUID:               "",
		State:              "",
	}

	for _, f := range m {
		f(o)
	}
	return o
}

func cdbObservation(m ...func(*v1alpha1.CloudantDatabaseObservation)) *v1alpha1.CloudantDatabaseObservation {
	o := &v1alpha1.CloudantDatabaseObservation{
		Cluster:            generateTestv1alpha1DatabaseInformationCluster(),
		CommittedUpdateSeq: "myCommittedUpdateSeq",
		CompactRunning:     false,
		CompactedSeq:       "myCompactedSeq",
		DiskFormatVersion:  int64(2),
		DocCount:           int64(2),
		DocDelCount:        int64(2),
		Engine:             "myEngine",
		Sizes:              generateTestv1alpha1ContentInformationSizes(),
		UpdateSeq:          "myUpdateSeq",
		UUID:               "myUUID",
		State:              "active",
	}

	for _, f := range m {
		f(o)
	}
	return o
}

func generateTestv1alpha1DatabaseInformationCluster() *v1alpha1.DatabaseInformationCluster {
	o := &v1alpha1.DatabaseInformationCluster{
		N: int64(2),
		R: int64(2),
		W: int64(2),
	}
	return o
}

func generateTestv1alpha1ContentInformationSizes() *v1alpha1.ContentInformationSizes {
	o := &v1alpha1.ContentInformationSizes{
		Active:   int64(2),
		External: int64(2),
		File:     int64(2),
	}
	return o
}

func cdbInstance(m ...func(*cv1.DatabaseInformation)) *cv1.DatabaseInformation {
	i := &cv1.DatabaseInformation{
		Cluster:            generateTestcv1DatabaseInformationCluster(),
		CommittedUpdateSeq: reference.ToPtrValue("myCommittedUpdateSeq"),
		CompactRunning:     ibmc.BoolPtr(false),
		CompactedSeq:       reference.ToPtrValue("myCompactedSeq"),
		DbName:             reference.ToPtrValue("mycloudantdatabase"),
		DiskFormatVersion:  ibmc.Int64Ptr(int64(2)),
		DocCount:           ibmc.Int64Ptr(int64(2)),
		DocDelCount:        ibmc.Int64Ptr(int64(2)),
		Engine:             reference.ToPtrValue("myEngine"),
		Props:              generateTestcv1DatabaseInformationProps(),
		Sizes:              generateTestcv1ContentInformationSizes(),
		UpdateSeq:          reference.ToPtrValue("myUpdateSeq"),
		UUID:               reference.ToPtrValue("myUUID"),
	}

	for _, f := range m {
		f(i)
	}
	return i
}

func generateTestcv1DatabaseInformationCluster() *cv1.DatabaseInformationCluster {
	o := &cv1.DatabaseInformationCluster{
		N: ibmc.Int64Ptr(int64(2)),
		Q: ibmc.Int64Ptr(int64(2)),
		R: ibmc.Int64Ptr(int64(2)),
		W: ibmc.Int64Ptr(int64(2)),
	}
	return o
}

func generateTestcv1DatabaseInformationProps() *cv1.DatabaseInformationProps {
	o := &cv1.DatabaseInformationProps{
		Partitioned: ibmc.BoolPtr(false),
	}
	return o
}

func generateTestcv1ContentInformationSizes() *cv1.ContentInformationSizes {
	o := &cv1.ContentInformationSizes{
		Active:   ibmc.Int64Ptr(int64(2)),
		External: ibmc.Int64Ptr(int64(2)),
		File:     ibmc.Int64Ptr(int64(2)),
	}
	return o
}

// Sets up a unit test http server, and creates an external cloudant db structure appropriate for unit test.
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
func setupServerAndGetUnitTestExternal(testingObj *testing.T, handlers *[]tstutil.Handler, kube *client.Client) (*cloudantdatabaseExternal, *httptest.Server, error) {
	mClient, tstServer, err := tstutil.SetupTestServerClient(testingObj, handlers)
	if err != nil || mClient == nil || tstServer == nil {
		return nil, nil, err
	}

	return &cloudantdatabaseExternal{
			kube:   *kube,
			client: *mClient,
			logger: logging.NewNopLogger(),
		},
		tstServer,
		nil
}
func TestCloudantDatabaseObserve(t *testing.T) {
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
						_ = json.NewEncoder(w).Encode(&cv1.DatabaseInformation{})
					},
				},
			},
			args: tstutil.Args{
				Managed: cloudantdatabase(),
			},
			want: want{
				mg:  cloudantdatabase(),
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
						_ = json.NewEncoder(w).Encode(&cv1.DatabaseInformation{})
					},
				},
			},
			args: tstutil.Args{
				Managed: cloudantdatabase(),
			},
			want: want{
				mg:  cloudantdatabase(),
				err: errors.New(errGetCloudantDatabaseFailed + ": Bad Request"),
			},
		},
		"ObservedCloudantDatabaseUpToDate": {
			handlers: []tstutil.Handler{
				{
					Path: "/",
					HandlerFunc: func(w http.ResponseWriter, r *http.Request) {
						_ = r.Body.Close()
						if diff := cmp.Diff(http.MethodGet, r.Method); diff != "" {
							t.Errorf("r: -want, +got:\n%s", diff)
						}
						w.Header().Set("Content-Type", "application/json")
						cdb := cdbInstance()
						_ = json.NewEncoder(w).Encode(cdb)
					},
				},
			},
			kube: &test.MockClient{
				MockUpdate: test.NewMockUpdateFn(nil),
			},
			args: tstutil.Args{
				Managed: cloudantdatabase(
					cdbWithExternalNameAnnotation(cdbName),
					cdbWithSpec(*cdbParams()),
					cdbWithStatus(*cdbEmptyObservation(func(p *v1alpha1.CloudantDatabaseObservation) { p.State = "active" })),
				),
			},
			want: want{
				mg: cloudantdatabase(cdbWithSpec(*cdbParams()),
					cdbWithConditions(cpv1alpha1.Available()),
					cdbWithStatus(*cdbObservation()),
				),
				obs: managed.ExternalObservation{
					ResourceExists:    true,
					ResourceUpToDate:  true,
					ConnectionDetails: nil,
				},
			},
		},
		// "ObservedCloudantDatabaseNotUpToDate": {
		// 	handlers: []handler{
		// 		{
		// 			path: "/",
		// 			handlerFunc: func(w http.ResponseWriter, r *http.Request) {
		// 				_ = r.Body.Close()
		// 				if diff := cmp.Diff(http.MethodGet, r.Method); diff != "" {
		// 					t.Errorf("r: -want, +got:\n%s", diff)
		// 				}
		// 				w.Header().Set("Content-Type", "application/json")
		// 				cdb := cdbInstance(func(p *cv1.DatabaseInformation) {
		// 					p.Props.Partitioned = nil
		// 				})
		// 				_ = json.NewEncoder(w).Encode(cdb)
		// 			},
		// 		},
		// 	},
		// 	kube: &test.MockClient{
		// 		MockUpdate: test.NewMockUpdateFn(nil),
		// 	},
		// 	args: args{
		// 		mg: cloudantdatabase(
		// 			cdbWithExternalNameAnnotation(cdbName),
		// 			cdbWithSpec(*cdbParams()),
		// 			cdbWithStatus(*cdbEmptyObservation(func(p *v1alpha1.CloudantDatabaseObservation) { p.State = "active" })),
		// 		),
		// 	},
		// 	want: want{
		// 		mg: cloudantdatabase(cdbWithSpec(*cdbParams()),
		// 			cdbWithConditions(cpv1alpha1.Available()),
		// 			cdbWithStatus(*cdbObservation()),
		// 		),
		// 		obs: managed.ExternalObservation{
		// 			ResourceExists:    true,
		// 			ResourceUpToDate:  false,
		// 			ConnectionDetails: nil,
		// 		},
		// 	},
		// },
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e, server, setupErr := setupServerAndGetUnitTestExternal(t, &tc.handlers, &tc.kube)
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

func TestCloudantDatabaseCreate(t *testing.T) {
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
						cdb := cdbInstance()
						_ = json.NewEncoder(w).Encode(cdb)
					},
				},
			},
			args: tstutil.Args{
				Managed: cloudantdatabase(cdbWithSpec(*cdbParams())),
			},
			want: want{
				mg: cloudantdatabase(cdbWithSpec(*cdbParams()),
					cdbWithConditions(cpv1alpha1.Creating()),
					cdbWithExternalNameAnnotation(cdbName)),
				cre: managed.ExternalCreation{ExternalNameAssigned: true},
				err: nil,
			},
		},
		"Failed": {
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
						cdb := cdbInstance()
						_ = json.NewEncoder(w).Encode(cdb)
					},
				},
			},
			args: tstutil.Args{
				Managed: cloudantdatabase(cdbWithSpec(*cdbParams())),
			},
			want: want{
				mg: cloudantdatabase(cdbWithSpec(*cdbParams()),
					cdbWithConditions(cpv1alpha1.Creating())),
				cre: managed.ExternalCreation{ExternalNameAssigned: false},
				err: errors.Wrap(errors.New(http.StatusText(http.StatusBadRequest)), errCreateCloudantDatabase),
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

func TestCloudantDatabaseDelete(t *testing.T) {
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
				Managed: cloudantdatabase(cdbWithExternalNameAnnotation(cdbName)),
			},
			want: want{
				mg:  cloudantdatabase(cdbWithExternalNameAnnotation(cdbName), cdbWithConditions(cpv1alpha1.Deleting()), cdbWithStatus(*cdbEmptyObservation(func(p *v1alpha1.CloudantDatabaseObservation) { p.State = "terminating" }))),
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
				Managed: cloudantdatabase(cdbWithExternalNameAnnotation(cdbName)),
			},
			want: want{
				mg:  cloudantdatabase(cdbWithExternalNameAnnotation(cdbName), cdbWithConditions(cpv1alpha1.Deleting())),
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
				Managed: cloudantdatabase(cdbWithExternalNameAnnotation(cdbName)),
			},
			want: want{
				mg:  cloudantdatabase(cdbWithExternalNameAnnotation(cdbName), cdbWithConditions(cpv1alpha1.Deleting())),
				err: errors.Wrap(errors.New(http.StatusText(http.StatusBadRequest)), errDeleteCloudantDatabase),
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
