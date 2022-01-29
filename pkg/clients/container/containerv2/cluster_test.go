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
	"strconv"
	"testing"

	ibmContainerV2 "github.com/IBM-Cloud/bluemix-go/api/container/containerv2"

	"github.com/crossplane-contrib/provider-ibm-cloud/apis/container/containerv2/v1alpha1"
	ibmc "github.com/crossplane-contrib/provider-ibm-cloud/pkg/clients"

	"github.com/google/go-cmp/cmp"
)

// Tests the GenerateCrossplaneClusterInfo function
func TestGenerateCrossplaneClusterInfo(t *testing.T) {
	ibmClusterInfo := GetContainerClusterInfo()

	t.Run("TestGenerateCrossplaneClusterInfo", func(t *testing.T) {
		crossPlaneClusterInfo, _ := GenerateCrossplaneClusterInfo(ibmClusterInfo)

		tests := map[string]struct {
			cloudVal            string
			crossplaneVal       string
			cloudValAddons      []ibmContainerV2.Addon
			crossplaneValAddons []v1alpha1.Addon
			isDateStr           bool `default:"false"`
		}{
			"CreatedDate": {
				cloudVal:      ibmClusterInfo.CreatedDate,
				crossplaneVal: strconv.FormatInt(crossPlaneClusterInfo.CreatedDate.Unix(), 10),
				isDateStr:     true,
			},
			"DataCenter": {
				cloudVal:      ibmClusterInfo.DataCenter,
				crossplaneVal: crossPlaneClusterInfo.DataCenter,
			},
			"ID": {
				cloudVal:      ibmClusterInfo.ID,
				crossplaneVal: crossPlaneClusterInfo.ID,
			},
			"Location": {
				cloudVal:      ibmClusterInfo.Location,
				crossplaneVal: crossPlaneClusterInfo.Location,
			},
			"Entitlement": {
				cloudVal:      ibmClusterInfo.Entitlement,
				crossplaneVal: crossPlaneClusterInfo.Entitlement,
			},
			"MasterKubeVersion": {
				cloudVal:      ibmClusterInfo.MasterKubeVersion,
				crossplaneVal: crossPlaneClusterInfo.MasterKubeVersion,
			},
			"Name": {
				cloudVal:      ibmClusterInfo.Name,
				crossplaneVal: crossPlaneClusterInfo.Name,
			},
			"Region": {
				cloudVal:      ibmClusterInfo.Region,
				crossplaneVal: crossPlaneClusterInfo.Region,
			},
			"ResourceGroupID": {
				cloudVal:      ibmClusterInfo.ResourceGroupID,
				crossplaneVal: crossPlaneClusterInfo.ResourceGroupID,
			},
			"State": {
				cloudVal:      ibmClusterInfo.State,
				crossplaneVal: crossPlaneClusterInfo.State,
			},
			"IsPaid": {
				cloudVal:      strconv.FormatBool(ibmClusterInfo.IsPaid),
				crossplaneVal: strconv.FormatBool(crossPlaneClusterInfo.IsPaid),
			},
			"Addons": {
				cloudValAddons:      ibmClusterInfo.Addons,
				crossplaneValAddons: crossPlaneClusterInfo.Addons,
			},
			"OwnerEmail": {
				cloudVal:      ibmClusterInfo.OwnerEmail,
				crossplaneVal: crossPlaneClusterInfo.OwnerEmail,
			},
			"Type": {
				cloudVal:      ibmClusterInfo.Type,
				crossplaneVal: crossPlaneClusterInfo.Type,
			},
			"TargetVersion": {
				cloudVal:      ibmClusterInfo.TargetVersion,
				crossplaneVal: crossPlaneClusterInfo.TargetVersion,
			},
			"ServiceSubnet": {
				cloudVal:      ibmClusterInfo.ServiceSubnet,
				crossplaneVal: crossPlaneClusterInfo.ServiceSubnet,
			},
			"ResourceGroupName": {
				cloudVal:      ibmClusterInfo.ResourceGroupName,
				crossplaneVal: crossPlaneClusterInfo.ResourceGroupName,
			},
			"Provider": {
				cloudVal:      ibmClusterInfo.Provider,
				crossplaneVal: crossPlaneClusterInfo.Provider,
			},
			"PodSubnet": {
				cloudVal:      ibmClusterInfo.PodSubnet,
				crossplaneVal: crossPlaneClusterInfo.PodSubnet,
			},
			"MultiAzCapable": {
				cloudVal:      strconv.FormatBool(ibmClusterInfo.MultiAzCapable),
				crossplaneVal: strconv.FormatBool(crossPlaneClusterInfo.MultiAzCapable),
			},
			"APIUser": {
				cloudVal:      ibmClusterInfo.APIUser,
				crossplaneVal: crossPlaneClusterInfo.APIUser,
			},
			"MasterURL": {
				cloudVal:      ibmClusterInfo.MasterURL,
				crossplaneVal: crossPlaneClusterInfo.MasterURL,
			},
			"DisableAutoUpdate": {
				cloudVal:      strconv.FormatBool(ibmClusterInfo.DisableAutoUpdate),
				crossplaneVal: strconv.FormatBool(crossPlaneClusterInfo.DisableAutoUpdate),
			},
			"PrivateServiceEndpointEnabled": {
				cloudVal:      strconv.FormatBool(ibmClusterInfo.ServiceEndpoints.PrivateServiceEndpointEnabled),
				crossplaneVal: strconv.FormatBool(crossPlaneClusterInfo.ServiceEndpoints.PrivateServiceEndpointEnabled),
			},
			"PrivateServiceEndpointURL": {
				cloudVal:      ibmClusterInfo.ServiceEndpoints.PrivateServiceEndpointURL,
				crossplaneVal: crossPlaneClusterInfo.ServiceEndpoints.PrivateServiceEndpointURL,
			},
			"PublicServiceEndpointEnabled": {
				cloudVal:      strconv.FormatBool(ibmClusterInfo.ServiceEndpoints.PublicServiceEndpointEnabled),
				crossplaneVal: strconv.FormatBool(crossPlaneClusterInfo.ServiceEndpoints.PublicServiceEndpointEnabled),
			},
			"PublicServiceEndpointURL": {
				cloudVal:      ibmClusterInfo.ServiceEndpoints.PublicServiceEndpointURL,
				crossplaneVal: crossPlaneClusterInfo.ServiceEndpoints.PublicServiceEndpointURL,
			},
			"ModifiedDate": {
				cloudVal:      ibmClusterInfo.Lifecycle.ModifiedDate,
				crossplaneVal: strconv.FormatInt(crossPlaneClusterInfo.Lifecycle.ModifiedDate.Unix(), 10),
				isDateStr:     true,
			},
			"MasterStatus": {
				cloudVal:      ibmClusterInfo.Lifecycle.MasterStatus,
				crossplaneVal: crossPlaneClusterInfo.Lifecycle.MasterStatus,
			},
			"MasterStatusModifiedDate": {
				cloudVal:      ibmClusterInfo.Lifecycle.MasterStatusModifiedDate,
				crossplaneVal: strconv.FormatInt(crossPlaneClusterInfo.Lifecycle.MasterStatusModifiedDate.Unix(), 10),
				isDateStr:     true,
			},
			"MasterHealth": {
				cloudVal:      ibmClusterInfo.Lifecycle.MasterHealth,
				crossplaneVal: crossPlaneClusterInfo.Lifecycle.MasterHealth,
			},
			"MasterState": {
				cloudVal:      ibmClusterInfo.Lifecycle.MasterState,
				crossplaneVal: crossPlaneClusterInfo.Lifecycle.MasterState,
			},
			"WorkerCount": {
				cloudVal:      strconv.Itoa(ibmClusterInfo.WorkerCount),
				crossplaneVal: strconv.Itoa(crossPlaneClusterInfo.WorkerCount),
			},
			"HostName": {
				cloudVal:      ibmClusterInfo.Ingress.HostName,
				crossplaneVal: crossPlaneClusterInfo.Ingress.HostName,
			},
			"SecretName": {
				cloudVal:      ibmClusterInfo.Ingress.SecretName,
				crossplaneVal: crossPlaneClusterInfo.Ingress.SecretName,
			},
			"KeyProtectEnabled": {
				cloudVal:      strconv.FormatBool(ibmClusterInfo.Features.KeyProtectEnabled),
				crossplaneVal: strconv.FormatBool(crossPlaneClusterInfo.Features.KeyProtectEnabled),
			},
		}

		for name, tc := range tests {
			t.Run(name, func(t *testing.T) {
				if !tc.isDateStr {
					if diff := cmp.Diff(tc.cloudVal, tc.crossplaneVal); diff != "" {
						t.Errorf("TestGenerateCrossplaneClusterInfo(...): -wanted, +got:\n%s", diff)
					}
				} else {
					cloudVal := ibmc.ParseMetaV1Time(tc.cloudVal).Unix()
					crossPlaneVal, _ := strconv.ParseInt(tc.crossplaneVal, 10, 64)

					if cloudVal != crossPlaneVal {
						t.Errorf("TestGenerateCrossplaneClusterInfo(...): -wanted %s, +got:%s\n", tc.cloudVal, tc.crossplaneVal)
					}
				}
			})
		}
	})
}

// Tests the GenerateClusterCreateRequest function
func TestGenerateClusterCreateRequest(t *testing.T) {
	crossplaneRequest := GetClusterCreateCrossplaneRequest()
	ibmCloudRequest := &ibmContainerV2.ClusterCreateRequest{}

	t.Run("TestGenerateClusterCreateRequest", func(t *testing.T) {
		err := GenerateClusterCreateRequest(crossplaneRequest, ibmCloudRequest)
		if err != nil {
			t.Errorf("TestGenerateClusterCreateRequest(...): -returned error\n%s", err.Error())

			return
		}

		tests := map[string]struct {
			crossplaneVal    string
			cloudVal         string
			isDateStr        bool `default:"false"`
			crossplaneLabels *map[string]string
			cloudLabels      map[string]string
		}{
			"DisablePublicServiceEndpoint": {
				crossplaneVal: strconv.FormatBool(crossplaneRequest.DisablePublicServiceEndpoint),
				cloudVal:      strconv.FormatBool(ibmCloudRequest.DisablePublicServiceEndpoint),
			},
			"KubeVersion": {
				crossplaneVal: crossplaneRequest.KubeVersion,
				cloudVal:      ibmCloudRequest.KubeVersion,
			},
			"Billing": {
				crossplaneVal: *crossplaneRequest.Billing,
				cloudVal:      ibmCloudRequest.Billing,
			},
			"PodSubnet": {
				crossplaneVal: crossplaneRequest.PodSubnet,
				cloudVal:      ibmCloudRequest.PodSubnet,
			},
			"Provider": {
				crossplaneVal: crossplaneRequest.Provider,
				cloudVal:      ibmCloudRequest.Provider,
			},
			"ServiceSubnet": {
				crossplaneVal: crossplaneRequest.ServiceSubnet,
				cloudVal:      ibmCloudRequest.ServiceSubnet,
			},
			"Name": {
				crossplaneVal: crossplaneRequest.Name,
				cloudVal:      ibmCloudRequest.Name,
			},
			"DefaultWorkerPoolEntitlement": {
				crossplaneVal: crossplaneRequest.DefaultWorkerPoolEntitlement,
				cloudVal:      ibmCloudRequest.DefaultWorkerPoolEntitlement,
			},
			"CosInstanceCRN": {
				crossplaneVal: crossplaneRequest.CosInstanceCRN,
				cloudVal:      ibmCloudRequest.CosInstanceCRN,
			},
			"DiskEncryption": {
				crossplaneVal: strconv.FormatBool(*crossplaneRequest.WorkerPools.DiskEncryption),
				cloudVal:      strconv.FormatBool(ibmCloudRequest.WorkerPools.DiskEncryption),
			},
			"Entitlement": {
				crossplaneVal: crossplaneRequest.WorkerPools.Entitlement,
				cloudVal:      ibmCloudRequest.WorkerPools.Entitlement,
			},
			"Flavor": {
				crossplaneVal: crossplaneRequest.WorkerPools.Flavor,
				cloudVal:      ibmCloudRequest.WorkerPools.Flavor,
			},
			"Isolation": {
				crossplaneVal: *crossplaneRequest.WorkerPools.Isolation,
				cloudVal:      ibmCloudRequest.WorkerPools.Isolation,
			},
			"Labels": {
				crossplaneLabels: crossplaneRequest.WorkerPools.Labels,
				cloudLabels:      ibmCloudRequest.WorkerPools.Labels,
			},
			"WorkerPools.Name": {
				crossplaneVal: crossplaneRequest.WorkerPools.Name,
				cloudVal:      ibmCloudRequest.WorkerPools.Name,
			},
			"VpcID": {
				crossplaneVal: crossplaneRequest.WorkerPools.VpcID,
				cloudVal:      ibmCloudRequest.WorkerPools.VpcID,
			},
			"WorkerCount": {
				crossplaneVal: strconv.Itoa(crossplaneRequest.WorkerPools.WorkerCount),
				cloudVal:      strconv.Itoa(ibmCloudRequest.WorkerPools.WorkerCount),
			},
		}

		for name, tc := range tests {
			t.Run(name, func(t *testing.T) {
				if !tc.isDateStr {
					if diff := cmp.Diff(tc.crossplaneVal, tc.cloudVal); diff != "" {
						t.Errorf("TestGenerateClusterCreateRequest(...): -wanted, +got:\n%s", diff)
					}
				} else {
					cloudVal := ibmc.ParseMetaV1Time(tc.cloudVal).Unix()
					crossPlaneVal, _ := strconv.ParseInt(tc.crossplaneVal, 10, 64)

					if cloudVal != crossPlaneVal {
						t.Errorf("TestGenerateClusterCreateRequest(...): -wanted %s, +got:%s\n", tc.crossplaneVal, tc.cloudVal)
					}
				}
			})
		}
	})
}
