package policy

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

// LateInitializeSpec fills optional and unassigned fields with the values in *iampmv1.Policy object.
func LateInitializeSpec(spec *v1alpha1.PolicyParameters, in *iampmv1.Policy) error { // nolint:gocyclo
	if spec.Description == nil {
		spec.Description = in.Description
	}
	if spec.Resources == nil {
		spec.Resources = GenerateCRResources(in.Resources)
	}
	for i, r := range spec.Resources {
		for j, attr := range r.Attributes {
			if attr.Name == nil {
				spec.Resources[i].Attributes[j].Name = in.Resources[i].Attributes[j].Name
			}
			if attr.Value == nil {
				spec.Resources[i].Attributes[j].Value = in.Resources[i].Attributes[j].Value
			}
			if attr.Operator == nil {
				spec.Resources[i].Attributes[j].Operator = in.Resources[i].Attributes[j].Operator
			}
		}
	}
	if spec.Roles == nil {
		spec.Roles = GenerateCRRoles(in.Roles)
	}
	for i, r := range spec.Roles {
		if r.RoleID == "" {
			spec.Roles[i].RoleID = reference.FromPtrValue(in.Roles[i].RoleID)
		}
	}
	if spec.Subjects == nil {
		spec.Subjects = GenerateCRSubjects(in.Subjects)
	}
	for i, s := range spec.Subjects {
		for j, attr := range s.Attributes {
			if attr.Name == nil {
				spec.Resources[i].Attributes[j].Name = in.Resources[i].Attributes[j].Name
			}
			if attr.Value == nil {
				spec.Resources[i].Attributes[j].Value = in.Resources[i].Attributes[j].Value
			}
		}
	}
	if spec.Type == "" {
		spec.Type = reference.FromPtrValue(in.Type)
	}
	return nil
}

// GenerateCreatePolicyOptions produces PolicyOptions object from PolicyParameters object.
func GenerateCreatePolicyOptions(in v1alpha1.PolicyParameters, o *iampmv1.CreatePolicyOptions) error {
	o.Description = in.Description
	o.Resources = GenerateSDKResources(in.Resources)
	o.Roles = GenerateSDKRoles(in.Roles)
	o.Subjects = GenerateSDKSubjects(in.Subjects)
	o.Type = reference.ToPtrValue(in.Type)
	return nil
}

// GenerateUpdatePolicyOptions produces UpdatePolicyOptions object from Policy object.
func GenerateUpdatePolicyOptions(id, eTag string, in v1alpha1.PolicyParameters, o *iampmv1.UpdatePolicyOptions) error {
	o.Description = in.Description
	o.Resources = GenerateSDKResources(in.Resources)
	o.Roles = GenerateSDKRoles(in.Roles)
	o.Subjects = GenerateSDKSubjects(in.Subjects)
	o.Type = reference.ToPtrValue(in.Type)
	o.PolicyID = &id
	o.SetIfMatch(eTag)
	return nil
}

// GenerateObservation produces PolicyObservation object from *iampmv1.Policy object.
func GenerateObservation(in *iampmv1.Policy) (v1alpha1.PolicyObservation, error) {
	o := v1alpha1.PolicyObservation{
		ID:               reference.FromPtrValue(in.ID),
		CreatedAt:        GenerateMetaV1Time(in.CreatedAt),
		LastModifiedAt:   GenerateMetaV1Time(in.LastModifiedAt),
		CreatedByID:      reference.FromPtrValue(in.CreatedByID),
		LastModifiedByID: reference.FromPtrValue(in.LastModifiedByID),
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
func IsUpToDate(in *v1alpha1.PolicyParameters, observed *iampmv1.Policy, l logging.Logger) (bool, error) {
	desired := in.DeepCopy()
	actual, err := GeneratePolicyParameters(observed)
	if err != nil {
		return false, err
	}

	l.Info(cmp.Diff(desired, actual, cmpopts.IgnoreTypes(&runtimev1alpha1.Reference{}, &runtimev1alpha1.Selector{}, []runtimev1alpha1.Reference{})))

	return cmp.Equal(desired, actual, cmpopts.EquateEmpty(),
		cmpopts.IgnoreFields(v1alpha1.PolicyParameters{}),
		cmpopts.IgnoreTypes(&runtimev1alpha1.Reference{}, &runtimev1alpha1.Selector{}, []runtimev1alpha1.Reference{})), nil
}

// GeneratePolicyParameters generates service instance parameters from resource instance
func GeneratePolicyParameters(in *iampmv1.Policy) (*v1alpha1.PolicyParameters, error) {
	o := &v1alpha1.PolicyParameters{
		Type:        reference.FromPtrValue(in.Type),
		Subjects:    GenerateCRSubjects(in.Subjects),
		Roles:       GenerateCRRoles(in.Roles),
		Resources:   GenerateCRResources(in.Resources),
		Description: in.Description,
	}
	return o, nil
}

// GenerateCRResources -
func GenerateCRResources(in []iampmv1.PolicyResource) []v1alpha1.PolicyResource {
	o := []v1alpha1.PolicyResource{}
	for _, res := range in {
		item := v1alpha1.PolicyResource{
			Attributes: GenerateCRResourceAttributes(res.Attributes),
		}
		o = append(o, item)
	}
	return o
}

// GenerateCRResourceAttributes -
func GenerateCRResourceAttributes(in []iampmv1.ResourceAttribute) []v1alpha1.ResourceAttribute {
	o := []v1alpha1.ResourceAttribute{}
	for _, attr := range in {
		item := v1alpha1.ResourceAttribute{
			Name:     attr.Name,
			Value:    attr.Value,
			Operator: attr.Operator,
		}
		o = append(o, item)
	}
	return o
}

// GenerateCRRoles -
func GenerateCRRoles(in []iampmv1.PolicyRole) []v1alpha1.PolicyRole {
	o := []v1alpha1.PolicyRole{}
	for _, pol := range in {
		item := v1alpha1.PolicyRole{
			RoleID: reference.FromPtrValue(pol.RoleID),
		}
		o = append(o, item)
	}
	return o
}

// GenerateCRSubjects -
func GenerateCRSubjects(in []iampmv1.PolicySubject) []v1alpha1.PolicySubject {
	o := []v1alpha1.PolicySubject{}
	for _, pol := range in {
		item := v1alpha1.PolicySubject{
			Attributes: GenerateCRSubjectAttributes(pol.Attributes),
		}
		o = append(o, item)
	}
	return o
}

// GenerateCRSubjectAttributes -
func GenerateCRSubjectAttributes(in []iampmv1.SubjectAttribute) []v1alpha1.SubjectAttribute {
	o := []v1alpha1.SubjectAttribute{}
	for _, attr := range in {
		item := v1alpha1.SubjectAttribute{
			Name:  attr.Name,
			Value: attr.Value,
		}
		o = append(o, item)
	}
	return o
}

// GenerateSDKResources -
func GenerateSDKResources(in []v1alpha1.PolicyResource) []iampmv1.PolicyResource {
	o := []iampmv1.PolicyResource{}
	for _, res := range in {
		item := iampmv1.PolicyResource{
			Attributes: GenerateSDKResourceAttributes(res.Attributes),
		}
		o = append(o, item)
	}
	return o
}

// GenerateSDKResourceAttributes -
func GenerateSDKResourceAttributes(in []v1alpha1.ResourceAttribute) []iampmv1.ResourceAttribute {
	o := []iampmv1.ResourceAttribute{}
	for _, attr := range in {
		item := iampmv1.ResourceAttribute{
			Name:     attr.Name,
			Value:    attr.Value,
			Operator: attr.Operator,
		}
		o = append(o, item)
	}
	return o
}

// GenerateSDKRoles -
func GenerateSDKRoles(in []v1alpha1.PolicyRole) []iampmv1.PolicyRole {
	o := []iampmv1.PolicyRole{}
	for _, pol := range in {
		item := iampmv1.PolicyRole{
			RoleID: &pol.RoleID,
		}
		o = append(o, item)
	}
	return o
}

// GenerateSDKSubjects -
func GenerateSDKSubjects(in []v1alpha1.PolicySubject) []iampmv1.PolicySubject {
	o := []iampmv1.PolicySubject{}
	for _, pol := range in {
		item := iampmv1.PolicySubject{
			Attributes: GenerateSDKSubjectAttributes(pol.Attributes),
		}
		o = append(o, item)
	}
	return o
}

// GenerateSDKSubjectAttributes -
func GenerateSDKSubjectAttributes(in []v1alpha1.SubjectAttribute) []iampmv1.SubjectAttribute {
	o := []iampmv1.SubjectAttribute{}
	for _, attr := range in {
		item := iampmv1.SubjectAttribute{
			Name:  attr.Name,
			Value: attr.Value,
		}
		o = append(o, item)
	}
	return o
}
