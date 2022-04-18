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

package subnet

import (
	ibmVPC "github.com/IBM/vpc-go-sdk/vpcv1"

	"github.com/crossplane-contrib/provider-ibm-cloud/apis/vpcv1/v1alpha1"
	ibmc "github.com/crossplane-contrib/provider-ibm-cloud/pkg/clients"

	"github.com/crossplane/crossplane-runtime/pkg/reference"
)

var (
	// Values below are the ones that will be used if we decide that some parameters has to be non-nil
	availableIpv4AddressCount = int64(ibmc.RandomInt(30))
	randomName                = ibmc.RandomString(true)
	// CrnVal is public b/c need to used it in the controller
	CrnVal                       = ibmc.RandomString(true)
	hrefVal                      = ibmc.RandomString(false)
	idVal                        = ibmc.RandomString(true)
	ipVersionVal                 = ibmc.RandomString(false)
	ipv4CIDRBlockVal             = ibmc.RandomString(false)
	statusVal                    = ibmc.RandomString(false)
	totalIpv4AddressCount        = int64(ibmc.RandomInt(30))
	createdAtVal                 = ibmc.ADateTimeInAYear(2012)
	networkACLCRN                = ibmc.RandomString(false)
	networkACLHref               = ibmc.RandomString(false)
	networkACLID                 = ibmc.RandomString(true)
	networkACLName               = ibmc.RandomString(false)
	networkACLDeletedMoreInfo    = ibmc.RandomString(false)
	publicGatewayCRN             = ibmc.RandomString(false)
	publicGatewayHref            = ibmc.RandomString(false)
	publicGatewayID              = ibmc.RandomString(true)
	publicGatewayName            = ibmc.RandomString(false)
	publicGatewayDeletedMoreInfo = ibmc.RandomString(false)
	routingTableDeletedMoreInfo  = ibmc.RandomString(false)
	routingTableHref             = ibmc.RandomString(false)
	routingTableID               = ibmc.RandomString(true)
	routingTableName             = ibmc.RandomString(false)
	routingTableResourceType     = ibmc.RandomString(false)
	resourceGroupName            = ibmc.RandomString(false)
	resourceGroupHref            = ibmc.RandomString(false)
	resourceGroupID              = ibmc.RandomString(true)
	vpcReferenceDeletedMoreInfo  = ibmc.RandomString(false)
	vpcReferenceHref             = ibmc.RandomString(false)
	vpcReferenceID               = ibmc.RandomString(true)
	vpcReferenceName             = ibmc.RandomString(false)
	vpcReferenceCRN              = ibmc.RandomString(false)
	zoneReferenceHref            = ibmc.RandomString(false)
	zoneReferenceName            = ibmc.RandomString(true)
)

// GetDummyObservation returns a dummy object, ready to be used in create-subnet-in-the-cloud request. Non-nil values will be the
// ones of the local constants above.
//
// Params
//    isByTotalCount - determines what kind of observation we get
//	  each other param is used in controlling the value of the similarly-named field in the returned structure.
//
// Returns
//	    an object appropriately populated
func GetDummyObservation( // nolint:gocyclo
	isByTotalCount bool,
	createdAtNil bool,
	crnNil bool,
	hrefNil bool,
	idNil bool,
	ipVersionNil bool,
	ipv4CIDRBlockNil bool,
	nameNil bool,
	networkACLNil bool,
	networkACLCRNNil bool,
	networkACLDeletedNil bool,
	networkACLDeletedMoreInfoNil bool,
	networkACLHrefNil bool,
	networkACLIDNil bool,
	networkACLNameNil bool,
	publicGatewayNil bool,
	publicGatewayCRNNil bool,
	publicGatewayDeletedNil bool,
	publicGatewayDeletedMoreInfoNil bool,
	publicGatewayHrefNil bool,
	publicGatewayIDNil bool,
	publicGatewayNameNil bool,
	publicGatewayResourceTypeNil bool,
	resourceGroupNil bool,
	resourceGroupNameNil bool,
	resourceGroupHrefNil bool,
	resourceGroupIDNil bool,
	routingTableNil bool,
	routingTableDeletedNil bool,
	routingTableDeletedMoreInfoNil,
	routingTableHrefNil bool,
	routingTableIDNil bool,
	routingTableNameNil bool,
	routingTableResourceTypeNil bool,
	statusNil bool,
	totalIpv4AddressCountNil bool,
	vpcReferenceNil bool,
	vpcReferenceDeletedNil bool,
	vpcReferenceDeletedMoreInfoNil,
	vpcReferenceHrefNil bool,
	vpcReferenceIDNil bool,
	vpcReferenceNameNil bool,
	vpcReferenceCRNNil bool,
	zoneReferenceNil bool,
	zoneReferenceHrefNil bool,
	zoneReferenceNameNil bool) ibmVPC.Subnet {

	result := ibmVPC.Subnet{
		AvailableIpv4AddressCount: &availableIpv4AddressCount,
		CreatedAt:                 ibmc.ReturnConditionalDate(!createdAtNil, createdAtVal),
		CRN:                       reference.ToPtrValue(CrnVal),
		Href:                      ibmc.ReturnConditionalStr(!hrefNil, hrefVal),
		ID:                        reference.ToPtrValue(idVal),
		IPVersion:                 reference.ToPtrValue(ipVersionVal),
		Name:                      reference.ToPtrValue(randomName),
		Status:                    ibmc.ReturnConditionalStr(!statusNil, statusVal),
	}

	if isByTotalCount {
		result.TotalIpv4AddressCount = &totalIpv4AddressCount
	} else {
		result.Ipv4CIDRBlock = reference.ToPtrValue(ipv4CIDRBlockVal)
	}

	if !networkACLNil {
		result.NetworkACL = &ibmVPC.NetworkACLReference{
			CRN:  ibmc.ReturnConditionalStr(!networkACLCRNNil, networkACLCRN),
			Href: ibmc.ReturnConditionalStr(!networkACLHrefNil, networkACLHref),
			ID:   ibmc.ReturnConditionalStr(!networkACLIDNil, networkACLID),
			Name: ibmc.ReturnConditionalStr(!networkACLNameNil, networkACLName),
		}

		if !networkACLDeletedNil {
			result.NetworkACL.Deleted = &ibmVPC.NetworkACLReferenceDeleted{
				MoreInfo: ibmc.ReturnConditionalStr(!networkACLDeletedMoreInfoNil, networkACLDeletedMoreInfo),
			}
		}
	}

	if !publicGatewayNil {
		result.PublicGateway = &ibmVPC.PublicGatewayReference{
			CRN:  ibmc.ReturnConditionalStr(!publicGatewayCRNNil, publicGatewayCRN),
			Href: ibmc.ReturnConditionalStr(!publicGatewayHrefNil, publicGatewayHref),
			ID:   ibmc.ReturnConditionalStr(!publicGatewayIDNil, publicGatewayID),
			Name: ibmc.ReturnConditionalStr(!publicGatewayNameNil, publicGatewayName),
		}

		if !publicGatewayDeletedNil {
			result.PublicGateway.Deleted = &ibmVPC.PublicGatewayReferenceDeleted{
				MoreInfo: ibmc.ReturnConditionalStr(!publicGatewayDeletedMoreInfoNil, publicGatewayDeletedMoreInfo),
			}
		}
	}

	if !resourceGroupNil {
		result.ResourceGroup = &ibmVPC.ResourceGroupReference{
			Href: ibmc.ReturnConditionalStr(!resourceGroupHrefNil, resourceGroupHref),
			ID:   ibmc.ReturnConditionalStr(!resourceGroupIDNil, resourceGroupID),
			Name: ibmc.ReturnConditionalStr(!resourceGroupNameNil, resourceGroupName),
		}
	}

	if !routingTableNil {
		result.RoutingTable = &ibmVPC.RoutingTableReference{
			Href:         ibmc.ReturnConditionalStr(!routingTableHrefNil, routingTableHref),
			ID:           ibmc.ReturnConditionalStr(!routingTableIDNil, routingTableID),
			Name:         ibmc.ReturnConditionalStr(!routingTableNameNil, routingTableName),
			ResourceType: ibmc.ReturnConditionalStr(!routingTableResourceTypeNil, routingTableResourceType),
		}

		if !routingTableDeletedNil {
			result.RoutingTable.Deleted = &ibmVPC.RoutingTableReferenceDeleted{
				MoreInfo: ibmc.ReturnConditionalStr(!routingTableDeletedMoreInfoNil, routingTableDeletedMoreInfo),
			}
		}
	}

	if !vpcReferenceNil {
		result.VPC = &ibmVPC.VPCReference{
			Href: ibmc.ReturnConditionalStr(!vpcReferenceHrefNil, vpcReferenceHref),
			ID:   ibmc.ReturnConditionalStr(!vpcReferenceIDNil, vpcReferenceID),
			Name: ibmc.ReturnConditionalStr(!vpcReferenceNameNil, vpcReferenceName),
			CRN:  ibmc.ReturnConditionalStr(!vpcReferenceCRNNil, vpcReferenceCRN),
		}

		if !vpcReferenceDeletedNil {
			result.VPC.Deleted = &ibmVPC.VPCReferenceDeleted{
				MoreInfo: ibmc.ReturnConditionalStr(!vpcReferenceDeletedMoreInfoNil, vpcReferenceDeletedMoreInfo),
			}
		}
	}

	if !zoneReferenceNil {
		result.Zone = &ibmVPC.ZoneReference{
			Href: ibmc.ReturnConditionalStr(!zoneReferenceHrefNil, zoneReferenceHref),
			Name: ibmc.ReturnConditionalStr(!zoneReferenceNameNil, zoneReferenceName),
		}
	}

	return result
}

// GetDummyCreateParams returns a dummy object, ready to be used in a create-VPC-in-k8s request. Non-nil values will be the
// ones of the local constants above
//
// Params
//      byTotalCount - whether the returned object is "ByTotalCount" or "ByCIDR"
//		ipVersionNil - whether to set the 'IPVersion' member to nil
// 		nameNil - whether to set the 'Name' member to nil
//      networkACLNil - whether to set the 'NetworkACL' member to nil
//      publicGatewayNil - whether to set the 'PublicGateway' member to nil..
//		resourceGroupNil - whether to set the 'ResourceGroup' member to nil
//		routingTableNil - whether to set the 'RoutingTable' member to nil
//      totalIpv4AddressCountNil - whether to set the 'TotalIpv4AddressCount' member to nil (does not apply if the returned object is 'ByTotalCount')
//      zoneNil - whether to set the 'Zone' member to nil  (does not apply if the returned object is 'ByTotalCount')
//      ipv4CIDRBlockNil - whether to set the 'Ipv4CIDRBlockNil' member to nil (does not apply if the returned object is 'ByCIDR')
//
// Returns
//	    an object appropriately populated
func GetDummyCreateParams(
	byTotalCount bool,
	ipVersionNil bool,
	nameNil bool,
	networkACLNil bool,
	publicGatewayNil bool,
	resourceGroupNil bool,
	routingTableNil bool,
	totalIpv4AddressCountNil bool,
	zoneNil bool,
	ipv4CIDRBlockNil bool) v1alpha1.SubnetParameters {

	var networkACL *v1alpha1.NetworkACLIdentity
	var publicGateway *v1alpha1.PublicGatewayIdentity
	var resourceGroup *v1alpha1.ResourceGroupIdentity
	var routingTable *v1alpha1.RoutingTableIdentity
	var zone *v1alpha1.ZoneIdentity

	if !networkACLNil {
		networkACL = &v1alpha1.NetworkACLIdentity{
			ID: networkACLID,
		}
	}

	if !publicGatewayNil {
		publicGateway = &v1alpha1.PublicGatewayIdentity{
			ID: publicGatewayID,
		}
	}

	if !resourceGroupNil {
		resourceGroup = &v1alpha1.ResourceGroupIdentity{
			ID: resourceGroupID,
		}
	}

	if !routingTableNil {
		routingTable = &v1alpha1.RoutingTableIdentity{
			ID: routingTableID,
		}
	}

	if !zoneNil {
		zone = &v1alpha1.ZoneIdentity{
			Name: zoneReferenceName,
		}
	}

	result := v1alpha1.SubnetParameters{}
	if byTotalCount {
		result.ByTocalCount = &v1alpha1.SubnetPrototypeSubnetByTotalCount{
			IPVersion:     ibmc.ReturnConditionalStr(!ipVersionNil, ipVersionVal),
			Name:          ibmc.ReturnConditionalStr(!nameNil, randomName),
			NetworkACL:    networkACL,
			PublicGateway: publicGateway,
			ResourceGroup: resourceGroup,
			RoutingTable:  routingTable,
			VPC: v1alpha1.VPCIdentity{
				ID: &vpcReferenceID,
			},
			TotalIpv4AddressCount: totalIpv4AddressCount,
			Zone: v1alpha1.ZoneIdentity{
				Name: zoneReferenceName,
			},
		}
	} else {
		result.ByCIDR = &v1alpha1.SubnetPrototypeSubnetByCIDR{
			IPVersion:     ibmc.ReturnConditionalStr(!ipVersionNil, ipVersionVal),
			Name:          ibmc.ReturnConditionalStr(!nameNil, randomName),
			NetworkACL:    networkACL,
			PublicGateway: publicGateway,
			ResourceGroup: resourceGroup,
			RoutingTable:  routingTable,
			VPC: v1alpha1.VPCIdentity{
				ID: &vpcReferenceID,
			},
			Zone:          zone,
			Ipv4CIDRBlock: ipv4CIDRBlockVal,
		}
	}

	return result
}
