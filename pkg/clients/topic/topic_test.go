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
	"testing"

	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/google/go-cmp/cmp"

	"github.com/crossplane/crossplane-runtime/pkg/reference"

	arv1 "github.com/IBM/eventstreams-go-sdk/pkg/adminrestv1"

	"github.com/crossplane-contrib/provider-ibm-cloud/apis/eventstreamsadminv1/v1alpha1"
	ibmc "github.com/crossplane-contrib/provider-ibm-cloud/pkg/clients"
)

func params(m ...func(*v1alpha1.TopicParameters)) *v1alpha1.TopicParameters {
	p := &v1alpha1.TopicParameters{
		Name:                  "myTopic",
		KafkaAdminURL:         reference.ToPtrValue("myKafkaAdminURL"),
		KafkaAdminURLRef:      &runtimev1alpha1.Reference{},
		KafkaAdminURLSelector: &runtimev1alpha1.Selector{},
		Partitions:            ibmc.Int64Ptr(int64(2)),
		PartitionCount:        ibmc.Int64Ptr(int64(2)),
		// can test empty ConfigCreate or generate ConfigCreate to test
		// Configs:               []v1alpha1.ConfigCreate{},
		Configs: generateTestv1alpha1ConfigCreate(),
	}

	for _, f := range m {
		f(p)
	}
	return p
}

func generateTestv1alpha1ConfigCreate() []v1alpha1.ConfigCreate {
	o := []v1alpha1.ConfigCreate{}

	c := v1alpha1.ConfigCreate{
		Name:  "cleanup.policy",
		Value: "myCleanupPolicy",
	}
	o = append(o, c)
	c = v1alpha1.ConfigCreate{
		Name:  "retention.bytes",
		Value: "myRetentionBytes",
	}
	o = append(o, c)
	c = v1alpha1.ConfigCreate{
		Name:  "retention.ms",
		Value: "myRetentionMs",
	}
	o = append(o, c)
	c = v1alpha1.ConfigCreate{
		Name:  "segment.bytes",
		Value: "mySegmentBytes",
	}
	o = append(o, c)
	c = v1alpha1.ConfigCreate{
		Name:  "segment.index.bytes",
		Value: "mySegmentIndexBytes",
	}
	o = append(o, c)
	c = v1alpha1.ConfigCreate{
		Name:  "segment.ms",
		Value: "mySegmentMs",
	}
	o = append(o, c)
	return o
}

func observation(m ...func(*v1alpha1.TopicObservation)) *v1alpha1.TopicObservation {
	o := &v1alpha1.TopicObservation{
		ReplicationFactor: int64(2),
		RetentionMs:       int64(2),
		CleanupPolicy:     "myCleanupPolicy",
		// Configs:           &v1alpha1.TopicConfigs{},
		Configs: generateTestv1alpha1TopicConfigs(),
		// ReplicaAssignments: []v1alpha1.ReplicaAssignment{},
		ReplicaAssignments: generateTestv1alpha1ReplicaAssignments(),
	}

	for _, f := range m {
		f(o)
	}
	return o
}

func generateTestv1alpha1TopicConfigs() *v1alpha1.TopicConfigs {
	o := &v1alpha1.TopicConfigs{
		CleanupPolicy:     "myCleanupPolicy",
		MinInsyncReplicas: "myMinInsyncReplicas",
		RetentionBytes:    "myRetentionBytes",
		RetentionMs:       "myRetentionMs",
		SegmentBytes:      "mySegmentBytes",
		SegmentIndexBytes: "mySegmentIndexBytes",
		SegmentMs:         "mySegmentMs",
	}
	return o
}

func generateTestv1alpha1ReplicaAssignments() []v1alpha1.ReplicaAssignment {
	o := []v1alpha1.ReplicaAssignment{}

	c := v1alpha1.ReplicaAssignment{
		ID:      int64(2),
		Brokers: generateTestv1alpha1ReplicaAssignmentBrokers(),
	}
	o = append(o, c)

	c = v1alpha1.ReplicaAssignment{
		ID:      int64(3),
		Brokers: generateTestv1alpha1ReplicaAssignmentBrokers(),
	}

	o = append(o, c)

	return o
}

func generateTestv1alpha1ReplicaAssignmentBrokers() *v1alpha1.ReplicaAssignmentBrokers {
	o := &v1alpha1.ReplicaAssignmentBrokers{
		Replicas: generateTestv1alpha1Replicas(),
	}
	return o
}

func generateTestv1alpha1Replicas() []int64 {
	o := []int64{}

	c := int64(2)

	o = append(o, c)

	c = int64(3)

	o = append(o, c)

	return o
}

func instance(m ...func(*arv1.TopicDetail)) *arv1.TopicDetail {
	i := &arv1.TopicDetail{
		Name:              reference.ToPtrValue("myTopic"),
		Partitions:        ibmc.Int64Ptr(int64(2)),
		ReplicationFactor: ibmc.Int64Ptr(int64(2)),
		RetentionMs:       ibmc.Int64Ptr(int64(2)),
		CleanupPolicy:     reference.ToPtrValue("myCleanupPolicy"),
		// Configs:           &arv1.TopicConfigs{},
		Configs: generateTestarv1TopicConfigs(),
		// ReplicaAssignments: []arv1.ReplicaAssignment{},
		ReplicaAssignments: generateTestarv1ReplicaAssignments(),
	}

	for _, f := range m {
		f(i)
	}
	return i
}

func generateTestarv1TopicConfigs() *arv1.TopicConfigs {
	o := &arv1.TopicConfigs{
		CleanupPolicy:     reference.ToPtrValue("myCleanupPolicy"),
		MinInsyncReplicas: reference.ToPtrValue("myMinInsyncReplicas"),
		RetentionBytes:    reference.ToPtrValue("myRetentionBytes"),
		RetentionMs:       reference.ToPtrValue("myRetentionMs"),
		SegmentBytes:      reference.ToPtrValue("mySegmentBytes"),
		SegmentIndexBytes: reference.ToPtrValue("mySegmentIndexBytes"),
		SegmentMs:         reference.ToPtrValue("mySegmentMs"),
	}
	return o
}

func generateTestarv1ReplicaAssignments() []arv1.ReplicaAssignment {
	o := []arv1.ReplicaAssignment{}

	c := arv1.ReplicaAssignment{
		ID:      ibmc.Int64Ptr(int64(2)),
		Brokers: generateTestarv1ReplicaAssignmentBrokers(),
	}
	o = append(o, c)

	c = arv1.ReplicaAssignment{
		ID:      ibmc.Int64Ptr(int64(3)),
		Brokers: generateTestarv1ReplicaAssignmentBrokers(),
	}

	o = append(o, c)

	return o
}

func generateTestarv1ReplicaAssignmentBrokers() *arv1.ReplicaAssignmentBrokers {
	o := &arv1.ReplicaAssignmentBrokers{
		Replicas: generateTestarv1Replicas(),
	}
	return o
}

func generateTestarv1Replicas() []int64 {
	o := []int64{}

	c := int64(2)

	o = append(o, c)

	c = int64(3)

	o = append(o, c)

	return o
}

func instanceOpts(m ...func(*arv1.CreateTopicOptions)) *arv1.CreateTopicOptions {
	i := &arv1.CreateTopicOptions{
		Name:           reference.ToPtrValue("myTopic"),
		Partitions:     ibmc.Int64Ptr(int64(2)),
		PartitionCount: ibmc.Int64Ptr(int64(2)),
		// Configs:        []arv1.ConfigCreate{},
		Configs: generateTestarv1ConfigCreate(),
	}

	for _, f := range m {
		f(i)
	}
	return i
}

func generateTestarv1ConfigCreate() []arv1.ConfigCreate {
	o := []arv1.ConfigCreate{}

	c := arv1.ConfigCreate{
		Name:  reference.ToPtrValue("cleanup.policy"),
		Value: reference.ToPtrValue("myCleanupPolicy"),
	}
	o = append(o, c)
	c = arv1.ConfigCreate{
		Name:  reference.ToPtrValue("retention.bytes"),
		Value: reference.ToPtrValue("myRetentionBytes"),
	}
	o = append(o, c)
	c = arv1.ConfigCreate{
		Name:  reference.ToPtrValue("retention.ms"),
		Value: reference.ToPtrValue("myRetentionMs"),
	}
	o = append(o, c)
	c = arv1.ConfigCreate{
		Name:  reference.ToPtrValue("segment.bytes"),
		Value: reference.ToPtrValue("mySegmentBytes"),
	}
	o = append(o, c)
	c = arv1.ConfigCreate{
		Name:  reference.ToPtrValue("segment.index.bytes"),
		Value: reference.ToPtrValue("mySegmentIndexBytes"),
	}
	o = append(o, c)
	c = arv1.ConfigCreate{
		Name:  reference.ToPtrValue("segment.ms"),
		Value: reference.ToPtrValue("mySegmentMs"),
	}
	o = append(o, c)
	return o
}

func instanceUpdOpts(m ...func(*arv1.UpdateTopicOptions)) *arv1.UpdateTopicOptions {
	i := &arv1.UpdateTopicOptions{
		TopicName:              reference.ToPtrValue("myTopic"),
		NewTotalPartitionCount: ibmc.Int64Ptr(int64(2)),
		// Configs:                []arv1.ConfigUpdate{},
		Configs: generateTestarv1ConfigUpdate(),
	}

	for _, f := range m {
		f(i)
	}
	return i
}

func generateTestarv1ConfigUpdate() []arv1.ConfigUpdate {
	o := []arv1.ConfigUpdate{}

	c := arv1.ConfigUpdate{
		Name:           reference.ToPtrValue("cleanup.policy"),
		Value:          reference.ToPtrValue("myCleanupPolicy"),
		ResetToDefault: ibmc.BoolPtr(false),
	}
	o = append(o, c)
	c = arv1.ConfigUpdate{
		Name:           reference.ToPtrValue("retention.bytes"),
		Value:          reference.ToPtrValue("myRetentionBytes"),
		ResetToDefault: ibmc.BoolPtr(false),
	}
	o = append(o, c)
	c = arv1.ConfigUpdate{
		Name:           reference.ToPtrValue("retention.ms"),
		Value:          reference.ToPtrValue("myRetentionMs"),
		ResetToDefault: ibmc.BoolPtr(false),
	}
	o = append(o, c)
	c = arv1.ConfigUpdate{
		Name:           reference.ToPtrValue("segment.bytes"),
		Value:          reference.ToPtrValue("mySegmentBytes"),
		ResetToDefault: ibmc.BoolPtr(false),
	}
	o = append(o, c)
	c = arv1.ConfigUpdate{
		Name:           reference.ToPtrValue("segment.index.bytes"),
		Value:          reference.ToPtrValue("mySegmentIndexBytes"),
		ResetToDefault: ibmc.BoolPtr(false),
	}
	o = append(o, c)
	c = arv1.ConfigUpdate{
		Name:           reference.ToPtrValue("segment.ms"),
		Value:          reference.ToPtrValue("mySegmentMs"),
		ResetToDefault: ibmc.BoolPtr(false),
	}
	o = append(o, c)
	return o
}

// Test GenerateCreateTopicOptions method
func TestGenerateCreateTopicOptions(t *testing.T) {
	type args struct {
		params v1alpha1.TopicParameters
	}
	type want struct {
		instance *arv1.CreateTopicOptions
	}
	cases := map[string]struct {
		args args
		want want
	}{
		"FullConversion": {
			args: args{params: *params()},
			want: want{instance: instanceOpts()},
		},
		"MissingFields": {
			args: args{
				params: *params(func(p *v1alpha1.TopicParameters) {
					p.Partitions = nil
				})},
			want: want{instance: instanceOpts(func(p *arv1.CreateTopicOptions) {
				p.Partitions = nil
			})},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {

			r := &arv1.CreateTopicOptions{}
			GenerateCreateTopicOptions(tc.args.params, r)
			if diff := cmp.Diff(tc.want.instance, r); diff != "" {
				t.Errorf("GenerateCreateTopicOptions(...): -want, +got:\n%s", diff)
			}
		})
	}
}

// Test GenerateUpdateTopicOptions method
func TestGenerateUpdateTopicOptions(t *testing.T) {
	type args struct {
		params v1alpha1.TopicParameters
	}
	type want struct {
		instance *arv1.UpdateTopicOptions
	}
	cases := map[string]struct {
		args args
		want want
	}{
		"FullConversion": {
			args: args{params: *params()},
			want: want{instance: instanceUpdOpts()},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			r := &arv1.UpdateTopicOptions{}
			GenerateUpdateTopicOptions(ibmc.Int64Ptr(1), tc.args.params, r)
			if diff := cmp.Diff(tc.want.instance, r); diff != "" {
				t.Errorf("GenerateUpdateTopicOptions(...): -want, +got:\n%s", diff)
			}
		})
	}
}

// Test LateInitializeSpecs method
func TestLateInitializeSpecs(t *testing.T) {
	type args struct {
		instance *arv1.TopicDetail
		params   *v1alpha1.TopicParameters
	}
	type want struct {
		params *v1alpha1.TopicParameters
	}
	cases := map[string]struct {
		args args
		want want
	}{
		"SomeFields": {
			args: args{
				params: params(func(p *v1alpha1.TopicParameters) {
					p.Partitions = nil
				}),
				instance: instance(func(p *arv1.TopicDetail) {
				}),
			},
			want: want{
				params: params(func(p *v1alpha1.TopicParameters) {
				})},
		},
		"AllFilledAlready": {
			args: args{
				params:   params(),
				instance: instance(),
			},
			want: want{
				params: params()},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			LateInitializeSpec(tc.args.params, tc.args.instance)
			if diff := cmp.Diff(tc.want.params, tc.args.params); diff != "" {
				t.Errorf("LateInitializeSpec(...): -want, +got:\n%s", diff)
			}
		})
	}
}

// Test GenerateObservation method
func TestTopicGenerateObservation(t *testing.T) {
	type args struct {
		instance *arv1.TopicDetail
	}
	type want struct {
		obs v1alpha1.TopicObservation
	}
	cases := map[string]struct {
		args args
		want want
	}{
		"FullConversion": {
			args: args{
				instance: instance(),
			},
			want: want{*observation()},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			o, err := GenerateObservation(tc.args.instance)
			if diff := cmp.Diff(nil, err); diff != "" {
				t.Errorf("GenerateObservation(...): want error != got error:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.obs, o); diff != "" {
				t.Errorf("GenerateObservation(...): -want, +got:\n%s", diff)
			}
		})
	}
}

// Test IsUpToDate method
func TestIsUpToDate(t *testing.T) {
	type args struct {
		params   *v1alpha1.TopicParameters
		instance *arv1.TopicDetail
	}
	type want struct {
		upToDate bool
		isErr    bool
	}
	cases := map[string]struct {
		args args
		want want
	}{
		"IsUpToDate": {
			args: args{
				params:   params(),
				instance: instance(),
			},
			want: want{upToDate: true, isErr: false},
		},
		"NeedsUpdate": {
			args: args{
				params: params(func(crp *v1alpha1.TopicParameters) {
					// crp.Partitions = ibmc.Int64Ptr(int64(3))
				}),
				instance: instance(func(i *arv1.TopicDetail) {
					i.Partitions = ibmc.Int64Ptr(int64(3))
				}),
			},
			want: want{upToDate: false, isErr: false},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {

			r, err := IsUpToDate(tc.args.params, tc.args.instance, logging.NewNopLogger())
			if err != nil && !tc.want.isErr {
				t.Error("IsUpToDate(...) unexpected error")
			}
			if diff := cmp.Diff(tc.want.upToDate, r); diff != "" {
				t.Errorf("IsUpToDate(...): -want, +got:\n%s", diff)
			}
		})
	}
}
