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

// CloudantDatabaseParameters are the configurable fields of a CloudantDatabase.
type CloudantDatabaseParameters struct {

	// The name of the database
	// +immutable
	Db string `validate:"required,ne="`

	// Query parameter to specify whether to enable database partitions when creating a database.
	// +immutable not doing update so these are immutable right ??
	// +optional
	Partitioned *bool `json:"partitioned,omitempty"`

	// The number of shards in the database. Each shard is a partition of the hash value range. Default is 8, unless
	// overridden in the `cluster config`.
	// +immutable not doing update so these are immutable right ??
	// +optional
	Q *int64 `json:"q,omitempty"`
}

// CloudantDatabaseObservation are the observable fields of a CloudantDatabase.
type CloudantDatabaseObservation struct {

	// should the json be omitempty here and all other validate:"required" in observation ??
	// Schema for database cluster information.
	Cluster *DatabaseInformationCluster `json:"cluster" validate:"required"`

	// An opaque string that describes the committed state of the database.
	CommittedUpdateSeq string `json:"committedUpdateSeq,omitempty"`

	// True if the database compaction routine is operating on this database.
	CompactRunning bool `json:"compactRunning" validate:"required"`

	// An opaque string that describes the compaction state of the database.
	CompactedSeq string `json:"compactedSeq,omitempty"`

	// name is in parameters so it shouldn't be here right ??
	// // The name of the database.
	// DbName *string `json:"db_name" validate:"required"`

	// The version of the physical format used for the data when it is stored on disk.
	DiskFormatVersion int64 `json:"diskFormatVersion" validate:"required"`

	// A count of the documents in the specified database.
	DocCount int64 `json:"docCount" validate:"required"`

	// Number of deleted documents.
	DocDelCount int64 `json:"docDelCount" validate:"required"`

	// The engine used for the database.
	Engine string `json:"engine,omitempty"`

	// Schema for database properties.
	Props *DatabaseInformationProps `json:"props" validate:"required"`

	// Schema for size information of content.
	Sizes *ContentInformationSizes `json:"sizes" validate:"required"`

	// An opaque string that describes the state of the database. Do not rely on this string for counting the number of
	// updates.
	UpdateSeq string `json:"updateSeq" validate:"required"`

	// The UUID of the database.
	UUID string `json:"uuid,omitempty"`
}

// DatabaseInformationCluster : Schema for database cluster information.
type DatabaseInformationCluster struct {
	// Schema for the number of replicas of a database in a cluster.
	N int64 `json:"n" validate:"required"`

	// shards is in parameters so it shouldn't be here right ??
	// // Schema for the number of shards in a database. Each shard is a partition of the hash value range.
	// Q int64 `json:"q" validate:"required"`

	// Read quorum. The number of consistent copies of a document that need to be read before a successful reply.
	R int64 `json:"r" validate:"required"`

	// Write quorum. The number of copies of a document that need to be written before a successful reply.
	W int64 `json:"w" validate:"required"`
}

// DatabaseInformationProps : Schema for database properties.
type DatabaseInformationProps struct {
	// The value is `true` for a partitioned database.
	Partitioned bool `json:"partitioned,omitempty"`
}

// ContentInformationSizes : Schema for size information of content.
type ContentInformationSizes struct {
	// The active size of the content, in bytes.
	Active int64 `json:"active" validate:"required"`

	// The total uncompressed size of the content, in bytes.
	External int64 `json:"external" validate:"required"`

	// The total size of the content as stored on disk, in bytes.
	File int64 `json:"file" validate:"required"`
}

// A CloudantDatabaseSpec defines the desired state of a CloudantDatabase.
type CloudantDatabaseSpec struct {
	runtimev1alpha1.ResourceSpec `json:",inline"`
	ForProvider                  CloudantDatabaseParameters `json:"forProvider"`
}

// A CloudantDatabaseStatus represents the observed state of a CloudantDatabase.
type CloudantDatabaseStatus struct {
	runtimev1alpha1.ResourceStatus `json:",inline"`
	AtProvider                     CloudantDatabaseObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A CloudantDatabase represents an instance of a managed service on IBM Cloud
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="STATUS",type="string",JSONPath=".status.bindingPhase"
// +kubebuilder:printcolumn:name="STATE",type="string",JSONPath=".status.atProvider.state"
// +kubebuilder:printcolumn:name="CLASS",type="string",JSONPath=".spec.classRef.name"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,ibmcloud}
type CloudantDatabase struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CloudantDatabaseSpec   `json:"spec"`
	Status CloudantDatabaseStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// CloudantDatabaseList contains a list of CloudantDatabase
type CloudantDatabaseList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CloudantDatabase `json:"items"`
}
