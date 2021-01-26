package customrole

import (
	"time"

	"github.com/go-openapi/strfmt"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/reference"

	iampmv1 "github.com/IBM/platform-services-go-sdk/iampolicymanagementv1"

	"github.com/crossplane-contrib/provider-ibm-cloud/apis/iampolicymanagementv1/v1alpha1"
)

const (
	// StateActive represents a policy in a running, available, and ready state
	StateActive = "active"
	// StateInactive represents an inactive policy
	StateInactive = "inactive"
)

// LateInitializeSpec fills optional and unassigned fields with the values in *iampmv1.CustomRole object.
func LateInitializeSpec(spec *v1alpha1.CustomRoleParameters, in *iampmv1.CustomRole) error { // nolint:gocyclo
	if spec.AccountID == "" {
		spec.AccountID = reference.FromPtrValue(in.AccountID)
	}
	if spec.Actions == nil {
		spec.Actions = in.Actions
	}
	if spec.Description == nil {
		spec.Description = in.Description
	}
	if spec.DisplayName == "" {
		spec.DisplayName = reference.FromPtrValue(in.DisplayName)
	}
	return nil
}

// GenerateCreateCustomRoleOptions produces CustomRoleOptions object from CustomRoleParameters object.
func GenerateCreateCustomRoleOptions(in v1alpha1.CustomRoleParameters, o *iampmv1.CreateRoleOptions) error {
	o.AccountID = reference.ToPtrValue(in.AccountID)
	o.Actions = in.Actions
	o.Description = in.Description
	o.DisplayName = reference.ToPtrValue(in.DisplayName)
	o.Name = reference.ToPtrValue(in.Name)
	o.ServiceName = reference.ToPtrValue(in.ServiceName)
	return nil
}

// GenerateUpdateCustomRoleOptions produces UpdateCustomRoleOptions object from CustomRole object.
func GenerateUpdateCustomRoleOptions(id, eTag string, in v1alpha1.CustomRoleParameters, o *iampmv1.UpdateRoleOptions) error {
	o.Actions = in.Actions
	o.Description = in.Description
	o.DisplayName = reference.ToPtrValue(in.DisplayName)
	o.RoleID = reference.ToPtrValue(id)
	o.SetIfMatch(eTag)
	return nil
}

// GenerateObservation produces CustomRoleObservation object from *iampmv1.CustomRole object.
func GenerateObservation(in *iampmv1.CustomRole) (v1alpha1.CustomRoleObservation, error) {
	o := v1alpha1.CustomRoleObservation{
		ID:               reference.FromPtrValue(in.ID),
		CreatedAt:        GenerateMetaV1Time(in.CreatedAt),
		LastModifiedAt:   GenerateMetaV1Time(in.LastModifiedAt),
		CreatedByID:      reference.FromPtrValue(in.CreatedByID),
		LastModifiedByID: reference.FromPtrValue(in.LastModifiedByID),
		CRN:              reference.FromPtrValue(in.CRN),
		Href:             reference.FromPtrValue(in.Href),
	}
	return o, nil
}

// GenerateMetaV1Time converts strfmt.DateTime to metav1.Time
// TODO - extract this to parent `clients` package
func GenerateMetaV1Time(t *strfmt.DateTime) *metav1.Time {
	if t == nil {
		return nil
	}
	tx := metav1.NewTime(time.Time(*t))
	return &tx
}

// IsUpToDate checks whether current state is up-to-date compared to the given
// set of parameters.
func IsUpToDate(in *v1alpha1.CustomRoleParameters, observed *iampmv1.CustomRole, l logging.Logger) (bool, error) {
	desired := in.DeepCopy()
	actual, err := GenerateCustomRoleParameters(observed)
	if err != nil {
		return false, err
	}

	l.Info(cmp.Diff(desired, actual, cmpopts.IgnoreTypes(&runtimev1alpha1.Reference{}, &runtimev1alpha1.Selector{}, []runtimev1alpha1.Reference{})))

	return cmp.Equal(desired, actual, cmpopts.EquateEmpty(),
		cmpopts.IgnoreFields(v1alpha1.CustomRoleParameters{}),
		cmpopts.IgnoreTypes(&runtimev1alpha1.Reference{}, &runtimev1alpha1.Selector{}, []runtimev1alpha1.Reference{})), nil
}

// GenerateCustomRoleParameters generates service instance parameters from resource instance
func GenerateCustomRoleParameters(in *iampmv1.CustomRole) (*v1alpha1.CustomRoleParameters, error) {
	o := &v1alpha1.CustomRoleParameters{
		DisplayName: reference.FromPtrValue(in.DisplayName),
		Actions:     in.Actions,
		Name:        reference.FromPtrValue(in.Name),
		AccountID:   reference.FromPtrValue(in.AccountID),
		ServiceName: reference.FromPtrValue(in.ServiceName),
		Description: in.Description,
	}
	return o, nil
}
