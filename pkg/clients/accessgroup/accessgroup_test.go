package accessgroup

import (
	"testing"

	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/go-openapi/strfmt"
	"github.com/google/go-cmp/cmp"

	iamagv2 "github.com/IBM/platform-services-go-sdk/iamaccessgroupsv2"

	"github.com/crossplane-contrib/provider-ibm-cloud/apis/iamaccessgroupsv2/v1alpha1"
	ibmc "github.com/crossplane-contrib/provider-ibm-cloud/pkg/clients"
)

var (
	accountID         = "aa5a00334eaf9eb9339d2ab48f20d7ff"
	agName            = "myAccessGroup"
	agDescription     = "myAccessGroup Description"
	agDescription2    = "myAccessGroup Description 2"
	agID              = "12345678-abcd-1a2b-a1b2-1234567890ab"
	createdAt, _      = strfmt.ParseDateTime("2020-10-31T02:33:06Z")
	lastModifiedAt, _ = strfmt.ParseDateTime("2020-10-31T03:33:06Z")
	hRef              = "https://iam.cloud.ibm.com/v2/accessgroups/12345678-abcd-1a2b-a1b2-1234567890ab"
	eTag              = "1-eb832c7ff8c8016a542974b9f880b55e"
	transactionID     = "12345-abcd-ef000-abac"
	createdByID       = "IBM-User-0001"
	isFederated       = false
)

func params(m ...func(*v1alpha1.AccessGroupParameters)) *v1alpha1.AccessGroupParameters {
	p := &v1alpha1.AccessGroupParameters{
		AccountID:     accountID,
		Name:          agName,
		Description:   &agDescription,
		TransactionID: &transactionID,
	}

	for _, f := range m {
		f(p)
	}
	return p
}

func observation(m ...func(*v1alpha1.AccessGroupObservation)) *v1alpha1.AccessGroupObservation {
	o := &v1alpha1.AccessGroupObservation{
		ID:               agID,
		CreatedAt:        ibmc.DateTimeToMetaV1Time(&createdAt),
		LastModifiedAt:   ibmc.DateTimeToMetaV1Time(&lastModifiedAt),
		CreatedByID:      createdByID,
		LastModifiedByID: createdByID,
		Href:             hRef,
		IsFederated:      isFederated,
	}

	for _, f := range m {
		f(o)
	}
	return o
}

func instance(m ...func(*iamagv2.Group)) *iamagv2.Group {
	i := &iamagv2.Group{
		ID:               &agID,
		Href:             &hRef,
		CreatedAt:        &createdAt,
		CreatedByID:      &createdByID,
		LastModifiedAt:   &lastModifiedAt,
		LastModifiedByID: &createdByID,
		Description:      &agDescription,
		Name:             &agName,
		AccountID:        &accountID,
		IsFederated:      &isFederated,
	}

	for _, f := range m {
		f(i)
	}
	return i
}

func instanceOpts(m ...func(*iamagv2.CreateAccessGroupOptions)) *iamagv2.CreateAccessGroupOptions {
	i := &iamagv2.CreateAccessGroupOptions{
		Name:          &agName,
		AccountID:     &accountID,
		Description:   &agDescription,
		TransactionID: &transactionID,
	}
	for _, f := range m {
		f(i)
	}
	return i
}

func instanceUpdOpts(m ...func(*iamagv2.UpdateAccessGroupOptions)) *iamagv2.UpdateAccessGroupOptions {
	i := &iamagv2.UpdateAccessGroupOptions{
		Name:          &agName,
		Description:   &agDescription,
		TransactionID: &transactionID,
		AccessGroupID: &agID,
		IfMatch:       &eTag,
	}

	for _, f := range m {
		f(i)
	}
	return i
}

func TestGenerateCreateAccessGroupOptions(t *testing.T) {
	type args struct {
		params v1alpha1.AccessGroupParameters
	}
	type want struct {
		instance *iamagv2.CreateAccessGroupOptions
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
				params: *params(func(p *v1alpha1.AccessGroupParameters) {
					p.Description = nil
				})},
			want: want{instance: instanceOpts(func(p *iamagv2.CreateAccessGroupOptions) {
				p.Description = nil
			})},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {

			r := &iamagv2.CreateAccessGroupOptions{}
			GenerateCreateAccessGroupOptions(tc.args.params, r)
			if diff := cmp.Diff(tc.want.instance, r); diff != "" {
				t.Errorf("GenerateCreateAccessGroupOptions(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestGenerateUpdateAccessGroupOptions(t *testing.T) {
	type args struct {
		id     string
		etag   string
		params v1alpha1.AccessGroupParameters
	}
	type want struct {
		instance *iamagv2.UpdateAccessGroupOptions
	}
	cases := map[string]struct {
		args args
		want want
	}{
		"FullConversion": {
			args: args{id: agID, etag: eTag, params: *params(func(crp *v1alpha1.AccessGroupParameters) {
				crp.Description = &agDescription2
			})},
			want: want{instance: instanceUpdOpts(func(upo *iamagv2.UpdateAccessGroupOptions) {
				upo.SetIfMatch(eTag)
				upo.Description = &agDescription2
			})},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			r := &iamagv2.UpdateAccessGroupOptions{}
			GenerateUpdateAccessGroupOptions(tc.args.id, tc.args.etag, tc.args.params, r)
			if diff := cmp.Diff(tc.want.instance, r); diff != "" {
				t.Errorf("GenerateUpdateAccessGroupOptions(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestLateInitializeSpecs(t *testing.T) {
	type args struct {
		instance *iamagv2.Group
		params   *v1alpha1.AccessGroupParameters
	}
	type want struct {
		params *v1alpha1.AccessGroupParameters
	}
	cases := map[string]struct {
		args args
		want want
	}{
		"SomeFields": {
			args: args{
				params: params(func(p *v1alpha1.AccessGroupParameters) {
					p.Description = nil
				}),
				instance: instance(func(p *iamagv2.Group) {
				}),
			},
			want: want{
				params: params(func(p *v1alpha1.AccessGroupParameters) {
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
		instance *iamagv2.Group
	}
	type want struct {
		obs v1alpha1.AccessGroupObservation
	}
	cases := map[string]struct {
		args args
		want want
	}{
		"FullConversion": {
			args: args{
				instance: instance(func(p *iamagv2.Group) {
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
		params   *v1alpha1.AccessGroupParameters
		instance *iamagv2.Group
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
				params: params(func(crp *v1alpha1.AccessGroupParameters) {
					crp.Description = &agDescription2
				}),
				instance: instance(func(i *iamagv2.Group) {
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
