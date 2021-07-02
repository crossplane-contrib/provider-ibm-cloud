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

package eventstreamsadminv1

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

	arv1 "github.com/IBM/eventstreams-go-sdk/pkg/adminrestv1"

	"github.com/crossplane-contrib/provider-ibm-cloud/apis/eventstreamsadminv1/v1alpha1"
	"github.com/crossplane-contrib/provider-ibm-cloud/apis/v1beta1"
	ibmc "github.com/crossplane-contrib/provider-ibm-cloud/pkg/clients"
	ibmct "github.com/crossplane-contrib/provider-ibm-cloud/pkg/clients/topic"
)

const (
	errNotTopic        = "managed resource is not a Topic custom resource"
	errCreateTopic     = "could not create Topic"
	errDeleteTopic     = "could not delete Topic"
	errGetTopicFailed  = "error getting Topic"
	errCreateTopicOpts = "error creating Topic"
	errUpdTopic        = "error updating Topic"
)

// SetupTopic adds a controller that reconciles Topic managed resources.
func SetupTopic(mgr ctrl.Manager, l logging.Logger) error {
	name := managed.ControllerName(v1alpha1.TopicGroupKind)
	log := l.WithValues("topic-controller", name)

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(v1alpha1.TopicGroupVersionKind),
		managed.WithExternalConnecter(&topicConnector{
			kube:     mgr.GetClient(),
			usage:    resource.NewProviderConfigUsageTracker(mgr.GetClient(), &v1beta1.ProviderConfigUsage{}),
			clientFn: ibmc.NewClient,
			logger:   log}),
		managed.WithInitializers(managed.NewDefaultProviderConfig(mgr.GetClient())),
		managed.WithLogger(log),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))))

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&v1alpha1.Topic{}).
		Complete(r)
}

// A topicConnector is expected to produce an ExternalClient when its Connect method
// is called.
type topicConnector struct {
	kube     client.Client
	usage    resource.Tracker
	clientFn func(optd ibmc.ClientOptions) (ibmc.ClientSession, error)
	logger   logging.Logger
}

// Connect produces an ExternalClient for IBM Cloud API
func (c *topicConnector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	opts, err := ibmc.GetAuthInfo(ctx, c.kube, mg)
	if err != nil {
		return nil, errors.Wrap(err, ibmc.ErrGetAuth)
	}

	cr, ok := mg.(*v1alpha1.Topic)
	if !ok {
		return nil, errors.New(errNotTopic)
	}

	opts.URL = reference.FromPtrValue(cr.Spec.ForProvider.KafkaAdminURL)

	service, err := c.clientFn(opts)
	if err != nil {
		return nil, errors.Wrap(err, ibmc.ErrNewClient)
	}

	return &topicExternal{client: service, kube: c.kube, logger: c.logger}, nil
}

// An topicExternal observes, then either creates, updates, or deletes an
// external resource to ensure it reflects the managed resource's desired state.
type topicExternal struct {
	client ibmc.ClientSession
	kube   client.Client
	logger logging.Logger
}

func (c *topicExternal) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) { // nolint:gocyclo
	cr, ok := mg.(*v1alpha1.Topic)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotTopic)
	}

	if meta.GetExternalName(cr) == "" {
		return managed.ExternalObservation{
			ResourceExists: false,
		}, nil
	}

	instance, _, err := c.client.AdminrestV1().GetTopic(&arv1.GetTopicOptions{TopicName: reference.ToPtrValue(meta.GetExternalName(cr))})
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(resource.Ignore(ibmc.IsResourceNotFound, err), errGetTopicFailed)
	}
	currentSpec := cr.Spec.ForProvider.DeepCopy()
	if err = ibmct.LateInitializeSpec(&cr.Spec.ForProvider, instance); err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, ibmc.ErrManagedUpdateFailed)
	}
	if !cmp.Equal(currentSpec, &cr.Spec.ForProvider) {
		if err := c.kube.Update(ctx, cr); err != nil {
			return managed.ExternalObservation{}, errors.Wrap(err, ibmc.ErrManagedUpdateFailed)
		}
	}

	cr.Status.AtProvider, err = ibmct.GenerateObservation(instance)
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

	upToDate, err := ibmct.IsUpToDate(&cr.Spec.ForProvider, instance, c.logger)
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, ibmc.ErrCheckUpToDate)
	}
	return managed.ExternalObservation{
		ResourceExists:    true,
		ResourceUpToDate:  upToDate,
		ConnectionDetails: nil,
	}, nil
}

func (c *topicExternal) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1alpha1.Topic)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotTopic)
	}

	cr.SetConditions(runtimev1alpha1.Creating())
	resInstanceOptions := &arv1.CreateTopicOptions{}
	if err := ibmct.GenerateCreateTopicOptions(cr.Spec.ForProvider, resInstanceOptions); err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreateTopicOpts)
	}

	_, err := c.client.AdminrestV1().CreateTopic(resInstanceOptions)
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreateTopic)
	}

	meta.SetExternalName(cr, cr.Spec.ForProvider.Name)

	return managed.ExternalCreation{ExternalNameAssigned: true}, nil
}

func (c *topicExternal) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1alpha1.Topic)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotTopic)
	}

	instance, _, err := c.client.AdminrestV1().GetTopic(&arv1.GetTopicOptions{TopicName: reference.ToPtrValue(meta.GetExternalName(cr))})
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(resource.Ignore(ibmc.IsResourceNotFound, err), errGetTopicFailed)
	}

	updInstanceOpts := &arv1.UpdateTopicOptions{}
	if err = ibmct.GenerateUpdateTopicOptions(instance.Partitions, cr.Spec.ForProvider, updInstanceOpts); err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errUpdTopic)
	}
	_, err = c.client.AdminrestV1().UpdateTopic(updInstanceOpts)
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errUpdTopic)
	}
	return managed.ExternalUpdate{}, nil
}

func (c *topicExternal) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*v1alpha1.Topic)
	if !ok {
		return errors.New(errNotTopic)
	}

	cr.SetConditions(runtimev1alpha1.Deleting())

	_, err := c.client.AdminrestV1().DeleteTopic(&arv1.DeleteTopicOptions{TopicName: reference.ToPtrValue(meta.GetExternalName(cr))})
	if err != nil {
		return errors.Wrap(resource.Ignore(ibmc.IsResourceGone, err), errDeleteTopic)
	}
	cr.Status.AtProvider.State = "terminating"

	return nil
}
