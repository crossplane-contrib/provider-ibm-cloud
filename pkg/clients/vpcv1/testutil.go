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
	addressPrefixVal                     = ibmc.RandomString()
	classicAccessVal                     = ibmc.RandomInt(2) == 0
	nameVal                              = ibmc.RandomString()
	resourceGroupIDVal                   = ibmc.RandomString()
	crnVal                               = ibmc.RandomString()
	hrefVal                              = ibmc.RandomString()
	idVal                                = ibmc.RandomString()
	statusVal                            = ibmc.RandomString()
	createdAtVal                         = ibmc.ADateTimeInAYear(2012)
	cseSourceIpsLen                      = ibmc.RandomInt(3)
	cseSourceIps_IP_Address              = ibmc.RandomString()
	cseSource_Zone_Href                  = ibmc.RandomString()
	cseSource_Zone_Name                  = ibmc.RandomString()
	defaultNetworkACL_CRN                = ibmc.RandomString()
	defaultNetworkACL_Href               = ibmc.RandomString()
	defaultNetworkACL_ID                 = ibmc.RandomString()
	defaultNetworkACL_Name               = ibmc.RandomString()
	defaultNetworkACL_Deleted_MoreInfo   = ibmc.RandomString()
	defaultRoutingTable_Deleted_MoreInfo = ibmc.RandomString()
	defaultRoutingTable_Href             = ibmc.RandomString()
	defaultRoutingTable_ID               = ibmc.RandomString()
	defaultRoutingTable_Name             = ibmc.RandomString()
	defaultRoutingTable_ResourceType     = ibmc.RandomString()
	defaultSecurityGroup_CRN             = ibmc.RandomString()
	defaultSecurityGroup_Href            = ibmc.RandomString()
	defaultSecurityGroup_ID              = ibmc.RandomString()
	defaultSecurityGroup_Name            = ibmc.RandomString()
	resourceGroup_Name                   = ibmc.RandomString()
	resourceGroup_Href                   = ibmc.RandomString()

	headersMapVal = map[string]string{"a": "b", "c": "d"} // maps cannot be constants hence var. Do not modify.
)

// GetDummyCloudVPCParams returns a dummy object, ready to be used in create-VPC-in-the-cloud request. Non-nil values will be the
// ones of the local constants above.
//
// Params
//		each param is used in controlling the value of the similarly-named field in the returned structure
//
// Returns
//	    an object appropriately populated
func GetDummyCloudVPCObservation(
	classicAccessNonNil bool,
	createdAtNonNil bool,
	crnNonNil bool,
	hrefNonNil bool,
	idNonNil bool,
	nameNonNil bool,
	statusNonNil bool,
	cseSourceIps_IP_AdressNonNil bool,
	cseSourceIps_Zone_NonNil bool,
	cseSourceIps_Zone_Href_NonNil bool,
	cseSourceIps_Zone_Name_NonNil bool,
	defaultNetworkACL_NonNil bool,
	defaultNetworkACL_CRN_NonNil bool,
	defaultNetworkACL_Deleted_NonNil bool,
	defaultNetworkACL_Deleted_MoreInfoNonNil bool,
	defaultNetworkACL_Href_NonNil bool,
	defaultNetworkACL_ID_NonNil bool,
	defaultNetworkACL_Name_NonNil bool,
	defaultRoutingTable_NonNil bool,
	defaultRoutingTable_Deleted_NonNil bool,
	defaultRoutingTable_Deleted_MoreInfoNonNil,
	defaultRoutingTable_HrefNonNil bool,
	defaultRoutingTable_IDNonNil bool,
	defaultRoutingTable_NameNonNil bool,
	defaultRoutingTable_ResourceTypeNonNil bool,
	defaultSecurityGroup_NonNil bool,
	defaultSecurityGroup_CRN_NonNil bool,
	defaultSecurityGroup_Deleted_NonNil bool,
	defaultSecurityGroup_Deleted_MoreInfoNonNil bool,
	defaultSecurityGroup_Href_NonNil bool,
	defaultSecurityGroup_ID_NonNil bool,
	defaultSecurityGroup_Name_NonNil bool,
	resourceGroupNonNil bool,
	resourceGroup_Name_NonNil bool,
	resourceGroup_Href_NonNil bool,
	resourceGroup_ID_NonNil bool) ibmVPC.VPC {

	result := ibmVPC.VPC{
		ClassicAccess: ibmc.ReturnConditionalBool(classicAccessNonNil, classicAccessVal),
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
					Address: ibmc.ReturnConditionalStr(cseSourceIps_IP_AdressNonNil, cseSourceIps_IP_Address),
				},
			}

			if cseSourceIps_Zone_NonNil {
				result.CseSourceIps[i].Zone = &ibmVPC.ZoneReference{
					Href: ibmc.ReturnConditionalStr(cseSourceIps_Zone_Href_NonNil, cseSource_Zone_Href),
					Name: ibmc.ReturnConditionalStr(cseSourceIps_Zone_Name_NonNil, cseSource_Zone_Name),
				}
			}
		}
	}

	if defaultNetworkACL_NonNil {
		result.DefaultNetworkACL = &ibmVPC.NetworkACLReference{
			CRN:  ibmc.ReturnConditionalStr(defaultNetworkACL_CRN_NonNil, defaultNetworkACL_CRN),
			Href: ibmc.ReturnConditionalStr(defaultNetworkACL_Href_NonNil, defaultNetworkACL_Href),
			ID:   ibmc.ReturnConditionalStr(defaultNetworkACL_ID_NonNil, defaultNetworkACL_ID),
			Name: ibmc.ReturnConditionalStr(defaultNetworkACL_Name_NonNil, defaultNetworkACL_Name),
		}

		if defaultNetworkACL_Deleted_NonNil {
			result.DefaultNetworkACL.Deleted = &ibmVPC.NetworkACLReferenceDeleted{
				MoreInfo: ibmc.ReturnConditionalStr(defaultNetworkACL_Deleted_MoreInfoNonNil, defaultNetworkACL_Deleted_MoreInfo),
			}
		}
	}

	if defaultRoutingTable_NonNil {
		result.DefaultRoutingTable = &ibmVPC.RoutingTableReference{
			Href:         ibmc.ReturnConditionalStr(defaultRoutingTable_HrefNonNil, defaultRoutingTable_Href),
			ID:           ibmc.ReturnConditionalStr(defaultRoutingTable_IDNonNil, defaultRoutingTable_ID),
			Name:         ibmc.ReturnConditionalStr(defaultRoutingTable_NameNonNil, defaultRoutingTable_Name),
			ResourceType: ibmc.ReturnConditionalStr(defaultRoutingTable_ResourceTypeNonNil, defaultRoutingTable_ResourceType),
		}

		if defaultRoutingTable_Deleted_NonNil {
			result.DefaultRoutingTable.Deleted = &ibmVPC.RoutingTableReferenceDeleted{
				MoreInfo: ibmc.ReturnConditionalStr(defaultRoutingTable_Deleted_MoreInfoNonNil, defaultRoutingTable_Deleted_MoreInfo),
			}
		}
	}

	if defaultSecurityGroup_NonNil {
		result.DefaultSecurityGroup = &ibmVPC.SecurityGroupReference{
			CRN:  ibmc.ReturnConditionalStr(defaultSecurityGroup_CRN_NonNil, defaultSecurityGroup_CRN),
			Href: ibmc.ReturnConditionalStr(defaultSecurityGroup_Href_NonNil, defaultSecurityGroup_Href),
			ID:   ibmc.ReturnConditionalStr(defaultSecurityGroup_ID_NonNil, defaultSecurityGroup_ID),
			Name: ibmc.ReturnConditionalStr(defaultSecurityGroup_Name_NonNil, defaultSecurityGroup_Name),
		}

		if defaultSecurityGroup_Deleted_NonNil {
			result.DefaultSecurityGroup.Deleted = &ibmVPC.SecurityGroupReferenceDeleted{
				MoreInfo: ibmc.ReturnConditionalStr(defaultNetworkACL_Deleted_MoreInfoNonNil, defaultNetworkACL_Deleted_MoreInfo),
			}
		}
	}

	if resourceGroupNonNil {
		result.ResourceGroup = &ibmVPC.ResourceGroupReference{
			Href: ibmc.ReturnConditionalStr(resourceGroup_Href_NonNil, resourceGroup_Href),
			ID:   ibmc.ReturnConditionalStr(resourceGroup_ID_NonNil, resourceGroupIDVal),
			Name: ibmc.ReturnConditionalStr(resourceGroup_Name_NonNil, resourceGroup_Name),
		}
	}

	return result
}

// GetDummyCrossplaneVPCParams returns a dummy object, ready to be used in a create-VPC-in-k8s request. Non-nil values will be the
// ones of the local constants above
//
// Params
//		addressNil - whether to set the 'AddressPrefixManagement' member to nil
//  	classicAccessNil - whether to set the 'ClassicAccess' member to nil
// 		nameNil - whether to set the 'Name' member to nil
//		resourceGroupIDNil - whether to set the 'resourceGroupIDNil' member to nil
//      noHeaders - whether to include headers
//
// Returns
//	    an object appropriately populated
func GetDummyCrossplaneVPCParams(addressNil bool, classicAccessNil bool, nameNil bool, resourceGroupIDNil bool, noHeaders bool) v1alpha1.VPCParameters {
	result := v1alpha1.VPCParameters{}

	if !addressNil {
		result.AddressPrefixManagement = reference.ToPtrValue(addressPrefixVal)
	}

	if !classicAccessNil {
		result.ClassicAccess = ibmc.BoolPtr(classicAccessVal)
	}

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
