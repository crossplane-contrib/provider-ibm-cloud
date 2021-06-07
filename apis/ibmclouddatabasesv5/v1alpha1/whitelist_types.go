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

// WhitelistParameters are the configurable fields of a Whitelist.
type WhitelistParameters struct {
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

	// An array of allowlist entries.
	IPAddresses []WhitelistEntry `json:"ipAddresses,omitempty"`

	// Verify that the current allowlist matches a provided ETag value. Use in conjunction with the GET operation's ETag
	// header to ensure synchronicity between clients.
	// +optional
	IfMatch *string `json:"IfMatch,omitempty"`
}

// WhitelistEntry : WhitelistEntry struct
type WhitelistEntry struct {
	// An IPv4 address or a CIDR range (netmasked IPv4 address).
	Address string `json:"address,omitempty"`

	// A human readable description of the address or range for identification purposes.
	// +optional
	Description *string `json:"description,omitempty"`
}

// WhitelistObservation are the observable fields of a Whitelist.
type WhitelistObservation struct {
	// The current state of the whitelist
	State string `json:"state,omitempty"`
}

// A WhitelistSpec defines the desired state of a Whitelist.
type WhitelistSpec struct {
	runtimev1alpha1.ResourceSpec `json:",inline"`
	ConnectionTemplates          map[string]string   `json:"connectionTemplates,omitempty"`
	ForProvider                  WhitelistParameters `json:"forProvider"`
}

// A WhitelistStatus represents the observed state of a Whitelist.
type WhitelistStatus struct {
	runtimev1alpha1.ResourceStatus `json:",inline"`
	AtProvider                     WhitelistObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A Whitelist represents an instance of a managed service on IBM Cloud
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="STATUS",type="string",JSONPath=".status.bindingPhase"
// +kubebuilder:printcolumn:name="STATE",type="string",JSONPath=".status.atProvider.state"
// +kubebuilder:printcolumn:name="CLASS",type="string",JSONPath=".spec.classRef.name"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,ibmcloud}
type Whitelist struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   WhitelistSpec   `json:"spec"`
	Status WhitelistStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// WhitelistList contains a list of Whitelist
type WhitelistList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Whitelist `json:"items"`
}
