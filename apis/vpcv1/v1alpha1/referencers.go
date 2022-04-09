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

	"github.com/crossplane/crossplane-runtime/pkg/reference"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ResolveReferences resolves the crossplane reference to the VPC
func (mg *Subnet) ResolveReferences(ctx context.Context, c client.Reader) error {
	r := reference.NewAPIResolver(c, mg)

	if mg.Spec.ForProvider.ByTocalCount != nil {
		rsp, err := r.Resolve(ctx, reference.ResolutionRequest{
			CurrentValue: reference.FromPtrValue(mg.Spec.ForProvider.ByTocalCount.VPC.ID),
			Reference:    mg.Spec.ForProvider.ByTocalCount.VPC.VPCRef,
			Selector:     mg.Spec.ForProvider.ByTocalCount.VPC.VPCSelector,
			To:           reference.To{Managed: &VPC{}, List: &VPCList{}},
			Extract:      vpcID(),
		})

		if err != nil {
			return errors.Wrap(err, "spec.forProvider.ByTocalCount")
		}

		mg.Spec.ForProvider.ByTocalCount.VPC.ID = reference.ToPtrValue(rsp.ResolvedValue)
		mg.Spec.ForProvider.ByTocalCount.VPC.VPCRef = rsp.ResolvedReference
	} else {
		rsp, err := r.Resolve(ctx, reference.ResolutionRequest{
			CurrentValue: reference.FromPtrValue(mg.Spec.ForProvider.ByCIDR.VPC.ID),
			Reference:    mg.Spec.ForProvider.ByCIDR.VPC.VPCRef,
			Selector:     mg.Spec.ForProvider.ByCIDR.VPC.VPCSelector,
			To:           reference.To{Managed: &VPC{}, List: &VPCList{}},
			Extract:      vpcID(),
		})

		if err != nil {
			return errors.Wrap(err, "spec.forProvider.ByCIDR")
		}

		mg.Spec.ForProvider.ByCIDR.VPC.ID = reference.ToPtrValue(rsp.ResolvedValue)
		mg.Spec.ForProvider.ByCIDR.VPC.VPCRef = rsp.ResolvedReference
	}

	return nil
}

// Extracts the resolved VPC ID - "" if it cannot
func vpcID() reference.ExtractValueFn {
	return func(mg resource.Managed) string {
		cr, ok := mg.(*VPC)
		if !ok {
			return ""
		}

		return cr.Status.AtProvider.ID
	}
}
