package autoscalinggroup

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/reference"

	icdv5 "github.com/IBM/experimental-go-sdk/ibmclouddatabasesv5"

	//  note that missing a  newline in between a built in package import and a github package import results in "File is not `goimports`-ed"
	"github.com/crossplane-contrib/provider-ibm-cloud/apis/ibmclouddatabasesv5/v1alpha1"
	ibmc "github.com/crossplane-contrib/provider-ibm-cloud/pkg/clients"
)

var (
	id                                         = "crn:v1:bluemix:public:databases-for-postgresql:us-south:a/0b5a00334eaf9eb9339d2ab48f20d7f5:dda29288-c259-4dc9-859c-154eb7939cd0::"
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
	cpuRatePeriodSeconds                       = 800
	cpuRateLimitCountPerMember                 = 20
	cpuRateUnits                               = "count"
)

func params(m ...func(*v1alpha1.AutoscalingGroupParameters)) *v1alpha1.AutoscalingGroupParameters {
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

func observation(m ...func(*v1alpha1.AutoscalingGroupObservation)) *v1alpha1.AutoscalingGroupObservation {
	o := &v1alpha1.AutoscalingGroupObservation{}
	for _, f := range m {
		f(o)
	}
	return o
}

func instance(m ...func(*icdv5.AutoscalingGroup)) *icdv5.AutoscalingGroup {
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

func instanceOpts(m ...func(*icdv5.SetAutoscalingConditionsOptions)) *icdv5.SetAutoscalingConditionsOptions {
	i := &icdv5.SetAutoscalingConditionsOptions{
		ID:      &id,
		GroupID: reference.ToPtrValue(MemberGroupID),
		Autoscaling: &icdv5.AutoscalingSetGroup{
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
		},
	}
	for _, f := range m {
		f(i)
	}
	return i
}

func TestGenerateSetDeploymentAutoscalingGroupOptions(t *testing.T) {
	type args struct {
		id     string
		params v1alpha1.AutoscalingGroupParameters
	}
	type want struct {
		instance *icdv5.SetAutoscalingConditionsOptions
	}
	cases := map[string]struct {
		args args
		want want
	}{
		"FullConversion": {
			args: args{id: id, params: *params()},
			want: want{instance: instanceOpts()},
		},
		"MissingFields": {
			args: args{
				id: id,
				params: *params(func(p *v1alpha1.AutoscalingGroupParameters) {
					p.Disk = nil
					p.CPU = nil
				})},
			want: want{instance: instanceOpts(
				func(p *icdv5.SetAutoscalingConditionsOptions) {
					p.Autoscaling = &icdv5.AutoscalingSetGroup{
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
					}
				},
			)},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			r := &icdv5.SetAutoscalingConditionsOptions{}
			GenerateSetAutoscalingConditionsOptions(tc.args.id, tc.args.params, r)
			if diff := cmp.Diff(tc.want.instance, r); diff != "" {
				t.Errorf("GenerateSetDeploymentAutoscalingGroupOptions(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestLateInitializeSpecs(t *testing.T) {
	type args struct {
		instance *icdv5.AutoscalingGroup
		params   *v1alpha1.AutoscalingGroupParameters
	}
	type want struct {
		params *v1alpha1.AutoscalingGroupParameters
	}
	cases := map[string]struct {
		args args
		want want
	}{
		"SomeFields": {
			args: args{
				params: params(func(p *v1alpha1.AutoscalingGroupParameters) {
					p.Disk = nil
				}),
				instance: instance(),
			},
			want: want{
				params: params(func(p *v1alpha1.AutoscalingGroupParameters) {
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
		instance *icdv5.AutoscalingGroup
	}
	type want struct {
		obs v1alpha1.AutoscalingGroupObservation
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
		params   *v1alpha1.AutoscalingGroupParameters
		instance *icdv5.AutoscalingGroup
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
				instance: instance(func(i *icdv5.AutoscalingGroup) {
					i.Memory = nil
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
