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

package resourcecontrollerv2

import (
	"context"

	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/reference"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	rcv2 "github.com/IBM/platform-services-go-sdk/resourcecontrollerv2"

	"github.com/crossplane-contrib/provider-ibm-cloud/apis/resourcecontrollerv2/v1alpha1"
	"github.com/crossplane-contrib/provider-ibm-cloud/apis/v1beta1"
	ibmc "github.com/crossplane-contrib/provider-ibm-cloud/pkg/clients"
	resclient "github.com/crossplane-contrib/provider-ibm-cloud/pkg/clients/resourceinstance"
)

const (
	errNotResourceInstance        = "managed resource is not a ResourceInstance custom resource"
	errCreateResourceInstance     = "could not create ResourceInstance"
	errDeleteResourceInstance     = "could not delete ResourceInstance"
	errGetResourceInstanceFailed  = "error getting ResourceInstance"
	errCreateResourceInstanceOpts = "error creating ResourceInstance"
	errUpdResourceInstance        = "error updating ResourceInstance"
)

// SetupResourceInstance adds a controller that reconciles ResourceInstance managed resources.
func SetupResourceInstance(mgr ctrl.Manager, l logging.Logger) error {
	name := managed.ControllerName(v1alpha1.ResourceInstanceGroupKind)
	log := l.WithValues("resourceinstance-controller", name)

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(v1alpha1.ResourceInstanceGroupVersionKind),
		managed.WithExternalConnecter(&resourceinstanceConnector{
			kube:     mgr.GetClient(),
			usage:    resource.NewProviderConfigUsageTracker(mgr.GetClient(), &v1beta1.ProviderConfigUsage{}),
			clientFn: ibmc.NewClient,
			logger:   log}),
		managed.WithInitializers(managed.NewDefaultProviderConfig(mgr.GetClient())),
		managed.WithLogger(log),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))))

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&v1alpha1.ResourceInstance{}).
		Complete(r)
}

// A resourceinstanceConnector is expected to produce an ExternalClient when its Connect method
// is called.
type resourceinstanceConnector struct {
	kube     client.Client
	usage    resource.Tracker
	clientFn func(optd ibmc.ClientOptions) (ibmc.ClientSession, error)
	logger   logging.Logger
}

// Connect produces an ExternalClient for IBM Cloud API
func (c *resourceinstanceConnector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	opts, err := ibmc.GetAuthInfo(ctx, c.kube, mg)
	if err != nil {
		return nil, errors.Wrap(err, ibmc.ErrGetAuth)
	}

	service, err := c.clientFn(opts)
	if err != nil {
		return nil, errors.Wrap(err, ibmc.ErrNewClient)
	}

	return &resourceinstanceExternal{client: service, kube: c.kube, logger: c.logger}, nil
}

// An resourceinstanceExternal observes, then either creates, updates, or deletes an
// external resource to ensure it reflects the managed resource's desired state.
type resourceinstanceExternal struct {
	client ibmc.ClientSession
	kube   client.Client
	logger logging.Logger
}

func (c *resourceinstanceExternal) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) { // nolint:gocyclo
	cr, ok := mg.(*v1alpha1.ResourceInstance)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotResourceInstance)
	}

	if meta.GetExternalName(cr) == "" {
		return managed.ExternalObservation{
			ResourceExists: false,
		}, nil
	}

	instance, _, err := c.client.ResourceControllerV2().GetResourceInstance(&rcv2.GetResourceInstanceOptions{ID: reference.ToPtrValue(meta.GetExternalName(cr))})
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(resource.Ignore(ibmc.IsResourceNotFound, err), errGetResourceInstanceFailed)
	}

	if !(reference.FromPtrValue(instance.State) == "active" ||
		reference.FromPtrValue(instance.State) == "inactive" ||
		reference.FromPtrValue(instance.State) == "provisioning") {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}

	currentSpec := cr.Spec.ForProvider.DeepCopy()
	if err = resclient.LateInitializeSpec(c.client, &cr.Spec.ForProvider, instance); err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, ibmc.ErrManagedUpdateFailed)
	}
	if !cmp.Equal(currentSpec, &cr.Spec.ForProvider) {
		if err := c.kube.Update(ctx, cr); err != nil {
			return managed.ExternalObservation{}, errors.Wrap(err, ibmc.ErrManagedUpdateFailed)
		}
	}

	cr.Status.AtProvider, err = resclient.GenerateObservation(c.client, instance)
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, ibmc.ErrGenObservation)
	}

	switch cr.Status.AtProvider.State {
	case "active":
		cr.Status.SetConditions(runtimev1alpha1.Available())
	case "inactive":
		cr.Status.SetConditions(runtimev1alpha1.Creating())
	case "provisioning":
		cr.Status.SetConditions(runtimev1alpha1.Creating())
	default:
		cr.Status.SetConditions(runtimev1alpha1.Unavailable())
	}

	cr.Status.SetConditions(runtimev1alpha1.Available())

	upToDate, err := resclient.IsUpToDate(c.client, &cr.Spec.ForProvider, instance, c.logger)
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, ibmc.ErrCheckUpToDate)
	}

	return managed.ExternalObservation{
		ResourceExists:    true,
		ResourceUpToDate:  upToDate,
		ConnectionDetails: nil,
	}, nil
}

func (c *resourceinstanceExternal) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1alpha1.ResourceInstance)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotResourceInstance)
	}

	cr.SetConditions(runtimev1alpha1.Creating())
	resInstanceOptions := &rcv2.CreateResourceInstanceOptions{}
	if err := resclient.GenerateCreateResourceInstanceOptions(c.client, cr.Spec.ForProvider, resInstanceOptions); err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreateResourceInstanceOpts)
	}

	instance, _, err := c.client.ResourceControllerV2().CreateResourceInstance(resInstanceOptions)
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreateResourceInstance)
	}

	meta.SetExternalName(cr, reference.FromPtrValue(instance.ID))
	return managed.ExternalCreation{ExternalNameAssigned: true}, nil
}

func (c *resourceinstanceExternal) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1alpha1.ResourceInstance)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotResourceInstance)
	}

	id := cr.Status.AtProvider.ID
	updInstanceOpts := &rcv2.UpdateResourceInstanceOptions{}
	if err := resclient.GenerateUpdateResourceInstanceOptions(c.client, id, cr.Spec.ForProvider, updInstanceOpts); err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errUpdResourceInstance)
	}

	_, _, err := c.client.ResourceControllerV2().UpdateResourceInstance(updInstanceOpts)
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errUpdResourceInstance)
	}

	if err = ibmc.UpdateResourceInstanceTags(c.client, cr.Status.AtProvider.CRN, cr.Spec.ForProvider.Tags); err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errUpdResourceInstance)
	}

	return managed.ExternalUpdate{}, nil
}

func (c *resourceinstanceExternal) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*v1alpha1.ResourceInstance)
	if !ok {
		return errors.New(errNotResourceInstance)
	}

	cr.SetConditions(runtimev1alpha1.Deleting())

	_, err := c.client.ResourceControllerV2().DeleteResourceInstance(&rcv2.DeleteResourceInstanceOptions{ID: &cr.Status.AtProvider.ID})
	if err != nil {
		return errors.Wrap(resource.Ignore(ibmc.IsResourceGone, err), errDeleteResourceInstance)
	}
	return nil
}
