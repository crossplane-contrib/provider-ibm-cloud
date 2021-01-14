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
	ibmcasg "github.com/crossplane-contrib/provider-ibm-cloud/pkg/clients/autoscalinggroup"
)

const (
	errNotAutoscalingGroup = "managed resource is not an AutoscalingGroup custom resource"
)

// SetupAutoscalingGroup adds a controller that reconciles AutoscalingGroup managed resources.
func SetupAutoscalingGroup(mgr ctrl.Manager, l logging.Logger) error {
	name := managed.ControllerName(v1alpha1.AutoscalingGroupKind)
	log := l.WithValues("AutoscalingGroup-controller", name)

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(v1alpha1.AutoscalingGroupGroupVersionKind),
		managed.WithExternalConnecter(&asgConnector{
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
		For(&v1alpha1.AutoscalingGroup{}).
		Complete(r)
}

// A asgConnector is expected to produce an ExternalClient when its Connect method
// is called.
type asgConnector struct {
	kube     client.Client
	usage    resource.Tracker
	clientFn func(optd ibmc.ClientOptions) (ibmc.ClientSession, error)
	logger   logging.Logger
}

// Connect produces an ExternalClient for IBM Cloud API
func (c *asgConnector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	opts, err := ibmc.GetAuthInfo(ctx, c.kube, mg)
	if err != nil {
		return nil, errors.Wrap(err, errGetAuth)
	}

	service, err := c.clientFn(opts)
	if err != nil {
		return nil, errors.Wrap(err, errNewClient)
	}

	return &asgExternal{client: service, kube: c.kube, logger: c.logger}, nil
}

// An asgExternal observes, then either creates, updates, or deletes an
// external resource to ensure it reflects the managed resource's desired state.
type asgExternal struct {
	client ibmc.ClientSession
	kube   client.Client
	logger logging.Logger
}

func (c *asgExternal) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) { // nolint:gocyclo
	cr, ok := mg.(*v1alpha1.AutoscalingGroup)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotAutoscalingGroup)
	}

	// since we do not really delete an external resource but rather we have a configuration on an existing service
	// we need to look at the deletion timestamp to figure out if the scaling group config was deleted.
	if meta.GetExternalName(cr) == "" || cr.DeletionTimestamp != nil {
		return managed.ExternalObservation{
			ResourceExists: false,
		}, nil
	}

	instance, _, err := c.client.IbmCloudDatabasesV5().GetAutoscalingConditions(&icdv5.GetAutoscalingConditionsOptions{
		ID:      reference.ToPtrValue(meta.GetExternalName(cr)),
		GroupID: reference.ToPtrValue(ibmcasg.MemberGroupID),
	})
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(resource.Ignore(ibmc.IsResourceNotFound, err), errGetInstanceFailed)
	}

	if instance.Autoscaling == nil {
		return managed.ExternalObservation{
			ResourceExists: false,
		}, nil
	}

	currentSpec := cr.Spec.ForProvider.DeepCopy()
	if err = ibmcasg.LateInitializeSpec(&cr.Spec.ForProvider, instance.Autoscaling); err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errLateInitSpec)
	}
	if !cmp.Equal(currentSpec, &cr.Spec.ForProvider) {
		if err := c.kube.Update(ctx, cr); err != nil {
			return managed.ExternalObservation{}, errors.Wrap(err, errUpdateCR)
		}
	}

	cr.Status.AtProvider, err = ibmcasg.GenerateObservation(instance.Autoscaling)
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errGenObservation)
	}

	cr.Status.SetConditions(cpv1alpha1.Available())
	cr.Status.AtProvider.State = string(cpv1alpha1.Available().Reason)

	upToDate, err := ibmcasg.IsUpToDate(meta.GetExternalName(cr), &cr.Spec.ForProvider, instance.Autoscaling, c.logger)
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errCheckUpToDate)
	}

	return managed.ExternalObservation{
		ResourceExists:    true,
		ResourceUpToDate:  upToDate,
		ConnectionDetails: nil,
	}, nil
}

func (c *asgExternal) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1alpha1.AutoscalingGroup)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotAutoscalingGroup)
	}

	cr.SetConditions(cpv1alpha1.Creating())
	if cr.Spec.ForProvider.ID == nil {
		return managed.ExternalCreation{}, errors.New(errResNotAvailable)
	}

	meta.SetExternalName(cr, reference.FromPtrValue(cr.Spec.ForProvider.ID))
	return managed.ExternalCreation{ExternalNameAssigned: true}, nil
}

func (c *asgExternal) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1alpha1.AutoscalingGroup)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotAutoscalingGroup)
	}

	opts := &icdv5.SetAutoscalingConditionsOptions{}
	err := ibmcasg.GenerateSetAutoscalingConditionsOptions(meta.GetExternalName(cr), cr.Spec.ForProvider, opts)
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errSetOpts)
	}

	_, resp, err := c.client.IbmCloudDatabasesV5().SetAutoscalingConditions(opts)

	if err != nil {
		return managed.ExternalUpdate{}, ibmc.ExtractErrorMessage(resp, err)
	}

	return managed.ExternalUpdate{}, nil
}

func (c *asgExternal) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*v1alpha1.AutoscalingGroup)
	if !ok {
		return errors.New(errNotAutoscalingGroup)
	}
	cr.SetConditions(cpv1alpha1.Deleting())
	return nil
}
