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

package vpcv1

import (
	"context"
	"fmt"
	"net/http"

	ibmVPC "github.com/IBM/vpc-go-sdk/vpcv1"

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

	"github.com/crossplane-contrib/provider-ibm-cloud/apis/v1beta1"
	"github.com/crossplane-contrib/provider-ibm-cloud/apis/vpcv1/v1alpha1"
	crossplaneClient "github.com/crossplane-contrib/provider-ibm-cloud/pkg/clients/vpcv1/subnet"

	ibmc "github.com/crossplane-contrib/provider-ibm-cloud/pkg/clients"
)

// Various errors...
const (
	errThisIsNotSubnet       = "managed resource is not a Subnet resource"
	errCreateSubnet          = "could not create a Subnet"
	errCreateReqSubnet       = "could not generate the input params for a Subnet"
	errDeleteSubnet          = "could not delete the Subnet"
	errGetFailedSubnet       = "error getting the Subnet"
	errUpdateSubnet          = "error updating the Subnet"
	errCouldNotGeneratePatch = "could not generate a diff patch"
)

// SetupSubnet adds a controller that reconciles Subnet objects
func SetupSubnet(mgr ctrl.Manager, l logging.Logger) error {
	name := managed.ControllerName(v1alpha1.SubnetGroupKind)
	log := l.WithValues("subnet-controller", name)

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(v1alpha1.SubnetGroupVersionKind),
		managed.WithExternalConnecter(&subnetConnector{
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
		For(&v1alpha1.Subnet{}).
		Complete(r)
}

// Expected to produce an object of type managed.ExternalClient when its Connect method
// is called.
type subnetConnector struct {
	kube     client.Client
	usage    resource.Tracker
	clientFn func(optd ibmc.ClientOptions) (ibmc.ClientSession, error)
	logger   logging.Logger
}

// Produces an ExternalClient for the IBM Cloud API
func (c *subnetConnector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	opts, err := ibmc.GetAuthInfo(ctx, c.kube, mg)
	if err != nil {
		return nil, errors.Wrap(err, ibmc.ErrGetAuth)
	}

	service, err := c.clientFn(opts)
	if err != nil {
		return nil, errors.Wrap(err, ibmc.ErrNewClient)
	}

	return &subnetExternal{client: service, kube: c.kube, logger: c.logger}, nil
}

// Observes, then either creates, updates, or deletes an
// external Subnet to ensure it reflects the managed resource's desired state.
type subnetExternal struct {
	client ibmc.ClientSession
	kube   client.Client
	logger logging.Logger
}

// Params
//
//	c - ...
//	crn - a subnet's crn
//
// Returns
//
//	the subnet with the given CRN, nil o/w
func getSubnet(c *subnetExternal, crn string) (*ibmVPC.Subnet, error) {
	subnetCollection, response, err := c.client.VPCClient().ListSubnets(&ibmVPC.ListSubnetsOptions{})
	if err != nil {
		return nil, err
	}

	if response.StatusCode != http.StatusOK {
		return nil, errors.New("ListSubnets returned status code: " + fmt.Sprint(response.StatusCode) + ", and response: " + response.String())
	}

	if subnetCollection != nil {
		for i := range subnetCollection.Subnets {
			cloudSubnet := subnetCollection.Subnets[i]

			if crn == *cloudSubnet.CRN {
				return &cloudSubnet, nil
			}
		}
	}

	return nil, nil
}

// Called by crossplane
func (c *subnetExternal) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) { // nolint:gocyclo
	crossplaneSubnet, ok := mg.(*v1alpha1.Subnet)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errThisIsNotSubnet)
	}

	externalName := meta.GetExternalName(crossplaneSubnet)
	if externalName == "" {
		return managed.ExternalObservation{
			ResourceExists: false,
		}, nil
	}

	found := false
	wasLateInitialized := false
	isUpToDate := false

	cloudSubnet, err := getSubnet(c, externalName)
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errGetFailedSubnet)
	}

	if cloudSubnet != nil {
		found = true

		currentSpec := crossplaneSubnet.Spec.ForProvider.DeepCopy()
		if wasLateInitialized, err = crossplaneClient.LateInitializeSpec(&crossplaneSubnet.Spec.ForProvider, cloudSubnet); err != nil {
			return managed.ExternalObservation{}, errors.Wrap(err, ibmc.ErrGenObservation)
		}

		if !cmp.Equal(currentSpec, &crossplaneSubnet.Spec.ForProvider) {
			if err := c.kube.Update(ctx, crossplaneSubnet); err != nil {
				return managed.ExternalObservation{}, errors.Wrap(err, ibmc.ErrGenObservation)
			}
		}

		if crossplaneSubnet.Status.AtProvider, err = crossplaneClient.GenerateObservation(cloudSubnet); err != nil {
			return managed.ExternalObservation{}, errors.Wrap(err, ibmc.ErrGenObservation)
		}

		if isUpToDate, err = crossplaneClient.IsUpToDate(&crossplaneSubnet.Spec.ForProvider, cloudSubnet, c.logger); err != nil {
			return managed.ExternalObservation{
				ResourceExists:          true,
				ResourceLateInitialized: wasLateInitialized,
			}, errors.Wrap(err, ibmc.ErrCheckUpToDate)
		}
	}

	return managed.ExternalObservation{
		ResourceExists:          found,
		ResourceUpToDate:        isUpToDate,
		ResourceLateInitialized: wasLateInitialized,
		ConnectionDetails:       nil,
	}, nil
}

// Called by crossplane
func (c *subnetExternal) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	crossplaneSubnet, ok := mg.(*v1alpha1.Subnet)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errThisIsNotSubnet)
	}

	crossplaneSubnet.SetConditions(runtimev1alpha1.Creating())

	createOptions, err := crossplaneClient.GenerateCreateOptions(&crossplaneSubnet.Spec.DeepCopy().ForProvider)
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreateReqSubnet)
	}

	subnet, _, err := c.client.VPCClient().CreateSubnet(&createOptions)
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreateSubnet)
	}

	meta.SetExternalName(crossplaneSubnet, *subnet.CRN)

	return managed.ExternalCreation{ExternalNameAssigned: true}, nil
}

// Called by crossplane
func (c *subnetExternal) Delete(ctx context.Context, mg resource.Managed) error {
	crossplaneSubnet, ok := mg.(*v1alpha1.Subnet)
	if !ok {
		return errors.New(errThisIsNotSubnet)
	}

	crossplaneSubnet.SetConditions(runtimev1alpha1.Deleting())

	if _, err := c.client.VPCClient().DeleteSubnet(&ibmVPC.DeleteSubnetOptions{
		ID: reference.ToPtrValue(crossplaneSubnet.Status.AtProvider.ID),
	}); err != nil {
		return errors.Wrap(resource.Ignore(ibmc.IsResourceNotFound, err), errDeleteSubnet)
	}

	return nil
}

// Called by crossplane
func (c *subnetExternal) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	crossplaneSubnet, ok := mg.(*v1alpha1.Subnet)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errThisIsNotSubnet)
	}

	externalName := meta.GetExternalName(crossplaneSubnet)
	cloudSubnet, err := getSubnet(c, externalName)
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errGetFailedSubnet)
	}

	if cloudSubnet == nil {
		return managed.ExternalUpdate{}, errors.Wrap(errors.New(errGetFailedSubnet), errUpdateSubnet)
	}

	subnetPatch, err := crossplaneClient.DiffPatch(&crossplaneSubnet.Spec.ForProvider, cloudSubnet)
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errCouldNotGeneratePatch)
	}

	updateOptions := ibmVPC.UpdateSubnetOptions{}
	updateOptions.SetID(crossplaneSubnet.Status.AtProvider.ID)
	updateOptions.SetSubnetPatch(subnetPatch)

	if _, _, err := c.client.VPCClient().UpdateSubnet(&updateOptions); err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errUpdateSubnet)
	}

	return managed.ExternalUpdate{}, nil
}
