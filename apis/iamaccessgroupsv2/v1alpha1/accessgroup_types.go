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

// AccessGroupParameters are the configurable fields of a AccessGroup.
type AccessGroupParameters struct {
	// IBM Cloud account identifier.
	AccountID string `json:"accountID"`

	// Assign the specified name to the Access Group. This field has a limit of 100 characters.
	Name string `json:"name"`

	// Assign a description for the Access Group. This field has a limit of 250 characters.
	//+optional
	Description *string `json:"description,omitempty"`

	// An optional transaction id for the request.
	//+optional
	TransactionID *string `json:"transactionID,omitempty"`
}

// AccessGroupObservation are the observable fields of a AccessGroup.
type AccessGroupObservation struct {
	// The group's Access Group ID.
	ID string `json:"id,omitempty"`

	// The UTC timestamp when the group was created.
	CreatedAt *metav1.Time `json:"createdAt,omitempty"`

	// The iam ID of the entity that created the group.
	CreatedByID string `json:"createdById,omitempty"`

	// The UTC timestamp when the group was last modified.
	LastModifiedAt *metav1.Time `json:"lastModifiedAt,omitempty"`

	// The iam ID of the entity that last modified the group.
	LastModifiedByID string `json:"lastModifiedById,omitempty"`

	// A url to the given group resource.
	Href string `json:"href,omitempty"`

	// This is set to true if rules exist for the group.
	IsFederated bool `json:"isFederated,omitempty"`

	// The current state of the group
	State string `json:"state,omitempty"`
}

// A AccessGroupSpec defines the desired state of a AccessGroup.
type AccessGroupSpec struct {
	runtimev1alpha1.ResourceSpec `json:",inline"`
	ForProvider                  AccessGroupParameters `json:"forProvider"`
}

// A AccessGroupStatus represents the observed state of a AccessGroup.
type AccessGroupStatus struct {
	runtimev1alpha1.ResourceStatus `json:",inline"`
	AtProvider                     AccessGroupObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A AccessGroup represents an instance of an IAM policy on IBM Cloud
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="STATUS",type="string",JSONPath=".status.bindingPhase"
// +kubebuilder:printcolumn:name="STATE",type="string",JSONPath=".status.atProvider.state"
// +kubebuilder:printcolumn:name="CLASS",type="string",JSONPath=".spec.classRef.name"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:scope=Cluster
type AccessGroup struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AccessGroupSpec   `json:"spec"`
	Status AccessGroupStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// AccessGroupList contains a list of AccessGroup
type AccessGroupList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AccessGroup `json:"items"`
}
