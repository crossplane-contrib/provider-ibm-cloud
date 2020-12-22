/*
Copyright 2020 The Crossplane Authors.

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
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane/crossplane-runtime/pkg/reference"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	rcv2 "github.com/crossplane-contrib/provider-ibm-cloud/apis/resourcecontrollerv2/v1alpha1"
)

// ResolveReferences of this ScalingGroup
func (mg *ScalingGroup) ResolveReferences(ctx context.Context, c client.Reader) error {
	r := reference.NewAPIResolver(c, mg)

	rsp, err := r.Resolve(ctx, reference.ResolutionRequest{
		CurrentValue: reference.FromPtrValue(mg.Spec.ForProvider.ID),
		Reference:    mg.Spec.ForProvider.IDRef,
		Selector:     mg.Spec.ForProvider.IDSelector,
		To:           reference.To{Managed: &rcv2.ResourceInstance{}, List: &rcv2.ResourceInstanceList{}},
		Extract:      ID(),
	})
	if err != nil {
		return errors.Wrap(err, "spec.forProvider.id")
	}
	mg.Spec.ForProvider.ID = reference.ToPtrValue(rsp.ResolvedValue)
	mg.Spec.ForProvider.IDRef = rsp.ResolvedReference
	return nil
}

// ResolveReferences of this Whitelist
func (mg *Whitelist) ResolveReferences(ctx context.Context, c client.Reader) error {
	r := reference.NewAPIResolver(c, mg)

	rsp, err := r.Resolve(ctx, reference.ResolutionRequest{
		CurrentValue: reference.FromPtrValue(mg.Spec.ForProvider.ID),
		Reference:    mg.Spec.ForProvider.IDRef,
		Selector:     mg.Spec.ForProvider.IDSelector,
		To:           reference.To{Managed: &rcv2.ResourceInstance{}, List: &rcv2.ResourceInstanceList{}},
		Extract:      ID(),
	})
	if err != nil {
		return errors.Wrap(err, "spec.forProvider.id")
	}
	mg.Spec.ForProvider.ID = reference.ToPtrValue(rsp.ResolvedValue)
	mg.Spec.ForProvider.IDRef = rsp.ResolvedReference
	return nil
}

// ID extracts the resolved ResourceInstance's ID
func ID() reference.ExtractValueFn {
	return func(mg resource.Managed) string {
		cr, ok := mg.(*rcv2.ResourceInstance)
		if !ok {
			return ""
		}
		return cr.Status.AtProvider.ID
	}
}
