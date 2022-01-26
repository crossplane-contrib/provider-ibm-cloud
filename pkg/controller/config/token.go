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

package config

import (
	"context"
	"net/http"
	"time"

	"github.com/pkg/errors"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	"github.com/IBM-Cloud/bluemix-go"
	"github.com/IBM-Cloud/bluemix-go/authentication"
	"github.com/IBM-Cloud/bluemix-go/endpoints"
	"github.com/IBM-Cloud/bluemix-go/rest"

	"github.com/crossplane-contrib/provider-ibm-cloud/apis/v1beta1"
	ibmc "github.com/crossplane-contrib/provider-ibm-cloud/pkg/clients"
)

const (
	timeout                  = 2 * time.Minute
	requeueTime              = 30 * time.Minute
	errGetPC                 = "cannot get ProviderConfig"
	errNoSecret              = "no credentials/secret reference was provided"
	errGetSecret             = "cannot get credentials/secret"
	errGetRefreshTokenSecret = "cannot get a refresh token from the server"
	errSaveSecret            = "cannot save new credentials from the cloud"
)

// SetupToken adds a controller that reconciles ProviderConfigs by accounting for
// their current usage.
func SetupToken(mgr ctrl.Manager, l logging.Logger) error {
	name := "TokenController"

	of := resource.ProviderConfigKinds{
		Config: v1beta1.ProviderConfigGroupVersionKind,
	}

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&v1beta1.ProviderConfig{}).
		Complete(NewTokenReconciler(mgr, of,
			WithLogger(l.WithValues("token-controller", name)),
			WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

// A TokenReconciler reconciles managed resources by creating and managing the
// lifecycle of an external resource, i.e. a resource in an external system such
// as a cloud provider API. Each controller must watch the managed resource kind
// for which it is responsible.
type TokenReconciler struct {
	client client.Client

	newConfig func() resource.ProviderConfig

	log    logging.Logger
	record event.Recorder
}

// A TokenReconcilerOption configures a Reconciler.
type TokenReconcilerOption func(*TokenReconciler)

// WithLogger specifies how the Reconciler should log messages.
func WithLogger(l logging.Logger) TokenReconcilerOption {
	return func(r *TokenReconciler) {
		r.log = l
	}
}

// WithRecorder specifies how the Reconciler should record events.
func WithRecorder(er event.Recorder) TokenReconcilerOption {
	return func(r *TokenReconciler) {
		r.record = er
	}
}

// NewTokenReconciler returns a Reconciler of ProviderConfigs.
func NewTokenReconciler(m manager.Manager, of resource.ProviderConfigKinds, o ...TokenReconcilerOption) *TokenReconciler {
	nc := func() resource.ProviderConfig {
		return resource.MustCreateObject(of.Config, m.GetScheme()).(resource.ProviderConfig)
	}

	// Panic early if we've been asked to reconcile a resource kind that has not
	// been registered with our controller manager's scheme.
	_ = nc()

	r := &TokenReconciler{
		client: m.GetClient(),

		newConfig: nc,

		log:    logging.NewNopLogger(),
		record: event.NewNopRecorder(),
	}

	for _, ro := range o {
		ro(r)
	}

	return r
}

// Reconcile a ProviderConfig by accounting for the managed resources that are
// using it, and ensuring it cannot be deleted until it is no longer in use.
func (r *TokenReconciler) Reconcile(req reconcile.Request) (reconcile.Result, error) {
	log := r.log.WithValues("request", req)
	log.Debug("Reconciling")

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	pc := &v1beta1.ProviderConfig{}
	if err := r.client.Get(ctx, req.NamespacedName, pc); err != nil {
		// In case object is not found, most likely the object was deleted and
		// then disappeared while the event was in the processing queue. We
		// don't need to take any action in that case.
		log.Debug(errGetPC, "error", err)
		return reconcile.Result{}, errors.Wrap(resource.IgnoreNotFound(err), errGetPC)
	}

	// Get the access token
	secretRef := pc.Spec.Credentials.SecretRef
	if secretRef == nil {
		return reconcile.Result{}, errors.New(errGetSecret)
	}

	secret := &v1.Secret{}
	if err := r.client.Get(ctx, types.NamespacedName{Name: secretRef.Name, Namespace: secretRef.Namespace}, secret); err != nil {
		return reconcile.Result{}, errors.Wrap(err, errGetSecret)
	}

	blueMixConfig := bluemix.Config{
		EndpointLocator: endpoints.NewEndpointLocator(getRegion(*pc)),
	}

	auth, err := authentication.NewIAMAuthRepository(&blueMixConfig, &rest.Client{HTTPClient: http.DefaultClient})
	if err != nil {
		return reconcile.Result{}, err
	}

	if err := auth.AuthenticateAPIKey(string(secret.Data[secretRef.Key])); err != nil {
		return reconcile.Result{}, err
	}

	secret.Data[ibmc.AccessTokenKey] = []byte(blueMixConfig.IAMAccessToken)
	secret.Data[ibmc.RefreshTokenKey] = []byte(blueMixConfig.IAMRefreshToken)
	if err := r.client.Update(ctx, secret); err != nil {
		return reconcile.Result{}, errors.Wrap(err, errSaveSecret)
	}

	return reconcile.Result{RequeueAfter: requeueTime}, nil
}

func getRegion(pc v1beta1.ProviderConfig) string {
	if pc.Spec.Region != "" {
		return pc.Spec.Region
	}

	return ibmc.DefaultRegion
}
