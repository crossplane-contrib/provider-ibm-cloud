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
	"strconv"
	"testing"

	ibmVPC "github.com/IBM/vpc-go-sdk/vpcv1"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/reference"

	"github.com/crossplane-contrib/provider-ibm-cloud/apis/vpcv1/v1alpha1"
	ibmc "github.com/crossplane-contrib/provider-ibm-cloud/pkg/clients"
	"github.com/crossplane-contrib/provider-ibm-cloud/pkg/clients/vpcv1"

	"github.com/google/go-cmp/cmp"
)

// Used in testing
type tstStruct struct {
	crossplaneVal interface{}
	cloudVal      interface{}
}

// Used in testing of IsUpToDate()
type tstStructUpToDate struct {
	spec     v1alpha1.SubnetParameters
	observed ibmVPC.Subnet
	want     bool
}

// Returns
//     a random boolean
func randomBool() bool {
	return ibmc.RandomInt(2) == 1
}

// Params
//    value - a value. Cannot be nil
//
// Returns
//    the value of the parameter (of the appopriate type, de-referenced if a pointer), or nil
func typeVal(value interface{}) interface{} {
	result := vpcv1.TypeVal(value)

	if result == nil {
		switch typed := value.(type) {
		case *ibmVPC.NetworkACLReference:
			if typed != nil {
				result = *typed
			}
		case *ibmVPC.PublicGatewayReference:
			if typed != nil {
				result = *typed
			}
		case *ibmVPC.ResourceGroupReference:
			if typed != nil {
				result = *typed
			}
		case *ibmVPC.RoutingTableReference:
			if typed != nil {
				result = *typed
			}
		case *ibmVPC.VPCReference:
			if typed != nil {
				result = *typed
			}
		case *ibmVPC.ZoneReference:
			if typed != nil {
				result = *typed
			}
		case *ibmVPC.NetworkACLIdentityByID:
			if typed != nil {
				result = *typed
			}
		case *ibmVPC.PublicGatewayIdentityPublicGatewayIdentityByID:
			if typed != nil {
				result = *typed
			}
		case *ibmVPC.ResourceGroupIdentityByID:
			if typed != nil {
				result = *typed
			}
		case *ibmVPC.RoutingTableIdentityByID:
			if typed != nil {
				result = *typed
			}
		case *ibmVPC.VPCIdentityByID:
			if typed != nil {
				result = *typed
			}
		case *ibmVPC.ZoneIdentityByName:
			if typed != nil {
				result = *typed
			}
		case v1alpha1.NetworkACLReference:
			result = typed
		case v1alpha1.PublicGatewayReference:
			result = typed
		case v1alpha1.ResourceGroupReference:
			result = typed
		case v1alpha1.RoutingTableReference:
			result = typed
		case v1alpha1.VPCReference:
			result = typed
		case v1alpha1.ZoneReference:
			result = typed
		case v1alpha1.SubnetParameters:
			result = typed
		}
	}

	return result
}

// Params
//    functionTested - name of the function getting tested
//    booleanComb - a permutation of 46 boolean variables
//
// Returns
//     a battery of tests. The crossplane values are structures or strings/numbers (they are never nil)
//     an error
func createTestsObservation(functionTested string, booleanComb []bool) (map[string]tstStruct, error) {
	ibmSubnetInfo := GetDummyObservation(
		booleanComb[0], booleanComb[1], booleanComb[2], booleanComb[3], booleanComb[4],
		booleanComb[5], booleanComb[6], booleanComb[7], booleanComb[8], booleanComb[9],
		booleanComb[10], booleanComb[11], booleanComb[12], booleanComb[13], booleanComb[14],
		booleanComb[15], booleanComb[16], booleanComb[17], booleanComb[18], booleanComb[19],
		booleanComb[20], booleanComb[21], booleanComb[22], booleanComb[23], booleanComb[24],
		booleanComb[25], booleanComb[26], booleanComb[27], booleanComb[28], booleanComb[29],
		booleanComb[30], booleanComb[31], booleanComb[32], booleanComb[33], booleanComb[34],
		booleanComb[35], booleanComb[36], booleanComb[37], booleanComb[38], booleanComb[39],
		booleanComb[40], booleanComb[41], booleanComb[42], booleanComb[43], booleanComb[44],
		booleanComb[45])
	crossplaneSubnetInfo, err := GenerateObservation(&ibmSubnetInfo)
	if err != nil {
		return nil, err
	}

	return map[string]tstStruct{
		"AvailableIpv4AddressCount": {
			cloudVal:      ibmSubnetInfo.AvailableIpv4AddressCount,
			crossplaneVal: crossplaneSubnetInfo.AvailableIpv4AddressCount,
		},
		"CreatedAt": {
			cloudVal:      ibmSubnetInfo.CreatedAt,
			crossplaneVal: crossplaneSubnetInfo.CreatedAt,
		},
		"CRN": {
			cloudVal:      ibmSubnetInfo.CRN,
			crossplaneVal: crossplaneSubnetInfo.CRN,
		},
		"Href": {
			cloudVal:      ibmSubnetInfo.Href,
			crossplaneVal: crossplaneSubnetInfo.Href,
		},
		"ID": {
			cloudVal:      ibmSubnetInfo.ID,
			crossplaneVal: crossplaneSubnetInfo.ID,
		},
		"IPVersion": {
			cloudVal:      ibmSubnetInfo.IPVersion,
			crossplaneVal: crossplaneSubnetInfo.IPVersion,
		},
		"Ipv4CIDRBlock": {
			cloudVal:      ibmSubnetInfo.Ipv4CIDRBlock,
			crossplaneVal: crossplaneSubnetInfo.Ipv4CIDRBlock,
		},
		"Name": {
			cloudVal:      ibmSubnetInfo.Name,
			crossplaneVal: crossplaneSubnetInfo.Name,
		},
		"NetworkACL": {
			cloudVal:      ibmSubnetInfo.NetworkACL,
			crossplaneVal: crossplaneSubnetInfo.NetworkACL,
		},
		"PublicGateway": {
			cloudVal:      ibmSubnetInfo.PublicGateway,
			crossplaneVal: crossplaneSubnetInfo.PublicGateway,
		},
		"ResourceGroup": {
			cloudVal:      ibmSubnetInfo.ResourceGroup,
			crossplaneVal: crossplaneSubnetInfo.ResourceGroup,
		},
		"RoutingTable": {
			cloudVal:      ibmSubnetInfo.RoutingTable,
			crossplaneVal: crossplaneSubnetInfo.RoutingTable,
		},
		"Status": {
			cloudVal:      ibmSubnetInfo.Status,
			crossplaneVal: crossplaneSubnetInfo.Status,
		},
		"TotalIpv4AddressCount": {
			cloudVal:      ibmSubnetInfo.TotalIpv4AddressCount,
			crossplaneVal: crossplaneSubnetInfo.TotalIpv4AddressCount,
		},
		"VPC": {
			cloudVal:      ibmSubnetInfo.VPC,
			crossplaneVal: crossplaneSubnetInfo.VPC,
		},
		"Zone": {
			cloudVal:      ibmSubnetInfo.Zone,
			crossplaneVal: crossplaneSubnetInfo.Zone,
		},
	}, nil
}

// Params
//    functionTested - name of the function getting tested
//    booleanComb - a permutation of 10 boolean variables
//
// Returns
//     a battery of tests
func createTestsCreateParams(functionTested string, booleanComb []bool) (map[string]tstStruct, error) {
	crossplaneSubnetInfo := GetDummyCreateParams(booleanComb[0], booleanComb[1], booleanComb[2], booleanComb[3],
		booleanComb[4], booleanComb[5], booleanComb[6], booleanComb[7], booleanComb[8], booleanComb[9])
	createSubnetOptions, err := GenerateCreateOptions(&crossplaneSubnetInfo)
	if err != nil {
		return nil, err
	}

	result := make(map[string]tstStruct)

	if crossplaneSubnetInfo.ByTocalCount != nil {
		ibmSubnetInfo := createSubnetOptions.SubnetPrototype.(*ibmVPC.SubnetPrototypeSubnetByTotalCount)
		result["IPVersion"] = tstStruct{
			cloudVal:      ibmSubnetInfo.IPVersion,
			crossplaneVal: crossplaneSubnetInfo.ByTocalCount.IPVersion,
		}

		result["Name"] = tstStruct{
			cloudVal:      ibmSubnetInfo.Name,
			crossplaneVal: crossplaneSubnetInfo.ByTocalCount.Name,
		}

		result["NetworkACL"] = tstStruct{
			cloudVal:      typeVal(ibmSubnetInfo.NetworkACL),
			crossplaneVal: crossplaneSubnetInfo.ByTocalCount.NetworkACL,
		}

		result["PublicGateway"] = tstStruct{
			cloudVal:      typeVal(ibmSubnetInfo.PublicGateway),
			crossplaneVal: crossplaneSubnetInfo.ByTocalCount.PublicGateway,
		}

		result["ResourceGroup"] = tstStruct{
			cloudVal:      typeVal(ibmSubnetInfo.ResourceGroup),
			crossplaneVal: crossplaneSubnetInfo.ByTocalCount.ResourceGroup,
		}

		result["RoutingTable"] = tstStruct{
			cloudVal:      typeVal(ibmSubnetInfo.RoutingTable),
			crossplaneVal: crossplaneSubnetInfo.ByTocalCount.RoutingTable,
		}

		result["TotalIpv4AddressCount"] = tstStruct{
			cloudVal:      *ibmSubnetInfo.TotalIpv4AddressCount,
			crossplaneVal: crossplaneSubnetInfo.ByTocalCount.TotalIpv4AddressCount,
		}

		result["VPC"] = tstStruct{
			cloudVal:      typeVal(ibmSubnetInfo.VPC),
			crossplaneVal: crossplaneSubnetInfo.ByTocalCount.VPC,
		}

		result["Zone"] = tstStruct{
			cloudVal:      typeVal(ibmSubnetInfo.Zone),
			crossplaneVal: crossplaneSubnetInfo.ByTocalCount.Zone,
		}
	} else {
		ibmSubnetInfo := createSubnetOptions.SubnetPrototype.(*ibmVPC.SubnetPrototypeSubnetByCIDR)
		result["IPVersion"] = tstStruct{
			cloudVal:      ibmSubnetInfo.IPVersion,
			crossplaneVal: crossplaneSubnetInfo.ByCIDR.IPVersion,
		}

		result["Name"] = tstStruct{
			cloudVal:      ibmSubnetInfo.Name,
			crossplaneVal: crossplaneSubnetInfo.ByCIDR.Name,
		}

		result["NetworkACL"] = tstStruct{
			cloudVal:      typeVal(ibmSubnetInfo.NetworkACL),
			crossplaneVal: crossplaneSubnetInfo.ByCIDR.NetworkACL,
		}

		result["PublicGateway"] = tstStruct{
			cloudVal:      typeVal(ibmSubnetInfo.PublicGateway),
			crossplaneVal: crossplaneSubnetInfo.ByCIDR.PublicGateway,
		}

		result["ResourceGroup"] = tstStruct{
			cloudVal:      typeVal(ibmSubnetInfo.ResourceGroup),
			crossplaneVal: crossplaneSubnetInfo.ByCIDR.ResourceGroup,
		}

		result["RoutingTable"] = tstStruct{
			cloudVal:      typeVal(ibmSubnetInfo.RoutingTable),
			crossplaneVal: crossplaneSubnetInfo.ByCIDR.RoutingTable,
		}

		result["VPC"] = tstStruct{
			cloudVal:      typeVal(ibmSubnetInfo.VPC),
			crossplaneVal: crossplaneSubnetInfo.ByCIDR.VPC,
		}

		result["Zone"] = tstStruct{
			cloudVal:      typeVal(ibmSubnetInfo.Zone),
			crossplaneVal: crossplaneSubnetInfo.ByCIDR.Zone,
		}

		result["Ipv4CIDRBlock"] = tstStruct{
			cloudVal:      *ibmSubnetInfo.Ipv4CIDRBlock,
			crossplaneVal: crossplaneSubnetInfo.ByCIDR.Ipv4CIDRBlock,
		}
	}

	return result, nil
}

// Params
//    functionTested - name of the function getting tested
//    booleanComb - a permutation of 46 boolean variables
//
// Returns
//     a battery of tests
func createTestsIsUpToDate(functionTested string, booleanComb []bool) (map[string]tstStructUpToDate, error) {
	isByTotalCount, nameNil, networkACLNil, publicGatewayNil, routingTableNil := booleanComb[0], booleanComb[1], booleanComb[2], booleanComb[3],
		booleanComb[4]

	crossplaneSubnetInfo := GetDummyCreateParams(isByTotalCount, randomBool(), nameNil, networkACLNil, publicGatewayNil, randomBool(),
		routingTableNil, randomBool(), randomBool(), randomBool())
	ibmSubnetInfo := GetDummyObservation(
		isByTotalCount, randomBool(), randomBool(), randomBool(), randomBool(),
		randomBool(), randomBool(), nameNil, networkACLNil, randomBool(),
		randomBool(), randomBool(), randomBool(), networkACLNil, randomBool(),
		publicGatewayNil, randomBool(), randomBool(), randomBool(), randomBool(),
		publicGatewayNil, randomBool(), randomBool(), randomBool(), randomBool(),
		randomBool(), randomBool(), routingTableNil, randomBool(), randomBool(),
		randomBool(), routingTableNil, randomBool(), randomBool(), randomBool(),
		randomBool(), false /* VPC must be != nil */, randomBool(), randomBool(), randomBool(),
		false /* VPC Id must be != nil */, randomBool(), randomBool(),
		!isByTotalCount /* zone depends on this */, randomBool(), !isByTotalCount /* zone depends on this */)

	result := make(map[string]tstStructUpToDate)

	result["uptodate!"] = tstStructUpToDate{
		spec:     crossplaneSubnetInfo,
		observed: ibmSubnetInfo,
		want:     true,
	}

	result["Name"] = tstStructUpToDate{
		spec:     withNewName(crossplaneSubnetInfo),
		observed: ibmSubnetInfo,
		want:     false,
	}

	result["NetworkACL"] = tstStructUpToDate{
		spec:     withNewNetworkACL(crossplaneSubnetInfo),
		observed: ibmSubnetInfo,
		want:     false,
	}

	result["PublicGateway"] = tstStructUpToDate{
		spec:     withNewPublicGateway(crossplaneSubnetInfo),
		observed: ibmSubnetInfo,
		want:     false,
	}

	result["RoutingTable"] = tstStructUpToDate{
		spec:     withNewRoutingTable(crossplaneSubnetInfo),
		observed: ibmSubnetInfo,
		want:     false,
	}

	return result, nil
}

// Compares 2 values
//
// Params
//    functionTested - the name of the tested function
//    variableCombination - binary string, indicating the nullness of the variables used for initializing the objects.
//                          Used mainly for logging/tracking
//    tstName - the name of the actual test
//    t - the testing object
//    cloudVal - a value from the cloud (or dummy thereof)
//    crossplaneVal - a crossplane value (or dummy therof)
//
// Returns
//     whether they are the same or not (also prints the diff)
func compareTst(functionTested string, variableCombination string, tstName string,
	t *testing.T, cloudVal interface{}, crossplaneVal interface{}) bool {
	fullTstName := functionTested + " " + variableCombination + " " + tstName

	cld := typeVal(cloudVal)
	cpv := typeVal(crossplaneVal)

	if observationEq(cld, cpv) {
		return false
	}

	if diff := cmp.Diff(cpv, cld); diff != "" {
		t.Errorf(fullTstName+": -wanted, +got:\n%s", diff)

		return false
	}

	return true
}

// Params
//    cloudVal - the cloud value ("observation" types in the sdk). If not nil, it contains a real object (not a pointer)
//    crossplaneVal - the crossplane value ("observation" types in v1alpha1)
//
// Returns
//    whether they point to equal strings OR cloudVal == nil && crossplaneVal == ""
func observationEq(cloudVal interface{}, crossplaneVal interface{}) bool {
	var result *bool

	switch crp := crossplaneVal.(type) {
	case v1alpha1.NetworkACLReference:
		if cloudVal != nil {
			cld := cloudVal.(ibmVPC.NetworkACLReference)

			result = ibmc.BoolPtr(crp.CRN == reference.FromPtrValue(cld.CRN) &&
				crp.Href == reference.FromPtrValue(cld.Href) &&
				crp.ID == reference.FromPtrValue(cld.ID) &&
				crp.Name == reference.FromPtrValue(cld.Name))

			if cld.Deleted != nil {
				result = ibmc.BoolPtr(*result && crp.Deleted.MoreInfo == reference.FromPtrValue(cld.Deleted.MoreInfo))
			}
		} else {
			result = ibmc.BoolPtr(crp.CRN == "" && crp.Href == "" && crp.ID == "" && crp.Name == "" && crp.Deleted == nil)
		}
	case v1alpha1.PublicGatewayReference:
		if cloudVal != nil {
			cld := cloudVal.(ibmVPC.PublicGatewayReference)

			result = ibmc.BoolPtr(crp.CRN == reference.FromPtrValue(cld.CRN) &&
				crp.Href == reference.FromPtrValue(cld.Href) &&
				crp.ID == reference.FromPtrValue(cld.ID) &&
				crp.Name == reference.FromPtrValue(cld.Name) &&
				crp.ResourceType == reference.FromPtrValue(cld.ResourceType))

			if cld.Deleted != nil {
				result = ibmc.BoolPtr(*result && crp.Deleted.MoreInfo == reference.FromPtrValue(cld.Deleted.MoreInfo))
			}
		} else {
			result = ibmc.BoolPtr(crp.CRN == "" && crp.Href == "" && crp.Name == "" && crp.ResourceType == "" && crp.Deleted == nil)
		}
	case v1alpha1.ResourceGroupReference:
		if cloudVal != nil {
			cld := cloudVal.(ibmVPC.ResourceGroupReference)

			result = ibmc.BoolPtr(crp.Href == reference.FromPtrValue(cld.Href) &&
				crp.ID == reference.FromPtrValue(cld.ID) &&
				crp.Name == reference.FromPtrValue(cld.Name))
		} else {
			result = ibmc.BoolPtr(crp.Href == "" && crp.ID == "" && crp.Name == "")
		}
	case v1alpha1.RoutingTableReference:
		if cloudVal != nil {
			cld := cloudVal.(ibmVPC.RoutingTableReference)

			result = ibmc.BoolPtr(crp.Href == reference.FromPtrValue(cld.Href) &&
				crp.ID == reference.FromPtrValue(cld.ID) &&
				crp.Name == reference.FromPtrValue(cld.Name) &&
				crp.ResourceType == reference.FromPtrValue(cld.ResourceType))

			if cld.Deleted != nil {
				result = ibmc.BoolPtr(*result && crp.Deleted.MoreInfo == reference.FromPtrValue(cld.Deleted.MoreInfo))
			}
		} else {
			result = ibmc.BoolPtr(crp.ID == "" && crp.Href == "" && crp.Name == "" && crp.ResourceType == "" && crp.Deleted == nil)
		}
	case v1alpha1.VPCReference:
		if cloudVal != nil {
			cld := cloudVal.(ibmVPC.VPCReference)

			result = ibmc.BoolPtr(crp.CRN == reference.FromPtrValue(cld.CRN) &&
				crp.Href == reference.FromPtrValue(cld.Href) &&
				crp.ID == reference.FromPtrValue(cld.ID) &&
				crp.Name == reference.FromPtrValue(cld.Name))

			if cld.Deleted != nil {
				result = ibmc.BoolPtr(*result && crp.Deleted.MoreInfo == reference.FromPtrValue(cld.Deleted.MoreInfo))
			}
		} else {
			result = ibmc.BoolPtr(crp.ID == "" && crp.Href == "" && crp.Name == "" && crp.CRN == "" && crp.Deleted == nil)
		}
	case v1alpha1.ZoneReference:
		if cloudVal != nil {
			cld := cloudVal.(ibmVPC.ZoneReference)

			result = ibmc.BoolPtr(crp.Href == reference.FromPtrValue(cld.Href) &&
				crp.Name == reference.FromPtrValue(cld.Name))
		} else {
			result = ibmc.BoolPtr(crp.Href == "" && crp.Name == "")
		}
	case string:
		if cloudVal != nil {
			cld := cloudVal.(string)

			result = ibmc.BoolPtr(crp == cld)
		} else {
			result = ibmc.BoolPtr(crp == "")
		}
	}

	if result == nil {
		result = ibmc.BoolPtr(false)
	}

	return *result
}

// Params
//    params - ...
//
// Returns
//    a copy of the argument but with the name set
func withNewName(params v1alpha1.SubnetParameters) v1alpha1.SubnetParameters {
	result := params.DeepCopy()

	if result.ByTocalCount != nil {
		result.ByTocalCount.Name = reference.ToPtrValue(reference.FromPtrValue(result.ByTocalCount.Name) + ibmc.RandomString(true))
	} else {
		result.ByCIDR.Name = reference.ToPtrValue(reference.FromPtrValue(result.ByCIDR.Name) + ibmc.RandomString(true))
	}

	return *result
}

// Params
//    params - ...
//
// Returns
//    a copy of the argument but with the name set
func withNewNetworkACL(params v1alpha1.SubnetParameters) v1alpha1.SubnetParameters {
	result := params.DeepCopy()

	var oldID string
	if result.ByTocalCount != nil && result.ByTocalCount.NetworkACL != nil {
		oldID = result.ByTocalCount.NetworkACL.ID
	} else if result.ByCIDR != nil && result.ByCIDR.NetworkACL != nil {
		oldID = result.ByCIDR.NetworkACL.ID
	}

	if result.ByTocalCount != nil {
		result.ByTocalCount.NetworkACL = &v1alpha1.NetworkACLIdentity{
			ID: ibmc.RandomString(true) + oldID,
		}
	} else {
		result.ByCIDR.NetworkACL = &v1alpha1.NetworkACLIdentity{
			ID: ibmc.RandomString(true) + oldID,
		}
	}

	return *result
}

// Params
//    params - ...
//
// Returns
//    a copy of the argument but with the name set
func withNewPublicGateway(params v1alpha1.SubnetParameters) v1alpha1.SubnetParameters {
	result := params.DeepCopy()

	var oldID string
	if result.ByTocalCount != nil && result.ByTocalCount.PublicGateway != nil {
		oldID = result.ByTocalCount.PublicGateway.ID
	} else if result.ByCIDR != nil && result.ByCIDR.PublicGateway != nil {
		oldID = result.ByCIDR.PublicGateway.ID
	}

	if result.ByTocalCount != nil {
		result.ByTocalCount.PublicGateway = &v1alpha1.PublicGatewayIdentity{
			ID: ibmc.RandomString(true) + oldID,
		}
	} else {
		result.ByCIDR.PublicGateway = &v1alpha1.PublicGatewayIdentity{
			ID: ibmc.RandomString(true) + oldID,
		}
	}

	return *result
}

// Params
//    params - ...
//
// Returns
//    a copy of the argument but with the name set
func withNewRoutingTable(params v1alpha1.SubnetParameters) v1alpha1.SubnetParameters {
	result := params.DeepCopy()

	var oldID string
	if result.ByTocalCount != nil && result.ByTocalCount.RoutingTable != nil {
		oldID = result.ByTocalCount.RoutingTable.ID
	} else if result.ByCIDR != nil && result.ByCIDR.RoutingTable != nil {
		oldID = result.ByCIDR.RoutingTable.ID
	}

	if result.ByTocalCount != nil {
		result.ByTocalCount.RoutingTable = &v1alpha1.RoutingTableIdentity{
			ID: ibmc.RandomString(true) + oldID,
		}
	} else {
		result.ByCIDR.RoutingTable = &v1alpha1.RoutingTableIdentity{
			ID: ibmc.RandomString(true) + oldID,
		}
	}

	return *result
}

// Tests the GenerateObservation function
func TestGenerateObservation(t *testing.T) {
	functionTstName := "GenerateObservation"

	numVars := 10 // as many as the params of booleanComb we will be using
	for i, booleanComb := range vpcv1.GenerateSomePermutations(numVars, 46, true) {
		varCombinationLogging := vpcv1.GetBinaryRep(i, numVars)
		if tests, err := createTestsObservation(functionTstName, booleanComb); err == nil {
			for name, tc := range tests {
				t.Run(functionTstName, func(t *testing.T) {
					_ = compareTst(functionTstName, varCombinationLogging, name, t, tc.cloudVal, tc.crossplaneVal)
				})
			}
		} else {
			t.Errorf(functionTstName + " " + varCombinationLogging + ", createTestsObservation() returned error: " + err.Error())
		}
	}
}

// Tests the GenerateCreateVPCOptions function
func TestGenerateCreateOptions(t *testing.T) {
	functionTstName := "GenerateCreateOptions"

	numVars := 10 // does not make sense to have more than the num of vars used...
	for i, booleanComb := range vpcv1.GeneratePermutations(numVars) {
		varCombinationLogging := vpcv1.GetBinaryRep(i, numVars)

		if tests, err := createTestsCreateParams(functionTstName, booleanComb); err == nil {
			for name, tc := range tests {
				t.Run(functionTstName, func(t *testing.T) {
					_ = compareTst(functionTstName, varCombinationLogging, name, t, tc.cloudVal, tc.crossplaneVal)
				})
			}
		} else {
			t.Errorf(functionTstName + " " + varCombinationLogging + ", createTestsCreateParams() returned error: " + err.Error())
		}
	}
}

// Tests the IsUpToDate function
func TestIsUpToDate(t *testing.T) {
	functionTstName := "IsUpToDate"

	numVars := 5 // as many as we will be using
	for i, booleanComb := range vpcv1.GeneratePermutations(numVars) {
		varCombinationLogging := vpcv1.GetBinaryRep(i, numVars)

		if tests, err := createTestsIsUpToDate(functionTstName, booleanComb); err == nil {
			for name, tc := range tests {
				t.Run(functionTstName, func(t *testing.T) {
					rc, _ := IsUpToDate(&tc.spec, &tc.observed, logging.NewNopLogger())
					if rc != tc.want {
						t.Errorf(functionTstName+" "+varCombinationLogging+" "+name+" IsUpToDate(...): -want:%t, +got:%t\n", tc.want, rc)
					}
				})
			}
		} else {
			t.Errorf(functionTstName + " " + varCombinationLogging + ", createTestsIsUpToDate() returned error: " + err.Error())
		}
	}
}

// Tests the LateInitializeSpec function
func TestLateInitializeSpec(t *testing.T) {
	functionTstName := "LateInitializeSpec"

	for _, byTotalCount := range []bool{true, false} {
		fullySpeced := GetDummyCreateParams(byTotalCount, false, false, false, false, false, false, false, false, false)

		cloudSubnetInfo := GetDummyObservation(
			byTotalCount, false, false, false, false, false, false, false, false, false, false,
			false, false, false, false, false, false, false, false, false, false,
			false, false, false, false, false, false, false, false, false, false,
			false, false, false, false, false, false, false, false, false, false,
			false, false, false, false, false)

		numVars := 9 // does not make sense to have more than the num of vars used... If we put too many,
		// then testing timeouts (30 secs)
		for i, booleanComb := range vpcv1.GenerateSomePermutations(numVars, 45, true) {
			varCombinationLogging := strconv.FormatBool(byTotalCount) + "-" + vpcv1.GetBinaryRep(i, numVars)

			spec := GetDummyCreateParams(byTotalCount, booleanComb[0], booleanComb[1], booleanComb[2],
				booleanComb[3], booleanComb[4], booleanComb[5],
				booleanComb[6], booleanComb[7], booleanComb[8])

			toLateInitialize := spec.DeepCopy()

			if _, err := LateInitializeSpec(toLateInitialize, &cloudSubnetInfo); err != nil {
				t.Errorf(functionTstName+" "+varCombinationLogging+": got error in LateInitializeSpec:\n%s", err)

				return
			}

			if diff := cmp.Diff(&fullySpeced, toLateInitialize); diff != "" {
				t.Errorf(functionTstName+" "+varCombinationLogging+" : -wanted, +got:\n%s", diff)
			}
		}
	}
}
