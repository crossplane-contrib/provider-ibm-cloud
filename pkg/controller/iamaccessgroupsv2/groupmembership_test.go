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

package iamaccessgroupsv2

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	cpv1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/reference"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/crossplane/crossplane-runtime/pkg/test"

	"github.com/IBM/go-sdk-core/core"
	iamagv2 "github.com/IBM/platform-services-go-sdk/iamaccessgroupsv2"

	"github.com/crossplane-contrib/provider-ibm-cloud/apis/iamaccessgroupsv2/v1alpha1"
	ibmc "github.com/crossplane-contrib/provider-ibm-cloud/pkg/clients"
	ibmcgm "github.com/crossplane-contrib/provider-ibm-cloud/pkg/clients/groupmembership"
)

var (
	accessGroupID = "12345678-abcd-1a2b-a1b2-1234567890ab"
	totalCount    = int64(2)
	memberIamID1  = "IBMid-user1"
	memberName1   = "memberName1"
	memberEmail1  = "memberName1@email.com"
	memberDescr1  = "member description 1"
	memberHRef1   = "https://iam.cloud.ibm.com/v2/accessgroups/" + accessGroupID + "members/" + memberIamID1
	memberIamID2  = "IBMid-user2"
	memberName2   = "memberName2"
	memberEmail2  = "memberName2@email.com"
	memberDescr2  = "member description 2"
	memberHRef2   = "https://iam.cloud.ibm.com/v2/accessgroups/" + accessGroupID + "members/" + memberIamID2
	memberIamID3  = "IBMid-user3"
)

var _ managed.ExternalConnecter = &gmConnector{}
var _ managed.ExternalClient = &gmExternal{}

type gmModifier func(*v1alpha1.GroupMembership)

func gm(im ...gmModifier) *v1alpha1.GroupMembership {
	i := &v1alpha1.GroupMembership{
		ObjectMeta: metav1.ObjectMeta{
			Name:       agName,
			Finalizers: []string{},
			Annotations: map[string]string{
				meta.AnnotationKeyExternalName: agID,
			},
		},
		Spec: v1alpha1.GroupMembershipSpec{
			ForProvider: v1alpha1.GroupMembershipParameters{},
		},
	}
	for _, m := range im {
		m(i)
	}
	return i
}

func gmWithExternalNameAnnotation(externalName string) gmModifier {
	return func(i *v1alpha1.GroupMembership) {
		if i.ObjectMeta.Annotations == nil {
			i.ObjectMeta.Annotations = make(map[string]string)
		}
		i.ObjectMeta.Annotations[meta.AnnotationKeyExternalName] = externalName
	}
}

func gmWithEtagAnnotation(eTag string) gmModifier {
	return func(i *v1alpha1.GroupMembership) {
		if i.ObjectMeta.Annotations == nil {
			i.ObjectMeta.Annotations = make(map[string]string)
		}
		i.ObjectMeta.Annotations[ibmc.ETagAnnotation] = eTag
	}
}

func gmWithSpec(p v1alpha1.GroupMembershipParameters) gmModifier {
	return func(r *v1alpha1.GroupMembership) { r.Spec.ForProvider = p }
}

func gmWithConditions(c ...cpv1alpha1.Condition) gmModifier {
	return func(i *v1alpha1.GroupMembership) { i.Status.SetConditions(c...) }
}

func gmWithStatus(p v1alpha1.GroupMembershipObservation) gmModifier {
	return func(r *v1alpha1.GroupMembership) { r.Status.AtProvider = p }
}

func gmParams(m ...func(*v1alpha1.GroupMembershipParameters)) *v1alpha1.GroupMembershipParameters {
	p := &v1alpha1.GroupMembershipParameters{
		AccessGroupID: &accessGroupID,
		Members: []v1alpha1.AddGroupMembersRequestMembersItem{
			{
				IamID: memberIamID1,
				Type:  ibmcgm.MemberTypeUser,
			},
			{
				IamID: memberIamID2,
				Type:  ibmcgm.MemberTypeUser,
			},
		},
		TransactionID: &transactionID,
	}
	for _, f := range m {
		f(p)
	}
	return p
}

func gmObservation(m ...func(*v1alpha1.GroupMembershipObservation)) *v1alpha1.GroupMembershipObservation {
	o := &v1alpha1.GroupMembershipObservation{
		Members: []v1alpha1.ListGroupMembersResponseMember{
			{
				IamID:       memberIamID1,
				Type:        ibmcgm.MemberTypeUser,
				Name:        memberName1,
				Email:       memberEmail1,
				Description: memberDescr1,
				Href:        memberHRef1,
				CreatedAt:   ibmcgm.GenerateMetaV1Time(&createdAt),
				CreatedByID: createdByID,
			},
			{
				IamID:       memberIamID2,
				Type:        ibmcgm.MemberTypeUser,
				Name:        memberName2,
				Email:       memberEmail2,
				Description: memberDescr2,
				Href:        memberHRef2,
				CreatedAt:   ibmcgm.GenerateMetaV1Time(&createdAt),
				CreatedByID: createdByID,
			},
		},
	}

	for _, f := range m {
		f(o)
	}
	return o
}

func gmInstance(m ...func(*iamagv2.GroupMembersList)) *iamagv2.GroupMembersList {
	i := &iamagv2.GroupMembersList{
		TotalCount: &totalCount,
		Members: []iamagv2.ListGroupMembersResponseMember{
			{
				IamID:       &memberIamID1,
				Type:        reference.ToPtrValue(ibmcgm.MemberTypeUser),
				Name:        &memberName1,
				Email:       &memberEmail1,
				Description: &memberDescr1,
				Href:        &memberHRef1,
				CreatedAt:   &createdAt,
				CreatedByID: &createdByID,
			},
			{
				IamID:       &memberIamID2,
				Type:        reference.ToPtrValue(ibmcgm.MemberTypeUser),
				Name:        &memberName2,
				Email:       &memberEmail2,
				Description: &memberDescr2,
				Href:        &memberHRef2,
				CreatedAt:   &createdAt,
				CreatedByID: &createdByID,
			},
		},
	}
	for _, f := range m {
		f(i)
	}
	return i
}

var membersCache map[string]iamagv2.ListGroupMembersResponseMember

// handler to mock client SDK call to iam API
var iamMembersHandler = func(w http.ResponseWriter, r *http.Request) {
	resp := iamagv2.GroupMembersList{}
	switch r.Method {
	case http.MethodPut:
		body, _ := ioutil.ReadAll(r.Body)
		req := iamagv2.AddMembersToAccessGroupOptions{}
		json.Unmarshal(body, &req)
		for _, m := range req.Members {
			membersCache[*m.IamID] = iamagv2.ListGroupMembersResponseMember{
				IamID: m.IamID,
				Type:  reference.ToPtrValue(ibmcgm.MemberTypeUser),
			}
		}
		return
	case http.MethodPost:
		body, _ := ioutil.ReadAll(r.Body)
		req := iamagv2.RemoveMembersFromAccessGroupOptions{}
		json.Unmarshal(body, &req)
		for _, m := range req.Members {
			delete(membersCache, m)
		}
		return
	}
	_ = r.Body.Close()
	w.Header().Set("Content-Type", "application/json")

	members := []iamagv2.ListGroupMembersResponseMember{}
	for _, v := range membersCache {
		members = append(members, v)
	}
	resp.Members = members
	resp.TotalCount = ibmc.Int64Ptr(int64(len(members)))
	_ = json.NewEncoder(w).Encode(resp)
}

func TestGroupMembershipObserve(t *testing.T) {
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
						_ = json.NewEncoder(w).Encode(&iamagv2.GroupMembersList{})
					},
				},
			},
			args: args{
				mg: gm(),
			},
			want: want{
				mg:  gm(),
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
						_ = json.NewEncoder(w).Encode(&iamagv2.GroupMembersList{})
					},
				},
			},
			args: args{
				mg: gm(),
			},
			want: want{
				mg:  gm(),
				err: errors.New(errAgBadRequest),
			},
		},
		"GetForbidden": {
			handlers: []handler{
				{
					path: "/",
					handlerFunc: func(w http.ResponseWriter, r *http.Request) {
						_ = r.Body.Close()
						if diff := cmp.Diff(http.MethodGet, r.Method); diff != "" {
							t.Errorf("r: -want, +got:\n%s", diff)
						}
						w.Header().Set("Content-Type", "application/json")
						w.WriteHeader(http.StatusForbidden)
						_ = json.NewEncoder(w).Encode(&iamagv2.GroupMembersList{})
					},
				},
			},
			args: args{
				mg: gm(),
			},
			want: want{
				mg:  gm(),
				err: errors.New(errAgForbidden),
			},
		},
		"UpToDate": {
			handlers: []handler{
				{
					path: "/",
					handlerFunc: func(w http.ResponseWriter, r *http.Request) {
						_ = r.Body.Close()
						if diff := cmp.Diff(http.MethodGet, r.Method); diff != "" {
							t.Errorf("r: -want, +got:\n%s", diff)
						}
						w.Header().Set("Content-Type", "application/json")
						w.Header().Set("ETag", eTag)
						cr := gmInstance()
						_ = json.NewEncoder(w).Encode(cr)
					},
				},
			},
			kube: &test.MockClient{
				MockUpdate: test.NewMockUpdateFn(nil),
			},
			args: args{
				mg: gm(
					gmWithExternalNameAnnotation(agID),
					gmWithSpec(*gmParams()),
				),
			},
			want: want{
				mg: gm(gmWithSpec(*gmParams()),
					gmWithConditions(cpv1alpha1.Available()),
					gmWithStatus(*gmObservation(func(cro *v1alpha1.GroupMembershipObservation) {
						cro.State = ibmcgm.StateActive
					})),
					gmWithEtagAnnotation(eTag)),
				obs: managed.ExternalObservation{
					ResourceExists:    true,
					ResourceUpToDate:  true,
					ConnectionDetails: nil,
				},
			},
		},
		"NotUpToDate": {
			handlers: []handler{
				{
					path: "/",
					handlerFunc: func(w http.ResponseWriter, r *http.Request) {
						_ = r.Body.Close()
						if diff := cmp.Diff(http.MethodGet, r.Method); diff != "" {
							t.Errorf("r: -want, +got:\n%s", diff)
						}
						w.Header().Set("Content-Type", "application/json")
						w.Header().Set("ETag", eTag)
						cr := gmInstance(func(p *iamagv2.GroupMembersList) {
							p.Members[0].IamID = &memberIamID3
						})
						_ = json.NewEncoder(w).Encode(cr)
					},
				},
			},
			kube: &test.MockClient{
				MockUpdate: test.NewMockUpdateFn(nil),
			},
			args: args{
				mg: gm(
					gmWithExternalNameAnnotation(agID),
					gmWithSpec(*gmParams()),
				),
			},
			want: want{
				mg: gm(gmWithSpec(*gmParams()),
					gmWithEtagAnnotation(eTag),
					gmWithConditions(cpv1alpha1.Available()),
					gmWithStatus(*gmObservation(func(cro *v1alpha1.GroupMembershipObservation) {
						cro.State = ibmcgm.StateActive
						cro.Members[0].IamID = memberIamID3
					}))),
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
			e := gmExternal{
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

func TestGroupMembershipCreate(t *testing.T) {
	type args struct {
		mg resource.Managed
	}
	type want struct {
		mg  resource.Managed
		cre managed.ExternalCreation
		err error
	}
	cases := map[string]struct {
		handlers []handler
		kube     client.Client
		args     args
		want     want
	}{
		"Successful": {
			handlers: []handler{
				{
					path: "/",
					handlerFunc: func(w http.ResponseWriter, r *http.Request) {
						if diff := cmp.Diff(http.MethodPut, r.Method); diff != "" {
							t.Errorf("r: -want, +got:\n%s", diff)
						}
						w.Header().Set("Content-Type", "application/json")
						w.WriteHeader(http.StatusCreated)
						_ = r.Body.Close()
						cr := gmInstance()
						_ = json.NewEncoder(w).Encode(cr)
					},
				},
			},
			args: args{
				mg: gm(gmWithSpec(*gmParams())),
			},
			want: want{
				mg: gm(gmWithSpec(*gmParams()),
					gmWithConditions(cpv1alpha1.Creating()),
					gmWithExternalNameAnnotation(agID)),
				cre: managed.ExternalCreation{ExternalNameAssigned: true},
				err: nil,
			},
		},
		"BadRequest": {
			handlers: []handler{
				{
					path: "/",
					handlerFunc: func(w http.ResponseWriter, r *http.Request) {
						if diff := cmp.Diff(http.MethodPut, r.Method); diff != "" {
							t.Errorf("r: -want, +got:\n%s", diff)
						}
						w.Header().Set("Content-Type", "application/json")
						w.WriteHeader(http.StatusBadRequest)
						_ = r.Body.Close()
						cr := gmInstance()
						_ = json.NewEncoder(w).Encode(cr)
					},
				},
			},
			args: args{
				mg: gm(gmWithSpec(*gmParams())),
			},
			want: want{
				mg: gm(gmWithSpec(*gmParams()),
					gmWithConditions(cpv1alpha1.Creating())),
				cre: managed.ExternalCreation{ExternalNameAssigned: false},
				err: errors.Wrap(errors.New(http.StatusText(http.StatusBadRequest)), errCreateGroupMembership),
			},
		},
		"Conflict": {
			handlers: []handler{
				{
					path: "/",
					handlerFunc: func(w http.ResponseWriter, r *http.Request) {
						if diff := cmp.Diff(http.MethodPut, r.Method); diff != "" {
							t.Errorf("r: -want, +got:\n%s", diff)
						}
						w.Header().Set("Content-Type", "application/json")
						w.WriteHeader(http.StatusConflict)
						_ = r.Body.Close()
						cr := gmInstance()
						_ = json.NewEncoder(w).Encode(cr)
					},
				},
			},
			args: args{
				mg: gm(gmWithSpec(*gmParams())),
			},
			want: want{
				mg: gm(gmWithSpec(*gmParams()),
					gmWithConditions(cpv1alpha1.Creating())),
				cre: managed.ExternalCreation{ExternalNameAssigned: false},
				err: errors.Wrap(errors.New(http.StatusText(http.StatusConflict)), errCreateGroupMembership),
			},
		},
		"Forbidden": {
			handlers: []handler{
				{
					path: "/",
					handlerFunc: func(w http.ResponseWriter, r *http.Request) {
						if diff := cmp.Diff(http.MethodPut, r.Method); diff != "" {
							t.Errorf("r: -want, +got:\n%s", diff)
						}
						w.Header().Set("Content-Type", "application/json")
						w.WriteHeader(http.StatusForbidden)
						_ = r.Body.Close()
						cr := gmInstance()
						_ = json.NewEncoder(w).Encode(cr)
					},
				},
			},
			args: args{
				mg: gm(gmWithSpec(*gmParams())),
			},
			want: want{
				mg: gm(gmWithSpec(*gmParams()),
					gmWithConditions(cpv1alpha1.Creating())),
				cre: managed.ExternalCreation{ExternalNameAssigned: false},
				err: errors.Wrap(errors.New(http.StatusText(http.StatusForbidden)), errCreateGroupMembership),
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
			e := gmExternal{
				kube:   tc.kube,
				client: mClient,
				logger: logging.NewNopLogger(),
			}
			cre, err := e.Create(context.Background(), tc.args.mg)
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
			if diff := cmp.Diff(tc.want.mg, tc.args.mg); diff != "" {
				t.Errorf("Create(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestGroupMembershipDelete(t *testing.T) {
	type args struct {
		mg resource.Managed
	}
	type want struct {
		mg  resource.Managed
		err error
	}
	cases := map[string]struct {
		handlers []handler
		kube     client.Client
		args     args
		want     want
	}{
		"Successful": {
			handlers: []handler{
				{
					path: "/",
					handlerFunc: func(w http.ResponseWriter, r *http.Request) {
						if diff := cmp.Diff(http.MethodPost, r.Method); diff != "" {
							t.Errorf("r: -want, +got:\n%s", diff)
						}
						w.Header().Set("Content-Type", "application/json")
						_ = r.Body.Close()
						gm := iamagv2.DeleteGroupBulkMembersResponse{}
						_ = json.NewEncoder(w).Encode(gm)
					},
				},
			},
			args: args{
				mg: gm(gmWithStatus(*gmObservation()), gmWithExternalNameAnnotation(accessGroupID)),
			},
			want: want{
				mg:  gm(gmWithStatus(*gmObservation()), gmWithConditions(cpv1alpha1.Deleting())),
				err: nil,
			},
		},
		"BadRequest": {
			handlers: []handler{
				{
					path: "/",
					handlerFunc: func(w http.ResponseWriter, r *http.Request) {
						if diff := cmp.Diff(http.MethodPost, r.Method); diff != "" {
							t.Errorf("r: -want, +got:\n%s", diff)
						}
						w.Header().Set("Content-Type", "application/json")
						w.WriteHeader(http.StatusBadRequest)
						_ = r.Body.Close()
					},
				},
			},
			args: args{
				mg: gm(gmWithStatus(*gmObservation()), gmWithExternalNameAnnotation(accessGroupID)),
			},
			want: want{
				mg:  gm(gmWithStatus(*gmObservation()), gmWithConditions(cpv1alpha1.Deleting())),
				err: errors.Wrap(errors.New(http.StatusText(http.StatusBadRequest)), errDeleteGroupMembership),
			},
		},
		"InvalidToken": {
			handlers: []handler{
				{
					path: "/",
					handlerFunc: func(w http.ResponseWriter, r *http.Request) {
						if diff := cmp.Diff(http.MethodPost, r.Method); diff != "" {
							t.Errorf("r: -want, +got:\n%s", diff)
						}
						w.Header().Set("Content-Type", "application/json")
						w.WriteHeader(http.StatusUnauthorized)
						_ = r.Body.Close()
					},
				},
			},
			args: args{
				mg: gm(gmWithStatus(*gmObservation()), gmWithExternalNameAnnotation(accessGroupID)),
			},
			want: want{
				mg:  gm(gmWithStatus(*gmObservation()), gmWithConditions(cpv1alpha1.Deleting())),
				err: errors.Wrap(errors.New(http.StatusText(http.StatusUnauthorized)), errDeleteGroupMembership),
			},
		},
		"Forbidden": {
			handlers: []handler{
				{
					path: "/",
					handlerFunc: func(w http.ResponseWriter, r *http.Request) {
						if diff := cmp.Diff(http.MethodPost, r.Method); diff != "" {
							t.Errorf("r: -want, +got:\n%s", diff)
						}
						w.Header().Set("Content-Type", "application/json")
						w.WriteHeader(http.StatusForbidden)
						_ = r.Body.Close()
					},
				},
			},
			args: args{
				mg: gm(gmWithStatus(*gmObservation()), gmWithExternalNameAnnotation(accessGroupID)),
			},
			want: want{
				mg:  gm(gmWithStatus(*gmObservation()), gmWithConditions(cpv1alpha1.Deleting())),
				err: errors.Wrap(errors.New(http.StatusText(http.StatusForbidden)), errDeleteGroupMembership),
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
			e := gmExternal{
				kube:   tc.kube,
				client: mClient,
				logger: logging.NewNopLogger(),
			}
			err := e.Delete(context.Background(), tc.args.mg)
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
			if diff := cmp.Diff(tc.want.mg, tc.args.mg); diff != "" {
				t.Errorf("Delete(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestGroupMembershipUpdate(t *testing.T) {
	type args struct {
		mg resource.Managed
	}
	type want struct {
		mg  resource.Managed
		upd managed.ExternalUpdate
		err error
	}
	cases := map[string]struct {
		handlers []handler
		kube     client.Client
		args     args
		want     want
	}{
		"Successful": {
			handlers: []handler{
				{
					path:        "/",
					handlerFunc: iamMembersHandler,
				},
			},
			args: args{
				mg: gm(gmWithSpec(*gmParams(
					func(gmp *v1alpha1.GroupMembershipParameters) {
						gmp.Members = append(gmp.Members, v1alpha1.AddGroupMembersRequestMembersItem{
							IamID: memberIamID3,
							Type:  ibmcgm.MemberTypeUser,
						})
					},
				)), gmWithStatus(*gmObservation()), gmWithEtagAnnotation(eTag)),
			},
			want: want{
				mg: gm(gmWithSpec(*gmParams(func(gmp *v1alpha1.GroupMembershipParameters) {
					gmp.Members = append(gmp.Members, v1alpha1.AddGroupMembersRequestMembersItem{
						IamID: memberIamID3,
						Type:  ibmcgm.MemberTypeUser,
					})
				})), gmWithStatus(*gmObservation()), gmWithEtagAnnotation(eTag)),
				upd: managed.ExternalUpdate{},
				err: nil,
			},
		},
		"BadRequest": {
			handlers: []handler{
				{
					path: "/",
					handlerFunc: func(w http.ResponseWriter, r *http.Request) {
						if diff := cmp.Diff(http.MethodPut, r.Method); diff != "" {
							t.Errorf("r: -want, +got:\n%s", diff)
						}
						w.Header().Set("Content-Type", "application/json")
						w.WriteHeader(http.StatusBadRequest)
						_ = r.Body.Close()
					},
				},
			},
			args: args{
				mg: gm(gmWithSpec(*gmParams(
					func(gmp *v1alpha1.GroupMembershipParameters) {
						gmp.Members = append(gmp.Members, v1alpha1.AddGroupMembersRequestMembersItem{
							IamID: memberIamID3,
							Type:  ibmcgm.MemberTypeUser,
						})
					},
				)), gmWithStatus(*gmObservation()), gmWithEtagAnnotation(eTag)),
			},
			want: want{
				mg:  gm(gmWithSpec(*gmParams()), gmWithStatus(*gmObservation())),
				err: errors.Wrap(errors.New(http.StatusText(http.StatusBadRequest)), errUpdGroupMembership),
			},
		},
		"NotFound": {
			handlers: []handler{
				{
					path: "/",
					handlerFunc: func(w http.ResponseWriter, r *http.Request) {
						if diff := cmp.Diff(http.MethodPut, r.Method); diff != "" {
							t.Errorf("r: -want, +got:\n%s", diff)
						}
						w.Header().Set("Content-Type", "application/json")
						w.WriteHeader(http.StatusNotFound)
						_ = r.Body.Close()
					},
				},
			},
			args: args{
				mg: gm(gmWithSpec(*gmParams(
					func(gmp *v1alpha1.GroupMembershipParameters) {
						gmp.Members = append(gmp.Members, v1alpha1.AddGroupMembersRequestMembersItem{
							IamID: memberIamID3,
							Type:  ibmcgm.MemberTypeUser,
						})
					},
				)), gmWithStatus(*gmObservation()), gmWithEtagAnnotation(eTag)),
			},
			want: want{
				mg:  gm(gmWithSpec(*gmParams()), gmWithStatus(*gmObservation())),
				err: errors.Wrap(errors.New(http.StatusText(http.StatusNotFound)), errUpdGroupMembership),
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

			membersCache = map[string]iamagv2.ListGroupMembersResponseMember{
				memberIamID1: {
					IamID: &memberIamID1,
					Type:  reference.ToPtrValue(ibmcgm.MemberTypeUser),
				},
				memberIamID2: {
					IamID: &memberIamID2,
					Type:  reference.ToPtrValue(ibmcgm.MemberTypeUser),
				},
			}

			opts := ibmc.ClientOptions{URL: server.URL, Authenticator: &core.BearerTokenAuthenticator{
				BearerToken: bearerTok,
			}}
			mClient, _ := ibmc.NewClient(opts)
			e := gmExternal{
				kube:   tc.kube,
				client: mClient,
				logger: logging.NewNopLogger(),
			}
			upd, err := e.Update(context.Background(), tc.args.mg)
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
				if diff := cmp.Diff(tc.want.mg, tc.args.mg); diff != "" {
					t.Errorf("Update(...): -want, +got:\n%s", diff)
				}
				if diff := cmp.Diff(tc.want.upd, upd); diff != "" {
					t.Errorf("Update(...): -want, +got:\n%s", diff)
				}
			}
		})
	}
}
