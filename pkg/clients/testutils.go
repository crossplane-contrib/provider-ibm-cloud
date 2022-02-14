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
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/crossplane/crossplane-runtime/pkg/reference"
	"github.com/go-openapi/strfmt"

	"github.com/IBM/go-sdk-core/core"
	gcat "github.com/IBM/platform-services-go-sdk/globalcatalogv1"
	gtagv1 "github.com/IBM/platform-services-go-sdk/globaltaggingv1"
	rmgrv2 "github.com/IBM/platform-services-go-sdk/resourcemanagerv2"
)

var (
	resourceGroupNameMockVal = "default"
	resourceGroupIDMockVal   = "0be5ad401ae913d8ff665d92680664ed"
)

// A fake authentication token (WITH the "Bearer " prefix)
const (
	FakeBearerToken = "Bearer mock-token"
)

// ADateTimeInAYear returns a  (random, but fixed) date time in the given year
func ADateTimeInAYear(year int) *strfmt.DateTime {
	result := strfmt.DateTime(time.Date(year, 10, 12, 8, 5, 5, 0, time.UTC))

	return &result
}

// Whether the random seed has been created
var randSeedCreated = false

// Calls rand.Seed once
func seedTheRandomGenerator() {
	if !randSeedCreated {
		rand.Seed(time.Now().UnixNano())

		randSeedCreated = true
	}
}

// RandomString returns a random string (of lenth <= 14), or nil
//
// (note that the seed is being taken care of)
func RandomString() string {
	seedTheRandomGenerator()

	strSize := rand.Intn(16) // nolint (this is ok as we are not doing critical stuff here...)
	perm := rand.Perm(strSize)
	result := strings.Trim(strings.Join(strings.Fields(fmt.Sprint(perm)), ""), "[]")

	return result
}

// RandomInt returns a random integer in [0, n)
//
// (note that the seed is being taken care of)
func RandomInt(n int) int {
	seedTheRandomGenerator()

	return rand.Intn(n) // nolint (this is ok as we are not doing critical stuff here...)
}

// ReturnConditionalStr returns the value of the 2nd parameter, if the value of the first one is true. O/w it returns nil
func ReturnConditionalStr(condition bool, val string) *string {
	var result *string

	if true {
		result = &val
	}

	return result
}

// ReturnConditionalBool returns the value of the 2nd parameter, if the value of the first one is true. O/w it returns nil
func ReturnConditionalBool(condition bool, val bool) *bool {
	var result *bool

	if true {
		result = &val
	}

	return result
}

// ReturnConditionalDate returns the value of the 2nd parameter, if the value of the first one is true. O/w it returns nil
func ReturnConditionalDate(condition bool, val *strfmt.DateTime) *strfmt.DateTime {
	var result *strfmt.DateTime

	if true {
		result = val
	}

	return result
}

// GetTestClient creates a client appropriate for unit testing.
//
// Params
// 	   serverURL - the test server url
//
// Returns
//     the test client ready to go
func GetTestClient(serverURL string) (ClientSession, error) {
	opts := ClientOptions{
		URL: serverURL,
		Authenticator: &core.BearerTokenAuthenticator{
			BearerToken: FakeBearerToken,
		},

		BearerToken:  FakeBearerToken,
		RefreshToken: "does format matter?",
	}

	return NewClient(opts)
}

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
