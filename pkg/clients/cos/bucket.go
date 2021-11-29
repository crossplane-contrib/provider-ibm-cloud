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
	"github.com/crossplane/crossplane-runtime/pkg/reference"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/IBM/ibm-cos-sdk-go/service/s3"

	"github.com/crossplane-contrib/provider-ibm-cloud/apis/cos/v1alpha1"
)

// GenerateBucketObservation sets the reported creation date to the IBM-cloud one
func GenerateBucketObservation(in *s3.Bucket) (v1alpha1.BucketObservation, error) {
	var newTimePtr *metav1.Time
	if in.CreationDate != nil {
		newTime := (metav1.NewTime(*in.CreationDate))
		newTimePtr = &newTime
	}

	result := v1alpha1.BucketObservation{
		CreationDate: newTimePtr,
	}

	return result, nil
}

// GenerateS3BucketInput populates the `out' object based on the values in the "in" object
func GenerateS3BucketInput(in *v1alpha1.BucketPararams, out *s3.CreateBucketInput) error {
	out.SetBucket(in.Name)
	out.SetIBMServiceInstanceId(*in.IbmServiceInstanceID)

	if in.IbmSSEKpEncryptionAlgorithm != nil {
		out.SetIBMSSEKPEncryptionAlgorithm(reference.FromPtrValue(in.IbmSSEKpEncryptionAlgorithm))
	}

	if in.IbmSSEKpCustomerRootKeyCrn != nil {
		out.SetIBMSSEKPCustomerRootKeyCrn(reference.FromPtrValue(in.IbmSSEKpCustomerRootKeyCrn))
	}

	bucketConf := &s3.CreateBucketConfiguration{}
	bucketConf.SetLocationConstraint(in.LocationConstraint)
	out.SetCreateBucketConfiguration(bucketConf)

	return nil
}
