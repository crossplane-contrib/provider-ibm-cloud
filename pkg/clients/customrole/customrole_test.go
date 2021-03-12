package customrole

import (
	"testing"

	"github.com/go-openapi/strfmt"
	"github.com/google/go-cmp/cmp"

	"github.com/crossplane/crossplane-runtime/pkg/logging"

	iampmv1 "github.com/IBM/platform-services-go-sdk/iampolicymanagementv1"

	"github.com/crossplane-contrib/provider-ibm-cloud/apis/iampolicymanagementv1/v1alpha1"
	ibmc "github.com/crossplane-contrib/provider-ibm-cloud/pkg/clients"
)

var (
	roleName          = "myCustomRole"
	roleDescription   = "role for my service"
	accountID         = "aa5a00334eaf9eb9339d2ab48f20d7ff"
	displayName       = "MyCustomRole"
	serviceName       = "mypostgres"
	createdByID       = "IBMid-123453user"
	action1           = "iam.policy.create"
	action2           = "iam.policy.update"
	action3           = "iam.policy.delete"
	roleID            = "12345678-abcd-1a2b-a1b2-1234567890ab"
	createdAt, _      = strfmt.ParseDateTime("2020-10-31T02:33:06Z")
	lastModifiedAt, _ = strfmt.ParseDateTime("2020-10-31T03:33:06Z")
	hRef              = "https://iam.cloud.ibm.com/v1/roles/12345678-abcd-1a2b-a1b2-1234567890ab"
	eTag              = "1-eb832c7ff8c8016a542974b9f880b55e"
	crn               = "crn:v1:bluemix:public:iam::::role:" + roleName
)

func params(m ...func(*v1alpha1.CustomRoleParameters)) *v1alpha1.CustomRoleParameters {
	p := &v1alpha1.CustomRoleParameters{
		DisplayName: displayName,
		Actions:     []string{action1, action2},
		Name:        roleName,
		AccountID:   accountID,
		ServiceName: serviceName,
		Description: &roleDescription,
	}

	for _, f := range m {
		f(p)
	}
	return p
}

func observation(m ...func(*v1alpha1.CustomRoleObservation)) *v1alpha1.CustomRoleObservation {
	o := &v1alpha1.CustomRoleObservation{
		ID:               roleID,
		CreatedAt:        ibmc.DateTimeToMetaV1Time(&createdAt),
		LastModifiedAt:   ibmc.DateTimeToMetaV1Time(&lastModifiedAt),
		CreatedByID:      createdByID,
		LastModifiedByID: createdByID,
		CRN:              crn,
		Href:             hRef,
	}

	for _, f := range m {
		f(o)
	}
	return o
}

func instance(m ...func(*iampmv1.CustomRole)) *iampmv1.CustomRole {
	i := &iampmv1.CustomRole{
		ID:               &roleID,
		Href:             &hRef,
		CreatedAt:        &createdAt,
		CreatedByID:      &createdByID,
		LastModifiedAt:   &lastModifiedAt,
		LastModifiedByID: &createdByID,
		DisplayName:      &displayName,
		Description:      &roleDescription,
		Actions:          []string{action1, action2},
		CRN:              &crn,
		Name:             &roleName,
		AccountID:        &accountID,
		ServiceName:      &serviceName,
	}

	for _, f := range m {
		f(i)
	}
	return i
}

func instanceOpts(m ...func(*iampmv1.CreateRoleOptions)) *iampmv1.CreateRoleOptions {
	i := &iampmv1.CreateRoleOptions{
		DisplayName: &displayName,
		Actions:     []string{action1, action2},
		Name:        &roleName,
		AccountID:   &accountID,
		ServiceName: &serviceName,
		Description: &roleDescription,
	}
	for _, f := range m {
		f(i)
	}
	return i
}

func instanceUpdOpts(m ...func(*iampmv1.UpdateRoleOptions)) *iampmv1.UpdateRoleOptions {
	i := &iampmv1.UpdateRoleOptions{
		RoleID:      &roleID,
		IfMatch:     &eTag,
		DisplayName: &displayName,
		Description: &roleDescription,
		Actions:     []string{action1, action2, action3},
	}

	for _, f := range m {
		f(i)
	}
	return i
}

func TestGenerateCreateCustomRoleOptions(t *testing.T) {
	type args struct {
		params v1alpha1.CustomRoleParameters
	}
	type want struct {
		instance *iampmv1.CreateRoleOptions
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
				params: *params(func(p *v1alpha1.CustomRoleParameters) {
					p.Description = nil
				})},
			want: want{instance: instanceOpts(func(p *iampmv1.CreateRoleOptions) {
				p.Description = nil
			})},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {

			r := &iampmv1.CreateRoleOptions{}
			GenerateCreateCustomRoleOptions(tc.args.params, r)
			if diff := cmp.Diff(tc.want.instance, r); diff != "" {
				t.Errorf("GenerateCreateCustomRoleOptions(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestGenerateUpdateCustomRoleOptions(t *testing.T) {
	type args struct {
		id     string
		etag   string
		params v1alpha1.CustomRoleParameters
	}
	type want struct {
		instance *iampmv1.UpdateRoleOptions
	}
	cases := map[string]struct {
		args args
		want want
	}{
		"FullConversion": {
			args: args{id: roleID, etag: eTag, params: *params(func(crp *v1alpha1.CustomRoleParameters) {
				crp.Actions = []string{action1, action2, action3}
			})},
			want: want{instance: instanceUpdOpts(func(upo *iampmv1.UpdateRoleOptions) { upo.SetIfMatch(eTag) })},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			r := &iampmv1.UpdateRoleOptions{}
			GenerateUpdateCustomRoleOptions(tc.args.id, tc.args.etag, tc.args.params, r)
			if diff := cmp.Diff(tc.want.instance, r); diff != "" {
				t.Errorf("GenerateUpdateCustomRoleOptions(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestLateInitializeSpecs(t *testing.T) {
	type args struct {
		instance *iampmv1.CustomRole
		params   *v1alpha1.CustomRoleParameters
	}
	type want struct {
		params *v1alpha1.CustomRoleParameters
	}
	cases := map[string]struct {
		args args
		want want
	}{
		"SomeFields": {
			args: args{
				params: params(func(p *v1alpha1.CustomRoleParameters) {
					p.DisplayName = ""
				}),
				instance: instance(func(p *iampmv1.CustomRole) {
				}),
			},
			want: want{
				params: params(func(p *v1alpha1.CustomRoleParameters) {
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
		instance *iampmv1.CustomRole
	}
	type want struct {
		obs v1alpha1.CustomRoleObservation
	}
	cases := map[string]struct {
		args args
		want want
	}{
		"FullConversion": {
			args: args{
				instance: instance(func(p *iampmv1.CustomRole) {
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
		params   *v1alpha1.CustomRoleParameters
		instance *iampmv1.CustomRole
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
				params: params(func(crp *v1alpha1.CustomRoleParameters) {
					crp.Actions = []string{action1, action2, action3}
				}),
				instance: instance(func(i *iampmv1.CustomRole) {
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
