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

// AccessGroupRuleParameters are the configurable fields of a AccessGroupRule.
type AccessGroupRuleParameters struct {
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

	// The number of hours that the rule lives for (Must be between 1 and 24).
	Expiration int64 `json:"expiration"`

	// The url of the identity provider.
	RealmName string `json:"realmName"`

	// A list of conditions the rule must satisfy.
	Conditions []RuleCondition `json:"conditions"`

	// The name of the rule.
	Name string `json:"name"`

	// An optional transaction id for the request.
	//+optional
	TransactionID *string `json:"transactionID,omitempty"`
}

// RuleCondition : The condition of a rule.
type RuleCondition struct {
	// The claim to evaluate against. This will be found in the `ext` claims of a user's login request.
	Claim string `json:"claim"`

	// The operation to perform on the claim. Valid operators are EQUALS, EQUALS_IGNORE_CASE, IN, NOT_EQUALS_IGNORE_CASE,
	// NOT_EQUALS, and CONTAINS.
	Operator string `json:"operator"`

	// The stringified JSON value that the claim is compared to using the operator.
	Value string `json:"value"`
}

// AccessGroupRuleObservation are the observable fields of a AccessGroupRule.
type AccessGroupRuleObservation struct {
	// The rule id.
	ID string `json:"id,omitempty"`

	// The account id that the group is in.
	AccountID string `json:"accountId,omitempty"`

	// The UTC timestamp when the rule was created.
	CreatedAt *metav1.Time `json:"createdAt,omitempty"`

	// The iam ID of the entity that created the rule.
	CreatedByID string `json:"createdById,omitempty"`

	// The UTC timestamp when the rule was last modified.
	LastModifiedAt *metav1.Time `json:"lastModifiedAt,omitempty"`

	// The iam ID of the entity that last modified the rule.
	LastModifiedByID string `json:"lastModifiedById,omitempty"`

	// The current state of the group
	State string `json:"state,omitempty"`
}

// A AccessGroupRuleSpec defines the desired state of a AccessGroupRule.
type AccessGroupRuleSpec struct {
	runtimev1alpha1.ResourceSpec `json:",inline"`
	ForProvider                  AccessGroupRuleParameters `json:"forProvider"`
}

// A AccessGroupRuleStatus represents the observed state of a AccessGroupRule.
type AccessGroupRuleStatus struct {
	runtimev1alpha1.ResourceStatus `json:",inline"`
	AtProvider                     AccessGroupRuleObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A AccessGroupRule represents an instance of an IAM policy on IBM Cloud
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="STATUS",type="string",JSONPath=".status.bindingPhase"
// +kubebuilder:printcolumn:name="STATE",type="string",JSONPath=".status.atProvider.state"
// +kubebuilder:printcolumn:name="CLASS",type="string",JSONPath=".spec.classRef.name"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:scope=Cluster
type AccessGroupRule struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AccessGroupRuleSpec   `json:"spec"`
	Status AccessGroupRuleStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// AccessGroupRuleList contains a list of AccessGroupRule
type AccessGroupRuleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AccessGroupRule `json:"items"`
}
