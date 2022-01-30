package groupmembership

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/reference"
	"github.com/go-openapi/strfmt"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	iamagv2 "github.com/IBM/platform-services-go-sdk/iamaccessgroupsv2"

	"github.com/crossplane-contrib/provider-ibm-cloud/apis/iamaccessgroupsv2/v1alpha1"
	ibmc "github.com/crossplane-contrib/provider-ibm-cloud/pkg/clients"
)

var (
	accessGroupID = "12345678-abcd-1a2b-a1b2-1234567890ab"
	createdAt, _  = strfmt.ParseDateTime("2020-10-31T02:33:06Z")
	transactionID = "12345-abcd-ef000-abac"
	createdByID   = "IBM-User-0001"
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

func params(m ...func(*v1alpha1.GroupMembershipParameters)) *v1alpha1.GroupMembershipParameters {
	p := &v1alpha1.GroupMembershipParameters{
		AccessGroupID: &accessGroupID,
		TransactionID: &transactionID,
		Members: []v1alpha1.AddGroupMembersRequestMembersItem{
			{
				IamID: memberIamID1,
				Type:  MemberTypeUser,
			},
			{
				IamID: memberIamID2,
				Type:  MemberTypeUser,
			},
		},
	}

	for _, f := range m {
		f(p)
	}
	return p
}

func observation(m ...func(*v1alpha1.GroupMembershipObservation)) *v1alpha1.GroupMembershipObservation {
	o := &v1alpha1.GroupMembershipObservation{
		Members: []v1alpha1.ListGroupMembersResponseMember{
			{
				IamID:       memberIamID1,
				Type:        MemberTypeUser,
				Name:        memberName1,
				Email:       memberEmail1,
				Description: memberDescr1,
				Href:        memberHRef1,
				CreatedAt:   GenerateMetaV1Time(&createdAt),
				CreatedByID: createdByID,
			},
			{
				IamID:       memberIamID2,
				Type:        MemberTypeUser,
				Name:        memberName2,
				Email:       memberEmail2,
				Description: memberDescr2,
				Href:        memberHRef2,
				CreatedAt:   GenerateMetaV1Time(&createdAt),
				CreatedByID: createdByID,
			},
		},
	}

	for _, f := range m {
		f(o)
	}
	return o
}

func instanceCreateOpts(m ...func(*iamagv2.AddMembersToAccessGroupOptions)) *iamagv2.AddMembersToAccessGroupOptions {
	i := &iamagv2.AddMembersToAccessGroupOptions{
		AccessGroupID: &accessGroupID,
		Members: []iamagv2.AddGroupMembersRequestMembersItem{
			{
				IamID: &memberIamID1,
				Type:  reference.ToPtrValue(MemberTypeUser),
			},
			{
				IamID: &memberIamID2,
				Type:  reference.ToPtrValue(MemberTypeUser),
			},
		},
		TransactionID: &transactionID,
	}
	for _, f := range m {
		f(i)
	}
	return i
}

func instanceList(m ...func(*iamagv2.GroupMembersList)) *iamagv2.GroupMembersList {
	i := &iamagv2.GroupMembersList{
		TotalCount: &totalCount,
		Members: []iamagv2.ListGroupMembersResponseMember{
			{
				IamID:       &memberIamID1,
				Type:        reference.ToPtrValue(MemberTypeUser),
				Name:        &memberName1,
				Email:       &memberEmail1,
				Description: &memberDescr1,
				Href:        &memberHRef1,
				CreatedAt:   &createdAt,
				CreatedByID: &createdByID,
			},
			{
				IamID:       &memberIamID2,
				Type:        reference.ToPtrValue(MemberTypeUser),
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

func TestGenerateCreateGroupMembershipOptions(t *testing.T) {
	type args struct {
		params v1alpha1.GroupMembershipParameters
	}
	type want struct {
		instance *iamagv2.AddMembersToAccessGroupOptions
	}
	cases := map[string]struct {
		args args
		want want
	}{
		"FullConversion": {
			args: args{params: *params()},
			want: want{instance: instanceCreateOpts()},
		},
		"MissingFields": {
			args: args{
				params: *params(func(p *v1alpha1.GroupMembershipParameters) {
					p.TransactionID = nil
				})},
			want: want{instance: instanceCreateOpts(func(p *iamagv2.AddMembersToAccessGroupOptions) {
				p.TransactionID = nil
			})},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {

			r := &iamagv2.AddMembersToAccessGroupOptions{}
			GenerateCreateGroupMembershipOptions(tc.args.params, r)
			if diff := cmp.Diff(tc.want.instance, r); diff != "" {
				t.Errorf("GenerateCreateGroupMembershipOptions(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestLateInitializeSpecs(t *testing.T) {
	type args struct {
		instance *iamagv2.GroupMembersList
		params   *v1alpha1.GroupMembershipParameters
	}
	type want struct {
		params *v1alpha1.GroupMembershipParameters
	}
	cases := map[string]struct {
		args args
		want want
	}{
		"AllFilledAlready": {
			args: args{
				params:   params(),
				instance: instanceList(),
			},
			want: want{
				params: params()},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			LateInitializeSpec(tc.args.params, tc.args.instance)
			if diff := cmp.Diff(tc.want.params, tc.args.params); diff != "" {
				t.Errorf("LateInitializeSpec(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestGenerateObservation(t *testing.T) {
	type args struct {
		instanceList *iamagv2.GroupMembersList
	}
	type want struct {
		obs v1alpha1.GroupMembershipObservation
	}
	cases := map[string]struct {
		args args
		want want
	}{
		"FullConversion": {
			args: args{
				instanceList: instanceList(func(p *iamagv2.GroupMembersList) {
				}),
			},
			want: want{*observation()},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			o, err := GenerateObservation(tc.args.instanceList)
			if diff := cmp.Diff(nil, err); diff != "" {
				t.Errorf("GenerateObservation(...): want error != got error:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.obs, o); diff != "" {
				t.Errorf("GenerateObservation(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestIsUpToDate(t *testing.T) {
	type args struct {
		params       *v1alpha1.GroupMembershipParameters
		instanceList *iamagv2.GroupMembersList
	}
	type want struct {
		upToDate bool
		isErr    bool
	}
	cases := map[string]struct {
		args args
		want want
	}{
		"IsUpToDate": {
			args: args{
				params:       params(),
				instanceList: instanceList(),
			},
			want: want{upToDate: true, isErr: false},
		},
		"NeedsUpdate": {
			args: args{
				params: params(func(crp *v1alpha1.GroupMembershipParameters) {
					crp.Members = []v1alpha1.AddGroupMembersRequestMembersItem{
						{
							IamID: memberIamID1,
							Type:  MemberTypeUser,
						},
					}
				}),
				instanceList: instanceList(func(i *iamagv2.GroupMembersList) {
				}),
			},
			want: want{upToDate: false, isErr: false},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {

			r, err := IsUpToDate(tc.args.params, tc.args.instanceList, logging.NewNopLogger())
			if err != nil && !tc.want.isErr {
				t.Error("IsUpToDate(...) unexpected error")
			}
			if diff := cmp.Diff(tc.want.upToDate, r); diff != "" {
				t.Errorf("IsUpToDate(...): -want, +got:\n%s", diff)
			}
		})
	}
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
				Type:  reference.ToPtrValue(MemberTypeUser),
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

func TestUpdateAccessGroupMembers(t *testing.T) {
	type args struct {
		gm v1alpha1.GroupMembership
	}
	type want struct {
		members *iamagv2.GroupMembersList
		err     error
	}
	cases := map[string]struct {
		args args
		want want
	}{
		"AddMembers": {
			args: args{
				v1alpha1.GroupMembership{
					Spec: v1alpha1.GroupMembershipSpec{
						ForProvider: *params(func(gmp *v1alpha1.GroupMembershipParameters) {
							gmp.Members = append(gmp.Members, v1alpha1.AddGroupMembersRequestMembersItem{
								IamID: memberIamID3,
								Type:  MemberTypeUser,
							})
						}),
					},
					Status: v1alpha1.GroupMembershipStatus{
						AtProvider: *observation(),
					},
				},
			},
			want: want{members: &iamagv2.GroupMembersList{
				TotalCount: ibmc.Int64Ptr(int64(3)),
				Members: []iamagv2.ListGroupMembersResponseMember{
					{
						IamID: &memberIamID1,
						Type:  reference.ToPtrValue(MemberTypeUser),
					},
					{
						IamID: &memberIamID2,
						Type:  reference.ToPtrValue(MemberTypeUser),
					},
					{
						IamID: &memberIamID3,
						Type:  reference.ToPtrValue(MemberTypeUser),
					},
				},
			}, err: nil},
		},
		"RemoveMembers": {
			args: args{
				v1alpha1.GroupMembership{
					Spec: v1alpha1.GroupMembershipSpec{
						ForProvider: *params(func(gmp *v1alpha1.GroupMembershipParameters) {
							gmp.Members = []v1alpha1.AddGroupMembersRequestMembersItem{
								{
									IamID: memberIamID1,
									Type:  MemberTypeUser,
								},
							}
						}),
					},
					Status: v1alpha1.GroupMembershipStatus{
						AtProvider: *observation(),
					},
				},
			},
			want: want{members: &iamagv2.GroupMembersList{
				TotalCount: ibmc.Int64Ptr(int64(1)),
				Members: []iamagv2.ListGroupMembersResponseMember{
					{
						IamID: &memberIamID1,
						Type:  reference.ToPtrValue(MemberTypeUser),
					},
				},
			}, err: nil},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			mux := http.NewServeMux()
			mux.HandleFunc("/groups/", iamMembersHandler)
			server := httptest.NewServer(mux)
			defer server.Close()

			membersCache = map[string]iamagv2.ListGroupMembersResponseMember{
				memberIamID1: {
					IamID: &memberIamID1,
					Type:  reference.ToPtrValue(MemberTypeUser),
				},
				memberIamID2: {
					IamID: &memberIamID2,
					Type:  reference.ToPtrValue(MemberTypeUser),
				},
			}

			mClient, _ := ibmc.GetTestClient(server.URL)
			err := UpdateAccessGroupMembers(mClient, tc.args.gm)
			if tc.want.err != nil && tc.want.err.Error() != err.Error() {
				t.Errorf("UpdateAccessGroupMembers(...): want: %s\ngot: %s\n", tc.want.err, err)
			}
			listOpts := &iamagv2.ListAccessGroupMembersOptions{
				AccessGroupID: &accessGroupID,
			}
			members, _, err := mClient.IamAccessGroupsV2().ListAccessGroupMembers(listOpts)
			if tc.want.err != nil && tc.want.err.Error() != err.Error() {
				t.Errorf("UpdateAccessGroupMembers(...): want: %s\ngot: %s\n", tc.want.err, err)
			}
			if diff := cmp.Diff(tc.want.members.Members, members.Members, cmpopts.SortSlices(func(x, y interface{}) bool {
				x1 := x.(iamagv2.ListGroupMembersResponseMember).IamID
				y1 := y.(iamagv2.ListGroupMembersResponseMember).IamID
				return fmt.Sprint("%# v", *x1) < fmt.Sprint("%# v", *y1)
			})); diff != "" {
				t.Errorf("UpdateAccessGroupMembers(...): -want, +got:\n%s", diff)
			}
		})
	}
}
