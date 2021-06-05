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

	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
)

// In spec mandatory fields should be by value, and optional fields pointers
// In status, all fields should be by value, except timestamps - metav1.Time, and runtime.RawExtension which requires special treatment
// https://github.com/crossplane/crossplane/blob/master/design/one-pager-managed-resource-api-design.md#pointer-types-and-markers

// ScalingGroupParameters are the configurable fields of a ScalingGroup.
type ScalingGroupParameters struct {
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

	// Members -
	Members *SetMembersGroupMembers `json:"members,omitempty"`

	// MemberMemory -
	// +optional
	MemberMemory *SetMemoryGroupMemory `json:"memberMemory,omitempty"`

	// MemberCPU -
	// +optional
	MemberCPU *SetCPUGroupCPU `json:"memberCpu,omitempty"`

	// MemberDisk -
	// +optional
	MemberDisk *SetDiskGroupDisk `json:"memberDisk,omitempty"`
}

// SetMembersGroupMembers : SetMembersGroupMembers struct
type SetMembersGroupMembers struct {
	// Allocated number of members.
	AllocationCount int64 `json:"allocationCount,omitempty"`
}

// SetMemoryGroupMemory : SetMemoryGroupMemory struct
type SetMemoryGroupMemory struct {
	// Allocated memory in MB.
	AllocationMb int64 `json:"allocationMb,omitempty"`
}

// SetCPUGroupCPU : SetCPUGroupCPU struct
type SetCPUGroupCPU struct {
	// Number of allocated CPUs.
	AllocationCount int64 `json:"allocationCount,omitempty"`
}

// SetDiskGroupDisk : SetDiskGroupDisk struct
type SetDiskGroupDisk struct {
	// Allocated storage in MB.
	AllocationMb int64 `json:"allocationMb,omitempty"`
}

// ScalingGroupObservation are the observable fields of a ScalingGroup.
type ScalingGroupObservation struct {
	Groups []Group `json:"groups,omitempty"`
	// The current state of the scaling group
	State string `json:"state,omitempty"`
}

// Group : Group struct
type Group struct {
	// Id/name for group.
	ID string `json:"id,omitempty"`

	// Number of entities in the group.
	Count int64 `json:"count,omitempty"`

	Members GroupMembers `json:"members,omitempty"`

	Memory GroupMemory `json:"memory,omitempty"`

	CPU GroupCPU `json:"cpu,omitempty"`

	Disk GroupDisk `json:"disk,omitempty"`
}

// GroupMembers -
type GroupMembers struct {
	// Allocated number of members.
	AllocationCount int64 `json:"allocationCount,omitempty"`

	// Units used for scaling number of members.
	Units *string `json:"units,omitempty"`

	// Minimum number of members.
	MinimumCount *int64 `json:"minimumCount,omitempty"`

	// Maximum number of members.
	MaximumCount *int64 `json:"maximumCount,omitempty"`

	// Step size for number of members.
	StepSizeCount *int64 `json:"stepSizeCount,omitempty"`

	// Is this deployment's number of members adjustable?.
	IsAdjustable *bool `json:"isAdjustable,omitempty"`

	// Is this deployments's number of members optional?.
	IsOptional *bool `json:"isOptional,omitempty"`

	// Can this deployment's number of members scale down?.
	CanScaleDown *bool `json:"canScaleDown,omitempty"`
}

// GroupMemory -
type GroupMemory struct {
	// Total allocated memory in MB.
	AllocationMb int64 `json:"allocationMb,omitempty"`

	// Allocated memory for member in MB.
	MemberAllocationMb int64 `json:"memberAllocationMb,omitempty"`

	// Units used for scaling memory.
	Units *string `json:"units,omitempty"`

	// Minimum memory in MB.
	MinimumMb *int64 `json:"minimumMb,omitempty"`

	// Maximum memory in MB.
	MaximumMb *int64 `json:"maximumMb,omitempty"`

	// Step size memory can be adjusted by in MB.
	StepSizeMb *int64 `json:"stepSizeMb,omitempty"`

	// Is this group's memory adjustable?.
	IsAdjustable *bool `json:"isAdjustable,omitempty"`

	// Is this group's memory optional?.
	IsOptional *bool `json:"isOptional,omitempty"`

	// Can this group's memory scale down?.
	CanScaleDown *bool `json:"canScaleDown,omitempty"`
}

// GroupDisk -
type GroupDisk struct {
	// Total allocated storage in MB.
	AllocationMb int64 `json:"allocationMb,omitempty"`

	// Allocated storage for member in MB.
	MemberAllocationMb int64 `json:"memberAllocationMb,omitempty"`

	// Units used for scaling storage.
	Units *string `json:"units,omitempty"`

	// Minimum allocated storage.
	MinimumMb *int64 `json:"minimumMb,omitempty"`

	// Maximum allocated storage.
	MaximumMb *int64 `json:"maximumMb,omitempty"`

	// Step size storage can be adjusted.
	StepSizeMb *int64 `json:"stepSizeMb,omitempty"`

	// Is this group's storage adjustable?.
	IsAdjustable *bool `json:"isAdjustable,omitempty"`

	// Is this group's storage optional?.
	IsOptional *bool `json:"isOptional,omitempty"`

	// Can this group's storage scale down?.
	CanScaleDown *bool `json:"can_scale_down,omitempty"`
}

// GroupCPU -
type GroupCPU struct {
	// Number of allocated CPUs.
	AllocationCount int64 `json:"allocationCount,omitempty"`

	// Number of allocated CPUs for member
	MemberAllocationCount int64 `json:"memberAllocationCount,omitempty"`

	// Units used for scaling cpu - count means the value is the number of the unit(s) available.
	Units *string `json:"units,omitempty"`

	// Minimum number of CPUs.
	MinimumCount *int64 `json:"minimumCount,omitempty"`

	// Maximum number of CPUs.
	MaximumCount *int64 `json:"maximumCount,omitempty"`

	// Step size CPUs can be adjusted.
	StepSizeCount *int64 `json:"stepSizeCount,omitempty"`

	// Is this group's CPU count adjustable.
	IsAdjustable *bool `json:"isAdjustable,omitempty"`

	// Is this group's CPU optional?.
	IsOptional *bool `json:"isOptional,omitempty"`

	// Can this group's CPU scale down?.
	CanScaleDown *bool `json:"canScaleDown,omitempty"`
}

// A ScalingGroupSpec defines the desired state of a ScalingGroup.
type ScalingGroupSpec struct {
	runtimev1alpha1.ResourceSpec `json:",inline"`
	ConnectionTemplates          map[string]string      `json:"connectionTemplates,omitempty"`
	ForProvider                  ScalingGroupParameters `json:"forProvider"`
}

// A ScalingGroupStatus represents the observed state of a ScalingGroup.
type ScalingGroupStatus struct {
	runtimev1alpha1.ResourceStatus `json:",inline"`
	AtProvider                     ScalingGroupObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A ScalingGroup represents an instance of a managed service on IBM Cloud
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="STATUS",type="string",JSONPath=".status.bindingPhase"
// +kubebuilder:printcolumn:name="STATE",type="string",JSONPath=".status.atProvider.state"
// +kubebuilder:printcolumn:name="CLASS",type="string",JSONPath=".spec.classRef.name"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,ibmcloud}
type ScalingGroup struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ScalingGroupSpec   `json:"spec"`
	Status ScalingGroupStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ScalingGroupList contains a list of ScalingGroup
type ScalingGroupList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ScalingGroup `json:"items"`
}
