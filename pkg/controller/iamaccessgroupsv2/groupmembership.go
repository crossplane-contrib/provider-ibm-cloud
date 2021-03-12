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
	ibmcgm "github.com/crossplane-contrib/provider-ibm-cloud/pkg/clients/groupmembership"
)

const (
	errNotGroupMembership        = "managed resource is not a GroupMembership custom resource"
	errCreateGroupMembership     = "could not create access group"
	errDeleteGroupMembership     = "could not delete access group"
	errGetGroupMembershipFailed  = "error getting access group"
	errCreateGroupMembershipOpts = "error creating access group options"
	errUpdGroupMembership        = "error updating access group"
)

// SetupGroupMembership adds a controller that reconciles GroupMembership managed resources.
func SetupGroupMembership(mgr ctrl.Manager, l logging.Logger) error {
	name := managed.ControllerName(v1alpha1.GroupMembershipGroupKind)
	log := l.WithValues("GroupMembership-controller", name)

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(v1alpha1.GroupMembershipGroupVersionKind),
		managed.WithExternalConnecter(&gmConnector{
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
		For(&v1alpha1.GroupMembership{}).
		Complete(r)
}

// A gmConnector is expected to produce an ExternalClient when its Connect method
// is called.
type gmConnector struct {
	kube     client.Client
	usage    resource.Tracker
	clientFn func(optd ibmc.ClientOptions) (ibmc.ClientSession, error)
	logger   logging.Logger
}

// Connect produces an ExternalClient for IBM Cloud API
func (c *gmConnector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	opts, err := ibmc.GetAuthInfo(ctx, c.kube, mg)
	if err != nil {
		return nil, errors.Wrap(err, errGetAuth)
	}

	service, err := c.clientFn(opts)
	if err != nil {
		return nil, errors.Wrap(err, errNewClient)
	}

	return &gmExternal{client: service, kube: c.kube, logger: c.logger}, nil
}

// An gmExternal observes, then either creates, updates, or deletes an
// external resource to ensure it reflects the managed resource's desired state.
type gmExternal struct {
	client ibmc.ClientSession
	kube   client.Client
	logger logging.Logger
}

func (c *gmExternal) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) { // nolint:gocyclo
	cr, ok := mg.(*v1alpha1.GroupMembership)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotGroupMembership)
	}

	if meta.GetExternalName(cr) == "" {
		return managed.ExternalObservation{
			ResourceExists: false,
		}, nil
	}

	instance, resp, err := c.client.IamAccessGroupsV2().ListAccessGroupMembers(&iamagv2.ListAccessGroupMembersOptions{AccessGroupID: reference.ToPtrValue(meta.GetExternalName(cr))})
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(resource.Ignore(ibmc.IsResourceNotFound, err), errGetGroupMembershipFailed)
	}
	ibmc.SetEtagAnnotation(cr, ibmc.GetEtag(resp))

	if len(instance.Members) == 0 {
		return managed.ExternalObservation{
			ResourceExists: false,
		}, nil
	}

	currentSpec := cr.Spec.ForProvider.DeepCopy()
	if err = ibmcgm.LateInitializeSpec(&cr.Spec.ForProvider, instance); err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errManagedUpdateFailed)
	}
	if !cmp.Equal(currentSpec, &cr.Spec.ForProvider) {
		if err := c.kube.Update(ctx, cr); err != nil {
			return managed.ExternalObservation{}, errors.Wrap(err, errManagedUpdateFailed)
		}
	}

	cr.Status.AtProvider, err = ibmcgm.GenerateObservation(instance)
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errGenObservation)
	}

	cr.Status.SetConditions(cpv1alpha1.Available())
	cr.Status.AtProvider.State = ibmcgm.StateActive

	upToDate, err := ibmcgm.IsUpToDate(&cr.Spec.ForProvider, instance, c.logger)
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errCheckUpToDate)
	}

	return managed.ExternalObservation{
		ResourceExists:    true,
		ResourceUpToDate:  upToDate,
		ConnectionDetails: nil,
	}, nil
}

func (c *gmExternal) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1alpha1.GroupMembership)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotGroupMembership)
	}

	cr.SetConditions(cpv1alpha1.Creating())
	createOptions := &iamagv2.AddMembersToAccessGroupOptions{}
	if err := ibmcgm.GenerateCreateGroupMembershipOptions(cr.Spec.ForProvider, createOptions); err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreateGroupMembershipOpts)
	}

	_, resp, err := c.client.IamAccessGroupsV2().AddMembersToAccessGroup(createOptions)
	err = ibmcgm.ExtractErrorMessage(resp, err)
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(resource.Ignore(ibmc.IsResourceInactive, err), errCreateGroupMembership)
	}
	meta.SetExternalName(cr, reference.FromPtrValue(createOptions.AccessGroupID))
	return managed.ExternalCreation{ExternalNameAssigned: true}, nil
}

func (c *gmExternal) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1alpha1.GroupMembership)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotGroupMembership)
	}

	err := ibmcgm.UpdateAccessGroupMembers(c.client, *cr)
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errUpdGroupMembership)
	}

	return managed.ExternalUpdate{}, nil
}

func (c *gmExternal) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*v1alpha1.GroupMembership)
	if !ok {
		return errors.New(errNotGroupMembership)
	}

	cr.SetConditions(cpv1alpha1.Deleting())

	_, _, err := c.client.IamAccessGroupsV2().RemoveMembersFromAccessGroup(&iamagv2.RemoveMembersFromAccessGroupOptions{
		AccessGroupID: reference.ToPtrValue(meta.GetExternalName(cr)),
		Members:       ibmcgm.GenerateSDKRemoveroupMembersRequestMembersItems(cr.Spec.ForProvider.Members),
	})
	if err != nil {
		return errors.Wrap(resource.Ignore(ibmc.IsResourceGone, err), errDeleteGroupMembership)
	}
	return nil
}
