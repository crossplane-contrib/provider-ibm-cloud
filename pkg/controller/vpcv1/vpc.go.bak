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
	"net/http"

	"github.com/IBM/vpc-go-sdk/vpcv1"
	"github.com/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	"github.com/crossplane-contrib/provider-ibm-cloud/apis/v1beta1"
	"github.com/crossplane-contrib/provider-ibm-cloud/apis/vpcv1/v1alpha1"
	crossplaneClient "github.com/crossplane-contrib/provider-ibm-cloud/pkg/clients/vpcv1"

	ibmc "github.com/crossplane-contrib/provider-ibm-cloud/pkg/clients"
)

// Various errors...
const (
	errThisIsNotAVPC = "managed resource is not a VPC resource"
	errCreateVPC     = "could not create a VPC"
	errCreateVPCReq  = "could not generate the input params for a VPC"
	errDeleteVPC     = "could not delete the VPC"
	errGetVPCFailed  = "error getting the VOC"
)

// SetupVPC adds a controller that reconciles VPC objects
func SetupVPC(mgr ctrl.Manager, l logging.Logger) error {
	name := managed.ControllerName(v1alpha1.VPCGroupKind)
	log := l.WithValues("vpc-controller", name)

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(v1alpha1.VPCGroupVersionKind),
		managed.WithExternalConnecter(&vpcConnector{
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
		For(&v1alpha1.VPC{}).
		Complete(r)
}

// Expected to produce an object of type managed.ExternalClient when its Connect method
// is called.
type vpcConnector struct {
	kube     client.Client
	usage    resource.Tracker
	clientFn func(optd ibmc.ClientOptions) (ibmc.ClientSession, error)
	logger   logging.Logger
}

// Produces an ExternalClient for the IBM Cloud API
func (c *vpcConnector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	opts, err := ibmc.GetAuthInfo(ctx, c.kube, mg)
	if err != nil {
		return nil, errors.Wrap(err, ibmc.ErrGetAuth)
	}

	service, err := c.clientFn(opts)
	if err != nil {
		return nil, errors.Wrap(err, ibmc.ErrNewClient)
	}

	return &vpcExternal{client: service, kube: c.kube, logger: c.logger}, nil
}

// Observes, then either creates, updates, or deletes an
// external VPC to ensure it reflects the managed resource's desired state.
type vpcExternal struct {
	client ibmc.ClientSession
	kube   client.Client
	logger logging.Logger
}

// Called by crossplane
func (c *vpcExternal) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) { // nolint:gocyclo
	crossplaneVPC, ok := mg.(*v1alpha1.VPC)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errThisIsNotAVPC)
	}

	externalClusterName := meta.GetExternalName(crossplaneVPC)
	if externalClusterName == "" {
		return managed.ExternalObservation{
			ResourceExists: false,
		}, nil
	}

	vpcCollection, response, err := c.client.VPCClient().ListVpcs(&vpcv1.ListVpcsOptions{})
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errGetVPCFailed)
	} else if response.StatusCode != http.StatusOK {
		return managed.ExternalObservation{}, errors.New("ListVpcs returned status code: " + string(response.StatusCode) + ", and response: " + response.String())
	} else if vpcCollection != nil {
		for cloudVPC := range vpcCollection.Vpcs {
			if *crossplaneVPC.Spec.ForProvider.Name == *cloudVPC.Name {
			}
		if err != nil {
			return managed.ExternalObservation{}, errors.Wrap(err, ibmc.ErrGenObservation)
		}
	}

	return managed.ExternalObservation{
		ResourceExists:    true, // ibmClusterInfo != nil,
		ResourceUpToDate:  true,
		ConnectionDetails: nil,
	}, nil
}

// Called by crossplane
func (c *vpcExternal) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	crossplaneVPC, ok := mg.(*v1alpha1.VPC)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errThisIsNotAVPC)
	}

	crossplaneVPC.SetConditions(runtimev1alpha1.Creating())

	createOptions, err := crossplaneClient.GenerateCloudVPCParams(&crossplaneVPC.Spec.ForProvider)
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreateVPCReq)
	}

	vpc, response, err := c.client.VPCClient().CreateVPC(&createOptions)
	if err != nil {
		if response != nil {
			response = nil
		}
		return managed.ExternalCreation{}, errors.Wrap(err, errCreateVPC)
	}

	meta.SetExternalName(crossplaneVPC, *vpc.Name)

	return managed.ExternalCreation{ExternalNameAssigned: true}, nil
}

// (Not) called by crossplane - as the bucket cannot be changed once created... Here only to satisfy the compiler
func (c *vpcExternal) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	return managed.ExternalUpdate{}, nil
}

// Called by crossplane
func (c *vpcExternal) Delete(ctx context.Context, mg resource.Managed) error {
	crossplaneCluster, ok := mg.(*v1alpha1.VPC)
	if !ok {
		return errors.New(errThisIsNotAVPC)
	}

	crossplaneCluster.SetConditions(runtimev1alpha1.Deleting())

	err := c.client.ClusterClientV2().Delete(crossplaneCluster.Spec.ForProvider.Name, ibmContainerV2.ClusterTargetHeader{})
	if err != nil {
		return errors.Wrap(resource.Ignore(ibmc.IsResourceNotFound, err), errDeleteCluster)
	}

	return nil
}
