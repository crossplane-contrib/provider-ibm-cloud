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

package resourcecontrollerv2

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
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	rcv2 "github.com/IBM/platform-services-go-sdk/resourcecontrollerv2"

	"github.com/crossplane-contrib/provider-ibm-cloud/apis/resourcecontrollerv2/v1alpha1"
	"github.com/crossplane-contrib/provider-ibm-cloud/apis/v1beta1"
	ibmc "github.com/crossplane-contrib/provider-ibm-cloud/pkg/clients"
	rcv2c "github.com/crossplane-contrib/provider-ibm-cloud/pkg/clients/resourcecontrollerv2"
)

const (
	errNotResourceInstance   = "managed resource is not a ResourceInstance custom resource"
	errManagedUpdateFailed   = "cannot update ResourceInstance custom resource"
	errNewClient             = "cannot create new Client"
	errCreateRes             = "could not create resource instance"
	errDeleteRes             = "could not delete resource instance"
	errGetInstanceFailed     = "error getting instance"
	errFindInstances         = "error finding instances"
	errCheckUpToDate         = "cannot determine if instance is up to date"
	errGetAuth               = "error getting auth info"
	errGenObservation        = "error generating observation"
	errNamedInstanceExists   = "resource instance with the name %s already exists"
	errCreateResInstanceOpts = "error creating resource instance options"
	errUpdRes                = "error updating instance"
)

// Setup adds a controller that reconciles ResourceInstance managed resources.
func Setup(mgr ctrl.Manager, l logging.Logger) error {
	name := managed.ControllerName(v1alpha1.ResourceInstanceGroupKind)
	log := l.WithValues("resourceinstance-controller", name)

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(v1alpha1.ResourceInstanceGroupVersionKind),
		managed.WithExternalConnecter(&riConnector{
			kube:     mgr.GetClient(),
			usage:    resource.NewProviderConfigUsageTracker(mgr.GetClient(), &v1beta1.ProviderConfigUsage{}),
			clientFn: ibmc.NewClient,
			logger:   log}),
		managed.WithInitializers(managed.NewNameAsExternalName(mgr.GetClient())),
		managed.WithLogger(log),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))))

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&v1alpha1.ResourceInstance{}).
		Complete(r)
}

// A riConnector is expected to produce an ExternalClient when its Connect method
// is called.
type riConnector struct {
	kube     client.Client
	usage    resource.Tracker
	clientFn func(optd ibmc.ClientOptions) (ibmc.ClientSession, error)
	logger   logging.Logger
}

// Connect produces an ExternalClient for IBM Cloud API
func (c *riConnector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	opts, err := ibmc.GetAuthInfo(ctx, c.kube, mg)
	if err != nil {
		return nil, errors.Wrap(err, errGetAuth)
	}

	service, err := c.clientFn(opts)
	if err != nil {
		return nil, errors.Wrap(err, errNewClient)
	}

	return &riExternal{client: service, kube: c.kube, logger: c.logger}, nil
}

// An riExternal observes, then either creates, updates, or deletes an
// external resource to ensure it reflects the managed resource's desired state.
type riExternal struct {
	client ibmc.ClientSession
	kube   client.Client
	logger logging.Logger
}

func (c *riExternal) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) { // nolint:gocyclo
	cr, ok := mg.(*v1alpha1.ResourceInstance)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotResourceInstance)
	}

	// SDK throws a validation error if a nil ID is provided, need to catch this situation before call to SDK
	if cr.Status.AtProvider.ID == "" {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}

	instance, _, err := c.client.ResourceControllerV2().GetResourceInstance(&rcv2.GetResourceInstanceOptions{ID: &cr.Status.AtProvider.ID})
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(resource.Ignore(rcv2c.IsInstanceNotFound, err), errGetInstanceFailed)
	}

	if ibmc.StringValue(instance.State) == rcv2c.StatePendingReclamation {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}

	currentSpec := cr.Spec.ForProvider.DeepCopy()
	if err = rcv2c.LateInitializeSpec(c.client, &cr.Spec.ForProvider, instance); err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errManagedUpdateFailed)
	}
	if !cmp.Equal(currentSpec, &cr.Spec.ForProvider) {
		if err := c.kube.Update(ctx, cr); err != nil {
			return managed.ExternalObservation{}, errors.Wrap(err, errManagedUpdateFailed)
		}
	}

	cr.Status.AtProvider, err = rcv2c.GenerateObservation(c.client, instance)
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errGenObservation)
	}

	switch cr.Status.AtProvider.State {
	case rcv2c.StateActive:
		cr.Status.SetConditions(cpv1alpha1.Available())
	case rcv2c.StateInactive:
		cr.Status.SetConditions(cpv1alpha1.Creating())
	default:
		cr.Status.SetConditions(cpv1alpha1.Unavailable())
	}

	upToDate, err := rcv2c.IsUpToDate(c.client, meta.GetExternalName(cr), &cr.Spec.ForProvider, instance, c.logger)
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errCheckUpToDate)
	}

	return managed.ExternalObservation{
		ResourceExists:    true,
		ResourceUpToDate:  upToDate,
		ConnectionDetails: managed.ConnectionDetails{},
	}, nil
}

func (c *riExternal) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1alpha1.ResourceInstance)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotResourceInstance)
	}

	cr.SetConditions(cpv1alpha1.Creating())
	resInstanceOptions := &rcv2.CreateResourceInstanceOptions{}
	if err := rcv2c.GenerateCreateResourceInstanceOptions(c.client, meta.GetExternalName(cr), cr.Spec.ForProvider, resInstanceOptions); err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreateResInstanceOpts)
	}

	list, err := ibmc.FindResourceInstancesByName(c.client, meta.GetExternalName(cr))
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errFindInstances)
	}
	if len(list.Resources) > 0 {
		return managed.ExternalCreation{}, errors.Errorf(errNamedInstanceExists, meta.GetExternalName(cr))
	}

	instance, _, err := c.client.ResourceControllerV2().CreateResourceInstance(resInstanceOptions)
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreateRes)
	}

	cr.Status.AtProvider, err = rcv2c.GenerateObservation(c.client, instance)
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errGenObservation)
	}

	return managed.ExternalCreation{
		// Optionally return any details that may be required to connect to the
		// external resource. These will be stored as the connection secret.
		ConnectionDetails: managed.ConnectionDetails{},
	}, nil
}

func (c *riExternal) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1alpha1.ResourceInstance)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotResourceInstance)
	}

	id := cr.Status.AtProvider.ID
	updInstanceOpts := &rcv2.UpdateResourceInstanceOptions{}
	if err := rcv2c.GenerateUpdateResourceInstanceOptions(c.client, meta.GetExternalName(cr), id, cr.Spec.ForProvider, updInstanceOpts); err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errUpdRes)
	}

	_, _, err := c.client.ResourceControllerV2().UpdateResourceInstance(updInstanceOpts)
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errUpdRes)
	}

	if err = ibmc.UpdateResourceInstanceTags(c.client, cr.Status.AtProvider.Crn, cr.Spec.ForProvider.Tags); err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errUpdRes)
	}

	return managed.ExternalUpdate{
		// Optionally return any details that may be required to connect to the
		// external resource. These will be stored as the connection secret.
		ConnectionDetails: managed.ConnectionDetails{},
	}, nil
}

func (c *riExternal) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*v1alpha1.ResourceInstance)
	if !ok {
		return errors.New(errNotResourceInstance)
	}

	cr.SetConditions(cpv1alpha1.Deleting())

	_, err := c.client.ResourceControllerV2().DeleteResourceInstance(&rcv2.DeleteResourceInstanceOptions{ID: &cr.Status.AtProvider.ID})
	if err != nil {
		return errors.Wrap(resource.Ignore(rcv2c.IsInstancePendingReclamation, err), errDeleteRes)
	}
	return nil
}
