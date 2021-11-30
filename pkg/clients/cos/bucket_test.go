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

package cos

import (
	"math/rand"
	"reflect"
	"testing"
	"time"

	"github.com/IBM/ibm-cos-sdk-go/service/s3"

	"github.com/crossplane/crossplane-runtime/pkg/reference"

	"github.com/crossplane-contrib/provider-ibm-cloud/apis/cos/v1alpha1"

	"github.com/google/go-cmp/cmp"
)

// Returns randomly a pointer to a string or nil
func randomOrNil() *string {
	var result *string = nil

	if rand.Intn(2) == 0 {
		aStr := "foo"
		result = &aStr
	}

	return result
}

// Tests the GenerateS3BucketInput function
func TestGenerateS3BucketInput(t *testing.T) {
	aRandomStrOrPtr := randomOrNil()
	theInstanceID := "the resource instance id in the cloud"

	crossplaneBucketParams := &v1alpha1.BucketPararams{
		Name:                         "harry",
		IbmServiceInstanceID:         &theInstanceID,
		IbmServiceInstanceIDRef:      nil,
		IbmServiceInstanceIDSelector: nil,
		IbmSSEKpEncryptionAlgorithm:  aRandomStrOrPtr,
		IbmSSEKpCustomerRootKeyCrn:   aRandomStrOrPtr,
		LocationConstraint:           "earth",
	}

	s3BucketParams := &s3.CreateBucketInput{}

	t.Run("TestGenerateS3BucketInput", func(t *testing.T) {
		GenerateS3BucketInput(crossplaneBucketParams, s3BucketParams)

		tests := map[string]struct {
			crossplaneVal *string
			s3Val         *string
		}{
			"bucketName": {
				crossplaneVal: &crossplaneBucketParams.Name,
				s3Val:         s3BucketParams.Bucket,
			},

			"IbmServiceInstanceID": {
				crossplaneVal: crossplaneBucketParams.IbmServiceInstanceID,
				s3Val:         s3BucketParams.IBMServiceInstanceId,
			},

			"IbmSSEKpEncryptionAlgorithm": {
				crossplaneVal: crossplaneBucketParams.IbmSSEKpEncryptionAlgorithm,
				s3Val:         s3BucketParams.IBMSSEKPEncryptionAlgorithm,
			},

			"IbmSSEKpCustomerRootKeyCrn": {
				crossplaneVal: crossplaneBucketParams.IbmSSEKpCustomerRootKeyCrn,
				s3Val:         s3BucketParams.IBMSSEKPCustomerRootKeyCrn,
			},

			"location": {
				crossplaneVal: &crossplaneBucketParams.LocationConstraint,
				s3Val:         s3BucketParams.CreateBucketConfiguration.LocationConstraint,
			},
		}

		for name, tc := range tests {
			t.Run(name, func(t *testing.T) {
				crossplaneValStr := reference.FromPtrValue(tc.crossplaneVal)
				s3ValStr := reference.FromPtrValue(tc.s3Val)

				if diff := cmp.Diff(crossplaneValStr, s3ValStr); diff != "" {
					t.Errorf("TestGenerateS3BucketInput(...): -wanted, +got:\n%s", diff)
				}
			})
		}
	})
}

func TestGenerateBucketObservation(t *testing.T) {
	tests := []time.Time{time.Date(2009, 11, 17, 20, 34, 58, 651387237, time.UTC)}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			s3Bucket := &s3.Bucket{CreationDate: &tt}
			obs, err := GenerateBucketObservation(s3Bucket)

			if err != nil {
				t.Errorf("GenerateObservation() returned an error: %s", err)
			} else if !reflect.DeepEqual(obs.CreationDate.Time, *s3Bucket.CreationDate) {
				t.Errorf("GenerateObservation() = %v, want %v", obs.CreationDate, s3Bucket.CreationDate)
			}
		})
	}
}
