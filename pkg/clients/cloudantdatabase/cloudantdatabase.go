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
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/reference"

	cv1 "github.com/IBM/cloudant-go-sdk/cloudantv1"

	"github.com/crossplane-contrib/provider-ibm-cloud/apis/cloudantv1/v1alpha1"
	ibmc "github.com/crossplane-contrib/provider-ibm-cloud/pkg/clients"
)

// LateInitializeSpec fills optional and unassigned fields with the values in *cv1.Databaseinformation object.
func LateInitializeSpec(spec *v1alpha1.CloudantDatabaseParameters, in *cv1.DatabaseInformation) error { // nolint:gocyclo

	if spec.Partitioned == nil {
		spec.Partitioned = in.Props.Partitioned
	}

	if spec.Q == nil {
		spec.Q = in.Cluster.Q
	}

	return nil
}

// GenerateCreateCloudantDatabaseOptions produces PutDatabaseOptions object from CloudantDatabaseParameters object.
func GenerateCreateCloudantDatabaseOptions(in v1alpha1.CloudantDatabaseParameters, o *cv1.PutDatabaseOptions) error {
	o.Db = reference.ToPtrValue(in.Db)
	o.Partitioned = in.Partitioned
	o.Q = in.Q
	return nil
}

// GenerateObservation produces CloudantDatabaseObservation object from *cv1.DatabaseInformation object.
func GenerateObservation(in *cv1.DatabaseInformation) (v1alpha1.CloudantDatabaseObservation, error) {
	o := v1alpha1.CloudantDatabaseObservation{
		Cluster:            Generatev1alpha1DatabaseInformationCluster(in.Cluster),
		CommittedUpdateSeq: reference.FromPtrValue(in.CommittedUpdateSeq),
		CompactRunning:     ibmc.BoolValue(in.CompactRunning),
		CompactedSeq:       reference.FromPtrValue(in.CompactedSeq),
		DiskFormatVersion:  ibmc.Int64Value(in.DiskFormatVersion),
		DocCount:           ibmc.Int64Value(in.DocCount),
		DocDelCount:        ibmc.Int64Value(in.DocDelCount),
		Engine:             reference.FromPtrValue(in.Engine),
		Sizes:              Generatev1alpha1ContentInformationSizes(in.Sizes),
		UpdateSeq:          reference.FromPtrValue(in.UpdateSeq),
		UUID:               reference.FromPtrValue(in.UUID),
	}
	return o, nil
}

// Generatev1alpha1DatabaseInformationCluster generates *v1alpha1.DatabaseinformationCluster from *cv1.DatabaseInformationCluster
func Generatev1alpha1DatabaseInformationCluster(in *cv1.DatabaseInformationCluster) *v1alpha1.DatabaseInformationCluster {
	if in == nil {
		return nil
	}
	o := &v1alpha1.DatabaseInformationCluster{
		N: ibmc.Int64Value(in.N),
		R: ibmc.Int64Value(in.R),
		W: ibmc.Int64Value(in.W),
	}
	return o
}

// Generatev1alpha1ContentInformationSizes generates *v1alpha1.ContentInformationSizes from *cv1.ContentInformationSizes
func Generatev1alpha1ContentInformationSizes(in *cv1.ContentInformationSizes) *v1alpha1.ContentInformationSizes {
	if in == nil {
		return nil
	}
	o := &v1alpha1.ContentInformationSizes{
		Active:   ibmc.Int64Value(in.Active),
		External: ibmc.Int64Value(in.External),
		File:     ibmc.Int64Value(in.File),
	}
	return o
}

// IsUpToDate checks whether current state is up-to-date compared to the given set of parameters.
func IsUpToDate(in *v1alpha1.CloudantDatabaseParameters, observed *cv1.DatabaseInformation, l logging.Logger) (bool, error) {
	desired := in.DeepCopy()
	actual, err := GenerateCloudantDatabaseParameters(observed)
	if err != nil {
		return false, err
	}

	diff := (cmp.Diff(desired, actual,
		cmpopts.EquateEmpty(),
		cmpopts.IgnoreFields(v1alpha1.CloudantDatabaseParameters{}, "CloudantAdminURL", "CloudantAdminURLRef", "CloudantAdminURLSelector"), cmpopts.IgnoreTypes(&runtimev1alpha1.Reference{}, &runtimev1alpha1.Selector{}, []runtimev1alpha1.Reference{})))

	if diff != "" {
		l.Info("IsUpToDate", "Diff", diff)
		return false, nil
	}

	return true, nil
}

// GenerateCloudantDatabaseParameters generates *v1alpha1.CloudantDatabaseParameters from *cv1.DatabaseInformation
func GenerateCloudantDatabaseParameters(in *cv1.DatabaseInformation) (*v1alpha1.CloudantDatabaseParameters, error) {

	o := &v1alpha1.CloudantDatabaseParameters{
		Db:          reference.FromPtrValue(in.DbName),
		Partitioned: in.Props.Partitioned,
		Q:           in.Cluster.Q,
	}
	return o, nil
}
