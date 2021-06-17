package eventstreamsadminv1

import (
	"testing"

	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/google/go-cmp/cmp"

	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
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
		Configs:               []v1alpha1.ConfigCreate{},
	}

	for _, f := range m {
		f(p)
	}
	return p
}

func observation(m ...func(*v1alpha1.TopicObservation)) *v1alpha1.TopicObservation {
	o := &v1alpha1.TopicObservation{
		ReplicationFactor:  int64(2),
		RetentionMs:        int64(2),
		CleanupPolicy:      "myCleanupPolicy",
		Configs:            &v1alpha1.TopicConfigs{},
		ReplicaAssignments: []v1alpha1.ReplicaAssignment{},
	}

	for _, f := range m {
		f(o)
	}
	return o
}

func instance(m ...func(*arv1.TopicDetail)) *arv1.TopicDetail {
	i := &arv1.TopicDetail{
		Name:               reference.ToPtrValue("myTopic"),
		Partitions:         ibmc.Int64Ptr(int64(2)),
		ReplicationFactor:  ibmc.Int64Ptr(int64(2)),
		RetentionMs:        ibmc.Int64Ptr(int64(2)),
		CleanupPolicy:      reference.ToPtrValue("myCleanupPolicy"),
		Configs:            &arv1.TopicConfigs{},
		ReplicaAssignments: []arv1.ReplicaAssignment{},
	}

	for _, f := range m {
		f(i)
	}
	return i
}

func instanceOpts(m ...func(*arv1.CreateTopicOptions)) *arv1.CreateTopicOptions {
	i := &arv1.CreateTopicOptions{
		Name:           reference.ToPtrValue("myTopic"),
		Partitions:     ibmc.Int64Ptr(int64(2)),
		PartitionCount: ibmc.Int64Ptr(int64(2)),
		Configs:        []arv1.ConfigCreate{},
	}

	for _, f := range m {
		f(i)
	}
	return i
}

func instanceUpdOpts(m ...func(*arv1.UpdateTopicOptions)) *arv1.UpdateTopicOptions {
	i := &arv1.UpdateTopicOptions{
		TopicName:              reference.ToPtrValue("myTopic"),
		NewTotalPartitionCount: ibmc.Int64Ptr(int64(2)),
		Configs:                []arv1.ConfigUpdate{},
	}

	for _, f := range m {
		f(i)
	}
	return i
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
			GenerateUpdateTopicOptions(tc.args.params, r)
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
				instance: instance(func(p *arv1.TopicDetail) {
				}),
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
					crp.Name = "differentName"
				}),
				instance: instance(func(i *arv1.TopicDetail) {
					// i.Name = reference.ToPtrValue("differentName")
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
