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
	ibmccr "github.com/crossplane-contrib/provider-ibm-cloud/pkg/clients/customrole"
)

const (
	errNotCustomRole        = "managed resource is not a CustomRole custom resource"
	errCreateCustomRole     = "could not create role"
	errDeleteCustomRole     = "could not delete role"
	errGetCustomRoleFailed  = "error getting role"
	errCreateCustomRoleOpts = "error creating role options"
	errUpdCustomRole        = "error updating role"
)

// SetupCustomRole adds a controller that reconciles CustomRole managed resources.
func SetupCustomRole(mgr ctrl.Manager, l logging.Logger) error {
	name := managed.ControllerName(v1alpha1.CustomRoleGroupKind)
	log := l.WithValues("CustomRole-controller", name)

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(v1alpha1.CustomRoleGroupVersionKind),
		managed.WithExternalConnecter(&crConnector{
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
		For(&v1alpha1.CustomRole{}).
		Complete(r)
}

// A crConnector is expected to produce an ExternalClient when its Connect method
// is called.
type crConnector struct {
	kube     client.Client
	usage    resource.Tracker
	clientFn func(optd ibmc.ClientOptions) (ibmc.ClientSession, error)
	logger   logging.Logger
}

// Connect produces an ExternalClient for IBM Cloud API
func (c *crConnector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	opts, err := ibmc.GetAuthInfo(ctx, c.kube, mg)
	if err != nil {
		return nil, errors.Wrap(err, errGetAuth)
	}

	service, err := c.clientFn(opts)
	if err != nil {
		return nil, errors.Wrap(err, errNewClient)
	}

	return &crExternal{client: service, kube: c.kube, logger: c.logger}, nil
}

// An crExternal observes, then either creates, updates, or deletes an
// external resource to ensure it reflects the managed resource's desired state.
type crExternal struct {
	client ibmc.ClientSession
	kube   client.Client
	logger logging.Logger
}

func (c *crExternal) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) { // nolint:gocyclo
	cr, ok := mg.(*v1alpha1.CustomRole)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotCustomRole)
	}

	if meta.GetExternalName(cr) == "" {
		return managed.ExternalObservation{
			ResourceExists: false,
		}, nil
	}

	instance, resp, err := c.client.IamPolicyManagementV1().GetRole(&iampmv1.GetRoleOptions{RoleID: reference.ToPtrValue(meta.GetExternalName(cr))})
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(resource.Ignore(ibmc.IsResourceNotFound, err), errGetCustomRoleFailed)
	}
	ibmc.SetEtagAnnotation(cr, ibmc.GetEtag(resp.Headers))

	currentSpec := cr.Spec.ForProvider.DeepCopy()
	if err = ibmccr.LateInitializeSpec(&cr.Spec.ForProvider, instance); err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errManagedUpdateFailed)
	}
	if !cmp.Equal(currentSpec, &cr.Spec.ForProvider) {
		if err := c.kube.Update(ctx, cr); err != nil {
			return managed.ExternalObservation{}, errors.Wrap(err, errManagedUpdateFailed)
		}
	}

	cr.Status.AtProvider, err = ibmccr.GenerateObservation(instance)
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errGenObservation)
	}

	cr.Status.SetConditions(cpv1alpha1.Available())
	cr.Status.AtProvider.State = ibmccr.StateActive

	upToDate, err := ibmccr.IsUpToDate(&cr.Spec.ForProvider, instance, c.logger)
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errCheckUpToDate)
	}

	return managed.ExternalObservation{
		ResourceExists:    true,
		ResourceUpToDate:  upToDate,
		ConnectionDetails: nil,
	}, nil
}

func (c *crExternal) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1alpha1.CustomRole)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotCustomRole)
	}

	cr.SetConditions(cpv1alpha1.Creating())
	resInstanceOptions := &iampmv1.CreateRoleOptions{}
	if err := ibmccr.GenerateCreateCustomRoleOptions(cr.Spec.ForProvider, resInstanceOptions); err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreateCustomRoleOpts)
	}

	instance, _, err := c.client.IamPolicyManagementV1().CreateRole(resInstanceOptions)
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(resource.Ignore(ibmc.IsResourceInactive, err), errCreateCustomRole)
	}

	meta.SetExternalName(cr, reference.FromPtrValue(instance.ID))
	return managed.ExternalCreation{ExternalNameAssigned: true}, nil
}

func (c *crExternal) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1alpha1.CustomRole)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotCustomRole)
	}

	id := cr.Status.AtProvider.ID
	eTag := ibmc.GetEtagAnnotation(cr)
	updInstanceOpts := &iampmv1.UpdateRoleOptions{}
	if err := ibmccr.GenerateUpdateCustomRoleOptions(id, eTag, cr.Spec.ForProvider, updInstanceOpts); err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errUpdCustomRole)
	}

	_, _, err := c.client.IamPolicyManagementV1().UpdateRole(updInstanceOpts)
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errUpdCustomRole)
	}

	return managed.ExternalUpdate{}, nil
}

func (c *crExternal) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*v1alpha1.CustomRole)
	if !ok {
		return errors.New(errNotCustomRole)
	}

	cr.SetConditions(cpv1alpha1.Deleting())

	_, err := c.client.IamPolicyManagementV1().DeleteRole(&iampmv1.DeleteRoleOptions{RoleID: &cr.Status.AtProvider.ID})
	if err != nil {
		return errors.Wrap(resource.Ignore(ibmc.IsResourceGone, err), errDeleteCustomRole)
	}
	return nil
}
