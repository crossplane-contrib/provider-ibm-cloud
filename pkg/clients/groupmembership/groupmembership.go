package groupmembership

import (
	"encoding/json"
	"sort"
	"time"

	"github.com/go-openapi/strfmt"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/pkg/errors"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/reference"

	corev4 "github.com/IBM/go-sdk-core/v4/core"
	iamagv2 "github.com/IBM/platform-services-go-sdk/iamaccessgroupsv2"

	"github.com/crossplane-contrib/provider-ibm-cloud/apis/iamaccessgroupsv2/v1alpha1"
	ibmc "github.com/crossplane-contrib/provider-ibm-cloud/pkg/clients"
)

const (
	// StateActive represents an access group in a running, available, and ready state
	StateActive = "active"
	// StateInactive represents an inactive access group
	StateInactive = "inactive"
	// MemberTypeUser represents a user member
	MemberTypeUser = "user"
	// MemberTypeService represents a service member
	MemberTypeService = "service"
)

// LateInitializeSpec fills optional and unassigned fields with the values in *iamagv2.GroupMembership object.
func LateInitializeSpec(spec *v1alpha1.GroupMembershipParameters, in *iamagv2.GroupMembersList) error { // nolint:gocyclo
	return nil
}

// GenerateCreateGroupMembershipOptions produces GroupMembershipOptions object from GroupMembershipParameters object.
func GenerateCreateGroupMembershipOptions(in v1alpha1.GroupMembershipParameters, o *iamagv2.AddMembersToAccessGroupOptions) error {
	o.AccessGroupID = in.AccessGroupID
	o.TransactionID = in.TransactionID
	o.Members = GenerateSDKAddGroupMembersRequestMembersItems(in.Members)
	return nil
}

// GenerateSDKAddGroupMembersRequestMembersItems -
func GenerateSDKAddGroupMembersRequestMembersItems(in []v1alpha1.AddGroupMembersRequestMembersItem) []iamagv2.AddGroupMembersRequestMembersItem {
	o := []iamagv2.AddGroupMembersRequestMembersItem{}
	for _, m := range in {
		item := iamagv2.AddGroupMembersRequestMembersItem{
			IamID: reference.ToPtrValue(m.IamID),
			Type:  reference.ToPtrValue(m.Type),
		}
		o = append(o, item)
	}
	return o
}

// GenerateSDKRemoveroupMembersRequestMembersItems -
func GenerateSDKRemoveroupMembersRequestMembersItems(in []v1alpha1.AddGroupMembersRequestMembersItem) []string {
	o := []string{}
	for _, m := range in {
		o = append(o, m.IamID)
	}
	return o
}

// GenerateObservation produces GroupMembershipObservation object from *iamagv2.Group object.
func GenerateObservation(in *iamagv2.GroupMembersList) (v1alpha1.GroupMembershipObservation, error) {
	o := v1alpha1.GroupMembershipObservation{
		Members: GenerateCRListGroupMembersResponseMembers(in),
	}
	return o, nil
}

// GenerateCRListGroupMembersResponseMembers -
func GenerateCRListGroupMembersResponseMembers(in *iamagv2.GroupMembersList) []v1alpha1.ListGroupMembersResponseMember {
	o := []v1alpha1.ListGroupMembersResponseMember{}
	if in == nil {
		return o
	}
	for _, m := range in.Members {
		item := v1alpha1.ListGroupMembersResponseMember{
			IamID:       reference.FromPtrValue(m.IamID),
			Type:        reference.FromPtrValue(m.Type),
			Name:        reference.FromPtrValue(m.Name),
			Email:       reference.FromPtrValue(m.Email),
			Description: reference.FromPtrValue(m.Description),
			Href:        reference.FromPtrValue(m.Href),
			CreatedAt:   GenerateMetaV1Time(m.CreatedAt),
			CreatedByID: reference.FromPtrValue(m.CreatedByID),
		}
		o = append(o, item)
	}
	return o
}

// GenerateMetaV1Time converts strfmt.DateTime to metav1.Time
// TODO - extract this to parent `clients` package
func GenerateMetaV1Time(t *strfmt.DateTime) *metav1.Time {
	if t == nil {
		return nil
	}
	tx := metav1.NewTime(time.Time(*t))
	return &tx
}

// IsUpToDate checks whether current state is up-to-date compared to the given
// set of parameters.
func IsUpToDate(in *v1alpha1.GroupMembershipParameters, observed *iamagv2.GroupMembersList, l logging.Logger) (bool, error) {
	desired := in.DeepCopy()
	actual, err := GenerateGroupMembershipParameters(observed)
	if err != nil {
		return false, err
	}
	sort.Slice(desired.Members, func(i, j int) bool {
		return desired.Members[i].IamID < desired.Members[j].IamID
	})
	sort.Slice(actual.Members, func(i, j int) bool {
		return actual.Members[i].IamID < actual.Members[j].IamID
	})

	l.Info(cmp.Diff(desired, actual, cmpopts.EquateEmpty(),
		cmpopts.IgnoreFields(v1alpha1.GroupMembershipParameters{}, "AccessGroupID"),
		cmpopts.IgnoreFields(v1alpha1.GroupMembershipParameters{}, "TransactionID"),
		cmpopts.IgnoreTypes(&runtimev1alpha1.Reference{}, &runtimev1alpha1.Selector{}, []runtimev1alpha1.Reference{})))

	return cmp.Equal(desired, actual, cmpopts.EquateEmpty(),
		cmpopts.IgnoreFields(v1alpha1.GroupMembershipParameters{}, "AccessGroupID"),
		cmpopts.IgnoreFields(v1alpha1.GroupMembershipParameters{}, "TransactionID"),
		cmpopts.IgnoreTypes(&runtimev1alpha1.Reference{}, &runtimev1alpha1.Selector{}, []runtimev1alpha1.Reference{})), nil
}

// GenerateGroupMembershipParameters generates service instance parameters from resource instance
func GenerateGroupMembershipParameters(in *iamagv2.GroupMembersList) (*v1alpha1.GroupMembershipParameters, error) {
	o := &v1alpha1.GroupMembershipParameters{
		Members: GenerateCRAddGroupMembersRequestMembersItem(in),
	}
	return o, nil
}

// GenerateCRAddGroupMembersRequestMembersItem -
func GenerateCRAddGroupMembersRequestMembersItem(in *iamagv2.GroupMembersList) []v1alpha1.AddGroupMembersRequestMembersItem {
	o := []v1alpha1.AddGroupMembersRequestMembersItem{}
	if in == nil {
		return o
	}
	for _, m := range in.Members {
		item := v1alpha1.AddGroupMembersRequestMembersItem{
			IamID: reference.FromPtrValue(m.IamID),
			Type:  reference.FromPtrValue(m.Type),
		}
		o = append(o, item)
	}
	return o
}

// MembersDiff computes the difference between desired members and actual membersfd and returns
// a list of tagcaqs to attach and to detach
func MembersDiff(desired v1alpha1.GroupMembershipParameters, actual v1alpha1.GroupMembershipObservation) ([]iamagv2.AddGroupMembersRequestMembersItem, []string) {
	toAdd := []iamagv2.AddGroupMembersRequestMembersItem{}
	toRemove := []string{}
	dMap := map[string]bool{}
	aMap := map[string]bool{}
	for _, d := range desired.Members {
		dMap[d.IamID] = true
	}
	for _, a := range actual.Members {
		aMap[a.IamID] = true
	}

	for _, d := range desired.Members {
		_, ok := aMap[d.IamID]
		if !ok {
			toAdd = append(toAdd, iamagv2.AddGroupMembersRequestMembersItem{
				IamID: reference.ToPtrValue(d.IamID),
				Type:  reference.ToPtrValue(d.Type),
			})
		}
	}

	for _, a := range actual.Members {
		_, ok := dMap[a.IamID]
		if !ok {
			toRemove = append(toRemove, a.IamID)
		}
	}
	return toAdd, toRemove
}

// UpdateAccessGroupMembers update members access group
func UpdateAccessGroupMembers(client ibmc.ClientSession, groupMembership v1alpha1.GroupMembership) error {
	toAdd, toRemove := MembersDiff(groupMembership.Spec.ForProvider, groupMembership.Status.AtProvider)

	if len(toAdd) > 0 {
		opts := &iamagv2.AddMembersToAccessGroupOptions{
			AccessGroupID: groupMembership.Spec.ForProvider.AccessGroupID,
			Members:       toAdd,
		}
		_, resp, err := client.IamAccessGroupsV2().AddMembersToAccessGroup(opts)
		err = ExtractErrorMessage(resp, err)
		if err != nil {
			return err
		}
	}

	if len(toRemove) > 0 {
		opts := &iamagv2.RemoveMembersFromAccessGroupOptions{
			AccessGroupID: groupMembership.Spec.ForProvider.AccessGroupID,
			TransactionID: groupMembership.Spec.ForProvider.TransactionID,
			Members:       toRemove,
		}
		_, resp, err := client.IamAccessGroupsV2().RemoveMembersFromAccessGroup(opts)
		err = ExtractErrorMessage(resp, err)
		if err != nil {
			return err
		}
	}

	return nil
}

// ExtractErrorMessage extracts the content of an error message from the detailed response (if any)
// and appends it to the error returned by the SDK
func ExtractErrorMessage(resp *corev4.DetailedResponse, err error) error { // nolint:gocyclo
	if resp == nil || resp != nil && resp.Result == nil {
		return err
	}
	rj, e := json.Marshal(resp.Result)
	if e != nil {
		return errors.Wrap(err, e.Error())
	}
	m := map[string]interface{}{}
	e = json.Unmarshal(rj, &m)
	if e != nil {
		return errors.Wrap(err, e.Error())
	}
	if o, ok := m["members"]; ok {
		if members, ok := o.([]interface{}); ok {
			for _, member := range members {
				if memberMap, ok := member.(map[string]interface{}); ok {
					if errs, ok := memberMap["errors"]; ok {
						jErr, e := json.Marshal(errs)
						if e != nil {
							return errors.Wrap(err, e.Error())
						}
						if err == nil {
							return errors.New(string(jErr))
						}
						return errors.Wrap(err, string(jErr))
					}
				}
			}
		}
	}
	return err
}
