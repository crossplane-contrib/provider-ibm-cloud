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
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/go-cmp/cmp"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	cpv1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/reference"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	cv1 "github.com/IBM/cloudant-go-sdk/cloudantv1"
	"github.com/IBM/go-sdk-core/core"

	"github.com/crossplane-contrib/provider-ibm-cloud/apis/cloudantv1/v1alpha1"
	ibmc "github.com/crossplane-contrib/provider-ibm-cloud/pkg/clients"
)

const (
	bearerTok = "mock-token"
)

var (
	cdbName = "myCloudantDatabase"
)

var _ managed.ExternalConnecter = &cloudantdatabaseConnector{}
var _ managed.ExternalClient = &cloudantdatabaseExternal{}

type handler struct {
	path        string
	handlerFunc func(w http.ResponseWriter, r *http.Request)
}

type cdbModifier func(*v1alpha1.CloudantDatabase)

func cloudantdatabase(im ...cdbModifier) *v1alpha1.CloudantDatabase {
	i := &v1alpha1.CloudantDatabase{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "myCloudantDatabase",
			Finalizers: []string{},
			Annotations: map[string]string{
				meta.AnnotationKeyExternalName: "myCloudantDatabase",
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
		Db:          "myDatabase",
		Partitioned: ibmc.BoolPtr(false),
		Q:           ibmc.Int64Ptr(int64(2)),
	}
	for _, f := range m {
		f(p)
	}
	return p
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
		// not including props for now because I don't think it should be in the API ??
		// Props: in.Props,
		Sizes:     generateTestv1alpha1ContentInformationSizes(),
		UpdateSeq: "myUpdateSeq",
		UUID:      "myUUID",
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
		DbName:             reference.ToPtrValue("myDatabase"),
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

func TestCloudantDatabaseObserve(t *testing.T) {
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
		// "NotFound": {
		// 	handlers: []handler{
		// 		{
		// 			path: "/",
		// 			handlerFunc: func(w http.ResponseWriter, r *http.Request) {
		// 				_ = r.Body.Close()
		// 				if diff := cmp.Diff(http.MethodGet, r.Method); diff != "" {
		// 					t.Errorf("r: -want, +got:\n%s", diff)
		// 				}
		// 				// content type should always set before writeHeader()
		// 				w.Header().Set("Content-Type", "application/json")
		// 				w.WriteHeader(http.StatusNotFound)
		// 				_ = json.NewEncoder(w).Encode(&cv1.DatabaseInformation{})
		// 			},
		// 		},
		// 	},
		// 	args: args{
		// 		mg: cloudantdatabase(),
		// 	},
		// 	want: want{
		// 		mg:  cloudantdatabase(),
		// 		err: nil,
		// 	},
		// },
		// "GetFailed": {
		// 	handlers: []handler{
		// 		{
		// 			path: "/",
		// 			handlerFunc: func(w http.ResponseWriter, r *http.Request) {
		// 				_ = r.Body.Close()
		// 				if diff := cmp.Diff(http.MethodGet, r.Method); diff != "" {
		// 					t.Errorf("r: -want, +got:\n%s", diff)
		// 				}
		// 				w.Header().Set("Content-Type", "application/json")
		// 				w.WriteHeader(http.StatusBadRequest)
		// 				_ = json.NewEncoder(w).Encode(&cv1.DatabaseInformation{})
		// 			},
		// 		},
		// 	},
		// 	args: args{
		// 		mg: cloudantdatabase(),
		// 	},
		// 	want: want{
		// 		mg:  cloudantdatabase(),
		// 		err: errors.New(errGetCloudantDatabaseFailed + ": Bad Request"),
		// 	},
		// },
		// "ObservedCloudantDatabaseUpToDate": {
		// 	handlers: []handler{
		// 		{
		// 			path: "/",
		// 			handlerFunc: func(w http.ResponseWriter, r *http.Request) {
		// 				_ = r.Body.Close()
		// 				if diff := cmp.Diff(http.MethodGet, r.Method); diff != "" {
		// 					t.Errorf("r: -want, +got:\n%s", diff)
		// 				}
		// 				w.Header().Set("Content-Type", "application/json")
		// 				cdb := cdbInstance()
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
		// 		),
		// 	},
		// 	want: want{
		// 		mg: cloudantdatabase(cdbWithSpec(*cdbParams()),
		// 			cdbWithConditions(cpv1alpha1.Available()),
		// 			cdbWithStatus(*cdbObservation()),
		// 		),
		// 		obs: managed.ExternalObservation{
		// 			ResourceExists:    true,
		// 			ResourceUpToDate:  true,
		// 			ConnectionDetails: nil,
		// 		},
		// 	},
		// },
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
			e := cloudantdatabaseExternal{
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
