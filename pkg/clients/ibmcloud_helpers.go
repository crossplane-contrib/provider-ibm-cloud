/*
Copyright 2019 The Crossplane Authors.

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

package clients

import (
	"github.com/pkg/errors"

	"github.com/IBM-Cloud/bluemix-go/crn"
	gcat "github.com/IBM/platform-services-go-sdk/globalcatalogv1"
	gtagv1 "github.com/IBM/platform-services-go-sdk/globaltaggingv1"
	rcv2 "github.com/IBM/platform-services-go-sdk/resourcecontrollerv2"
	rmgrv2 "github.com/IBM/platform-services-go-sdk/resourcemanagerv2"
)

const (
	errServiceNotFound       = "service not found in catalog"
	errListServiceCatEntries = "error listing service entries from catalog"
	errListPlanCatEntries    = "error listing plan entries from catalog"
	errPlanIDNotFound        = "could not find plan ID for plan name"
	errPlanNameNotFound      = "could not find plan name for plan id"
	errListRG                = "could not list resource groups"
	errRGIDNotFound          = "could not find resource group id"
	errRGNameNotFound        = "could not find resource group name"
	errGetTags               = "could not get tags"
)

// GetResourcePlanID gets a resource plan ID from a service name and resource plan name for a given service
func GetResourcePlanID(client ClientSession, serviceName, planName string) (*string, error) {
	planEntry, err := getPlanEntries(client, serviceName)
	if err != nil {
		return nil, errors.Wrap(err, errListPlanCatEntries)
	}

	for _, p := range planEntry.Resources {
		if *p.Name == planName {
			return p.ID, nil
		}
	}

	return nil, errors.New(errPlanIDNotFound)
}

// GetResourcePlanName gets a resource plan ID from a service name and resource plan name for a given service
func GetResourcePlanName(client ClientSession, serviceName, planID string) (*string, error) {
	planEntry, err := getPlanEntries(client, serviceName)
	if err != nil {
		return nil, errors.Wrap(err, errListPlanCatEntries)
	}

	for _, p := range planEntry.Resources {
		if *p.ID == planID {
			return p.Name, nil
		}
	}
	return nil, errors.New(errPlanNameNotFound)
}

func getPlanEntries(client ClientSession, serviceName string) (*gcat.EntrySearchResult, error) {
	listCEOpts := &gcat.ListCatalogEntriesOptions{
		Q:       StringPtr(serviceName),
		Include: StringPtr("*"),
	}

	svcEntries, _, err := client.GlobalCatalogV1().ListCatalogEntries(listCEOpts)
	if err != nil {
		return nil, errors.Wrap(err, errListServiceCatEntries)
	}

	if len(svcEntries.Resources) == 0 {
		return nil, errors.New(errServiceNotFound)
	}

	id := svcEntries.Resources[0].Metadata.Ui.PrimaryOfferingID

	getChildOptions := &gcat.GetChildObjectsOptions{
		ID:   id,
		Kind: StringPtr("*"),
	}
	planEntry, _, err := client.GlobalCatalogV1().GetChildObjects(getChildOptions)

	return planEntry, err
}

// GetResourceGroupID gets a resource group ID from a resource group name
func GetResourceGroupID(client ClientSession, rgName string) (*string, error) {
	opts := &rmgrv2.ListResourceGroupsOptions{}

	entries, _, err := client.ResourceManagerV2().ListResourceGroups(opts)
	if err != nil {
		return nil, errors.Wrap(err, errListRG)
	}

	for _, rg := range entries.Resources {
		if *rg.Name == rgName {
			return rg.ID, nil
		}
	}

	return nil, errors.New(errRGIDNotFound)
}

// GetResourceGroupName gets a resource group name from a resource group ID
func GetResourceGroupName(client ClientSession, rgID string) (string, error) {
	opts := &rmgrv2.ListResourceGroupsOptions{}

	entries, _, err := client.ResourceManagerV2().ListResourceGroups(opts)
	if err != nil {
		return "", errors.Wrap(err, errListRG)
	}

	for _, rg := range entries.Resources {
		if *rg.ID == rgID {
			return StringValue(rg.Name), nil
		}
	}

	return "", errors.New(errRGNameNotFound)
}

// GetResourceInstanceTags gets tags for a resource instance
func GetResourceInstanceTags(client ClientSession, crn string) ([]string, error) {
	listTagsOpts := &gtagv1.ListTagsOptions{
		AttachedTo: &crn,
	}
	entries, _, err := client.GlobalTaggingV1().ListTags(listTagsOpts)
	if err != nil {
		return nil, errors.Wrap(err, errGetTags)
	}

	if len(entries.Items) == 0 {
		return nil, nil
	}

	tags := []string{}
	for _, tag := range entries.Items {
		tags = append(tags, *tag.Name)
	}

	return tags, nil
}

// UpdateResourceInstanceTags update tags for the instance as needed
func UpdateResourceInstanceTags(client ClientSession, crn string, tags []string) error {
	actualTags, err := GetResourceInstanceTags(client, crn)
	if err != nil {
		return err
	}
	toAttach, toDetach := TagsDiff(tags, actualTags)

	if len(toAttach) > 0 {
		attachTagsOpts := &gtagv1.AttachTagOptions{
			TagNames:  toAttach,
			Resources: []gtagv1.Resource{{ResourceID: &crn}},
		}
		_, _, err = client.GlobalTaggingV1().AttachTag(attachTagsOpts)
		if err != nil {
			return err
		}
	}

	if len(toDetach) > 0 {
		detachTagsOpts := &gtagv1.DetachTagOptions{
			TagNames:  toDetach,
			Resources: []gtagv1.Resource{{ResourceID: &crn}},
		}
		_, _, err = client.GlobalTaggingV1().DetachTag(detachTagsOpts)
		if err != nil {
			return err
		}
	}

	return nil
}

// GetServiceName gets ServiceName from Crn
func GetServiceName(in *rcv2.ResourceInstance) string {
	if in.Crn == nil {
		return ""
	}
	crn, err := crn.Parse(*in.Crn)
	if err != nil {
		return ""
	}
	return crn.ServiceName
}

// FindResourceInstancesByName finds resources instances matching name
func FindResourceInstancesByName(client ClientSession, name string) (*rcv2.ResourceInstancesList, error) {
	queryOpts := &rcv2.ListResourceInstancesOptions{
		Name: &name,
	}
	return QueryResourceInstances(client, queryOpts)
}

// QueryResourceInstances finds resource instances based on query options
func QueryResourceInstances(client ClientSession, queryOpts *rcv2.ListResourceInstancesOptions) (*rcv2.ResourceInstancesList, error) {
	list, _, err := client.ResourceControllerV2().ListResourceInstances(queryOpts)
	return list, err
}
