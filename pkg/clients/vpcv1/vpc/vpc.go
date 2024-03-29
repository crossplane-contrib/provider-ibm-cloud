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
	"fmt"

	ibmVPC "github.com/IBM/vpc-go-sdk/vpcv1"
	"github.com/google/go-cmp/cmp"

	"github.com/crossplane-contrib/provider-ibm-cloud/apis/vpcv1/v1alpha1"
	ibmc "github.com/crossplane-contrib/provider-ibm-cloud/pkg/clients"

	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/reference"
)

// LateInitializeSpec fills optional and unassigned fields with the values in the spec, from the info that comes from the cloud
//
// Params
// 	  spec - what we get from k8s
// 	  fromIBMCloud - ...what comes from the cloud
//
// Returns
//    whether late initialization happened
//    currently, always nil
func LateInitializeSpec(spec *v1alpha1.VPCParameters, fromIBMCloud *ibmVPC.VPC) (bool, error) { // nolint:gocyclo
	result := false

	if spec.Name == nil {
		if fromIBMCloud.Name != nil {
			spec.Name = new(string)
			*spec.Name = *fromIBMCloud.Name

			result = true
		}
	} else {
		if *spec.Name == "" {
			if fromIBMCloud.Name != nil && *fromIBMCloud.Name != "" {
				*spec.Name = *fromIBMCloud.Name

				result = true
			}
		}
	}

	if spec.ResourceGroup == nil {
		if fromIBMCloud.ResourceGroup != nil && fromIBMCloud.ResourceGroup.ID != nil {
			spec.ResourceGroup = &v1alpha1.ResourceGroupIdentity{
				ID: reference.FromPtrValue(fromIBMCloud.ResourceGroup.ID),
			}

			result = true
		}
	} else {
		if spec.ResourceGroup.ID == "" {
			if fromIBMCloud.ResourceGroup != nil && fromIBMCloud.ResourceGroup.ID != nil && *fromIBMCloud.ResourceGroup.ID != "" {
				spec.ResourceGroup.ID = reference.FromPtrValue(fromIBMCloud.ResourceGroup.ID)

				result = true
			}
		}
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
func GenerateObservation(in *ibmVPC.VPC) (v1alpha1.VPCObservation, error) { // nolint:gocyclo
	result := v1alpha1.VPCObservation{
		ClassicAccess: ibmc.BoolValue(in.ClassicAccess),
		CreatedAt:     ibmc.DateTimeToMetaV1Time(in.CreatedAt),
		CRN:           reference.FromPtrValue(in.CRN),
		Href:          reference.FromPtrValue(in.Href),
		ID:            reference.FromPtrValue(in.ID),
		Name:          reference.FromPtrValue(in.Name),
		Status:        reference.FromPtrValue(in.Status),
	}

	if len(in.CseSourceIps) > 0 {
		result.CseSourceIps = []v1alpha1.VpccseSourceIP{}
		for _, cl := range in.CseSourceIps {
			if (cl.IP != nil && cl.IP.Address != nil) ||
				(cl.Zone != nil && (cl.Zone.Href != nil || cl.Zone.Name != nil)) {

				newVSIP := v1alpha1.VpccseSourceIP{}

				if cl.IP != nil && cl.IP.Address != nil {
					newVSIP.IP = v1alpha1.IP{
						Address: reference.FromPtrValue(cl.IP.Address),
					}
				}

				if cl.Zone != nil && (cl.Zone.Href != nil || cl.Zone.Name != nil) {
					newVSIP.Zone = v1alpha1.ZoneReference{
						Href: reference.FromPtrValue(cl.Zone.Href),
						Name: reference.FromPtrValue(cl.Zone.Name),
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
			result.DefaultNetworkACL = v1alpha1.NetworkACLReference{
				CRN:  reference.FromPtrValue(in.DefaultNetworkACL.CRN),
				Href: reference.FromPtrValue(in.DefaultNetworkACL.Href),
				ID:   reference.FromPtrValue(in.DefaultNetworkACL.ID),
				Name: reference.FromPtrValue(in.DefaultNetworkACL.Name),
			}

			if in.DefaultNetworkACL.Deleted != nil && in.DefaultNetworkACL.Deleted.MoreInfo != nil {
				result.DefaultNetworkACL.Deleted = &v1alpha1.NetworkACLReferenceDeleted{
					MoreInfo: reference.FromPtrValue(in.DefaultNetworkACL.Deleted.MoreInfo),
				}
			}
		}
	}

	if in.DefaultRoutingTable != nil {
		if (in.DefaultRoutingTable.Deleted != nil && in.DefaultRoutingTable.Deleted.MoreInfo != nil) ||
			in.DefaultRoutingTable.Href != nil ||
			in.DefaultRoutingTable.Name != nil ||
			in.DefaultRoutingTable.ResourceType != nil {
			result.DefaultRoutingTable = v1alpha1.RoutingTableReference{
				Href:         reference.FromPtrValue(in.DefaultRoutingTable.Href),
				ID:           reference.FromPtrValue(in.DefaultRoutingTable.ID),
				Name:         reference.FromPtrValue(in.DefaultRoutingTable.Name),
				ResourceType: reference.FromPtrValue(in.DefaultRoutingTable.ResourceType),
			}

			if in.DefaultRoutingTable.Deleted != nil && in.DefaultRoutingTable.Deleted.MoreInfo != nil {
				result.DefaultRoutingTable.Deleted = &v1alpha1.RoutingTableReferenceDeleted{
					MoreInfo: reference.FromPtrValue(in.DefaultRoutingTable.Deleted.MoreInfo),
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
			result.DefaultSecurityGroup = v1alpha1.SecurityGroupReference{
				CRN:  reference.FromPtrValue(in.DefaultSecurityGroup.CRN),
				Href: reference.FromPtrValue(in.DefaultSecurityGroup.Href),
				ID:   reference.FromPtrValue(in.DefaultSecurityGroup.ID),
				Name: reference.FromPtrValue(in.DefaultSecurityGroup.Name),
			}

			if in.DefaultSecurityGroup.Deleted != nil && in.DefaultSecurityGroup.Deleted.MoreInfo != nil {
				result.DefaultSecurityGroup.Deleted = &v1alpha1.SecurityGroupReferenceDeleted{
					MoreInfo: reference.FromPtrValue(in.DefaultSecurityGroup.Deleted.MoreInfo),
				}
			}
		}
	}

	if in.ResourceGroup != nil {
		if in.ResourceGroup.Name != nil || in.ResourceGroup.Href != nil || in.ResourceGroup.ID != nil {
			result.ResourceGroup = v1alpha1.ResourceGroupReference{
				Name: reference.FromPtrValue(in.ResourceGroup.Name),
				Href: reference.FromPtrValue(in.ResourceGroup.Href),
				ID:   reference.FromPtrValue(in.ResourceGroup.ID),
			}
		}
	}

	return result, nil
}

// GenerateCreateOptions returns a cloud-compliant version of the VPC creation parameters
//
// Params
//    in - the creation options, crossplane style
//
//  Returns
//     the struct to use in the cloud call
func GenerateCreateOptions(in *v1alpha1.VPCParameters) (ibmVPC.CreateVPCOptions, error) {
	result := ibmVPC.CreateVPCOptions{
		AddressPrefixManagement: in.AddressPrefixManagement,
		ClassicAccess:           &in.ClassicAccess,
		Name:                    in.Name,
	}

	dc := in.DeepCopy()

	if dc.AddressPrefixManagement != nil {
		result.SetAddressPrefixManagement(*dc.AddressPrefixManagement)
	}

	if dc.Name != nil {
		result.SetName(*dc.Name)
	}

	result.SetClassicAccess(dc.ClassicAccess)

	if dc.ResourceGroup != nil {
		result.SetResourceGroup(&ibmVPC.ResourceGroupIdentity{
			ID: reference.ToPtrValue(dc.ResourceGroup.ID),
		})
	}

	return result, nil
}

// IsUpToDate checks whether the current VPC config (in the cloud) is up-to-date compared to the crossplane one (only
// the name is checked, as this is the only one that can be updated).
//
// No error is currently returned
func IsUpToDate(crossplane *v1alpha1.VPCParameters, observed *ibmVPC.VPC, l logging.Logger) (bool, error) {
	result := false

	if crossplane.Name == nil && observed.Name == nil {
		result = true
	} else if crossplane.Name != nil && observed.Name != nil {
		if result = (*crossplane.Name == *observed.Name); !result {
			diff := cmp.Diff(observed.Name, crossplane.Name)
			fmt.Printf(">>> %s\n", diff)
			l.Info("IsUpToDate", "Diff", diff)

			return false, nil
		}
	}

	return result, nil
}
