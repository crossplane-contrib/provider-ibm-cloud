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

// ResourceInstanceParameters are the configurable fields of a ResourceInstance.
type ResourceInstanceParameters struct {
	// An human-readable name of the instance.
	Name string `json:"name"`

	// The deployment location where the instance should be hosted.
	// +immutable
	Target string `json:"target"`

	// The name of the resource group where the instance is deployed
	// +immutable
	ResourceGroupName string `json:"resourceGroupName"`

	// The name of the service offering like cloud-object-storage, kms etc
	// +immutable
	ServiceName string `json:"serviceName"`

	// The name of the plan associated with the offering. This value is provided by and stored in the global catalog.
	ResourcePlanName string `json:"resourcePlanName"`

	// Tags that are attached to the instance after provisioning. These tags can be searched and managed through the
	// Tagging API in IBM Cloud.
	// +optional
	Tags []string `json:"tags,omitempty"`

	// A boolean that dictates if the resource instance should be deleted (cleaned up) during the processing of a region
	// instance delete call.
	// +optional
	AllowCleanup *bool `json:"allowCleanup,omitempty"`

	// Configuration options represented as key-value pairs that are passed through to the target resource brokers.
	// +optional
	Parameters *runtime.RawExtension `json:"parameters,omitempty"`

	// Indicates if the resource instance is locked for further update or delete operations. It does not affect actions
	// performed on child resources like aliases, bindings or keys. False by default.
	// +optional
	EntityLock *string `json:"entityLock,omitempty"`
}

// ResourceInstanceObservation are the observable fields of a ResourceInstance.
type ResourceInstanceObservation struct {
	// The ID associated with the instance.
	ID string `json:"id,omitempty"`

	// When you create a new resource, a globally unique identifier (GUID) is assigned. This GUID is a unique internal
	// identifier managed by the resource controller that corresponds to the instance.
	GUID string `json:"guid,omitempty"`

	// The full Cloud Resource Name (CRN) associated with the instance. For more information about this format, see [Cloud
	// Resource Names](https://cloud.ibm.com/docs/overview?topic=overview-crn).
	Crn string `json:"crn,omitempty"`

	// When you provision a new resource, a relative URL path is created identifying the location of the instance.
	URL string `json:"url,omitempty"`

	// An alpha-numeric value identifying the account ID.
	AccountID string `json:"accountId,omitempty"`

	// The short ID of the resource group.
	ResourceGroupID string `json:"resourceGroupId,omitempty"`

	// The long ID (full CRN) of the resource group.
	ResourceGroupCrn string `json:"resourceGroupCrn,omitempty"`

	// ResourceID is the unique ID of the offering. This value is provided by and stored in the global catalog.
	ResourceID string `json:"resourceId,omitempty"`

	// The unique ID of the plan associated with the offering. This value is provided by and stored in the global catalog.
	ResourcePlanID string `json:"resourcePlanId,omitempty"`

	// The full deployment CRN as defined in the global catalog. The Cloud Resource Name (CRN) of the deployment location
	// where the instance is provisioned.
	TargetCrn string `json:"targetCrn,omitempty"`

	// The current state of the instance. For example, if the instance is deleted, it will return removed.
	State string `json:"state,omitempty"`

	// The type of the instance, e.g. `service_instance`.
	Type string `json:"type,omitempty"`

	// The sub-type of instance, e.g. `cfaas`.
	SubType string `json:"subType,omitempty"`

	// The status of the last operation requested on the instance.
	LastOperation *runtime.RawExtension `json:"lastOperation,omitempty"`

	// The resource-broker-provided URL to access administrative features of the instance.
	DashboardURL string `json:"dashboardUrl,omitempty"`

	// The plan history of the instance.
	PlanHistory []PlanHistoryItem `json:"planHistory,omitempty"`

	// The relative path to the resource aliases for the instance.
	ResourceAliasesURL string `json:"resourceAliasesUrl,omitempty"`

	// The relative path to the resource bindings for the instance.
	ResourceBindingsURL string `json:"resourceBindingsUrl,omitempty"`

	// The relative path to the resource keys for the instance.
	ResourceKeysURL string `json:"resourceKeysUrl,omitempty"`

	// The date when the instance was created.
	CreatedAt *metav1.Time `json:"createdAt,omitempty"`

	// The subject who created the instance.
	CreatedBy string `json:"createdBy,omitempty"`

	// The date when the instance was last updated.
	UpdatedAt *metav1.Time `json:"updatedAt,omitempty"`

	// The subject who updated the instance.
	UpdatedBy string `json:"updatedBy,omitempty"`

	// The date when the instance was deleted.
	DeletedAt *metav1.Time `json:"deletedAt,omitempty"`

	// The subject who deleted the instance.
	DeletedBy string `json:"deletedBy,omitempty"`

	// The date when the instance was scheduled for reclamation.
	ScheduledReclaimAt *metav1.Time `json:"scheduledReclaimAt,omitempty"`

	// The subject who initiated the instance reclamation.
	ScheduledReclaimBy string `json:"scheduledReclaimBy,omitempty"`

	// The date when the instance under reclamation was restored.
	RestoredAt *metav1.Time `json:"restoredAt,omitempty"`

	// The subject who restored the instance back from reclamation.
	RestoredBy string `json:"restoredBy,omitempty"`
}

// PlanHistoryItem : An element of the plan history of the instance.
type PlanHistoryItem struct {
	// The unique ID of the plan associated with the offering. This value is provided by and stored in the global catalog.
	ResourcePlanID string `json:"resourcePlanId"`

	// The date on which the plan was changed.
	StartDate *metav1.Time `json:"startDate"`
}

// A ResourceInstanceSpec defines the desired state of a ResourceInstance.
type ResourceInstanceSpec struct {
	runtimev1alpha1.ResourceSpec `json:",inline"`
	ForProvider                  ResourceInstanceParameters `json:"forProvider"`
}

// A ResourceInstanceStatus represents the observed state of a ResourceInstance.
type ResourceInstanceStatus struct {
	runtimev1alpha1.ResourceStatus `json:",inline"`
	AtProvider                     ResourceInstanceObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A ResourceInstance represents an instance of a managed service on IBM Cloud
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="STATUS",type="string",JSONPath=".status.bindingPhase"
// +kubebuilder:printcolumn:name="STATE",type="string",JSONPath=".status.atProvider.state"
// +kubebuilder:printcolumn:name="CLASS",type="string",JSONPath=".spec.classRef.name"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:scope=Cluster
type ResourceInstance struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ResourceInstanceSpec   `json:"spec"`
	Status ResourceInstanceStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ResourceInstanceList contains a list of ResourceInstance
type ResourceInstanceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ResourceInstance `json:"items"`
}
