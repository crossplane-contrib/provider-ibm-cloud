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
	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// BucketConfigParams are the configurable aspects of a bucket
type BucketConfigParams struct {
	// The name of the bucket. Non-mutable.
	//
	// Note:
	//    One of 'Name', 'NameRef' should be specified...
	//
	// +immutable
	// +optional
	Name *string `json:"name,omitempty"`

	// Crossplane reference of the bucket name.
	//
	// Note:
	//    One of 'Name', 'NameRef' should be specified
	//
	// +immutable
	// +optional
	NameRef *string `json:"nameRef,omitempty"`

	// Maximum bytes for this bucket. If set to 0, quota is disabled
	//
	// +optional
	HardQuota *int64 `json:"hardQuota,omitempty"`

	// An access control mechanism based on the network (IP address) where request originated. Requests not originating
	// from IP addresses listed in the `allowed_ip` field will be denied regardless of any access policies (including
	// public access) that might otherwise permit the request.  Viewing or updating the `Firewall` element requires the
	// requester to have the `manager` role.
	//
	// +optional
	Firewall *Firewall `json:"firewall,omitempty"`

	// Enables sending log data to Activity Tracker and LogDNA to provide visibility into object read and write events. All
	// object events are sent to the activity tracker instance defined in the `activity_tracker_crn` field.
	//
	// +optional
	ActivityTracking *ActivityTracking `json:"activityTracking,omitempty"`

	// Enables sending metrics to IBM Cloud Monitoring. All metrics are sent to the IBM Cloud Monitoring instance defined
	// in the `monitoring_crn` field.
	//
	// +optional
	MetricsMonitoring *MetricsMonitoring `json:"metricsMonitoring,omitempty"`

	// Allows users to set headers to be GDPR compliant
	//
	// +optional
	Headers *map[string]string `json:"headers,omitempty"`
}

// BucketConfigObservation are the observable fields of a bucket configuration
type BucketConfigObservation struct {
	// The CRN of the bucket
	CRN string `json:"crn"`

	// Id of the service instance that holds the bucket
	ServiceInstanceID string `json:"serviceInstanceID,omitempty"`

	// RN of the service instance that holds the bucket.
	ServiceInstanceCRN string `json:"serviceInstanceCRN,omitempty"`

	// The creation time of the bucket in RFC 3339 format
	TimeCreated metav1.Time `json:"timeCreated,omitempty"`

	// The modification time of the bucket in RFC 3339 format.
	TimeUpdated metav1.Time `json:"timeUpdated,omitempty"`

	// Total number of objects in the bucket
	ObjectCount int64 `json:"objectCount,omitempty"`

	// Total size of all objects in the bucket
	BytesUsed int64 `json:"bytesUsed,omitempty"`

	// Number of non-current object versions in the bucket. Non-mutable.
	NoncurrentObjectCount int64 `json:"noncurrentObjectCount,omitempty"`

	// Total size of all non-current object versions in the bucket. Non-mutable.
	NoncurrentBytesUsed int64 `json:"noncurrentBytesUsed,omitempty"`

	// Total number of delete markers in the bucket. Non-mutable.
	DeleteMarkerCount int64 `json:"deleteMarkerCount,omitempty"`
}

// BucketConfigSpec - desired end-state of a bucket in the IBM cloud
type BucketConfigSpec struct {
	runtimev1alpha1.ResourceSpec `json:",inline"`

	// Info the IBM cloud needs to create a bucket
	ForProvider BucketConfigParams `json:"forProvider"`
}

// BucketConfigStatus - whatever the status is (the IBM cloud decides that)
type BucketConfigStatus struct {
	runtimev1alpha1.ResourceStatus `json:",inline"`

	// Info the IBM cloud returns about a bucket
	AtProvider BucketConfigObservation `json:"atProvider,omitempty"`
}

// Firewall : An access control mechanism based on the network (IP address) where request originated. Requests not originating from
// IP addresses listed in the `allowed_ip` field will be denied regardless of any access policies (including public
// access) that might otherwise permit the request.  Viewing or updating the `Firewall` element requires the requester
// to have the `manager` role.
type Firewall struct {
	// List of IPv4 or IPv6 addresses in CIDR notation to be affected by firewall in CIDR notation is supported. Passing an
	// empty array will lift the IP address filter.  The `allowed_ip` array can contain a maximum of 1000 items.
	AllowedIP []string `json:"allowedIP"`
}

// ActivityTracking contains the parameters used to configure activity tracking
//
// Setting the CRN to "" signals that we want to disable tracking (as this not a valid CRN)
type ActivityTracking struct {
	// If set to `true`, all object read events (i.e. downloads) will be sent to Activity Tracker.
	//
	// +optional
	ReadDataEvents *bool `json:"readDataEvents,omitempty"`

	// If set to `true`, all object write events (i.e. uploads) will be sent to Activity Tracker.
	//
	// +optional
	WriteDataEvents *bool `json:"writeDataEvents,omitempty"`

	// Required the first time Cctivity Tracking is configured. The is the CRN of the instance of Activity Tracker that will receive object
	// event data. The format is "crn:v1:bluemix:public:logdnaat:{bucket location}:a/{storage account}:{activity tracker
	// service instance}::"
	//
	// If set to "", tracking is disabled (independently of the values of the other paremeters)'
	//
	// +optional
	ActivityTrackerCRN *string `json:"activityTrackerCRN,omitempty"`
}

// MetricsMonitoring contains the parameters used to configure metrics monitoring.
//
// Setting the CRN to "" signals that we want to disable monitoring (as this not a valid CRN)
type MetricsMonitoring struct {
	// If set to `true`, all usage metrics (i.e. `bytes_used`) will be sent to the monitoring service.
	//
	// +optional
	UsageMetricsEnabled *bool `json:"usageMetricsEnabled,omitempty"`

	// If set to `true`, all request metrics (i.e. `rest.object.head`) will be sent to the monitoring service.
	//
	// +optional
	RequestMetricsEnabled *bool `json:"requestMetricsEnabled,omitempty"`

	// Required the first time monitoring is be configured. This is CRN the instance of IBM Cloud Monitoring that will receive
	// the bucket metrics. The format is "crn:v1:bluemix:public:logdnaat:{bucket location}:a/{storage account}:{monitoring
	// service instance}::".
	//
	// If set to "", monitoring is disabled (independently of the values of the other paremeters)
	//
	// +optional
	MetricsMonitoringCRN *string `json:"metricsMonitoringCRN,omitempty"`
}

// MetricsMonitoring enables sending metrics to IBM Cloud Monitoring. All metrics are sent to the IBM Cloud Monitoring instance defined in
// the `monitoring_crn` field.

// +kubebuilder:object:root=true

// BucketConfig contains all the info (spec + status) for a bucket configuration
//
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="STATUS",type="string",JSONPath=".status.bindingPhase"
// +kubebuilder:printcolumn:name="STATE",type="string",JSONPath=".status.atProvider.state"
// +kubebuilder:printcolumn:name="CLASS",type="string",JSONPath=".spec.classRef.name"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,ibmcloud}
type BucketConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BucketConfigSpec   `json:"spec"`
	Status BucketConfigStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// BucketConfigList - list of existing bucket configs
type BucketConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	// List of buckets configs returned
	Items []BucketConfig `json:"bucketConfigs"`
}
