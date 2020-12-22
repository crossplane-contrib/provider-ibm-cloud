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

package ibmclouddatabasesv5

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

	icdv5 "github.com/IBM/experimental-go-sdk/ibmclouddatabasesv5"

	"github.com/crossplane-contrib/provider-ibm-cloud/apis/ibmclouddatabasesv5/v1alpha1"
	"github.com/crossplane-contrib/provider-ibm-cloud/apis/v1beta1"
	ibmc "github.com/crossplane-contrib/provider-ibm-cloud/pkg/clients"
	ibmcwl "github.com/crossplane-contrib/provider-ibm-cloud/pkg/clients/whitelist"
)

const (
	errNotWhitelist  = "managed resource is not a Whitelist custom resource"
	errWhiteListOpts = "error setting whitelist options"
)

// SetupWhitelist adds a controller that reconciles Whitelist managed resources.
func SetupWhitelist(mgr ctrl.Manager, l logging.Logger) error {
	name := managed.ControllerName(v1alpha1.WhitelistGroupKind)
	log := l.WithValues("Whitelist-controller", name)

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(v1alpha1.WhitelistGroupVersionKind),
		managed.WithExternalConnecter(&wlConnector{
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
		For(&v1alpha1.Whitelist{}).
		Complete(r)
}

// A wlConnector is expected to produce an ExternalClient when its Connect method
// is called.
type wlConnector struct {
	kube     client.Client
	usage    resource.Tracker
	clientFn func(optd ibmc.ClientOptions) (ibmc.ClientSession, error)
	logger   logging.Logger
}

// Connect produces an ExternalClient for IBM Cloud API
func (c *wlConnector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	opts, err := ibmc.GetAuthInfo(ctx, c.kube, mg)
	if err != nil {
		return nil, errors.Wrap(err, errGetAuth)
	}

	service, err := c.clientFn(opts)
	if err != nil {
		return nil, errors.Wrap(err, errNewClient)
	}

	return &wlExternal{client: service, kube: c.kube, logger: c.logger}, nil
}

// An wlExternal observes, then either creates, updates, or deletes an
// external resource to ensure it reflects the managed resource's desired state.
type wlExternal struct {
	client ibmc.ClientSession
	kube   client.Client
	logger logging.Logger
}

func (c *wlExternal) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) { // nolint:gocyclo
	cr, ok := mg.(*v1alpha1.Whitelist)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotWhitelist)
	}

	if meta.GetExternalName(cr) == "" {
		return managed.ExternalObservation{
			ResourceExists: false,
		}, nil
	}

	instance, _, err := c.client.IbmCloudDatabasesV5().GetWhitelist(&icdv5.GetWhitelistOptions{ID: reference.ToPtrValue(meta.GetExternalName(cr))})
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(resource.Ignore(ibmc.IsResourceNotFound, err), errGetInstanceFailed)
	}

	if len(instance.IpAddresses) == 0 {
		return managed.ExternalObservation{
			ResourceExists: false,
		}, nil
	}

	currentSpec := cr.Spec.ForProvider.DeepCopy()
	if err = ibmcwl.LateInitializeSpec(&cr.Spec.ForProvider, instance); err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errLateInitSpec)
	}
	if !cmp.Equal(currentSpec, &cr.Spec.ForProvider) {
		if err := c.kube.Update(ctx, cr); err != nil {
			return managed.ExternalObservation{}, errors.Wrap(err, errUpdateCR)
		}
	}

	cr.Status.AtProvider, err = ibmcwl.GenerateObservation(instance)
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errGenObservation)
	}

	if instance.IpAddresses != nil {
		cr.Status.SetConditions(cpv1alpha1.Available())
		cr.Status.AtProvider.State = string(cpv1alpha1.Available().Reason)
	} else {
		cr.Status.SetConditions(cpv1alpha1.Unavailable())
		cr.Status.AtProvider.State = string(cpv1alpha1.Unavailable().Reason)
	}

	upToDate, err := ibmcwl.IsUpToDate(meta.GetExternalName(cr), &cr.Spec.ForProvider, instance, c.logger)
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errCheckUpToDate)
	}

	return managed.ExternalObservation{
		ResourceExists:    true,
		ResourceUpToDate:  upToDate,
		ConnectionDetails: nil,
	}, nil
}

func (c *wlExternal) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1alpha1.Whitelist)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotWhitelist)
	}

	cr.SetConditions(cpv1alpha1.Creating())
	if cr.Spec.ForProvider.ID == nil {
		return managed.ExternalCreation{}, errors.New(errResNotAvailable)
	}

	opts := &icdv5.ReplaceWhitelistOptions{}
	err := ibmcwl.GenerateReplaceWhitelistOptions(reference.FromPtrValue(cr.Spec.ForProvider.ID), cr.Spec.ForProvider, opts)
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errWhiteListOpts)
	}

	_, _, err = c.client.IbmCloudDatabasesV5().ReplaceWhitelist(opts)
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errWhiteListOpts)
	}

	meta.SetExternalName(cr, reference.FromPtrValue(cr.Spec.ForProvider.ID))
	return managed.ExternalCreation{ExternalNameAssigned: true}, nil
}

func (c *wlExternal) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1alpha1.Whitelist)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotWhitelist)
	}

	opts := &icdv5.ReplaceWhitelistOptions{}
	err := ibmcwl.GenerateReplaceWhitelistOptions(meta.GetExternalName(cr), cr.Spec.ForProvider, opts)
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errWhiteListOpts)
	}

	_, _, err = c.client.IbmCloudDatabasesV5().ReplaceWhitelist(opts)
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errWhiteListOpts)
	}

	return managed.ExternalUpdate{}, nil
}

func (c *wlExternal) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*v1alpha1.Whitelist)
	if !ok {
		return errors.New(errNotWhitelist)
	}
	cr.SetConditions(cpv1alpha1.Deleting())

	opts := &icdv5.ReplaceWhitelistOptions{
		ID:          reference.ToPtrValue(meta.GetExternalName(cr)),
		IpAddresses: []icdv5.WhitelistEntry{},
	}
	_, _, err := c.client.IbmCloudDatabasesV5().ReplaceWhitelist(opts)
	if err != nil {
		return errors.Wrap(err, errWhiteListOpts)
	}

	return nil
}
