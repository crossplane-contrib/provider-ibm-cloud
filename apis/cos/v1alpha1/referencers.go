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

	"github.com/pkg/errors"

	rc2 "github.com/crossplane-contrib/provider-ibm-cloud/apis/resourcecontrollerv2/v1alpha1"

	"github.com/crossplane/crossplane-runtime/pkg/reference"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ResolveReferences resolves the crossplane reference id to the IBM Cloud reference instance id
func (mg *Bucket) ResolveReferences(ctx context.Context, c client.Reader) error {
	r := reference.NewAPIResolver(c, mg)

	rsp, err := r.Resolve(ctx, reference.ResolutionRequest{
		CurrentValue: reference.FromPtrValue(mg.Spec.ForProvider.IbmServiceInstanceID),
		Reference:    mg.Spec.ForProvider.IbmServiceInstanceIDRef,
		Selector:     mg.Spec.ForProvider.IbmServiceInstanceIDSelector,
		To:           reference.To{Managed: &rc2.ResourceInstance{}, List: &rc2.ResourceInstanceList{}},
		Extract:      sourceGUID(),
	})

	if err != nil {
		return errors.Wrap(err, "spec.forProvider.Source")
	}

	mg.Spec.ForProvider.IbmServiceInstanceID = reference.ToPtrValue(rsp.ResolvedValue)
	mg.Spec.ForProvider.IbmServiceInstanceIDRef = rsp.ResolvedReference

	return nil
}

// ResolveReferences resolves the crossplane reference
func (mg *BucketConfig) ResolveReferences(ctx context.Context, c client.Reader) error {
	r := reference.NewAPIResolver(c, mg)

	rsp, err := r.Resolve(ctx, reference.ResolutionRequest{
		CurrentValue: reference.FromPtrValue(mg.Spec.ForProvider.Name),
		Reference:    mg.Spec.ForProvider.NameRef,
		Selector:     mg.Spec.ForProvider.NameSelector,
		To:           reference.To{Managed: &Bucket{}, List: &BucketList{}},
		Extract:      sourceName(),
	})

	if err != nil {
		return errors.Wrap(err, "spec.forProvider.Source")
	}

	mg.Spec.ForProvider.Name = reference.ToPtrValue(rsp.ResolvedValue)
	mg.Spec.ForProvider.NameRef = rsp.ResolvedReference

	return nil
}

// Extracts the resolved ResourceInstance's GUID - "" if it cannot
func sourceGUID() reference.ExtractValueFn {
	return func(mg resource.Managed) string {
		cr, ok := mg.(*rc2.ResourceInstance)
		if !ok {
			return ""
		}

		return cr.Status.AtProvider.GUID
	}
}

// Extracts the resolved Bucket's name - "" if it cannot
func sourceName() reference.ExtractValueFn {
	return func(mg resource.Managed) string {
		cr, ok := mg.(*Bucket)
		if !ok {
			return ""
		}

		return cr.Spec.ForProvider.Name
	}
}
