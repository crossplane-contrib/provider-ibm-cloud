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
	"github.com/crossplane/crossplane-runtime/pkg/reference"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	rcv2 "github.com/IBM/platform-services-go-sdk/resourcecontrollerv2"

	"github.com/crossplane-contrib/provider-ibm-cloud/apis/resourcecontrollerv2/v1alpha1"
	"github.com/crossplane-contrib/provider-ibm-cloud/apis/v1beta1"
	ibmc "github.com/crossplane-contrib/provider-ibm-cloud/pkg/clients"
	ibmcrk "github.com/crossplane-contrib/provider-ibm-cloud/pkg/clients/resourcekey"
)

const (
	errNotResourceKey = "managed resource is not a ResourceKey custom resource"
	errGetConnDetails = "error getting connection details"
)

// SetupResourceKey adds a controller that reconciles ResourceKey managed resources.
func SetupResourceKey(mgr ctrl.Manager, l logging.Logger) error {
	name := managed.ControllerName(v1alpha1.ResourceKeyGroupKind)
	log := l.WithValues("ResourceKey-controller", name)

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(v1alpha1.ResourceKeyGroupVersionKind),
		managed.WithExternalConnecter(&rkConnector{
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
		For(&v1alpha1.ResourceKey{}).
		Complete(r)
}

// A rkConnector is expected to produce an ExternalClient when its Connect method
// is called.
type rkConnector struct {
	kube     client.Client
	usage    resource.Tracker
	clientFn func(optd ibmc.ClientOptions) (ibmc.ClientSession, error)
	logger   logging.Logger
}

// Connect produces an ExternalClient for IBM Cloud API
func (c *rkConnector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	opts, err := ibmc.GetAuthInfo(ctx, c.kube, mg)
	if err != nil {
		return nil, errors.Wrap(err, errGetAuth)
	}

	service, err := c.clientFn(opts)
	if err != nil {
		return nil, errors.Wrap(err, errNewClient)
	}

	return &rkExternal{client: service, kube: c.kube, logger: c.logger}, nil
}

// An rkExternal observes, then either creates, updates, or deletes an
// external resource to ensure it reflects the managed resource's desired state.
type rkExternal struct {
	client ibmc.ClientSession
	kube   client.Client
	logger logging.Logger
}

func (c *rkExternal) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) { // nolint:gocyclo
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
		return managed.ExternalObservation{}, errors.Wrap(resource.Ignore(ibmc.IsResourceNotFound, err), errGetInstanceFailed)
	}

	if reference.FromPtrValue(instance.State) == ibmcrk.StateRemoved {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}

	currentSpec := cr.Spec.ForProvider.DeepCopy()
	if err = ibmcrk.LateInitializeSpec(&cr.Spec.ForProvider, instance); err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errManagedUpdateFailed)
	}
	if !cmp.Equal(currentSpec, &cr.Spec.ForProvider) {
		if err := c.kube.Update(ctx, cr); err != nil {
			return managed.ExternalObservation{}, errors.Wrap(err, errManagedUpdateFailed)
		}
	}

	cr.Status.AtProvider, err = ibmcrk.GenerateObservation(instance)
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errGenObservation)
	}

	switch cr.Status.AtProvider.State {
	case ibmcrk.StateActive:
		cr.Status.SetConditions(cpv1alpha1.Available())
	case ibmcrk.StateInactive:
		cr.Status.SetConditions(cpv1alpha1.Creating())
	default:
		cr.Status.SetConditions(cpv1alpha1.Unavailable())
	}

	upToDate, err := ibmcrk.IsUpToDate(meta.GetExternalName(cr), &cr.Spec.ForProvider, instance, c.logger)
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errCheckUpToDate)
	}

	cd, err := ibmcrk.GetConnectionDetails(cr, instance)
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errGetConnDetails)
	}

	return managed.ExternalObservation{
		ResourceExists:    true,
		ResourceUpToDate:  upToDate,
		ConnectionDetails: cd,
	}, nil
}

func (c *rkExternal) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1alpha1.ResourceKey)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotResourceKey)
	}

	cr.SetConditions(cpv1alpha1.Creating())
	resInstanceOptions := &rcv2.CreateResourceKeyOptions{}
	if err := ibmcrk.GenerateCreateResourceKeyOptions(meta.GetExternalName(cr), cr.Spec.ForProvider, resInstanceOptions); err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreateResInstanceOpts)
	}

	instance, _, err := c.client.ResourceControllerV2().CreateResourceKey(resInstanceOptions)
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(resource.Ignore(ibmc.IsResourceInactive, err), errCreateRes)
	}

	meta.SetExternalName(cr, reference.FromPtrValue(instance.ID))
	return managed.ExternalCreation{ExternalNameAssigned: true}, nil
}

func (c *rkExternal) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1alpha1.ResourceKey)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotResourceKey)
	}

	id := cr.Status.AtProvider.ID
	updInstanceOpts := &rcv2.UpdateResourceKeyOptions{}
	if err := ibmcrk.GenerateUpdateResourceKeyOptions(meta.GetExternalName(cr), id, cr.Spec.ForProvider, updInstanceOpts); err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errUpdRes)
	}

	_, _, err := c.client.ResourceControllerV2().UpdateResourceKey(updInstanceOpts)
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errUpdRes)
	}

	return managed.ExternalUpdate{}, nil
}

func (c *rkExternal) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*v1alpha1.ResourceKey)
	if !ok {
		return errors.New(errNotResourceKey)
	}

	cr.SetConditions(cpv1alpha1.Deleting())

	_, err := c.client.ResourceControllerV2().DeleteResourceKey(&rcv2.DeleteResourceKeyOptions{ID: &cr.Status.AtProvider.ID})
	if err != nil {
		return errors.Wrap(resource.Ignore(ibmc.IsResourceGone, err), errDeleteRes)
	}
	return nil
}
