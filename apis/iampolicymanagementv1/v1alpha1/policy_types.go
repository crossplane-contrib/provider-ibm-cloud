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

// PolicyParameters are the configurable fields of a Policy.
type PolicyParameters struct {
	// The policy type; either 'access' or 'authorization'.
	Type string `json:"type"`

	// The subjects associated with a policy.
	Subjects []PolicySubject `json:"subjects"`

	// A set of role cloud resource names (CRNs) granted by the policy.
	Roles []PolicyRole `json:"roles"`

	// The resources associated with a policy.
	Resources []PolicyResource `json:"resources"`

	// Customer-defined description.
	// +optional
	Description *string `json:"description,omitempty"`
}

// PolicyRole : A role associated with a policy.
type PolicyRole struct {
	// The role cloud resource name granted by the policy.
	RoleID string `json:"roleId"`

	// TODO - this should not be generated as the policy management API does not accept updates to this field
	// The display name of the role.
	// +optional
	// DisplayName *string `json:"displayName,omitempty"`

	// TODO - this should not be generated as the policy management API does not accept updates to this field
	// The description of the role.
	// +optional
	// Description *string `json:"description,omitempty"`
}

// PolicyResource : The attributes of the resource. Note that only one resource is allowed in a policy.
type PolicyResource struct {
	// List of resource attributes.
	Attributes []ResourceAttribute `json:"attributes,omitempty"`
}

// ResourceAttribute : An attribute associated with a resource.
type ResourceAttribute struct {
	// The name of an attribute.
	Name *string `json:"name" validate:"required"`

	// The value of an attribute.
	Value *string `json:"value" validate:"required"`

	// The operator of an attribute.
	Operator *string `json:"operator,omitempty"`
}

// PolicySubject : The subject attribute values that must match in order for this policy to apply in a permission decision.
type PolicySubject struct {
	// List of subject attributes.
	Attributes []SubjectAttribute `json:"attributes,omitempty"`
}

// SubjectAttribute : An attribute associated with a subject.
type SubjectAttribute struct {
	// The name of an attribute.
	Name *string `json:"name" validate:"required"`

	// The value of an attribute.
	Value *string `json:"value" validate:"required"`
}

// PolicyObservation are the observable fields of a Policy.
type PolicyObservation struct {
	// The policy ID.
	ID string `json:"id,omitempty"`

	// The href link back to the policy.
	Href string `json:"href,omitempty"`

	// The UTC timestamp when the policy was created.
	CreatedAt *metav1.Time `json:"createdAt,omitempty"`

	// The iam ID of the entity that created the policy.
	CreatedByID string `json:"createdById,omitempty"`

	// The UTC timestamp when the policy was last modified.
	LastModifiedAt *metav1.Time `json:"lastModifiedAt,omitempty"`

	// The iam ID of the entity that last modified the policy.
	LastModifiedByID string `json:"lastModifiedById,omitempty"`

	// The current state of the policy
	State string `json:"state,omitempty"`
}

// A PolicySpec defines the desired state of a Policy.
type PolicySpec struct {
	runtimev1alpha1.ResourceSpec `json:",inline"`
	ForProvider                  PolicyParameters `json:"forProvider"`
}

// A PolicyStatus represents the observed state of a Policy.
type PolicyStatus struct {
	runtimev1alpha1.ResourceStatus `json:",inline"`
	AtProvider                     PolicyObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A Policy represents an instance of an IAM policy on IBM Cloud
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="STATUS",type="string",JSONPath=".status.bindingPhase"
// +kubebuilder:printcolumn:name="STATE",type="string",JSONPath=".status.atProvider.state"
// +kubebuilder:printcolumn:name="CLASS",type="string",JSONPath=".spec.classRef.name"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,ibmcloud}
type Policy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PolicySpec   `json:"spec"`
	Status PolicyStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// PolicyList contains a list of Policy
type PolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Policy `json:"items"`
}
