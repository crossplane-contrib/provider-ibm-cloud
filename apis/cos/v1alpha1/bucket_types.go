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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
)

// BucketPararams are input params when creating a bucket
type BucketPararams struct {
	// Name of the bucket. Must be globally unique and DNS-compliant; names between 3 and 63 characters long must be made of lowercase letters, numbers, and dashes. Must begin and end with a lowercase letter or number.
	//
	// Names resembling IP addresses are not allowed. Names must be unique because all buckets in the public cloud share a global namespace, allowing access to buckets without the need to provide any service instance or account information.
	//
	// It is not possible to create a bucket with a name beginning with cosv1- or account- as these prefixes are reserved by the system.
	Name string `json:"bucket"`

	// References the resource service instance where the bucket will be created and to which data usage will be billed. This value can be either the full Cloud Resource Name (CRN) or just the GUID segment that identifies the service instance.
	//
	// Note:
	//    Only one of 'IbmServiceInstanceID', 'IbmServiceInstanceIDRef', 'IbmServiceInstanceIDSelector' should be != nil
	//
	// Example: d6f76k03-6k4f-4a82-n165-697654o63903
	// +immutable
	// +optional
	IbmServiceInstanceID *string `json:"ibmServiceInstanceID,omitempty"`

	// A reference to a resource instance containing the bucket
	//
	// Note:
	//    Only one of 'IbmServiceInstanceID', 'IbmServiceInstanceIDRef', 'IbmServiceInstanceIDSelector' should be != nil
	//
	// +immutable
	// +optional
	IbmServiceInstanceIDRef *runtimev1alpha1.Reference `json:"ibmServiceInstanceIDRef,omitempty"`

	// Selects a reference to a resource instance containing the bucket
	//
	// Note:
	//    Only one of 'IbmServiceInstanceID', 'IbmServiceInstanceIDRef', 'IbmServiceInstanceIDSelector' should be != nil
	//
	// +immutable
	// +optional
	IbmServiceInstanceIDSelector *runtimev1alpha1.Selector `json:"ibmServiceInstanceIDSelector,omitempty"`

	// The algorithm and key size used to for the managed encryption root key. Required if IbmSSEKpCustomerRootKeyCrn is also present.
	//
	// Allowable values: ``AES256''
	// +immutable
	// +optional
	IbmSSEKpEncryptionAlgorithm *string `json:"ibmSSEKpEncryptionAlgorithm,omitempty"`

	// The CRN of the root key used to encrypt the bucket. Required ifIbmSSEKpEncryptionAlgorithm is also present.
	//
	// Example: crn:v1:bluemix:public:kms:us-south:a/f047b55a3362ac06afad8a3f2f5586ea:12e8c9c2-a162-472d-b7d6-8b9a86b815a6:key:02fd6835-6001-4482-a892-13bd2085f75d
	// +immutable
	// +optional
	IbmSSEKpCustomerRootKeyCrn *string `json:"ibmSSEKpCustomerRootKeyCrn,omitempty"`

	// Allowable values: ``us-standard'', ``us-cold''
	// +immutable
	LocationConstraint string `json:"locationConstraint"`
}

// BucketObservation contains the fields of a bucket that are "set" by the IBM cloud
type BucketObservation struct {
	// When the bucket was created. Can change when making changes to it -
	// such as editing its policy
	CreationDate *metav1.Time `json:"creationDate"`
}

// BucketSpec - desired end-state of a Bucket on the IBM cloud
type BucketSpec struct {
	runtimev1alpha1.ResourceSpec `json:",inline"`

	// Info the IBM cloud needs to create a bucket
	ForProvider BucketPararams `json:"forProvider"`
}

// BucketStatus - whatever the status is (the IBM cloud decides that)
type BucketStatus struct {
	runtimev1alpha1.ResourceStatus `json:",inline"`

	// Info the IBM cloud returns about a bucket
	AtProvider BucketObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// Bucket contains all the info (spec + status) for a bucket
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="STATUS",type="string",JSONPath=".status.bindingPhase"
// +kubebuilder:printcolumn:name="STATE",type="string",JSONPath=".status.atProvider.state"
// +kubebuilder:printcolumn:name="CLASS",type="string",JSONPath=".spec.classRef.name"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,ibmcloud}
type Bucket struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BucketSpec   `json:"spec"`
	Status BucketStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// BucketList - list of existing buckets...
type BucketList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	// List of buckets returned
	Items []Bucket `json:"buckets"`
}
