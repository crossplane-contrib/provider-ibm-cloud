package resourcekey

import (
	"fmt"
	"time"

	"github.com/go-openapi/strfmt"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/jeremywohl/flatten"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/reference"

	rcv2 "github.com/IBM/platform-services-go-sdk/resourcecontrollerv2"

	"github.com/crossplane-contrib/provider-ibm-cloud/apis/resourcecontrollerv2/v1alpha1"
	ibmc "github.com/crossplane-contrib/provider-ibm-cloud/pkg/clients"
)

const (
	// StateActive represents a service instance in a running, available, and ready state
	StateActive = "active"
	// StateInactive represents a service instance in a not running state
	StateInactive = "inactive"
	// StateRemoved means that delete has been initiated
	StateRemoved = "removed"
)

// LateInitializeSpec fills optional and unassigned fields with the values in *rcv2.ResourceKey object.
func LateInitializeSpec(spec *v1alpha1.ResourceKeyParameters, in *rcv2.ResourceKey) error { // nolint:gocyclo
	if spec.Role == nil {
		spec.Role = in.Role
	}
	// TODO -add parameters once https://github.com/IBM/platform-services-go-sdk/issues/57 is resolved
	return nil
}

// GenerateCreateResourceKeyOptions produces ResourceKeyOptions object from ResourceKeyParameters object.
func GenerateCreateResourceKeyOptions(name string, in v1alpha1.ResourceKeyParameters, o *rcv2.CreateResourceKeyOptions) error {
	o.Name = reference.ToPtrValue(in.Name)
	// TODO o.Parameters = helper
	o.Role = in.Role
	o.Source = in.Source
	return nil
}

// GenerateUpdateResourceKeyOptions produces UpdateResourceKeyOptions object from ResourceKey object.
func GenerateUpdateResourceKeyOptions(name, id string, in v1alpha1.ResourceKeyParameters, o *rcv2.UpdateResourceKeyOptions) error {
	o.Name = reference.ToPtrValue(in.Name)
	o.ID = reference.ToPtrValue(id)
	return nil
}

// GenerateObservation produces ResourceKeyObservation object from *rcv2.ResourceKey object.
func GenerateObservation(in *rcv2.ResourceKey) (v1alpha1.ResourceKeyObservation, error) {
	o := v1alpha1.ResourceKeyObservation{
		AccountID:           reference.FromPtrValue(in.AccountID),
		CreatedBy:           reference.FromPtrValue(in.CreatedBy),
		DeletedBy:           reference.FromPtrValue(in.DeletedBy),
		IamCompatible:       ibmc.BoolValue(in.IamCompatible),
		ResourceInstanceURL: reference.FromPtrValue(in.ResourceInstanceURL),
		UpdatedBy:           reference.FromPtrValue(in.UpdatedBy),
		CreatedAt:           GenerateMetaV1Time(in.CreatedAt),
		CRN:                 reference.FromPtrValue(in.CRN),
		DeletedAt:           GenerateMetaV1Time(in.DeletedAt),
		GUID:                reference.FromPtrValue(in.GUID),
		ID:                  reference.FromPtrValue(in.ID),
		ResourceGroupID:     reference.FromPtrValue(in.ResourceGroupID),
		State:               reference.FromPtrValue(in.State),
		URL:                 reference.FromPtrValue(in.URL),
		UpdatedAt:           GenerateMetaV1Time(in.UpdatedAt),
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
func IsUpToDate(name string, in *v1alpha1.ResourceKeyParameters, observed *rcv2.ResourceKey, l logging.Logger) (bool, error) {
	desired := in.DeepCopy()
	actual, err := GenerateResourceKeyParameters(observed)
	if err != nil {
		return false, err
	}

	l.Info(cmp.Diff(desired, actual, cmpopts.IgnoreTypes(&runtimev1alpha1.Reference{}, &runtimev1alpha1.Selector{}, []runtimev1alpha1.Reference{})))

	return cmp.Equal(desired, actual, cmpopts.EquateEmpty(),
		cmpopts.IgnoreFields(v1alpha1.ResourceKeyParameters{}),
		cmpopts.IgnoreTypes(&runtimev1alpha1.Reference{}, &runtimev1alpha1.Selector{}, []runtimev1alpha1.Reference{})), nil
}

// GenerateResourceKeyParameters generates service instance parameters from resource instance
func GenerateResourceKeyParameters(in *rcv2.ResourceKey) (*v1alpha1.ResourceKeyParameters, error) {
	o := &v1alpha1.ResourceKeyParameters{
		Name:   reference.FromPtrValue(in.Name),
		Role:   in.Role,
		Source: in.SourceCRN,
		// TODO - need resolution for https://github.com/IBM/platform-services-go-sdk/issues/57
		// Parameters: GenerateResourceKeyPostParameters(in.),
	}
	return o, nil
}

// GenerateResourceKeyPostParameters generates v1alpha1.ResourceKeyPostParameters from rcv2.ResourceKeyPostParameters
func GenerateResourceKeyPostParameters(in *rcv2.ResourceKeyPostParameters) *v1alpha1.ResourceKeyPostParameters {
	o := &v1alpha1.ResourceKeyPostParameters{
		ServiceidCRN: reference.FromPtrValue(in.ServiceidCRN),
	}
	return o
}

// GetConnectionDetails generate the connection details from the *rcv2.ResourceKey in a format ready to be set into a secret
func GetConnectionDetails(cr *v1alpha1.ResourceKey, in *rcv2.ResourceKey) (managed.ConnectionDetails, error) {
	if in.Credentials == nil {
		return managed.ConnectionDetails{}, nil
	}
	if cr.Spec.ConnectionTemplates != nil {
		return handleTemplatedConnectionVars(cr, in)
	}
	return handleFlettenedConnectionVars(in)
}

func handleTemplatedConnectionVars(cr *v1alpha1.ResourceKey, in *rcv2.ResourceKey) (managed.ConnectionDetails, error) {
	creds, err := ibmc.ConvertStructToMap(in.Credentials)
	if err != nil {
		return nil, err
	}
	parser := ibmc.NewTemplateParser(cr.Spec.ConnectionTemplates, creds)
	values, err := parser.Parse()
	if err != nil {
		return nil, err
	}
	return ibmc.ConvertVarsMap(values), nil
}

func handleFlettenedConnectionVars(in *rcv2.ResourceKey) (managed.ConnectionDetails, error) {
	m := managed.ConnectionDetails{
		"apikey":               ibmc.StrPtr2Bytes(in.Credentials.Apikey),
		"iamApikeyDescription": ibmc.StrPtr2Bytes(in.Credentials.IamApikeyDescription),
		"iamApikeyName":        ibmc.StrPtr2Bytes(in.Credentials.IamApikeyName),
		"iamRoleCRN":           ibmc.StrPtr2Bytes(in.Credentials.IamRoleCRN),
		"iamServiceidCRN":      ibmc.StrPtr2Bytes(in.Credentials.IamServiceidCRN),
	}
	f, err := flatten.Flatten(in.Credentials.GetProperties(), "", flatten.DotStyle)
	if err != nil {
		return nil, err
	}
	for k, v := range f {
		switch v := v.(type) {
		case int:
			m[k] = []byte(fmt.Sprintf("%d", v))
		case float64:
			m[k] = []byte(fmt.Sprintf("%f", v))
		default:
			m[k] = []byte(fmt.Sprintf("%s", v))
		}
	}
	return m, nil
}
