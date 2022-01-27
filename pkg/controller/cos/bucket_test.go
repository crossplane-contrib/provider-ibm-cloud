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

package cos

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/IBM/ibm-cos-sdk-go/aws/credentials"
	"github.com/IBM/ibm-cos-sdk-go/service/s3"
	cpv1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	"github.com/crossplane-contrib/provider-ibm-cloud/apis/cos/v1alpha1"
	"github.com/crossplane-contrib/provider-ibm-cloud/pkg/controller/tstutil"
)

// Constants that we could not use as such because no address...
var (
	createdAt                 = metav1.Time{Time: time.Date(2009, 11, 17, 20, 34, 58, 651387237, time.UTC)}.Rfc3339Copy()
	aBucketName               = "a-bucket"
	resourceInstanceIDInCloud = "cloudy"
)

// Used in testing, to hold the arguments passed to the crossplane functions
type args struct {
	mg resource.Managed
}

// Applies a list of functions to a bucket observation creted locally
func bucketObservation(m ...func(*v1alpha1.BucketObservation)) *v1alpha1.BucketObservation {
	result := &v1alpha1.BucketObservation{
		CreationDate: &createdAt,
	}

	for _, f := range m {
		f(result)
	}

	return result
}

// Interface to a function that takes as argument a bucket and modifies it
type bucketModifier func(*v1alpha1.Bucket)

// Applies the given external name to a bucket, if it has no annotations
func withBucketExternalNameAnnotation(externalName string) bucketModifier {
	return func(bucket *v1alpha1.Bucket) {
		if bucket.ObjectMeta.Annotations == nil {
			bucket.ObjectMeta.Annotations = make(map[string]string)
		}

		bucket.ObjectMeta.Annotations[meta.AnnotationKeyExternalName] = externalName
	}
}

// Returns a function that sets the ForProvider part of a bucket
func withBucketForProvider(p *v1alpha1.BucketPararams) bucketModifier {
	return func(b *v1alpha1.Bucket) {
		b.Spec.ForProvider = *p
	}
}

// Returns a function that sets the AtProvider part of a bucket
func withBucketAtProvider(p v1alpha1.BucketObservation) bucketModifier {
	return func(b *v1alpha1.Bucket) {
		b.Status.AtProvider = p
	}
}

// Returns a function that sets the bucket condition
func withConditions(c ...cpv1alpha1.Condition) bucketModifier {
	return func(i *v1alpha1.Bucket) {
		i.Status.SetConditions(c...)
	}
}

// Returns params used in tests
func forBucketProvider() *v1alpha1.BucketPararams {
	return &v1alpha1.BucketPararams{
		Name:                 aBucketName,
		IbmServiceInstanceID: &resourceInstanceIDInCloud,
		LocationConstraint:   "does not matter",
	}
}

// Converts a crossplane bucket to an S3 one
func crossplaneToS3(b *v1alpha1.Bucket) *s3.Bucket {
	return &s3.Bucket{
		Name:         &b.Spec.ForProvider.Name,
		CreationDate: &b.Status.AtProvider.DeepCopy().CreationDate.Time,
	}
}

// Creates a crossplane bucket, by creating a generic one + applying a list of modifiers
func createCrossplaneBucket(im ...bucketModifier) *v1alpha1.Bucket {
	i := &v1alpha1.Bucket{
		ObjectMeta: metav1.ObjectMeta{
			Name:       aBucketName,
			Finalizers: []string{},
			Annotations: map[string]string{
				meta.AnnotationKeyExternalName: aBucketName,
			},
		},
		Spec: v1alpha1.BucketSpec{
			ForProvider: v1alpha1.BucketPararams{Name: aBucketName, IbmServiceInstanceID: &resourceInstanceIDInCloud, LocationConstraint: "timbuktu"},
		},
		Status: v1alpha1.BucketStatus{
			AtProvider: v1alpha1.BucketObservation{CreationDate: &createdAt},
		},
	}

	for _, m := range im {
		m(i)
	}

	return i
}

// Converts an array of buckets to XML (imitating the IBM cloud api's response)
//
//
// Params
//    s3BucketArray - the array of buckets
//
// Response
//     an XML string containing only the names and the creation dates of the buckets
func toXML(s3BucketArray []*s3.Bucket) string {
	result := "<?xml version=\"1.0\" encoding=\"UTF-8\" standalone=\"yes\"?><ListAllMyBucketsResult xmlns=\"http://s3.amazonaws.com/doc/2006-03-01/\"><Owner><ID>6967f0d3-dc43-4b9e-beaa-40a79a3128d2</ID><DisplayName>6967f0d3-dc43-4b9e-beaa-40a79a3128d2</DisplayName></Owner><Buckets>"

	for _, b := range s3BucketArray {
		formattedDate := b.CreationDate.Format(time.RFC3339)
		result += "<Bucket><Name>" + *b.Name + "</Name><CreationDate>" + formattedDate + "</CreationDate></Bucket>"
	}

	result += "</Buckets></ListAllMyBucketsResult>"

	return result
}

// Sets up a unit test http server, and creates an external bucket appropriate for unit test.
//
// Params
//	   testingObj - the test object
//	   handlers - the handlers that create the responses
//	   client - the controller runtime client
//
// Returns
//		- the external bucket, ready for unit test
//		- the test http server, on which the caller should call 'defer ....Close()' (reason for this is we need to keep it around to prevent
//		  garbage collection)
//      -- an error (if...)

func setupServerAndGetUnitTestExternalBucket(testingObj *testing.T, handlers *[]tstutil.Handler, kube *client.Client) (*bucketExternal, *httptest.Server, error) {
	mClient, tstServer, err := tstutil.SetupTestServerClient(testingObj, handlers)
	if err != nil || mClient == nil || tstServer == nil {
		return nil, nil, err
	}

	return &bucketExternal{
			unitTestRegionAndCredentials: &unitTestRegionAndCredentials{
				credentials: credentials.AnonymousCredentials,
				region:      "does not matter",
			},
			kube:   *kube,
			client: *mClient,
			logger: logging.NewNopLogger(),
		},
		tstServer,
		nil
}

func TestBucketCreate(t *testing.T) {
	type want struct {
		mg  resource.Managed
		cre managed.ExternalCreation
		err error
	}

	cases := map[string]struct {
		handlers []tstutil.Handler
		kube     client.Client
		args     args
		want     want
	}{
		"Successful": {
			handlers: []tstutil.Handler{
				{
					Path: "/",
					HandlerFunc: func(w http.ResponseWriter, r *http.Request) {
						_ = r.Body.Close()

						if diff := cmp.Diff(http.MethodPut, r.Method); diff != "" {
							t.Errorf("r: -want, +got:\n%s", diff)
						}
						w.Header().Set("Content-Type", "application/json")
						w.WriteHeader(http.StatusCreated)

						s3Bucket := crossplaneToS3(createCrossplaneBucket(withBucketForProvider(forBucketProvider())))
						_ = json.NewEncoder(w).Encode(s3Bucket)
					},
				},
			},
			args: args{
				mg: createCrossplaneBucket(withBucketForProvider(forBucketProvider())),
			},
			want: want{
				mg: createCrossplaneBucket(withBucketForProvider(forBucketProvider()),
					withConditions(cpv1alpha1.Creating()),
					withBucketExternalNameAnnotation(aBucketName)),
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

						if diff := cmp.Diff(http.MethodPut, r.Method); diff != "" {
							t.Errorf("r: -want, +got:\n%s", diff)
						}
						w.Header().Set("Content-Type", "application/json")
						w.WriteHeader(http.StatusBadRequest)

						s3Bucket := crossplaneToS3(createCrossplaneBucket(withBucketForProvider(forBucketProvider())))
						_ = json.NewEncoder(w).Encode(s3Bucket)
					},
				},
			},
			args: args{
				mg: createCrossplaneBucket(withBucketForProvider(forBucketProvider())),
			},
			want: want{
				mg: createCrossplaneBucket(withBucketForProvider(forBucketProvider()),
					withConditions(cpv1alpha1.Creating())),
				cre: managed.ExternalCreation{ExternalNameAssigned: false},
				err: errors.Wrap(errors.New(http.StatusText(http.StatusBadRequest)), errCreateBucket),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e, server, err := setupServerAndGetUnitTestExternalBucket(t, &tc.handlers, &tc.kube)
			if err != nil {
				t.Errorf("Create(...): problem setting up the test server %s", err)
			}

			defer server.Close()

			cre, err := e.Create(context.Background(), tc.args.mg)
			if tc.want.err != nil && err != nil {
				// the case where our mock server returns error, is tricky, as the returned error string is long/spans multiple lines
				expectedNoSpace := strings.ReplaceAll(tc.want.err.Error(), " ", "")
				returnedNoSpace := strings.ReplaceAll(err.Error(), " ", "")
				if strings.HasPrefix(returnedNoSpace, expectedNoSpace) == false {
					diff := cmp.Diff(tc.want.err.Error(), err.Error())
					t.Errorf("Create(...): -want, +got:\n%s", diff)
				}
			} else if diff := cmp.Diff(tc.want.err, err); diff != "" {
				t.Errorf("Create(...): -want, +got:\n%s", diff)
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

func TestBucketDelete(t *testing.T) {
	type want struct {
		mg  resource.Managed
		err error
	}

	cases := map[string]struct {
		handlers []tstutil.Handler
		kube     client.Client
		args     args
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
			args: args{
				mg: createCrossplaneBucket(withBucketAtProvider(*bucketObservation())),
			},
			want: want{
				mg:  createCrossplaneBucket(withBucketAtProvider(*bucketObservation()), withConditions(cpv1alpha1.Deleting())),
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
			args: args{
				mg: createCrossplaneBucket(withBucketAtProvider(*bucketObservation())),
			},
			want: want{
				mg:  createCrossplaneBucket(withBucketAtProvider(*bucketObservation()), withConditions(cpv1alpha1.Deleting())),
				err: nil,
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
			args: args{
				mg: createCrossplaneBucket(withBucketAtProvider(*bucketObservation())),
			},
			want: want{
				mg:  createCrossplaneBucket(withBucketAtProvider(*bucketObservation()), withConditions(cpv1alpha1.Deleting())),
				err: errors.Wrap(errors.New(http.StatusText(http.StatusBadRequest)), errDeleteBucket),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e, server, errCr := setupServerAndGetUnitTestExternalBucket(t, &tc.handlers, &tc.kube)
			if errCr != nil {
				t.Errorf("Delete(...): problem setting up the test server %s", errCr)
			}

			defer server.Close()

			err := e.Delete(context.Background(), tc.args.mg)
			if tc.want.err != nil && err != nil {
				// the case where our mock server returns error, is tricky, as the returned error string is long/spans multiple lines
				expectedNoSpace := strings.ReplaceAll(tc.want.err.Error(), " ", "")
				returnedNoSpace := strings.ReplaceAll(err.Error(), " ", "")
				if strings.HasPrefix(returnedNoSpace, expectedNoSpace) == false {
					diff := cmp.Diff(tc.want.err.Error(), err.Error())
					t.Errorf("Delete(...): -want, +got:\n%s", diff)
				}
			} else if diff := cmp.Diff(tc.want.err, err); diff != "" {
				t.Errorf("Delete(...): -want, +got:\n%s", diff)
			}

			if diff := cmp.Diff(tc.want.mg, tc.args.mg); diff != "" {
				t.Errorf("Delete(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestBucketObserve(t *testing.T) {
	type want struct {
		mg  resource.Managed
		obs managed.ExternalObservation
		err error
	}

	cases := map[string]struct {
		handlers []tstutil.Handler
		kube     client.Client
		args     args
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
			args: args{
				mg: createCrossplaneBucket(withBucketForProvider(forBucketProvider())),
			},
			want: want{
				mg:  createCrossplaneBucket(withBucketForProvider(forBucketProvider())),
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
			args: args{
				mg: createCrossplaneBucket(withBucketForProvider(forBucketProvider())),
			},
			want: want{
				mg:  createCrossplaneBucket(withBucketForProvider(forBucketProvider())),
				obs: managed.ExternalObservation{},
				err: errors.Wrap(errors.New(http.StatusText(http.StatusBadRequest)), errGetBucketFailed),
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
			args: args{
				mg: createCrossplaneBucket(withBucketForProvider(forBucketProvider())),
			},
			want: want{
				mg:  createCrossplaneBucket(withBucketForProvider(forBucketProvider())),
				obs: managed.ExternalObservation{},
				err: errors.Wrap(errors.New(http.StatusText(http.StatusForbidden)), errGetBucketFailed),
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

						crossplaneBucket := createCrossplaneBucket(withBucketAtProvider(*bucketObservation()))
						s3BucketArray := []*s3.Bucket{crossplaneToS3(crossplaneBucket)}
						xmlStr := toXML(s3BucketArray)
						w.Write([]byte(xmlStr))
					},
				},
			},
			args: args{
				mg: createCrossplaneBucket(withBucketAtProvider(*bucketObservation())),
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
			e, server, errCr := setupServerAndGetUnitTestExternalBucket(t, &tc.handlers, &tc.kube)
			if errCr != nil {
				t.Errorf("Delete(...): problem setting up the test server %s", errCr)
			}

			defer server.Close()

			obs, err := e.Observe(context.Background(), tc.args.mg)
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

			if diff := cmp.Diff(tc.want.mg, tc.args.mg); diff != "" {
				t.Errorf("Observe(...): -want, +got:\n%s", diff)
			}
		})
	}
}
