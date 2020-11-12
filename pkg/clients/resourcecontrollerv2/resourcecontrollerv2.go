package resourcecontrollerv2

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-openapi/strfmt"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/pkg/errors"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/crossplane/crossplane-runtime/pkg/logging"

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
	errGetResPlaID          = "error getting resource plan ID"
	errGetResGroupID        = "error getting resource group ID"
	errServiceNotFound      = "Service: does not exist"
	errPendingReclamation   = "Instance is pending reclamation"
)

// LateInitializeSpec fills unassigned fields with the values in *rcv2.ResourceInstance object.
func LateInitializeSpec(client ibmc.ClientSession, spec *v1alpha1.ResourceInstanceParameters, in *rcv2.ResourceInstance) error { // nolint:gocyclo
	if spec.Target == "" {
		spec.Target = GenerateTarget(in)
	}
	if spec.AllowCleanup == nil {
		spec.AllowCleanup = in.AllowCleanup
	}
	if spec.EntityLock == nil {
		spec.EntityLock = ibmc.StringPtr(strconv.FormatBool(*in.Locked))
	}
	if spec.Parameters == nil {
		spec.Parameters = ibmc.GenerateRawExtensionFromMap(in.Parameters)
	}
	if spec.ResourcePlanName == "" {
		pName, err := ibmc.GetResourcePlanName(client, spec.ServiceName, *in.ResourcePlanID)
		if err != nil {
			return err
		}
		spec.ResourcePlanName = ibmc.StringValue(pName)
	}
	if spec.ResourceGroupName == "" {
		rgName, err := ibmc.GetResourceGroupName(client, ibmc.StringValue(in.ResourceGroupID))
		if err != nil {
			return err
		}
		spec.ResourceGroupName = rgName
	}
	if spec.ServiceName == "" {
		spec.ServiceName = ibmc.GetServiceName(in)
	}
	if spec.Tags == nil {
		tags, err := ibmc.GetResourceInstanceTags(client, ibmc.StringValue(in.TargetCrn))
		if err != nil {
			return err
		}
		spec.Tags = tags
	}
	return nil
}

// GenerateCreateResourceInstanceOptions produces ResourceInstanceOptions object from ResourceInstanceParameters object.
func GenerateCreateResourceInstanceOptions(client ibmc.ClientSession, name string, in v1alpha1.ResourceInstanceParameters, o *rcv2.CreateResourceInstanceOptions) error {
	rPlanID, err := ibmc.GetResourcePlanID(client, in.ServiceName, in.ResourcePlanName)
	if err != nil {
		return errors.Wrap(err, errGetResPlaID)
	}

	rgID, err := ibmc.GetResourceGroupID(client, in.ResourceGroupName)
	if err != nil {
		return errors.Wrap(err, errGetResGroupID)
	}

	o.Name = ibmc.StringPtr(name)
	o.Tags = in.Tags
	o.ResourcePlanID = rPlanID
	o.ResourceGroup = rgID
	o.Target = ibmc.StringPtr(in.Target)
	o.AllowCleanup = in.AllowCleanup
	o.EntityLock = in.EntityLock
	o.Parameters = ibmc.GenerateMapFromRawExtension(in.Parameters)
	return nil
}

// GenerateUpdateResourceInstanceOptions produces UpdateResourceInstanceOptions object from ResourceInstance object.
func GenerateUpdateResourceInstanceOptions(client ibmc.ClientSession, name, id string, in v1alpha1.ResourceInstanceParameters, o *rcv2.UpdateResourceInstanceOptions) error {
	rPlanID, err := ibmc.GetResourcePlanID(client, in.ServiceName, in.ResourcePlanName)
	if err != nil {
		return errors.Wrap(err, errGetResPlaID)
	}

	o.Name = ibmc.StringPtr(name)
	o.ResourcePlanID = rPlanID
	o.AllowCleanup = in.AllowCleanup
	o.ID = ibmc.StringPtr(id)
	o.Parameters = ibmc.GenerateMapFromRawExtension(in.Parameters)

	return nil
}

// GenerateObservation produces ResourceInstanceObservation object from *rcv2.ResourceInstance object.
func GenerateObservation(client ibmc.ClientSession, in *rcv2.ResourceInstance) (v1alpha1.ResourceInstanceObservation, error) {
	o := v1alpha1.ResourceInstanceObservation{
		AccountID:           ibmc.StringValue(in.AccountID),
		AllowCleanup:        *in.AllowCleanup,
		CreatedAt:           GenerateMetaV1Time(in.CreatedAt),
		Crn:                 ibmc.StringValue(in.Crn),
		DashboardURL:        ibmc.StringValue(in.DashboardURL),
		DeletedAt:           GenerateMetaV1Time(in.DeletedAt),
		GUID:                ibmc.StringValue(in.Guid),
		ID:                  ibmc.StringValue(in.ID),
		LastOperation:       ibmc.GenerateRawExtensionFromMap(in.LastOperation),
		Locked:              *in.Locked,
		Name:                ibmc.StringValue(in.Name),
		PlanHistory:         GeneratePlanHistory(in.PlanHistory),
		ResourceAliasesURL:  ibmc.StringValue(in.ResourceAliasesURL),
		ResourceBindingsURL: ibmc.StringValue(in.ResourceBindingsURL),
		ResourceGroupCrn:    ibmc.StringValue(in.ResourceGroupCrn),
		ResourceGroupID:     ibmc.StringValue(in.ResourceGroupID),
		ResourceID:          ibmc.StringValue(in.ResourceID),
		ResourceKeysURL:     ibmc.StringValue(in.ResourceKeysURL),
		ResourcePlanID:      ibmc.StringValue(in.ResourcePlanID),
		State:               ibmc.StringValue(in.State),
		SubType:             ibmc.StringValue(in.SubType),
		Target:              GenerateTarget(in),
		TargetCrn:           ibmc.StringValue(in.TargetCrn),
		Type:                ibmc.StringValue(in.Type),
		URL:                 ibmc.StringValue(in.URL),
		UpdatedAt:           GenerateMetaV1Time(in.UpdatedAt),
		Parameters:          ibmc.GenerateRawExtensionFromMap(in.Parameters),
	}
	// ServiceEndpoints can be found in instance.Parameters["service-endpoints"]
	tags, err := ibmc.GetResourceInstanceTags(client, ibmc.StringValue(in.Crn))
	if err != nil {
		return o, err
	}
	o.Tags = tags
	return o, nil
}

// GenerateTarget generates Target from Crn
func GenerateTarget(in *rcv2.ResourceInstance) string {
	if in.Crn == nil {
		return ""
	}
	crn, err := crn.Parse(*in.Crn)
	if err != nil {
		return ""
	}
	return crn.Region
}

// GenerateMetaV1Time converts strfmt.DateTime to metav1.Time
func GenerateMetaV1Time(t *strfmt.DateTime) *metav1.Time {
	if t == nil {
		return nil
	}
	tx := metav1.NewTime(time.Time(*t))
	return &tx
}

// GeneratePlanHistory generates v1alpha1.PlanHistoryItem[] from []rcv2.PlanHistoryItem
func GeneratePlanHistory(in []rcv2.PlanHistoryItem) []v1alpha1.PlanHistoryItem {
	if in == nil {
		return nil
	}
	o := make([]v1alpha1.PlanHistoryItem, 0)
	for _, phi := range in {
		item := v1alpha1.PlanHistoryItem{
			ResourcePlanID: ibmc.StringValue(phi.ResourcePlanID),
			StartDate:      GenerateMetaV1Time(phi.StartDate),
		}
		o = append(o, item)
	}
	return o
}

// IsUpToDate checks whether current state is up-to-date compared to the given
// set of parameters.
func IsUpToDate(client ibmc.ClientSession, name string, in *v1alpha1.ResourceInstanceParameters, observed *rcv2.ResourceInstance, l logging.Logger) (bool, error) {
	desired := in.DeepCopy()
	actual, err := GenerateResourceInstanceParameters(client, observed)
	if err != nil {
		return false, err
	}

	l.Info(cmp.Diff(desired, actual))

	// name needs special treatment as it is not present in ResourceInstanceParameters
	if name != *observed.Name {
		l.Info(cmp.Diff(name, *observed.Name))
		return false, nil
	}
	return cmp.Equal(desired, actual, cmpopts.EquateEmpty(), cmpopts.IgnoreFields(v1alpha1.ResourceInstanceParameters{})), nil
}

// GenerateResourceInstanceParameters generates service instance parameters from resource instance
func GenerateResourceInstanceParameters(client ibmc.ClientSession, in *rcv2.ResourceInstance) (*v1alpha1.ResourceInstanceParameters, error) {
	o := &v1alpha1.ResourceInstanceParameters{
		Target:       GenerateTarget(in),
		AllowCleanup: in.AllowCleanup,
		EntityLock:   ibmc.StringPtr(strconv.FormatBool(*in.Locked)),
		ServiceName:  ibmc.GetServiceName(in),
		Parameters:   ibmc.GenerateRawExtensionFromMap(in.Parameters),
	}
	rgName, err := ibmc.GetResourceGroupName(client, ibmc.StringValue(in.ResourceGroupID))
	if err != nil {
		return nil, err
	}
	o.ResourceGroupName = rgName
	pName, err := ibmc.GetResourcePlanName(client, o.ServiceName, ibmc.StringValue(in.ResourcePlanID))
	if err != nil {
		return nil, err
	}
	o.ResourcePlanName = ibmc.StringValue(pName)
	tags, err := ibmc.GetResourceInstanceTags(client, ibmc.StringValue(in.Crn))
	if err != nil {
		return nil, err
	}
	o.Tags = tags
	return o, nil
}

// IsInstanceNotFound returns true if the SDK returns a not found error
func IsInstanceNotFound(err error) bool {
	return strings.HasPrefix(err.Error(), errServiceNotFound)
}

// IsInstancePendingReclamation returns true if instance is being already deleted
func IsInstancePendingReclamation(err error) bool {
	return strings.Contains(err.Error(), errPendingReclamation) ||
		strings.Contains(err.Error(), http.StatusText(http.StatusNotFound))
}
