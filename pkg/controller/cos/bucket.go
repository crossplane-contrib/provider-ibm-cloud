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

	"github.com/IBM/ibm-cos-sdk-go/aws/credentials"
	"github.com/IBM/ibm-cos-sdk-go/service/s3"
	"github.com/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	"github.com/crossplane-contrib/provider-ibm-cloud/apis/cos/v1alpha1"
	"github.com/crossplane-contrib/provider-ibm-cloud/apis/v1beta1"
	ibmc "github.com/crossplane-contrib/provider-ibm-cloud/pkg/clients"
	crossplane_client "github.com/crossplane-contrib/provider-ibm-cloud/pkg/clients/cos"
)

// SetupBucket adds a controller that reconciles Bucket objects
func SetupBucket(mgr ctrl.Manager, l logging.Logger) error {
	name := managed.ControllerName(v1alpha1.BucketGroupKind)
	log := l.WithValues("bucket-controller", name)

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(v1alpha1.BucketGroupVersionKind),
		managed.WithExternalConnecter(&bucketConnector{
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
		For(&v1alpha1.Bucket{}).
		Complete(r)
}

// Expected to produce an object of type managed.ExternalClient when its Connect method
// is called.
type bucketConnector struct {
	kube     client.Client
	usage    resource.Tracker
	clientFn func(optd ibmc.ClientOptions) (ibmc.ClientSession, error)
	logger   logging.Logger
}

// Produces an ExternalClient for the IBM Cloud API
func (c *bucketConnector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	opts, err := ibmc.GetAuthInfo(ctx, c.kube, mg)
	if err != nil {
		return nil, errors.Wrap(err, ibmc.ErrGetAuth)
	}

	service, err := c.clientFn(opts)
	if err != nil {
		return nil, errors.Wrap(err, ibmc.ErrNewClient)
	}

	return &bucketExternal{client: service, kube: c.kube, logger: c.logger}, nil
}

// Various errors...
const (
	errThisIsNotABucket = "managed resource is not a bucket resource"
	errCreateBucket     = "could not create a bucket"
	errCreateBucketInp  = "could not generate the input paraks for a bucket"
	errDeleteBucket     = "could not delete the bucket"
	errGetBucketFailed  = "error getting the bucket"
	errUpdBucket        = "error updating the bucket"
)

// Because we use the Amazon S3 API, it requires more stuff when in unit-test mode than when actually
// interacting with the IBM cloud. This type helps keep truck of what modus operandi we are in
type unitTestRegionAndCredentials struct {
	credentials *credentials.Credentials
	region      string
}

// Observes, then either creates, updates, or deletes an
// external bucket to ensure it reflects the managed resource's desired state.
type bucketExternal struct {
	unitTestRegionAndCredentials *unitTestRegionAndCredentials
	client                       ibmc.ClientSession
	kube                         client.Client
	logger                       logging.Logger
}

// Retrieves the bucket with the given name, and for the same instance id, as the parameters if one exists.
//
// Params
//	   crossplaneBucket - a "crossplane" bucket
//
// Returns
//	   the bucket retrieved (if error != nil), or nil - if no bucket was found
//	   the error that happened when trying to retrieve all the buckets; nil if none exists
func (c *bucketExternal) retrieveBucket(crossplaneBucket *v1alpha1.Bucket) (*s3.Bucket, error) {
	var result *s3.Bucket

	s3Client := c.generateClient()
	bo, err := s3Client.ListBuckets(&s3.ListBucketsInput{IBMServiceInstanceId: crossplaneBucket.Spec.ForProvider.IbmServiceInstanceID})
	if err == nil {
		for _, bi := range bo.Buckets {
			if *bi.Name == crossplaneBucket.Spec.ForProvider.Name {
				result = bi

				break
			}
		}
	}

	return result, err
}

// Called by crossplane
func (c *bucketExternal) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) { // nolint:gocyclo
	crossplaneBucket, ok := mg.(*v1alpha1.Bucket)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errThisIsNotABucket)
	}

	externalBucketName := meta.GetExternalName(crossplaneBucket)
	if externalBucketName == "" {
		return managed.ExternalObservation{
			ResourceExists: false,
		}, nil
	}

	s3Bucket, err := c.retrieveBucket(crossplaneBucket)
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(resource.Ignore(ibmc.IsResourceNotFound, err), errGetBucketFailed)
	} else if s3Bucket != nil {
		crossplaneBucket.Status.AtProvider, err = crossplane_client.GenerateObservation(s3Bucket)
		if err != nil {
			return managed.ExternalObservation{}, errors.Wrap(err, errUpdBucket)
		}
	}

	return managed.ExternalObservation{
		ResourceExists:    s3Bucket != nil,
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
	if err := crossplane_client.GenerateS3BucketInput(crossplaneBucket.Spec.ForProvider.DeepCopy(), &s3BucketInp); err != nil {
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
