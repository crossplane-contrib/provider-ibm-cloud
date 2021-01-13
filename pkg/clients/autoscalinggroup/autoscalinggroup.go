package autoscalinggroup

import (
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/reference"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"

	icdv5 "github.com/IBM/experimental-go-sdk/ibmclouddatabasesv5"

	"github.com/crossplane-contrib/provider-ibm-cloud/apis/ibmclouddatabasesv5/v1alpha1"
	ibmc "github.com/crossplane-contrib/provider-ibm-cloud/pkg/clients"
)

// MemberGroupID is the default ID for members group
const MemberGroupID = "member"

// LateInitializeSpec fills optional and unassigned fields with the values in *icdv5.AutoscalingGroup object.
func LateInitializeSpec(spec *v1alpha1.AutoscalingGroupParameters, in *icdv5.AutoscalingGroup) error { // nolint:gocyclo
	if in == nil {
		return nil
	}
	if spec.Disk == nil && in.Disk != nil {
		spec.Disk = &v1alpha1.AutoscalingDiskGroupDisk{
			Scalers: &v1alpha1.AutoscalingDiskGroupDiskScalers{
				Capacity: &v1alpha1.AutoscalingDiskGroupDiskScalersCapacity{
					Enabled:                  in.Disk.Scalers.Capacity.Enabled,
					FreeSpaceLessThanPercent: in.Disk.Scalers.Capacity.FreeSpaceLessThanPercent,
				},
				IoUtilization: &v1alpha1.AutoscalingDiskGroupDiskScalersIoUtilization{
					Enabled:      in.Disk.Scalers.IoUtilization.Enabled,
					OverPeriod:   in.Disk.Scalers.IoUtilization.OverPeriod,
					AbovePercent: in.Disk.Scalers.IoUtilization.AbovePercent,
				},
			},
			Rate: &v1alpha1.AutoscalingDiskGroupDiskRate{
				IncreasePercent:  ibmc.Float64PtrToInt64Ptr(in.Disk.Rate.IncreasePercent),
				PeriodSeconds:    in.Disk.Rate.PeriodSeconds,
				LimitMbPerMember: ibmc.Float64PtrToInt64Ptr(in.Disk.Rate.LimitMbPerMember),
				Units:            in.Disk.Rate.Units,
			},
		}
	}
	if spec.Memory == nil && in.Memory != nil {
		spec.Memory = &v1alpha1.AutoscalingMemoryGroupMemory{
			Scalers: &v1alpha1.AutoscalingMemoryGroupMemoryScalers{
				IoUtilization: &v1alpha1.AutoscalingMemoryGroupMemoryScalersIoUtilization{
					Enabled:      in.Memory.Scalers.IoUtilization.Enabled,
					OverPeriod:   in.Memory.Scalers.IoUtilization.OverPeriod,
					AbovePercent: in.Memory.Scalers.IoUtilization.AbovePercent,
				},
			},
			Rate: &v1alpha1.AutoscalingMemoryGroupMemoryRate{
				IncreasePercent:  ibmc.Float64PtrToInt64Ptr(in.Memory.Rate.IncreasePercent),
				PeriodSeconds:    in.Memory.Rate.PeriodSeconds,
				LimitMbPerMember: ibmc.Float64PtrToInt64Ptr(in.Memory.Rate.LimitMbPerMember),
				Units:            in.Memory.Rate.Units,
			},
		}
	}
	if spec.CPU == nil && in.Cpu != nil {
		spec.CPU = &v1alpha1.AutoscalingCPUGroupCPU{
			Scalers: ibmc.InterfaceToRawExtension(in.Cpu.Scalers),
			Rate: &v1alpha1.AutoscalingCPUGroupCPURate{
				IncreasePercent:     ibmc.Float64PtrToInt64Ptr(in.Cpu.Rate.IncreasePercent),
				LimitCountPerMember: in.Cpu.Rate.LimitCountPerMember,
				PeriodSeconds:       in.Cpu.Rate.PeriodSeconds,
				Units:               in.Cpu.Rate.Units,
			},
		}
	}

	return nil
}

// GenerateSetAutoscalingConditionsOptions produces SetAutoscalingConditionsOptions object from AutoscalingGroupParameters object.
func GenerateSetAutoscalingConditionsOptions(id string, in v1alpha1.AutoscalingGroupParameters, o *icdv5.SetAutoscalingConditionsOptions) error {
	o.ID = reference.ToPtrValue(id)
	o.GroupID = reference.ToPtrValue(MemberGroupID)
	o.Autoscaling = &icdv5.AutoscalingSetGroup{}
	autoscalingSetGroupAutoscaling := &icdv5.AutoscalingSetGroup{}
	o.Autoscaling = autoscalingSetGroupAutoscaling
	if in.Disk != nil {
		autoscalingSetGroupAutoscaling.Disk = &icdv5.AutoscalingDiskGroupDisk{
			Scalers: &icdv5.AutoscalingDiskGroupDiskScalers{
				Capacity: &icdv5.AutoscalingDiskGroupDiskScalersCapacity{
					Enabled:                  in.Disk.Scalers.Capacity.Enabled,
					FreeSpaceLessThanPercent: in.Disk.Scalers.Capacity.FreeSpaceLessThanPercent,
				},
				IoUtilization: &icdv5.AutoscalingDiskGroupDiskScalersIoUtilization{
					Enabled:      in.Disk.Scalers.IoUtilization.Enabled,
					OverPeriod:   in.Disk.Scalers.IoUtilization.OverPeriod,
					AbovePercent: in.Disk.Scalers.IoUtilization.AbovePercent,
				},
			},
			Rate: &icdv5.AutoscalingDiskGroupDiskRate{
				IncreasePercent:  ibmc.Int64PtrToFloat64Ptr(in.Disk.Rate.IncreasePercent),
				PeriodSeconds:    in.Disk.Rate.PeriodSeconds,
				LimitMbPerMember: ibmc.Int64PtrToFloat64Ptr(in.Disk.Rate.LimitMbPerMember),
				Units:            in.Disk.Rate.Units,
			},
		}
	}
	if in.Memory != nil {
		autoscalingSetGroupAutoscaling.Memory = &icdv5.AutoscalingMemoryGroupMemory{
			Scalers: &icdv5.AutoscalingMemoryGroupMemoryScalers{
				IoUtilization: &icdv5.AutoscalingMemoryGroupMemoryScalersIoUtilization{
					Enabled:      in.Memory.Scalers.IoUtilization.Enabled,
					OverPeriod:   in.Memory.Scalers.IoUtilization.OverPeriod,
					AbovePercent: in.Memory.Scalers.IoUtilization.AbovePercent,
				},
			},
			Rate: &icdv5.AutoscalingMemoryGroupMemoryRate{
				IncreasePercent:  ibmc.Int64PtrToFloat64Ptr(in.Memory.Rate.IncreasePercent),
				PeriodSeconds:    in.Memory.Rate.PeriodSeconds,
				LimitMbPerMember: ibmc.Int64PtrToFloat64Ptr(in.Memory.Rate.LimitMbPerMember),
				Units:            in.Memory.Rate.Units,
			},
		}
	}
	if in.CPU != nil {
		autoscalingSetGroupAutoscaling.Cpu = &icdv5.AutoscalingCPUGroupCPU{
			Scalers: ibmc.RawExtensionToInterface(in.CPU.Scalers),
			Rate: &icdv5.AutoscalingCPUGroupCPURate{
				IncreasePercent:     ibmc.Int64PtrToFloat64Ptr(in.CPU.Rate.IncreasePercent),
				PeriodSeconds:       in.CPU.Rate.PeriodSeconds,
				LimitCountPerMember: in.CPU.Rate.LimitCountPerMember,
				Units:               in.CPU.Rate.Units,
			},
		}
	}
	return nil
}

// GenerateObservation produces AutoscalingGroupObservation object from *icdv5.AutoscalingGroup object.
func GenerateObservation(in *icdv5.AutoscalingGroup) (v1alpha1.AutoscalingGroupObservation, error) {
	o := v1alpha1.AutoscalingGroupObservation{}
	return o, nil
}

// IsUpToDate checks whether current state is up-to-date compared to the given
// set of parameters.
func IsUpToDate(id string, in *v1alpha1.AutoscalingGroupParameters, observed *icdv5.AutoscalingGroup, l logging.Logger) (bool, error) {
	desired := in.DeepCopy()
	actual, err := GenerateAutoscalingGroupParameters(id, observed)
	if err != nil {
		return false, err
	}

	l.Info(cmp.Diff(desired, actual, cmpopts.IgnoreTypes(&runtimev1alpha1.Reference{}, &runtimev1alpha1.Selector{}, []runtimev1alpha1.Reference{})))

	return cmp.Equal(desired, actual, cmpopts.EquateEmpty(),
		cmpopts.IgnoreFields(v1alpha1.AutoscalingGroupParameters{}),
		cmpopts.IgnoreTypes(&runtimev1alpha1.Reference{}, &runtimev1alpha1.Selector{}, []runtimev1alpha1.Reference{})), nil
}

// GenerateAutoscalingGroupParameters generates autoscaling group parameters from AutoscalingGroup
func GenerateAutoscalingGroupParameters(id string, in *icdv5.AutoscalingGroup) (*v1alpha1.AutoscalingGroupParameters, error) {
	o := &v1alpha1.AutoscalingGroupParameters{
		ID: &id,
	}
	if in != nil {
		if in.Disk != nil {
			o.Disk = &v1alpha1.AutoscalingDiskGroupDisk{
				Scalers: &v1alpha1.AutoscalingDiskGroupDiskScalers{
					Capacity: &v1alpha1.AutoscalingDiskGroupDiskScalersCapacity{
						Enabled:                  in.Disk.Scalers.Capacity.Enabled,
						FreeSpaceLessThanPercent: in.Disk.Scalers.Capacity.FreeSpaceLessThanPercent,
					},
					IoUtilization: &v1alpha1.AutoscalingDiskGroupDiskScalersIoUtilization{
						Enabled:      in.Disk.Scalers.IoUtilization.Enabled,
						OverPeriod:   in.Disk.Scalers.IoUtilization.OverPeriod,
						AbovePercent: in.Disk.Scalers.IoUtilization.AbovePercent,
					},
				},
				Rate: &v1alpha1.AutoscalingDiskGroupDiskRate{
					IncreasePercent:  ibmc.Float64PtrToInt64Ptr(in.Disk.Rate.IncreasePercent),
					PeriodSeconds:    in.Disk.Rate.PeriodSeconds,
					LimitMbPerMember: ibmc.Float64PtrToInt64Ptr(in.Disk.Rate.LimitMbPerMember),
					Units:            in.Disk.Rate.Units,
				},
			}
		}
		if in.Memory != nil {
			o.Memory = &v1alpha1.AutoscalingMemoryGroupMemory{
				Scalers: &v1alpha1.AutoscalingMemoryGroupMemoryScalers{
					IoUtilization: &v1alpha1.AutoscalingMemoryGroupMemoryScalersIoUtilization{
						Enabled:      in.Memory.Scalers.IoUtilization.Enabled,
						OverPeriod:   in.Memory.Scalers.IoUtilization.OverPeriod,
						AbovePercent: in.Memory.Scalers.IoUtilization.AbovePercent,
					},
				},
				Rate: &v1alpha1.AutoscalingMemoryGroupMemoryRate{
					IncreasePercent:  ibmc.Float64PtrToInt64Ptr(in.Memory.Rate.IncreasePercent),
					PeriodSeconds:    in.Memory.Rate.PeriodSeconds,
					LimitMbPerMember: ibmc.Float64PtrToInt64Ptr(in.Memory.Rate.LimitMbPerMember),
					Units:            in.Memory.Rate.Units,
				},
			}
		}
		if in.Cpu != nil {
			o.CPU = &v1alpha1.AutoscalingCPUGroupCPU{
				Scalers: ibmc.InterfaceToRawExtension(in.Cpu.Scalers),
				Rate: &v1alpha1.AutoscalingCPUGroupCPURate{
					IncreasePercent:     ibmc.Float64PtrToInt64Ptr(in.Cpu.Rate.IncreasePercent),
					LimitCountPerMember: in.Cpu.Rate.LimitCountPerMember,
					PeriodSeconds:       in.Cpu.Rate.PeriodSeconds,
					Units:               in.Cpu.Rate.Units,
				},
			}
		}
	}
	return o, nil
}
