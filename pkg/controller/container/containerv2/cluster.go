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

package cos

import (
	"context"

	"github.com/aws/aws-sdk-go/service/s3"
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
	errCreateClusterInp  = "could not generate the input params for a cluster"
	errDeleteCluster     = "could not delete the cluster"
	errGetClusterFailed  = "error getting the cluster"
	errUpdCluster        = "error updating the cluster"
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

	ibmCluster, err := c.client.ClustersClientV2().GetCluster(crossplaneCluster.Name, ibmContainerV2.ClusterTargetHeader{})
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(resource.Ignore(ibmc.IsResourceNotFound, err), errGetClusterFailed)
	} else if ibmCluster != nil {
		crossplaneCluster.Status.AtProvider, err = crossplaneClient.GenerateClusterInfo(ibmCluster)
		if err != nil {
			return managed.ExternalObservation{}, errors.Wrap(err, errUpdBucket)
		}
	}

	return managed.ExternalObservation{
		ResourceExists:    ibmCluster != nil,
		ResourceUpToDate:  true,
		ConnectionDetails: nil,
	}, nil
}

// Called by crossplane
func (c *bucketExternal) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	crossplaneBucket, ok := mg.(*v1alpha1.Bucket)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errCreateBucket)
	}

	crossplaneBucket.SetConditions(runtimev1alpha1.Creating())

	s3BucketInp := s3.CreateBucketInput{}
	if err := crossplaneClient.GenerateS3BucketInput(crossplaneBucket.Spec.ForProvider.DeepCopy(), &s3BucketInp); err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreateBucketInp)
	}

	s3Client := c.generateClient()
	_, err := s3Client.CreateBucket(&s3BucketInp)
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreateBucket)
	}

	meta.SetExternalName(crossplaneBucket, crossplaneBucket.Spec.ForProvider.Name)

	return managed.ExternalCreation{ExternalNameAssigned: true}, nil
}

// (Not) called by crossplane - as the bucket cannot be changed once created... Here only to satisfy the compiler
func (c *bucketExternal) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	return managed.ExternalUpdate{}, nil
}

// Called by crossplane
func (c *bucketExternal) Delete(ctx context.Context, mg resource.Managed) error {
	crossplaneBucket, ok := mg.(*v1alpha1.Bucket)
	if !ok {
		return errors.New(errThisIsNotABucket)
	}

	crossplaneBucket.SetConditions(runtimev1alpha1.Deleting())

	s3Client := c.generateClient()

	_, err := s3Client.DeleteBucket(&s3.DeleteBucketInput{Bucket: &crossplaneBucket.Spec.ForProvider.Name})
	if err != nil {
		return errors.Wrap(resource.Ignore(ibmc.IsResourceNotFound, err), errDeleteBucket)
	}

	return nil
}

// Returns a client, potentially adapted for unit test running (unit test requires extra params)
func (c *bucketExternal) generateClient() *s3.S3 {
	result := c.client.S3Client()

	if c.unitTestRegionAndCredentials != nil {
		result.Config.Credentials = c.unitTestRegionAndCredentials.credentials
		result.Config.Region = &c.unitTestRegionAndCredentials.region
		result.ClientInfo.SigningRegion = c.unitTestRegionAndCredentials.region
	}

	return result
}
