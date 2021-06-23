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

package eventstreamsadminv1

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	cpv1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/reference"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/crossplane/crossplane-runtime/pkg/test"

	arv1 "github.com/IBM/eventstreams-go-sdk/pkg/adminrestv1"
	"github.com/IBM/go-sdk-core/core"

	"github.com/crossplane-contrib/provider-ibm-cloud/apis/eventstreamsadminv1/v1alpha1"
	ibmc "github.com/crossplane-contrib/provider-ibm-cloud/pkg/clients"
)

const (
	bearerTok = "mock-token"
)

var (
	tName = "myTopic"
)

var _ managed.ExternalConnecter = &topicConnector{}
var _ managed.ExternalClient = &topicExternal{}

type handler struct {
	path        string
	handlerFunc func(w http.ResponseWriter, r *http.Request)
}

type tModifier func(*v1alpha1.Topic)

func topic(im ...tModifier) *v1alpha1.Topic {
	i := &v1alpha1.Topic{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "myTopic",
			Finalizers: []string{},
			Annotations: map[string]string{
				meta.AnnotationKeyExternalName: "myTopic",
			},
		},
		Spec: v1alpha1.TopicSpec{
			ForProvider: v1alpha1.TopicParameters{},
		},
	}
	for _, m := range im {
		m(i)
	}
	return i
}

func tWithExternalNameAnnotation(externalName string) tModifier {
	return func(i *v1alpha1.Topic) {
		if i.ObjectMeta.Annotations == nil {
			i.ObjectMeta.Annotations = make(map[string]string)
		}
		i.ObjectMeta.Annotations[meta.AnnotationKeyExternalName] = externalName
	}
}

func tWithSpec(p v1alpha1.TopicParameters) tModifier {
	return func(r *v1alpha1.Topic) { r.Spec.ForProvider = p }
}

func tWithConditions(c ...cpv1alpha1.Condition) tModifier {
	return func(i *v1alpha1.Topic) { i.Status.SetConditions(c...) }
}

func tWithStatus(p v1alpha1.TopicObservation) tModifier {
	return func(r *v1alpha1.Topic) { r.Status.AtProvider = p }
}

func tParams(m ...func(*v1alpha1.TopicParameters)) *v1alpha1.TopicParameters {
	p := &v1alpha1.TopicParameters{
		Name:                  "myTopic",
		KafkaAdminURL:         reference.ToPtrValue("myKafkaAdminURL"),
		KafkaAdminURLRef:      &cpv1alpha1.Reference{},
		KafkaAdminURLSelector: &cpv1alpha1.Selector{},
		Partitions:            ibmc.Int64Ptr(int64(2)),
		PartitionCount:        ibmc.Int64Ptr(int64(2)),
		// Configs:               []v1alpha1.ConfigCreate{},
		// can test empty ConfigCreate or generate ConfigCreate to test
		Configs: generateTestv1alpha1ConfigCreate(),
	}
	for _, f := range m {
		f(p)
	}
	return p
}

func generateTestv1alpha1ConfigCreate() []v1alpha1.ConfigCreate {
	o := []v1alpha1.ConfigCreate{}

	c := v1alpha1.ConfigCreate{
		Name:  "CleanupPolicy",
		Value: "myCleanupPolicy",
	}
	o = append(o, c)
	c = v1alpha1.ConfigCreate{
		Name:  "MinInsyncReplicas",
		Value: "myMinInsyncReplicas",
	}
	o = append(o, c)
	c = v1alpha1.ConfigCreate{
		Name:  "RetentionBytes",
		Value: "myRetentionBytes",
	}
	o = append(o, c)
	c = v1alpha1.ConfigCreate{
		Name:  "RetentionMs",
		Value: "myRetentionMs",
	}
	o = append(o, c)
	c = v1alpha1.ConfigCreate{
		Name:  "SegmentBytes",
		Value: "mySegmentBytes",
	}
	o = append(o, c)
	c = v1alpha1.ConfigCreate{
		Name:  "SegmentIndexBytes",
		Value: "mySegmentIndexBytes",
	}
	o = append(o, c)
	c = v1alpha1.ConfigCreate{
		Name:  "SegmentMs",
		Value: "mySegmentMs",
	}
	o = append(o, c)
	return o
}

func tObservation(m ...func(*v1alpha1.TopicObservation)) *v1alpha1.TopicObservation {
	o := &v1alpha1.TopicObservation{
		ReplicationFactor: int64(2),
		RetentionMs:       int64(2),
		CleanupPolicy:     "myCleanupPolicy",
		// Configs:           &v1alpha1.TopicConfigs{},
		Configs: generateTestv1alpha1TopicConfigs(),
		// ReplicaAssignments: []v1alpha1.ReplicaAssignment{},
		ReplicaAssignments: generateTestv1alpha1ReplicaAssignments(),
	}

	for _, f := range m {
		f(o)
	}
	return o
}

func generateTestv1alpha1TopicConfigs() *v1alpha1.TopicConfigs {
	o := &v1alpha1.TopicConfigs{
		CleanupPolicy:     "myCleanupPolicy",
		MinInsyncReplicas: "myMinInsyncReplicas",
		RetentionBytes:    "myRetentionBytes",
		RetentionMs:       "myRetentionMs",
		SegmentBytes:      "mySegmentBytes",
		SegmentIndexBytes: "mySegmentIndexBytes",
		SegmentMs:         "mySegmentMs",
	}
	return o
}

func generateTestv1alpha1ReplicaAssignments() []v1alpha1.ReplicaAssignment {
	o := []v1alpha1.ReplicaAssignment{}

	c := v1alpha1.ReplicaAssignment{
		ID:      int64(2),
		Brokers: generateTestv1alpha1ReplicaAssignmentBrokers(),
	}
	o = append(o, c)

	c = v1alpha1.ReplicaAssignment{
		ID:      int64(3),
		Brokers: generateTestv1alpha1ReplicaAssignmentBrokers(),
	}

	o = append(o, c)

	return o
}

func generateTestv1alpha1ReplicaAssignmentBrokers() *v1alpha1.ReplicaAssignmentBrokers {
	o := &v1alpha1.ReplicaAssignmentBrokers{
		Replicas: generateTestv1alpha1Replicas(),
	}
	return o
}

func generateTestv1alpha1Replicas() []int64 {
	o := []int64{}

	c := int64(2)

	o = append(o, c)

	c = int64(3)

	o = append(o, c)

	return o
}

func tInstance(m ...func(*arv1.TopicDetail)) *arv1.TopicDetail {
	i := &arv1.TopicDetail{
		Name:              reference.ToPtrValue("myTopic"),
		Partitions:        ibmc.Int64Ptr(int64(2)),
		ReplicationFactor: ibmc.Int64Ptr(int64(2)),
		RetentionMs:       ibmc.Int64Ptr(int64(2)),
		CleanupPolicy:     reference.ToPtrValue("myCleanupPolicy"),
		// Configs:           &arv1.TopicConfigs{},
		Configs: generateTestarv1TopicConfigs(),
		// ReplicaAssignments: []arv1.ReplicaAssignment{},
		ReplicaAssignments: generateTestarv1ReplicaAssignments(),
	}

	for _, f := range m {
		f(i)
	}
	return i
}

func generateTestarv1TopicConfigs() *arv1.TopicConfigs {
	o := &arv1.TopicConfigs{
		CleanupPolicy:     reference.ToPtrValue("myCleanupPolicy"),
		MinInsyncReplicas: reference.ToPtrValue("myMinInsyncReplicas"),
		RetentionBytes:    reference.ToPtrValue("myRetentionBytes"),
		RetentionMs:       reference.ToPtrValue("myRetentionMs"),
		SegmentBytes:      reference.ToPtrValue("mySegmentBytes"),
		SegmentIndexBytes: reference.ToPtrValue("mySegmentIndexBytes"),
		SegmentMs:         reference.ToPtrValue("mySegmentMs"),
	}
	return o
}

func generateTestarv1ReplicaAssignments() []arv1.ReplicaAssignment {
	o := []arv1.ReplicaAssignment{}

	c := arv1.ReplicaAssignment{
		ID:      ibmc.Int64Ptr(int64(2)),
		Brokers: generateTestarv1ReplicaAssignmentBrokers(),
	}
	o = append(o, c)

	c = arv1.ReplicaAssignment{
		ID:      ibmc.Int64Ptr(int64(3)),
		Brokers: generateTestarv1ReplicaAssignmentBrokers(),
	}

	o = append(o, c)

	return o
}

func generateTestarv1ReplicaAssignmentBrokers() *arv1.ReplicaAssignmentBrokers {
	o := &arv1.ReplicaAssignmentBrokers{
		Replicas: generateTestarv1Replicas(),
	}
	return o
}

func generateTestarv1Replicas() []int64 {
	o := []int64{}

	c := int64(2)

	o = append(o, c)

	c = int64(3)

	o = append(o, c)

	return o
}

// the error comes when I run the test using none of the generate functions,
// ie all generate test array functions above are commented and the empty arrays and structs are used instead
// I think when the tInstance is encoded ie (_ = json.NewEncoder(w).Encode(t)) the empty ReplicaAssignments array is maybe encoded as nil (but I'm not entirely sure)
// references: https://golang.org/pkg/encoding/json/     https://pkg.go.dev/encoding/json#Encoder.Encode
// so when
// instance, _, err := c.client.AdminrestV1().GetTopic(&arv1.GetTopicOptions{TopicName: reference.ToPtrValue(meta.GetExternalName(cr))}) is run in Observe
// ReplicaAssignments is nil even though it should be an empty array ??

// but the array should never be empty, it should either have values in it or be nil (so the encoder should never encounter an empty array),
// so I don't think this problem is a big problem
// also the test works when the arrays are full, ie when the generate array functions are used for ReplicaAssignments etc.
func TestTopicObserve(t *testing.T) {
	type args struct {
		mg resource.Managed
	}
	type want struct {
		mg  resource.Managed
		obs managed.ExternalObservation
		err error
	}
	cases := map[string]struct {
		handlers []handler
		kube     client.Client
		args     args
		want     want
	}{
		"NotFound": {
			handlers: []handler{
				{
					path: "/",
					handlerFunc: func(w http.ResponseWriter, r *http.Request) {
						_ = r.Body.Close()
						if diff := cmp.Diff(http.MethodGet, r.Method); diff != "" {
							t.Errorf("r: -want, +got:\n%s", diff)
						}
						// content type should always set before writeHeader()
						w.Header().Set("Content-Type", "application/json")
						w.WriteHeader(http.StatusNotFound)
						_ = json.NewEncoder(w).Encode(&arv1.TopicDetail{})
					},
				},
			},
			args: args{
				mg: topic(),
			},
			want: want{
				mg:  topic(),
				err: nil,
			},
		},
		"GetFailed": {
			handlers: []handler{
				{
					path: "/",
					handlerFunc: func(w http.ResponseWriter, r *http.Request) {
						_ = r.Body.Close()
						if diff := cmp.Diff(http.MethodGet, r.Method); diff != "" {
							t.Errorf("r: -want, +got:\n%s", diff)
						}
						w.Header().Set("Content-Type", "application/json")
						w.WriteHeader(http.StatusBadRequest)
						_ = json.NewEncoder(w).Encode(&arv1.TopicDetail{})
					},
				},
			},
			args: args{
				mg: topic(),
			},
			want: want{
				mg:  topic(),
				err: errors.New(errGetTopicFailed + ": Bad Request"),
			},
		},
		"ObservedTopicUpToDate": {
			handlers: []handler{
				{
					path: "/",
					handlerFunc: func(w http.ResponseWriter, r *http.Request) {
						_ = r.Body.Close()
						if diff := cmp.Diff(http.MethodGet, r.Method); diff != "" {
							t.Errorf("r: -want, +got:\n%s", diff)
						}
						w.Header().Set("Content-Type", "application/json")
						t := tInstance()
						_ = json.NewEncoder(w).Encode(t)
					},
				},
			},
			kube: &test.MockClient{
				MockUpdate: test.NewMockUpdateFn(nil),
			},
			args: args{
				mg: topic(
					tWithExternalNameAnnotation(tName),
					tWithSpec(*tParams()),
				),
			},
			want: want{
				mg: topic(tWithSpec(*tParams()),
					tWithConditions(cpv1alpha1.Available()),
					tWithStatus(*tObservation()),
				),
				obs: managed.ExternalObservation{
					ResourceExists:    true,
					ResourceUpToDate:  true,
					ConnectionDetails: nil,
				},
			},
		},
		"ObservedTopicNotUpToDate": {
			handlers: []handler{
				{
					path: "/",
					handlerFunc: func(w http.ResponseWriter, r *http.Request) {
						_ = r.Body.Close()
						if diff := cmp.Diff(http.MethodGet, r.Method); diff != "" {
							t.Errorf("r: -want, +got:\n%s", diff)
						}
						w.Header().Set("Content-Type", "application/json")
						t := tInstance(func(p *arv1.TopicDetail) {
							p.Partitions = ibmc.Int64Ptr(int64(3))
						})
						_ = json.NewEncoder(w).Encode(t)
					},
				},
			},
			kube: &test.MockClient{
				MockUpdate: test.NewMockUpdateFn(nil),
			},
			args: args{
				mg: topic(
					tWithExternalNameAnnotation(tName),
					tWithSpec(*tParams()),
				),
			},
			want: want{
				mg: topic(tWithSpec(*tParams()),
					tWithConditions(cpv1alpha1.Available()),
					tWithStatus(*tObservation()),
				),
				obs: managed.ExternalObservation{
					ResourceExists:    true,
					ResourceUpToDate:  false,
					ConnectionDetails: nil,
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			mux := http.NewServeMux()
			for _, h := range tc.handlers {
				mux.HandleFunc(h.path, h.handlerFunc)
			}
			server := httptest.NewServer(mux)
			defer server.Close()

			opts := ibmc.ClientOptions{URL: server.URL, Authenticator: &core.BearerTokenAuthenticator{
				BearerToken: bearerTok,
			}}
			mClient, _ := ibmc.NewClient(opts)
			e := topicExternal{
				kube:   tc.kube,
				client: mClient,
				logger: logging.NewNopLogger(),
			}
			obs, err := e.Observe(context.Background(), tc.args.mg)
			if tc.want.err != nil && err != nil {
				// the case where our mock server returns error.
				if diff := cmp.Diff(tc.want.err.Error(), err.Error()); diff != "" {
					t.Errorf("Observe(...): want error string != got error string:\n%s", diff)
				}
			} else {
				if diff := cmp.Diff(tc.want.err, err); diff != "" {
					t.Errorf("Observe(...): want error != got error:\n%s", diff)
				}
			}
			if diff := cmp.Diff(tc.want.obs, obs); diff != "" {
				t.Errorf("Observe(...): -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.mg, tc.args.mg); diff != "" {
				t.Errorf("Observe(...): -want, +got:\n%s", diff)
			}
		})
	}
}
