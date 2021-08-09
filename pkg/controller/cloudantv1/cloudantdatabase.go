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

package cloudantv1

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

	cv1 "github.com/IBM/cloudant-go-sdk/cloudantv1"

	"github.com/crossplane-contrib/provider-ibm-cloud/apis/cloudantv1/v1alpha1"
	"github.com/crossplane-contrib/provider-ibm-cloud/apis/v1beta1"
	ibmc "github.com/crossplane-contrib/provider-ibm-cloud/pkg/clients"
	ibmccdb "github.com/crossplane-contrib/provider-ibm-cloud/pkg/clients/cloudantdatabase"
)

const (
	errNotCloudantDatabase        = "managed resource is not a CloudantDatabase custom resource"
	errCreateCloudantDatabase     = "could not create CloudantDatabase"
	errDeleteCloudantDatabase     = "could not delete CloudantDatabase"
	errGetCloudantDatabaseFailed  = "error getting CloudantDatabase"
	errCreateCloudantDatabaseOpts = "error creating CloudantDatabase"
)

// SetupCloudantDatabase adds a controller that reconciles CloudantDatabase managed resources.
func SetupCloudantDatabase(mgr ctrl.Manager, l logging.Logger) error {
	name := managed.ControllerName(v1alpha1.CloudantDatabaseGroupKind)
	log := l.WithValues("cloudantdatabase-controller", name)

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(v1alpha1.CloudantDatabaseGroupVersionKind),
		managed.WithExternalConnecter(&cloudantdatabaseConnector{
			kube:     mgr.GetClient(),
			usage:    resource.NewProviderConfigUsageTracker(mgr.GetClient(), &v1beta1.ProviderConfigUsage{}),
			clientFn: ibmc.NewClient,
			logger:   log}),
		managed.WithInitializers(managed.NewDefaultProviderConfig(mgr.GetClient())),
		managed.WithLogger(log),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))))

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&v1alpha1.CloudantDatabase{}).
		Complete(r)
}

// A cloudantdatabaseConnector is expected to produce an ExternalClient when its Connect method
// is called.
type cloudantdatabaseConnector struct {
	kube     client.Client
	usage    resource.Tracker
	clientFn func(optd ibmc.ClientOptions) (ibmc.ClientSession, error)
	logger   logging.Logger
}

// Connect produces an ExternalClient for IBM Cloud API
func (c *cloudantdatabaseConnector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	opts, err := ibmc.GetAuthInfo(ctx, c.kube, mg)
	if err != nil {
		return nil, errors.Wrap(err, ibmc.ErrGetAuth)
	}

	cr, ok := mg.(*v1alpha1.CloudantDatabase)
	if !ok {
		return nil, errors.New(errNotCloudantDatabase)
	}

	opts.URL = reference.FromPtrValue(cr.Spec.ForProvider.CloudantAdminURL)

	service, err := c.clientFn(opts)
	if err != nil {
		return nil, errors.Wrap(err, ibmc.ErrNewClient)
	}

	return &cloudantdatabaseExternal{client: service, kube: c.kube, logger: c.logger}, nil
}

// An cloudantdatabaseExternal observes, then either creates, updates, or deletes an
// external resource to ensure it reflects the managed resource's desired state.
type cloudantdatabaseExternal struct {
	client ibmc.ClientSession
	kube   client.Client
	logger logging.Logger
}

func (c *cloudantdatabaseExternal) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) { // nolint:gocyclo
	cr, ok := mg.(*v1alpha1.CloudantDatabase)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotCloudantDatabase)
	}

	if meta.GetExternalName(cr) == "" {
		return managed.ExternalObservation{
			ResourceExists: false,
		}, nil
	}

	instance, _, err := c.client.CloudantV1().GetDatabaseInformation(&cv1.GetDatabaseInformationOptions{Db: reference.ToPtrValue(meta.GetExternalName(cr))})
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(resource.Ignore(ibmc.IsResourceNotFound, err), errGetCloudantDatabaseFailed)
	}

	currentSpec := cr.Spec.ForProvider.DeepCopy()
	if err = ibmccdb.LateInitializeSpec(&cr.Spec.ForProvider, instance); err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, ibmc.ErrManagedUpdateFailed)
	}
	if !cmp.Equal(currentSpec, &cr.Spec.ForProvider) {
		if err := c.kube.Update(ctx, cr); err != nil {
			return managed.ExternalObservation{}, errors.Wrap(err, ibmc.ErrManagedUpdateFailed)
		}
	}

	cr.Status.AtProvider, err = ibmccdb.GenerateObservation(instance)
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, ibmc.ErrGenObservation)
	}

	cr.Status.AtProvider.State = "active"

	switch cr.Status.AtProvider.State {
	case "active":
		cr.Status.SetConditions(runtimev1alpha1.Available())
	default:
		cr.Status.SetConditions(runtimev1alpha1.Unavailable())
	}
	cr.Status.SetConditions(runtimev1alpha1.Available())

	// have to ensure isuptodate always return true ?? just a note
	upToDate, err := ibmccdb.IsUpToDate(&cr.Spec.ForProvider, instance, c.logger)
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, ibmc.ErrCheckUpToDate)
	}
	return managed.ExternalObservation{
		ResourceExists:    true,
		ResourceUpToDate:  upToDate,
		ConnectionDetails: nil,
	}, nil
}

func (c *cloudantdatabaseExternal) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1alpha1.CloudantDatabase)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotCloudantDatabase)
	}

	cr.SetConditions(runtimev1alpha1.Creating())
	resInstanceOptions := &cv1.PutDatabaseOptions{}
	if err := ibmccdb.GenerateCreateCloudantDatabaseOptions(cr.Spec.ForProvider, resInstanceOptions); err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreateCloudantDatabaseOpts)
	}

	_, _, err := c.client.CloudantV1().PutDatabase(resInstanceOptions)
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreateCloudantDatabase)
	}

	meta.SetExternalName(cr, cr.Spec.ForProvider.Db)

	return managed.ExternalCreation{ExternalNameAssigned: true}, nil
}

func (c *cloudantdatabaseExternal) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	return managed.ExternalUpdate{}, nil
}

func (c *cloudantdatabaseExternal) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*v1alpha1.CloudantDatabase)
	if !ok {
		return errors.New(errNotCloudantDatabase)
	}

	cr.SetConditions(runtimev1alpha1.Deleting())

	_, _, err := c.client.CloudantV1().DeleteDatabase(&cv1.DeleteDatabaseOptions{Db: reference.ToPtrValue(meta.GetExternalName(cr))})
	if err != nil {
		return errors.Wrap(resource.Ignore(ibmc.IsResourceGone, err), errDeleteCloudantDatabase)
	}
	cr.Status.AtProvider.State = "terminating"

	return nil
}
