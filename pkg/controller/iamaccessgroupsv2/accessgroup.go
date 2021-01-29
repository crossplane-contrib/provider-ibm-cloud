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

package iamaccessgroupsv2

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

	iamagv2 "github.com/IBM/platform-services-go-sdk/iamaccessgroupsv2"

	"github.com/crossplane-contrib/provider-ibm-cloud/apis/iamaccessgroupsv2/v1alpha1"
	"github.com/crossplane-contrib/provider-ibm-cloud/apis/v1beta1"
	ibmc "github.com/crossplane-contrib/provider-ibm-cloud/pkg/clients"
	ibmcag "github.com/crossplane-contrib/provider-ibm-cloud/pkg/clients/accessgroup"
)

const (
	errNewClient             = "cannot create new Client"
	errNotAccessGroup        = "managed resource is not a AccessGroup custom resource"
	errCreateAccessGroup     = "could not create access group"
	errDeleteAccessGroup     = "could not delete access group"
	errGetAccessGroupFailed  = "error getting access group"
	errCreateAccessGroupOpts = "error creating access group options"
	errUpdAccessGroup        = "error updating access group"
	errCheckUpToDate         = "cannot determine if instance is up to date"
	errGetAuth               = "error getting auth info"
	errManagedUpdateFailed   = "cannot update ResourceInstance custom resource"
	errGenObservation        = "error generating observation"
)

// SetupAccessGroup adds a controller that reconciles AccessGroup managed resources.
func SetupAccessGroup(mgr ctrl.Manager, l logging.Logger) error {
	name := managed.ControllerName(v1alpha1.AccessGroupGroupKind)
	log := l.WithValues("AccessGroup-controller", name)

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(v1alpha1.AccessGroupGroupVersionKind),
		managed.WithExternalConnecter(&agConnector{
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
		For(&v1alpha1.AccessGroup{}).
		Complete(r)
}

// A agConnector is expected to produce an ExternalClient when its Connect method
// is called.
type agConnector struct {
	kube     client.Client
	usage    resource.Tracker
	clientFn func(optd ibmc.ClientOptions) (ibmc.ClientSession, error)
	logger   logging.Logger
}

// Connect produces an ExternalClient for IBM Cloud API
func (c *agConnector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	opts, err := ibmc.GetAuthInfo(ctx, c.kube, mg)
	if err != nil {
		return nil, errors.Wrap(err, errGetAuth)
	}

	service, err := c.clientFn(opts)
	if err != nil {
		return nil, errors.Wrap(err, errNewClient)
	}

	return &agExternal{client: service, kube: c.kube, logger: c.logger}, nil
}

// An agExternal observes, then either creates, updates, or deletes an
// external resource to ensure it reflects the managed resource's desired state.
type agExternal struct {
	client ibmc.ClientSession
	kube   client.Client
	logger logging.Logger
}

func (c *agExternal) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) { // nolint:gocyclo
	cr, ok := mg.(*v1alpha1.AccessGroup)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotAccessGroup)
	}

	if meta.GetExternalName(cr) == "" {
		return managed.ExternalObservation{
			ResourceExists: false,
		}, nil
	}

	instance, resp, err := c.client.IamAccessGroupsV2().GetAccessGroup(&iamagv2.GetAccessGroupOptions{AccessGroupID: reference.ToPtrValue(meta.GetExternalName(cr))})
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(resource.Ignore(ibmc.IsResourceNotFound, err), errGetAccessGroupFailed)
	}
	ibmc.SetEtagAnnotation(cr, ibmc.GetEtag(resp))

	currentSpec := cr.Spec.ForProvider.DeepCopy()
	if err = ibmcag.LateInitializeSpec(&cr.Spec.ForProvider, instance); err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errManagedUpdateFailed)
	}
	if !cmp.Equal(currentSpec, &cr.Spec.ForProvider) {
		if err := c.kube.Update(ctx, cr); err != nil {
			return managed.ExternalObservation{}, errors.Wrap(err, errManagedUpdateFailed)
		}
	}

	cr.Status.AtProvider, err = ibmcag.GenerateObservation(instance)
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errGenObservation)
	}

	cr.Status.SetConditions(cpv1alpha1.Available())
	cr.Status.AtProvider.State = ibmcag.StateActive

	upToDate, err := ibmcag.IsUpToDate(&cr.Spec.ForProvider, instance, c.logger)
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errCheckUpToDate)
	}

	return managed.ExternalObservation{
		ResourceExists:    true,
		ResourceUpToDate:  upToDate,
		ConnectionDetails: nil,
	}, nil
}

func (c *agExternal) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1alpha1.AccessGroup)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotAccessGroup)
	}

	cr.SetConditions(cpv1alpha1.Creating())
	createOptions := &iamagv2.CreateAccessGroupOptions{}
	if err := ibmcag.GenerateCreateAccessGroupOptions(cr.Spec.ForProvider, createOptions); err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreateAccessGroupOpts)
	}

	instance, _, err := c.client.IamAccessGroupsV2().CreateAccessGroup(createOptions)
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(resource.Ignore(ibmc.IsResourceInactive, err), errCreateAccessGroup)
	}

	meta.SetExternalName(cr, reference.FromPtrValue(instance.ID))
	return managed.ExternalCreation{ExternalNameAssigned: true}, nil
}

func (c *agExternal) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1alpha1.AccessGroup)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotAccessGroup)
	}

	id := cr.Status.AtProvider.ID
	eTag := ibmc.GetEtagAnnotation(cr)
	updInstanceOpts := &iamagv2.UpdateAccessGroupOptions{}
	if err := ibmcag.GenerateUpdateAccessGroupOptions(id, eTag, cr.Spec.ForProvider, updInstanceOpts); err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errUpdAccessGroup)
	}

	_, _, err := c.client.IamAccessGroupsV2().UpdateAccessGroup(updInstanceOpts)
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errUpdAccessGroup)
	}

	return managed.ExternalUpdate{}, nil
}

func (c *agExternal) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*v1alpha1.AccessGroup)
	if !ok {
		return errors.New(errNotAccessGroup)
	}

	cr.SetConditions(cpv1alpha1.Deleting())

	_, err := c.client.IamAccessGroupsV2().DeleteAccessGroup(&iamagv2.DeleteAccessGroupOptions{AccessGroupID: &cr.Status.AtProvider.ID})
	if err != nil {
		return errors.Wrap(resource.Ignore(ibmc.IsResourceGone, err), errDeleteAccessGroup)
	}
	return nil
}
