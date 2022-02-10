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
	"math"
	"reflect"
	"strconv"
	"testing"

	ibmVPC "github.com/IBM/vpc-go-sdk/vpcv1"

	"github.com/crossplane-contrib/provider-ibm-cloud/apis/vpcv1/v1alpha1"

	"github.com/google/go-cmp/cmp"
)

const numVariables = 5

// All the combinations of variables' values (with null ones) to use for testing
var allBooleanCombinations = generateCombinations(numVariables)

// Params
//      i  - an integer >= 0
//      size  >= 2^i
//
// Returns
//      a string with binary representation of the integer, of length == size
func getBinaryRep(i int, size int) string {
	result := strconv.FormatInt(int64(i), 2)

	for len(result) < size {
		result = "0" + result
	}

	return result
}

// Returns all the combinations (of booleans) for a given number of elements
//
// Params
// 	  numElems - the number of elementds
//
// Returns
//    an array of boolean arrays
func generateCombinations(numElems int) [][]bool {
	result := make([][]bool, 0)

	for i := 0; i < int(math.Pow(2, float64(numElems))); i++ {
		str := getBinaryRep(i, numElems)

		boolArray := make([]bool, numElems)
		boolArrayIdx := len(boolArray) - 1
		for j := len(str) - 1; j >= 0; j-- {
			boolArray[boolArrayIdx] = (str[j] == '1')

			boolArrayIdx--
		}

		result = append(result, boolArray)
	}

	return result
}

// Params
//    value - a value. May be nil
//
// Returns
//    the  value of the parameter (of the appopriate type), or nil
func typeVal(value interface{}) interface{} {
	var result interface{}

	switch typed := value.(type) {
	case *string:
		result = typed
	case *bool:
		result = typed
	case *map[string]string:
		result = typed
	case map[string]string:
		result = typed
	case *v1alpha1.ResourceGroupIdentityAlsoByID:
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
// Returns
//    whether they point to the same underlying resource
func sameResource(crossplaneRGIntf interface{}, cloudRGIntf interface{}) bool {
	result := false

	crossplaneRG := crossplaneRGIntf.(*v1alpha1.ResourceGroupIdentityAlsoByID)
	if crossplaneRG.IsByID {
		cloudRGStructVal, ok := cloudRGIntf.(*ibmVPC.ResourceGroupIdentityByID)
		if ok && cloudRGStructVal.ID != nil {
			result = (crossplaneRG.ID == *cloudRGStructVal.ID)
		}
	} else {
		cloudRGStructVal, ok := cloudRGIntf.(*ibmVPC.ResourceGroupIdentity)
		if ok && cloudRGStructVal.ID != nil {
			result = (crossplaneRG.ID == *cloudRGStructVal.ID)
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
func createTests(ibmVPCInfo *ibmVPC.CreateVPCOptions, crossplaneVPCInfo *v1alpha1.VPCParameters) map[string]struct {
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

// Tests the GenerateCrossplaneVPCParams function
func TestGenerateCrossplaneVPCParams(t *testing.T) {
	functionTstName := "TestGenerateCrossplaneVPCParams"

	for i, booleanComb := range allBooleanCombinations {
		varCombination := getBinaryRep(i, numVariables)

		ibmVPCInfo := GetDummyCloudVPCParams(booleanComb[0], booleanComb[1], booleanComb[2], booleanComb[3], booleanComb[4])
		crossplaneVPCInfo, err := GenerateCrossplaneVPCParams(&ibmVPCInfo)
		if err != nil {
			t.Errorf(functionTstName + " " + varCombination + ": function GenerateCrossplaneVPCParams() returned error: " + err.Error())

			return
		}

		tests := createTests(&ibmVPCInfo, &crossplaneVPCInfo)
		for name, tc := range tests {
			t.Run(functionTstName, func(t *testing.T) {
				fullTstName := functionTstName + " " + varCombination + " " + name

				cloudVal := typeVal(tc.cloudVal)
				crossplaneVal := typeVal(tc.crossplaneVal)

				if (cloudVal == nil || reflect.ValueOf(cloudVal).IsNil()) &&
					(crossplaneVal == nil || reflect.ValueOf(crossplaneVal).IsNil()) {
					return
				}

				if reflect.TypeOf(crossplaneVal).String() == "*v1alpha1.ResourceGroupIdentityAlsoByID" {
					if !sameResource(crossplaneVal, cloudVal) {
						t.Errorf(fullTstName+": different IDs - cloudVal=%s, crossplaneVal=%s", cloudVal, crossplaneVal)
					}
				} else if reflect.TypeOf(crossplaneVal).String() == "*map[string]string" {
					if diff := cmp.Diff(cloudVal, *crossplaneVal.(*map[string]string)); diff != "" {
						t.Errorf(fullTstName+": -wanted, +got:\n%s", diff)
					}
				} else if diff := cmp.Diff(cloudVal, crossplaneVal); diff != "" {
					t.Errorf(fullTstName+": -wanted, +got:\n%s", diff)
				}
			})
		}
	}
}

// Tests the GenerateCloudVPCParams function
func TestGenerateCloudVPCParams(t *testing.T) {
	functionTstName := "TestGenerateCloudVPCParams"

	for i, booleanComb := range allBooleanCombinations {
		varCombination := getBinaryRep(i, numVariables)

		crossplaneVPCInfo := GetDummyCrossplaneVPCParams(booleanComb[0], booleanComb[1], booleanComb[2], booleanComb[3], booleanComb[4])
		ibmVPCInfo, err := GenerateCloudVPCParams(&crossplaneVPCInfo)
		if err != nil {
			t.Errorf(functionTstName + " " + varCombination + ": function GenerateCrossplaneVPCParams() returned error: " + err.Error())

			return
		}

		tests := createTests(&ibmVPCInfo, &crossplaneVPCInfo)
		for name, tc := range tests {
			t.Run(functionTstName, func(t *testing.T) {
				fullTstName := functionTstName + " " + varCombination + " " + name

				cloudVal := typeVal(tc.cloudVal)
				crossplaneVal := typeVal(tc.crossplaneVal)

				if (cloudVal == nil || reflect.ValueOf(cloudVal).IsNil()) &&
					(crossplaneVal == nil || reflect.ValueOf(crossplaneVal).IsNil()) {
					return
				}

				if reflect.TypeOf(crossplaneVal).String() == "*v1alpha1.ResourceGroupIdentityAlsoByID" {
					if !sameResource(crossplaneVal, cloudVal) {
						t.Errorf(fullTstName+": different IDs - cloudVal=%s, crossplaneVal=%s", cloudVal, crossplaneVal)
					}
				} else if reflect.TypeOf(crossplaneVal).String() == "*map[string]string" {
					if diff := cmp.Diff(*crossplaneVal.(*map[string]string), cloudVal); diff != "" {
						t.Errorf(fullTstName+": -wanted, +got:\n%s", diff)
					}
				} else if diff := cmp.Diff(crossplaneVal, cloudVal); diff != "" {
					t.Errorf(fullTstName+": -wanted, +got:\n%s", diff)
				}
			})
		}
	}
}
