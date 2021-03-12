package resourceinstance

import (
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/pkg/errors"

	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/reference"

	"github.com/IBM-Cloud/bluemix-go/crn"
	rcv2 "github.com/IBM/platform-services-go-sdk/resourcecontrollerv2"

	"github.com/crossplane-contrib/provider-ibm-cloud/apis/resourcecontrollerv2/v1alpha1"
	ibmc "github.com/crossplane-contrib/provider-ibm-cloud/pkg/clients"
)

const (
	// StateActive represents a service instance in a running, available, and ready state
	StateActive = "active"
	// StateInactive represents a service instance in a not running state
	StateInactive = "inactive"
	// StatePendingReclamation means that delete has been initiated
	StatePendingReclamation = "pending_reclamation"
	// StateRemoved means that delete has been completed
	StateRemoved     = "removed"
	errGetResPlaID   = "error getting resource plan ID"
	errGetResGroupID = "error getting resource group ID"
)

// LateInitializeSpec fills optional and unsassigned fields with the values in *rcv2.ResourceInstance object.
func LateInitializeSpec(client ibmc.ClientSession, spec *v1alpha1.ResourceInstanceParameters, in *rcv2.ResourceInstance) error { // nolint:gocyclo
	if spec.AllowCleanup == nil {
		spec.AllowCleanup = in.AllowCleanup
	}
	if spec.EntityLock == nil {
		spec.EntityLock = in.Locked
	}
	if spec.Parameters == nil {
		spec.Parameters = ibmc.MapToRawExtension(in.Parameters)
	}
	if spec.Tags == nil {
		tags, err := ibmc.GetResourceInstanceTags(client, reference.FromPtrValue(in.ID))
		if err != nil {
			return err
		}
		spec.Tags = tags
	}
	return nil
}

// GenerateCreateResourceInstanceOptions produces ResourceInstanceOptions object from ResourceInstanceParameters object.
func GenerateCreateResourceInstanceOptions(client ibmc.ClientSession, in v1alpha1.ResourceInstanceParameters, o *rcv2.CreateResourceInstanceOptions) error {
	rPlanID, err := ibmc.GetResourcePlanID(client, in.ServiceName, in.ResourcePlanName)
	if err != nil {
		return errors.Wrap(err, errGetResPlaID)
	}

	rgID, err := ibmc.GetResourceGroupID(client, in.ResourceGroupName)
	if err != nil {
		return errors.Wrap(err, errGetResGroupID)
	}

	o.Name = reference.ToPtrValue(in.Name)
	o.Tags = in.Tags
	o.ResourcePlanID = rPlanID
	o.ResourceGroup = rgID
	o.Target = reference.ToPtrValue(in.Target)
	o.AllowCleanup = in.AllowCleanup
	o.EntityLock = in.EntityLock
	o.Parameters = ibmc.RawExtensionToMap(in.Parameters)
	return nil
}

// GenerateUpdateResourceInstanceOptions produces UpdateResourceInstanceOptions object from ResourceInstance object.
func GenerateUpdateResourceInstanceOptions(client ibmc.ClientSession, id string, in v1alpha1.ResourceInstanceParameters, o *rcv2.UpdateResourceInstanceOptions) error {
	rPlanID, err := ibmc.GetResourcePlanID(client, in.ServiceName, in.ResourcePlanName)
	if err != nil {
		return errors.Wrap(err, errGetResPlaID)
	}

	o.Name = reference.ToPtrValue(in.Name)
	o.ResourcePlanID = rPlanID
	o.AllowCleanup = in.AllowCleanup
	o.ID = reference.ToPtrValue(id)
	o.Parameters = ibmc.RawExtensionToMap(in.Parameters)
	return nil
}

// GenerateObservation produces ResourceInstanceObservation object from *rcv2.ResourceInstance object.
func GenerateObservation(client ibmc.ClientSession, in *rcv2.ResourceInstance) (v1alpha1.ResourceInstanceObservation, error) {
	o := v1alpha1.ResourceInstanceObservation{
		AccountID:           reference.FromPtrValue(in.AccountID),
		CreatedAt:           ibmc.DateTimeToMetaV1Time(in.CreatedAt),
		CRN:                 reference.FromPtrValue(in.CRN),
		DashboardURL:        reference.FromPtrValue(in.DashboardURL),
		DeletedAt:           ibmc.DateTimeToMetaV1Time(in.DeletedAt),
		GUID:                reference.FromPtrValue(in.GUID),
		ID:                  reference.FromPtrValue(in.ID),
		LastOperation:       ibmc.MapToRawExtension(in.LastOperation),
		PlanHistory:         GeneratePlanHistory(in.PlanHistory),
		ResourceAliasesURL:  reference.FromPtrValue(in.ResourceAliasesURL),
		ResourceBindingsURL: reference.FromPtrValue(in.ResourceBindingsURL),
		ResourceGroupCRN:    reference.FromPtrValue(in.ResourceGroupCRN),
		ResourceGroupID:     reference.FromPtrValue(in.ResourceGroupID),
		ResourceID:          reference.FromPtrValue(in.ResourceID),
		ResourceKeysURL:     reference.FromPtrValue(in.ResourceKeysURL),
		ResourcePlanID:      reference.FromPtrValue(in.ResourcePlanID),
		State:               reference.FromPtrValue(in.State),
		SubType:             reference.FromPtrValue(in.SubType),
		TargetCRN:           reference.FromPtrValue(in.TargetCRN),
		Type:                reference.FromPtrValue(in.Type),
		URL:                 reference.FromPtrValue(in.URL),
		UpdatedAt:           ibmc.DateTimeToMetaV1Time(in.UpdatedAt),
		CreatedBy:           reference.FromPtrValue(in.CreatedBy),
		DeletedBy:           reference.FromPtrValue(in.DeletedBy),
		RestoredAt:          ibmc.DateTimeToMetaV1Time(in.RestoredAt),
		RestoredBy:          reference.FromPtrValue(in.RestoredBy),
		ScheduledReclaimAt:  ibmc.DateTimeToMetaV1Time(in.ScheduledReclaimAt),
		ScheduledReclaimBy:  reference.FromPtrValue(in.ScheduledReclaimBy),
		UpdatedBy:           reference.FromPtrValue(in.UpdatedBy),
	}
	// ServiceEndpoints can be found in instance.Parameters["service-endpoints"]
	return o, nil
}

// GenerateTarget generates Target from CRN
func GenerateTarget(in *rcv2.ResourceInstance) string {
	if in.CRN == nil {
		return ""
	}
	crn, err := crn.Parse(*in.CRN)
	if err != nil {
		return ""
	}
	return crn.Region
}

// GeneratePlanHistory generates v1alpha1.PlanHistoryItem[] from []rcv2.PlanHistoryItem
func GeneratePlanHistory(in []rcv2.PlanHistoryItem) []v1alpha1.PlanHistoryItem {
	if in == nil {
		return nil
	}
	o := make([]v1alpha1.PlanHistoryItem, 0)
	for _, phi := range in {
		item := v1alpha1.PlanHistoryItem{
			ResourcePlanID: reference.FromPtrValue(phi.ResourcePlanID),
			StartDate:      ibmc.DateTimeToMetaV1Time(phi.StartDate),
		}
		o = append(o, item)
	}
	return o
}

// IsUpToDate checks whether current state is up-to-date compared to the given set of parameters.
func IsUpToDate(client ibmc.ClientSession, in *v1alpha1.ResourceInstanceParameters, observed *rcv2.ResourceInstance, l logging.Logger) (bool, error) {
	desired := in.DeepCopy()
	actual, err := GenerateResourceInstanceParameters(client, observed)
	if err != nil {
		return false, err
	}

	l.Info(cmp.Diff(desired, actual))

	return cmp.Equal(desired, actual, cmpopts.EquateEmpty(), cmpopts.IgnoreFields(v1alpha1.ResourceInstanceParameters{})), nil
}

// GenerateResourceInstanceParameters generates service instance parameters from resource instance
func GenerateResourceInstanceParameters(client ibmc.ClientSession, in *rcv2.ResourceInstance) (*v1alpha1.ResourceInstanceParameters, error) {
	o := &v1alpha1.ResourceInstanceParameters{
		Name:         reference.FromPtrValue(in.Name),
		Target:       GenerateTarget(in),
		AllowCleanup: in.AllowCleanup,
		EntityLock:   in.Locked,
		ServiceName:  ibmc.GetServiceName(in),
		Parameters:   ibmc.MapToRawExtension(in.Parameters),
	}
	rgName, err := ibmc.GetResourceGroupName(client, reference.FromPtrValue(in.ResourceGroupID))
	if err != nil {
		return nil, err
	}
	o.ResourceGroupName = rgName
	pName, err := ibmc.GetResourcePlanName(client, o.ServiceName, reference.FromPtrValue(in.ResourcePlanID))
	if err != nil {
		return nil, err
	}
	o.ResourcePlanName = reference.FromPtrValue(pName)
	tags, err := ibmc.GetResourceInstanceTags(client, reference.FromPtrValue(in.CRN))
	if err != nil {
		return nil, err
	}
	o.Tags = tags
	return o, nil
}
