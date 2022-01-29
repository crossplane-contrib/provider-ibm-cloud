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

	"github.com/crossplane-contrib/provider-ibm-cloud/apis/container/containerv2/v1alpha1"
	ibmc "github.com/crossplane-contrib/provider-ibm-cloud/pkg/clients"

	"github.com/crossplane/crossplane-runtime/pkg/reference"
)

const (
	clusterName   = "aClusterName" // cannot contain spaces
	anEntitlement = "kids these days"
	aCRN          = "fake crn"
	zoneName1     = "zone 1"
	zoneName2     = "zone 2"
	zoneName3     = "zone 3"
)

// Returns a pointer to a random map
func randomMap() *map[string]string {
	result := map[string]string{
		"one":   "two",
		"three": "two",
		"four":  "two",
	}

	return &result
}

// GetClusterCreateCrossplaneRequest returns a crossplane request to generate a cluster
func GetClusterCreateCrossplaneRequest() *v1alpha1.ClusterCreateRequest {
	return &v1alpha1.ClusterCreateRequest{
		DisablePublicServiceEndpoint: false,
		KubeVersion:                  "a version",
		Billing:                      reference.ToPtrValue("billing"),
		PodSubnet:                    "a subnet",
		Provider:                     "a provider",
		ServiceSubnet:                "a service net",
		Name:                         clusterName,
		DefaultWorkerPoolEntitlement: anEntitlement,
		CosInstanceCRN:               aCRN,
		WorkerPools: v1alpha1.WorkerPoolConfig{
			DiskEncryption: ibmc.BoolPtr(true),
			Entitlement:    "so entitled",
			Flavor:         "banana",
			Isolation:      reference.ToPtrValue("...due to COVID"),
			Labels:         randomMap(),
			Name:           "another name",
			VpcID:          "whoooooa",
			WorkerCount:    33,
			Zones: []v1alpha1.Zone{{ID: reference.ToPtrValue(zoneName1), SubnetID: reference.ToPtrValue("subnet uno")},
				{ID: reference.ToPtrValue(zoneName2), SubnetID: reference.ToPtrValue("subnet due")},
				{ID: reference.ToPtrValue(zoneName2), SubnetID: reference.ToPtrValue("subnet tre")}},
		},
	}
}

// GetContainerClusterInfo returns an object like the one returned by the cloud
func GetContainerClusterInfo() *ibmContainerV2.ClusterInfo {
	return &ibmContainerV2.ClusterInfo{
		CreatedDate:       "2006-01-02 15:04:05",
		DataCenter:        "a data center",
		ID:                "an id",
		Location:          anEntitlement,
		Entitlement:       "kids these days...",
		MasterKubeVersion: "grand master",
		Name:              clusterName,
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
		WorkerZones:       []string{zoneName2, zoneName1, zoneName3}, // different order than in the "create" request
		Vpcs:              []string{"vpcs3", "zone1", "vpcs2"},
		CRN:               aCRN,
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
}
