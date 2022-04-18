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
	"github.com/google/go-cmp/cmp"

	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/reference"

	"github.com/crossplane-contrib/provider-ibm-cloud/apis/vpcv1/v1alpha1"
	ibmc "github.com/crossplane-contrib/provider-ibm-cloud/pkg/clients"
)

// Contastants are keys that will be sent to the server when updating a subnet
const (
	nameKey          = "name"
	idKey            = "id"
	networkACLKey    = "network_acl"
	publicGatewayKey = "public_gateway"
	routingTableKey  = "routing_table"
)

// Params
//    someStrings - one or more strings
//
// Returns
//    the  number of non-null strings
func numNonNul(someStrings ...*string) int {
	result := 0
	for _, aStr := range someStrings {
		if aStr != nil {
			result++
		}
	}

	return result
}

// Params
//    aStr - a pointer to string - potentially nil
//
// Returns
//    a new string with a copy of the value of the param (or nil)
func copyVal(aStr *string) *string {
	var result *string

	if aStr != nil {
		result = new(string)

		*result = *aStr
	}

	return result
}

// Params
//    aVal - a value. Nay be nil
//    anotherVal - another value
//    changedTillNow - Cannot be nil. if true, it leaves it alone. If false, it changes it to true
//                     "if the return value is != aVal"
//
// Returns
//    if aVal == nil and anotherVal != "", it returns (anotherVal, true). O/w it returns (aVal, false)
func assignIfAppropriateStr(aVal *string, anotherVal string, changedTillNow *bool) string {
	var result string

	if aVal == nil && anotherVal != "" {
		result = anotherVal

		*changedTillNow = true
	} else {
		result = reference.FromPtrValue(aVal)
	}

	return result
}

// Params
//    aVal - a value. May be nil
//    anotherVal - another value. May be nil
//    changedTillNow - Cannot be nil. if true, it leaves it alone. If false, it changes it to true
//                     "if the return value is != aVal"
//
// Returns
//    if aVal == nil and anotherVal != nil, it returns (a copy of) anotherVal; o/w it return aVal
func assignIfAppropriatePtr(aVal *string, anotherVal *string, changedTillNow *bool) *string {
	var result *string

	if aVal == nil && anotherVal != nil {
		result = copyVal(anotherVal)

		*changedTillNow = true
	} else {
		result = aVal
	}

	return result
}

// Params
//     spec - the spec
//
// Returns
//     specIPVersion, specName, specNetworkACL, specPublicGateway, specResourceGroup, specRoutingTable, specVPC, specTotalIpv4AddressCount, specZone, specIpv4CIDRBlock
func getParameters(spec *v1alpha1.SubnetParameters) (*string, *string, *v1alpha1.NetworkACLIdentity, *v1alpha1.PublicGatewayIdentity,
	*v1alpha1.ResourceGroupIdentity, *v1alpha1.RoutingTableIdentity, v1alpha1.VPCIdentity, *int64, *v1alpha1.ZoneIdentity, *string) {
	var specIPVersion *string
	var specName *string
	var specNetworkACL *v1alpha1.NetworkACLIdentity
	var specPublicGateway *v1alpha1.PublicGatewayIdentity
	var specResourceGroup *v1alpha1.ResourceGroupIdentity
	var specRoutingTable *v1alpha1.RoutingTableIdentity
	var specVPC v1alpha1.VPCIdentity
	var specTotalIpv4AddressCount *int64
	var specZone *v1alpha1.ZoneIdentity
	var specIpv4CIDRBlock *string

	if spec.ByTocalCount != nil {
		specIPVersion = spec.ByTocalCount.IPVersion
		specName = spec.ByTocalCount.Name
		specNetworkACL = spec.ByTocalCount.NetworkACL
		specPublicGateway = spec.ByTocalCount.PublicGateway
		specResourceGroup = spec.ByTocalCount.ResourceGroup
		specRoutingTable = spec.ByTocalCount.RoutingTable
		specVPC = spec.ByTocalCount.VPC
		specTotalIpv4AddressCount = &spec.ByTocalCount.TotalIpv4AddressCount
		specZone = &spec.ByTocalCount.Zone
	} else {
		specIPVersion = spec.ByCIDR.IPVersion
		specName = spec.ByCIDR.Name
		specNetworkACL = spec.ByCIDR.NetworkACL
		specPublicGateway = spec.ByCIDR.PublicGateway
		specResourceGroup = spec.ByCIDR.ResourceGroup
		specRoutingTable = spec.ByCIDR.RoutingTable
		specVPC = spec.ByCIDR.VPC
		specZone = spec.ByCIDR.Zone
		specIpv4CIDRBlock = &spec.ByCIDR.Ipv4CIDRBlock
	}

	return specIPVersion, specName, specNetworkACL, specPublicGateway, specResourceGroup, specRoutingTable, specVPC, specTotalIpv4AddressCount, specZone, specIpv4CIDRBlock
}

// LateInitializeSpec fills optional and unassigned fields with the values in the spec, from the info that comes from the cloud
//
// Params
// 	  spec - what we get from k8s
// 	  fromIBMCloud - ...what comes from the cloud
//
// Returns
//    whether late initialization happened
//    currently, always nil
func LateInitializeSpec(spec *v1alpha1.SubnetParameters, fromIBMCloud *ibmVPC.Subnet) (bool, error) { // nolint:gocyclo
	result := false

	specIPVersion, specName, specNetworkACL, specPublicGateway, specResourceGroup,
		specRoutingTable, specVPC, specTotalIpv4AddressCount, specZone, specIpv4CIDRBlock := getParameters(spec)

	specIPVersion = assignIfAppropriatePtr(specIPVersion, fromIBMCloud.IPVersion, &result)
	specName = assignIfAppropriatePtr(specName, fromIBMCloud.Name, &result)

	if fromIBMCloud.NetworkACL != nil && fromIBMCloud.NetworkACL.ID != nil {
		if specNetworkACL == nil {
			specNetworkACL = &v1alpha1.NetworkACLIdentity{
				ID: *fromIBMCloud.NetworkACL.ID,
			}

			result = true
		} else {
			specNetworkACL.ID = assignIfAppropriateStr(&specNetworkACL.ID, *fromIBMCloud.NetworkACL.ID, &result)
		}
	}

	if fromIBMCloud.PublicGateway != nil && fromIBMCloud.PublicGateway.ID != nil {
		if specPublicGateway == nil {
			specPublicGateway = &v1alpha1.PublicGatewayIdentity{
				ID: *fromIBMCloud.PublicGateway.ID,
			}

			result = true
		} else {
			specPublicGateway.ID = assignIfAppropriateStr(&specPublicGateway.ID, *fromIBMCloud.PublicGateway.ID, &result)
		}
	}

	if fromIBMCloud.ResourceGroup != nil && fromIBMCloud.ResourceGroup.ID != nil {
		if specResourceGroup == nil {
			specResourceGroup = &v1alpha1.ResourceGroupIdentity{
				ID: reference.FromPtrValue(fromIBMCloud.ResourceGroup.ID),
			}

			result = true
		} else {
			specResourceGroup.ID = assignIfAppropriateStr(&specResourceGroup.ID, *fromIBMCloud.ResourceGroup.ID, &result)
		}
	}

	if fromIBMCloud.RoutingTable != nil && fromIBMCloud.RoutingTable.ID != nil {
		if specRoutingTable == nil {
			specRoutingTable = &v1alpha1.RoutingTableIdentity{
				ID: *fromIBMCloud.RoutingTable.ID,
			}

			result = true
		} else {
			specRoutingTable.ID = assignIfAppropriateStr(&specRoutingTable.ID, *fromIBMCloud.RoutingTable.ID, &result)
		}
	}

	if fromIBMCloud.VPC != nil && fromIBMCloud.VPC.ID != nil {
		specVPC.ID = assignIfAppropriatePtr(specVPC.ID, fromIBMCloud.VPC.ID, &result)
	}

	if spec.ByTocalCount != nil {
		if fromIBMCloud.TotalIpv4AddressCount != nil && specTotalIpv4AddressCount == nil {
			specTotalIpv4AddressCount = new(int64)
			*specTotalIpv4AddressCount = *fromIBMCloud.TotalIpv4AddressCount

			result = true
		}
	}

	if fromIBMCloud.Zone != nil && fromIBMCloud.Zone.Name != nil {
		if specZone == nil {
			specZone = &v1alpha1.ZoneIdentity{
				Name: *fromIBMCloud.Zone.Name,
			}

			result = true
		} else {
			specZone.Name = assignIfAppropriateStr(&specZone.Name, *fromIBMCloud.Zone.Name, &result)
		}
	}

	if spec.ByCIDR != nil {
		specIpv4CIDRBlock = assignIfAppropriatePtr(specIpv4CIDRBlock, fromIBMCloud.Ipv4CIDRBlock, &result)
	}

	if spec.ByTocalCount != nil {
		spec.ByTocalCount.IPVersion = specIPVersion
		spec.ByTocalCount.Name = specName
		spec.ByTocalCount.NetworkACL = specNetworkACL
		spec.ByTocalCount.PublicGateway = specPublicGateway
		spec.ByTocalCount.ResourceGroup = specResourceGroup
		spec.ByTocalCount.RoutingTable = specRoutingTable
		spec.ByTocalCount.VPC = specVPC
		spec.ByTocalCount.TotalIpv4AddressCount = *specTotalIpv4AddressCount
		spec.ByTocalCount.Zone = *specZone
	} else {
		spec.ByCIDR.IPVersion = specIPVersion
		spec.ByCIDR.Name = specName
		spec.ByCIDR.NetworkACL = specNetworkACL
		spec.ByCIDR.PublicGateway = specPublicGateway
		spec.ByCIDR.ResourceGroup = specResourceGroup
		spec.ByCIDR.RoutingTable = specRoutingTable
		spec.ByCIDR.VPC = specVPC
		spec.ByCIDR.Zone = specZone
		spec.ByCIDR.Ipv4CIDRBlock = *specIpv4CIDRBlock
	}

	return result, nil
}

// GenerateObservation returns a crossplane version of the cloud observation results parameters
//
// Params
//     in - the create options, in IBM-cloud-style
//
// Returns
//     the status, crossplane-style
func GenerateObservation(in *ibmVPC.Subnet) (v1alpha1.SubnetObservation, error) { // nolint:gocyclo
	result := v1alpha1.SubnetObservation{
		CreatedAt:     ibmc.DateTimeToMetaV1Time(in.CreatedAt),
		CRN:           reference.FromPtrValue(in.CRN),
		Href:          reference.FromPtrValue(in.Href),
		ID:            reference.FromPtrValue(in.ID),
		IPVersion:     reference.FromPtrValue(in.IPVersion),
		Ipv4CIDRBlock: reference.FromPtrValue(in.Ipv4CIDRBlock),
		Name:          reference.FromPtrValue(in.Name),
		Status:        reference.FromPtrValue(in.Status),
	}

	if in.AvailableIpv4AddressCount != nil {
		result.AvailableIpv4AddressCount = *in.AvailableIpv4AddressCount
	}

	if in.TotalIpv4AddressCount != nil {
		result.TotalIpv4AddressCount = *in.TotalIpv4AddressCount
	}

	if in.NetworkACL != nil {
		if numNonNul(in.NetworkACL.CRN, in.NetworkACL.Href, in.NetworkACL.ID, in.NetworkACL.Name) > 0 ||
			(in.NetworkACL.Deleted != nil && in.NetworkACL.Deleted.MoreInfo != nil) {
			result.NetworkACL = v1alpha1.NetworkACLReference{
				CRN:  reference.FromPtrValue(in.NetworkACL.CRN),
				Href: reference.FromPtrValue(in.NetworkACL.Href),
				ID:   reference.FromPtrValue(in.NetworkACL.ID),
				Name: reference.FromPtrValue(in.NetworkACL.Name),
			}

			if in.NetworkACL.Deleted != nil && in.NetworkACL.Deleted.MoreInfo != nil {
				result.NetworkACL.Deleted = &v1alpha1.NetworkACLReferenceDeleted{
					MoreInfo: reference.FromPtrValue(in.NetworkACL.Deleted.MoreInfo),
				}
			}
		}
	}

	if in.PublicGateway != nil {
		if numNonNul(in.PublicGateway.CRN, in.PublicGateway.Href, in.PublicGateway.ID, in.PublicGateway.Name, in.PublicGateway.ResourceType) > 0 ||
			(in.PublicGateway.Deleted != nil && in.PublicGateway.Deleted.MoreInfo != nil) {
			result.PublicGateway = v1alpha1.PublicGatewayReference{
				CRN:          reference.FromPtrValue(in.PublicGateway.CRN),
				Href:         reference.FromPtrValue(in.PublicGateway.Href),
				ID:           reference.FromPtrValue(in.PublicGateway.ID),
				Name:         reference.FromPtrValue(in.PublicGateway.Name),
				ResourceType: reference.FromPtrValue(in.PublicGateway.ResourceType),
			}

			if in.PublicGateway.Deleted != nil && in.PublicGateway.Deleted.MoreInfo != nil {
				result.PublicGateway.Deleted = &v1alpha1.PublicGatewayReferenceDeleted{
					MoreInfo: reference.FromPtrValue(in.PublicGateway.Deleted.MoreInfo),
				}
			}
		}
	}

	if in.ResourceGroup != nil {
		if numNonNul(in.ResourceGroup.Name, in.ResourceGroup.Href, in.ResourceGroup.ID) > 0 {
			result.ResourceGroup = v1alpha1.ResourceGroupReference{
				Name: reference.FromPtrValue(in.ResourceGroup.Name),
				Href: reference.FromPtrValue(in.ResourceGroup.Href),
				ID:   reference.FromPtrValue(in.ResourceGroup.ID),
			}
		}
	}

	if in.RoutingTable != nil {
		if numNonNul(in.RoutingTable.Href, in.RoutingTable.ID, in.RoutingTable.Name, in.RoutingTable.ResourceType) > 0 ||
			(in.RoutingTable.Deleted != nil && in.RoutingTable.Deleted.MoreInfo != nil) {
			result.RoutingTable = v1alpha1.RoutingTableReference{
				Href:         reference.FromPtrValue(in.RoutingTable.Href),
				ID:           reference.FromPtrValue(in.RoutingTable.ID),
				Name:         reference.FromPtrValue(in.RoutingTable.Name),
				ResourceType: reference.FromPtrValue(in.RoutingTable.ResourceType),
			}

			if in.RoutingTable.Deleted != nil && in.RoutingTable.Deleted.MoreInfo != nil {
				result.RoutingTable.Deleted = &v1alpha1.RoutingTableReferenceDeleted{
					MoreInfo: reference.FromPtrValue(in.RoutingTable.Deleted.MoreInfo),
				}
			}
		}
	}

	if in.VPC != nil {
		if numNonNul(in.VPC.CRN, in.VPC.Href, in.VPC.ID, in.VPC.Name) > 0 ||
			(in.VPC.Deleted != nil && in.VPC.Deleted.MoreInfo != nil) {
			result.VPC = v1alpha1.VPCReference{
				CRN:  reference.FromPtrValue(in.VPC.CRN),
				Href: reference.FromPtrValue(in.VPC.Href),
				ID:   reference.FromPtrValue(in.VPC.ID),
				Name: reference.FromPtrValue(in.VPC.Name),
			}

			if in.VPC.Deleted != nil && in.VPC.Deleted.MoreInfo != nil {
				result.VPC.Deleted = &v1alpha1.VPCReferenceDeleted{
					MoreInfo: reference.FromPtrValue(in.VPC.Deleted.MoreInfo),
				}
			}
		}
	}

	if in.Zone != nil {
		if numNonNul(in.Zone.Href, in.Zone.Name) > 0 {
			result.Zone = v1alpha1.ZoneReference{
				Href: reference.FromPtrValue(in.Zone.Href),
				Name: reference.FromPtrValue(in.Zone.Name),
			}
		}
	}

	return result, nil
}

// GenerateCreateSubnetParameters gives us a create-from-crossplane version of the cloud observation results parameters
//
// Params
//     isByTotalCount - whether the subnet was created via total address count
//     in - the create options, in IBM-cloud-style
//
// Returns
//     the params, crossplane-style
func GenerateCreateSubnetParameters(isByTotalCount bool, in *ibmVPC.Subnet) (v1alpha1.SubnetParameters, error) { // nolint:gocyclo
	ipVersion := copyVal(in.IPVersion)
	name := copyVal(in.Name)
	var networkACL *v1alpha1.NetworkACLIdentity
	var publicGateway *v1alpha1.PublicGatewayIdentity
	var resourceGroup *v1alpha1.ResourceGroupIdentity
	var routingTable *v1alpha1.RoutingTableIdentity
	var vpc v1alpha1.VPCIdentity
	var totalIpv4AddressCount *int64
	var zone *v1alpha1.ZoneIdentity
	var iPv4CIDRBlock *string

	if in.NetworkACL != nil && in.NetworkACL.ID != nil {
		networkACL = &v1alpha1.NetworkACLIdentity{
			ID: *in.NetworkACL.ID,
		}
	}

	if in.PublicGateway != nil && in.PublicGateway.ID != nil {
		publicGateway = &v1alpha1.PublicGatewayIdentity{
			ID: *in.PublicGateway.ID,
		}
	}

	if in.ResourceGroup != nil && in.ResourceGroup.ID != nil {
		resourceGroup = &v1alpha1.ResourceGroupIdentity{
			ID: reference.FromPtrValue(in.ResourceGroup.ID),
		}
	}

	if in.RoutingTable != nil && in.RoutingTable.ID != nil {
		routingTable = &v1alpha1.RoutingTableIdentity{
			ID: *in.RoutingTable.ID,
		}
	}

	if in.VPC != nil {
		vpc = v1alpha1.VPCIdentity{
			ID: copyVal(in.VPC.ID),
		}
	}

	if in.TotalIpv4AddressCount != nil {
		totalIpv4AddressCount = new(int64)
		*totalIpv4AddressCount = *in.TotalIpv4AddressCount
	}

	if in.Zone != nil && in.Zone.Name != nil {
		zone = &v1alpha1.ZoneIdentity{
			Name: *in.Zone.Name,
		}
	}

	iPv4CIDRBlock = copyVal(in.Ipv4CIDRBlock)

	result := v1alpha1.SubnetParameters{}

	if isByTotalCount {
		result.ByTocalCount = &v1alpha1.SubnetPrototypeSubnetByTotalCount{
			IPVersion:             ipVersion,
			Name:                  name,
			NetworkACL:            networkACL,
			PublicGateway:         publicGateway,
			ResourceGroup:         resourceGroup,
			RoutingTable:          routingTable,
			VPC:                   vpc,
			TotalIpv4AddressCount: *totalIpv4AddressCount,
			Zone:                  *zone,
		}
	} else {
		result.ByCIDR = &v1alpha1.SubnetPrototypeSubnetByCIDR{
			IPVersion:     ipVersion,
			Name:          name,
			NetworkACL:    networkACL,
			PublicGateway: publicGateway,
			ResourceGroup: resourceGroup,
			RoutingTable:  routingTable,
			VPC:           vpc,
			Zone:          zone,
			Ipv4CIDRBlock: *iPv4CIDRBlock,
		}
	}

	return result, nil
}

// GenerateCreateOptions returns a cloud-compliant version of the crossplane creation parameters
//
// Params
//    in - the creation options, crossplane style
//
//  Returns
//     the struct to use in the cloud call
//     error - always nil for now
func GenerateCreateOptions(in *v1alpha1.SubnetParameters) (ibmVPC.CreateSubnetOptions, error) { // nolint:gocyclo
	dcIPVersion, dcName, dcNetworkACL, dcPublicGateway, dcResourceGroup,
		dcRoutingTable, dcVPC, dcTotalIpv4AddressCount, dcZone, dcIpv4CIDRBlock := getParameters(in.DeepCopy())

	var cloudNetworkACL ibmVPC.NetworkACLIdentityIntf
	var cloudPublicGateway ibmVPC.PublicGatewayIdentityIntf
	var cloudResourceGroup ibmVPC.ResourceGroupIdentityIntf
	var cloudRoutingTable ibmVPC.RoutingTableIdentityIntf
	var cloudVPC ibmVPC.VPCIdentityIntf
	var cloudZone ibmVPC.ZoneIdentityIntf

	if dcNetworkACL != nil {
		cloudNetworkACL = &ibmVPC.NetworkACLIdentityByID{
			ID: reference.ToPtrValue(dcNetworkACL.ID),
		}
	}

	if dcPublicGateway != nil {
		cloudPublicGateway = &ibmVPC.PublicGatewayIdentityPublicGatewayIdentityByID{
			ID: reference.ToPtrValue(dcPublicGateway.ID),
		}
	}

	if dcResourceGroup != nil {
		cloudResourceGroup = &ibmVPC.ResourceGroupIdentityByID{
			ID: reference.ToPtrValue(dcResourceGroup.ID),
		}
	}

	if dcRoutingTable != nil {
		cloudRoutingTable = &ibmVPC.RoutingTableIdentityByID{
			ID: reference.ToPtrValue(dcRoutingTable.ID),
		}
	}

	cloudVPC = &ibmVPC.VPCIdentityByID{
		ID: dcVPC.ID,
	}

	if dcZone != nil {
		cloudZone = &ibmVPC.ZoneIdentityByName{
			Name: reference.ToPtrValue(dcZone.Name),
		}
	}

	result := ibmVPC.CreateSubnetOptions{}

	if dcTotalIpv4AddressCount != nil {
		result.SubnetPrototype = &ibmVPC.SubnetPrototypeSubnetByTotalCount{
			IPVersion:             dcIPVersion,
			Name:                  dcName,
			NetworkACL:            cloudNetworkACL,
			PublicGateway:         cloudPublicGateway,
			ResourceGroup:         cloudResourceGroup,
			RoutingTable:          cloudRoutingTable,
			VPC:                   cloudVPC,
			TotalIpv4AddressCount: dcTotalIpv4AddressCount,
			Zone:                  cloudZone,
		}
	} else {
		result.SubnetPrototype = &ibmVPC.SubnetPrototypeSubnetByCIDR{
			IPVersion:     dcIPVersion,
			Name:          dcName,
			NetworkACL:    cloudNetworkACL,
			PublicGateway: cloudPublicGateway,
			ResourceGroup: cloudResourceGroup,
			RoutingTable:  cloudRoutingTable,
			VPC:           cloudVPC,
			Zone:          cloudZone,
			Ipv4CIDRBlock: dcIpv4CIDRBlock,
		}
	}

	return result, nil
}

// DiffPatch checks whether the current subnet config (in the cloud) is up-to-date compared to the crossplane one (after late initialization)
//
// Params
//    in - the crossplane parameters
//    observed - what came from the cloud
//
// Returns
//    - A map of the values that should get updated. Note that, in case that more than one option to identify a
//      resource exists (eg id, href, crn), only one is used
//    - error - nil for now
func DiffPatch(in *v1alpha1.SubnetParameters, observed *ibmVPC.Subnet) (map[string]interface{}, error) { // nolint:gocyclo
	result := make(map[string]interface{})

	observedParams, err := GenerateCreateSubnetParameters(in.ByTocalCount != nil, observed)
	if err != nil {
		return nil, err
	}

	_, observedName, observedNetworkACL, observedPublicGateway, _,
		observedRoutingTable, _, _, _, _ := getParameters(&observedParams)

	_, specName, specNetworkACL, specPublicGateway, _,
		specRoutingTable, _, _, _, _ := getParameters(in.DeepCopy())

	if diff := cmp.Diff(specName, observedName); diff != "" {
		result[nameKey] = specName
	}

	if diff := cmp.Diff(specNetworkACL, observedNetworkACL); diff != "" {
		networkACLMap := make(map[string]interface{})

		if specNetworkACL != nil {
			if observedNetworkACL != nil {
				if diff := cmp.Diff(specNetworkACL.ID, observedNetworkACL.ID); diff != "" {
					networkACLMap[idKey] = specNetworkACL.ID
				}
			} else {
				networkACLMap[idKey] = specNetworkACL.ID
			}
		}

		result[networkACLKey] = networkACLMap
	}

	if diff := cmp.Diff(specPublicGateway, observedPublicGateway); diff != "" {
		publicGatewayMap := make(map[string]interface{})

		if specPublicGateway != nil {
			if observedPublicGateway != nil {
				if diff := cmp.Diff(specPublicGateway.ID, observedPublicGateway.ID); diff != "" {
					publicGatewayMap[idKey] = specPublicGateway.ID
				}
			} else {
				publicGatewayMap[idKey] = specPublicGateway.ID
			}
		}

		result[publicGatewayKey] = publicGatewayMap
	}

	if diff := cmp.Diff(specRoutingTable, observedRoutingTable); diff != "" {
		routingTableMap := make(map[string]interface{})

		if specRoutingTable != nil {
			if observedRoutingTable != nil {
				if diff := cmp.Diff(specRoutingTable.ID, observedRoutingTable.ID); diff != "" {
					routingTableMap[idKey] = specRoutingTable.ID
				}
			} else {
				routingTableMap[idKey] = specRoutingTable.ID
			}
		}

		result[routingTableKey] = routingTableMap
	}

	return result, nil
}

// IsUpToDate checks whether the current subnet config (in the cloud) is up-to-date compared to the crossplane one
//
// No error is currently returned
func IsUpToDate(in *v1alpha1.SubnetParameters, observed *ibmVPC.Subnet, l logging.Logger) (bool, error) { // nolint:gocyclo
	result := true

	actualParams, err := GenerateCreateSubnetParameters(in.ByTocalCount != nil, observed)
	if err != nil {
		return false, err
	}

	patch, err := DiffPatch(in, observed)
	if err != nil {
		return false, err
	}

	if len(patch) > 0 {
		result = false

		_, actualName, actualNetworkACL, actualPublicGateway, _,
			actualRoutingTable, _, _, _, _ := getParameters(&actualParams)

		_, specName, specNetworkACL, specPublicGateway, _,
			specRoutingTable, _, _, _, _ := getParameters(in.DeepCopy())

		if _, ok := patch[nameKey]; ok {
			l.Info("IsUpToDate", nameKey, cmp.Diff(specName, actualName))
		}

		if _, ok := patch[networkACLKey]; ok {
			l.Info("IsUpToDate", networkACLKey, cmp.Diff(specNetworkACL, actualNetworkACL))
		}

		if _, ok := patch[publicGatewayKey]; ok {
			l.Info("IsUpToDate", publicGatewayKey, cmp.Diff(specPublicGateway, actualPublicGateway))
		}

		if _, ok := patch[routingTableKey]; ok {
			l.Info("IsUpToDate", routingTableKey, cmp.Diff(specRoutingTable, actualRoutingTable))
		}
	}

	return result, nil
}
