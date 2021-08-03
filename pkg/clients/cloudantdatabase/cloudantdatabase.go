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

package cloudantdatabase

import (
	cv1 "github.com/IBM/cloudant-go-sdk/cloudantv1"

	"github.com/crossplane-contrib/provider-ibm-cloud/apis/cloudantv1/v1alpha1"
)

// LateInitializeSpec fills optional and unassigned fields with the values in *cv1.PutDatabaseOptions object.
func LateInitializeSpec(spec *v1alpha1.CloudantDatabaseParameters, in *cv1.PutDatabaseOptions) error { // nolint:gocyclo

	if spec.Partitioned == nil {
		spec.Partitioned = in.Partitioned
	}

	if spec.Q == nil {
		spec.Q = in.Q
	}

	return nil
}
