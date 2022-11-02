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

package v1alpha1

import (
	"context"
	"encoding/json"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane/crossplane-runtime/pkg/reference"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	rcv2 "github.com/crossplane-contrib/provider-ibm-cloud/apis/resourcecontrollerv2/v1alpha1"
)

const (
	kafkaAdminURLKey = "kafka_admin_url"
	errKeyNotFound   = "key not found"
)

// ResolveReferences of this ResourceKey
// Crossplane resolvers are not designed to resolve fields from non-Crossplane resources.
// There is a design doc in progress to support this type of scenario at https://github.com/crossplane/crossplane/pull/2385
// At this time the only solution is a two steps approach:
// 1. use the resolver on a resource key to obtain the namespage and name of the secret from writeConnectionSecretToRef
// 2. use that namespace and name with the client.Reader to get the secret and extract the kafka_admin_url from there
func (mg *Topic) ResolveReferences(ctx context.Context, c client.Reader) error {
	r := reference.NewAPIResolver(c, mg)

	rsp, err := r.Resolve(ctx, reference.ResolutionRequest{
		CurrentValue: reference.FromPtrValue(mg.Spec.ForProvider.KafkaAdminURL),
		Reference:    mg.Spec.ForProvider.KafkaAdminURLRef,
		Selector:     mg.Spec.ForProvider.KafkaAdminURLSelector,
		To:           reference.To{Managed: &rcv2.ResourceKey{}, List: &rcv2.ResourceKeyList{}},
		Extract:      ConnSecretRef(),
	})
	if err != nil {
		return errors.Wrap(err, "spec.forProvider.KafkaAdminURL")
	}

	nsName, err := jsonStringToNamespacedName(rsp.ResolvedValue)
	if err != nil {
		return nil
	}
	creds := &corev1.Secret{}
	if err = c.Get(ctx, nsName, creds); err != nil {
		return err
	}
	url, ok := creds.Data[kafkaAdminURLKey]
	if !ok {
		return errors.Wrap(errors.New(errKeyNotFound), kafkaAdminURLKey)
	}
	mg.Spec.ForProvider.KafkaAdminURL = reference.ToPtrValue(string(url))
	mg.Spec.ForProvider.KafkaAdminURLRef = rsp.ResolvedReference
	return nil
}

// ConnSecretRef extracts the connection secret namespace and name
func ConnSecretRef() reference.ExtractValueFn {
	return func(mg resource.Managed) string {
		cr, ok := mg.(*rcv2.ResourceKey)
		if !ok {
			return ""
		}
		n, err := namespacedNameToJSONString(cr.Spec.WriteConnectionSecretToReference.Namespace, cr.Spec.WriteConnectionSecretToReference.Name)
		if err != nil {
			klog.Errorf("ConnSecretRef: %s", err)
		}
		return n
	}
}

func namespacedNameToJSONString(namespace, name string) (string, error) {
	nsName := types.NamespacedName{
		Namespace: namespace,
		Name:      name}
	b, err := json.Marshal(nsName)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func jsonStringToNamespacedName(nsNameStr string) (types.NamespacedName, error) {
	nsName := &types.NamespacedName{}
	err := json.Unmarshal([]byte(nsNameStr), nsName)
	if err != nil {
		return *nsName, err
	}
	return *nsName, nil
}
