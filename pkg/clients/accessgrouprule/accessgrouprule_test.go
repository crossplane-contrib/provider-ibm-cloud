package accessgrouprule

import (
	"testing"

	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/reference"
	"github.com/go-openapi/strfmt"
	"github.com/google/go-cmp/cmp"

	iamagv2 "github.com/IBM/platform-services-go-sdk/iamaccessgroupsv2"

	"github.com/crossplane-contrib/provider-ibm-cloud/apis/iamaccessgroupsv2/v1alpha1"
	ibmc "github.com/crossplane-contrib/provider-ibm-cloud/pkg/clients"
)

var (
	accountID     = "aa5a00334eaf9eb9339d2ab48f20d7ff"
	accessGroupID = "12345678-abcd-1a2b-a1b2-1234567890ab"
	ruleID        = "abcd-12345689-1a2b-a1b2-123456789000"
	ruleName      = "Manager group rule"
	realmName     = "https://idp.example.org/SAML2"
	createdAt, _  = strfmt.ParseDateTime("2020-10-31T02:33:06Z")
	transactionID = "12345-abcd-ef000-abac"
	createdByID   = "IBM-User-0001"
	expiration    = 24
	claim1        = "isManager"
	opEquals      = "EQUALS"
	claim2        = "isViewer"
	eTag          = "1-eb832c7ff8c8016a542974b9f880b55e"
)

func params(m ...func(*v1alpha1.AccessGroupRuleParameters)) *v1alpha1.AccessGroupRuleParameters {
	p := &v1alpha1.AccessGroupRuleParameters{
		Name:          ruleName,
		AccessGroupID: &accessGroupID,
		TransactionID: &transactionID,
		Expiration:    int64(expiration),
		RealmName:     "https://idp.example.org/SAML2",
		Conditions: []v1alpha1.RuleCondition{
			{
				Claim:    claim1,
				Operator: opEquals,
				Value:    "true",
			},
		},
	}

	for _, f := range m {
		f(p)
	}
	return p
}

func observation(m ...func(*v1alpha1.AccessGroupRuleObservation)) *v1alpha1.AccessGroupRuleObservation {
	o := &v1alpha1.AccessGroupRuleObservation{
		ID:               ruleID,
		AccountID:        accountID,
		CreatedAt:        GenerateMetaV1Time(&createdAt),
		CreatedByID:      createdByID,
		LastModifiedAt:   GenerateMetaV1Time(&createdAt),
		LastModifiedByID: createdByID,
	}

	for _, f := range m {
		f(o)
	}
	return o
}

func instanceCreateOpts(m ...func(*iamagv2.AddAccessGroupRuleOptions)) *iamagv2.AddAccessGroupRuleOptions {
	i := &iamagv2.AddAccessGroupRuleOptions{
		Name:          &ruleName,
		AccessGroupID: &accessGroupID,
		Expiration:    ibmc.Int64Ptr(int64(expiration)),
		TransactionID: &transactionID,
		RealmName:     &realmName,
		Conditions: []iamagv2.RuleConditions{
			{
				Claim:    &claim1,
				Operator: &opEquals,
				Value:    reference.ToPtrValue("true"),
			},
		},
	}
	for _, f := range m {
		f(i)
	}
	return i
}

func instanceReplaceOpts(m ...func(*iamagv2.ReplaceAccessGroupRuleOptions)) *iamagv2.ReplaceAccessGroupRuleOptions {
	i := &iamagv2.ReplaceAccessGroupRuleOptions{
		RuleID:        &ruleID,
		Name:          &ruleName,
		IfMatch:       &eTag,
		AccessGroupID: &accessGroupID,
		Expiration:    ibmc.Int64Ptr(int64(expiration)),
		TransactionID: &transactionID,
		RealmName:     &realmName,
		Conditions: []iamagv2.RuleConditions{
			{
				Claim:    &claim1,
				Operator: &opEquals,
				Value:    reference.ToPtrValue("true"),
			},
		},
	}
	for _, f := range m {
		f(i)
	}
	return i
}

func instance(m ...func(*iamagv2.Rule)) *iamagv2.Rule {
	i := &iamagv2.Rule{
		ID:            &ruleID,
		Name:          &ruleName,
		Expiration:    ibmc.Int64Ptr(int64(expiration)),
		RealmName:     &realmName,
		AccessGroupID: &accessGroupID,
		AccountID:     &accountID,
		Conditions: []iamagv2.RuleConditions{
			{
				Claim:    &claim1,
				Operator: &opEquals,
				Value:    reference.ToPtrValue("true"),
			},
		},
		CreatedAt:        &createdAt,
		CreatedByID:      &createdByID,
		LastModifiedAt:   &createdAt,
		LastModifiedByID: &createdByID,
	}
	for _, f := range m {
		f(i)
	}
	return i
}

func TestGenerateCreateAccessGroupRuleOptions(t *testing.T) {
	type args struct {
		params v1alpha1.AccessGroupRuleParameters
	}
	type want struct {
		instance *iamagv2.AddAccessGroupRuleOptions
	}
	cases := map[string]struct {
		args args
		want want
	}{
		"FullConversion": {
			args: args{params: *params()},
			want: want{instance: instanceCreateOpts()},
		},
		"MissingFields": {
			args: args{
				params: *params(func(p *v1alpha1.AccessGroupRuleParameters) {
					p.TransactionID = nil
				})},
			want: want{instance: instanceCreateOpts(func(p *iamagv2.AddAccessGroupRuleOptions) {
				p.TransactionID = nil
			})},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {

			r := &iamagv2.AddAccessGroupRuleOptions{}
			GenerateAddAccessGroupRuleOptions(tc.args.params, r)
			if diff := cmp.Diff(tc.want.instance, r); diff != "" {
				t.Errorf("GenerateCreateAccessGroupRuleOptions(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestGenerateReplaceAccessGroupRuleOptions(t *testing.T) {
	type args struct {
		params v1alpha1.AccessGroupRuleParameters
		id     string
		eTag   string
	}
	type want struct {
		instance *iamagv2.ReplaceAccessGroupRuleOptions
	}
	cases := map[string]struct {
		args args
		want want
	}{
		"FullConversion": {
			args: args{params: *params(),
				id:   ruleID,
				eTag: eTag,
			},
			want: want{instance: instanceReplaceOpts()},
		},
		"MissingFields": {
			args: args{
				params: *params(func(p *v1alpha1.AccessGroupRuleParameters) {
					p.TransactionID = nil
				}),
				id:   ruleID,
				eTag: eTag,
			},
			want: want{instance: instanceReplaceOpts(func(p *iamagv2.ReplaceAccessGroupRuleOptions) {
				p.TransactionID = nil
			})},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			r := &iamagv2.ReplaceAccessGroupRuleOptions{}
			GenerateReplaceAccessGroupRuleOptions(tc.args.id, tc.args.eTag, tc.args.params, r)
			if diff := cmp.Diff(tc.want.instance, r); diff != "" {
				t.Errorf("GenerateReplaceAccessGroupRuleOptions(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestLateInitializeSpecs(t *testing.T) {
	type args struct {
		instance *iamagv2.Rule
		params   *v1alpha1.AccessGroupRuleParameters
	}
	type want struct {
		params *v1alpha1.AccessGroupRuleParameters
	}
	cases := map[string]struct {
		args args
		want want
	}{
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
		instance *iamagv2.Rule
	}
	type want struct {
		obs v1alpha1.AccessGroupRuleObservation
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

func TestIsUpToDate(t *testing.T) {
	type args struct {
		params   *v1alpha1.AccessGroupRuleParameters
		instance *iamagv2.Rule
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
				params: params(func(crp *v1alpha1.AccessGroupRuleParameters) {
					crp.Conditions = append(crp.Conditions, v1alpha1.RuleCondition{
						Claim:    claim2,
						Operator: opEquals,
						Value:    "true",
					})
				}),
				instance: instance(),
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
