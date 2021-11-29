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
	"fmt"
	"sort"

	ibmBucketConf "github.com/IBM/ibm-cos-sdk-go-config/resourceconfigurationv1"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/reference"
	"github.com/google/go-cmp/cmp"

	"github.com/crossplane-contrib/provider-ibm-cloud/apis/cos/v1alpha1"
	ibmc "github.com/crossplane-contrib/provider-ibm-cloud/pkg/clients"
)

// Returns whether an activity tracking, or monitoring, is disabled
func isDisabled(aCRN *string) bool {
	return aCRN != nil && *aCRN == "0"
}

// LateInitializeSpec fills optional and unassigned fields with the values in the spec, from the info that comes from the cloud
//
// Params
// 	  spec - what we get from k8s
// 	  fromIBMCloud - ...what comes from the cloud
//
// Returns
//    whether the resource was late-initialized, any error
func LateInitializeSpec(spec *v1alpha1.BucketConfigParams, fromIBMCloud *ibmBucketConf.Bucket) (bool, error) { // nolint:gocyclo
	wasLateInitialized := false

	if spec.Name == nil {
		spec.Name = fromIBMCloud.Name

		wasLateInitialized = true
	}

	if spec.HardQuota == nil && fromIBMCloud.HardQuota != nil {
		spec.HardQuota = fromIBMCloud.HardQuota

		wasLateInitialized = (fromIBMCloud.HardQuota != nil)
	}

	if spec.Firewall == nil && fromIBMCloud.Firewall != nil {
		spec.Firewall = &v1alpha1.Firewall{
			AllowedIP: fromIBMCloud.Firewall.AllowedIp,
		}

		wasLateInitialized = true
	}

	if spec.ActivityTracking == nil {
		if fromIBMCloud.ActivityTracking != nil {
			spec.ActivityTracking = &v1alpha1.ActivityTracking{
				ActivityTrackerCRN: fromIBMCloud.ActivityTracking.ActivityTrackerCrn,
				ReadDataEvents:     fromIBMCloud.ActivityTracking.ReadDataEvents,
				WriteDataEvents:    fromIBMCloud.ActivityTracking.WriteDataEvents,
			}

			wasLateInitialized = true
		}
	} else {
		// Only initialize if we are not actually deleting the tracking activity
		if !isDisabled(spec.ActivityTracking.ActivityTrackerCRN) {
			if fromIBMCloud.ActivityTracking != nil {
				if spec.ActivityTracking.ReadDataEvents == nil {
					spec.ActivityTracking.ReadDataEvents = fromIBMCloud.ActivityTracking.ReadDataEvents

					wasLateInitialized = (fromIBMCloud.ActivityTracking.ReadDataEvents != nil)
				}

				if spec.ActivityTracking.WriteDataEvents == nil {
					spec.ActivityTracking.WriteDataEvents = fromIBMCloud.ActivityTracking.WriteDataEvents

					wasLateInitialized = (fromIBMCloud.ActivityTracking.WriteDataEvents != nil)
				}

				if spec.ActivityTracking.ActivityTrackerCRN == nil {
					spec.ActivityTracking.ActivityTrackerCRN = fromIBMCloud.ActivityTracking.ActivityTrackerCrn

					wasLateInitialized = (fromIBMCloud.ActivityTracking.ActivityTrackerCrn != nil)
				}
			}
		}
	}

	if spec.MetricsMonitoring == nil {
		if fromIBMCloud.MetricsMonitoring != nil {
			spec.MetricsMonitoring = &v1alpha1.MetricsMonitoring{
				MetricsMonitoringCRN:  fromIBMCloud.MetricsMonitoring.MetricsMonitoringCrn,
				UsageMetricsEnabled:   fromIBMCloud.MetricsMonitoring.UsageMetricsEnabled,
				RequestMetricsEnabled: fromIBMCloud.MetricsMonitoring.RequestMetricsEnabled,
			}

			wasLateInitialized = true
		}
	} else {
		// Only initialize if we are not actually deleting the monitoring thingy
		if !isDisabled(spec.MetricsMonitoring.MetricsMonitoringCRN) {
			if fromIBMCloud.MetricsMonitoring != nil {
				if spec.MetricsMonitoring.UsageMetricsEnabled == nil {
					spec.MetricsMonitoring.UsageMetricsEnabled = fromIBMCloud.MetricsMonitoring.UsageMetricsEnabled

					wasLateInitialized = (fromIBMCloud.MetricsMonitoring.UsageMetricsEnabled != nil)
				}

				if spec.MetricsMonitoring.RequestMetricsEnabled == nil {
					spec.MetricsMonitoring.RequestMetricsEnabled = fromIBMCloud.MetricsMonitoring.RequestMetricsEnabled

					wasLateInitialized = (fromIBMCloud.MetricsMonitoring.RequestMetricsEnabled != nil)
				}

				if spec.MetricsMonitoring.MetricsMonitoringCRN == nil {
					spec.MetricsMonitoring.MetricsMonitoringCRN = fromIBMCloud.MetricsMonitoring.MetricsMonitoringCrn

					wasLateInitialized = (fromIBMCloud.MetricsMonitoring.MetricsMonitoringCrn != nil)
				}
			}
		}
	}

	return wasLateInitialized, nil
}

// GenerateBucketConfigObservation returns an observation object, created with values taken from the 'in' parameter
func GenerateBucketConfigObservation(in *ibmBucketConf.Bucket) (v1alpha1.BucketConfigObservation, error) {
	result := v1alpha1.BucketConfigObservation{
		CRN:                reference.FromPtrValue(in.Crn),
		ServiceInstanceID:  reference.FromPtrValue(in.ServiceInstanceID),
		ServiceInstanceCRN: reference.FromPtrValue(in.ServiceInstanceCrn),
	}

	if in.TimeCreated != nil {
		result.TimeCreated = *ibmc.DateTimeToMetaV1Time(in.TimeCreated)
	}

	if in.TimeUpdated != nil {
		result.TimeUpdated = *ibmc.DateTimeToMetaV1Time(in.TimeUpdated)
	}

	if in.ObjectCount != nil {
		result.ObjectCount = ibmc.Int64Value(in.ObjectCount)
	}

	if in.BytesUsed != nil {
		result.BytesUsed = ibmc.Int64Value(in.BytesUsed)
	}

	if in.NoncurrentObjectCount != nil {
		result.NoncurrentObjectCount = ibmc.Int64Value(in.NoncurrentObjectCount)
	}

	if in.NoncurrentBytesUsed != nil {
		result.NoncurrentBytesUsed = ibmc.Int64Value(in.NoncurrentBytesUsed)
	}

	if in.DeleteMarkerCount != nil {
		result.DeleteMarkerCount = ibmc.Int64Value(in.DeleteMarkerCount)
	}

	return result, nil
}

// GenerateCloudBucketConfig populates the `out' object with values from the `in' object and the 'eTag'
func GenerateCloudBucketConfig(spec *v1alpha1.BucketConfigParams, eTag *string) (*ibmBucketConf.UpdateBucketConfigOptions, error) { // nolint:gocyclo
	result := ibmBucketConf.UpdateBucketConfigOptions{}

	if eTag != nil {
		result.SetIfMatch(*eTag)
	}

	result.SetBucket(*spec.Name)

	if spec.HardQuota != nil {
		result.SetHardQuota(*spec.HardQuota)
	}

	if spec.Headers != nil {
		result.SetHeaders(*spec.Headers)
	}

	if spec.Firewall != nil {
		result.SetFirewall(&ibmBucketConf.Firewall{
			AllowedIp: spec.Firewall.AllowedIP,
		})
	}

	if spec.ActivityTracking != nil {
		if isDisabled(spec.ActivityTracking.ActivityTrackerCRN) { // we want to delete
			result.SetActivityTracking(&ibmBucketConf.ActivityTracking{})
		} else if spec.ActivityTracking.ReadDataEvents != nil || spec.ActivityTracking.WriteDataEvents != nil {
			result.SetActivityTracking(
				&ibmBucketConf.ActivityTracking{
					ReadDataEvents:     spec.ActivityTracking.ReadDataEvents,
					WriteDataEvents:    spec.ActivityTracking.WriteDataEvents,
					ActivityTrackerCrn: spec.ActivityTracking.ActivityTrackerCRN,
				},
			)
		}
	}

	if spec.MetricsMonitoring != nil {
		if isDisabled(spec.MetricsMonitoring.MetricsMonitoringCRN) { // we want to delete
			result.SetMetricsMonitoring(&ibmBucketConf.MetricsMonitoring{})
		} else if spec.MetricsMonitoring.UsageMetricsEnabled != nil || spec.MetricsMonitoring.RequestMetricsEnabled != nil {
			result.SetMetricsMonitoring(
				&ibmBucketConf.MetricsMonitoring{
					UsageMetricsEnabled:   spec.MetricsMonitoring.UsageMetricsEnabled,
					RequestMetricsEnabled: spec.MetricsMonitoring.RequestMetricsEnabled,
					MetricsMonitoringCrn:  spec.MetricsMonitoring.MetricsMonitoringCRN,
				},
			)
		}
	}

	return &result, nil
}

// IsUpToDate checks whether the current bucket config (in the cloud) is up-to-date compared to the crossplane one
func IsUpToDate(in *v1alpha1.BucketConfigParams, observed *ibmBucketConf.Bucket, l logging.Logger) (bool, error) { //nolint:gocyclo
	desired := in.DeepCopy()
	actual, err := GenerateBucketConfigFromServerParams(observed)
	if err != nil {
		return false, err
	}

	// HardQuota comparison
	if actual.HardQuota == nil {
		if desired.HardQuota != nil && *desired.HardQuota != 0 {
			return false, nil
		}
	} else {
		diff := cmp.Diff(desired.HardQuota, actual.HardQuota)
		if diff != "" {
			fmt.Printf(">>> %s\n", diff)
			l.Info("IsUpToDate", "Diff", diff)

			return false, nil
		}
	}

	// Firewall comparison
	if desired.Firewall != nil && len(desired.Firewall.AllowedIP) != 0 {
		if actual.Firewall == nil {
			return false, nil
		}

		if len(desired.Firewall.AllowedIP) != len(actual.Firewall.AllowedIP) {
			return false, nil
		}

		sort.Strings(desired.Firewall.AllowedIP)
		sort.Strings(actual.Firewall.AllowedIP)

		diff := cmp.Diff(desired.Firewall, actual.Firewall)
		if diff != "" {
			fmt.Printf(">>> %s\n", diff)
			l.Info("IsUpToDate", "Diff", diff)

			return false, nil
		}
	}

	// ActivityTracking comparison
	if desired.ActivityTracking != nil {
		if isDisabled(desired.ActivityTracking.ActivityTrackerCRN) {
			if actual.ActivityTracking != nil {
				return false, nil
			}
		} else {
			diff := cmp.Diff(desired.ActivityTracking, actual.ActivityTracking)
			if diff != "" {
				fmt.Printf(">>> %s\n", diff)
				l.Info("IsUpToDate", "Diff", diff)

				return false, nil
			}
		}
	}

	// MetricsMonitoring comparison
	if desired.MetricsMonitoring != nil {
		if isDisabled(desired.MetricsMonitoring.MetricsMonitoringCRN) {
			if actual.MetricsMonitoring != nil {
				return false, nil
			}
		} else {
			diff := cmp.Diff(desired.MetricsMonitoring, actual.MetricsMonitoring)
			if diff != "" {
				fmt.Printf(">>> %s\n", diff)
				l.Info("IsUpToDate", "Diff", diff)

				return false, nil
			}
		}
	}

	return true, nil
}

// GenerateBucketConfigFromServerParams generates parameters for the crossplane object (bucket), from the one in the
// cloud.
// Moreover, if sets the 'Enabled' field to true whenever there is data returned for HardQuota, ActivityTracking etc
func GenerateBucketConfigFromServerParams(in *ibmBucketConf.Bucket) (*v1alpha1.BucketConfigParams, error) {
	result := &v1alpha1.BucketConfigParams{
		Name: in.Name,
	}

	if in.HardQuota != nil {
		result.HardQuota = ibmc.Int64Ptr(*in.HardQuota)
	}

	if in.Firewall != nil {
		result.Firewall = &v1alpha1.Firewall{
			AllowedIP: in.Firewall.AllowedIp,
		}
	}

	if in.ActivityTracking != nil {
		result.ActivityTracking = &v1alpha1.ActivityTracking{
			ActivityTrackerCRN: in.ActivityTracking.ActivityTrackerCrn,
			ReadDataEvents:     in.ActivityTracking.ReadDataEvents,
			WriteDataEvents:    in.ActivityTracking.WriteDataEvents,
		}
	}

	if in.MetricsMonitoring != nil {
		result.MetricsMonitoring = &v1alpha1.MetricsMonitoring{
			MetricsMonitoringCRN:  in.MetricsMonitoring.MetricsMonitoringCrn,
			RequestMetricsEnabled: in.MetricsMonitoring.RequestMetricsEnabled,
			UsageMetricsEnabled:   in.MetricsMonitoring.UsageMetricsEnabled,
		}
	}

	return result, nil
}
