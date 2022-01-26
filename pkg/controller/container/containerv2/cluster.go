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

package containerv2

import (
	"context"

	"github.com/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	ibmContainerV2 "github.com/IBM-Cloud/bluemix-go/api/container/containerv2"
	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	"github.com/crossplane-contrib/provider-ibm-cloud/apis/container/containerv2/v1alpha1"
	"github.com/crossplane-contrib/provider-ibm-cloud/apis/v1beta1"
	ibmc "github.com/crossplane-contrib/provider-ibm-cloud/pkg/clients"
	crossplaneClient "github.com/crossplane-contrib/provider-ibm-cloud/pkg/clients/container/containerv2"
)

// Various errors...
const (
	errThisIsNotACluster = "managed resource is not a cluster resource"
	errCreateCluster     = "could not create a cluster"
	errCreateClusterReq  = "could not generate the input params for a cluster"
	errDeleteCluster     = "could not delete the cluster"
	errGetClusterFailed  = "error getting the cluster"
)

// SetupCluster adds a controller that reconciles Cluster objects
func SetupCluster(mgr ctrl.Manager, l logging.Logger) error {
	name := managed.ControllerName(v1alpha1.ClusterGroupKind)
	log := l.WithValues("cluster-controller", name)

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(v1alpha1.ClusterGroupVersionKind),
		managed.WithExternalConnecter(&clusterConnector{
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
		For(&v1alpha1.Cluster{}).
		Complete(r)
}

// Expected to produce an object of type managed.ExternalClient when its Connect method
// is called.
type clusterConnector struct {
	kube     client.Client
	usage    resource.Tracker
	clientFn func(optd ibmc.ClientOptions) (ibmc.ClientSession, error)
	logger   logging.Logger
}

// Produces an ExternalClient for the IBM Cloud API
func (c *clusterConnector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	opts, err := ibmc.GetAuthInfo(ctx, c.kube, mg)
	if err != nil {
		return nil, errors.Wrap(err, ibmc.ErrGetAuth)
	}

	service, err := c.clientFn(opts)
	if err != nil {
		return nil, errors.Wrap(err, ibmc.ErrNewClient)
	}

	return &clusterExternal{client: service, kube: c.kube, logger: c.logger}, nil
}

// Observes, then either creates, updates, or deletes an
// external cluster to ensure it reflects the managed resource's desired state.
type clusterExternal struct {
	client ibmc.ClientSession
	kube   client.Client
	logger logging.Logger
}

// Called by crossplane
func (c *clusterExternal) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) { // nolint:gocyclo
	crossplaneCluster, ok := mg.(*v1alpha1.Cluster)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errThisIsNotACluster)
	}

	externalClusterName := meta.GetExternalName(crossplaneCluster)
	if externalClusterName == "" {
		return managed.ExternalObservation{
			ResourceExists: false,
		}, nil
	}

	ibmClusterInfo, err := c.client.ClusterClientV2().GetCluster(crossplaneCluster.Spec.ForProvider.Name, ibmContainerV2.ClusterTargetHeader{})
	if err != nil {
		if ibmc.IsResourceNotFound(err) {
			return managed.ExternalObservation{ResourceExists: false}, errors.Wrap(resource.Ignore(ibmc.IsResourceNotFound, err), errGetClusterFailed)
		} else {
			return managed.ExternalObservation{}, errors.Wrap(resource.Ignore(ibmc.IsResourceNotFound, err), errGetClusterFailed)
		}
	} else if ibmClusterInfo != nil {
		crossplaneCluster.Status.AtProvider, err = crossplaneClient.GenerateCrossplaneClusterInfo(ibmClusterInfo)
		if err != nil {
			return managed.ExternalObservation{}, errors.Wrap(err, ibmc.ErrGenObservation)
		}
	}

	return managed.ExternalObservation{
		ResourceExists:    ibmClusterInfo != nil,
		ResourceUpToDate:  true,
		ConnectionDetails: nil,
	}, nil
}

// Called by crossplane
func (c *clusterExternal) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	crossplaneCluster, ok := mg.(*v1alpha1.Cluster)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errThisIsNotACluster)
	}

	crossplaneCluster.SetConditions(runtimev1alpha1.Creating())

	createRequest := ibmContainerV2.ClusterCreateRequest{}
	if err := crossplaneClient.GenerateClusterCreateRequest(crossplaneCluster.Spec.ForProvider.DeepCopy(), &createRequest); err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreateClusterReq)
	}

	_, err := c.client.ClusterClientV2().Create(createRequest, ibmContainerV2.ClusterTargetHeader{})
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreateCluster)
	}

	meta.SetExternalName(crossplaneCluster, crossplaneCluster.Spec.ForProvider.Name)

	return managed.ExternalCreation{ExternalNameAssigned: true}, nil
}

// (Not) called by crossplane - as the bucket cannot be changed once created... Here only to satisfy the compiler
func (c *clusterExternal) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	return managed.ExternalUpdate{}, nil
}

// Called by crossplane
func (c *clusterExternal) Delete(ctx context.Context, mg resource.Managed) error {
	crossplaneCluster, ok := mg.(*v1alpha1.Cluster)
	if !ok {
		return errors.New(errThisIsNotACluster)
	}

	crossplaneCluster.SetConditions(runtimev1alpha1.Deleting())

	err := c.client.ClusterClientV2().Delete(crossplaneCluster.Spec.ForProvider.Name, ibmContainerV2.ClusterTargetHeader{})
	if err != nil {
		return errors.Wrap(resource.Ignore(ibmc.IsResourceNotFound, err), errDeleteCluster)
	}

	return nil
}
