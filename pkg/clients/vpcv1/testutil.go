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

const (
	// Values below are the ones that will be used if we decide that some parameters has to be non-nil
	addressPrefixVal   = "an address"
	classicAccessVal   = false
	nameVal            = "a name"
	resourceGroupIDVal = "a resource group id"
)

var (
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
	createedAtNonNil bool,
	crnNonNil bool,
	hrefNonNil bool,
	idNonNil bool,
	nameNonNil bool,
	statusNonNil bool,
	cseSourceIpsNonNil bool,
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
	defaultRoutingTable_NameNil bool,
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

	result := ibmVPC.VPC{}

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
