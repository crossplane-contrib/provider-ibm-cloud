package accessgrouprule

import (
	"sort"
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
	// StateActive represents an access group rule in a running, available, and ready state
	StateActive = "active"
	// StateInactive represents an inactive access group rule
	StateInactive = "inactive"
)

// LateInitializeSpec fills optional and unassigned fields with the values in *iamagv2.Rule object.
func LateInitializeSpec(spec *v1alpha1.AccessGroupRuleParameters, in *iamagv2.Rule) error { // nolint:gocyclo
	return nil
}

// GenerateAddAccessGroupRuleOptions produces AddAccessGroupRuleOptions object from AccessGroupRuleParameters object.
func GenerateAddAccessGroupRuleOptions(in v1alpha1.AccessGroupRuleParameters, o *iamagv2.AddAccessGroupRuleOptions) error {
	o.AccessGroupID = in.AccessGroupID
	o.TransactionID = in.TransactionID
	o.Conditions = GenerateSDKRuleConditions(in.Conditions)
	o.Expiration = ibmc.Int64Ptr(in.Expiration)
	o.Name = reference.ToPtrValue(in.Name)
	o.RealmName = reference.ToPtrValue(in.RealmName)
	return nil
}

// GenerateReplaceAccessGroupRuleOptions produces ReplaceAccessGroupRuleOptions object from AccessGroupRuleParameters object.
func GenerateReplaceAccessGroupRuleOptions(id, eTag string, in v1alpha1.AccessGroupRuleParameters, o *iamagv2.ReplaceAccessGroupRuleOptions) error {
	o.RuleID = reference.ToPtrValue(id)
	o.AccessGroupID = in.AccessGroupID
	o.TransactionID = in.TransactionID
	o.Conditions = GenerateSDKRuleConditions(in.Conditions)
	o.Expiration = ibmc.Int64Ptr(in.Expiration)
	o.Name = reference.ToPtrValue(in.Name)
	o.RealmName = reference.ToPtrValue(in.RealmName)
	o.IfMatch = reference.ToPtrValue(eTag)
	return nil
}

// GenerateSDKRuleConditions -
func GenerateSDKRuleConditions(in []v1alpha1.RuleCondition) []iamagv2.RuleConditions {
	o := []iamagv2.RuleConditions{}
	for _, c := range in {
		item := iamagv2.RuleConditions{
			Claim:    reference.ToPtrValue(c.Claim),
			Operator: reference.ToPtrValue(c.Operator),
			Value:    reference.ToPtrValue(c.Value),
		}
		o = append(o, item)
	}
	return o
}

// GenerateCRRuleConditions -
func GenerateCRRuleConditions(in []iamagv2.RuleConditions) []v1alpha1.RuleCondition {
	o := []v1alpha1.RuleCondition{}
	for _, c := range in {
		item := v1alpha1.RuleCondition{
			Claim:    reference.FromPtrValue(c.Claim),
			Operator: reference.FromPtrValue(c.Operator),
			Value:    reference.FromPtrValue(c.Value),
		}
		o = append(o, item)
	}
	return o
}

// GenerateObservation produces AccessGroupRuleObservation object from *iamagv2.Group object.
func GenerateObservation(in *iamagv2.Rule) (v1alpha1.AccessGroupRuleObservation, error) {
	o := v1alpha1.AccessGroupRuleObservation{
		ID:               reference.FromPtrValue(in.ID),
		AccountID:        reference.FromPtrValue(in.AccountID),
		CreatedAt:        GenerateMetaV1Time(in.CreatedAt),
		CreatedByID:      reference.FromPtrValue(in.CreatedByID),
		LastModifiedAt:   GenerateMetaV1Time(in.LastModifiedAt),
		LastModifiedByID: reference.FromPtrValue(in.LastModifiedByID),
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
func IsUpToDate(in *v1alpha1.AccessGroupRuleParameters, observed *iamagv2.Rule, l logging.Logger) (bool, error) {
	desired := in.DeepCopy()
	actual, err := GenerateAccessGroupRuleParameters(observed)
	if err != nil {
		return false, err
	}
	sort.Slice(desired.Conditions, func(i, j int) bool {
		return desired.Conditions[i].Claim < desired.Conditions[j].Claim
	})
	sort.Slice(actual.Conditions, func(i, j int) bool {
		return actual.Conditions[i].Claim < actual.Conditions[j].Claim
	})

	l.Info(cmp.Diff(desired, actual, cmpopts.EquateEmpty(),
		cmpopts.IgnoreFields(v1alpha1.AccessGroupRuleParameters{}, "AccessGroupID"),
		cmpopts.IgnoreFields(v1alpha1.AccessGroupRuleParameters{}, "TransactionID"),
		cmpopts.IgnoreTypes(&runtimev1alpha1.Reference{}, &runtimev1alpha1.Selector{}, []runtimev1alpha1.Reference{})))

	return cmp.Equal(desired, actual, cmpopts.EquateEmpty(),
		cmpopts.IgnoreFields(v1alpha1.AccessGroupRuleParameters{}, "AccessGroupID"),
		cmpopts.IgnoreFields(v1alpha1.AccessGroupRuleParameters{}, "TransactionID"),
		cmpopts.IgnoreTypes(&runtimev1alpha1.Reference{}, &runtimev1alpha1.Selector{}, []runtimev1alpha1.Reference{})), nil
}

// GenerateAccessGroupRuleParameters generates service instance parameters from resource instance
func GenerateAccessGroupRuleParameters(in *iamagv2.Rule) (*v1alpha1.AccessGroupRuleParameters, error) {
	o := &v1alpha1.AccessGroupRuleParameters{
		AccessGroupID: in.AccessGroupID,
		Expiration:    ibmc.Int64Value(in.Expiration),
		RealmName:     reference.FromPtrValue(in.RealmName),
		Conditions:    GenerateCRRuleConditions(in.Conditions),
		Name:          reference.FromPtrValue(in.Name),
	}
	return o, nil
}

// GenerateCRAddGroupMembersRequestMembersItem -
func GenerateCRAddGroupMembersRequestMembersItem(in *iamagv2.GroupMembersList) []v1alpha1.AddGroupMembersRequestMembersItem {
	o := []v1alpha1.AddGroupMembersRequestMembersItem{}
	if in == nil {
		return o
	}
	for _, m := range in.Members {
		item := v1alpha1.AddGroupMembersRequestMembersItem{
			IamID: reference.FromPtrValue(m.IamID),
			Type:  reference.FromPtrValue(m.Type),
		}
		o = append(o, item)
	}
	return o
}
