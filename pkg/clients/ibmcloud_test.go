package clients

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	crv1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"

	"github.com/IBM/go-sdk-core/core"

	"github.com/crossplane-contrib/provider-ibm-cloud/apis/resourcecontrollerv2/v1alpha1"
	"github.com/crossplane-contrib/provider-ibm-cloud/apis/v1beta1"
)

const (
	providerName = "provider-ibm-cloud"
	secretName   = "ibm-cloud-creds"
	creds        = "xxx-yyy-zzz"
	key          = "apikey"
	instanceName = "myinstance"
	uid          = "1fce7df8-8615-4b5c-97cb-2309cd0b9b23"
	btok         = "xyz"
)

var (
	fakeTok = "Bearer " + btok
)

func getInitializedMockClient(t *testing.T) client.Client {
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

	objs := []runtime.Object{pc, pcu}
	s := scheme.Scheme
	s.AddKnownTypes(v1beta1.SchemeGroupVersion, pc, pcu)
	c := fake.NewFakeClient(objs...)

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: secretName,
		},
		Data: map[string][]byte{
			key:            []byte(creds),
			AccessTokenKey: []byte(fakeTok),
		},
	}
	err := c.Create(context.TODO(), secret)
	if err != nil {
		t.Error("unexpected error: ", err)
	}
	return c
}

func resourceInstance() *v1alpha1.ResourceInstance {
	i := &v1alpha1.ResourceInstance{
		ObjectMeta: metav1.ObjectMeta{
			Name: instanceName,
			UID:  uid,
		},
		Spec: v1alpha1.ResourceInstanceSpec{},
	}
	i.SetProviderConfigReference(&crv1alpha1.Reference{Name: providerName})
	return i
}

func TestGetAuthInfo(t *testing.T) {
	type args struct {
		cr *v1alpha1.ResourceInstance
	}
	type want struct {
		auth core.Authenticator
	}
	cases := map[string]struct {
		args args
		want want
	}{
		"AddTags": {
			args: args{
				cr: resourceInstance(),
			},
			want: want{
				auth: &core.BearerTokenAuthenticator{
					BearerToken: btok,
				},
			},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			mClient := getInitializedMockClient(t)
			opts, err := GetAuthInfo(context.TODO(), mClient, tc.args.cr)
			if err != nil {
				t.Error("unexpected error: ", err)
			}

			if diff := cmp.Diff(tc.want.auth, opts.Authenticator); diff != "" {
				t.Errorf("TestGetAuthInfo(...): -want, +got:\n%s", diff)
			}

		})
	}
}

func TestTagsDiff(t *testing.T) {
	type args struct {
		desired []string
		actual  []string
	}
	type want struct {
		toAttach []string
		toDetach []string
	}
	cases := map[string]struct {
		args args
		want want
	}{
		"AddTags": {
			args: args{
				desired: []string{"tag1", "tag2", "tag3"},
				actual:  []string{"tag1"},
			},
			want: want{
				toAttach: []string{"tag2", "tag3"},
				toDetach: []string{},
			},
		},
		"DetachTags": {
			args: args{
				desired: []string{"tag1"},
				actual:  []string{"tag1", "tag2", "tag3"},
			},
			want: want{
				toAttach: []string{},
				toDetach: []string{"tag2", "tag3"},
			},
		},
		"NoAction": {
			args: args{
				desired: []string{"tag1", "tag2", "tag3"},
				actual:  []string{"tag1", "tag2", "tag3"},
			},
			want: want{
				toAttach: []string{},
				toDetach: []string{},
			},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			toAttach, toDetach := TagsDiff(tc.args.desired, tc.args.actual)
			if diff := cmp.Diff(tc.want.toAttach, toAttach); diff != "" {
				t.Errorf("TestTagsDiff(...):toAttach: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.toDetach, toDetach); diff != "" {
				t.Errorf("TestTagsDiff(...):toDetach: -want, +got:\n%s", diff)
			}

		})
	}
}
