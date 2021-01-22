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

package iampolicymanagementv1

import (
	"context"

	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	cpv1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/reference"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	iampmv1 "github.com/IBM/platform-services-go-sdk/iampolicymanagementv1"

	"github.com/crossplane-contrib/provider-ibm-cloud/apis/iampolicymanagementv1/v1alpha1"
	"github.com/crossplane-contrib/provider-ibm-cloud/apis/v1beta1"
	ibmc "github.com/crossplane-contrib/provider-ibm-cloud/pkg/clients"
	ibmcp "github.com/crossplane-contrib/provider-ibm-cloud/pkg/clients/policy"
)

const (
	errNotPolicy           = "managed resource is not a Policy custom resource"
	errNewClient           = "cannot create new Client"
	errCreatePolicy        = "could not create policy"
	errDeletePolicy        = "could not delete policy"
	errGetPolicyFailed     = "error getting policy"
	errCheckUpToDate       = "cannot determine if instance is up to date"
	errGetAuth             = "error getting auth info"
	errManagedUpdateFailed = "cannot update ResourceInstance custom resource"
	errGenObservation      = "error generating observation"
	errCreatePolicyOpts    = "error creating policy options"
	errUpdPolicy           = "error updating policy"
)

// SetupPolicy adds a controller that reconciles Policy managed resources.
func SetupPolicy(mgr ctrl.Manager, l logging.Logger) error {
	name := managed.ControllerName(v1alpha1.PolicyGroupKind)
	log := l.WithValues("Policy-controller", name)

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(v1alpha1.PolicyGroupVersionKind),
		managed.WithExternalConnecter(&pConnector{
			kube:     mgr.GetClient(),
			usage:    resource.NewProviderConfigUsageTracker(mgr.GetClient(), &v1beta1.ProviderConfigUsage{}),
			clientFn: ibmc.NewClient,
			logger:   log}),
		managed.WithInitializers(managed.NewDefaultProviderConfig(mgr.GetClient())),
		managed.WithReferenceResolver(managed.NewAPISimpleReferenceResolver(mgr.GetClient())),
		managed.WithLogger(log),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))))

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&v1alpha1.Policy{}).
		Complete(r)
}

// A pConnector is expected to produce an ExternalClient when its Connect method
// is called.
type pConnector struct {
	kube     client.Client
	usage    resource.Tracker
	clientFn func(optd ibmc.ClientOptions) (ibmc.ClientSession, error)
	logger   logging.Logger
}

// Connect produces an ExternalClient for IBM Cloud API
func (c *pConnector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	opts, err := ibmc.GetAuthInfo(ctx, c.kube, mg)
	if err != nil {
		return nil, errors.Wrap(err, errGetAuth)
	}

	service, err := c.clientFn(opts)
	if err != nil {
		return nil, errors.Wrap(err, errNewClient)
	}

	return &pExternal{client: service, kube: c.kube, logger: c.logger}, nil
}

// An pExternal observes, then either creates, updates, or deletes an
// external resource to ensure it reflects the managed resource's desired state.
type pExternal struct {
	client ibmc.ClientSession
	kube   client.Client
	logger logging.Logger
}

func (c *pExternal) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) { // nolint:gocyclo
	cr, ok := mg.(*v1alpha1.Policy)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotPolicy)
	}

	if meta.GetExternalName(cr) == "" {
		return managed.ExternalObservation{
			ResourceExists: false,
		}, nil
	}

	instance, resp, err := c.client.IamPolicyManagementV1().GetPolicy(&iampmv1.GetPolicyOptions{PolicyID: reference.ToPtrValue(meta.GetExternalName(cr))})
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(resource.Ignore(ibmc.IsResourceNotFound, err), errGetPolicyFailed)
	}
	ibmc.SetEtagAnnotation(cr, ibmc.GetEtag(resp))

	currentSpec := cr.Spec.ForProvider.DeepCopy()
	if err = ibmcp.LateInitializeSpec(&cr.Spec.ForProvider, instance); err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errManagedUpdateFailed)
	}
	if !cmp.Equal(currentSpec, &cr.Spec.ForProvider) {
		if err := c.kube.Update(ctx, cr); err != nil {
			return managed.ExternalObservation{}, errors.Wrap(err, errManagedUpdateFailed)
		}
	}

	cr.Status.AtProvider, err = ibmcp.GenerateObservation(instance)
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errGenObservation)
	}

	cr.Status.SetConditions(cpv1alpha1.Available())

	upToDate, err := ibmcp.IsUpToDate(&cr.Spec.ForProvider, instance, c.logger)
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errCheckUpToDate)
	}

	return managed.ExternalObservation{
		ResourceExists:    true,
		ResourceUpToDate:  upToDate,
		ConnectionDetails: nil,
	}, nil
}

func (c *pExternal) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1alpha1.Policy)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotPolicy)
	}

	cr.SetConditions(cpv1alpha1.Creating())
	resInstanceOptions := &iampmv1.CreatePolicyOptions{}
	if err := ibmcp.GenerateCreatePolicyOptions(cr.Spec.ForProvider, resInstanceOptions); err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreatePolicyOpts)
	}

	instance, _, err := c.client.IamPolicyManagementV1().CreatePolicy(resInstanceOptions)
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(resource.Ignore(ibmc.IsResourceInactive, err), errCreatePolicy)
	}

	meta.SetExternalName(cr, reference.FromPtrValue(instance.ID))
	return managed.ExternalCreation{ExternalNameAssigned: true}, nil
}

func (c *pExternal) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1alpha1.Policy)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotPolicy)
	}

	id := cr.Status.AtProvider.ID
	eTag := ibmc.GetEtagAnnotation(cr)
	updInstanceOpts := &iampmv1.UpdatePolicyOptions{}
	if err := ibmcp.GenerateUpdatePolicyOptions(id, eTag, cr.Spec.ForProvider, updInstanceOpts); err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errUpdPolicy)
	}

	_, _, err := c.client.IamPolicyManagementV1().UpdatePolicy(updInstanceOpts)
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errUpdPolicy)
	}

	return managed.ExternalUpdate{}, nil
}

func (c *pExternal) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*v1alpha1.Policy)
	if !ok {
		return errors.New(errNotPolicy)
	}

	cr.SetConditions(cpv1alpha1.Deleting())

	_, err := c.client.IamPolicyManagementV1().DeletePolicy(&iampmv1.DeletePolicyOptions{PolicyID: &cr.Status.AtProvider.ID})
	if err != nil {
		return errors.Wrap(resource.Ignore(ibmc.IsResourceGone, err), errDeletePolicy)
	}
	return nil
}
