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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
)

// In spec mandatory fields should be by value, and optional fields pointers
// In status, all fields should be by value, except timestamps - metav1.Time, and runtime.RawExtension which requires special treatment
// https://github.com/crossplane/crossplane/blob/master/design/one-pager-managed-resource-api-design.md#pointer-types-and-markers

// CustomRoleParameters are the configurable fields of a CustomRole.
type CustomRoleParameters struct {
	// The display name of the role that is shown in the console.
	DisplayName string `json:"displayName"`

	// The actions of the role.
	Actions []string `json:"actions"`

	// The name of the role that is used in the CRN. Can only be alphanumeric and has to be capitalized.
	Name string `json:"name"`

	// The account GUID.
	AccountID string `json:"accountId"`

	// The service name.
	ServiceName string `json:"serviceName"`

	// The description of the role.
	//+optional
	Description *string `json:"description,omitempty"`
}

// CustomRoleObservation are the observable fields of a CustomRole.
type CustomRoleObservation struct {
	// The role ID.
	ID string `json:"id,omitempty"`

	// The role CRN.
	CRN string `json:"crn,omitempty"`

	// The UTC timestamp when the role was created.
	CreatedAt *metav1.Time `json:"createdAt,omitempty"`

	// The iam ID of the entity that created the role.
	CreatedByID string `json:"createdById,omitempty"`

	// The UTC timestamp when the role was last modified.
	LastModifiedAt *metav1.Time `json:"lastModifiedAt,omitempty"`

	// The iam ID of the entity that last modified the role.
	LastModifiedByID string `json:"lastModifiedById,omitempty"`

	// The href link back to the role.
	Href string `json:"href,omitempty"`

	// The current state of the role
	State string `json:"state,omitempty"`
}

// A CustomRoleSpec defines the desired state of a CustomRole.
type CustomRoleSpec struct {
	runtimev1alpha1.ResourceSpec `json:",inline"`
	ForProvider                  CustomRoleParameters `json:"forProvider"`
}

// A CustomRoleStatus represents the observed state of a CustomRole.
type CustomRoleStatus struct {
	runtimev1alpha1.ResourceStatus `json:",inline"`
	AtProvider                     CustomRoleObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A CustomRole represents an instance of an IAM policy on IBM Cloud
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="STATUS",type="string",JSONPath=".status.bindingPhase"
// +kubebuilder:printcolumn:name="STATE",type="string",JSONPath=".status.atProvider.state"
// +kubebuilder:printcolumn:name="CLASS",type="string",JSONPath=".spec.classRef.name"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:scope=Cluster
type CustomRole struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CustomRoleSpec   `json:"spec"`
	Status CustomRoleStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// CustomRoleList contains a list of CustomRole
type CustomRoleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CustomRole `json:"items"`
}
