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
	"testing"

	ibmBucketConf "github.com/IBM/ibm-cos-sdk-go-config/resourceconfigurationv1"
	"github.com/go-openapi/strfmt"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/reference"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	"github.com/crossplane-contrib/provider-ibm-cloud/apis/cos/v1alpha1"
	ibmc "github.com/crossplane-contrib/provider-ibm-cloud/pkg/clients"
	cosClient "github.com/crossplane-contrib/provider-ibm-cloud/pkg/clients/cos"
	"github.com/crossplane-contrib/provider-ibm-cloud/pkg/controller/tstutil"
)

// Interface to a function that takes as argument a bucket config and modifies it
type bucketConfigModifier func(*v1alpha1.BucketConfig)

// Applies a list of functions to a bucket observation creted locally
func bucketConfigObservation(m ...func(*v1alpha1.BucketConfigObservation)) *v1alpha1.BucketConfigObservation {
	result := &v1alpha1.BucketConfigObservation{
		CRN:                   "fooooo",
		ServiceInstanceID:     "what an id",
		ServiceInstanceCRN:    "some crn",
		TimeCreated:           *ibmc.DateTimeToMetaV1Time(ibmc.ADateTimeInAYear(1)),
		TimeUpdated:           *ibmc.DateTimeToMetaV1Time(ibmc.ADateTimeInAYear(2)),
		ObjectCount:           33,
		BytesUsed:             31,
		NoncurrentObjectCount: 41,
		NoncurrentBytesUsed:   42,
		DeleteMarkerCount:     43,
	}

	for _, f := range m {
		f(result)
	}

	return result
}

// Returns a fully-populated made-in-the-IBM-cloud bucket configuration
func getIBMBucketConfig() *ibmBucketConf.Bucket {
	result := &ibmBucketConf.Bucket{
		Name:      reference.ToPtrValue(aBucketName),
		HardQuota: ibmc.Int64Ptr(int64(300)),
		Firewall: &ibmBucketConf.Firewall{
			AllowedIp: cosClient.AStrArray(),
		},
		ActivityTracking: &ibmBucketConf.ActivityTracking{
			ReadDataEvents:     ibmc.BoolPtr(true),
			WriteDataEvents:    ibmc.BoolPtr(false),
			ActivityTrackerCrn: reference.ToPtrValue("mama mia"),
		},
		MetricsMonitoring: &ibmBucketConf.MetricsMonitoring{
			UsageMetricsEnabled:   ibmc.BoolPtr(true),
			RequestMetricsEnabled: ibmc.BoolPtr(false),
			MetricsMonitoringCrn:  reference.ToPtrValue("mama due"),
		},
	}

	obs := *bucketConfigObservation()
	result.Crn = &obs.CRN
	result.ServiceInstanceID = &obs.ServiceInstanceID
	result.ServiceInstanceCrn = &obs.ServiceInstanceCRN
	result.TimeCreated = (*strfmt.DateTime)(&obs.TimeCreated.Time)
	result.TimeUpdated = (*strfmt.DateTime)(&obs.TimeUpdated.Time)
	result.ObjectCount = &obs.ObjectCount
	result.BytesUsed = &obs.BytesUsed
	result.NoncurrentObjectCount = &obs.NoncurrentObjectCount
	result.NoncurrentBytesUsed = &obs.NoncurrentBytesUsed
	result.DeleteMarkerCount = &obs.DeleteMarkerCount

	return result
}

// Returns params used in tests
func forBucketConfigProvider() *v1alpha1.BucketConfigParams {
	result, _ := cosClient.GenerateBucketConfigFromServerParams(getIBMBucketConfig())

	return result
}

// Returns a function that sets the ForProvider part of a bucket config
func withBucketConfigForProvider(p *v1alpha1.BucketConfigParams) bucketConfigModifier {
	return func(b *v1alpha1.BucketConfig) {
		b.Spec.ForProvider = *p
	}
}

// Creates a crossplane bucket config, by creating a generic one + applying a list of modifiers
func createCrossplaneBucketConfig(im ...bucketConfigModifier) *v1alpha1.BucketConfig {
	i := &v1alpha1.BucketConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:       aBucketName,
			Finalizers: []string{},
			Annotations: map[string]string{
				meta.AnnotationKeyExternalName: aBucketName,
			},
		},
		Spec: v1alpha1.BucketConfigSpec{
			ForProvider: v1alpha1.BucketConfigParams{Name: reference.ToPtrValue(aBucketName)},
		},
		Status: v1alpha1.BucketConfigStatus{
			AtProvider: *bucketConfigObservation(),
		},
	}

	for _, m := range im {
		m(i)
	}

	return i
}

// Sets up a unit test http server, and creates an external bucket config structure appropriate for unit test.
//
// Params
//	   testingObj - the test object
//	   handlers - the handlers that create the responses
//	   client - the controller runtime client
//
// Returns
//		- the external bucket config, ready for unit test
//		- the test http server, on which the caller should call 'defer ....Close()' (reason for this is we need to keep it around to prevent
//		  garbage collection)
//      - an error (iff...)
func setupServerAndGetUnitTestExternalBucketConfig(testingObj *testing.T, handlers *[]tstutil.Handler, kube *client.Client) (*bucketConfigExternal, *httptest.Server, error) {
	mClient, tstServer, err := tstutil.SetupTestServerClient(testingObj, handlers)
	if err != nil {
		return nil, nil, err
	}

	return &bucketConfigExternal{
			kube:   *kube,
			client: *mClient,
			logger: logging.NewNopLogger(),
		},
		tstServer,
		nil
}

func TestBucketConfigObserve(t *testing.T) {
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
		name     string // used for debugging convenience
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
				Managed: createCrossplaneBucketConfig(withBucketConfigForProvider(forBucketConfigProvider())),
			},
			want: want{
				err: errors.Wrap(errors.New(http.StatusText(http.StatusNotFound)), errGetBucketConfigFailed),
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

						// content type should always set before writeHeader()
						w.Header().Set("Content-Type", "application/json")
						w.WriteHeader(http.StatusBadRequest)
					},
				},
			},
			args: tstutil.Args{
				Managed: createCrossplaneBucketConfig(withBucketConfigForProvider(forBucketConfigProvider())),
			},
			want: want{
				err: errors.Wrap(errors.New(http.StatusText(http.StatusBadRequest)), errGetBucketConfigFailed),
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

						// content type should always set before writeHeader()
						w.Header().Set("Content-Type", "application/json")
						w.WriteHeader(http.StatusForbidden)
					},
				},
			},
			args: tstutil.Args{
				Managed: createCrossplaneBucketConfig(withBucketConfigForProvider(forBucketConfigProvider())),
			},
			want: want{
				err: errors.Wrap(errors.New(http.StatusText(http.StatusForbidden)), errGetBucketConfigFailed),
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

						// content type should always set before writeHeader()
						w.Header().Set("Content-Type", "application/json")
						w.WriteHeader(http.StatusOK)

						bc := getIBMBucketConfig()
						_ = json.NewEncoder(w).Encode(bc)
					},
				},
			},
			args: tstutil.Args{
				Managed: createCrossplaneBucketConfig(withBucketConfigForProvider(forBucketConfigProvider())),
			},
			want: want{
				obs: managed.ExternalObservation{
					ResourceExists:    true,
					ResourceUpToDate:  true,
					ConnectionDetails: nil,
				},
			},
		},
	}

	for name, tc := range cases {
		tc.name = name
		t.Run(name, func(t *testing.T) {
			e, server, errCr := setupServerAndGetUnitTestExternalBucketConfig(t, &tc.handlers, &tc.kube)
			if errCr != nil {
				t.Errorf("Delete(...): problem setting up the test server %s", errCr)
			}

			defer server.Close()

			obs, err := e.Observe(context.Background(), tc.args.Managed)
			if tc.want.err != nil && err != nil {
				if diff := cmp.Diff(tc.want.err.Error(), err.Error()); diff != "" {
					t.Errorf("Observe(...): want error string != got error string:\n%s", diff)
				}
			} else if diff := cmp.Diff(tc.want.err, err); diff != "" {
				t.Errorf("Observe(...): want error != got error:\n%s", diff)
			}

			if diff := cmp.Diff(tc.want.obs, obs); diff != "" {
				t.Errorf("Observe(...): -want, +got:\n%s", diff)
			}

			if tc.want.mg != nil {
				if diff := cmp.Diff(tc.want.mg, tc.args.Managed); diff != "" {
					t.Errorf("Observe(...): -want, +got:\n%s", diff)
				}
			}
		})
	}
}

func TestBucketConfigUpdate(t *testing.T) {
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
		name     string // used for debugging convenience
	}{
		"Successful": {
			handlers: []tstutil.Handler{
				{
					Path: "/",
					HandlerFunc: func(w http.ResponseWriter, r *http.Request) {
						_ = r.Body.Close()

						if diff := cmp.Diff(http.MethodPatch, r.Method); diff != "" {
							t.Errorf("r: -want, +got:\n%s", diff)
						}

						// content type should always set before writeHeader()
						w.Header().Set("Content-Type", "application/json")
						w.WriteHeader(http.StatusOK)
					},
				},
			},
			args: tstutil.Args{
				Managed: createCrossplaneBucketConfig(withBucketConfigForProvider(forBucketConfigProvider())),
			},
			want: want{
				upd: managed.ExternalUpdate{},
				err: nil,
			},
		},
		"Failed": {
			handlers: []tstutil.Handler{
				{
					Path: "/",
					HandlerFunc: func(w http.ResponseWriter, r *http.Request) {
						_ = r.Body.Close()

						if diff := cmp.Diff(http.MethodPatch, r.Method); diff != "" {
							t.Errorf("r: -want, +got:\n%s", diff)
						}

						// content type should always set before writeHeader()
						w.Header().Set("Content-Type", "application/json")
						w.WriteHeader(http.StatusBadRequest)
					},
				},
			},
			args: tstutil.Args{
				Managed: createCrossplaneBucketConfig(withBucketConfigForProvider(forBucketConfigProvider())),
			},
			want: want{
				upd: managed.ExternalUpdate{},
				err: errors.Wrap(errors.New(http.StatusText(http.StatusBadRequest)), errUpdBucketConfig),
			},
		},
	}

	for name, tc := range cases {
		tc.name = name
		t.Run(name, func(t *testing.T) {
			e, server, errCr := setupServerAndGetUnitTestExternalBucketConfig(t, &tc.handlers, &tc.kube)
			if errCr != nil {
				t.Errorf("Delete(...): problem setting up the test server %s", errCr)
			}

			defer server.Close()

			xu, err := e.Update(context.Background(), tc.args.Managed)
			if tc.want.err != nil && err != nil {
				if diff := cmp.Diff(tc.want.err.Error(), err.Error()); diff != "" {
					t.Errorf("Observe(...): want error string != got error string:\n%s", diff)
				}
			} else if diff := cmp.Diff(tc.want.err, err); diff != "" {
				t.Errorf("Observe(...): want error != got error:\n%s", diff)
			}

			if diff := cmp.Diff(tc.want.upd, xu); diff != "" {
				t.Errorf("Observe(...): -want, +got:\n%s", diff)
			}

			if tc.want.mg != nil {
				if diff := cmp.Diff(tc.want.mg, tc.args.Managed); diff != "" {
					t.Errorf("Observe(...): -want, +got:\n%s", diff)
				}
			}
		})
	}
}
