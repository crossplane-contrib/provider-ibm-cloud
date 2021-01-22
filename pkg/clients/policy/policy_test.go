package policy

import (
	"testing"

	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/go-openapi/strfmt"
	"github.com/google/go-cmp/cmp"

	iampmv1 "github.com/IBM/platform-services-go-sdk/iampolicymanagementv1"

	"github.com/crossplane-contrib/provider-ibm-cloud/apis/iampolicymanagementv1/v1alpha1"
	ibmc "github.com/crossplane-contrib/provider-ibm-cloud/pkg/clients"
)

var (
	policyTypeAccess     = "access"
	policyTypeAuth       = "authorization"
	policyAttributeName  = "iam_id"
	policyAttributeValue = "IBMid-123453user"
	createdByID          = "IBMid-123453user"
	roleID               = "crn:v1:bluemix:public:iam::::role:Editor"
	resAttr1Name         = "accountId"
	resAttr1Value        = "my-account-id"
	resAttr2Name         = "serviceName"
	resAttr2Value        = "cos"
	resAttr3Name         = "resource"
	resAttr3Value        = "mycos"
	resAttr3Operator     = "stringEquals"
	policyDescription    = "this is my policy 1"
	policyID             = "12345678-abcd-1a2b-a1b2-1234567890ab"
	createdAt, _         = strfmt.ParseDateTime("2020-10-31T02:33:06Z")
	lastModifiedAt, _    = strfmt.ParseDateTime("2020-10-31T03:33:06Z")
	hRef                 = "https://iam.cloud.ibm.com/v1/policies/12345678-abcd-1a2b-a1b2-1234567890ab"
	eTag                 = "1-eb832c7ff8c8016a542974b9f880b55e"
)

func params(m ...func(*v1alpha1.PolicyParameters)) *v1alpha1.PolicyParameters {
	p := &v1alpha1.PolicyParameters{
		Type: policyTypeAccess,
		Subjects: []v1alpha1.PolicySubject{
			{
				Attributes: []v1alpha1.SubjectAttribute{
					{
						Name:  &policyAttributeName,
						Value: &policyAttributeValue,
					},
				},
			},
		},
		Roles: []v1alpha1.PolicyRole{
			{
				RoleID: roleID,
			},
		},
		Resources: []v1alpha1.PolicyResource{
			{
				Attributes: []v1alpha1.ResourceAttribute{
					{
						Name:  &resAttr1Name,
						Value: &resAttr1Value,
					},
					{
						Name:  &resAttr2Name,
						Value: &resAttr2Value,
					},
					{
						Name:     &resAttr3Name,
						Value:    &resAttr3Value,
						Operator: &resAttr3Operator,
					},
				},
			},
		},
		Description: &policyDescription,
	}

	for _, f := range m {
		f(p)
	}
	return p
}

func observation(m ...func(*v1alpha1.PolicyObservation)) *v1alpha1.PolicyObservation {
	o := &v1alpha1.PolicyObservation{
		ID:               policyID,
		CreatedAt:        ibmc.DateTimeToMetaV1Time(&createdAt),
		LastModifiedAt:   ibmc.DateTimeToMetaV1Time(&lastModifiedAt),
		CreatedByID:      policyAttributeValue,
		LastModifiedByID: policyAttributeValue,
		Href:             hRef,
	}

	for _, f := range m {
		f(o)
	}
	return o
}

func instance(m ...func(*iampmv1.Policy)) *iampmv1.Policy {
	i := &iampmv1.Policy{
		ID:          &policyID,
		Type:        &policyTypeAccess,
		Description: &policyDescription,
		Subjects: []iampmv1.PolicySubject{
			{
				Attributes: []iampmv1.SubjectAttribute{
					{
						Name:  &policyAttributeName,
						Value: &policyAttributeValue,
					},
				},
			},
		},
		Roles: []iampmv1.PolicyRole{
			{
				RoleID: &roleID,
			},
		},
		Resources: []iampmv1.PolicyResource{
			{
				Attributes: []iampmv1.ResourceAttribute{
					{
						Name:  &resAttr1Name,
						Value: &resAttr1Value,
					},
					{
						Name:  &resAttr2Name,
						Value: &resAttr2Value,
					},
					{
						Name:     &resAttr3Name,
						Value:    &resAttr3Value,
						Operator: &resAttr3Operator,
					},
				},
			},
		},
		Href:             &hRef,
		CreatedAt:        &createdAt,
		CreatedByID:      &createdByID,
		LastModifiedAt:   &lastModifiedAt,
		LastModifiedByID: &createdByID,
	}

	for _, f := range m {
		f(i)
	}
	return i
}

func instanceOpts(m ...func(*iampmv1.CreatePolicyOptions)) *iampmv1.CreatePolicyOptions {
	i := &iampmv1.CreatePolicyOptions{
		Type: &policyTypeAccess,
		Subjects: []iampmv1.PolicySubject{
			{
				Attributes: []iampmv1.SubjectAttribute{
					{
						Name:  &policyAttributeName,
						Value: &policyAttributeValue,
					},
				},
			},
		},
		Roles: []iampmv1.PolicyRole{
			{
				RoleID: &roleID,
			},
		},
		Resources: []iampmv1.PolicyResource{
			{
				Attributes: []iampmv1.ResourceAttribute{
					{
						Name:  &resAttr1Name,
						Value: &resAttr1Value,
					},
					{
						Name:  &resAttr2Name,
						Value: &resAttr2Value,
					},
					{
						Name:     &resAttr3Name,
						Value:    &resAttr3Value,
						Operator: &resAttr3Operator,
					},
				},
			},
		},
		Description: &policyDescription,
	}
	for _, f := range m {
		f(i)
	}
	return i
}

func instanceUpdOpts(m ...func(*iampmv1.UpdatePolicyOptions)) *iampmv1.UpdatePolicyOptions {
	i := &iampmv1.UpdatePolicyOptions{
		PolicyID: &policyID,
		Type:     &policyTypeAccess,
		Subjects: []iampmv1.PolicySubject{
			{
				Attributes: []iampmv1.SubjectAttribute{
					{
						Name:  &policyAttributeName,
						Value: &policyAttributeValue,
					},
				},
			},
		},
		Roles: []iampmv1.PolicyRole{
			{
				RoleID: &roleID,
			},
		},
		Resources: []iampmv1.PolicyResource{
			{
				Attributes: []iampmv1.ResourceAttribute{
					{
						Name:  &resAttr1Name,
						Value: &resAttr1Value,
					},
					{
						Name:  &resAttr2Name,
						Value: &resAttr2Value,
					},
					{
						Name:     &resAttr3Name,
						Value:    &resAttr3Value,
						Operator: &resAttr3Operator,
					},
				},
			},
		},
		Description: &policyDescription,
	}

	for _, f := range m {
		f(i)
	}
	return i
}

func TestGenerateCreatePolicyOptions(t *testing.T) {
	type args struct {
		params v1alpha1.PolicyParameters
	}
	type want struct {
		instance *iampmv1.CreatePolicyOptions
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
				params: *params(func(p *v1alpha1.PolicyParameters) {
					p.Type = ""
				})},
			want: want{instance: instanceOpts(func(p *iampmv1.CreatePolicyOptions) {
				p.Type = nil
			})},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {

			r := &iampmv1.CreatePolicyOptions{}
			GenerateCreatePolicyOptions(tc.args.params, r)
			if diff := cmp.Diff(tc.want.instance, r); diff != "" {
				t.Errorf("GenerateCreatePolicyOptions(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestGenerateUpdatePolicyOptions(t *testing.T) {
	type args struct {
		id     string
		etag   string
		params v1alpha1.PolicyParameters
	}
	type want struct {
		instance *iampmv1.UpdatePolicyOptions
	}
	cases := map[string]struct {
		args args
		want want
	}{
		"FullConversion": {
			args: args{id: policyID, etag: eTag, params: *params()},
			want: want{instance: instanceUpdOpts(func(upo *iampmv1.UpdatePolicyOptions) { upo.SetIfMatch(eTag) })},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			r := &iampmv1.UpdatePolicyOptions{}
			GenerateUpdatePolicyOptions(tc.args.id, tc.args.etag, tc.args.params, r)
			if diff := cmp.Diff(tc.want.instance, r); diff != "" {
				t.Errorf("GenerateUpdatePolicyOptions(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestLateInitializeSpecs(t *testing.T) {
	type args struct {
		instance *iampmv1.Policy
		params   *v1alpha1.PolicyParameters
	}
	type want struct {
		params *v1alpha1.PolicyParameters
	}
	cases := map[string]struct {
		args args
		want want
	}{
		"SomeFields": {
			args: args{
				params: params(func(p *v1alpha1.PolicyParameters) {
					p.Roles = nil
				}),
				instance: instance(func(p *iampmv1.Policy) {
				}),
			},
			want: want{
				params: params(func(p *v1alpha1.PolicyParameters) {
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

func TestGenerateObservation(t *testing.T) {
	type args struct {
		instance *iampmv1.Policy
	}
	type want struct {
		obs v1alpha1.PolicyObservation
	}
	cases := map[string]struct {
		args args
		want want
	}{
		"FullConversion": {
			args: args{
				instance: instance(func(p *iampmv1.Policy) {
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

func TestIsUpToDate(t *testing.T) {
	type args struct {
		params   *v1alpha1.PolicyParameters
		instance *iampmv1.Policy
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
				params: params(),
				instance: instance(func(i *iampmv1.Policy) {
					i.Type = &policyTypeAuth
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
