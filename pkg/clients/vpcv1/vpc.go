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

	ibmVPC "github.com/IBM/vpc-go-sdk/vpcv1"

	ibmc "github.com/crossplane-contrib/provider-ibm-cloud/pkg/clients"
	"github.com/crossplane/crossplane-runtime/pkg/reference"

	"github.com/crossplane-contrib/provider-ibm-cloud/apis/vpcv1/v1alpha1"
)

// LateInitializeSpec fills optional and unassigned fields with the values in the spec, from the info that comes from the cloud
//
// Params
// 	  spec - what we get from k8s
// 	  fromIBMCloud - ...what comes from the cloud
//
// Returns
//    whether the resource was late-initialized, any error
func LateInitializeSpec(spec *v1alpha1.VPCParameters, fromIBMCloud *ibmVPC.CreateVPCOptions) (bool, error) {
	wasLateInitialized := false

	spec.AddressPrefixManagement, wasLateInitialized = ibmc.LateInitializeStr(spec.AddressPrefixManagement, fromIBMCloud.AddressPrefixManagement)
	spec.ClassicAccess, wasLateInitialized = ibmc.LateInitializeBool(spec.ClassicAccess, fromIBMCloud.ClassicAccess)
	spec.Name, wasLateInitialized = ibmc.LateInitializeStr(spec.Name, fromIBMCloud.Name)

	if spec.ResourceGroup == nil && fromIBMCloud.ResourceGroup != nil && !reflect.ValueOf(fromIBMCloud.ResourceGroup).IsNil() {
		spec.ResourceGroup = &v1alpha1.ResourceGroupIdentityAlsoByID{
			ID: fromIBMCloud.ResourceGroup.ID,
		}
	}

	return wasLateInitialized, nil
}

// GenerateCrossplaneVPCParams returns a crossplane version of the VPC creation parameters
//
// Params
//     in - the create options, in IBM-cloud-style
//
// Returns
//     the create options, crossplane-style
func GenerateCrossplaneVPCParams(in *ibmVPC.CreateVPCOptions) (v1alpha1.VPCParameters, error) {
	result := v1alpha1.VPCParameters{
		AddressPrefixManagement: in.AddressPrefixManagement,
		ClassicAccess:           in.ClassicAccess,
		Name:                    in.Name,
	}

	if len(in.Headers) > 0 {
		result.Headers = &in.Headers
	}

	if in.ResourceGroup != nil {
		ref, ok := in.ResourceGroup.(*ibmVPC.ResourceGroupIdentity)
		if ok && ref.ID != nil {
			result.ResourceGroup = &v1alpha1.ResourceGroupIdentityAlsoByID{
				ID:     *ref.ID,
				IsByID: false,
			}
		}

		refByID, ok := in.ResourceGroup.(*ibmVPC.ResourceGroupIdentityByID)
		if ok && refByID.ID != nil {
			result.ResourceGroup = &v1alpha1.ResourceGroupIdentityAlsoByID{
				ID:     *refByID.ID,
				IsByID: true,
			}
		}

	}

	return result, nil
}

// GenerateCloudVPCParams returns a cloud-compliant version of the VPC creation parameters
//
// Params
//    in - the creation options, crossplane style
func GenerateCloudVPCParams(in *v1alpha1.VPCParameters) (ibmVPC.CreateVPCOptions, error) {
	result := ibmVPC.CreateVPCOptions{
		AddressPrefixManagement: in.AddressPrefixManagement,
		ClassicAccess:           in.ClassicAccess,
		Name:                    in.Name,
	}

	if in.Headers != nil && len(*in.Headers) > 0 {
		result.SetHeaders(*in.DeepCopy().Headers)
	}

	if in.ResourceGroup != nil {
		if in.ResourceGroup.IsByID {
			result.ResourceGroup = &ibmVPC.ResourceGroupIdentityByID{
				ID: reference.ToPtrValue(in.ResourceGroup.ID),
			}
		} else {
			result.ResourceGroup = &ibmVPC.ResourceGroupIdentity{
				ID: reference.ToPtrValue(in.ResourceGroup.ID),
			}
		}
	}

	return result, nil
}
