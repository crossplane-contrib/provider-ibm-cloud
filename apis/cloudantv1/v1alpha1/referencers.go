// /*
// Copyright 2021 The Crossplane Authors.

// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at

//     http://www.apache.org/licenses/LICENSE-2.0

// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
// */

package v1alpha1

import (
	"context"
	"encoding/json"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane/crossplane-runtime/pkg/reference"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	rcv2 "github.com/crossplane-contrib/provider-ibm-cloud/apis/resourcecontrollerv2/v1alpha1"
)

const (
	cloudantAdminURLKey = "url"
	errKeyNotFound      = "key not found"
)

// ResolveReferences of this ResourceKey
// Crossplane resolvers are not designed to resolve fields from non-Crossplane resources.
// There is a design doc in progress to support this type of scenario at https://github.com/crossplane/crossplane/pull/2385
// At this time the only solution is a two steps approach:
// 1. use the resolver on a resource key to obtain the namespage and name of the secret from writeConnectionSecretToRef
// 2. use that namespace and name with the client.Reader to get the secret and extract the cloudant_admin_url from there
func (mg *CloudantDatabase) ResolveReferences(ctx context.Context, c client.Reader) error {
	r := reference.NewAPIResolver(c, mg)

	rsp, err := r.Resolve(ctx, reference.ResolutionRequest{
		CurrentValue: reference.FromPtrValue(mg.Spec.ForProvider.CloudantAdminURL),
		Reference:    mg.Spec.ForProvider.CloudantAdminURLRef,
		Selector:     mg.Spec.ForProvider.CloudantAdminURLSelector,
		To:           reference.To{Managed: &rcv2.ResourceKey{}, List: &rcv2.ResourceKeyList{}},
		Extract:      ConnSecretRef(),
	})
	if err != nil {
		return errors.Wrap(err, "spec.forProvider.CloudantAdminURL")
	}

	nsName, err := jsonStringToNamespacedName(rsp.ResolvedValue)
	if err != nil {
		return nil
	}
	creds := &corev1.Secret{}
	if err = c.Get(ctx, nsName, creds); err != nil {
		return err
	}
	url, ok := creds.Data[cloudantAdminURLKey]
	if !ok {
		return errors.Wrap(errors.New(errKeyNotFound), cloudantAdminURLKey)
	}
	mg.Spec.ForProvider.CloudantAdminURL = reference.ToPtrValue(string(url))
	mg.Spec.ForProvider.CloudantAdminURLRef = rsp.ResolvedReference
	return nil
}

// ConnSecretRef extracts the connection secret namespace and name
func ConnSecretRef() reference.ExtractValueFn {
	return func(mg resource.Managed) string {
		cr, ok := mg.(*rcv2.ResourceKey)
		if !ok {
			return ""
		}
		return namespacedNameToJSONString(cr.Spec.WriteConnectionSecretToReference.Namespace, cr.Spec.WriteConnectionSecretToReference.Name)
	}
}

func namespacedNameToJSONString(namespace, name string) string {
	nsName := types.NamespacedName{
		Namespace: namespace,
		Name:      name}
	b, _ := json.Marshal(nsName)
	return string(b)
}

func jsonStringToNamespacedName(nsNameStr string) (types.NamespacedName, error) {
	nsName := &types.NamespacedName{}
	err := json.Unmarshal([]byte(nsNameStr), nsName)
	if err != nil {
		return *nsName, err
	}
	return *nsName, nil
}
