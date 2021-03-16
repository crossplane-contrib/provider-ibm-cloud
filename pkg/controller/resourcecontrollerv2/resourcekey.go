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
	resclient "github.com/crossplane-contrib/provider-ibm-cloud/pkg/clients/resourcekey"
)

const (
	errNotResourceKey        = "managed resource is not a ResourceKey custom resource"
	errCreateResourceKey     = "could not create ResourceKey"
	errDeleteResourceKey     = "could not delete ResourceKey"
	errGetResourceKeyFailed  = "error getting ResourceKey"
	errCreateResourceKeyOpts = "error creating ResourceKey"
	errUpdResourceKey        = "error updating ResourceKey"
)

// SetupResourceKey adds a controller that reconciles ResourceKey managed resources.
func SetupResourceKey(mgr ctrl.Manager, l logging.Logger) error {
	name := managed.ControllerName(v1alpha1.ResourceKeyGroupKind)
	log := l.WithValues("resourcekey-controller", name)

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(v1alpha1.ResourceKeyGroupVersionKind),
		managed.WithExternalConnecter(&resourcekeyConnector{
			kube:     mgr.GetClient(),
			usage:    resource.NewProviderConfigUsageTracker(mgr.GetClient(), &v1beta1.ProviderConfigUsage{}),
			clientFn: ibmc.NewClient,
			logger:   log}),
		managed.WithInitializers(managed.NewDefaultProviderConfig(mgr.GetClient())),
		managed.WithLogger(log),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))))

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&v1alpha1.ResourceKey{}).
		Complete(r)
}

// A resourcekeyConnector is expected to produce an ExternalClient when its Connect method
// is called.
type resourcekeyConnector struct {
	kube     client.Client
	usage    resource.Tracker
	clientFn func(optd ibmc.ClientOptions) (ibmc.ClientSession, error)
	logger   logging.Logger
}

// Connect produces an ExternalClient for IBM Cloud API
func (c *resourcekeyConnector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	opts, err := ibmc.GetAuthInfo(ctx, c.kube, mg)
	if err != nil {
		return nil, errors.Wrap(err, ibmc.ErrGetAuth)
	}

	service, err := c.clientFn(opts)
	if err != nil {
		return nil, errors.Wrap(err, ibmc.ErrNewClient)
	}

	return &resourcekeyExternal{client: service, kube: c.kube, logger: c.logger}, nil
}

// An resourcekeyExternal observes, then either creates, updates, or deletes an
// external resource to ensure it reflects the managed resource's desired state.
type resourcekeyExternal struct {
	client ibmc.ClientSession
	kube   client.Client
	logger logging.Logger
}

func (c *resourcekeyExternal) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) { // nolint:gocyclo
	cr, ok := mg.(*v1alpha1.ResourceKey)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotResourceKey)
	}

	if meta.GetExternalName(cr) == "" {
		return managed.ExternalObservation{
			ResourceExists: false,
		}, nil
	}

	instance, _, err := c.client.ResourceControllerV2().GetResourceKey(&rcv2.GetResourceKeyOptions{ID: reference.ToPtrValue(meta.GetExternalName(cr))})
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(resource.Ignore(ibmc.IsResourceNotFound, err), errGetResourceKeyFailed)
	}

	if !(reference.FromPtrValue(instance.State) == "active" ||
		reference.FromPtrValue(instance.State) == "inactive") {
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
	default:
		cr.Status.SetConditions(runtimev1alpha1.Unavailable())
	}

	cr.Status.SetConditions(runtimev1alpha1.Available())

	upToDate, err := resclient.IsUpToDate(c.client, &cr.Spec.ForProvider, instance, c.logger)
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, ibmc.ErrCheckUpToDate)
	}

	cd, err := resclient.GetConnectionDetails(cr, instance)
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, ibmc.ErrGetConnDetails)
	}

	return managed.ExternalObservation{
		ResourceExists:    true,
		ResourceUpToDate:  upToDate,
		ConnectionDetails: cd,
	}, nil
}

func (c *resourcekeyExternal) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1alpha1.ResourceKey)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotResourceKey)
	}

	cr.SetConditions(runtimev1alpha1.Creating())
	resInstanceOptions := &rcv2.CreateResourceKeyOptions{}
	if err := resclient.GenerateCreateResourceKeyOptions(c.client, cr.Spec.ForProvider, resInstanceOptions); err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreateResourceKeyOpts)
	}

	instance, _, err := c.client.ResourceControllerV2().CreateResourceKey(resInstanceOptions)
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreateResourceKey)
	}

	meta.SetExternalName(cr, reference.FromPtrValue(instance.ID))
	return managed.ExternalCreation{ExternalNameAssigned: true}, nil
}

func (c *resourcekeyExternal) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1alpha1.ResourceKey)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotResourceKey)
	}

	id := cr.Status.AtProvider.ID
	updInstanceOpts := &rcv2.UpdateResourceKeyOptions{}
	if err := resclient.GenerateUpdateResourceKeyOptions(c.client, id, cr.Spec.ForProvider, updInstanceOpts); err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errUpdResourceKey)
	}

	_, _, err := c.client.ResourceControllerV2().UpdateResourceKey(updInstanceOpts)
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errUpdResourceKey)
	}

	return managed.ExternalUpdate{}, nil
}

func (c *resourcekeyExternal) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*v1alpha1.ResourceKey)
	if !ok {
		return errors.New(errNotResourceKey)
	}

	cr.SetConditions(runtimev1alpha1.Deleting())

	_, err := c.client.ResourceControllerV2().DeleteResourceKey(&rcv2.DeleteResourceKeyOptions{ID: &cr.Status.AtProvider.ID})
	if err != nil {
		return errors.Wrap(resource.Ignore(ibmc.IsResourceGone, err), errDeleteResourceKey)
	}
	return nil
}
