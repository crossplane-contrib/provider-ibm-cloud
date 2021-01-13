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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
)

// In spec mandatory fields should be by value, and optional fields pointers
// In status, all fields should be by value, except timestamps - metav1.Time, and runtime.RawExtension which requires special treatment
// https://github.com/crossplane/crossplane/blob/master/design/one-pager-managed-resource-api-design.md#pointer-types-and-markers

// NOTE: Type of IncreasePercent, LimitMbPerMember, is float64 in the SDK but
// float is not supported by controller-runtime.
// See https://github.com/kubernetes-sigs/controller-tools/issues/245

// AutoscalingGroupParameters are the configurable fields of a AutoscalingGroup.
type AutoscalingGroupParameters struct {
	// Deployment ID.
	// +immutable
	// +optional
	ID *string `json:"id,omitempty"`

	// IDRef is a reference to an ICD resource instance used to set ID
	// +immutable
	// +optional
	IDRef *runtimev1alpha1.Reference `json:"idRef,omitempty"`

	// SourceSelector selects a reference to an ICD resource instance used to set ID.
	// +immutable
	// +optional
	IDSelector *runtimev1alpha1.Selector `json:"idSelector,omitempty"`

	// Disk -
	// +optional
	Disk *AutoscalingDiskGroupDisk `json:"disk,omitempty"`

	// 	Memory -
	// +optional
	Memory *AutoscalingMemoryGroupMemory `json:"memory,omitempty"`

	// CPU -
	// +optional
	CPU *AutoscalingCPUGroupCPU `json:"cpu,omitempty"`
}

// AutoscalingDiskGroupDisk : AutoscalingDiskGroupDisk struct
type AutoscalingDiskGroupDisk struct {
	Scalers *AutoscalingDiskGroupDiskScalers `json:"scalers,omitempty"`

	Rate *AutoscalingDiskGroupDiskRate `json:"rate,omitempty"`
}

// AutoscalingDiskGroupDiskScalers : AutoscalingDiskGroupDiskScalers struct
type AutoscalingDiskGroupDiskScalers struct {
	Capacity *AutoscalingDiskGroupDiskScalersCapacity `json:"capacity,omitempty"`

	IoUtilization *AutoscalingDiskGroupDiskScalersIoUtilization `json:"ioUtilization,omitempty"`
}

// AutoscalingDiskGroupDiskScalersIoUtilization : AutoscalingDiskGroupDiskScalersIoUtilization struct
type AutoscalingDiskGroupDiskScalersIoUtilization struct {
	Enabled *bool `json:"enabled,omitempty"`

	OverPeriod *string `json:"overPeriod,omitempty"`

	AbovePercent *int64 `json:"abovePercent,omitempty"`
}

// AutoscalingDiskGroupDiskScalersCapacity : AutoscalingDiskGroupDiskScalersCapacity struct
type AutoscalingDiskGroupDiskScalersCapacity struct {
	Enabled *bool `json:"enabled,omitempty"`

	FreeSpaceLessThanPercent *int64 `json:"freeSpaceLessThanPercent,omitempty"`
}

// AutoscalingDiskGroupDiskRate : AutoscalingDiskGroupDiskRate struct
type AutoscalingDiskGroupDiskRate struct {
	IncreasePercent *int64 `json:"increasePercent,omitempty"`

	PeriodSeconds *int64 `json:"periodSeconds,omitempty"`

	LimitMbPerMember *int64 `json:"limitMbPerMember,omitempty"`

	Units *string `json:"units,omitempty"`
}

// AutoscalingMemoryGroupMemory : AutoscalingMemoryGroupMemory struct
type AutoscalingMemoryGroupMemory struct {
	Scalers *AutoscalingMemoryGroupMemoryScalers `json:"scalers,omitempty"`

	Rate *AutoscalingMemoryGroupMemoryRate `json:"rate,omitempty"`
}

// AutoscalingMemoryGroupMemoryScalers : AutoscalingMemoryGroupMemoryScalers struct
type AutoscalingMemoryGroupMemoryScalers struct {
	IoUtilization *AutoscalingMemoryGroupMemoryScalersIoUtilization `json:"ioUtilization,omitempty"`
}

// AutoscalingMemoryGroupMemoryScalersIoUtilization : AutoscalingMemoryGroupMemoryScalersIoUtilization struct
type AutoscalingMemoryGroupMemoryScalersIoUtilization struct {
	Enabled *bool `json:"enabled,omitempty"`

	OverPeriod *string `json:"overPeriod,omitempty"`

	AbovePercent *int64 `json:"abovePercent,omitempty"`
}

// AutoscalingMemoryGroupMemoryRate : AutoscalingMemoryGroupMemoryRate struct
type AutoscalingMemoryGroupMemoryRate struct {
	IncreasePercent *int64 `json:"increasePercent,omitempty"`

	PeriodSeconds *int64 `json:"periodSeconds,omitempty"`

	LimitMbPerMember *int64 `json:"limitMbPerMember,omitempty"`

	Units *string `json:"units,omitempty"`
}

// AutoscalingCPUGroupCPU : AutoscalingCPUGroupCPU struct
type AutoscalingCPUGroupCPU struct {
	Scalers *runtime.RawExtension `json:"scalers,omitempty"`

	Rate *AutoscalingCPUGroupCPURate `json:"rate,omitempty"`
}

// AutoscalingCPUGroupCPURate : AutoscalingCPUGroupCPURate struct
type AutoscalingCPUGroupCPURate struct {
	IncreasePercent *int64 `json:"increasePercent,omitempty"`

	PeriodSeconds *int64 `json:"periodSeconds,omitempty"`

	LimitCountPerMember *int64 `json:"limitCountPerMember,omitempty"`

	Units *string `json:"units,omitempty"`
}

// A AutoscalingGroupSpec defines the desired state of a AutoscalingGroup.
type AutoscalingGroupSpec struct {
	runtimev1alpha1.ResourceSpec `json:",inline"`
	ConnectionTemplates          map[string]string          `json:"connectionTemplates,omitempty"`
	ForProvider                  AutoscalingGroupParameters `json:"forProvider"`
}

// AutoscalingGroupObservation are the observable fields of a Autoscaling Group.
type AutoscalingGroupObservation struct {
	// The current state of the whitelist
	State string `json:"state,omitempty"`
}

// A AutoscalingGroupStatus represents the observed state of a AutoscalingGroup.
type AutoscalingGroupStatus struct {
	runtimev1alpha1.ResourceStatus `json:",inline"`
	AtProvider                     AutoscalingGroupObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A AutoscalingGroup represents an instance of a managed service on IBM Cloud
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="STATUS",type="string",JSONPath=".status.bindingPhase"
// +kubebuilder:printcolumn:name="STATE",type="string",JSONPath=".status.atProvider.state"
// +kubebuilder:printcolumn:name="CLASS",type="string",JSONPath=".spec.classRef.name"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:scope=Cluster
type AutoscalingGroup struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AutoscalingGroupSpec   `json:"spec"`
	Status AutoscalingGroupStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// AutoscalingGroupList contains a list of AutoscalingGroup
type AutoscalingGroupList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AutoscalingGroup `json:"items"`
}
