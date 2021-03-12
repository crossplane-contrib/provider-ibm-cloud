package accessgroup

import (
	"time"

	"github.com/go-openapi/strfmt"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/reference"

	iamagv2 "github.com/IBM/platform-services-go-sdk/iamaccessgroupsv2"

	"github.com/crossplane-contrib/provider-ibm-cloud/apis/iamaccessgroupsv2/v1alpha1"
	ibmc "github.com/crossplane-contrib/provider-ibm-cloud/pkg/clients"
)

const (
	// StateActive represents an access group in a running, available, and ready state
	StateActive = "active"
	// StateInactive represents an inactive access group
	StateInactive = "inactive"
)

// LateInitializeSpec fills optional and unassigned fields with the values in *iamagv2.AccessGroup object.
func LateInitializeSpec(spec *v1alpha1.AccessGroupParameters, in *iamagv2.Group) error { // nolint:gocyclo
	if spec.Description == nil {
		spec.Description = in.Description
	}

	return nil
}

// GenerateCreateAccessGroupOptions produces AccessGroupOptions object from AccessGroupParameters object.
func GenerateCreateAccessGroupOptions(in v1alpha1.AccessGroupParameters, o *iamagv2.CreateAccessGroupOptions) error {
	o.AccountID = reference.ToPtrValue(in.AccountID)
	o.Description = in.Description
	o.Name = reference.ToPtrValue(in.Name)
	o.TransactionID = in.TransactionID
	return nil
}

// GenerateUpdateAccessGroupOptions produces UpdateAccessGroupOptions object from AccessGroup object.
func GenerateUpdateAccessGroupOptions(id, eTag string, in v1alpha1.AccessGroupParameters, o *iamagv2.UpdateAccessGroupOptions) error {
	o.AccessGroupID = reference.ToPtrValue(id)
	o.Description = in.Description
	o.Name = reference.ToPtrValue(in.Name)
	o.TransactionID = in.TransactionID
	o.SetIfMatch(eTag)
	return nil
}

// GenerateObservation produces AccessGroupObservation object from *iamagv2.Group object.
func GenerateObservation(in *iamagv2.Group) (v1alpha1.AccessGroupObservation, error) {
	o := v1alpha1.AccessGroupObservation{
		ID:               reference.FromPtrValue(in.ID),
		CreatedAt:        GenerateMetaV1Time(in.CreatedAt),
		LastModifiedAt:   GenerateMetaV1Time(in.LastModifiedAt),
		CreatedByID:      reference.FromPtrValue(in.CreatedByID),
		LastModifiedByID: reference.FromPtrValue(in.LastModifiedByID),
		Href:             reference.FromPtrValue(in.Href),
		IsFederated:      ibmc.BoolValue(in.IsFederated),
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
func IsUpToDate(in *v1alpha1.AccessGroupParameters, observed *iamagv2.Group, l logging.Logger) (bool, error) {
	desired := in.DeepCopy()
	actual, err := GenerateAccessGroupParameters(observed)
	if err != nil {
		return false, err
	}

	l.Info(cmp.Diff(desired, actual, cmpopts.IgnoreTypes(&runtimev1alpha1.Reference{}, &runtimev1alpha1.Selector{}, []runtimev1alpha1.Reference{})))

	return cmp.Equal(desired, actual, cmpopts.EquateEmpty(),
		cmpopts.IgnoreFields(v1alpha1.AccessGroupParameters{}, "TransactionID"),
		cmpopts.IgnoreTypes(&runtimev1alpha1.Reference{}, &runtimev1alpha1.Selector{}, []runtimev1alpha1.Reference{})), nil
}

// GenerateAccessGroupParameters generates service instance parameters from resource instance
func GenerateAccessGroupParameters(in *iamagv2.Group) (*v1alpha1.AccessGroupParameters, error) {
	o := &v1alpha1.AccessGroupParameters{
		Name:        reference.FromPtrValue(in.Name),
		AccountID:   reference.FromPtrValue(in.AccountID),
		Description: in.Description,

		// TODO - TransactionID has no match in the Group object - need to ignore it in IsUpToDate
	}
	return o, nil
}
