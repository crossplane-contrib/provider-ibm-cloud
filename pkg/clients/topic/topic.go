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
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/reference"

	arv1 "github.com/IBM/eventstreams-go-sdk/pkg/adminrestv1"

	"github.com/crossplane-contrib/provider-ibm-cloud/apis/eventstreamsadminv1/v1alpha1"
	ibmc "github.com/crossplane-contrib/provider-ibm-cloud/pkg/clients"
)

// LateInitializeSpec fills optional and unassigned fields with the values in *arv1.TopicDetail object.
func LateInitializeSpec(spec *v1alpha1.TopicParameters, in *arv1.TopicDetail) error { // nolint:gocyclo

	if spec.PartitionCount != nil {
		spec.Partitions = spec.PartitionCount
	}

	if spec.Partitions == nil {
		spec.Partitions = in.Partitions
	}

	if spec.PartitionCount == nil {
		spec.PartitionCount = in.Partitions
	}

	// this works but it still updates configs everytime ?? could potentially optimize but would be difficult and probably increase cyclomatic complexity ??
	spec.Configs = Generatev1alpha1ConfigCreateLateInitializeSpec(spec.Configs, in.Configs)

	return nil
}

// Generatev1alpha1ConfigCreateLateInitializeSpec generates []v1alpha1.ConfigCreate from *arv1.TopicConfigs, only adding configs that aren't already in specConfigs
func Generatev1alpha1ConfigCreateLateInitializeSpec(specConfigs []v1alpha1.ConfigCreate, in *arv1.TopicConfigs) []v1alpha1.ConfigCreate { // nolint:gocyclo
	if in == nil {
		return nil
	}
	o := []v1alpha1.ConfigCreate{}
	m := map[string]string{}
	for _, item := range specConfigs {
		m[item.Name] = item.Value
	}

	if m["cleanup.policy"] != "" {
		c := v1alpha1.ConfigCreate{
			Name:  "cleanup.policy",
			Value: m["cleanup.policy"],
		}
		o = append(o, c)
	} else if in.CleanupPolicy != nil {
		c := v1alpha1.ConfigCreate{
			Name:  "cleanup.policy",
			Value: reference.FromPtrValue(in.CleanupPolicy),
		}
		o = append(o, c)
	}
	if m["retention.bytes"] != "" {
		c := v1alpha1.ConfigCreate{
			Name:  "retention.bytes",
			Value: m["retention.bytes"],
		}
		o = append(o, c)
	} else if in.RetentionBytes != nil {
		c := v1alpha1.ConfigCreate{
			Name:  "retention.bytes",
			Value: reference.FromPtrValue(in.RetentionBytes),
		}
		o = append(o, c)
	}
	if m["retention.ms"] != "" {
		c := v1alpha1.ConfigCreate{
			Name:  "retention.ms",
			Value: m["retention.ms"],
		}
		o = append(o, c)
	} else if in.RetentionBytes != nil {
		c := v1alpha1.ConfigCreate{
			Name:  "retention.ms",
			Value: reference.FromPtrValue(in.RetentionMs),
		}
		o = append(o, c)
	}
	if m["segment.bytes"] != "" {
		c := v1alpha1.ConfigCreate{
			Name:  "segment.bytes",
			Value: m["segment.bytes"],
		}
		o = append(o, c)
	} else if in.SegmentBytes != nil {
		c := v1alpha1.ConfigCreate{
			Name:  "segment.bytes",
			Value: reference.FromPtrValue(in.SegmentBytes),
		}
		o = append(o, c)
	}
	if m["segment.index.bytes"] != "" {
		c := v1alpha1.ConfigCreate{
			Name:  "segment.index.bytes",
			Value: m["segment.index.bytes"],
		}
		o = append(o, c)
	} else if in.SegmentIndexBytes != nil {
		c := v1alpha1.ConfigCreate{
			Name:  "segment.index.bytes",
			Value: reference.FromPtrValue(in.SegmentIndexBytes),
		}
		o = append(o, c)
	}
	if m["segment.ms"] != "" {
		c := v1alpha1.ConfigCreate{
			Name:  "segment.ms",
			Value: m["segment.ms"],
		}
		o = append(o, c)
	} else if in.SegmentMs != nil {
		c := v1alpha1.ConfigCreate{
			Name:  "segment.ms",
			Value: reference.FromPtrValue(in.SegmentMs),
		}
		o = append(o, c)
	}
	return o
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
func GenerateUpdateTopicOptions(currentPartitions *int64, in v1alpha1.TopicParameters, o *arv1.UpdateTopicOptions) error {
	o.TopicName = reference.ToPtrValue(in.Name)
	if ibmc.Int64Value(in.PartitionCount) > ibmc.Int64Value(currentPartitions) {
		o.NewTotalPartitionCount = in.PartitionCount
	}
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
			Name:           reference.ToPtrValue(m.Name),
			Value:          reference.ToPtrValue(m.Value),
			ResetToDefault: ibmc.BoolPtr(false),
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
func IsUpToDate(in *v1alpha1.TopicParameters, observed *arv1.TopicDetail, l logging.Logger) (bool, error) {
	desired := in.DeepCopy()
	actual, err := GenerateTopicParameters(observed)
	if err != nil {
		return false, err
	}
	diff := (cmp.Diff(desired, actual,
		cmpopts.EquateEmpty(),
		cmpopts.IgnoreFields(v1alpha1.TopicParameters{}, "KafkaAdminURL", "KafkaAdminURLRef", "KafkaAdminURLSelector"), cmpopts.IgnoreTypes(&runtimev1alpha1.Reference{}, &runtimev1alpha1.Selector{}, []runtimev1alpha1.Reference{})))

	if diff != "" {
		l.Info("IsUpToDate", "Diff", diff)
		return false, nil
	}

	return true, nil
}

// GenerateTopicParameters generates *v1alpha1.TopicParameters from *arv1.TopicDetail
func GenerateTopicParameters(in *arv1.TopicDetail) (*v1alpha1.TopicParameters, error) {

	o := &v1alpha1.TopicParameters{
		Name:           reference.FromPtrValue(in.Name),
		Partitions:     in.Partitions,
		PartitionCount: in.Partitions,
		Configs:        Generatev1alpha1ConfigCreate(in.Configs),
	}
	return o, nil
}

// Generatev1alpha1ConfigCreate generates []v1alpha1.ConfigCreate from *arv1.TopicConfigs
func Generatev1alpha1ConfigCreate(in *arv1.TopicConfigs) []v1alpha1.ConfigCreate {
	if in == nil {
		return nil
	}
	o := []v1alpha1.ConfigCreate{}
	if in.CleanupPolicy != nil {
		c := v1alpha1.ConfigCreate{
			Name:  "cleanup.policy",
			Value: reference.FromPtrValue(in.CleanupPolicy),
		}
		o = append(o, c)
	}
	if in.RetentionBytes != nil {
		c := v1alpha1.ConfigCreate{
			Name:  "retention.bytes",
			Value: reference.FromPtrValue(in.RetentionBytes),
		}
		o = append(o, c)
	}
	if in.RetentionMs != nil {
		c := v1alpha1.ConfigCreate{
			Name:  "retention.ms",
			Value: reference.FromPtrValue(in.RetentionMs),
		}
		o = append(o, c)
	}
	if in.SegmentBytes != nil {
		c := v1alpha1.ConfigCreate{
			Name:  "segment.bytes",
			Value: reference.FromPtrValue(in.SegmentBytes),
		}
		o = append(o, c)
	}
	if in.SegmentIndexBytes != nil {
		c := v1alpha1.ConfigCreate{
			Name:  "segment.index.bytes",
			Value: reference.FromPtrValue(in.SegmentIndexBytes),
		}
		o = append(o, c)
	}
	if in.SegmentMs != nil {
		c := v1alpha1.ConfigCreate{
			Name:  "segment.ms",
			Value: reference.FromPtrValue(in.SegmentMs),
		}
		o = append(o, c)
	}
	return o
}
