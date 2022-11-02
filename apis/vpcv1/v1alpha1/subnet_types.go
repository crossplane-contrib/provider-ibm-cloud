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

// NetworkACLIdentity identifies ... a network ACL.
// We allow only 1 parameter - ID - for simplicity; the IBM API has 3 parameters
type NetworkACLIdentity struct {
	// The unique identifier for this network ACL.
	ID string `json:"id"`
}

// PublicGatewayIdentity identifies a public gateway by a unique property.
// We allow only 1 parameter - ID - for simplicity; the IBM API has 3 parameters
type PublicGatewayIdentity struct {
	// The unique identifier for this public gateway.
	ID string `json:"id"`
}

// RoutingTableIdentity : Identifies a routing table by a unique property.
// We allow only 1 parameter - ID - for simplicity; the IBM API has 2 parameters
type RoutingTableIdentity struct {
	// The unique identifier for this routing table.
	ID string `json:"id"`
}

// VPCIdentity identifies a VPC by a unique property. Only one element should be set (we only allow one parameter - ID - for
// simplicity; the IBM cloud API allows 3)
type VPCIdentity struct {
	// Crossplane reference of the VPC name
	//
	// +immutable
	// +optional
	VPCRef *runtimev1alpha1.Reference `json:"vpcRef,omitempty"`

	// Selects a reference to a VPC
	//
	// +immutable
	// +optional
	VPCSelector *runtimev1alpha1.Selector `json:"vpcSelector,omitempty"`

	// The unique identifier for this VPC.
	ID *string `json:"id,omitempty"`
}

// ZoneIdentity ...for subnet creation only (we only allow one parameter - Name - for simplicity;
// IBM cloud API allows 2)
type ZoneIdentity struct {
	// The globally unique name for this zone.
	Name string `json:"name"`
}

// SubnetPrototypeSubnetByCIDR are input params when creating a Subnet
type SubnetPrototypeSubnetByCIDR struct {
	// The IP version(s) to support for this subnet. Only current allowable value is 'ipv4'
	//
	// +immutable
	// +optional
	IPVersion *string `json:"ip_version,omitempty"`

	// The user-defined name for this subnet. Names must be unique within the VPC the subnet resides in. If unspecified,
	// the name will be a hyphenated list of randomly-selected words.
	//
	// +optional
	Name *string `json:"name,omitempty"`

	// The network ACL to use for this subnet.
	//
	// +optional
	NetworkACL *NetworkACLIdentity `json:"networkACL,omitempty"`

	// The public gateway to use for internet-bound traffic for this subnet. If
	// unspecified, the subnet will not be attached to a public gateway.
	//
	// +optional
	PublicGateway *PublicGatewayIdentity `json:"publicGateway,omitempty"`

	// The resource group to use. If unspecified, the account's [default resource
	// group](https://cloud.ibm.com/apidocs/resource-manager#introduction) is used.
	//
	// +immutable
	// +optional
	ResourceGroup *ResourceGroupIdentity `json:"resourceGroup,omitempty"`

	// The routing table to use for this subnet. If unspecified, the default routing table
	// for the VPC is used. The routing table properties `route_direct_link_ingress`,
	// `route_transit_gateway_ingress`, and `route_vpc_zone_ingress` must be `false`.
	//
	// +optional
	RoutingTable *RoutingTableIdentity `json:"routingTable,omitempty"`

	// The VPC the subnet is to be a part of.
	//
	// +immutable
	VPC VPCIdentity `json:"vpc"`

	// The zone this subnet will reside in.
	//
	// +immutable
	// +optional
	Zone *ZoneIdentity `json:"zone,omitempty"`

	// The IPv4 range of the subnet, expressed in CIDR format. The prefix length of the subnet's CIDR must be between `/9`
	// (8,388,608 addresses) and `/29` (8 addresses). The IPv4 range of the subnet's CIDR must fall within an existing
	// address prefix in the VPC. The subnet will be created in the zone of the address prefix that contains the IPv4 CIDR.
	// If zone is specified, it must match the zone of the address prefix that contains the subnet's IPv4 CIDR.
	//
	// +immutable
	Ipv4CIDRBlock string `json:"ipv4CIDRBlock"`
}

// SubnetPrototypeSubnetByTotalCount are input params when creating a Subnet
type SubnetPrototypeSubnetByTotalCount struct {
	// The IP version(s) to support for this subnet. Only current allowable value is 'ipv4'
	//
	// +immutable
	// +optional
	IPVersion *string `json:"ip_version,omitempty"`

	// The user-defined name for this subnet. Names must be unique within the VPC the subnet resides in. If unspecified,
	// the name will be a hyphenated list of randomly-selected words.
	//
	// +optional
	Name *string `json:"name,omitempty"`

	// The network ACL to use for this subnet.
	//
	// +optional
	NetworkACL *NetworkACLIdentity `json:"networkACL,omitempty"`

	// The public gateway to use for internet-bound traffic for this subnet. If
	// unspecified, the subnet will not be attached to a public gateway.
	//
	// +optional
	PublicGateway *PublicGatewayIdentity `json:"publicGateway,omitempty"`

	// The resource group to use. If unspecified, the account's [default resource
	// group](https://cloud.ibm.com/apidocs/resource-manager#introduction) is used.
	//
	// +immutable
	// +optional
	ResourceGroup *ResourceGroupIdentity `json:"resourceGroup,omitempty"`

	// The routing table to use for this subnet. If unspecified, the default routing table
	// for the VPC is used. The routing table properties `route_direct_link_ingress`,
	// `route_transit_gateway_ingress`, and `route_vpc_zone_ingress` must be `false`.
	//
	// +optional
	RoutingTable *RoutingTableIdentity `json:"routingTable,omitempty"`

	// The VPC the subnet is to be a part of.
	//
	// +immutable
	VPC VPCIdentity `json:"vpc"`

	// The total number of IPv4 addresses required. Must be a power of 2. The VPC must have a default address prefix in the
	// specified zone, and that prefix must have a free CIDR range with at least this number of addresses.
	//
	// +immutable
	TotalIpv4AddressCount int64 `json:"totalIpv4AddressCount"`

	// The zone this subnet will reside in.
	//
	// +immutable
	Zone ZoneIdentity `json:"zone"`
}

// SubnetParameters are the subnet parameters. Only one of the members must be non-nil
type SubnetParameters struct {
	// First way to specify the subnet
	//
	// +immutable
	// +optional
	ByTocalCount *SubnetPrototypeSubnetByTotalCount `json:"byTocalCount,omitempty"`

	// Second way to specify the subnet
	//
	// +immutable
	// _optional
	ByCIDR *SubnetPrototypeSubnetByCIDR `json:"byCIDR,omitempty"`
}

// SubnetSpec is the desired end-state of a subnet in the IBM cloud
type SubnetSpec struct {
	runtimev1alpha1.ResourceSpec `json:",inline"`

	// Info the IBM cloud needs to create a subnet
	ForProvider SubnetParameters `json:"forProvider"`
}

// PublicGatewayReferenceDeleted : If present, this property indicates the referenced resource has been deleted and provides some supplementary
// information.
type PublicGatewayReferenceDeleted struct {
	// Link to documentation about deleted resources.
	MoreInfo string `json:"moreInfo"`
}

// PublicGatewayReference ...
type PublicGatewayReference struct {
	// The CRN for this public gateway.
	CRN string `json:"crn,omitempty"`

	// If present, this property indicates the referenced resource has been deleted and provides
	// some supplementary information.
	Deleted *PublicGatewayReferenceDeleted `json:"deleted,omitempty"`

	// The URL for this public gateway.
	Href string `json:"href,omitempty"`

	// The unique identifier for this public gateway.
	ID string `json:"id,omitempty"`

	// The user-defined name for this public gateway.
	Name string `json:"name,omitempty"`

	// The resource type.
	ResourceType string `json:"resourceType,omitempty"`
}

// VPCReferenceDeleted  If present, this property indicates the referenced resource has been deleted and provides some supplementary
// information.
type VPCReferenceDeleted struct {
	// Link to documentation about deleted resources.
	MoreInfo string `json:"moreInfo,omitempty"`
}

// VPCReference ...
type VPCReference struct {
	// The CRN for this VPC.
	CRN string `json:"crn,omitempty"`

	// If present, this property indicates the referenced resource has been deleted and provides
	// some supplementary information.
	Deleted *VPCReferenceDeleted `json:"deleted,omitempty"`

	// The URL for this VPC.
	Href string `json:"href,omitempty"`

	// The unique identifier for this VPC.
	ID string `json:"id,omitempty"`

	// The unique user-defined name for this VPC.
	Name string `json:"name,omitempty"`
}

// SubnetObservation ...what comes back from the IBM cloud
type SubnetObservation struct {
	// The number of IPv4 addresses in this subnet that are not in-use, and have not been reserved by the user or the
	// provider.
	AvailableIpv4AddressCount int64 `json:"availableIpv4AddressCount"`

	// The date and time that the subnet was created.
	CreatedAt *metav1.Time `json:"createdAt"`

	// The CRN for this subnet.
	CRN string `json:"crn"`

	// The URL for this subnet.
	Href string `json:"href"`

	// The unique identifier for this subnet.
	ID string `json:"id"`

	// The IP version(s) supported by this subnet.
	IPVersion string `json:"ipVersion"`

	// The IPv4 range of the subnet, expressed in CIDR format.
	Ipv4CIDRBlock string `json:"opv4CIDRBlock,omitempty"`

	// The user-defined name for this subnet.
	Name string `json:"name"`

	// The network ACL for this subnet.
	NetworkACL NetworkACLReference `json:"NetworkACL,omitempty"`

	// The public gateway to use for internet-bound traffic for this subnet.
	PublicGateway PublicGatewayReference `json:"publicGateway,omitempty"`

	// The resource group for this subnet.
	ResourceGroup ResourceGroupReference `json:"resourceGroup"`

	// The routing table for this subnet.
	RoutingTable RoutingTableReference `json:"routingTable,omitempty"`

	// The status of the subnet.
	Status string `json:"status"`

	// The total number of IPv4 addresses in this subnet.
	//
	// Note: This is calculated as 2<sup>(32 − prefix length)</sup>. For example, the prefix length `/24` gives:<br>
	// 2<sup>(32 − 24)</sup> = 2<sup>8</sup> = 256 addresses.
	TotalIpv4AddressCount int64 `json:"totalIpv4AddressCount,omitempty"`

	// The VPC this subnet is a part of.
	VPC VPCReference `json:"vpc"`

	// The zone this subnet resides in.
	Zone ZoneReference `json:"zone"`
}

// SubnetStatus - whatever the status is (the IBM cloud decides that)
type SubnetStatus struct {
	runtimev1alpha1.ResourceStatus `json:",inline"`

	// Info the IBM cloud returns about a subnet
	AtProvider SubnetObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// Subnet contains all the info (spec + status) for a Subnet
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="STATE",type="string",JSONPath=".status.atProvider.status"
// +kubebuilder:printcolumn:name="CLASS",type="string",JSONPath=".spec.classRef.name"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,ibmcloud}
type Subnet struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SubnetSpec   `json:"spec"`
	Status SubnetStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// SubnetList - list of existing subnets...
type SubnetList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	// List of subnets returned
	Items []Subnet `json:"subnets"`
}
