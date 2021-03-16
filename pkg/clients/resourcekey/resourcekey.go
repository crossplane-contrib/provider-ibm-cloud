/*
Copyright 2021 The Crossplane Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package resourcekey

import (
	"fmt"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/jeremywohl/flatten"

	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/reference"

	rcv2 "github.com/IBM/platform-services-go-sdk/resourcecontrollerv2"

	"github.com/crossplane-contrib/provider-ibm-cloud/apis/resourcecontrollerv2/v1alpha1"
	ibmc "github.com/crossplane-contrib/provider-ibm-cloud/pkg/clients"
)

// LateInitializeSpec fills optional and unassigned fields with the values in *rcv2.ResourceKey object.
func LateInitializeSpec(client ibmc.ClientSession, spec *v1alpha1.ResourceKeyParameters, in *rcv2.ResourceKey) error { // nolint:gocyclo
	if spec.Role == nil {
		spec.Role = in.Role
	}
	return nil
}

// GenerateCreateResourceKeyOptions produces CreateResourceKeyOptions object from ResourceKeyParameters object.
func GenerateCreateResourceKeyOptions(client ibmc.ClientSession, in v1alpha1.ResourceKeyParameters, o *rcv2.CreateResourceKeyOptions) error {
	o.Name = reference.ToPtrValue(in.Name)
	o.Source = in.Source
	o.Parameters = GenerateParameters(in.Parameters)
	o.Role = in.Role

	return nil
}

// GenerateParameters generates rcv2.ResourceKeyPostParameters from v1alpha1.ResourceKeyPostParameters
func GenerateParameters(in *v1alpha1.ResourceKeyPostParameters) *rcv2.ResourceKeyPostParameters {
	if in == nil {
		return nil
	}
	o := &rcv2.ResourceKeyPostParameters{
		ServiceidCRN: &in.ServiceidCRN,
	}
	return o
}

// GenerateUpdateResourceKeyOptions produces UpdateResourceKeyOptions object from ResourceKeyParameters object.
func GenerateUpdateResourceKeyOptions(client ibmc.ClientSession, id string, in v1alpha1.ResourceKeyParameters, o *rcv2.UpdateResourceKeyOptions) error {
	o.ID = reference.ToPtrValue(id)
	o.Name = reference.ToPtrValue(in.Name)
	return nil
}

// GenerateObservation produces ResourceKeyObservation object from *rcv2.ResourceKey object.
func GenerateObservation(client ibmc.ClientSession, in *rcv2.ResourceKey) (v1alpha1.ResourceKeyObservation, error) {
	o := v1alpha1.ResourceKeyObservation{
		ID:                  reference.FromPtrValue(in.ID),
		GUID:                reference.FromPtrValue(in.GUID),
		CRN:                 reference.FromPtrValue(in.CRN),
		URL:                 reference.FromPtrValue(in.URL),
		AccountID:           reference.FromPtrValue(in.AccountID),
		ResourceGroupID:     reference.FromPtrValue(in.ResourceGroupID),
		SourceCRN:           reference.FromPtrValue(in.SourceCRN),
		State:               reference.FromPtrValue(in.State),
		IamCompatible:       ibmc.BoolValue(in.IamCompatible),
		ResourceInstanceURL: reference.FromPtrValue(in.ResourceInstanceURL),
		CreatedAt:           ibmc.DateTimeToMetaV1Time(in.CreatedAt),
		UpdatedAt:           ibmc.DateTimeToMetaV1Time(in.UpdatedAt),
		DeletedAt:           ibmc.DateTimeToMetaV1Time(in.DeletedAt),
		CreatedBy:           reference.FromPtrValue(in.CreatedBy),
		UpdatedBy:           reference.FromPtrValue(in.UpdatedBy),
		DeletedBy:           reference.FromPtrValue(in.DeletedBy),
	}
	// ServiceEndpoints can be found in instance.Parameters["service-endpoints"]
	return o, nil
}

// GenerateCredentials generates v1alpha1.Credentials from rcv2.Credentials
func GenerateCredentials(in rcv2.Credentials) v1alpha1.Credentials {
	credentials := v1alpha1.Credentials{
		Apikey:               reference.FromPtrValue(in.Apikey),
		IamApikeyDescription: reference.FromPtrValue(in.IamApikeyDescription),
		IamApikeyName:        reference.FromPtrValue(in.IamApikeyName),
		IamRoleCRN:           reference.FromPtrValue(in.IamRoleCRN),
		IamServiceidCRN:      reference.FromPtrValue(in.IamServiceidCRN),
	}
	return credentials
}

// IsUpToDate checks whether current state is up-to-date compared to the given set of parameters.
func IsUpToDate(client ibmc.ClientSession, in *v1alpha1.ResourceKeyParameters, observed *rcv2.ResourceKey, l logging.Logger) (bool, error) {
	desired := in.DeepCopy()
	actual, err := GenerateResourceKeyParameters(client, observed)
	if err != nil {
		return false, err
	}

	diff := (cmp.Diff(desired, actual,
		cmpopts.EquateEmpty(),
		cmpopts.IgnoreFields(v1alpha1.ResourceKeyParameters{}, "Source", "Parameters"), cmpopts.IgnoreTypes(&runtimev1alpha1.Reference{}, &runtimev1alpha1.Selector{}, []runtimev1alpha1.Reference{})))

	if diff != "" {
		l.Info("IsUpToDate", "Diff", diff)
		return false, nil
	}

	return true, nil
}

// GenerateResourceKeyParameters generates service instance parameters from resource instance
func GenerateResourceKeyParameters(client ibmc.ClientSession, in *rcv2.ResourceKey) (*v1alpha1.ResourceKeyParameters, error) {
	o := &v1alpha1.ResourceKeyParameters{
		Name: reference.FromPtrValue(in.Name),
		Role: in.Role,
	}

	return o, nil
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
		"iamRoleCrn":           ibmc.StrPtr2Bytes(in.Credentials.IamRoleCRN),
		"iamServiceidCrn":      ibmc.StrPtr2Bytes(in.Credentials.IamServiceidCRN),
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
