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

package vpcv1

import (
	ibmVPC "github.com/IBM/vpc-go-sdk/vpcv1"

	"github.com/crossplane-contrib/provider-ibm-cloud/apis/vpcv1/v1alpha1"
	ibmc "github.com/crossplane-contrib/provider-ibm-cloud/pkg/clients"

	"github.com/crossplane/crossplane-runtime/pkg/reference"
)

var (
	// Values below are the ones that will be used if we decide that some parameters has to be non-nil
	addressPrefixVal                   = ibmc.RandomString()
	classicAccessVal                   = ibmc.RandomInt(2) == 0
	nameVal                            = ibmc.RandomString()
	resourceGroupIDVal                 = ibmc.RandomString()
	crnVal                             = ibmc.RandomString()
	hrefVal                            = ibmc.RandomString()
	idVal                              = ibmc.RandomString()
	statusVal                          = ibmc.RandomString()
	createdAtVal                       = ibmc.ADateTimeInAYear(2012)
	cseSourceIpsLen                    = ibmc.RandomInt(3)
	cseSourceIpsIPAddress              = ibmc.RandomString()
	cseSourceZoneHref                  = ibmc.RandomString()
	cseSourceZoneName                  = ibmc.RandomString()
	defaultNetworkACLCRN               = ibmc.RandomString()
	defaultNetworkACLHref              = ibmc.RandomString()
	defaultNetworkACLID                = ibmc.RandomString()
	defaultNetworkACLName              = ibmc.RandomString()
	defaultNetworkACLDeletedMoreInfo   = ibmc.RandomString()
	defaultRoutingTableDeletedMoreInfo = ibmc.RandomString()
	defaultRoutingTableHref            = ibmc.RandomString()
	defaultRoutingTableID              = ibmc.RandomString()
	defaultRoutingTableName            = ibmc.RandomString()
	defaultRoutingTableResourceType    = ibmc.RandomString()
	defaultSecurityGroupCRN            = ibmc.RandomString()
	defaultSecurityGroupHref           = ibmc.RandomString()
	defaultSecurityGroupID             = ibmc.RandomString()
	defaultSecurityGroupName           = ibmc.RandomString()
	resourceGroupName                  = ibmc.RandomString()
	resourceGroupHref                  = ibmc.RandomString()

	headersMapVal = map[string]string{"a": "b", "c": "d"} // maps cannot be constants hence var. Do not modify.
)

// GetDummyCloudVPCObservation returns a dummy object, ready to be used in create-VPC-in-the-cloud request. Non-nil values will be the
// ones of the local constants above.
//
// Params
//		each param is used in controlling the value of the similarly-named field in the returned structure
//
// Returns
//	    an object appropriately populated
func GetDummyCloudVPCObservation( // nolint:gocyclo
	classicAccessNonNil bool,
	createdAtNonNil bool,
	crnNonNil bool,
	hrefNonNil bool,
	idNonNil bool,
	nameNonNil bool,
	statusNonNil bool,
	cseSourceIpsIPAdressNonNil bool,
	cseSourceIpsZoneNonNil bool,
	cseSourceIpsZoneHrefNonNil bool,
	cseSourceIpsZoneNameNonNil bool,
	defaultNetworkACLNonNil bool,
	defaultNetworkACLCRNNonNil bool,
	defaultNetworkACLDeletedNonNil bool,
	defaultNetworkACLDeletedMoreInfoNonNil bool,
	defaultNetworkACLHrefNonNil bool,
	defaultNetworkACLIDNonNil bool,
	defaultNetworkACLNameNonNil bool,
	defaultRoutingTableNonNil bool,
	defaultRoutingTableDeletedNonNil bool,
	defaultRoutingTableDeletedMoreInfoNonNil,
	defaultRoutingTableHrefNonNil bool,
	defaultRoutingTableIDNonNil bool,
	defaultRoutingTableNameNonNil bool,
	defaultRoutingTableResourceTypeNonNil bool,
	defaultSecurityGroupNonNil bool,
	defaultSecurityGroupCRNNonNil bool,
	defaultSecurityGroupDeletedNonNil bool,
	defaultSecurityGroupDeletedMoreInfoNonNil bool,
	defaultSecurityGroupHrefNonNil bool,
	defaultSecurityGroupIDNonNil bool,
	defaultSecurityGroupNameNonNil bool,
	resourceGroupNonNil bool,
	resourceGroupNameNonNil bool,
	resourceGroupHrefNonNil bool,
	resourceGroupIDNonNil bool) ibmVPC.VPC {

	result := ibmVPC.VPC{
		ClassicAccess: &classicAccessVal,
		CreatedAt:     ibmc.ReturnConditionalDate(createdAtNonNil, createdAtVal),
		CRN:           ibmc.ReturnConditionalStr(crnNonNil, crnVal),
		Href:          ibmc.ReturnConditionalStr(hrefNonNil, hrefVal),
		ID:            ibmc.ReturnConditionalStr(idNonNil, idVal),
		Name:          ibmc.ReturnConditionalStr(nameNonNil, nameVal),
		Status:        ibmc.ReturnConditionalStr(statusNonNil, statusVal),
	}

	if cseSourceIpsLen > 0 {
		result.CseSourceIps = make([]ibmVPC.VpccseSourceIP, cseSourceIpsLen)

		for i := 0; i < cseSourceIpsLen; i++ {
			result.CseSourceIps[i] = ibmVPC.VpccseSourceIP{
				IP: &ibmVPC.IP{
					Address: ibmc.ReturnConditionalStr(cseSourceIpsIPAdressNonNil, cseSourceIpsIPAddress),
				},
			}

			if cseSourceIpsZoneNonNil {
				result.CseSourceIps[i].Zone = &ibmVPC.ZoneReference{
					Href: ibmc.ReturnConditionalStr(cseSourceIpsZoneHrefNonNil, cseSourceZoneHref),
					Name: ibmc.ReturnConditionalStr(cseSourceIpsZoneNameNonNil, cseSourceZoneName),
				}
			}
		}
	}

	if defaultNetworkACLNonNil {
		result.DefaultNetworkACL = &ibmVPC.NetworkACLReference{
			CRN:  ibmc.ReturnConditionalStr(defaultNetworkACLCRNNonNil, defaultNetworkACLCRN),
			Href: ibmc.ReturnConditionalStr(defaultNetworkACLHrefNonNil, defaultNetworkACLHref),
			ID:   ibmc.ReturnConditionalStr(defaultNetworkACLIDNonNil, defaultNetworkACLID),
			Name: ibmc.ReturnConditionalStr(defaultNetworkACLNameNonNil, defaultNetworkACLName),
		}

		if defaultNetworkACLDeletedNonNil {
			result.DefaultNetworkACL.Deleted = &ibmVPC.NetworkACLReferenceDeleted{
				MoreInfo: ibmc.ReturnConditionalStr(defaultNetworkACLDeletedMoreInfoNonNil, defaultNetworkACLDeletedMoreInfo),
			}
		}
	}

	if defaultRoutingTableNonNil {
		result.DefaultRoutingTable = &ibmVPC.RoutingTableReference{
			Href:         ibmc.ReturnConditionalStr(defaultRoutingTableHrefNonNil, defaultRoutingTableHref),
			ID:           ibmc.ReturnConditionalStr(defaultRoutingTableIDNonNil, defaultRoutingTableID),
			Name:         ibmc.ReturnConditionalStr(defaultRoutingTableNameNonNil, defaultRoutingTableName),
			ResourceType: ibmc.ReturnConditionalStr(defaultRoutingTableResourceTypeNonNil, defaultRoutingTableResourceType),
		}

		if defaultRoutingTableDeletedNonNil {
			result.DefaultRoutingTable.Deleted = &ibmVPC.RoutingTableReferenceDeleted{
				MoreInfo: ibmc.ReturnConditionalStr(defaultRoutingTableDeletedMoreInfoNonNil, defaultRoutingTableDeletedMoreInfo),
			}
		}
	}

	if defaultSecurityGroupNonNil {
		result.DefaultSecurityGroup = &ibmVPC.SecurityGroupReference{
			CRN:  ibmc.ReturnConditionalStr(defaultSecurityGroupCRNNonNil, defaultSecurityGroupCRN),
			Href: ibmc.ReturnConditionalStr(defaultSecurityGroupHrefNonNil, defaultSecurityGroupHref),
			ID:   ibmc.ReturnConditionalStr(defaultSecurityGroupIDNonNil, defaultSecurityGroupID),
			Name: ibmc.ReturnConditionalStr(defaultSecurityGroupNameNonNil, defaultSecurityGroupName),
		}

		if defaultSecurityGroupDeletedNonNil {
			result.DefaultSecurityGroup.Deleted = &ibmVPC.SecurityGroupReferenceDeleted{
				MoreInfo: ibmc.ReturnConditionalStr(defaultNetworkACLDeletedMoreInfoNonNil, defaultNetworkACLDeletedMoreInfo),
			}
		}
	}

	if resourceGroupNonNil {
		result.ResourceGroup = &ibmVPC.ResourceGroupReference{
			Href: ibmc.ReturnConditionalStr(resourceGroupHrefNonNil, resourceGroupHref),
			ID:   ibmc.ReturnConditionalStr(resourceGroupIDNonNil, resourceGroupIDVal),
			Name: ibmc.ReturnConditionalStr(resourceGroupNameNonNil, resourceGroupName),
		}
	}

	return result
}

// GetDummyCrossplaneVPCParams returns a dummy object, ready to be used in a create-VPC-in-k8s request. Non-nil values will be the
// ones of the local constants above
//
// Params
//		addressNil - whether to set the 'AddressPrefixManagement' member to nil
// 		nameNil - whether to set the 'Name' member to nil
//		resourceGroupIDNil - whether to set the 'resourceGroupIDNil' member to nil
//      noHeaders - whether to include headers
//
// Returns
//	    an object appropriately populated
func GetDummyCrossplaneVPCParams(addressNil bool, nameNil bool, resourceGroupIDNil bool, noHeaders bool) v1alpha1.VPCParameters {
	result := v1alpha1.VPCParameters{}

	if !addressNil {
		result.AddressPrefixManagement = reference.ToPtrValue(addressPrefixVal)
	}

	result.ClassicAccess = classicAccessVal

	if !nameNil {
		result.Name = reference.ToPtrValue(nameVal)
	}

	if !resourceGroupIDNil {
		result.ResourceGroup = &v1alpha1.ResourceGroupIdentity{
			ID: resourceGroupIDVal,
		}
	}

	if !noHeaders {
		result.Headers = &headersMapVal
	}

	return result
}
