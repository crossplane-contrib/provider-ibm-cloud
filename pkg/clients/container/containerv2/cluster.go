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

package containerv2

import (
	ibmContainerV2 "github.com/IBM-Cloud/bluemix-go/api/container/containerv2"
	"github.com/crossplane/crossplane-runtime/pkg/reference"

	"github.com/crossplane-contrib/provider-ibm-cloud/apis/container/containerv2/v1alpha1"
	ibmc "github.com/crossplane-contrib/provider-ibm-cloud/pkg/clients"
)

// GenerateCrossplaneClusterInfo returns a crossplane version of the Cluster info (built from the one returned by the IBM cloud)
func GenerateCrossplaneClusterInfo(in *ibmContainerV2.ClusterInfo) (v1alpha1.ClusterInfo, error) {
	result := v1alpha1.ClusterInfo{
		CreatedDate:       ibmc.ParseMetaV1Time(in.CreatedDate),
		DataCenter:        in.DataCenter,
		ID:                in.ID,
		Location:          in.Location,
		Entitlement:       in.Entitlement,
		MasterKubeVersion: in.MasterKubeVersion,
		Name:              in.Name,
		Region:            in.Region,
		ResourceGroupID:   in.ResourceGroupID,
		State:             in.State,
		IsPaid:            in.IsPaid,
		OwnerEmail:        in.OwnerEmail,
		Type:              in.Type,
		TargetVersion:     in.TargetVersion,
		ServiceSubnet:     in.ServiceSubnet,
		ResourceGroupName: in.ResourceGroupName,
		Provider:          in.Provider,
		PodSubnet:         in.PodSubnet,
		MultiAzCapable:    in.MultiAzCapable,
		APIUser:           in.APIUser,
		MasterURL:         in.MasterURL,
		DisableAutoUpdate: in.DisableAutoUpdate,
		WorkerZones:       in.WorkerZones,
		Vpcs:              in.Vpcs,
		CRN:               in.CRN,
		VersionEOS:        in.VersionEOS,
		ServiceEndpoints: v1alpha1.Endpoints{
			PrivateServiceEndpointEnabled: in.ServiceEndpoints.PrivateServiceEndpointEnabled,
			PrivateServiceEndpointURL:     in.ServiceEndpoints.PrivateServiceEndpointURL,
			PublicServiceEndpointEnabled:  in.ServiceEndpoints.PublicServiceEndpointEnabled,
			PublicServiceEndpointURL:      in.ServiceEndpoints.PublicServiceEndpointURL,
		},
		Lifecycle: v1alpha1.LifeCycleInfo{
			ModifiedDate:             ibmc.ParseMetaV1Time(in.Lifecycle.ModifiedDate),
			MasterStatus:             in.Lifecycle.MasterStatus,
			MasterStatusModifiedDate: ibmc.ParseMetaV1Time(in.Lifecycle.MasterStatusModifiedDate),
			MasterHealth:             in.Lifecycle.MasterHealth,
			MasterState:              in.Lifecycle.MasterState,
		},
		WorkerCount: in.WorkerCount,
		Ingress: v1alpha1.IngresInfo{
			HostName:   in.Ingress.HostName,
			SecretName: in.Ingress.SecretName,
		},
		Features: v1alpha1.Feat{
			KeyProtectEnabled: in.Features.KeyProtectEnabled,
			PullSecretApplied: in.Features.PullSecretApplied,
		},
	}

	if len(in.Addons) > 0 {
		result.Addons = make([]v1alpha1.Addon, len(in.Addons))
		for i, ao := range in.Addons {
			result.Addons[i] = v1alpha1.Addon{
				Name:    ao.Name,
				Version: ao.Version,
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
