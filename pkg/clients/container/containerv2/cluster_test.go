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
	"github.com/crossplane/crossplane-runtime/pkg/reference"

	"github.com/crossplane-contrib/provider-ibm-cloud/apis/container/containerv2/v1alpha1"
	ibmc "github.com/crossplane-contrib/provider-ibm-cloud/pkg/clients"

	"github.com/google/go-cmp/cmp"
)

// Tests the GenerateCrossplaneClusterInfo function
func TestGenerateCrossplaneClusterInfo(t *testing.T) {
	ibmClusterInfo := &ibmContainerV2.ClusterInfo{
		CreatedDate:       "2006-01-02 15:04:05",
		DataCenter:        "a data center",
		ID:                "an id",
		Location:          "location-location",
		Entitlement:       "young people these days are like that",
		MasterKubeVersion: "grand master",
		Name:              "harry",
		Region:            "a region",
		ResourceGroupID:   "what an id",
		State:             "doing great!",
		IsPaid:            false,
		Addons:            []ibmContainerV2.Addon{{Name: "name1", Version: "version 1"}, {Name: "name1", Version: "version 2"}},
		OwnerEmail:        "foo@bar",
		Type:              "residential",
		TargetVersion:     "a version",
		ServiceSubnet:     "a subnet",
		ResourceGroupName: "a resource group",
		Provider:          "a provider",
		PodSubnet:         "a pod subnet",
		MultiAzCapable:    false,
		APIUser:           "some user",
		MasterURL:         "what a url!",
		DisableAutoUpdate: true,
		WorkerZones:       []string{"zone2", "zone1", "zone0"},
		Vpcs:              []string{"vpcs3", "zone1", "vpcs2"},
		CRN:               "a CRN",
		VersionEOS:        "a version",
		ServiceEndpoints: ibmContainerV2.Endpoints{
			PrivateServiceEndpointEnabled: false,
			PrivateServiceEndpointURL:     "some url",
			PublicServiceEndpointEnabled:  true,
			PublicServiceEndpointURL:      "another url",
		},
		Lifecycle: ibmContainerV2.LifeCycleInfo{
			ModifiedDate:             "2006-01-02 15:04:05",
			MasterStatus:             "a status",
			MasterStatusModifiedDate: "1006-01-02 15:04:05",
			MasterHealth:             "fairly healthy",
			MasterState:              "NY State",
		},
		WorkerCount: 4,
		Ingress: ibmContainerV2.IngresInfo{
			HostName:   "ingress host name",
			SecretName: "ingress secret name",
		},
		Features: ibmContainerV2.Feat{
			KeyProtectEnabled: false,
			PullSecretApplied: true,
		},
	}

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

// Returns a pointer to a random map
func randomMap() *map[string]string {
	result := map[string]string{
		"one":   "two",
		"three": "two",
		"four":  "two",
	}

	return &result
}

// Tests the GenerateClusterCreateRequest function
func TestGenerateClusterCreateRequest(t *testing.T) {
	crossplaneRequest := &v1alpha1.ClusterCreateRequest{
		DisablePublicServiceEndpoint: false,
		KubeVersion:                  "a version",
		Billing:                      reference.ToPtrValue("billing"),
		PodSubnet:                    "a subnet",
		Provider:                     "a provider",
		ServiceSubnet:                "a service net",
		Name:                         "a name",
		DefaultWorkerPoolEntitlement: "an entitlement",
		CosInstanceCRN:               "a crn",
		WorkerPools: v1alpha1.WorkerPoolConfig{
			DiskEncryption: ibmc.BoolPtr(true),
			Entitlement:    "so entitled",
			Flavor:         "banana",
			Isolation:      reference.ToPtrValue("...due to COVID"),
			Labels:         randomMap(),
			Name:           "another name",
			VpcID:          "whoooooa",
			WorkerCount:    33,
			Zones: []v1alpha1.Zone{{ID: reference.ToPtrValue("name 1"), SubnetID: reference.ToPtrValue("verston 2")},
				{ID: reference.ToPtrValue("name 1"), SubnetID: reference.ToPtrValue("v2")}},
		},
	}

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
