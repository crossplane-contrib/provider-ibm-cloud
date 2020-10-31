package clients

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	crv1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"

	"github.com/IBM/go-sdk-core/core"

	"github.com/crossplane-contrib/provider-ibm-cloud/apis/resourcecontrollerv2/v1alpha1"
	"github.com/crossplane-contrib/provider-ibm-cloud/apis/v1beta1"
)

func TestGetAuthInfo(t *testing.T) {
	const (
		providerName = "provider-ibm-cloud"
		secretName   = "ibm-cloud-creds"
		creds        = "xxx-yyy-zzz"
		key          = "apikey"
		instanceName = "myinstance"
		uid          = "1fce7df8-8615-4b5c-97cb-2309cd0b9b23"
		fakeTok      = "Bearer xyz"
	)

	pcu := &v1beta1.ProviderConfigUsage{
		ObjectMeta: metav1.ObjectMeta{
			Name: providerName,
		},
	}

	pc := &v1beta1.ProviderConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name: providerName,
		},
		Spec: v1beta1.ProviderConfigSpec{
			ProviderConfigSpec: crv1alpha1.ProviderConfigSpec{
				Credentials: crv1alpha1.ProviderCredentials{
					SecretRef: &crv1alpha1.SecretKeySelector{
						SecretReference: crv1alpha1.SecretReference{
							Name: secretName,
						},
						Key: key,
					},
				},
			},
		},
	}

	m := &v1alpha1.ResourceInstance{
		ObjectMeta: metav1.ObjectMeta{
			Name: instanceName,
			UID:  uid,
		},
		Spec: v1alpha1.ResourceInstanceSpec{},
	}
	m.SetProviderConfigReference(&crv1alpha1.Reference{Name: providerName})

	objs := []runtime.Object{pc, pcu}
	s := scheme.Scheme
	s.AddKnownTypes(v1beta1.SchemeGroupVersion, pc, pcu)
	mockClient := fake.NewFakeClient(objs...)

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: secretName,
		},
		Data: map[string][]byte{
			key:            []byte(creds),
			AccessTokenKey: []byte(fakeTok),
		},
	}

	err := mockClient.Create(context.TODO(), secret)
	assert.NoError(t, err)

	opts, err := GetAuthInfo(context.TODO(), mockClient, m)
	assert.NoError(t, err)

	expAuth := &core.BearerTokenAuthenticator{
		BearerToken: "xyz",
	}

	assert.Equal(t, expAuth, opts.Authenticator)
}

func TestTagsDiff(t *testing.T) {
	desired := []string{"tag1", "tag2", "tag3"}
	actual := []string{"tag1"}
	expToAttach := []string{"tag2", "tag3"}
	expToDetach := []string{}

	toAttach, toDetach := TagsDiff(desired, actual)
	assert.Equal(t, expToAttach, toAttach)
	assert.Equal(t, expToDetach, toDetach)

	desired = []string{"tag1"}
	actual = []string{"tag1", "tag2", "tag3"}
	expToAttach = []string{}
	expToDetach = []string{"tag2", "tag3"}

	toAttach, toDetach = TagsDiff(desired, actual)
	assert.Equal(t, expToAttach, toAttach)
	assert.Equal(t, expToDetach, toDetach)

	desired = []string{"tag1", "tag2", "tag3"}
	actual = []string{"tag1", "tag2", "tag3"}
	expToAttach = []string{}
	expToDetach = []string{}

	toAttach, toDetach = TagsDiff(desired, actual)
	assert.Equal(t, expToAttach, toAttach)
	assert.Equal(t, expToDetach, toDetach)

}
