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
	"testing"

	"github.com/google/go-cmp/cmp"

	ibmBucketConf "github.com/IBM/ibm-cos-sdk-go-config/resourceconfigurationv1"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/reference"

	"github.com/crossplane-contrib/provider-ibm-cloud/apis/cos/v1alpha1"
	ibmc "github.com/crossplane-contrib/provider-ibm-cloud/pkg/clients"
)

// Returns some headers
func someHeaders() *map[string]string {
	result := map[string]string{
		"alpha": "beta",
		"gamma": "delta",
		"":      "bar",
	}

	return &result
}

// Test the LateInitializeSpec method
func TestLateInitializeSpec(t *testing.T) {
	cases := map[string]struct {
		spec      v1alpha1.BucketConfigParams
		ibmConfig ibmBucketConf.Bucket
		want      v1alpha1.BucketConfigParams
	}{
		"TestLateInitializeSpec-1": {
			spec: v1alpha1.BucketConfigParams{},
			ibmConfig: ibmBucketConf.Bucket{
				Name:      reference.ToPtrValue("foo"),
				HardQuota: ibmc.Int64Ptr(int64(300)),
				Firewall: &ibmBucketConf.Firewall{
					AllowedIp: AStrArray(),
				},
				ActivityTracking: &ibmBucketConf.ActivityTracking{
					ReadDataEvents:     ibmc.BoolPtr(true),
					WriteDataEvents:    ibmc.BoolPtr(false),
					ActivityTrackerCrn: reference.ToPtrValue("mama mia"),
				},
			},
			want: v1alpha1.BucketConfigParams{
				Name:      reference.ToPtrValue("foo"),
				HardQuota: ibmc.Int64Ptr(int64(300)),
				Firewall: &v1alpha1.Firewall{
					AllowedIP: AStrArray(),
				},
				ActivityTracking: &v1alpha1.ActivityTracking{
					ReadDataEvents:     ibmc.BoolPtr(true),
					WriteDataEvents:    ibmc.BoolPtr(false),
					ActivityTrackerCRN: reference.ToPtrValue("mama mia"),
				},
			},
		},
		"TestLateInitializeSpec-2": {
			spec: v1alpha1.BucketConfigParams{
				Name: reference.ToPtrValue("foo"),
				ActivityTracking: &v1alpha1.ActivityTracking{
					ReadDataEvents:     ibmc.BoolPtr(false),
					WriteDataEvents:    ibmc.BoolPtr(false),
					ActivityTrackerCRN: reference.ToPtrValue("nothing"),
				},
				MetricsMonitoring: &v1alpha1.MetricsMonitoring{
					UsageMetricsEnabled:  ibmc.BoolPtr(true),
					MetricsMonitoringCRN: reference.ToPtrValue("fap"),
				},
			},
			ibmConfig: ibmBucketConf.Bucket{
				Name:      reference.ToPtrValue("foo"),
				HardQuota: ibmc.Int64Ptr(int64(300)),
				Firewall: &ibmBucketConf.Firewall{
					AllowedIp: AStrArray(),
				},
				ActivityTracking: &ibmBucketConf.ActivityTracking{
					ReadDataEvents:     ibmc.BoolPtr(true),
					WriteDataEvents:    ibmc.BoolPtr(false),
					ActivityTrackerCrn: reference.ToPtrValue("mama mia"),
				},
			},
			want: v1alpha1.BucketConfigParams{
				Name:      reference.ToPtrValue("foo"),
				HardQuota: ibmc.Int64Ptr(int64(300)),
				Firewall: &v1alpha1.Firewall{
					AllowedIP: AStrArray(),
				},
				ActivityTracking: &v1alpha1.ActivityTracking{
					ReadDataEvents:     ibmc.BoolPtr(false),
					WriteDataEvents:    ibmc.BoolPtr(false),
					ActivityTrackerCRN: reference.ToPtrValue("nothing"),
				},
				MetricsMonitoring: &v1alpha1.MetricsMonitoring{
					UsageMetricsEnabled:  ibmc.BoolPtr(true),
					MetricsMonitoringCRN: reference.ToPtrValue("fap"),
				},
			},
		},
		"TestLateInitializeSpec-3": {
			spec: v1alpha1.BucketConfigParams{
				Name:      reference.ToPtrValue("foo"),
				HardQuota: ibmc.Int64Ptr(int64(300)),
				MetricsMonitoring: &v1alpha1.MetricsMonitoring{
					UsageMetricsEnabled:  ibmc.BoolPtr(true),
					MetricsMonitoringCRN: reference.ToPtrValue("fap"),
				},
			},
			ibmConfig: ibmBucketConf.Bucket{
				Name:      reference.ToPtrValue("foo"),
				HardQuota: ibmc.Int64Ptr(int64(100)),
				Firewall: &ibmBucketConf.Firewall{
					AllowedIp: AStrArray(),
				},
				ActivityTracking: &ibmBucketConf.ActivityTracking{
					ReadDataEvents:     ibmc.BoolPtr(true),
					WriteDataEvents:    ibmc.BoolPtr(false),
					ActivityTrackerCrn: reference.ToPtrValue("mama mia"),
				},
			},
			want: v1alpha1.BucketConfigParams{
				Name:      reference.ToPtrValue("foo"),
				HardQuota: ibmc.Int64Ptr(int64(300)),
				Firewall: &v1alpha1.Firewall{
					AllowedIP: AStrArray(),
				},
				ActivityTracking: &v1alpha1.ActivityTracking{
					ReadDataEvents:     ibmc.BoolPtr(true),
					WriteDataEvents:    ibmc.BoolPtr(false),
					ActivityTrackerCRN: reference.ToPtrValue("mama mia"),
				},
				MetricsMonitoring: &v1alpha1.MetricsMonitoring{
					UsageMetricsEnabled:  ibmc.BoolPtr(true),
					MetricsMonitoringCRN: reference.ToPtrValue("fap"),
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			LateInitializeSpec(&tc.spec, &tc.ibmConfig)
			if diff := cmp.Diff(tc.want, tc.spec); diff != "" {
				t.Errorf("LateInitializeSpec(...): -want, +got:\n%s", diff)
			}
		})
	}
}

// Tests the GenerateBucketConfigObservation function
func TestGenerateBucketConfigObservation(t *testing.T) {
	cases := map[string]struct {
		ibmConfig ibmBucketConf.Bucket
		want      v1alpha1.BucketConfigObservation
	}{
		"TestGenerateBucketConfigObservation": {
			ibmConfig: ibmBucketConf.Bucket{
				Crn:                   reference.ToPtrValue("fooooo"),
				ServiceInstanceID:     reference.ToPtrValue("what an id"),
				ServiceInstanceCrn:    reference.ToPtrValue("some crn"),
				TimeCreated:           ibmc.ADateTimeInAYear(1),
				TimeUpdated:           ibmc.ADateTimeInAYear(2),
				ObjectCount:           ibmc.Int64Ptr(33),
				BytesUsed:             ibmc.Int64Ptr(31),
				NoncurrentObjectCount: ibmc.Int64Ptr(41),
				NoncurrentBytesUsed:   ibmc.Int64Ptr(42),
				DeleteMarkerCount:     ibmc.Int64Ptr(43),
			},
			want: v1alpha1.BucketConfigObservation{
				CRN:                   "fooooo",
				ServiceInstanceID:     "what an id",
				ServiceInstanceCRN:    "some crn",
				TimeCreated:           *ibmc.DateTimeToMetaV1Time(ibmc.ADateTimeInAYear(1)),
				TimeUpdated:           *ibmc.DateTimeToMetaV1Time(ibmc.ADateTimeInAYear(2)),
				ObjectCount:           33,
				BytesUsed:             31,
				NoncurrentObjectCount: 41,
				NoncurrentBytesUsed:   42,
				DeleteMarkerCount:     43,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			obs, _ := GenerateBucketConfigObservation(&tc.ibmConfig)

			if diff := cmp.Diff(tc.want, obs); diff != "" {
				t.Errorf("GenerateBucketConfigObservation(...): -want, +got:\n%s", diff)
			}
		})
	}
}

// Tests the GenerateCloudBucketConfig function
func TestGenerateCloudBucketConfig(t *testing.T) {
	cases := map[string]struct {
		kubeBC v1alpha1.BucketConfigParams
		want   ibmBucketConf.UpdateBucketConfigOptions
	}{
		"TestGenerateCloudBucketConfig-1": {
			kubeBC: v1alpha1.BucketConfigParams{
				Name: reference.ToPtrValue("name-1"),
			},
			want: ibmBucketConf.UpdateBucketConfigOptions{
				Bucket: reference.ToPtrValue("name-1"),
			},
		},
		"TestGenerateCloudBucketConfig-2": {
			kubeBC: v1alpha1.BucketConfigParams{
				Name:      reference.ToPtrValue("name-2"),
				HardQuota: ibmc.Int64Ptr(0),
			},
			want: ibmBucketConf.UpdateBucketConfigOptions{
				Bucket:    reference.ToPtrValue("name-2"),
				HardQuota: ibmc.Int64Ptr(0),
			},
		},
		"TestGenerateCloudBucketConfig-3": {
			kubeBC: v1alpha1.BucketConfigParams{
				Name:      reference.ToPtrValue("name-3"),
				HardQuota: ibmc.Int64Ptr(6000),
			},
			want: ibmBucketConf.UpdateBucketConfigOptions{
				Bucket:    reference.ToPtrValue("name-3"),
				HardQuota: ibmc.Int64Ptr(6000),
			},
		},
		"TestGenerateCloudBucketConfig-4": {
			kubeBC: v1alpha1.BucketConfigParams{
				Name:      reference.ToPtrValue("name-3"),
				HardQuota: ibmc.Int64Ptr(6000),
				Headers:   someHeaders(),
			},
			want: ibmBucketConf.UpdateBucketConfigOptions{
				Bucket:    reference.ToPtrValue("name-3"),
				HardQuota: ibmc.Int64Ptr(6000),
				Headers:   *someHeaders(),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			obs, _ := GenerateCloudBucketConfig(&tc.kubeBC, nil)

			if diff := cmp.Diff(tc.want, *obs); diff != "" {
				t.Errorf("GenerateCloudBucketConfig(...): -want, +got:\n%s", diff)
			}
		})
	}
}

// Tests the GenerateBucketConfigFromServerParams function
func TestGenerateBucketConfigFromServerParams(t *testing.T) {
	cases := map[string]struct {
		ibmConfig ibmBucketConf.Bucket
		want      v1alpha1.BucketConfigParams
	}{
		"TestGenerateBucketConfigFromServerParams": {
			ibmConfig: ibmBucketConf.Bucket{
				Name:      reference.ToPtrValue("foo"),
				HardQuota: ibmc.Int64Ptr(int64(300)),
				Firewall: &ibmBucketConf.Firewall{
					AllowedIp: AStrArray(),
				},
				ActivityTracking: &ibmBucketConf.ActivityTracking{
					ReadDataEvents:     ibmc.BoolPtr(true),
					WriteDataEvents:    ibmc.BoolPtr(false),
					ActivityTrackerCrn: reference.ToPtrValue("mama mia"),
				},
				MetricsMonitoring: &ibmBucketConf.MetricsMonitoring{
					UsageMetricsEnabled:   ibmc.BoolPtr(true),
					RequestMetricsEnabled: ibmc.BoolPtr(false),
					MetricsMonitoringCrn:  reference.ToPtrValue("mama due"),
				},
			},
			want: v1alpha1.BucketConfigParams{
				Name:      reference.ToPtrValue("foo"),
				HardQuota: ibmc.Int64Ptr(300),
				Firewall: &v1alpha1.Firewall{
					AllowedIP: AStrArray(),
				},
				ActivityTracking: &v1alpha1.ActivityTracking{
					ReadDataEvents:     ibmc.BoolPtr(true),
					WriteDataEvents:    ibmc.BoolPtr(false),
					ActivityTrackerCRN: reference.ToPtrValue("mama mia"),
				},
				MetricsMonitoring: &v1alpha1.MetricsMonitoring{
					UsageMetricsEnabled:   ibmc.BoolPtr(true),
					RequestMetricsEnabled: ibmc.BoolPtr(false),
					MetricsMonitoringCRN:  reference.ToPtrValue("mama due"),
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			kubeBC, _ := GenerateBucketConfigFromServerParams(&tc.ibmConfig)

			if diff := cmp.Diff(tc.want, *kubeBC); diff != "" {
				t.Errorf("GenerateBucketConfigFromServerParams(...): -want, +got:\n%s", diff)
			}
		})
	}
}

// Test the IsUpToDate method
func TestIsUpToDate(t *testing.T) {
	cases := map[string]struct {
		spec     v1alpha1.BucketConfigParams
		observed ibmBucketConf.Bucket
		want     bool
	}{
		"TestIsUpToDate-1": {
			spec: v1alpha1.BucketConfigParams{
				Name:      reference.ToPtrValue("foo"),
				HardQuota: ibmc.Int64Ptr(int64(300)),
				Firewall: &v1alpha1.Firewall{
					AllowedIP: AStrArray(),
				},
				ActivityTracking: &v1alpha1.ActivityTracking{
					ReadDataEvents:     ibmc.BoolPtr(true),
					WriteDataEvents:    ibmc.BoolPtr(false),
					ActivityTrackerCRN: reference.ToPtrValue("mama mia"),
				},
			},
			observed: ibmBucketConf.Bucket{
				Name:      reference.ToPtrValue("foo"),
				HardQuota: ibmc.Int64Ptr(int64(300)),
				Firewall: &ibmBucketConf.Firewall{
					AllowedIp: AStrArray(),
				},
				ActivityTracking: &ibmBucketConf.ActivityTracking{
					ReadDataEvents:     ibmc.BoolPtr(true),
					WriteDataEvents:    ibmc.BoolPtr(false),
					ActivityTrackerCrn: reference.ToPtrValue("mama mia"),
				},
			},
			want: true,
		},
		"TestIsUpToDate-2": {
			spec: v1alpha1.BucketConfigParams{
				Name:      reference.ToPtrValue("foo"),
				HardQuota: ibmc.Int64Ptr(int64(300)),
				Firewall: &v1alpha1.Firewall{
					AllowedIP: AStrArray(),
				},
				ActivityTracking: &v1alpha1.ActivityTracking{
					ReadDataEvents:     ibmc.BoolPtr(true),
					WriteDataEvents:    ibmc.BoolPtr(false),
					ActivityTrackerCRN: reference.ToPtrValue("mama mia"),
				},
			},
			observed: ibmBucketConf.Bucket{
				Name:      reference.ToPtrValue("foo"),
				HardQuota: ibmc.Int64Ptr(int64(300)),
				Firewall: &ibmBucketConf.Firewall{
					AllowedIp: AStrArray(),
				},
				ActivityTracking: &ibmBucketConf.ActivityTracking{
					ReadDataEvents:     ibmc.BoolPtr(false),
					WriteDataEvents:    ibmc.BoolPtr(false),
					ActivityTrackerCrn: reference.ToPtrValue("mama mia"),
				},
			},
			want: false,
		},
		"TestIsUpToDate-3": {
			spec: v1alpha1.BucketConfigParams{
				Name:      reference.ToPtrValue("foo"),
				HardQuota: ibmc.Int64Ptr(int64(300)),
			},
			observed: ibmBucketConf.Bucket{
				Name:      reference.ToPtrValue("foo"),
				HardQuota: ibmc.Int64Ptr(int64(300)),
				Firewall: &ibmBucketConf.Firewall{
					AllowedIp: AStrArray(),
				},
			},
			want: true,
		},
		"TestIsUpToDate-4": {
			spec: v1alpha1.BucketConfigParams{
				Name:      reference.ToPtrValue("foo"),
				HardQuota: ibmc.Int64Ptr(int64(300)),
				ActivityTracking: &v1alpha1.ActivityTracking{
					ReadDataEvents:     ibmc.BoolPtr(true),
					WriteDataEvents:    ibmc.BoolPtr(false),
					ActivityTrackerCRN: reference.ToPtrValue("mama mia"),
				},
			},
			observed: ibmBucketConf.Bucket{
				Name:      reference.ToPtrValue("foo"),
				HardQuota: ibmc.Int64Ptr(int64(300)),
			},
			want: false,
		},
		"TestIsUpToDate-5": {
			spec: v1alpha1.BucketConfigParams{
				Name: reference.ToPtrValue("foo"),
				MetricsMonitoring: &v1alpha1.MetricsMonitoring{
					UsageMetricsEnabled:   ibmc.BoolPtr(true),
					RequestMetricsEnabled: ibmc.BoolPtr(false),
					MetricsMonitoringCRN:  reference.ToPtrValue("mama due"),
				},
			},
			observed: ibmBucketConf.Bucket{
				Name: reference.ToPtrValue("foo"),
				MetricsMonitoring: &ibmBucketConf.MetricsMonitoring{
					UsageMetricsEnabled:   ibmc.BoolPtr(true),
					RequestMetricsEnabled: ibmc.BoolPtr(false),
					MetricsMonitoringCrn:  reference.ToPtrValue("mama due"),
				},
			},
			want: true,
		},
		"TestIsUpToDate-6": {
			spec: v1alpha1.BucketConfigParams{
				Name: reference.ToPtrValue("foo"),
				MetricsMonitoring: &v1alpha1.MetricsMonitoring{
					UsageMetricsEnabled:   ibmc.BoolPtr(true),
					RequestMetricsEnabled: ibmc.BoolPtr(false),
					MetricsMonitoringCRN:  reference.ToPtrValue("mama due"),
				},
			},
			observed: ibmBucketConf.Bucket{
				Name: reference.ToPtrValue("foo"),
				MetricsMonitoring: &ibmBucketConf.MetricsMonitoring{
					UsageMetricsEnabled:   ibmc.BoolPtr(false),
					RequestMetricsEnabled: ibmc.BoolPtr(false),
					MetricsMonitoringCrn:  reference.ToPtrValue("mama due"),
				},
			},
			want: false,
		},
		"TestIsUpToDate-7": {
			spec: v1alpha1.BucketConfigParams{
				Name: reference.ToPtrValue("foo"),
				MetricsMonitoring: &v1alpha1.MetricsMonitoring{
					UsageMetricsEnabled:   ibmc.BoolPtr(true),
					RequestMetricsEnabled: ibmc.BoolPtr(false),
					MetricsMonitoringCRN:  reference.ToPtrValue("mama due"),
				},
			},
			observed: ibmBucketConf.Bucket{
				Name: reference.ToPtrValue("foo"),
				MetricsMonitoring: &ibmBucketConf.MetricsMonitoring{
					UsageMetricsEnabled:   ibmc.BoolPtr(true),
					RequestMetricsEnabled: ibmc.BoolPtr(false),
					MetricsMonitoringCrn:  reference.ToPtrValue("mama due"),
				},
			},
			want: true,
		},
	}

	for name, tc := range cases {
		if name != "TestIsUpToDate-3" {
			continue
		}
		t.Run(name, func(t *testing.T) {
			rc, _ := IsUpToDate(&tc.spec, &tc.observed, logging.NewNopLogger())
			if rc != tc.want {
				t.Errorf("IsUpToDate(...): -want:%t, +got:%t\n", tc.want, rc)
			}
		})
	}
}
