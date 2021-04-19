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
	"encoding/json"
	"net/http"

	"github.com/crossplane/crossplane-runtime/pkg/reference"

	gcat "github.com/IBM/platform-services-go-sdk/globalcatalogv1"
	gtagv1 "github.com/IBM/platform-services-go-sdk/globaltaggingv1"
	rmgrv2 "github.com/IBM/platform-services-go-sdk/resourcemanagerv2"
)

var (
	resourceGroupNameMockVal = "default"
	resourceGroupIDMockVal   = "0be5ad401ae913d8ff665d92680664ed"
)

// TagsTestHandler handler to mock client SDK call to global tags API
var TagsTestHandler = func(w http.ResponseWriter, r *http.Request) {
	_ = r.Body.Close()
	w.Header().Set("Content-Type", "application/json")
	tags := gtagv1.TagList{
		Items: []gtagv1.Tag{
			{
				Name: reference.ToPtrValue("testString"),
			},
		},
	}
	_ = json.NewEncoder(w).Encode(tags)
}

// RgTestHandler handler to mock client SDK call to resource manager API
var RgTestHandler = func(w http.ResponseWriter, r *http.Request) {
	_ = r.Body.Close()
	w.Header().Set("Content-Type", "application/json")
	rgl := rmgrv2.ResourceGroupList{
		Resources: []rmgrv2.ResourceGroup{
			{
				ID:   reference.ToPtrValue(resourceGroupIDMockVal),
				Name: reference.ToPtrValue(resourceGroupNameMockVal),
			},
		},
	}
	_ = json.NewEncoder(w).Encode(rgl)
}

// PcatTestHandler handler to mock client SDK call to global catalog API for plans
var PcatTestHandler = func(planName string, planId string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		_ = r.Body.Close()
		w.Header().Set("Content-Type", "application/json")
		planEntry := gcat.EntrySearchResult{
			Resources: []gcat.CatalogEntry{
				{
					Name: reference.ToPtrValue(planName),
					ID:   reference.ToPtrValue(planId),
				},
			},
		}
		_ = json.NewEncoder(w).Encode(planEntry)
	}
}

// SvcatTestHandler handler to mock client SDK call to global catalog API for services
var SvcatTestHandler = func(serviceName string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		_ = r.Body.Close()
		w.Header().Set("Content-Type", "application/json")
		catEntry := gcat.EntrySearchResult{
			Resources: []gcat.CatalogEntry{
				{
					Metadata: &gcat.CatalogEntryMetadata{
						UI: &gcat.UIMetaData{
							PrimaryOfferingID: reference.ToPtrValue(serviceName),
						},
					},
				},
			},
		}
		_ = json.NewEncoder(w).Encode(catEntry)
	}
}
