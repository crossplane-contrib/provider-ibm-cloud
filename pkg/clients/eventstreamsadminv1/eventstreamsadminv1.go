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

package topic

import (
	arv1 "github.com/IBM/eventstreams-go-sdk/pkg/adminrestv1"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"

	"github.com/crossplane-contrib/provider-ibm-cloud/apis/eventstreamsadminv1/v1alpha1"
	ibmc "github.com/crossplane-contrib/provider-ibm-cloud/pkg/clients"
	"github.com/crossplane/crossplane-runtime/pkg/reference"
)

// LateInitializeSpec fills optional and unassigned fields with the values in *arv1.TopicDetail object.
func LateInitializeSpec(spec *v1alpha1.TopicParameters, in *arv1.TopicDetail) error { // nolint:gocyclo

	if spec.Partitions == nil {
		spec.Partitions = in.Partitions
	}

	return nil
}

// GenerateCreateTopicOptions produces CreateTopicOptions object from TopicParameters object.
func GenerateCreateTopicOptions(in v1alpha1.TopicParameters, o *arv1.CreateTopicOptions) error {
	o.Name = reference.ToPtrValue(in.Name)
	o.Partitions = in.Partitions
	o.PartitionCount = in.PartitionCount
	o.Configs = Generatearv1ConfigCreate(in.Configs)
	return nil
}

// Generatearv1ConfigCreate generates []arv1.ConfigCreate from []v1alpha1.ConfigCreate
func Generatearv1ConfigCreate(in []v1alpha1.ConfigCreate) []arv1.ConfigCreate {
	if in == nil {
		return nil
	}
	o := []arv1.ConfigCreate{}
	for _, m := range in {
		item := arv1.ConfigCreate{
			Name:  reference.ToPtrValue(m.Name),
			Value: reference.ToPtrValue(m.Value),
		}
		o = append(o, item)
	}
	return o
}

// GenerateUpdateTopicOptions produces UpdateTopicOptions object from TopicParameters object.
func GenerateUpdateTopicOptions(in v1alpha1.TopicParameters, o *arv1.UpdateTopicOptions) error {
	o.TopicName = reference.ToPtrValue(in.Name)
	o.NewTotalPartitionCount = in.PartitionCount
	o.Configs = GenerateConfigUpdate(in.Configs)
	return nil
}

// GenerateConfigUpdate generates []arv1.ConfigUpdate from []v1alpha1.ConfigCreate
func GenerateConfigUpdate(in []v1alpha1.ConfigCreate) []arv1.ConfigUpdate {
	if in == nil {
		return nil
	}
	o := []arv1.ConfigUpdate{}
	for _, m := range in {
		item := arv1.ConfigUpdate{
			Name:  reference.ToPtrValue(m.Name),
			Value: reference.ToPtrValue(m.Value),
			//ResetToDefault: needs to be a pointer to a boolean value, should be true or false??
		}
		o = append(o, item)
	}
	return o
}

// GenerateObservation produces TopicObservation object from *arv1.TopicDetail object.
func GenerateObservation(in *arv1.TopicDetail) (v1alpha1.TopicObservation, error) {

	o := v1alpha1.TopicObservation{
		ReplicationFactor:  ibmc.Int64Value(in.ReplicationFactor),
		RetentionMs:        ibmc.Int64Value(in.RetentionMs),
		CleanupPolicy:      reference.FromPtrValue(in.CleanupPolicy),
		Configs:            Generatev1alpha1TopicConfigs(in.Configs),
		ReplicaAssignments: GenerateReplicaAssignments(in.ReplicaAssignments),
	}
	return o, nil
}

// Generatev1alpha1TopicConfigs generates *v1alpha1.TopicConfigs from *arv1.TopicConfigs
func Generatev1alpha1TopicConfigs(in *arv1.TopicConfigs) *v1alpha1.TopicConfigs {
	if in == nil {
		return nil
	}
	o := &v1alpha1.TopicConfigs{
		CleanupPolicy:     reference.FromPtrValue(in.CleanupPolicy),
		MinInsyncReplicas: reference.FromPtrValue(in.MinInsyncReplicas),
		RetentionBytes:    reference.FromPtrValue(in.RetentionBytes),
		RetentionMs:       reference.FromPtrValue(in.RetentionMs),
		SegmentBytes:      reference.FromPtrValue(in.SegmentBytes),
		SegmentIndexBytes: reference.FromPtrValue(in.SegmentIndexBytes),
		SegmentMs:         reference.FromPtrValue(in.SegmentMs),
	}
	return o
}

// GenerateReplicaAssignments generates []v1alpha1.ReplicaAssignment from []arv1.ReplicaAssignment
func GenerateReplicaAssignments(in []arv1.ReplicaAssignment) []v1alpha1.ReplicaAssignment {
	if in == nil {
		return nil
	}
	o := []v1alpha1.ReplicaAssignment{}
	for _, m := range in {
		item := v1alpha1.ReplicaAssignment{
			ID:      ibmc.Int64Value(m.ID),
			Brokers: GenerateBrokers(m.Brokers),
		}
		o = append(o, item)
	}
	return o
}

// GenerateBrokers generates *v1alpha1.ReplicaAssignmentBrokers from *arv1.ReplicaAssignmentBrokers
func GenerateBrokers(in *arv1.ReplicaAssignmentBrokers) *v1alpha1.ReplicaAssignmentBrokers {
	if in == nil {
		return nil
	}
	o := &v1alpha1.ReplicaAssignmentBrokers{
		Replicas: in.Replicas,
	}
	return o
}

// IsUpToDate checks whether current state is up-to-date compared to the given set of parameters.
func IsUpToDate(in *v1alpha1.TopicParameters, observed *arv1.TopicDetail) (bool, error) {
	desired := in.DeepCopy()
	actual, err := GenerateTopicParameters(observed)
	if err != nil {
		return false, err
	}

	// it says logger isn't found even when logger is imported??
	// l.Info(cmp.Diff(desired, actual, cmpopts.IgnoreTypes(&runtimev1alpha1.Reference{}, &runtimev1alpha1.Selector{}, []runtimev1alpha1.Reference{})))

	// what is it being compared to, ie which fields should be ignored??
	return cmp.Equal(desired, actual, cmpopts.EquateEmpty(),
		cmpopts.IgnoreFields(v1alpha1.TopicParameters{}, "KafkaAdminURL", "KafkaAdminURLRef", "KafkaAdminURLSelector"),
		cmpopts.IgnoreTypes(&runtimev1alpha1.Reference{}, &runtimev1alpha1.Selector{}, []runtimev1alpha1.Reference{})), nil
}

// GenerateTopicParameters generates *v1alpha1.TopicParameters from *arv1.TopicDetail
func GenerateTopicParameters(in *arv1.TopicDetail) (*v1alpha1.TopicParameters, error) {

	o := &v1alpha1.TopicParameters{
		Name: reference.FromPtrValue(in.Name),
		// if in.Partitions is nil make it 1??
		Partitions:     in.Partitions,
		PartitionCount: in.Partitions,
		// Configs: Generatev1alpha1ConfigCreate(in.Configs)

	}
	return o, nil
}

// // Generatev1alpha1ConfigCreate generates []v1alpha1.ConfigCreate from *arv1.TopicConfigs
// func Generatev1alpha1ConfigCreate(in *arv1.TopicConfigs) []v1alpha1.ConfigCreate {
// 	if in == nil {
// 		return nil
// 	}

// 	return o
// }
