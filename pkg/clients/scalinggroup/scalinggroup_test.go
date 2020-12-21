package scalinggroup

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/reference"

	icdv5 "github.com/IBM/experimental-go-sdk/ibmclouddatabasesv5"
	ibmc "github.com/crossplane-contrib/provider-ibm-cloud/pkg/clients"

	"github.com/crossplane-contrib/provider-ibm-cloud/apis/ibmclouddatabasesv5/v1alpha1"
)

var (
	role                   = "Manager"
	role2                  = "Reader"
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

func instanceOpts(m ...func(*icdv5.SetDeploymentScalingGroupOptions)) *icdv5.SetDeploymentScalingGroupOptions {
	i := &icdv5.SetDeploymentScalingGroupOptions{
		ID:      &id,
		GroupID: reference.ToPtrValue(MemberGroupID),
		SetDeploymentScalingGroupRequest: &icdv5.SetDeploymentScalingGroupRequest{
			Members: &icdv5.SetMembersGroupMembers{
				AllocationCount: ibmc.Int64Ptr(int64(membersAllocationCount)),
			},
			Cpu: &icdv5.SetCPUGroupCPU{
				AllocationCount: ibmc.Int64Ptr(int64(cpuAllocationCount)),
			},
			Memory: &icdv5.SetMemoryGroupMemory{
				AllocationMb: ibmc.Int64Ptr(int64(memoryAllocationMb)),
			},
			Disk: &icdv5.SetDiskGroupDisk{
				AllocationMb: ibmc.Int64Ptr(int64(diskAllocationMb)),
			},
		},
	}
	for _, f := range m {
		f(i)
	}
	return i
}

func cr(m ...func(*v1alpha1.ScalingGroup)) *v1alpha1.ScalingGroup {
	i := &v1alpha1.ScalingGroup{
		Spec: v1alpha1.ScalingGroupSpec{
			ForProvider: *params(),
		},
		Status: v1alpha1.ScalingGroupStatus{
			AtProvider: *observation(),
		},
	}
	for _, f := range m {
		f(i)
	}
	return i
}

func TestGenerateSetDeploymentScalingGroupOptions(t *testing.T) {
	type args struct {
		id string
		cr v1alpha1.ScalingGroup
	}
	type want struct {
		instance *icdv5.SetDeploymentScalingGroupOptions
	}
	cases := map[string]struct {
		args args
		want want
	}{
		"FullConversion": {
			args: args{id: id, cr: *cr()},
			want: want{instance: instanceOpts()},
		},
		"MissingFields": {
			args: args{
				id: id,
				cr: *cr(func(p *v1alpha1.ScalingGroup) {
					p.Spec.ForProvider.MemberDisk = nil
					p.Spec.ForProvider.Members = nil

				})},
			want: want{instance: instanceOpts(
				func(p *icdv5.SetDeploymentScalingGroupOptions) {
					p.SetDeploymentScalingGroupRequest = &icdv5.SetDeploymentScalingGroupRequest{
						Memory: &icdv5.SetMemoryGroupMemory{
							AllocationMb: ibmc.Int64Ptr(int64(memoryAllocationMb)),
						},
						Cpu: &icdv5.SetCPUGroupCPU{
							AllocationCount: ibmc.Int64Ptr(int64(cpuAllocationCount)),
						},
					}
				},
			)},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			r := &icdv5.SetDeploymentScalingGroupOptions{}
			GenerateSetDeploymentScalingGroupOptions(tc.args.id, tc.args.cr, r)
			if diff := cmp.Diff(tc.want.instance, r); diff != "" {
				t.Errorf("GenerateSetDeploymentScalingGroupOptions(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestLateInitializeSpecs(t *testing.T) {
	type args struct {
		instance *icdv5.Groups
		params   *v1alpha1.ScalingGroupParameters
	}
	type want struct {
		params *v1alpha1.ScalingGroupParameters
	}
	cases := map[string]struct {
		args args
		want want
	}{
		"SomeFields": {
			args: args{
				params: params(func(p *v1alpha1.ScalingGroupParameters) {
					p.Members = nil
				}),
				instance: instance(),
			},
			want: want{
				params: params(func(p *v1alpha1.ScalingGroupParameters) {
				})},
		},
		"AllFilledAlready": {
			args: args{
				params:   params(),
				instance: instance(),
			},
			want: want{
				params: params()},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			LateInitializeSpec(tc.args.params, tc.args.instance)
			if diff := cmp.Diff(tc.want.params, tc.args.params); diff != "" {
				t.Errorf("LateInitializeSpec(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestGenerateObservation(t *testing.T) {
	type args struct {
		instance *icdv5.Groups
	}
	type want struct {
		obs v1alpha1.ScalingGroupObservation
	}
	cases := map[string]struct {
		args args
		want want
	}{
		"FullConversion": {
			args: args{
				instance: instance(),
			},
			want: want{*observation()},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			o, err := GenerateObservation(tc.args.instance)
			if diff := cmp.Diff(nil, err); diff != "" {
				t.Errorf("GenerateObservation(...): want error != got error:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.obs, o); diff != "" {
				t.Errorf("GenerateObservation(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestIsUpToDate(t *testing.T) {
	type args struct {
		params   *v1alpha1.ScalingGroupParameters
		instance *icdv5.Groups
	}
	type want struct {
		upToDate bool
		isErr    bool
	}
	cases := map[string]struct {
		args args
		want want
	}{
		"IsUpToDate": {
			args: args{
				params:   params(),
				instance: instance(),
			},
			want: want{upToDate: true, isErr: false},
		},
		"NeedsUpdate": {
			args: args{
				params: params(),
				instance: instance(func(i *icdv5.Groups) {
					k := instance().Groups
					k[0].Cpu.AllocationCount = ibmc.Int64Ptr(int64(cpuAllocationCount + 2))
					i.Groups = k
				}),
			},
			want: want{upToDate: false, isErr: false},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {

			r, err := IsUpToDate(id, tc.args.params, tc.args.instance, logging.NewNopLogger())
			if err != nil && !tc.want.isErr {
				t.Error("IsUpToDate(...) unexpected error")
			}
			if diff := cmp.Diff(tc.want.upToDate, r); diff != "" {
				t.Errorf("IsUpToDate(...): -want, +got:\n%s", diff)
			}
		})
	}
}
