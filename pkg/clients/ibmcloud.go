/*
Copyright 2019 The Crossplane Authors.

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

package clients

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-openapi/strfmt"
	"github.com/pkg/errors"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane/crossplane-runtime/pkg/resource"

	"github.com/IBM/go-sdk-core/core"
	gcat "github.com/IBM/platform-services-go-sdk/globalcatalogv1"
	gtagv1 "github.com/IBM/platform-services-go-sdk/globaltaggingv1"
	rcv2 "github.com/IBM/platform-services-go-sdk/resourcecontrollerv2"
	rmgrv2 "github.com/IBM/platform-services-go-sdk/resourcemanagerv2"

	"github.com/crossplane-contrib/provider-ibm-cloud/apis/v1beta1"
)

const (
	// AccessTokenKey key for IBM Cloud API access token
	AccessTokenKey = "access_token"
	// DefaultRegion is the default region for the IBM Cloud API
	DefaultRegion = "us-south"

	errTokNotFound        = "IAM access token key not found in provider config secret"
	errGetSecret          = "cannot get credentials secret"
	errGetTracker         = "error setting up provider config usage tracker"
	errGetProviderCfg     = "error getting provider config"
	errNoSecret           = "no credentials secret reference was provided"
	errGetGcat            = "error initializing GlobalCatalogV1 client"
	errGetGtag            = "error initializing GlobalTaggingV1 client"
	errParseTok           = "error parsig IAM access token"
	errNotFound           = "Not Found"
	errPendingReclamation = "Instance is pending reclamation"
	errGone               = "Gone"
	errRemovedInvalid     = "The resource instance is removed/invalid"
)

// ClientOptions provides info to initialize a client for the IBM Cloud APIs
type ClientOptions struct {
	ServiceName   string
	URL           string
	Authenticator core.Authenticator
}

// GetAuthInfo returns the necessary authentication information that is necessary
// to use when the controller connects to GCP API in order to reconcile the managed
// resource.
func GetAuthInfo(ctx context.Context, c client.Client, mg resource.Managed) (opts ClientOptions, err error) {
	pc := &v1beta1.ProviderConfig{}
	t := resource.NewProviderConfigUsageTracker(c, &v1beta1.ProviderConfigUsage{})
	if err := t.Track(ctx, mg); err != nil {
		return ClientOptions{}, errors.Wrap(err, errGetTracker)
	}
	if err := c.Get(ctx, types.NamespacedName{Name: mg.GetProviderConfigReference().Name}, pc); err != nil {
		return ClientOptions{}, errors.Wrap(err, errGetProviderCfg)
	}

	ref := pc.Spec.Credentials.SecretRef
	if ref == nil {
		return ClientOptions{}, errors.New(errNoSecret)
	}

	s := &v1.Secret{}
	if err := c.Get(ctx, types.NamespacedName{Name: ref.Name, Namespace: ref.Namespace}, s); err != nil {
		return ClientOptions{}, errors.Wrap(err, errGetSecret)
	}
	authenticator, err := getAuthenticator(s)
	if err != nil {
		return ClientOptions{}, err
	}

	return ClientOptions{Authenticator: authenticator}, nil
}

func getAuthenticator(s *v1.Secret) (core.Authenticator, error) {
	aTok, ok := s.Data[AccessTokenKey]
	if !ok {
		return nil, errors.New(errTokNotFound)
	}

	bearerTok, err := getBearerFromAccessToken(string(aTok))
	if err != nil {
		return nil, err
	}

	authenticator := &core.BearerTokenAuthenticator{
		BearerToken: bearerTok,
	}
	return authenticator, nil
}

func getBearerFromAccessToken(aTok string) (string, error) {
	toks := strings.Split(aTok, " ")
	if len(toks) != 2 {
		return "", errors.New(errParseTok)
	}
	return toks[1], nil
}

// NewClient returns an IBM API client
func NewClient(opts ClientOptions) (ClientSession, error) {
	var err error
	cs := clientSessionImpl{}

	rcv2Opts := &rcv2.ResourceControllerV2Options{
		ServiceName:   opts.ServiceName,
		Authenticator: opts.Authenticator,
		URL:           opts.URL,
	}
	cs.resourceControllerV2, err = rcv2.NewResourceControllerV2(rcv2Opts)
	if err != nil {
		return nil, errors.Wrap(err, errGetGcat)
	}

	gcatOpts := &gcat.GlobalCatalogV1Options{
		ServiceName:   opts.ServiceName,
		Authenticator: opts.Authenticator,
		URL:           opts.URL,
	}
	cs.globalCatalogV1, err = gcat.NewGlobalCatalogV1(gcatOpts)
	if err != nil {
		return nil, errors.Wrap(err, errGetGcat)
	}

	rmgrOpts := &rmgrv2.ResourceManagerV2Options{
		ServiceName:   opts.ServiceName,
		Authenticator: opts.Authenticator,
		URL:           opts.URL,
	}
	cs.resourceManagerV2, err = rmgrv2.NewResourceManagerV2(rmgrOpts)
	if err != nil {
		return nil, errors.Wrap(err, errGetGcat)
	}

	gtagsOpts := &gtagv1.GlobalTaggingV1Options{
		ServiceName:   opts.ServiceName,
		Authenticator: opts.Authenticator,
		URL:           opts.URL,
	}
	cs.globalTaggingV1, err = gtagv1.NewGlobalTaggingV1(gtagsOpts)
	if err != nil {
		return nil, errors.Wrap(err, errGetGtag)
	}

	return &cs, err
}

// ClientSession provides an interface for IBM Cloud APIs
type ClientSession interface {
	ResourceControllerV2() *rcv2.ResourceControllerV2
	GlobalCatalogV1() *gcat.GlobalCatalogV1
	ResourceManagerV2() *rmgrv2.ResourceManagerV2
	GlobalTaggingV1() *gtagv1.GlobalTaggingV1
}

type clientSessionImpl struct {
	resourceControllerV2 *rcv2.ResourceControllerV2
	globalCatalogV1      *gcat.GlobalCatalogV1
	resourceManagerV2    *rmgrv2.ResourceManagerV2
	globalTaggingV1      *gtagv1.GlobalTaggingV1
}

func (c *clientSessionImpl) ResourceControllerV2() *rcv2.ResourceControllerV2 {
	return c.resourceControllerV2
}

func (c *clientSessionImpl) GlobalCatalogV1() *gcat.GlobalCatalogV1 {
	return c.globalCatalogV1
}

func (c *clientSessionImpl) ResourceManagerV2() *rmgrv2.ResourceManagerV2 {
	return c.resourceManagerV2
}

func (c *clientSessionImpl) GlobalTaggingV1() *gtagv1.GlobalTaggingV1 {
	return c.globalTaggingV1
}

// StrPtr2Bytes converts the supplied string pointer to a byte array
// and returns nil for nil pointer
func StrPtr2Bytes(v *string) []byte {
	if v == nil {
		return nil
	}
	return []byte(*v)
}

// BoolValue converts the supplied bool pointer to an bool, returning false if
// the pointer is nil.
func BoolValue(v *bool) bool {
	if v == nil {
		return false
	}
	return *v
}

// Int64Ptr converts the supplied int64 to a pointer to that int64.
func Int64Ptr(p int64) *int64 { return &p }

// BoolPtr converts the supplied bool to a pointer to that bool
func BoolPtr(p bool) *bool { return &p }

// MapToRawExtension - create a Map from a RawExtension
func MapToRawExtension(in map[string]interface{}) *runtime.RawExtension {
	if len(in) == 0 {
		return nil
	}
	js, _ := json.Marshal(in)
	o := &runtime.RawExtension{
		Raw: js,
	}
	return o
}

// RawExtensionToMap - create a RawExtension from a Map
func RawExtensionToMap(in *runtime.RawExtension) map[string]interface{} {
	if in == nil {
		return nil
	}
	o := make(map[string]interface{})
	_ = json.Unmarshal(in.Raw, &o)
	return o
}

// DateTimeToMetaV1Time converts strfmt.DateTime to metav1.Time
func DateTimeToMetaV1Time(t *strfmt.DateTime) *metav1.Time {
	if t == nil {
		return nil
	}
	tx := metav1.NewTime(time.Time(*t))
	return &tx
}

// TagsDiff computes the difference between desired tags and actual tags and returns
// a list of tags to attach and to detach
func TagsDiff(desired, actual []string) (toAttach, toDetach []string) {
	toAttach = []string{}
	toDetach = []string{}
	dMap := map[string]bool{}
	aMap := map[string]bool{}
	for _, d := range desired {
		dMap[d] = true
	}
	for _, a := range actual {
		aMap[a] = true
	}

	for _, d := range desired {
		_, ok := aMap[d]
		if !ok {
			toAttach = append(toAttach, d)
		}
	}

	for _, a := range actual {
		_, ok := dMap[a]
		if !ok {
			toDetach = append(toDetach, a)
		}
	}
	return toAttach, toDetach
}

// ConvertVarsMap connection vars map to format used in secret
func ConvertVarsMap(in map[string]interface{}) map[string][]byte {
	o := map[string][]byte{}
	for k, v := range in {
		o[k] = []byte(fmt.Sprintf("%s", v))
	}
	return o
}

// ConvertStructToMap - converts any struct to a map[string]interface{}
func ConvertStructToMap(in interface{}) (map[string]interface{}, error) {
	o := map[string]interface{}{}
	j, err := json.Marshal(in)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(j, &o); err != nil {
		return nil, err
	}
	return o, nil
}

// IsResourceGone returns true if resource is gone
func IsResourceGone(err error) bool {
	return strings.Contains(err.Error(), errGone) ||
		strings.Contains(err.Error(), http.StatusText(http.StatusNotFound))
}

// IsResourceInactive returns true if resource is inactive
func IsResourceInactive(err error) bool {
	return strings.Contains(err.Error(), errRemovedInvalid)
}

// IsResourceNotFound returns true if the SDK returns a not found error
func IsResourceNotFound(err error) bool {
	return strings.Contains(err.Error(), errNotFound)
}

// IsResourcePendingReclamation returns true if instance is being already deleted
func IsResourcePendingReclamation(err error) bool {
	return strings.Contains(err.Error(), errPendingReclamation) ||
		strings.Contains(err.Error(), http.StatusText(http.StatusNotFound))
}