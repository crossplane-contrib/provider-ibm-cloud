package scalinggroup

import (
	"testing"

	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/google/go-cmp/cmp"

	icdv5 "github.com/IBM/experimental-go-sdk/ibmclouddatabasesv5"

	"github.com/crossplane-contrib/provider-ibm-cloud/apis/ibmclouddatabasesv5/v1alpha1"
)

var (
	id   = "crn:v1:bluemix:public:databases-for-postgresql:us-south:a/0b5a00334eaf9eb9339d2ab48f20d7f5:dda29288-c259-4dc9-859c-154eb7939cd0::"
	ip1  = "195.212.0.0/16"
	ip1d = "Dev IP space 1"
	ip2  = "195.0.0.0/8"
	ip2d = "Dev IP space 2"
	ip3  = "46.5.0.0/16"

	eTag = "myEtag"
)

func params(m ...func(*v1alpha1.WhitelistParameters)) *v1alpha1.WhitelistParameters {
	p := &v1alpha1.WhitelistParameters{
		ID: &id,
		IPAddresses: []v1alpha1.WhitelistEntry{
			{
				Address:     ip1,
				Description: &ip1d,
			},
			{
				Address:     ip2,
				Description: &ip2d,
			},
		},
		IfMatch: &eTag,
	}
	for _, f := range m {
		f(p)
	}
	return p
}

func observation(m ...func(*v1alpha1.WhitelistObservation)) *v1alpha1.WhitelistObservation {
	o := &v1alpha1.WhitelistObservation{}

	for _, f := range m {
		f(o)
	}
	return o
}

func instance(m ...func(*icdv5.Whitelist)) *icdv5.Whitelist {
	i := &icdv5.Whitelist{
		IpAddresses: []icdv5.WhitelistEntry{
			{
				Address:     &ip1,
				Description: &ip1d,
			},
			{
				Address:     &ip2,
				Description: &ip2d,
			},
		},
	}

	for _, f := range m {
		f(i)
	}
	return i
}

func instanceOpts(m ...func(*icdv5.ReplaceWhitelistOptions)) *icdv5.ReplaceWhitelistOptions {
	i := &icdv5.ReplaceWhitelistOptions{
		ID: &id,
		IpAddresses: []icdv5.WhitelistEntry{
			{
				Address:     &ip1,
				Description: &ip1d,
			},
			{
				Address:     &ip2,
				Description: &ip2d,
			},
		},
		IfMatch: &eTag,
	}
	for _, f := range m {
		f(i)
	}
	return i
}

// func cr(m ...func(*v1alpha1.Whitelist)) *v1alpha1.Whitelist {
// 	i := &v1alpha1.Whitelist{
// 		Spec: v1alpha1.WhitelistSpec{
// 			ForProvider: *params(),
// 		},
// 		Status: v1alpha1.WhitelistStatus{
// 			AtProvider: *observation(),
// 		},
// 	}
// 	for _, f := range m {
// 		f(i)
// 	}
// 	return i
// }

func TestGenerateReplaceWhitelistOptions(t *testing.T) {
	type args struct {
		id         string
		parameters v1alpha1.WhitelistParameters
	}
	type want struct {
		instance *icdv5.ReplaceWhitelistOptions
	}
	cases := map[string]struct {
		args args
		want want
	}{
		"FullConversion": {
			args: args{id: id, parameters: *params()},
			want: want{instance: instanceOpts()},
		},
		"MissingFields": {
			args: args{
				id: id,
				parameters: *params(func(p *v1alpha1.WhitelistParameters) {
					p.IPAddresses = []v1alpha1.WhitelistEntry{
						{
							Address:     ip1,
							Description: &ip1d,
						},
					}
				})},
			want: want{instance: instanceOpts(func(p *icdv5.ReplaceWhitelistOptions) {
				p.IpAddresses = []icdv5.WhitelistEntry{
					{
						Address:     &ip1,
						Description: &ip1d,
					},
				}
			},
			)},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			r := &icdv5.ReplaceWhitelistOptions{}
			GenerateReplaceWhitelistOptions(tc.args.id, tc.args.parameters, r)
			if diff := cmp.Diff(tc.want.instance, r); diff != "" {
				t.Errorf("GenerateReplaceWhitelistOptions(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestLateInitializeSpecs(t *testing.T) {
	type args struct {
		instance *icdv5.Whitelist
		params   *v1alpha1.WhitelistParameters
	}
	type want struct {
		params *v1alpha1.WhitelistParameters
	}
	cases := map[string]struct {
		args args
		want want
	}{
		"SomeFields": {
			args: args{
				params: params(func(p *v1alpha1.WhitelistParameters) {
					p.IPAddresses = nil
				}),
				instance: instance(),
			},
			want: want{
				params: params()},
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
		instance *icdv5.Whitelist
	}
	type want struct {
		obs v1alpha1.WhitelistObservation
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
		params   *v1alpha1.WhitelistParameters
		instance *icdv5.Whitelist
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
				instance: instance(func(i *icdv5.Whitelist) {
					i.IpAddresses[0].Address = &ip3
				}),
			},
			want: want{upToDate: false, isErr: false},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {

			r, err := IsUpToDate(id, tc.args.params, tc.args.instance, logging.NewNopLogger())
			if err != nil && !tc.want.isErr {
				t.Error("IsUpToDate(...) unexpected error")
			}
			if diff := cmp.Diff(tc.want.upToDate, r); diff != "" {
				t.Errorf("IsUpToDate(...): -want, +got:\n%s", diff)
			}
		})
	}
}
