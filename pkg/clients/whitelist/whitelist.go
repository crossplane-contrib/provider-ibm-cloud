package scalinggroup

import (
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/reference"

	icdv5 "github.com/IBM/experimental-go-sdk/ibmclouddatabasesv5"

	"github.com/crossplane-contrib/provider-ibm-cloud/apis/ibmclouddatabasesv5/v1alpha1"
)

// MemberGroupID is the default ID for members group
const MemberGroupID = "member"

// LateInitializeSpec fills optional and unassigned fields with the values in *icdv5.Group object.
func LateInitializeSpec(spec *v1alpha1.WhitelistParameters, in *icdv5.Whitelist) error { // nolint:gocyclo
	if in.IpAddresses == nil || in.IpAddresses != nil && len(in.IpAddresses) == 0 {
		return nil
	}

	if spec.IPAddresses == nil {
		spec.IPAddresses = []v1alpha1.WhitelistEntry{}
		for _, wl := range in.IpAddresses {
			spec.IPAddresses = append(spec.IPAddresses, v1alpha1.WhitelistEntry{
				Address:     reference.FromPtrValue(wl.Address),
				Description: wl.Description,
			})
		}
	}

	return nil
}

// GenerateReplaceWhitelistOptions produces SetDeploymentWhitelistOptions object from WhitelistParameters object.
func GenerateReplaceWhitelistOptions(id string, in v1alpha1.WhitelistParameters, o *icdv5.ReplaceWhitelistOptions) error {
	o.ID = reference.ToPtrValue(id)
	if in.IPAddresses != nil {
		o.IpAddresses = []icdv5.WhitelistEntry{}
		for i := range in.IPAddresses {
			o.IpAddresses = append(o.IpAddresses, icdv5.WhitelistEntry{
				Address:     &in.IPAddresses[i].Address,
				Description: in.IPAddresses[i].Description,
			})
		}
	}
	if in.IfMatch != nil {
		o.IfMatch = in.IfMatch
	}
	return nil
}

// GenerateObservation produces WhitelistObservation object from *icdv5.Whitelist object.
func GenerateObservation(in *icdv5.Whitelist) (v1alpha1.WhitelistObservation, error) {
	o := v1alpha1.WhitelistObservation{}

	return o, nil
}

// IsUpToDate checks whether current state is up-to-date compared to the given
// set of parameters.
func IsUpToDate(id string, in *v1alpha1.WhitelistParameters, observed *icdv5.Whitelist, l logging.Logger) (bool, error) {
	desired := in.DeepCopy()
	actual, err := GenerateWhitelistParameters(id, observed)
	if err != nil {
		return false, err
	}

	l.Info(cmp.Diff(desired, actual, cmpopts.IgnoreTypes(&runtimev1alpha1.Reference{}, &runtimev1alpha1.Selector{}, []runtimev1alpha1.Reference{})))

	return cmp.Equal(desired, actual, cmpopts.EquateEmpty(),
		cmpopts.IgnoreFields(v1alpha1.WhitelistParameters{}, "IfMatch"),
		cmpopts.IgnoreTypes(&runtimev1alpha1.Reference{}, &runtimev1alpha1.Selector{}, []runtimev1alpha1.Reference{})), nil
}

// GenerateWhitelistParameters generates white list parameters from whitelist
func GenerateWhitelistParameters(id string, in *icdv5.Whitelist) (*v1alpha1.WhitelistParameters, error) {
	o := &v1alpha1.WhitelistParameters{
		ID: &id,
	}
	o.IPAddresses = []v1alpha1.WhitelistEntry{}
	for _, wl := range in.IpAddresses {
		o.IPAddresses = append(o.IPAddresses, v1alpha1.WhitelistEntry{
			Address:     reference.FromPtrValue(wl.Address),
			Description: wl.Description,
		})
	}
	return o, nil
}
