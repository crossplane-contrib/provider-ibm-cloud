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
)

// Zone info for the workes
type Zone struct {
	// +immutable
	// +optional
	ID *string `json:"id,omitempty" description:"The id"`

	// +immutable
	// +optional
	SubnetID *string `json:"subnetID,omitempty"`
}

// WorkerPoolConfig is needin order to create a cluster
type WorkerPoolConfig struct {
	// +immutable
	// +optional
	DiskEncryption *bool `json:"diskEncryption,omitempty"`

	// +immutable
	Entitlement string `json:"entitlement"`

	// +immutable
	Flavor string `json:"flavor"`

	// +immutable
	Isolation *string `json:"isolation,omitempty"`

	// +immutable
	// +optional
	Labels *map[string]string `json:"labels,omitempty"`

	// +immutable
	Name string `json:"name" description:"The workerpool's name"`

	// +immutable
	VpcID string `json:"vpcID"`

	// +immutable
	WorkerCount int `json:"workerCount"`

	// +immutable
	Zones []Zone `json:"zones"`
}

// ClusterCreateRequest contains the params used to create a cluster
type ClusterCreateRequest struct {
	// +immutable
	DisablePublicServiceEndpoint bool `json:"disablePublicServiceEndpoint"`

	// +immutable
	KubeVersion string `json:"kubeVersion" description:"kubeversion of cluster"`

	// +immutable
	// +optional
	Billing *string `json:"billing,omitempty"`

	// +immutable
	PodSubnet string `json:"podSubnet"`

	// +immutable
	Provider string `json:"provider"`

	// +immutable
	ServiceSubnet string `json:"serviceSubnet"`

	// +immutable
	Name string `json:"name" description:"The cluster's name"`

	// +immutable
	DefaultWorkerPoolEntitlement string `json:"defaultWorkerPoolEntitlement"`

	// +immutable
	CosInstanceCRN string `json:"cosInstanceCRN"`

	// +immutable
	WorkerPools WorkerPoolConfig `json:"workerPool"`
}

// Feat ...
type Feat struct {
	// +immutable
	KeyProtectEnabled bool `json:"keyProtectEnabled"`

	// +immutable
	PullSecretApplied bool `json:"pullSecretApplied"`
}

// IngresInfo ...
type IngresInfo struct {
	// +immutable
	HostName string `json:"hostname"`

	// +immutable
	SecretName string `json:"secretName"`
}

// LifeCycleInfo ...
type LifeCycleInfo struct {
	ModifiedDate             *metav1.Time `json:"modifiedDate"`
	MasterStatus             string       `json:"masterStatus"`
	MasterStatusModifiedDate string       `json:"masterStatusModifiedDate"`
	MasterHealth             string       `json:"masterHealth"`
	MasterState              string       `json:"masterState"`
}

// Endpoints ...
type Endpoints struct {
	PrivateServiceEndpointEnabled bool   `json:"privateServiceEndpointEnabled"`
	PrivateServiceEndpointURL     string `json:"privateServiceEndpointURL"`
	PublicServiceEndpointEnabled  bool   `json:"publicServiceEndpointEnabled"`
	PublicServiceEndpointURL      string `json:"publicServiceEndpointURL"`
}

// Addon ...
type Addon struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// ClusterInfo contains the "observation" info
type ClusterInfo struct {
	CreatedDate       *metav1.Time  `json:"createdDate"`
	DataCenter        string        `json:"dataCenter"`
	ID                string        `json:"id"`
	Location          string        `json:"location"`
	Entitlement       string        `json:"entitlement"`
	MasterKubeVersion string        `json:"masterKubeVersion"`
	Name              string        `json:"name"`
	Region            string        `json:"region"`
	ResourceGroupID   string        `json:"resourceGroup"`
	State             string        `json:"state"`
	IsPaid            bool          `json:"isPaid"`
	Addons            []Addon       `json:"addons"`
	OwnerEmail        string        `json:"ownerEmail"`
	Type              string        `json:"type"`
	TargetVersion     string        `json:"targetVersion"`
	ServiceSubnet     string        `json:"serviceSubnet"`
	ResourceGroupName string        `json:"resourceGroupName"`
	Provider          string        `json:"provider"`
	PodSubnet         string        `json:"podSubnet"`
	MultiAzCapable    bool          `json:"multiAzCapable"`
	APIUser           string        `json:"apiUser"`
	ServerURL         string        `json:"serverURL"`
	MasterURL         string        `json:"masterURL"`
	MasterStatus      string        `json:"masterStatus"`
	DisableAutoUpdate bool          `json:"disableAutoUpdate"`
	WorkerZones       []string      `json:"workerZones"`
	Vpcs              []string      `json:"vpcs"`
	CRN               string        `json:"crn"`
	VersionEOS        string        `json:"versionEOS"`
	ServiceEndpoints  Endpoints     `json:"serviceEndpoints"`
	Lifecycle         LifeCycleInfo `json:"lifecycle"`
	WorkerCount       int           `json:"workerCount"`
	Ingress           IngresInfo    `json:"ingress"`
	Features          Feat          `json:"features"`
}

// +kubebuilder:object:root=true

// Cluster contains all the info (spec + status) for a cluster
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="STATUS",type="string",JSONPath=".status.bindingPhase"
// +kubebuilder:printcolumn:name="STATE",type="string",JSONPath=".status.atProvider.state"
// +kubebuilder:printcolumn:name="CLASS",type="string",JSONPath=".spec.classRef.name"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,ibmcloud}
type Cluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ClusterCreateRequest `json:"spec"`
	Status ClusterInfo          `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ClusterList  list of existing clusters...
type ClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	// List of clusters returned
	Items []Cluster `json:"clusters"`
}
