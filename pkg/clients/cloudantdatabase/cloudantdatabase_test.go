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
	"testing"

	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/google/go-cmp/cmp"

	"github.com/crossplane/crossplane-runtime/pkg/reference"

	cv1 "github.com/IBM/cloudant-go-sdk/cloudantv1"

	"github.com/crossplane-contrib/provider-ibm-cloud/apis/cloudantv1/v1alpha1"
	ibmc "github.com/crossplane-contrib/provider-ibm-cloud/pkg/clients"
)

func params(m ...func(*v1alpha1.CloudantDatabaseParameters)) *v1alpha1.CloudantDatabaseParameters {
	p := &v1alpha1.CloudantDatabaseParameters{
		Db:          "mydatabase",
		Partitioned: ibmc.BoolPtr(false),
		Q:           ibmc.Int64Ptr(int64(2)),
	}

	for _, f := range m {
		f(p)
	}
	return p
}

func observation(m ...func(*v1alpha1.CloudantDatabaseObservation)) *v1alpha1.CloudantDatabaseObservation {
	o := &v1alpha1.CloudantDatabaseObservation{
		Cluster:            generateTestv1alpha1DatabaseInformationCluster(),
		CommittedUpdateSeq: "myCommittedUpdateSeq",
		CompactRunning:     false,
		CompactedSeq:       "myCompactedSeq",
		DiskFormatVersion:  int64(2),
		DocCount:           int64(2),
		DocDelCount:        int64(2),
		Engine:             "myEngine",
		Sizes:              generateTestv1alpha1ContentInformationSizes(),
		UpdateSeq:          "myUpdateSeq",
		UUID:               "myUUID",
	}

	for _, f := range m {
		f(o)
	}
	return o
}

func generateTestv1alpha1DatabaseInformationCluster() *v1alpha1.DatabaseInformationCluster {
	o := &v1alpha1.DatabaseInformationCluster{
		N: int64(2),
		R: int64(2),
		W: int64(2),
	}
	return o
}

func generateTestv1alpha1ContentInformationSizes() *v1alpha1.ContentInformationSizes {
	o := &v1alpha1.ContentInformationSizes{
		Active:   int64(2),
		External: int64(2),
		File:     int64(2),
	}
	return o
}

func instance(m ...func(*cv1.DatabaseInformation)) *cv1.DatabaseInformation {
	i := &cv1.DatabaseInformation{
		Cluster:            generateTestcv1DatabaseInformationCluster(),
		CommittedUpdateSeq: reference.ToPtrValue("myCommittedUpdateSeq"),
		CompactRunning:     ibmc.BoolPtr(false),
		CompactedSeq:       reference.ToPtrValue("myCompactedSeq"),
		DbName:             reference.ToPtrValue("mydatabase"),
		DiskFormatVersion:  ibmc.Int64Ptr(int64(2)),
		DocCount:           ibmc.Int64Ptr(int64(2)),
		DocDelCount:        ibmc.Int64Ptr(int64(2)),
		Engine:             reference.ToPtrValue("myEngine"),
		Props:              generateTestcv1DatabaseInformationProps(),
		Sizes:              generateTestcv1ContentInformationSizes(),
		UpdateSeq:          reference.ToPtrValue("myUpdateSeq"),
		UUID:               reference.ToPtrValue("myUUID"),
	}

	for _, f := range m {
		f(i)
	}
	return i
}

func generateTestcv1DatabaseInformationCluster() *cv1.DatabaseInformationCluster {
	o := &cv1.DatabaseInformationCluster{
		N: ibmc.Int64Ptr(int64(2)),
		Q: ibmc.Int64Ptr(int64(2)),
		R: ibmc.Int64Ptr(int64(2)),
		W: ibmc.Int64Ptr(int64(2)),
	}
	return o
}

func generateTestcv1DatabaseInformationProps() *cv1.DatabaseInformationProps {
	o := &cv1.DatabaseInformationProps{
		Partitioned: ibmc.BoolPtr(false),
	}
	return o
}

func generateTestcv1ContentInformationSizes() *cv1.ContentInformationSizes {
	o := &cv1.ContentInformationSizes{
		Active:   ibmc.Int64Ptr(int64(2)),
		External: ibmc.Int64Ptr(int64(2)),
		File:     ibmc.Int64Ptr(int64(2)),
	}
	return o
}

func instanceOpts(m ...func(*cv1.PutDatabaseOptions)) *cv1.PutDatabaseOptions {
	i := &cv1.PutDatabaseOptions{
		Db:          reference.ToPtrValue("mydatabase"),
		Partitioned: ibmc.BoolPtr(false),
		Q:           ibmc.Int64Ptr(int64(2)),
	}

	for _, f := range m {
		f(i)
	}
	return i
}

// Test GenerateCreateCloudantDatabaseOptions method
func TestGenerateCreateCloudantDatabaseOptions(t *testing.T) {
	type args struct {
		params v1alpha1.CloudantDatabaseParameters
	}
	type want struct {
		instance *cv1.PutDatabaseOptions
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
				params: *params(func(p *v1alpha1.CloudantDatabaseParameters) {
					p.Partitioned = nil
				})},
			want: want{instance: instanceOpts(func(p *cv1.PutDatabaseOptions) {
				p.Partitioned = nil
			})},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {

			r := &cv1.PutDatabaseOptions{}
			GenerateCreateCloudantDatabaseOptions(tc.args.params, r)
			if diff := cmp.Diff(tc.want.instance, r); diff != "" {
				t.Errorf("GenerateCreateCloudantDatabaseOptions(...): -want, +got:\n%s", diff)
			}
		})
	}
}

// Test LateInitializeSpecs method
func TestLateInitializeSpecs(t *testing.T) {
	type args struct {
		instance *cv1.DatabaseInformation
		params   *v1alpha1.CloudantDatabaseParameters
	}
	type want struct {
		params *v1alpha1.CloudantDatabaseParameters
	}
	cases := map[string]struct {
		args args
		want want
	}{
		"SomeFields": {
			args: args{
				params: params(func(p *v1alpha1.CloudantDatabaseParameters) {
					p.Partitioned = nil
				}),
				instance: instance(func(p *cv1.DatabaseInformation) {
				}),
			},
			want: want{
				params: params(func(p *v1alpha1.CloudantDatabaseParameters) {
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
func TestCloudantDatabaseGenerateObservation(t *testing.T) {
	type args struct {
		instance *cv1.DatabaseInformation
	}
	type want struct {
		obs v1alpha1.CloudantDatabaseObservation
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
		params   *v1alpha1.CloudantDatabaseParameters
		instance *cv1.DatabaseInformation
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
				params: params(func(crp *v1alpha1.CloudantDatabaseParameters) {
					crp.Partitioned = ibmc.BoolPtr(true)
				}),
				instance: instance(func(i *cv1.DatabaseInformation) {
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
