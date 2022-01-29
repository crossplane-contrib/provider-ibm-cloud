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

	"github.com/crossplane-contrib/provider-ibm-cloud/apis/eventstreamsadminv1/v1alpha1"
	ibmc "github.com/crossplane-contrib/provider-ibm-cloud/pkg/clients"
	"github.com/crossplane-contrib/provider-ibm-cloud/pkg/controller/tstutil"
)

var (
	tName = "myTopic"
)

var _ managed.ExternalConnecter = &topicConnector{}
var _ managed.ExternalClient = &topicExternal{}

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
		Name:  "cleanup.policy",
		Value: "myCleanupPolicy",
	}
	o = append(o, c)
	c = v1alpha1.ConfigCreate{
		Name:  "retention.bytes",
		Value: "myRetentionBytes",
	}
	o = append(o, c)
	c = v1alpha1.ConfigCreate{
		Name:  "retention.ms",
		Value: "myRetentionMs",
	}
	o = append(o, c)
	c = v1alpha1.ConfigCreate{
		Name:  "segment.bytes",
		Value: "mySegmentBytes",
	}
	o = append(o, c)
	c = v1alpha1.ConfigCreate{
		Name:  "segment.index.bytes",
		Value: "mySegmentIndexBytes",
	}
	o = append(o, c)
	c = v1alpha1.ConfigCreate{
		Name:  "segment.ms",
		Value: "mySegmentMs",
	}
	o = append(o, c)
	return o
}

func tEmptyObservation(m ...func(*v1alpha1.TopicObservation)) *v1alpha1.TopicObservation {
	o := &v1alpha1.TopicObservation{
		ReplicationFactor:  0,
		RetentionMs:        0,
		CleanupPolicy:      "",
		Configs:            nil,
		ReplicaAssignments: nil,
		State:              "",
	}

	for _, f := range m {
		f(o)
	}
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
		State:              "active",
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

// Sets up a unit test http server, and creates an external topic structure appropriate for unit test.
//
// Params
//	   testingObj - the test object
//	   handlers - the handlers that create the responses
//	   client - the controller runtime client
//
// Returns
//		- the external object, ready for unit test
//		- the test http server, on which the caller should call 'defer ....Close()' (reason for this is we need to keep it around to prevent
//		  garbage collection)
//      -- an error (if...)
func setupServerAndGetUnitTestExternal(testingObj *testing.T, handlers *[]tstutil.Handler, kube *client.Client) (*topicExternal, *httptest.Server, error) {
	mClient, tstServer, err := tstutil.SetupTestServerClient(testingObj, handlers)
	if err != nil || mClient == nil || tstServer == nil {
		return nil, nil, err
	}

	return &topicExternal{
			kube:   *kube,
			client: *mClient,
			logger: logging.NewNopLogger(),
		},
		tstServer,
		nil
}

func TestTopicObserve(t *testing.T) {
	type want struct {
		mg  resource.Managed
		obs managed.ExternalObservation
		err error
	}
	cases := map[string]struct {
		handlers []tstutil.Handler
		kube     client.Client
		args     tstutil.Args
		want     want
	}{
		"NotFound": {
			handlers: []tstutil.Handler{
				{
					Path: "/",
					HandlerFunc: func(w http.ResponseWriter, r *http.Request) {
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
			args: tstutil.Args{
				Managed: topic(),
			},
			want: want{
				mg:  topic(),
				err: nil,
			},
		},
		"GetFailed": {
			handlers: []tstutil.Handler{
				{
					Path: "/",
					HandlerFunc: func(w http.ResponseWriter, r *http.Request) {
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
			args: tstutil.Args{
				Managed: topic(),
			},
			want: want{
				mg:  topic(),
				err: errors.New(errGetTopicFailed + ": Bad Request"),
			},
		},
		"ObservedTopicUpToDate": {
			handlers: []tstutil.Handler{
				{
					Path: "/",
					HandlerFunc: func(w http.ResponseWriter, r *http.Request) {
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
			args: tstutil.Args{
				Managed: topic(
					tWithExternalNameAnnotation(tName),
					tWithSpec(*tParams()),
					tWithStatus(*tEmptyObservation(func(p *v1alpha1.TopicObservation) { p.State = "active" })),
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
			handlers: []tstutil.Handler{
				{
					Path: "/",
					HandlerFunc: func(w http.ResponseWriter, r *http.Request) {
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
			args: tstutil.Args{
				Managed: topic(
					tWithExternalNameAnnotation(tName),
					tWithSpec(*tParams()),
					tWithStatus(*tEmptyObservation(func(p *v1alpha1.TopicObservation) { p.State = "active" })),
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
			e, server, err := setupServerAndGetUnitTestExternal(t, &tc.handlers, &tc.kube)
			if err != nil {
				t.Errorf("Create(...): problem setting up the test server %s", err)
			}

			defer server.Close()

			obs, err := e.Observe(context.Background(), tc.args.Managed)
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
			if diff := cmp.Diff(tc.want.mg, tc.args.Managed); diff != "" {
				t.Errorf("Observe(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestTopicCreate(t *testing.T) {
	type want struct {
		mg  resource.Managed
		cre managed.ExternalCreation
		err error
	}
	cases := map[string]struct {
		handlers []tstutil.Handler
		kube     client.Client
		args     tstutil.Args
		want     want
	}{
		"Successful": {
			handlers: []tstutil.Handler{
				{
					Path: "/",
					HandlerFunc: func(w http.ResponseWriter, r *http.Request) {
						if diff := cmp.Diff(http.MethodPost, r.Method); diff != "" {
							t.Errorf("r: -want, +got:\n%s", diff)
						}
						w.Header().Set("Content-Type", "application/json")
						w.WriteHeader(http.StatusCreated)
						_ = r.Body.Close()
						t := tInstance()
						_ = json.NewEncoder(w).Encode(t)
					},
				},
			},
			args: tstutil.Args{
				Managed: topic(tWithSpec(*tParams())),
			},
			want: want{
				mg: topic(tWithSpec(*tParams()),
					tWithConditions(cpv1alpha1.Creating()),
					tWithExternalNameAnnotation(tName)),
				cre: managed.ExternalCreation{ExternalNameAssigned: true},
				err: nil,
			},
		},
		"Failed": {
			handlers: []tstutil.Handler{
				{
					Path: "/",
					HandlerFunc: func(w http.ResponseWriter, r *http.Request) {
						if diff := cmp.Diff(http.MethodPost, r.Method); diff != "" {
							t.Errorf("r: -want, +got:\n%s", diff)
						}
						w.Header().Set("Content-Type", "application/json")
						w.WriteHeader(http.StatusBadRequest)
						_ = r.Body.Close()
						t := tInstance()
						_ = json.NewEncoder(w).Encode(t)
					},
				},
			},
			args: tstutil.Args{
				Managed: topic(tWithSpec(*tParams())),
			},
			want: want{
				mg: topic(tWithSpec(*tParams()),
					tWithConditions(cpv1alpha1.Creating())),
				cre: managed.ExternalCreation{ExternalNameAssigned: false},
				err: errors.Wrap(errors.New(http.StatusText(http.StatusBadRequest)), errCreateTopic),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e, server, err := setupServerAndGetUnitTestExternal(t, &tc.handlers, &tc.kube)
			if err != nil {
				t.Errorf("Create(...): problem setting up the test server %s", err)
			}

			defer server.Close()

			cre, err := e.Create(context.Background(), tc.args.Managed)
			if tc.want.err != nil && err != nil {
				// the case where our mock server returns error.
				if diff := cmp.Diff(tc.want.err.Error(), err.Error()); diff != "" {
					t.Errorf("Create(...): -want, +got:\n%s", diff)
				}
			} else {
				if diff := cmp.Diff(tc.want.err, err); diff != "" {
					t.Errorf("Create(...): -want, +got:\n%s", diff)
				}
			}
			if diff := cmp.Diff(tc.want.cre, cre); diff != "" {
				t.Errorf("Create(...): -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.mg, tc.args.Managed); diff != "" {
				t.Errorf("Create(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestTopicDelete(t *testing.T) {
	type want struct {
		mg  resource.Managed
		err error
	}
	cases := map[string]struct {
		handlers []tstutil.Handler
		kube     client.Client
		args     tstutil.Args
		want     want
	}{
		"Successful": {
			handlers: []tstutil.Handler{
				{
					Path: "/",
					HandlerFunc: func(w http.ResponseWriter, r *http.Request) {
						if diff := cmp.Diff(http.MethodDelete, r.Method); diff != "" {
							t.Errorf("r: -want, +got:\n%s", diff)
						}
						w.Header().Set("Content-Type", "application/json")
						w.WriteHeader(http.StatusAccepted)
						_ = r.Body.Close()
					},
				},
			},
			args: tstutil.Args{
				Managed: topic(tWithExternalNameAnnotation(tName)),
			},
			want: want{
				mg:  topic(tWithExternalNameAnnotation(tName), tWithConditions(cpv1alpha1.Deleting()), tWithStatus(*tEmptyObservation(func(p *v1alpha1.TopicObservation) { p.State = "terminating" }))),
				err: nil,
			},
		},
		"AlreadyGone": {
			handlers: []tstutil.Handler{
				{
					Path: "/",
					HandlerFunc: func(w http.ResponseWriter, r *http.Request) {
						if diff := cmp.Diff(http.MethodDelete, r.Method); diff != "" {
							t.Errorf("r: -want, +got:\n%s", diff)
						}
						w.Header().Set("Content-Type", "application/json")
						w.WriteHeader(http.StatusNotFound)
						_ = r.Body.Close()
					},
				},
			},
			args: tstutil.Args{
				Managed: topic(tWithExternalNameAnnotation(tName)),
			},
			want: want{
				mg:  topic(tWithExternalNameAnnotation(tName), tWithConditions(cpv1alpha1.Deleting())),
				err: nil,
			},
		},
		"Failed": {
			handlers: []tstutil.Handler{
				{
					Path: "/",
					HandlerFunc: func(w http.ResponseWriter, r *http.Request) {
						if diff := cmp.Diff(http.MethodDelete, r.Method); diff != "" {
							t.Errorf("r: -want, +got:\n%s", diff)
						}
						w.Header().Set("Content-Type", "application/json")
						w.WriteHeader(http.StatusBadRequest)
						_ = r.Body.Close()
					},
				},
			},
			args: tstutil.Args{
				Managed: topic(tWithExternalNameAnnotation(tName)),
			},
			want: want{
				mg:  topic(tWithExternalNameAnnotation(tName), tWithConditions(cpv1alpha1.Deleting())),
				err: errors.Wrap(errors.New(http.StatusText(http.StatusBadRequest)), errDeleteTopic),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e, server, errServer := setupServerAndGetUnitTestExternal(t, &tc.handlers, &tc.kube)
			if errServer != nil {
				t.Errorf("Create(...): problem setting up the test server %s", errServer)
			}

			defer server.Close()

			err := e.Delete(context.Background(), tc.args.Managed)
			if tc.want.err != nil && err != nil {
				// the case where our mock server returns error.
				if diff := cmp.Diff(tc.want.err.Error(), err.Error()); diff != "" {
					t.Errorf("Delete(...): -want, +got:\n%s", diff)
				}
			} else {
				if diff := cmp.Diff(tc.want.err, err); diff != "" {
					t.Errorf("Delete(...): -want, +got:\n%s", diff)
				}
			}
			if diff := cmp.Diff(tc.want.mg, tc.args.Managed); diff != "" {
				t.Errorf("Delete(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestTopicUpdate(t *testing.T) {
	type want struct {
		mg  resource.Managed
		upd managed.ExternalUpdate
		err error
	}
	cases := map[string]struct {
		handlers []tstutil.Handler
		kube     client.Client
		args     tstutil.Args
		want     want
	}{
		"Successful": {
			handlers: []tstutil.Handler{
				{
					Path: "/",
					HandlerFunc: func(w http.ResponseWriter, r *http.Request) {
						if diff := cmp.Diff(http.MethodPatch, r.Method); diff == "" {
							return
						}
						w.Header().Set("Content-Type", "application/json")
						w.WriteHeader(http.StatusOK)
						_ = r.Body.Close()
						t := tInstance()
						_ = json.NewEncoder(w).Encode(t)
					},
				},
			},
			args: tstutil.Args{
				Managed: topic(tWithSpec(*tParams()), tWithExternalNameAnnotation(tName)),
			},
			want: want{
				mg:  topic(tWithSpec(*tParams())),
				upd: managed.ExternalUpdate{ConnectionDetails: nil},
				err: nil,
			},
		},
		"PatchFails": {
			handlers: []tstutil.Handler{
				{
					Path: "/",
					HandlerFunc: func(w http.ResponseWriter, r *http.Request) {
						if diff := cmp.Diff(http.MethodPatch, r.Method); diff == "" {
							return
						}
						w.Header().Set("Content-Type", "application/json")
						w.WriteHeader(http.StatusBadRequest)
						_ = r.Body.Close()
					},
				},
			},

			args: tstutil.Args{
				Managed: topic(tWithSpec(*tParams())),
			},
			want: want{
				mg:  topic(tWithSpec(*tParams())),
				err: errors.Wrap(errors.New(http.StatusText(http.StatusBadRequest)), errGetTopicFailed),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e, server, errServer := setupServerAndGetUnitTestExternal(t, &tc.handlers, &tc.kube)
			if errServer != nil {
				t.Errorf("Create(...): problem setting up the test server %s", errServer)
			}

			defer server.Close()

			upd, err := e.Update(context.Background(), tc.args.Managed)
			if tc.want.err != nil && err != nil {
				// the case where our mock server returns error.
				if diff := cmp.Diff(tc.want.err.Error(), err.Error()); diff != "" {
					t.Errorf("Update(...): -want, +got:\n%s", diff)
				}
			} else {
				if diff := cmp.Diff(tc.want.err, err); diff != "" {
					t.Errorf("Update(...): -want, +got:\n%s", diff)
				}
			}
			if tc.want.err == nil {
				if diff := cmp.Diff(tc.want.mg, tc.args.Managed); diff != "" {
					t.Errorf("Update(...): -want, +got:\n%s", diff)
				}
				if diff := cmp.Diff(tc.want.upd, upd); diff != "" {
					t.Errorf("Update(...): -want, +got:\n%s", diff)
				}
			}
		})
	}
}
