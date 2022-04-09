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

package vpc

import (
	"reflect"
	"strconv"
	"strings"
	"testing"

	ibmVPC "github.com/IBM/vpc-go-sdk/vpcv1"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/reference"

	"github.com/crossplane-contrib/provider-ibm-cloud/apis/vpcv1/v1alpha1"
	"github.com/crossplane-contrib/provider-ibm-cloud/pkg/clients/vpcv1"

	"github.com/google/go-cmp/cmp"
)

// Params
//    value - a value. May be nil
//
// Returns
//    the value of the parameter, (of the appopriate type, dereferenced if a pointer), or nil
func typeVal(value interface{}) interface{} {
	if value == nil {
		return nil
	}

	result := vpcv1.TypeVal(value)

	if result == nil && strings.Contains(reflect.TypeOf(value).String(), ".ResourceGroupIdentity") &&
		!reflect.ValueOf(value).IsNil() {
		switch typed := value.(type) {
		case *v1alpha1.ResourceGroupIdentity:
			result = *typed
		case *ibmVPC.ResourceGroupIdentity:
			result = *typed
		case *ibmVPC.ResourceGroupIdentityByID:
			result = *typed
		}
	}

	return result
}

// Params
//    crossplaneRGIntf - a crossplane resource group (must be of type v1alpha1.ResourceGroupIdentityAlsoByID; interface for convenience). May be nil
//    cloudRGIntf - a cloud resource group (must be of type ibmVPC.ResourceGroupIdentityIntf; interface for convenience). May be nil
//
//    Note that both params should NOT be at the same time nil (or point to nil structures)
//
// Returns
//    crossplaneID, cloudID
func getResourceIds(crossplaneRGIntf interface{}, cloudRGIntf interface{}) (string, string) {
	var crossplaneID, cloudID string

	if crossplaneRGIntf != nil && cloudRGIntf != nil &&
		reflect.TypeOf(crossplaneRGIntf).String() == "v1alpha1.ResourceGroupIdentity" {
		crossplaneRG := crossplaneRGIntf.(v1alpha1.ResourceGroupIdentity)
		crossplaneID = crossplaneRG.ID

		cloudRG, ok := cloudRGIntf.(ibmVPC.ResourceGroupIdentity)
		if ok && cloudRG.ID != nil {
			cloudID = reference.FromPtrValue(cloudRG.ID)
		} else {
			cloudRGByID, ok := cloudRGIntf.(ibmVPC.ResourceGroupIdentityByID)
			if ok && cloudRGByID.ID != nil {
				cloudID = reference.FromPtrValue(cloudRGByID.ID)
			}
		}
	}

	return crossplaneID, cloudID
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
				spec:            *orig.DeepCopy(),
				setName:         nameIsNil,
				setReourceGroup: rgIsNil,
				expects:         withResourceGroup(withName(*orig.DeepCopy(), nameIsNil), rgIsNil),
			}
		}
	}

	return result
}

// Tests the GenerateObservation function
func TestGenerateObservation(t *testing.T) {
	functionTstName := "GenerateObservation"

	numVars := 10 // as many as the params of booleanComb we will be using
	for i, booleanComb := range vpcv1.GenerateSomePermutations(numVars, 32, true) {
		varCombinationLogging := vpcv1.GetBinaryRep(i, numVars)

		ibmVPCInfo := GetDummyObservation(
			booleanComb[0], booleanComb[1], booleanComb[2], booleanComb[3], booleanComb[4],
			booleanComb[5], booleanComb[6], booleanComb[7], booleanComb[8], booleanComb[9],
			booleanComb[10], booleanComb[11], booleanComb[12], booleanComb[13], booleanComb[14],
			booleanComb[15], booleanComb[16], booleanComb[17], booleanComb[18], booleanComb[19],
			booleanComb[20], booleanComb[21], booleanComb[22], booleanComb[23], booleanComb[24],
			booleanComb[25], booleanComb[26], booleanComb[27], booleanComb[28], booleanComb[29],
			booleanComb[30], booleanComb[31])
		crossplaneVPCInfo, err := GenerateObservation(&ibmVPCInfo)
		if err != nil {
			t.Errorf(functionTstName + " " + varCombinationLogging + ", GenerateObservation() returned error: " + err.Error())

			return
		}

		tests := createTestsObservation(&ibmVPCInfo, &crossplaneVPCInfo)
		for name, tc := range tests {
			t.Run(functionTstName, func(t *testing.T) {
				fullTstName := functionTstName + " " + varCombinationLogging + " " + name

				cloudVal := typeVal(tc.cloudVal)
				crossplaneVal := typeVal(tc.crossplaneVal)

				if diff := cmp.Diff(cloudVal, crossplaneVal); diff != "" {
					t.Errorf(fullTstName+": -wanted, +got:\n%s", diff)
				}
			})
		}
	}
}

// Tests the GenerateCreateOptions function
func TestGenerateCreateOptions(t *testing.T) {
	functionTstName := "GenerateCreateOptions"

	numVars := 3 // does not make sense to have more than the num of vars used...
	for i, booleanComb := range vpcv1.GeneratePermutations(numVars) {
		varCombinationLogging := vpcv1.GetBinaryRep(i, numVars)

		crossplaneVPCInfo := GetDummyCreateParams(booleanComb[0], booleanComb[1], booleanComb[2])
		ibmVPCInfo, err := GenerateCreateOptions(&crossplaneVPCInfo)
		if err != nil {
			t.Errorf(functionTstName + " " + varCombinationLogging + ", GenerateCreateOptions() returned error: " + err.Error())

			return
		}

		tests := createTestsCreateParams(&ibmVPCInfo, &crossplaneVPCInfo)
		for name, tc := range tests {
			t.Run(functionTstName, func(t *testing.T) {
				fullTstName := functionTstName + " " + varCombinationLogging + " " + name

				cloudVal := typeVal(tc.cloudVal)
				crossplaneVal := typeVal(tc.crossplaneVal)

				crossplaneID, cloudID := getResourceIds(crossplaneVal, cloudVal)
				if diff := cmp.Diff(crossplaneID, cloudID); diff != "" {
					t.Errorf(fullTstName+": -wanted, +got:\n%s", diff)
				}
			})
		}
	}
}

// Tests the IsUpToDate function
func TestIsUpToDate(t *testing.T) {
	functionTstName := "IsUpToDate"

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
				t.Errorf(functionTstName+" "+name+", IsUpToDate(...): -want:%t, +got:%t\n", tc.want, rc)
			}
		})
	}
}

// Tests the LateInitializeSpec function
func TestLateInitializeSpec(t *testing.T) {
	functionTstName := "LateInitializeSpec"

	numVars := 3 // does not make sense to have more than the num of vars used... If we put too many,
	// then testing timeouts (30 secs)
	for i, booleanComb := range vpcv1.GenerateSomePermutations(numVars, 32, true) {
		varCombinationLogging := vpcv1.GetBinaryRep(i, numVars)

		crossplaneVPCInfo := GetDummyCreateParams(booleanComb[0], booleanComb[1], booleanComb[2])

		tests := createLateInitializeSpecTests(crossplaneVPCInfo)
		for name, tc := range tests {
			t.Run(functionTstName, func(t *testing.T) {
				fullTstName := functionTstName + " " + varCombinationLogging + " " + name

				cloudVPC := GetDummyObservation(booleanComb[0], booleanComb[1], booleanComb[2], booleanComb[3], booleanComb[4],
					booleanComb[5], booleanComb[6], booleanComb[7], booleanComb[8], booleanComb[9],
					booleanComb[10], booleanComb[11], booleanComb[12], booleanComb[13], booleanComb[14],
					booleanComb[15], booleanComb[16], booleanComb[17], booleanComb[18], booleanComb[19],
					booleanComb[20], booleanComb[21], booleanComb[22], booleanComb[23], booleanComb[24],
					booleanComb[25], booleanComb[26], booleanComb[27], booleanComb[28], booleanComb[29],
					booleanComb[30], booleanComb[31])

				if tc.setName {
					cloudVPC.Name = &randomName
				} else {
					cloudVPC.Name = nil
				}

				if tc.setReourceGroup {
					cloudVPC.ResourceGroup = &ibmVPC.ResourceGroupReference{
						ID: &randomResourceGroupID,
					}
				} else {
					cloudVPC.ResourceGroup = nil
				}

				if _, err := LateInitializeSpec(&tc.spec, &cloudVPC); err != nil {
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
