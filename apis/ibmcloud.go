/*
Copyright 2020 The Crossplane Authors.

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

// Package apis contains Kubernetes API for IMB Cloud provider.
package apis

import (
	"k8s.io/apimachinery/pkg/runtime"

	cv1 "github.com/crossplane-contrib/provider-ibm-cloud/apis/cloudantv1/v1alpha1"
	contv2 "github.com/crossplane-contrib/provider-ibm-cloud/apis/container/containerv2/v1alpha1"
	cos "github.com/crossplane-contrib/provider-ibm-cloud/apis/cos/v1alpha1"
	esav1 "github.com/crossplane-contrib/provider-ibm-cloud/apis/eventstreamsadminv1/v1alpha1"
	iamagv2 "github.com/crossplane-contrib/provider-ibm-cloud/apis/iamaccessgroupsv2/v1alpha1"
	iampmv1 "github.com/crossplane-contrib/provider-ibm-cloud/apis/iampolicymanagementv1/v1alpha1"
	icdv5 "github.com/crossplane-contrib/provider-ibm-cloud/apis/ibmclouddatabasesv5/v1alpha1"
	rcv2 "github.com/crossplane-contrib/provider-ibm-cloud/apis/resourcecontrollerv2/v1alpha1"
	"github.com/crossplane-contrib/provider-ibm-cloud/apis/v1beta1"
)

func init() {
	// Register the types with the Scheme so the components can map objects to GroupVersionKinds and back
	AddToSchemes = append(AddToSchemes,
		v1beta1.SchemeBuilder.AddToScheme,
		rcv2.SchemeBuilder.AddToScheme,
		icdv5.SchemeBuilder.AddToScheme,
		iampmv1.SchemeBuilder.AddToScheme,
		iamagv2.SchemeBuilder.AddToScheme,
		esav1.SchemeBuilder.AddToScheme,
		cv1.SchemeBuilder.AddToScheme,
		cos.SchemeBuilder.AddToScheme,
		contv2.SchemeBuilder.AddToScheme,
	)
}

// AddToSchemes may be used to add all resources defined in the project to a Scheme
var AddToSchemes runtime.SchemeBuilder

// AddToScheme adds all Resources to the Scheme
func AddToScheme(s *runtime.Scheme) error {
	return AddToSchemes.AddToScheme(s)
}
