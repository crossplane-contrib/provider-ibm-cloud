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

// TopicParameters are the configurable fields of a Topic.
type TopicParameters struct {

	// The name of topic to be created.
	// +immutable
	Name string `json:"name"`

	// KafkaAdminURL is the URL to the Event Streams instance admin endpoint
	// +immutable
	// +optional
	KafkaAdminURL *string `json:"kafkaAdminUrl,omitempty"`

	// A reference to the Event Streams Secret Key used to set KafkaAdminURL
	// +immutable
	// +optional
	KafkaAdminURLRef *runtimev1alpha1.Reference `json:"kafkaAdminUrlRef,omitempty"`

	// SourceSelector selects a reference to a resource used to set Source
	// +immutable
	// +optional
	KafkaAdminURLSelector *runtimev1alpha1.Selector `json:"kafkaAdminUrlSelector,omitempty"`

	// The number of partitions.
	// +immutable
	// +optional
	Partitions *int64 `json:"partitions,omitempty"`

	// The number of partitions, this field takes precedence over 'partitions'. Default value is 1 if not specified.
	// +immutable
	// +optional
	PartitionCount *int64 `json:"partitionCount,omitempty"`

	// The config properties to be set for the new topic.
	// +immutable
	// +optional
	Configs []ConfigCreate `json:"configs,omitempty"`
}

// ConfigCreate : ConfigCreate struct
type ConfigCreate struct {

	// The name of the config property.
	Name string `json:"name,omitempty"`

	// The value for a config property.
	Value string `json:"value,omitempty"`
}

// TopicObservation are the observable fields of a Topic.
type TopicObservation struct {

	// The number of partitions.
	// I think partitions can change, so even though it is
	// in parameters, should it also be in observation??
	Partitions int64 `json:"partitions,omitempty"`

	// The number of replication factor.
	ReplicationFactor int64 `json:"replicationFactor,omitempty"`

	// The value of config property 'retention.ms'.
	RetentionMs int64 `json:"retentionMs,omitempty"`

	// The value of config property 'cleanup.policy'.
	CleanupPolicy string `json:"cleanupPolicy,omitempty"`

	// The config properties of the topic.
	Configs *TopicConfigs `json:"configs,omitempty"`

	// The replia assignment of the topic.
	ReplicaAssignments []ReplicaAssignment `json:"replicaAssignments,omitempty"`
}

// TopicConfigs : TopicConfigs struct
type TopicConfigs struct {
	// The value of config property 'cleanup.policy'.
	CleanupPolicy string `json:"cleanup.policy,omitempty"`

	// The value of config property 'min.insync.replicas'.
	MinInsyncReplicas string `json:"min.insync.replicas,omitempty"`

	// The value of config property 'retention.bytes'.
	RetentionBytes string `json:"retention.bytes,omitempty"`

	// The value of config property 'retention.ms'.
	RetentionMs string `json:"retention.ms,omitempty"`

	// The value of config property 'segment.bytes'.
	SegmentBytes string `json:"segment.bytes,omitempty"`

	// The value of config property 'segment.index.bytes'.
	SegmentIndexBytes string `json:"segment.index.bytes,omitempty"`

	// The value of config property 'segment.ms'.
	SegmentMs string `json:"segment.ms,omitempty"`
}

// ReplicaAssignment : ReplicaAssignment struct
type ReplicaAssignment struct {
	// The ID of the partition.
	ID int64 `json:"id,omitempty"`

	Brokers *ReplicaAssignmentBrokers `json:"brokers,omitempty"`
}

// ReplicaAssignmentBrokers : ReplicaAssignmentBrokers struct
type ReplicaAssignmentBrokers struct {
	Replicas []int64 `json:"replicas,omitempty"`
}

// A TopicSpec defines the desired state of a Topic.
type TopicSpec struct {
	runtimev1alpha1.ResourceSpec `json:",inline"`
	// Connection Templates??
	ForProvider TopicParameters `json:"forProvider"`
}

// A TopicStatus represents the observed state of a Topic.
type TopicStatus struct {
	runtimev1alpha1.ResourceStatus `json:",inline"`
	AtProvider                     TopicObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A Topic represents an instance of a managed service on IBM Cloud
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="STATUS",type="string",JSONPath=".status.bindingPhase"
// +kubebuilder:printcolumn:name="STATE",type="string",JSONPath=".status.atProvider.state"
// +kubebuilder:printcolumn:name="CLASS",type="string",JSONPath=".spec.classRef.name"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,ibmcloud}
type Topic struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TopicSpec   `json:"spec"`
	Status TopicStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// TopicList contains a list of Topic
type TopicList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Topic `json:"items"`
}
