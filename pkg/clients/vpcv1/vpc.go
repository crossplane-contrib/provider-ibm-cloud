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

	"github.com/crossplane-contrib/provider-ibm-cloud/apis/vpcv1/v1alpha1"
	ibmc "github.com/crossplane-contrib/provider-ibm-cloud/pkg/clients"

	"github.com/crossplane/crossplane-runtime/pkg/reference"
)

// LateInitializeSpec fills optional and unassigned fields with the values in the spec, from the info that comes from the cloud
//
// Params
// 	  spec - what we get from k8s
// 	  fromIBMCloud - ...what comes from the cloud
//
// Returns
//    whether the resource was late-initialized, any error
func LateInitializeSpec(spec *v1alpha1.VPCParameters, fromIBMCloud *ibmVPC.VPC) (bool, error) {
	wasLateInitializedCA := false
	wasLateInitializedName := false
	wasLateInitializedNameRG := false

	if spec.ClassicAccess == nil && fromIBMCloud.ClassicAccess != nil {
		spec.ClassicAccess = fromIBMCloud.ClassicAccess

		wasLateInitializedCA = true
	}

	if spec.Name == nil && fromIBMCloud.Name != nil {
		spec.Name = fromIBMCloud.Name

		wasLateInitializedName = true
	}

	if spec.ResourceGroup == nil && fromIBMCloud.ResourceGroup != nil && !reflect.ValueOf(fromIBMCloud.ResourceGroup).IsNil() {
		spec.ResourceGroup = &v1alpha1.ResourceGroupIdentity{
			ID: *fromIBMCloud.ResourceGroup.ID,
		}

		wasLateInitializedNameRG = true
	}

	return wasLateInitializedCA || wasLateInitializedName || wasLateInitializedNameRG, nil
}

// GenerateCrossplaneVPCObservation returns a crossplane version of the cloud observation results parameters
//
// Params
//     in - the create options, in IBM-cloud-style
//
// Returns
//     the status, crossplane-style
func GenerateCrossplaneVPCObservation(in *ibmVPC.VPC) (v1alpha1.VPCObservation, error) { // nolint:gocyclo
	result := v1alpha1.VPCObservation{
		ClassicAccess: in.ClassicAccess,
		CreatedAt:     ibmc.DateTimeToMetaV1Time(in.CreatedAt),
		CRN:           in.CRN,
		Href:          in.Href,
		ID:            in.ID,
		Name:          in.Name,
		Status:        in.Status,
	}

	if len(in.CseSourceIps) > 0 {
		result.CseSourceIps = []v1alpha1.VpccseSourceIP{}
		for _, cl := range in.CseSourceIps {
			if (cl.IP != nil && cl.IP.Address != nil) ||
				(cl.Zone != nil && (cl.Zone.Href != nil || cl.Zone.Name != nil)) {

				newVSIP := v1alpha1.VpccseSourceIP{}

				if cl.IP != nil && cl.IP.Address != nil {
					newVSIP.IP = &v1alpha1.IP{
						Address: *cl.IP.Address,
					}
				}

				if cl.Zone != nil && (cl.Zone.Href != nil || cl.Zone.Name != nil) {
					newVSIP.Zone = &v1alpha1.ZoneReference{
						Href: cl.Zone.Href,
						Name: cl.Zone.Name,
					}
				}

				result.CseSourceIps = append(result.CseSourceIps, newVSIP)
			}
		}
	}

	if in.DefaultNetworkACL != nil {
		if in.DefaultNetworkACL.CRN != nil ||
			(in.DefaultNetworkACL.Deleted != nil && in.DefaultNetworkACL.Deleted.MoreInfo != nil) ||
			in.DefaultNetworkACL.Href != nil ||
			in.DefaultNetworkACL.ID != nil ||
			in.DefaultNetworkACL.Name != nil {
			result.DefaultNetworkACL = &v1alpha1.NetworkACLReference{
				CRN:  in.DefaultNetworkACL.CRN,
				Href: in.DefaultNetworkACL.Href,
				ID:   in.DefaultNetworkACL.ID,
				Name: in.DefaultNetworkACL.Name,
			}

			if in.DefaultNetworkACL.Deleted != nil && in.DefaultNetworkACL.Deleted.MoreInfo != nil {
				result.DefaultNetworkACL.Deleted = &v1alpha1.NetworkACLReferenceDeleted{
					MoreInfo: *in.DefaultNetworkACL.Deleted.MoreInfo,
				}
			}
		}
	}

	if in.DefaultRoutingTable != nil {
		if (in.DefaultRoutingTable.Deleted != nil && in.DefaultRoutingTable.Deleted.MoreInfo != nil) ||
			in.DefaultRoutingTable.Href != nil ||
			in.DefaultRoutingTable.Name != nil ||
			in.DefaultRoutingTable.ResourceType != nil {
			result.DefaultRoutingTable = &v1alpha1.RoutingTableReference{
				Href:         in.DefaultRoutingTable.Href,
				ID:           in.DefaultRoutingTable.ID,
				Name:         in.DefaultRoutingTable.Name,
				ResourceType: in.DefaultRoutingTable.ResourceType,
			}

			if in.DefaultRoutingTable.Deleted != nil && in.DefaultRoutingTable.Deleted.MoreInfo != nil {
				result.DefaultRoutingTable.Deleted = &v1alpha1.RoutingTableReferenceDeleted{
					MoreInfo: *in.DefaultRoutingTable.Deleted.MoreInfo,
				}
			}
		}
	}

	if in.DefaultSecurityGroup != nil {
		if in.DefaultSecurityGroup.CRN != nil ||
			(in.DefaultSecurityGroup.Deleted != nil && in.DefaultSecurityGroup.Deleted.MoreInfo != nil) ||
			in.DefaultSecurityGroup.Href != nil ||
			in.DefaultSecurityGroup.ID != nil ||
			in.DefaultSecurityGroup.Name != nil {
			result.DefaultSecurityGroup = &v1alpha1.SecurityGroupReference{
				CRN:  in.DefaultSecurityGroup.CRN,
				Href: in.DefaultSecurityGroup.Href,
				ID:   in.DefaultSecurityGroup.ID,
				Name: in.DefaultSecurityGroup.Name,
			}

			if in.DefaultSecurityGroup.Deleted != nil && in.DefaultSecurityGroup.Deleted.MoreInfo != nil {
				result.DefaultSecurityGroup.Deleted = &v1alpha1.SecurityGroupReferenceDeleted{
					MoreInfo: *in.DefaultSecurityGroup.Deleted.MoreInfo,
				}
			}
		}
	}

	if in.ResourceGroup != nil {
		if in.ResourceGroup.Name != nil || in.ResourceGroup.Href != nil || in.ResourceGroup.ID != nil {
			result.ResourceGroup = &v1alpha1.ResourceGroupReference{
				Name: in.ResourceGroup.Name,
				Href: in.ResourceGroup.Href,
				ID:   in.ResourceGroup.ID,
			}
		}
	}

	return result, nil
}

// GenerateCloudVPCParams returns a cloud-compliant version of the VPC creation parameters
//
// Params
//    in - the creation options, crossplane style
//
//  Returns
//     the struct to use in the cloud call
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
		result.ResourceGroup = &ibmVPC.ResourceGroupIdentity{
			ID: reference.ToPtrValue(in.ResourceGroup.ID),
		}
	}

	return result, nil
}
