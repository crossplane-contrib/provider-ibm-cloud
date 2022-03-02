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
	"reflect"
	"strconv"
	"testing"

	ibmVPC "github.com/IBM/vpc-go-sdk/vpcv1"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/reference"

	"github.com/crossplane-contrib/provider-ibm-cloud/apis/vpcv1/v1alpha1"

	"github.com/google/go-cmp/cmp"
)

// Params
//    value - a value. May be nil
//
// Returns
//    the value of the parameter (of the appopriate type), or nil
func typeVal(value interface{}) interface{} {
	var result interface{}

	switch typed := value.(type) {
	case string:
		result = &typed
	case *string:
		result = typed
	case bool:
		result = &typed
	case *bool:
		result = typed
	case *map[string]string:
		result = typed
	case map[string]string:
		if typed == nil {
			result = nil
		} else {
			result = &typed
		}
	case *v1alpha1.ResourceGroupIdentity:
		result = typed
	case *ibmVPC.ResourceGroupIdentity:
		result = typed
	case *ibmVPC.ResourceGroupIdentityByID:
		result = typed
	}

	return result
}

// Params
//    crossplaneRGIntf - a crossplane resource group (must be of type *v1alpha1.ResourceGroupIdentityAlsoByID; interface for convenience). May be nil
//    cloudRGIntf - a cloud resource group (must be of type *ibmVPC.ResourceGroupIdentityIntf; interface for convenience). May be nil
//
//    Note that both params should NOT be at the same time nil (or point to nil structures)
//
// Returns
//    whether they point to the same underlying resource
func sameResource(crossplaneRGIntf interface{}, cloudRGIntf interface{}) bool {
	result := false

	if crossplaneRGIntf != nil && !reflect.ValueOf(crossplaneRGIntf).IsNil() && cloudRGIntf != nil && !reflect.ValueOf(cloudRGIntf).IsNil() {
		crossplaneRG := crossplaneRGIntf.(*v1alpha1.ResourceGroupIdentity)

		cloudRG, ok := cloudRGIntf.(*ibmVPC.ResourceGroupIdentity)
		if ok && cloudRG.ID != nil {
			result = (crossplaneRG.ID == *cloudRG.ID)
		}

		cloudRGByID, ok := cloudRGIntf.(*ibmVPC.ResourceGroupIdentityByID)
		if ok && cloudRGByID.ID != nil {
			result = (crossplaneRG.ID == *cloudRGByID.ID)
		}
	}

	return result
}

// Params
//     ibmVPCInfo - for/from the cloud (dummy)
//     crossplaneVPCInfo - or/from the cloud (dummy)
//
// Returns
//     a battery of tests
func createTestsObservation(ibmVPCInfo *ibmVPC.VPC, crossplaneVPCInfo *v1alpha1.VPCObservation) map[string]struct {
	cloudVal      interface{}
	crossplaneVal interface{}
} {
	return map[string]struct {
		cloudVal      interface{}
		crossplaneVal interface{}
	}{
		"ClassicAccess": {
			cloudVal:      ibmVPCInfo.ClassicAccess,
			crossplaneVal: crossplaneVPCInfo.ClassicAccess,
		},
		"CreatedAt": {
			cloudVal:      ibmVPCInfo.CreatedAt,
			crossplaneVal: crossplaneVPCInfo.CreatedAt,
		},
		"CRN": {
			cloudVal:      ibmVPCInfo.CRN,
			crossplaneVal: crossplaneVPCInfo.CRN,
		},
		"Href": {
			cloudVal:      ibmVPCInfo.Href,
			crossplaneVal: crossplaneVPCInfo.Href,
		},
		"ID": {
			cloudVal:      ibmVPCInfo.ID,
			crossplaneVal: crossplaneVPCInfo.ID,
		},
		"Name": {
			cloudVal:      ibmVPCInfo.Name,
			crossplaneVal: crossplaneVPCInfo.Name,
		},
		"Status": {
			cloudVal:      ibmVPCInfo.Status,
			crossplaneVal: crossplaneVPCInfo.Status,
		},
		"CseSourceIps": {
			cloudVal:      ibmVPCInfo.CseSourceIps,
			crossplaneVal: crossplaneVPCInfo.CseSourceIps,
		},
		"DefaultNetworkACL": {
			cloudVal:      ibmVPCInfo.DefaultNetworkACL,
			crossplaneVal: crossplaneVPCInfo.DefaultNetworkACL,
		},
		"DefaultRoutingTable": {
			cloudVal:      ibmVPCInfo.DefaultRoutingTable,
			crossplaneVal: crossplaneVPCInfo.DefaultRoutingTable,
		},
		"DefaultSecurityGroup": {
			cloudVal:      ibmVPCInfo.DefaultSecurityGroup,
			crossplaneVal: crossplaneVPCInfo.DefaultSecurityGroup,
		},
		"ResourceGroup": {
			cloudVal:      ibmVPCInfo.ResourceGroup,
			crossplaneVal: crossplaneVPCInfo.ResourceGroup,
		},
	}
}

// Params
//     ibmVPCInfo - for/from the cloud (dummy)
//     crossplaneVPCInfo - or/from the cloud (dummy)
//
// Returns
//     a battery of tests
func createTestsCreateParams(ibmVPCInfo *ibmVPC.CreateVPCOptions, crossplaneVPCInfo *v1alpha1.VPCParameters) map[string]struct {
	cloudVal      interface{}
	crossplaneVal interface{}
} {
	return map[string]struct {
		cloudVal      interface{}
		crossplaneVal interface{}
	}{
		"AddressPrefixManagement": {
			cloudVal:      ibmVPCInfo.AddressPrefixManagement,
			crossplaneVal: crossplaneVPCInfo.AddressPrefixManagement,
		},
		"ClassicAccess": {
			cloudVal:      ibmVPCInfo.ClassicAccess,
			crossplaneVal: crossplaneVPCInfo.ClassicAccess,
		},
		"Name": {
			cloudVal:      ibmVPCInfo.Name,
			crossplaneVal: crossplaneVPCInfo.Name,
		},
		"ResourceGroup": {
			cloudVal:      ibmVPCInfo.ResourceGroup,
			crossplaneVal: crossplaneVPCInfo.ResourceGroup,
		},
		"Headers": {
			cloudVal:      ibmVPCInfo.Headers,
			crossplaneVal: crossplaneVPCInfo.Headers,
		},
	}
}

// Compares 2 interface values for nilness (there own or the variable they point to..)
//
// Params
//    a - an inteface. May be nil
//    b - an interaface. May be nil
//
// Returns
//    whether they are both pointing to nil values or are nil
func areEquallyNil(a interface{}, b interface{}) bool {
	result := false

	if (a == nil || reflect.ValueOf(a).IsNil()) &&
		(b == nil || reflect.ValueOf(b).IsNil()) {
		result = true
	}

	return result
}

// Tests the GenerateCrossplaneVPCObservation function
func TestGenerateCrossplaneVPCObservation(t *testing.T) {
	functionTstName := "TestGenerateCrossplaneVPCObservation"

	numVars := 10 // as many as the params of booleanComb we will be using
<<<<<<< HEAD
	for i, booleanComb := range GenerateSomeCombinations(numVars, 34, true) {
=======
	for i, booleanComb := range GenerateSomeCombinations(numVars, 32, true) {
>>>>>>> 90b51371a5c47187a61bd83aa140e7bea15ff394
		varCombinationLogging := GetBinaryRep(i, numVars)

		ibmVPCInfo := GetDummyCloudVPCObservation(
			booleanComb[0], booleanComb[1], booleanComb[2], booleanComb[3], booleanComb[4],
			booleanComb[5], booleanComb[6], booleanComb[7], booleanComb[8], booleanComb[9],
			booleanComb[10], booleanComb[11], booleanComb[12], booleanComb[13], booleanComb[14],
			booleanComb[15], booleanComb[16], booleanComb[17], booleanComb[18], booleanComb[19],
			booleanComb[20], booleanComb[21], booleanComb[22], booleanComb[23], booleanComb[24],
			booleanComb[25], booleanComb[26], booleanComb[27], booleanComb[28], booleanComb[29],
<<<<<<< HEAD
			booleanComb[30], booleanComb[31], booleanComb[32], booleanComb[33])
=======
			booleanComb[30], booleanComb[31])
>>>>>>> 90b51371a5c47187a61bd83aa140e7bea15ff394
		crossplaneVPCInfo, err := GenerateCrossplaneVPCObservation(&ibmVPCInfo)
		if err != nil {
			t.Errorf(functionTstName + " " + varCombinationLogging + ": function GenerateCrossplaneVPCParams() returned error: " + err.Error())

			return
		}

		tests := createTestsObservation(&ibmVPCInfo, &crossplaneVPCInfo)
		for name, tc := range tests {
			t.Run(functionTstName, func(t *testing.T) {
				fullTstName := functionTstName + " " + varCombinationLogging + " " + name

				cloudVal := typeVal(tc.cloudVal)
				crossplaneVal := typeVal(tc.crossplaneVal)

				if areEquallyNil(cloudVal, crossplaneVal) {
					return
				}

				if diff := cmp.Diff(cloudVal, crossplaneVal); diff != "" {
					t.Errorf(fullTstName+": -wanted, +got:\n%s", diff)
				}
			})
		}
	}
}

// Tests the GenerateCloudVPCParams function
func TestGenerateCloudVPCParams(t *testing.T) {
	functionTstName := "TestGenerateCloudVPCParams"

	numVars := 4 // does not make sense to have more than the num of vars used...
	for i, booleanComb := range generateCombinations(numVars) {
		varCombinationLogging := GetBinaryRep(i, numVars)

		crossplaneVPCInfo := GetDummyCrossplaneVPCParams(booleanComb[0], booleanComb[1], booleanComb[2], booleanComb[3])
		ibmVPCInfo, err := GenerateCloudVPCParams(&crossplaneVPCInfo)
		if err != nil {
			t.Errorf(functionTstName + " " + varCombinationLogging + ": function GenerateCrossplaneVPCParams() returned error: " + err.Error())

			return
		}

		tests := createTestsCreateParams(&ibmVPCInfo, &crossplaneVPCInfo)
		for name, tc := range tests {
			t.Run(functionTstName, func(t *testing.T) {
				fullTstName := functionTstName + " " + varCombinationLogging + " " + name

				cloudVal := typeVal(tc.cloudVal)
				crossplaneVal := typeVal(tc.crossplaneVal)

				if areEquallyNil(crossplaneVal, cloudVal) {
					return
				}

				if crossplaneVal != nil {
					if reflect.TypeOf(crossplaneVal).String() == "*v1alpha1.ResourceGroupIdentity" {
						if !sameResource(crossplaneVal, cloudVal) {
							t.Errorf(fullTstName+": different IDs - cloudVal=%s, crossplaneVal=%s", cloudVal, crossplaneVal)
						}

						return
					} else if reflect.TypeOf(crossplaneVal).String() == "*map[string]string" {
						if diff := cmp.Diff(crossplaneVal, cloudVal); diff != "" {
							t.Errorf(fullTstName+": -wanted, +got:\n%s", diff)
						}

						return
					}
				}

				if diff := cmp.Diff(crossplaneVal, cloudVal); diff != "" {
					t.Errorf(fullTstName+": -wanted, +got:\n%s", diff)
				}
			})
		}
	}
}

// Params
//    params - ...
//    setName - whether to set the name
//
// Returns
//    a copy of the argument but with the name set
func withName(params v1alpha1.VPCParameters, setName bool) v1alpha1.VPCParameters {
	result := params.DeepCopy()

	if setName && result.Name != nil {
		result.Name = &randomName
	}

	return *result
}

// Params
//    params - ...
//    setRG - whether to set the resource group id
//
// Returns
//    a copy of the argument but with the name set
func withResourceGroup(params v1alpha1.VPCParameters, setRG bool) v1alpha1.VPCParameters {
	result := params.DeepCopy()

	if setRG && result.ResourceGroup != nil {
		result.ResourceGroup = &v1alpha1.ResourceGroupIdentity{
			ID: randomResourceGroupID,
		}
	}

	return *result
}

// Used to old the late initialization spec tests
type lateInitTstSpec struct {
	spec            v1alpha1.VPCParameters
	setName         bool
	setReourceGroup bool
	expects         v1alpha1.VPCParameters
}

// Creates tests for LateInitializeSpec
func createLateInitializeSpecTests(orig v1alpha1.VPCParameters) map[string]lateInitTstSpec {
	result := make(map[string]lateInitTstSpec)

	for _, nameIsNil := range []bool{true, false} {
		for _, rgIsNil := range []bool{true, false} {
			result[strconv.FormatBool(nameIsNil)+","+strconv.FormatBool(nameIsNil)] = lateInitTstSpec{
				spec:            orig,
				setName:         nameIsNil,
				setReourceGroup: rgIsNil,
				expects:         withResourceGroup(withName(orig, nameIsNil), rgIsNil),
			}
		}
	}

	return result
}

// Tests the IsUpToDate function
func TestIsUpToDate(t *testing.T) {
	functionTstName := "TestIsUpToDate"

	cases := map[string]struct {
		spec     v1alpha1.VPCParameters
		observed ibmVPC.VPC
		want     bool
	}{
		"TestIsUpToDate-1": {
			spec: v1alpha1.VPCParameters{
				Name: reference.ToPtrValue("foo"),
			},
			observed: ibmVPC.VPC{
				Name: reference.ToPtrValue("foo"),
			},
			want: true,
		},
		"TestIsUpToDate-2": {
			spec: v1alpha1.VPCParameters{
				Name: reference.ToPtrValue("foo"),
			},
			observed: ibmVPC.VPC{
				Name: reference.ToPtrValue("bar"),
			},
			want: false,
		},
		"TestIsUpToDate-3": {
			spec: v1alpha1.VPCParameters{
				Name: nil,
			},
			observed: ibmVPC.VPC{
				Name: reference.ToPtrValue("bar"),
			},
			want: false,
		},
		"TestIsUpToDate-4": {
			spec: v1alpha1.VPCParameters{
				Name: reference.ToPtrValue("bar"),
			},
			observed: ibmVPC.VPC{
				Name: nil,
			},
			want: false,
		},
		"TestIsUpToDate-5": {
			spec: v1alpha1.VPCParameters{
				Name: nil,
			},
			observed: ibmVPC.VPC{
				Name: nil,
			},
			want: true,
		},
		"TestIsUpToDate-6": {
			spec: v1alpha1.VPCParameters{
				Name: reference.ToPtrValue(""),
			},
			observed: ibmVPC.VPC{
				Name: nil,
			},
			want: true,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			rc, _ := IsUpToDate(&tc.spec, &tc.observed, logging.NewNopLogger())
			if rc != tc.want {
				t.Errorf(functionTstName+" "+name+"IsUpToDate(...): -want:%t, +got:%t\n", tc.want, rc)
			}
		})
	}
}

// Tests the LateInitializeSpec function
func TestLateInitializeSpec(t *testing.T) {
	functionTstName := "TestLateInitializeSpec"

	numVars := 10 // does not make sense to have more than the num of vars used... If we put too many,
	// then testing timeouts (30 secs)
<<<<<<< HEAD
	for i, booleanComb := range GenerateSomeCombinations(numVars, 34, true) {
=======
	for i, booleanComb := range GenerateSomeCombinations(numVars, 32, true) {
>>>>>>> 90b51371a5c47187a61bd83aa140e7bea15ff394
		varCombinationLogging := GetBinaryRep(i, numVars)

		crossplaneVPCInfo := GetDummyCrossplaneVPCParams(booleanComb[0], booleanComb[1], booleanComb[2], booleanComb[3])

		tests := createLateInitializeSpecTests(crossplaneVPCInfo)
		for name, tc := range tests {
			t.Run(functionTstName, func(t *testing.T) {
				fullTstName := functionTstName + " " + varCombinationLogging + " " + name

				cloudVPC := GetDummyCloudVPCObservation(booleanComb[0], booleanComb[1], booleanComb[2], booleanComb[3], booleanComb[4],
					booleanComb[5], booleanComb[6], booleanComb[7], booleanComb[8], booleanComb[9],
					booleanComb[10], booleanComb[11], booleanComb[12], booleanComb[13], booleanComb[14],
					booleanComb[15], booleanComb[16], booleanComb[17], booleanComb[18], booleanComb[19],
					booleanComb[20], booleanComb[21], booleanComb[22], booleanComb[23], booleanComb[24],
					booleanComb[25], booleanComb[26], booleanComb[27], booleanComb[28], booleanComb[29],
<<<<<<< HEAD
					booleanComb[30], booleanComb[31], booleanComb[32], booleanComb[33])
=======
					booleanComb[30], booleanComb[31])
>>>>>>> 90b51371a5c47187a61bd83aa140e7bea15ff394

				if tc.setName {
					cloudVPC.Name = &randomName
				} else {
					cloudVPC.Name = nil
				}

				if tc.setReourceGroup {
					cloudVPC.ResourceGroup = &ibmVPC.ResourceGroupReference{
						ID: &resourceGroupName,
					}
				} else {
					cloudVPC.ResourceGroup = nil
				}

				if _, err := LateInitializeSpec(crossplaneVPCInfo.DeepCopy(), &cloudVPC); err != nil {
					t.Errorf(fullTstName+": got error in LateInitializeSpec:\n%s", err)

					return
				}

				if diff := cmp.Diff(tc.spec, tc.expects); diff != "" {
					t.Errorf(fullTstName+": -wanted, +got:\n%s", diff)
				}
			})
		}
	}
}
