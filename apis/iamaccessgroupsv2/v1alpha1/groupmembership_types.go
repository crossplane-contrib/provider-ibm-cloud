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

// GroupMembershipParameters are the configurable fields of a GroupMembership.
type GroupMembershipParameters struct {
	// The Access Group identifier.
	// +immutable
	// +optional
	AccessGroupID *string `json:"accessGroupId"`

	// Reference to AccessGroupID
	// +immutable
	// +optional
	AccessGroupIDRef *runtimev1alpha1.Reference `json:"accessGroupIdRef,omitempty"`

	// Selector for AccessGroupID
	// +immutable
	// +optional
	AccessGroupIDSelector *runtimev1alpha1.Selector `json:"accessGroupIdSelector,omitempty"`

	// An array of member objects to add to an access group.
	Members []AddGroupMembersRequestMembersItem `json:"members"`

	// An optional transaction id for the request.
	//+optional
	TransactionID *string `json:"transactionID,omitempty"`
}

// AddGroupMembersRequestMembersItem : AddGroupMembersRequestMembersItem struct
type AddGroupMembersRequestMembersItem struct {
	// The IBMid or Service Id of the member.
	IamID string `json:"iamId"`

	// The type of the member, must be either "user" or "service".
	Type string `json:"type"`
}

// GroupMembershipObservation are the observable fields of a GroupMembership.
type GroupMembershipObservation struct {
	// The members of an access group.
	Members []ListGroupMembersResponseMember `json:"members,omitempty"`

	// The current state of the group
	State string `json:"state,omitempty"`
}

// ListGroupMembersResponseMember : A single member of an access group in a list.
type ListGroupMembersResponseMember struct {
	// The IBMid or Service Id of the member.
	IamID string `json:"iamId,omitempty"`

	// The member type - either `user` or `service`.
	Type string `json:"type,omitempty"`

	// The user's or service id's name.
	Name string `json:"name,omitempty"`

	// If the member type is user, this is the user's email.
	Email string `json:"email,omitempty"`

	// If the member type is service, this is the service id's description.
	Description string `json:"description,omitempty"`

	// A url to the given member resource.
	Href string `json:"href,omitempty"`

	// The timestamp the membership was created at.
	CreatedAt *metav1.Time `json:"createdAt,omitempty"`

	// The `iam_id` of the entity that created the membership.
	CreatedByID string `json:"createdById,omitempty"`
}

// A GroupMembershipSpec defines the desired state of a GroupMembership.
type GroupMembershipSpec struct {
	runtimev1alpha1.ResourceSpec `json:",inline"`
	ForProvider                  GroupMembershipParameters `json:"forProvider"`
}

// A GroupMembershipStatus represents the observed state of a GroupMembership.
type GroupMembershipStatus struct {
	runtimev1alpha1.ResourceStatus `json:",inline"`
	AtProvider                     GroupMembershipObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A GroupMembership represents an instance of an IAM policy on IBM Cloud
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="STATUS",type="string",JSONPath=".status.bindingPhase"
// +kubebuilder:printcolumn:name="STATE",type="string",JSONPath=".status.atProvider.state"
// +kubebuilder:printcolumn:name="CLASS",type="string",JSONPath=".spec.classRef.name"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:scope=Cluster
type GroupMembership struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GroupMembershipSpec   `json:"spec"`
	Status GroupMembershipStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// GroupMembershipList contains a list of GroupMembership
type GroupMembershipList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GroupMembership `json:"items"`
}
