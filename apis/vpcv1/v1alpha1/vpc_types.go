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

// ResourceGroupIdentity is supposed to contain either a client ResourceGroupIdentity or a ResourceGroupIdentityByID
type ResourceGroupIdentityBoth struct {
	// The unique identifier for this resource group.
	ID string `json:"id,omitempty"`

	// Whether this is a by-id on the client side
	IsByID bool `json:"isByID,omitempty"`
}

// VPCParameters are input params when creating a VOC
type VPCParameters struct {
	// Indicates whether a default address prefix should be automatically created for each zone in this VPC. If `manual`,
	// this VPC will be created with no default address prefixes.
	//
	// +immutable
	// +optional
	AddressPrefixManagement *string `json:"addressPrefixManagement,omitempty"`

	// Indicates whether this VPC should be connected to Classic Infrastructure. If true, this VPC's resources will have
	// private network connectivity to the account's Classic Infrastructure resources. Only one VPC, per region, may be
	// connected in this way. This value is set at creation and subsequently immutable.
	//
	// +immutable
	// +optional
	ClassicAccess *bool `json:"classicAccess,omitempty"`

	// The unique user-defined name for this VPC. If unspecified, the name will be a hyphenated list of randomly-selected
	// words.
	//
	// +immutable
	// +optional
	Name *string `json:"name,omitempty"`

	// The resource group to use. If unspecified, the account's [default resource
	// group](https://cloud.ibm.com/apidocs/resource-manager#introduction) is used.
	//
	// +immutable
	// +optional
	ResourceGroup *ResourceGroupIdentityBoth `json:"resourceGroup,omitempty"`

	// Allows users to set headers on API requests
	Headers *map[string]string `json:"headers,omitempty"`
}

// VPCSpec is the desired end-state of a VPC in the IBM cloud
type VPCSpec struct {
	runtimev1alpha1.ResourceSpec `json:",inline"`

	// Info the IBM cloud needs to create a VPC
	ForProvider VPCParameters `json:"forProvider"`
}

// ZoneReference : ZoneReference struct
type ZoneReference struct {
	// The URL for this zone.
	Href string `json:"href,omitempty"`

	// The globally unique name for this zone.
	Name string `json:"name,omitempty"`
}

// IP contains the ip address
type IP struct {
	// The IP address. This property may add support for IPv6 addresses in the future. When processing a value in this
	// property, verify that the address is in an expected format. If it is not, log an error. Optionally halt processing
	// and surface the error, or bypass the resource on which the unexpected IP address format was encountered.
	Address string `json:"address,omitempty"`
}

// VpccseSourceIP ...
type VpccseSourceIP struct {
	// The cloud service endpoint source IP address for this zone.
	IP IP `json:"ip,omitempty"`

	// The zone this cloud service endpoint source IP resides in.
	Zone ZoneReference `json:"zone,omitempty"`
}

// NetworkACLReferenceDeleted : If present, this property indicates the referenced resource has been deleted and provides some supplementary
// information.
type NetworkACLReferenceDeleted struct {
	// Link to documentation about deleted resources.
	MoreInfo string `json:"moreIinfo,omitempty"`
}

// NetworkACLReference ...
type NetworkACLReference struct {
	// The CRN for this network ACL.
	CRN string `json:"crn,omitempty"`

	// If present, this property indicates the referenced resource has been deleted and provides
	// some supplementary information.
	Deleted *NetworkACLReferenceDeleted `json:"deleted,omitempty"`

	// The URL for this network ACL.
	Href string `json:"href,omitempty"`

	// The unique identifier for this network ACL.
	ID string `json:"id,omitempty"`

	// The user-defined name for this network ACL.
	Name string `json:"name,omitempty"`
}

// RoutingTableReferenceDeleted : If present, this property indicates the referenced resource has been deleted and provides some supplementary
// information.
type RoutingTableReferenceDeleted struct {
	// Link to documentation about deleted resources.
	MoreInfo string `json:"moreInfo,omitempty"`
}

// RoutingTableReference : RoutingTableReference struct
type RoutingTableReference struct {
	// If present, this property indicates the referenced resource has been deleted and provides
	// some supplementary information.
	Deleted RoutingTableReferenceDeleted `json:"deleted,omitempty"`

	// The URL for this routing table.
	Href string `json:"href,omitempty"`

	// The unique identifier for this routing table.
	ID *string `json:"id,omitempty"`

	// The user-defined name for this routing table.
	Name string `json:"name,omitempty"`

	// The resource type.
	ResourceType string `json:"resourceType,omitempty"`
}

// SecurityGroupReferenceDeleted : If present, this property indicates the referenced resource has been deleted and provides some supplementary
// information.
type SecurityGroupReferenceDeleted struct {
	// Link to documentation about deleted resources.
	MoreInfo string `json:"moreInfo,omitempty"`
}

// SecurityGroupReference : SecurityGroupReference struct
type SecurityGroupReference struct {
	// The security group's CRN.
	CRN string `json:"crn,omitempty"`

	// If present, this property indicates the referenced resource has been deleted and provides
	// some supplementary information.
	Deleted SecurityGroupReferenceDeleted `json:"deleted,omitempty"`

	// The security group's canonical URL.
	Href string `json:"href,omitempty"`

	// The unique identifier for this security group.
	ID string `json:"id,omitempty"`

	// The user-defined name for this security group. Names must be unique within the VPC the security group resides in.
	Name string `json:"name,omitempty"`
}

// ResourceGroupReference ...
type ResourceGroupReference struct {
	// The URL for this resource group.
	Href string `json:"href,omitempty"`

	// The unique identifier for this resource group.
	ID string `json:"id,omitempty"`

	// The user-defined name for this resource group.
	Name string `json:"name,omitempty"`
}

// VPCObservation ...what comes back
type VPCObservation struct {
	// Indicates whether this VPC is connected to Classic Infrastructure. If true, this VPC's resources have private
	// network connectivity to the account's Classic Infrastructure resources. Only one VPC, per region, may be connected
	// in this way. This value is set at creation and subsequently immutable.
	ClassicAccess bool `json:"classicAccess,omitempty"`

	// The date and time that the VPC was created.
	CreatedAt *metav1.Time `json:"createdAt,omitempty"`

	// The CRN for this VPC.
	CRN string `json:"crn,omitempty"`

	// Array of CSE ([Cloud Service Endpoint](https://cloud.ibm.com/docs/resources?topic=resources-service-endpoints))
	// source IP addresses for the VPC. The VPC will have one CSE source IP address per zone.
	CseSourceIps []VpccseSourceIP `json:"cseSourceIps,omitempty"`

	// The default network ACL to use for subnets created in this VPC.
	DefaultNetworkACL NetworkACLReference `json:"defaultNetworkAcl,omitempty"`

	// The default routing table to use for subnets created in this VPC.
	DefaultRoutingTable RoutingTableReference `json:"defaultRoutingTable,omitempty"`

	// The default security group to use for network interfaces created in this VPC.
	DefaultSecurityGroup SecurityGroupReference `json:"defaultSecurityGroup,omitempty"`

	// The URL for this VPC.
	Href string `json:"href,omitempty"`

	// The unique identifier for this VPC.
	ID string `json:"id,omitempty"`

	// The unique user-defined name for this VPC.
	Name string `json:"name,omitempty"`

	// The resource group for this VPC.
	ResourceGroup ResourceGroupReference `json:"resourceGroup,omitempty"`

	// The status of this VPC.
	Status string `json:"status,omitempty"`
}

// VPCStatus - whatever the status is (the IBM cloud decides that)
type VPCStatus struct {
	runtimev1alpha1.ResourceStatus `json:",inline"`

	// Info the IBM cloud returns about a bucket
	AtProvider VPCObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// VPC contains all the info (spec + status) for a VPC
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="STATUS",type="string",JSONPath=".status.bindingPhase"
// +kubebuilder:printcolumn:name="STATE",type="string",JSONPath=".status.atProvider.state"
// +kubebuilder:printcolumn:name="CLASS",type="string",JSONPath=".spec.classRef.name"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,ibmcloud}
type VPC struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   VPCSpec   `json:"spec"`
	Status VPCStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// VPCList - list of existing buckets...
type VPCList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	// List of buckets returned
	Items []VPC `json:"vpcs"`
}
