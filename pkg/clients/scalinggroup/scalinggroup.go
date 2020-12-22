package scalinggroup

import (
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/reference"

	icdv5 "github.com/IBM/experimental-go-sdk/ibmclouddatabasesv5"

	"github.com/crossplane-contrib/provider-ibm-cloud/apis/ibmclouddatabasesv5/v1alpha1"
	ibmc "github.com/crossplane-contrib/provider-ibm-cloud/pkg/clients"
)

// MemberGroupID is the default ID for members group
const MemberGroupID = "member"

// LateInitializeSpec fills optional and unassigned fields with the values in *icdv5.Group object.
func LateInitializeSpec(spec *v1alpha1.ScalingGroupParameters, in *icdv5.Groups) error { // nolint:gocyclo
	if in.Groups == nil || in.Groups != nil && len(in.Groups) == 0 {
		return nil
	}
	group := in.Groups[0]
	if group.Members == nil || group.Members != nil && group.Members.AllocationCount == nil {
		return nil
	}
	if spec.MemberCPU == nil && group.Cpu != nil {
		spec.MemberCPU = &v1alpha1.SetCPUGroupCPU{
			AllocationCount: *group.Cpu.AllocationCount / *group.Members.AllocationCount,
		}
	}
	if spec.MemberDisk == nil && group.Disk != nil {
		spec.MemberDisk = &v1alpha1.SetDiskGroupDisk{
			AllocationMb: *group.Disk.AllocationMb / *group.Members.AllocationCount,
		}
	}
	if spec.MemberMemory == nil && group.Memory != nil {
		spec.MemberMemory = &v1alpha1.SetMemoryGroupMemory{
			AllocationMb: *group.Memory.AllocationMb / *group.Members.AllocationCount,
		}
	}
	if spec.Members == nil && group.Members != nil {
		spec.Members = &v1alpha1.SetMembersGroupMembers{
			AllocationCount: *group.Members.AllocationCount,
		}
	}
	return nil
}

// GenerateSetDeploymentScalingGroupOptions produces SetDeploymentScalingGroupOptions object from ScalingGroupParameters object.
func GenerateSetDeploymentScalingGroupOptions(id string, in v1alpha1.ScalingGroup, o *icdv5.SetDeploymentScalingGroupOptions) error {
	pars := in.Spec.ForProvider
	// the ICDv5 API allocates memory, disk and CPU based on the current number of members.
	// To find the total allocation we need to multiply the current number of members for the allocation per member
	// the ICDv5 API then allocate the amount set for the old members to the new members
	currentMambers := in.Status.AtProvider.Groups[0].Members.AllocationCount
	o.ID = reference.ToPtrValue(id)
	o.GroupID = reference.ToPtrValue(MemberGroupID)
	setDeploymentScalingGroupRequest := &icdv5.SetDeploymentScalingGroupRequest{}
	o.SetDeploymentScalingGroupRequest = setDeploymentScalingGroupRequest
	if pars.Members != nil {
		setDeploymentScalingGroupRequest.Members = &icdv5.SetMembersGroupMembers{
			AllocationCount: &pars.Members.AllocationCount,
		}
	}
	if pars.MemberCPU != nil {
		setDeploymentScalingGroupRequest.Cpu = &icdv5.SetCPUGroupCPU{
			AllocationCount: ibmc.Int64Ptr(pars.MemberCPU.AllocationCount * currentMambers),
		}
	}
	if pars.MemberDisk != nil {
		setDeploymentScalingGroupRequest.Disk = &icdv5.SetDiskGroupDisk{
			AllocationMb: ibmc.Int64Ptr(pars.MemberDisk.AllocationMb * currentMambers),
		}
	}
	if pars.MemberMemory != nil {
		setDeploymentScalingGroupRequest.Memory = &icdv5.SetMemoryGroupMemory{
			AllocationMb: ibmc.Int64Ptr(pars.MemberMemory.AllocationMb * currentMambers),
		}
	}
	return nil
}

// GenerateObservation produces ScalingGroupObservation object from *icdv5.Groups object.
func GenerateObservation(in *icdv5.Groups) (v1alpha1.ScalingGroupObservation, error) {
	o := v1alpha1.ScalingGroupObservation{
		Groups: []v1alpha1.Group{},
	}
	for _, g := range in.Groups {
		if g.Members == nil || g.Members != nil && g.Members.AllocationCount == nil {
			continue
		}
		o.Groups = append(o.Groups, v1alpha1.Group{
			ID:    *g.ID,
			Count: *g.Count,
			Members: v1alpha1.GroupMembers{
				AllocationCount: *g.Members.AllocationCount,
				Units:           g.Members.Units,
				MinimumCount:    g.Members.MinimumCount,
				MaximumCount:    g.Members.MaximumCount,
				StepSizeCount:   g.Members.StepSizeCount,
				IsAdjustable:    g.Members.IsAdjustable,
				IsOptional:      g.Members.IsOptional,
				CanScaleDown:    g.Members.CanScaleDown,
			},
			Memory: v1alpha1.GroupMemory{
				AllocationMb:       *g.Memory.AllocationMb,
				MemberAllocationMb: *g.Memory.AllocationMb / *g.Members.AllocationCount,
				Units:              g.Memory.Units,
				MinimumMb:          g.Memory.MinimumMb,
				MaximumMb:          g.Memory.MaximumMb,
				StepSizeMb:         g.Memory.StepSizeMb,
				IsAdjustable:       g.Memory.IsAdjustable,
				IsOptional:         g.Memory.IsOptional,
				CanScaleDown:       g.Memory.CanScaleDown,
			},
			Disk: v1alpha1.GroupDisk{
				AllocationMb:       *g.Disk.AllocationMb,
				MemberAllocationMb: *g.Disk.AllocationMb / *g.Members.AllocationCount,
				Units:              g.Disk.Units,
				MinimumMb:          g.Disk.MinimumMb,
				MaximumMb:          g.Disk.MaximumMb,
				StepSizeMb:         g.Disk.StepSizeMb,
				IsAdjustable:       g.Disk.IsAdjustable,
				IsOptional:         g.Disk.IsOptional,
				CanScaleDown:       g.Disk.CanScaleDown,
			},
			CPU: v1alpha1.GroupCPU{
				AllocationCount:       *g.Cpu.AllocationCount,
				MemberAllocationCount: *g.Cpu.AllocationCount / *g.Members.AllocationCount,
				Units:                 g.Cpu.Units,
				MinimumCount:          g.Cpu.MinimumCount,
				MaximumCount:          g.Cpu.MaximumCount,
				StepSizeCount:         g.Cpu.StepSizeCount,
				IsAdjustable:          g.Cpu.IsAdjustable,
				IsOptional:            g.Cpu.IsOptional,
				CanScaleDown:          g.Cpu.CanScaleDown,
			},
		})
	}
	return o, nil
}

// IsUpToDate checks whether current state is up-to-date compared to the given
// set of parameters.
func IsUpToDate(id string, in *v1alpha1.ScalingGroupParameters, observed *icdv5.Groups, l logging.Logger) (bool, error) {
	desired := in.DeepCopy()
	actual, err := GenerateScalingGroupParameters(id, observed)
	if err != nil {
		return false, err
	}

	l.Info(cmp.Diff(desired, actual, cmpopts.IgnoreTypes(&runtimev1alpha1.Reference{}, &runtimev1alpha1.Selector{}, []runtimev1alpha1.Reference{})))

	return cmp.Equal(desired, actual, cmpopts.EquateEmpty(),
		cmpopts.IgnoreFields(v1alpha1.ScalingGroupParameters{}),
		cmpopts.IgnoreTypes(&runtimev1alpha1.Reference{}, &runtimev1alpha1.Selector{}, []runtimev1alpha1.Reference{})), nil
}

// GenerateScalingGroupParameters generates scaling group parameters from groups
func GenerateScalingGroupParameters(id string, in *icdv5.Groups) (*v1alpha1.ScalingGroupParameters, error) {
	o := &v1alpha1.ScalingGroupParameters{
		ID: &id,
	}
	if len(in.Groups) > 0 && in.Groups[0].Members != nil && in.Groups[0].Members.AllocationCount != nil {
		if in.Groups[0].Cpu != nil {
			o.MemberCPU = &v1alpha1.SetCPUGroupCPU{
				AllocationCount: *in.Groups[0].Cpu.AllocationCount / *in.Groups[0].Members.AllocationCount,
			}
		}
		if in.Groups[0].Disk != nil {
			o.MemberDisk = &v1alpha1.SetDiskGroupDisk{
				AllocationMb: *in.Groups[0].Disk.AllocationMb / *in.Groups[0].Members.AllocationCount,
			}
		}
		if in.Groups[0].Memory != nil {
			o.MemberMemory = &v1alpha1.SetMemoryGroupMemory{
				AllocationMb: *in.Groups[0].Memory.AllocationMb / *in.Groups[0].Members.AllocationCount,
			}
		}
		if in.Groups[0].Members != nil {
			o.Members = &v1alpha1.SetMembersGroupMembers{
				AllocationCount: *in.Groups[0].Members.AllocationCount,
			}
		}

	}
	return o, nil
}
