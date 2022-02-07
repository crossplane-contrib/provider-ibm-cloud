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

package vpc1

/*
// GenerateCrossplaneVPCParams returns a crossplane version of the VPC creation parameters
func GenerateCrossplaneVPCParams(in *vpcv1.CreateVPCOptions) (v1alpha1.VPCParameters, error) {
	result := v1alpha1.VPCParameters{
		AddressPrefixManagement: in.AddressPrefixManagement,
		ClassicAccess:           in.ClassicAccess,
		Name:                    in.Name,
		Headers:                 &in.Headers,
	}

	if in.ResourceGroup != nil {
		ref, ok := in.ResourceGroup.(vpcv1.ResourceGroupIdentity)
		if ok && ref.ID != nil {
			result.ResourceGroup := &v1alpha1.ResourceGroupIdentityBoth{
				ID: *ref.ID,
				IsByID: false,
			}
		}

		ref, ok := in.ResourceGroup.(vpcv1.ResourceGroupIdentityByID)
		if ok && ref.ID != nil {
			result.ResourceGroup := &v1alpha1.ResourceGroupIdentityBoth{
				ID: *ref.ID,
				IsByID: true,
			}
		}
	}

	return result, nil
}

// GenerateClusterCreateRequest populates the 'out' object from the 'in' one
func GenerateClusterCreateRequest(in *v1alpha1.ClusterCreateRequest, out *ibmContainerV2.ClusterCreateRequest) error {
	out.DisablePublicServiceEndpoint = in.DisablePublicServiceEndpoint
	out.KubeVersion = in.KubeVersion
	out.Billing = reference.FromPtrValue(in.Billing)
	out.PodSubnet = in.PodSubnet
	out.Provider = in.Provider
	out.ServiceSubnet = in.ServiceSubnet
	out.Name = in.Name
	out.DefaultWorkerPoolEntitlement = in.DefaultWorkerPoolEntitlement
	out.CosInstanceCRN = in.CosInstanceCRN
	out.WorkerPools.DiskEncryption = ibmc.BoolValue(in.WorkerPools.DiskEncryption)
	out.WorkerPools.Entitlement = in.WorkerPools.Entitlement
	out.WorkerPools.Flavor = in.WorkerPools.Flavor
	out.WorkerPools.Isolation = reference.FromPtrValue(in.WorkerPools.Isolation)

	if in.WorkerPools.Labels != nil {
		out.WorkerPools.Labels = map[string]string{}

		for k, v := range *in.WorkerPools.Labels {
			out.WorkerPools.Labels[k] = v
		}
	}

	out.WorkerPools.Name = in.WorkerPools.Name
	out.WorkerPools.VpcID = in.WorkerPools.VpcID
	out.WorkerPools.WorkerCount = in.WorkerPools.WorkerCount

	if len(in.WorkerPools.Zones) > 0 {
		out.WorkerPools.Zones = make([]ibmContainerV2.Zone, len(in.WorkerPools.Zones))

		for i, zo := range in.WorkerPools.Zones {
			out.WorkerPools.Zones[i] = ibmContainerV2.Zone{
				ID:       reference.FromPtrValue(zo.ID),
				SubnetID: reference.FromPtrValue(zo.SubnetID),
			}
		}
	}

	return nil
}
*/
