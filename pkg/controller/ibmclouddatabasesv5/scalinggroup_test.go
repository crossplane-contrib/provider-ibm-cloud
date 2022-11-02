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

const (
	errBadRequest = "error getting instance: Bad Request"
	errForbidden  = "error getting instance: Forbidden"
	wtfConst      = "crossplane.io/external-name"
)

var (
	sgName                 = "postgres-sg"
	id                     = "crn:v1:bluemix:public:databases-for-postgresql:us-south:a/0b5a00334eaf9eb9339d2ab48f20d7f5:dda29288-c259-4dc9-859c-154eb7939cd0::"
	membersUnits           = "count"
	membersAllocationCount = 2
	membersMinimumCount    = 2
	membersMaximumCount    = 20
	membersStepSizeCount   = 1
	membersIsAdjustable    = true
	membersIsOptional      = false
	membersCanScaleDown    = false
	memoryUnits            = "mb"
	memoryAllocationMb     = 25600
	memoryMinimumMb        = 2048
	memoryMaximumMb        = 229376
	memoryStepSizeMb       = 256
	memoryIsAdjustable     = true
	memoryIsOptional       = false
	memoryCanScaleDown     = true
	cpuUnits               = "count"
	cpuAllocationCount     = 6
	cpuMinimumCount        = 6
	cpuMaximumCount        = 56
	cpuStepSizeCount       = 2
	cpuIsAdjustable        = true
	cpuIsOptional          = true
	cpuCanScaleDown        = true
	diskUnits              = "mb"
	diskAllocationMb       = 35840
	diskMinimumMb          = 35840
	diskMaximumMb          = 7340032
	diskStepSizeMb         = 1024
	diskIsAdjustable       = true
	diskIsOptional         = false
	diskCanScaleDown       = false
)

var _ managed.ExternalConnecter = &sgConnector{}
var _ managed.ExternalClient = &sgExternal{}

type sgModifier func(*v1alpha1.ScalingGroup)

func sg(im ...sgModifier) *v1alpha1.ScalingGroup {
	i := &v1alpha1.ScalingGroup{
		ObjectMeta: metav1.ObjectMeta{
			Name:       sgName,
			Finalizers: []string{},
			Annotations: map[string]string{
				meta.AnnotationKeyExternalName: id,
			},
		},
		Spec: v1alpha1.ScalingGroupSpec{
			ForProvider: v1alpha1.ScalingGroupParameters{},
		},
	}
	for _, m := range im {
		m(i)
	}
	return i
}

func sgWithExternalNameAnnotation(externalName string) sgModifier {
	return func(i *v1alpha1.ScalingGroup) {
		if i.ObjectMeta.Annotations == nil {
			i.ObjectMeta.Annotations = make(map[string]string)
		}
		i.ObjectMeta.Annotations[wtfConst] = externalName
	}
}

func sgWithSpec(p v1alpha1.ScalingGroupParameters) sgModifier {
	return func(r *v1alpha1.ScalingGroup) { r.Spec.ForProvider = p }
}

func sgWithConditions(c ...cpv1alpha1.Condition) sgModifier {
	return func(i *v1alpha1.ScalingGroup) { i.Status.SetConditions(c...) }
}

func sgWithStatus(p v1alpha1.ScalingGroupObservation) sgModifier {
	return func(r *v1alpha1.ScalingGroup) { r.Status.AtProvider = p }
}

func params(m ...func(*v1alpha1.ScalingGroupParameters)) *v1alpha1.ScalingGroupParameters {
	p := &v1alpha1.ScalingGroupParameters{
		ID: &id,
		Members: &v1alpha1.SetMembersGroupMembers{
			AllocationCount: int64(membersAllocationCount),
		},
		MemberMemory: &v1alpha1.SetMemoryGroupMemory{
			AllocationMb: int64(memoryAllocationMb / membersAllocationCount),
		},
		MemberDisk: &v1alpha1.SetDiskGroupDisk{
			AllocationMb: int64(diskAllocationMb / membersAllocationCount),
		},
		MemberCPU: &v1alpha1.SetCPUGroupCPU{
			AllocationCount: int64(cpuAllocationCount / membersAllocationCount),
		},
	}
	for _, f := range m {
		f(p)
	}
	return p
}

func observation(m ...func(*v1alpha1.ScalingGroupObservation)) *v1alpha1.ScalingGroupObservation {
	o := &v1alpha1.ScalingGroupObservation{
		State: string(cpv1alpha1.Available().Reason),
		Groups: []v1alpha1.Group{
			{
				ID:    id,
				Count: int64(membersAllocationCount),
				Members: v1alpha1.GroupMembers{
					AllocationCount: int64(membersAllocationCount),
					Units:           &membersUnits,
					MinimumCount:    ibmc.Int64Ptr(int64(membersMinimumCount)),
					MaximumCount:    ibmc.Int64Ptr(int64(membersMaximumCount)),
					StepSizeCount:   ibmc.Int64Ptr(int64(membersStepSizeCount)),
					IsAdjustable:    ibmc.BoolPtr(membersIsAdjustable),
					IsOptional:      ibmc.BoolPtr(membersIsOptional),
					CanScaleDown:    ibmc.BoolPtr(membersCanScaleDown),
				},
				Memory: v1alpha1.GroupMemory{
					AllocationMb:       int64(memoryAllocationMb),
					MemberAllocationMb: int64(memoryAllocationMb / membersAllocationCount),
					Units:              &memoryUnits,
					MinimumMb:          ibmc.Int64Ptr(int64(memoryMinimumMb)),
					MaximumMb:          ibmc.Int64Ptr(int64(memoryMaximumMb)),
					StepSizeMb:         ibmc.Int64Ptr(int64(memoryStepSizeMb)),
					IsAdjustable:       ibmc.BoolPtr(memoryIsAdjustable),
					IsOptional:         ibmc.BoolPtr(memoryIsOptional),
					CanScaleDown:       ibmc.BoolPtr(memoryCanScaleDown),
				},
				Disk: v1alpha1.GroupDisk{
					AllocationMb:       int64(diskAllocationMb),
					MemberAllocationMb: int64(diskAllocationMb / membersAllocationCount),
					Units:              &diskUnits,
					MinimumMb:          ibmc.Int64Ptr(int64(diskMinimumMb)),
					MaximumMb:          ibmc.Int64Ptr(int64(diskMaximumMb)),
					StepSizeMb:         ibmc.Int64Ptr(int64(diskStepSizeMb)),
					IsAdjustable:       ibmc.BoolPtr(diskIsAdjustable),
					IsOptional:         ibmc.BoolPtr(diskIsOptional),
					CanScaleDown:       ibmc.BoolPtr(diskCanScaleDown),
				},
				CPU: v1alpha1.GroupCPU{
					AllocationCount:       int64(cpuAllocationCount),
					MemberAllocationCount: int64(cpuAllocationCount / membersAllocationCount),
					Units:                 &cpuUnits,
					MinimumCount:          ibmc.Int64Ptr(int64(cpuMinimumCount)),
					MaximumCount:          ibmc.Int64Ptr(int64(cpuMaximumCount)),
					StepSizeCount:         ibmc.Int64Ptr(int64(cpuStepSizeCount)),
					IsAdjustable:          ibmc.BoolPtr(cpuIsAdjustable),
					IsOptional:            ibmc.BoolPtr(cpuIsOptional),
					CanScaleDown:          ibmc.BoolPtr(cpuCanScaleDown),
				},
			},
		},
	}

	for _, f := range m {
		f(o)
	}
	return o
}

func instance(m ...func(*icdv5.Groups)) *icdv5.Groups {
	i := &icdv5.Groups{
		Groups: []icdv5.Group{
			{
				ID:    &id,
				Count: ibmc.Int64Ptr(int64(membersAllocationCount)),
				Members: &icdv5.GroupMembers{
					AllocationCount: ibmc.Int64Ptr(int64(membersAllocationCount)),
					Units:           &membersUnits,
					MinimumCount:    ibmc.Int64Ptr(int64(membersMinimumCount)),
					MaximumCount:    ibmc.Int64Ptr(int64(membersMaximumCount)),
					StepSizeCount:   ibmc.Int64Ptr(int64(membersStepSizeCount)),
					IsAdjustable:    ibmc.BoolPtr(membersIsAdjustable),
					IsOptional:      ibmc.BoolPtr(membersIsOptional),
					CanScaleDown:    ibmc.BoolPtr(membersCanScaleDown),
				},
				Memory: &icdv5.GroupMemory{
					AllocationMb: ibmc.Int64Ptr(int64(memoryAllocationMb)),
					Units:        &memoryUnits,
					MinimumMb:    ibmc.Int64Ptr(int64(memoryMinimumMb)),
					MaximumMb:    ibmc.Int64Ptr(int64(memoryMaximumMb)),
					StepSizeMb:   ibmc.Int64Ptr(int64(memoryStepSizeMb)),
					IsAdjustable: ibmc.BoolPtr(memoryIsAdjustable),
					IsOptional:   ibmc.BoolPtr(memoryIsOptional),
					CanScaleDown: ibmc.BoolPtr(memoryCanScaleDown),
				},
				Disk: &icdv5.GroupDisk{
					AllocationMb: ibmc.Int64Ptr(int64(diskAllocationMb)),
					Units:        &diskUnits,
					MinimumMb:    ibmc.Int64Ptr(int64(diskMinimumMb)),
					MaximumMb:    ibmc.Int64Ptr(int64(diskMaximumMb)),
					StepSizeMb:   ibmc.Int64Ptr(int64(diskStepSizeMb)),
					IsAdjustable: ibmc.BoolPtr(diskIsAdjustable),
					IsOptional:   ibmc.BoolPtr(diskIsOptional),
					CanScaleDown: ibmc.BoolPtr(diskCanScaleDown),
				},
				Cpu: &icdv5.GroupCpu{
					AllocationCount: ibmc.Int64Ptr(int64(cpuAllocationCount)),
					Units:           &cpuUnits,
					MinimumCount:    ibmc.Int64Ptr(int64(cpuMinimumCount)),
					MaximumCount:    ibmc.Int64Ptr(int64(cpuMaximumCount)),
					StepSizeCount:   ibmc.Int64Ptr(int64(cpuStepSizeCount)),
					IsAdjustable:    ibmc.BoolPtr(cpuIsAdjustable),
					IsOptional:      ibmc.BoolPtr(cpuIsOptional),
					CanScaleDown:    ibmc.BoolPtr(cpuCanScaleDown),
				},
			},
		},
	}

	for _, f := range m {
		f(i)
	}
	return i
}

// Sets up a unit test http server, and creates an external scaling group structure appropriate for unit test.
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
func setupServerAndGetUnitTestExternalSG(testingObj *testing.T, handlers *[]tstutil.Handler, kube *client.Client) (*sgExternal, *httptest.Server, error) {
	mClient, tstServer, err := tstutil.SetupTestServerClient(testingObj, handlers)
	if err != nil || mClient == nil || tstServer == nil {
		return nil, nil, err
	}

	return &sgExternal{
			kube:   *kube,
			client: *mClient,
			logger: logging.NewNopLogger(),
		},
		tstServer,
		nil
}

func TestScalingGroupObserve(t *testing.T) {
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
						err := json.NewEncoder(w).Encode(&icdv5.Groups{})
						if err != nil {
							klog.Errorf("%s", err)
						}
					},
				},
			},
			args: tstutil.Args{
				Managed: sg(),
			},
			want: want{
				mg:  sg(),
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
						err := json.NewEncoder(w).Encode(&icdv5.Groups{})
						if err != nil {
							klog.Errorf("%s", err)
						}
					},
				},
			},
			args: tstutil.Args{
				Managed: sg(),
			},
			want: want{
				mg:  sg(),
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
						err := json.NewEncoder(w).Encode(&icdv5.Groups{})
						if err != nil {
							klog.Errorf("%s", err)
						}
					},
				},
			},
			args: tstutil.Args{
				Managed: sg(),
			},
			want: want{
				mg:  sg(),
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
						sg := instance()
						err := json.NewEncoder(w).Encode(sg)
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
				Managed: sg(
					sgWithExternalNameAnnotation(id),
					sgWithSpec(*params()),
				),
			},
			want: want{
				mg: sg(sgWithSpec(*params()),
					sgWithConditions(cpv1alpha1.Available()),
					sgWithStatus(*observation())),
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
						sg := instance(func(p *icdv5.Groups) {
							p.Groups = instance().Groups
							p.Groups[0].Disk.AllocationMb = ibmc.Int64Ptr(int64(diskAllocationMb * 2))
						})
						err := json.NewEncoder(w).Encode(sg)
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
				Managed: sg(
					sgWithExternalNameAnnotation(id),
					sgWithSpec(*params()),
				),
			},
			want: want{
				mg: sg(sgWithSpec(*params()),
					sgWithConditions(cpv1alpha1.Available()),
					sgWithStatus(*observation(func(p *v1alpha1.ScalingGroupObservation) {
						p.Groups = observation().Groups
						p.Groups[0].Disk.AllocationMb = int64(diskAllocationMb * 2)
						p.Groups[0].Disk.MemberAllocationMb = int64(diskAllocationMb)
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
			e, server, setupErr := setupServerAndGetUnitTestExternalSG(t, &tc.handlers, &tc.kube)
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

func TestScalingGroupCreate(t *testing.T) {
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
						r.Body.Close()
						sg := instance()
						err := json.NewEncoder(w).Encode(sg)
						if err != nil {
							klog.Errorf("%s", err)
						}
					},
				},
			},
			args: tstutil.Args{
				Managed: sg(sgWithSpec(*params())),
			},
			want: want{
				mg: sg(sgWithSpec(*params()),
					sgWithConditions(cpv1alpha1.Creating()),
					sgWithExternalNameAnnotation(id)),
				cre: managed.ExternalCreation{ExternalNameAssigned: true},
				err: nil,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e, server, setupErr := setupServerAndGetUnitTestExternalSG(t, &tc.handlers, &tc.kube)
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

func TestScalingGroupDelete(t *testing.T) {
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
				Managed: sg(sgWithStatus(*observation())),
			},
			want: want{
				mg:  sg(sgWithStatus(*observation()), sgWithConditions(cpv1alpha1.Deleting())),
				err: nil,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e, server, setupErr := setupServerAndGetUnitTestExternalSG(t, &tc.handlers, &tc.kube)
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

func TestScalingGroupUpdate(t *testing.T) {
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
						sg := instance()
						err := json.NewEncoder(w).Encode(sg)
						if err != nil {
							klog.Errorf("%s", err)
						}
					},
				},
			},
			args: tstutil.Args{
				Managed: sg(sgWithSpec(*params()), sgWithStatus(*observation())),
			},
			want: want{
				mg:  sg(sgWithSpec(*params()), sgWithStatus(*observation())),
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
				Managed: sg(sgWithSpec(*params()), sgWithStatus(*observation())),
			},
			want: want{
				mg:  sg(sgWithSpec(*params()), sgWithStatus(*observation())),
				err: errors.New(http.StatusText(http.StatusBadRequest)),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e, server, setupErr := setupServerAndGetUnitTestExternalSG(t, &tc.handlers, &tc.kube)
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
