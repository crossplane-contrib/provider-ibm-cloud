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
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"

	cpv1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/crossplane/crossplane-runtime/pkg/test"

	icdv5 "github.com/IBM/experimental-go-sdk/ibmclouddatabasesv5"

	"github.com/crossplane-contrib/provider-ibm-cloud/apis/ibmclouddatabasesv5/v1alpha1"
	ibmc "github.com/crossplane-contrib/provider-ibm-cloud/pkg/clients"
	"github.com/crossplane-contrib/provider-ibm-cloud/pkg/controller/tstutil"
)

var (
	asgName                                    = "postgres-asg"
	diskScalerCapacityEnabled                  = true
	diskScalerCapacityFreeSpaceLessThanPercent = 10
	diskScalerIoUtilizationEnabled             = true
	diskScalerIoUtilizationOverPeriod          = "30m"
	diskScalerIoUtilizationAbovePercent        = 45
	diskRateIncreasePercent                    = 20
	diskRatePeriodSeconds                      = 900
	diskRateLimitMbPerMember                   = 3670016
	diskRateUnits                              = "mb"
	memoryScalerIoUtilizationEnabled           = true
	memoryScalerIoUtilizationOverPeriod        = "5m"
	memoryScalerIoUtilizationAbovePercent      = 90
	memoryRateIncreasePercent                  = 10
	memoryRatePeriodSeconds                    = 300
	memoryRateLimitMbPerMember                 = 125952
	memoryRateUnits                            = "mb"
	cpuRateIncreasePercent                     = 15
	cpuRateIncreasePercent2                    = 20
	cpuRatePeriodSeconds                       = 800
	cpuRateLimitCountPerMember                 = 20
	cpuRateUnits                               = "count"
)

var _ managed.ExternalConnecter = &asgConnector{}
var _ managed.ExternalClient = &asgExternal{}

type asgModifier func(*v1alpha1.AutoscalingGroup)

func asg(im ...asgModifier) *v1alpha1.AutoscalingGroup {
	i := &v1alpha1.AutoscalingGroup{
		ObjectMeta: metav1.ObjectMeta{
			Name:       asgName,
			Finalizers: []string{},
			Annotations: map[string]string{
				meta.AnnotationKeyExternalName: id,
			},
		},
		Spec: v1alpha1.AutoscalingGroupSpec{
			ForProvider: v1alpha1.AutoscalingGroupParameters{},
		},
	}
	for _, m := range im {
		m(i)
	}
	return i
}

func asgWithExternalNameAnnotation(externalName string) asgModifier {
	return func(i *v1alpha1.AutoscalingGroup) {
		if i.ObjectMeta.Annotations == nil {
			i.ObjectMeta.Annotations = make(map[string]string)
		}
		i.ObjectMeta.Annotations[wtfConst] = externalName
	}
}

func asgWithSpec(p v1alpha1.AutoscalingGroupParameters) asgModifier {
	return func(r *v1alpha1.AutoscalingGroup) { r.Spec.ForProvider = p }
}

func asgWithConditions(c ...cpv1alpha1.Condition) asgModifier {
	return func(i *v1alpha1.AutoscalingGroup) { i.Status.SetConditions(c...) }
}

func asgWithStatus(p v1alpha1.AutoscalingGroupObservation) asgModifier {
	return func(r *v1alpha1.AutoscalingGroup) { r.Status.AtProvider = p }
}

func asgParams(m ...func(*v1alpha1.AutoscalingGroupParameters)) *v1alpha1.AutoscalingGroupParameters {
	p := &v1alpha1.AutoscalingGroupParameters{
		ID: &id,
		Disk: &v1alpha1.AutoscalingDiskGroupDisk{
			Scalers: &v1alpha1.AutoscalingDiskGroupDiskScalers{
				Capacity: &v1alpha1.AutoscalingDiskGroupDiskScalersCapacity{
					Enabled:                  &diskScalerCapacityEnabled,
					FreeSpaceLessThanPercent: ibmc.Int64Ptr(int64(diskScalerCapacityFreeSpaceLessThanPercent)),
				},
				IoUtilization: &v1alpha1.AutoscalingDiskGroupDiskScalersIoUtilization{
					Enabled:      &diskScalerIoUtilizationEnabled,
					OverPeriod:   &diskScalerIoUtilizationOverPeriod,
					AbovePercent: ibmc.Int64Ptr(int64(diskScalerIoUtilizationAbovePercent)),
				},
			},
			Rate: &v1alpha1.AutoscalingDiskGroupDiskRate{
				IncreasePercent:  ibmc.Int64Ptr(int64(diskRateIncreasePercent)),
				PeriodSeconds:    ibmc.Int64Ptr(int64(diskRatePeriodSeconds)),
				LimitMbPerMember: ibmc.Int64Ptr(int64(diskRateLimitMbPerMember)),
				Units:            &diskRateUnits,
			},
		},
		Memory: &v1alpha1.AutoscalingMemoryGroupMemory{
			Scalers: &v1alpha1.AutoscalingMemoryGroupMemoryScalers{
				IoUtilization: &v1alpha1.AutoscalingMemoryGroupMemoryScalersIoUtilization{
					Enabled:      &memoryScalerIoUtilizationEnabled,
					OverPeriod:   &memoryScalerIoUtilizationOverPeriod,
					AbovePercent: ibmc.Int64Ptr(int64(memoryScalerIoUtilizationAbovePercent)),
				},
			},
			Rate: &v1alpha1.AutoscalingMemoryGroupMemoryRate{
				IncreasePercent:  ibmc.Int64Ptr(int64(memoryRateIncreasePercent)),
				PeriodSeconds:    ibmc.Int64Ptr(int64(memoryRatePeriodSeconds)),
				LimitMbPerMember: ibmc.Int64Ptr(int64(memoryRateLimitMbPerMember)),
				Units:            &memoryRateUnits,
			},
		},
		CPU: &v1alpha1.AutoscalingCPUGroupCPU{
			Rate: &v1alpha1.AutoscalingCPUGroupCPURate{
				IncreasePercent:     ibmc.Int64Ptr(int64(cpuRateIncreasePercent)),
				PeriodSeconds:       ibmc.Int64Ptr(int64(cpuRatePeriodSeconds)),
				LimitCountPerMember: ibmc.Int64Ptr(int64(cpuRateLimitCountPerMember)),
				Units:               &cpuRateUnits,
			},
		},
	}
	for _, f := range m {
		f(p)
	}
	return p
}

func asgObservation(m ...func(*v1alpha1.AutoscalingGroupObservation)) *v1alpha1.AutoscalingGroupObservation {
	o := &v1alpha1.AutoscalingGroupObservation{
		State: string(cpv1alpha1.Available().Reason),
	}

	for _, f := range m {
		f(o)
	}
	return o
}

func asgInstance(m ...func(*icdv5.AutoscalingGroup)) *icdv5.AutoscalingGroup {
	i := &icdv5.AutoscalingGroup{
		Disk: &icdv5.AutoscalingDiskGroupDisk{
			Scalers: &icdv5.AutoscalingDiskGroupDiskScalers{
				Capacity: &icdv5.AutoscalingDiskGroupDiskScalersCapacity{
					Enabled:                  &diskScalerCapacityEnabled,
					FreeSpaceLessThanPercent: ibmc.Int64Ptr(int64(diskScalerCapacityFreeSpaceLessThanPercent)),
				},
				IoUtilization: &icdv5.AutoscalingDiskGroupDiskScalersIoUtilization{
					Enabled:      &diskScalerIoUtilizationEnabled,
					OverPeriod:   &diskScalerIoUtilizationOverPeriod,
					AbovePercent: ibmc.Int64Ptr(int64(diskScalerIoUtilizationAbovePercent)),
				},
			},
			Rate: &icdv5.AutoscalingDiskGroupDiskRate{
				IncreasePercent:  ibmc.Int64PtrToFloat64Ptr(ibmc.Int64Ptr(int64(diskRateIncreasePercent))),
				PeriodSeconds:    ibmc.Int64Ptr(int64(diskRatePeriodSeconds)),
				LimitMbPerMember: ibmc.Int64PtrToFloat64Ptr(ibmc.Int64Ptr(int64(diskRateLimitMbPerMember))),
				Units:            &diskRateUnits,
			},
		},
		Memory: &icdv5.AutoscalingMemoryGroupMemory{
			Scalers: &icdv5.AutoscalingMemoryGroupMemoryScalers{
				IoUtilization: &icdv5.AutoscalingMemoryGroupMemoryScalersIoUtilization{
					Enabled:      &memoryScalerIoUtilizationEnabled,
					OverPeriod:   &memoryScalerIoUtilizationOverPeriod,
					AbovePercent: ibmc.Int64Ptr(int64(memoryScalerIoUtilizationAbovePercent)),
				},
			},
			Rate: &icdv5.AutoscalingMemoryGroupMemoryRate{
				IncreasePercent:  ibmc.Int64PtrToFloat64Ptr(ibmc.Int64Ptr(int64(memoryRateIncreasePercent))),
				PeriodSeconds:    ibmc.Int64Ptr(int64(memoryRatePeriodSeconds)),
				LimitMbPerMember: ibmc.Int64PtrToFloat64Ptr(ibmc.Int64Ptr(int64(memoryRateLimitMbPerMember))),
				Units:            &memoryRateUnits,
			},
		},
		Cpu: &icdv5.AutoscalingCPUGroupCPU{
			Rate: &icdv5.AutoscalingCPUGroupCPURate{
				IncreasePercent:     ibmc.Int64PtrToFloat64Ptr(ibmc.Int64Ptr(int64(cpuRateIncreasePercent))),
				PeriodSeconds:       ibmc.Int64Ptr(int64(cpuRatePeriodSeconds)),
				LimitCountPerMember: ibmc.Int64Ptr(int64(cpuRateLimitCountPerMember)),
				Units:               &cpuRateUnits,
			},
		},
	}

	for _, f := range m {
		f(i)
	}
	return i
}

// Sets up a unit test http server, and creates an external autoscaling group structure appropriate for unit test.
//
// Params
//
//	testingObj - the test object
//	handlers - the handlers that create the responses
//	client - the controller runtime client
//
// Returns
//   - the external object, ready for unit test
//   - the test http server, on which the caller should call 'defer ....Close()' (reason for this is we need to keep it around to prevent
//     garbage collection)
//     -- an error (if...)
func setupServerAndGetUnitTestExternalASG(testingObj *testing.T, handlers *[]tstutil.Handler, kube *client.Client) (*asgExternal, *httptest.Server, error) {
	mClient, tstServer, err := tstutil.SetupTestServerClient(testingObj, handlers)
	if err != nil || mClient == nil || tstServer == nil {
		return nil, nil, err
	}

	return &asgExternal{
			kube:   *kube,
			client: *mClient,
			logger: logging.NewNopLogger(),
		},
		tstServer,
		nil
}

func TestAutoscalingGroupObserve(t *testing.T) {
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
						err := json.NewEncoder(w).Encode(&icdv5.GetAutoscalingConditionsResponse{})
						if err != nil {
							klog.Errorf("%s", err)
						}
					},
				},
			},
			args: tstutil.Args{
				Managed: asg(),
			},
			want: want{
				mg:  asg(),
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
						err := json.NewEncoder(w).Encode(&icdv5.GetAutoscalingConditionsResponse{})
						if err != nil {
							klog.Errorf("%s", err)
						}
					},
				},
			},
			args: tstutil.Args{
				Managed: asg(),
			},
			want: want{
				mg:  asg(),
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
						err := json.NewEncoder(w).Encode(&icdv5.GetAutoscalingConditionsResponse{})
						if err != nil {
							klog.Errorf("%s", err)
						}
					},
				},
			},
			args: tstutil.Args{
				Managed: asg(),
			},
			want: want{
				mg:  asg(),
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
						err := json.NewEncoder(w).Encode(&icdv5.GetAutoscalingConditionsResponse{Autoscaling: asgInstance()})
						if err != nil {
							klog.Errorf("%s", err)
						}
					},
				},
			},
			kube: &test.MockClient{
				MockUpdate: test.NewMockUpdateFn(nil),
			},
			args: tstutil.Args{
				Managed: asg(
					asgWithExternalNameAnnotation(id),
					asgWithSpec(*asgParams()),
				),
			},
			want: want{
				mg: asg(asgWithSpec(*asgParams()),
					asgWithConditions(cpv1alpha1.Available()),
					asgWithStatus(*asgObservation())),
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
						asg := &icdv5.GetAutoscalingConditionsResponse{Autoscaling: asgInstance(
							func(ag *icdv5.AutoscalingGroup) {
								ag.Cpu.Rate.IncreasePercent = ibmc.Int64PtrToFloat64Ptr(ibmc.Int64Ptr(int64(cpuRateIncreasePercent2)))
							},
						)}
						err := json.NewEncoder(w).Encode(asg)
						if err != nil {
							klog.Errorf("%s", err)
						}
					},
				},
			},
			kube: &test.MockClient{
				MockUpdate: test.NewMockUpdateFn(nil),
			},
			args: tstutil.Args{
				Managed: asg(
					asgWithExternalNameAnnotation(id),
					asgWithSpec(*asgParams()),
				),
			},
			want: want{
				mg: asg(asgWithSpec(*asgParams()),
					asgWithConditions(cpv1alpha1.Available()),
					asgWithStatus(*asgObservation(func(p *v1alpha1.AutoscalingGroupObservation) {
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
			e, server, setupErr := setupServerAndGetUnitTestExternalASG(t, &tc.handlers, &tc.kube)
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

func TestAutoscalingGroupCreate(t *testing.T) {
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
						resp := icdv5.SetAutoscalingConditionsResponse{Task: &icdv5.Task{}}
						err := json.NewEncoder(w).Encode(resp)
						if err != nil {
							klog.Errorf("%s", err)
						}
					},
				},
			},
			args: tstutil.Args{
				Managed: asg(asgWithSpec(*asgParams())),
			},
			want: want{
				mg: asg(asgWithSpec(*asgParams()),
					asgWithConditions(cpv1alpha1.Creating()),
					asgWithExternalNameAnnotation(id)),
				cre: managed.ExternalCreation{ExternalNameAssigned: true},
				err: nil,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e, server, setupErr := setupServerAndGetUnitTestExternalASG(t, &tc.handlers, &tc.kube)
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

func TestAutoscalingGroupDelete(t *testing.T) {
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
				Managed: asg(asgWithStatus(*asgObservation())),
			},
			want: want{
				mg:  asg(asgWithStatus(*asgObservation()), asgWithConditions(cpv1alpha1.Deleting())),
				err: nil,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e, server, setupErr := setupServerAndGetUnitTestExternalASG(t, &tc.handlers, &tc.kube)
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

func TestAutoscalingGroupUpdate(t *testing.T) {
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
						if diff := cmp.Diff(http.MethodPatch, r.Method); diff != "" {
							t.Errorf("r: -want, +got:\n%s", diff)
						}
						w.Header().Set("Content-Type", "application/json")
						w.WriteHeader(http.StatusOK)
						_ = r.Body.Close()
						resp := icdv5.SetAutoscalingConditionsResponse{Task: &icdv5.Task{}}
						err := json.NewEncoder(w).Encode(resp)
						if err != nil {
							klog.Errorf("%s", err)
						}
					},
				},
			},
			args: tstutil.Args{
				Managed: asg(asgWithSpec(*asgParams()), asgWithStatus(*asgObservation())),
			},
			want: want{
				mg:  asg(asgWithSpec(*asgParams()), asgWithStatus(*asgObservation())),
				upd: managed.ExternalUpdate{},
				err: nil,
			},
		},
		"PatchFails": {
			handlers: []tstutil.Handler{
				{
					Path: "/",
					HandlerFunc: func(w http.ResponseWriter, r *http.Request) {
						if diff := cmp.Diff(http.MethodPatch, r.Method); diff != "" {
							t.Errorf("r: -want, +got:\n%s", diff)
						}
						w.Header().Set("Content-Type", "application/json")
						w.WriteHeader(http.StatusBadRequest)
						_ = r.Body.Close()
					},
				},
			},
			args: tstutil.Args{
				Managed: asg(asgWithSpec(*asgParams()), asgWithStatus(*asgObservation())),
			},
			want: want{
				mg:  asg(asgWithSpec(*asgParams()), asgWithStatus(*asgObservation())),
				err: errors.New(http.StatusText(http.StatusBadRequest)),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e, server, setupErr := setupServerAndGetUnitTestExternalASG(t, &tc.handlers, &tc.kube)
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
