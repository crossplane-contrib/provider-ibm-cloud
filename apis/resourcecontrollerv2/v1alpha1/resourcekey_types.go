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

// ResourceKeyParameters are the configurable fields of a ResourceKey.
type ResourceKeyParameters struct {
	// The name of the key.
	Name string `json:"name"`

	// The short or long ID of resource instance or alias.
	// +immutable
	// +optional
	Source *string `json:"source,omitempty"`

	// A reference to a resource used to set Source
	// +immutable
	// +optional
	SourceRef *runtimev1alpha1.Reference `json:"sourceRef,omitempty"`

	// SourceSelector selects a reference to a resource used to set Source
	// +immutable
	// +optional
	SourceSelector *runtimev1alpha1.Selector `json:"sourceSelector,omitempty"`

	// Configuration options represented as key-value pairs. Service defined options are passed through to the target
	// resource brokers, whereas platform defined options are not.
	// +optional
	Parameters *ResourceKeyPostParameters `json:"parameters,omitempty"`

	// The role name or it's CRN.
	// +optional
	Role *string `json:"role,omitempty"`
}

// ResourceKeyPostParameters : Configuration options represented as key-value pairs. Service defined options are passed through to the target
// resource brokers, whereas platform defined options are not.
type ResourceKeyPostParameters struct {
	// An optional platform defined option to reuse an existing IAM serviceId for the role assignment.
	ServiceidCRN string `json:"serviceidCrn,omitempty"`
}

// ResourceKeyObservation are the observable fields of a ResourceKey.
type ResourceKeyObservation struct {
	// The ID associated with the key.
	ID string `json:"id,omitempty"`

	// When you create a new key, a globally unique identifier (GUID) is assigned. This GUID is a unique internal
	// identifier managed by the resource controller that corresponds to the key.
	GUID string `json:"guid,omitempty"`

	// The full Cloud Resource Name (CRN) associated with the key. For more information about this format, see [Cloud
	// Resource Names](https://cloud.ibm.com/docs/overview?topic=overview-crn).
	CRN string `json:"crn,omitempty"`

	// When you created a new key, a relative URL path is created identifying the location of the key.
	URL string `json:"url,omitempty"`

	// An alpha-numeric value identifying the account ID.
	AccountID string `json:"accountId,omitempty"`

	// The short ID of the resource group.
	ResourceGroupID string `json:"resourceGroupId,omitempty"`

	// The CRN of resource instance or alias associated to the key.
	SourceCRN string `json:"sourceCrn,omitempty"`

	// The state of the key.
	State string `json:"state,omitempty"`

	// Specifies whether the key’s credentials support IAM.
	IamCompatible bool `json:"iamCompatible,omitempty"`

	// The relative path to the resource.
	ResourceInstanceURL string `json:"resourceInstanceUrl,omitempty"`

	// The date when the key was created.
	CreatedAt *metav1.Time `json:"createdAt,omitempty"`

	// The date when the key was last updated.
	UpdatedAt *metav1.Time `json:"updatedAt,omitempty"`

	// The date when the key was deleted.
	DeletedAt *metav1.Time `json:"deletedAt,omitempty"`

	// The subject who created the key.
	CreatedBy string `json:"createdBy,omitempty"`

	// The subject who updated the key.
	UpdatedBy string `json:"updatedBy,omitempty"`

	// The subject who deleted the key.
	DeletedBy string `json:"deletedBy,omitempty"`
}

// Credentials : The credentials for a resource.
type Credentials struct {
	// The API key for the credentials.
	Apikey string `json:"apikey,omitempty"`

	// The optional description of the API key.
	IamApikeyDescription string `json:"iamApikeyDescription,omitempty"`

	// The name of the API key.
	IamApikeyName string `json:"iamApikeyName,omitempty"`

	// The Cloud Resource Name for the role of the credentials.
	IamRoleCRN string `json:"iamRoleCrn,omitempty"`

	// The Cloud Resource Name for the service ID of the credentials.
	IamServiceidCRN string `json:"iamServiceidCrn,omitempty"`
}

// A ResourceKeySpec defines the desired state of a ResourceKey.
type ResourceKeySpec struct {
	runtimev1alpha1.ResourceSpec `json:",inline"`
	ConnectionTemplates          map[string]string     `json:"connectionTemplates,omitempty"`
	ForProvider                  ResourceKeyParameters `json:"forProvider"`
}

// A ResourceKeyStatus represents the observed state of a ResourceKey.
type ResourceKeyStatus struct {
	runtimev1alpha1.ResourceStatus `json:",inline"`
	AtProvider                     ResourceKeyObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A ResourceKey represents an instance of a managed service on IBM Cloud
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="STATUS",type="string",JSONPath=".status.bindingPhase"
// +kubebuilder:printcolumn:name="STATE",type="string",JSONPath=".status.atProvider.state"
// +kubebuilder:printcolumn:name="CLASS",type="string",JSONPath=".spec.classRef.name"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:scope=Cluster
type ResourceKey struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ResourceKeySpec   `json:"spec"`
	Status ResourceKeyStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ResourceKeyList contains a list of ResourceKey
type ResourceKeyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ResourceKey `json:"items"`
}
