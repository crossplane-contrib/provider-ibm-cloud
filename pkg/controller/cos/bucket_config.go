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

	"github.com/google/go-cmp/cmp"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/pkg/errors"

	"github.com/crossplane-contrib/provider-ibm-cloud/apis/cos/v1alpha1"
	"github.com/crossplane-contrib/provider-ibm-cloud/apis/v1beta1"
	ibmc "github.com/crossplane-contrib/provider-ibm-cloud/pkg/clients"
	crossplaneClient "github.com/crossplane-contrib/provider-ibm-cloud/pkg/clients/cos"
)

// Various errors...
const (
	errThisIsNotABucketConfig = "managed resource is not a bucket configuration resource"
	errCreateBucketConfig     = "could not create a bucket configuration. This should never happen"
	errGetBucketConfigFailed  = "error getting the bucket configuration"
	errCreatePatchForUpdate   = "error creating the update structure to send to the server"
	errUpdBucketConfig        = "error updating the bucket configuration"
)

// SetupBucketConfig adds a controller that reconciles Bucket objects
func SetupBucketConfig(mgr ctrl.Manager, l logging.Logger) error {
	name := managed.ControllerName(v1alpha1.BucketConfigGroupKind)
	log := l.WithValues("bucket-config-controller", name)

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(v1alpha1.BucketConfigGroupVersionKind),
		managed.WithExternalConnecter(&bucketConfigConnector{
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
		For(&v1alpha1.BucketConfig{}).
		Complete(r)
}

// Expected to produce an object of type managed.ExternalClient when its Connect method
// is called.
type bucketConfigConnector struct {
	kube     client.Client
	usage    resource.Tracker
	clientFn func(optd ibmc.ClientOptions) (ibmc.ClientSession, error)
	logger   logging.Logger
}

// Produces an ExternalClient for the IBM Cloud API
func (c *bucketConfigConnector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	opts, err := ibmc.GetAuthInfo(ctx, c.kube, mg)
	if err != nil {
		return nil, errors.Wrap(err, ibmc.ErrGetAuth)
	}

	service, err := c.clientFn(opts)
	if err != nil {
		return nil, errors.Wrap(err, ibmc.ErrNewClient)
	}

	return &bucketConfigExternal{client: service, kube: c.kube, logger: c.logger}, nil
}

// Observes, then either creates, updates, or deletes an
// external bucket to ensure it reflects the managed resource's desired state.
type bucketConfigExternal struct {
	client ibmc.ClientSession
	kube   client.Client
	logger logging.Logger
}

// Called by crossplane. Does not create anything (a configuration always exists when a bucket exists) - justs sets the
// external name
func (c *bucketConfigExternal) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	crossplaneBucketConfig, ok := mg.(*v1alpha1.BucketConfig)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errCreateBucketConfig)
	}

	meta.SetExternalName(crossplaneBucketConfig, *crossplaneBucketConfig.Spec.ForProvider.Name)

	return managed.ExternalCreation{ExternalNameAssigned: true}, nil
}

// Called by crossplane
func (c *bucketConfigExternal) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) { // nolint:gocyclo
	crossplaneBucketConfig, ok := mg.(*v1alpha1.BucketConfig)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errThisIsNotABucketConfig)
	}

	externalBucketConfigName := meta.GetExternalName(crossplaneBucketConfig)
	if externalBucketConfigName == "" {
		return managed.ExternalObservation{
			ResourceExists: false,
		}, nil
	}

	configClient := c.client.BucketConfigClient()
	bucketConfigOptions := configClient.NewGetBucketConfigOptions(crossplaneBucketConfig.Name)
	ibmBucketConfig, _, err := configClient.GetBucketConfig(bucketConfigOptions)
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errGetBucketConfigFailed)
	}

	wasLateInitialized := false
	currentSpecCopy := crossplaneBucketConfig.Spec.ForProvider.DeepCopy()
	if wasLateInitialized, err = crossplaneClient.LateInitializeSpec(&crossplaneBucketConfig.Spec.ForProvider, ibmBucketConfig); err != nil {
		return managed.ExternalObservation{
			ResourceExists:          true,
			ResourceLateInitialized: wasLateInitialized,
		}, errors.Wrap(err, ibmc.ErrManagedUpdateFailed)
	}

	if !cmp.Equal(currentSpecCopy, &crossplaneBucketConfig.Spec.ForProvider) {
		if err := c.kube.Update(ctx, crossplaneBucketConfig); err != nil {
			return managed.ExternalObservation{
				ResourceExists:          true,
				ResourceLateInitialized: wasLateInitialized,
			}, errors.Wrap(err, ibmc.ErrManagedUpdateFailed)
		}
	}

	crossplaneBucketConfig.Status.AtProvider, err = crossplaneClient.GenerateBucketConfigObservation(ibmBucketConfig)
	if err != nil {
		return managed.ExternalObservation{
			ResourceExists:          true,
			ResourceLateInitialized: wasLateInitialized,
		}, errors.Wrap(err, ibmc.ErrGenObservation)
	}

	upToDate, err := crossplaneClient.IsUpToDate(&crossplaneBucketConfig.Spec.ForProvider, ibmBucketConfig, c.logger)
	if err != nil {
		return managed.ExternalObservation{
			ResourceExists:          true,
			ResourceLateInitialized: wasLateInitialized,
		}, errors.Wrap(err, ibmc.ErrCheckUpToDate)
	}

	return managed.ExternalObservation{
		ResourceExists:    true,
		ResourceUpToDate:  upToDate,
		ConnectionDetails: nil,
	}, nil
}

// Called by crossplane. Does not do anything, really, as the configuration always exists when a bucket exists
func (c *bucketConfigExternal) Delete(ctx context.Context, mg resource.Managed) error {
	_, ok := mg.(*v1alpha1.BucketConfig)
	if !ok {
		return errors.New(errThisIsNotABucketConfig)
	}

	return nil
}

// Called by crossplane
func (c *bucketConfigExternal) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	crossplaneBucketConfig, ok := mg.(*v1alpha1.BucketConfig)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errThisIsNotABucketConfig)
	}

	eTag := ibmc.GetEtagAnnotation(crossplaneBucketConfig)
	updateBucketInServerOptions, err := crossplaneClient.GenerateCloudBucketConfig(&crossplaneBucketConfig.Spec.ForProvider, &eTag)
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errCreatePatchForUpdate)
	}

	configClient := c.client.BucketConfigClient()
	_, err = configClient.UpdateBucketConfig(updateBucketInServerOptions)
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errUpdBucketConfig)
	}

	return managed.ExternalUpdate{}, nil
}
