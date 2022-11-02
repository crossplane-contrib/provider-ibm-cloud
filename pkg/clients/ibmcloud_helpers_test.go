package clients

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/pkg/errors"
	"k8s.io/klog/v2"

	"github.com/crossplane/crossplane-runtime/pkg/reference"

	gcat "github.com/IBM/platform-services-go-sdk/globalcatalogv1"
	gtagv1 "github.com/IBM/platform-services-go-sdk/globaltaggingv1"
	rmgrv2 "github.com/IBM/platform-services-go-sdk/resourcemanagerv2"
)

var (
	resourceGroupID    = "mock-resource-group-id"
	resourcePlanID     = "744bfc56-d12c-4866-88d5-dac9139e0e5d"
	resourceGroupName  = "default"
	resourcePlanName   = "standard"
	serviceName        = "cloud-object-storage"
	invalidServiceName = "invalidServiceName"
	invalidPlanName    = "invalidPlanName"
	invalidPlanID      = "invalidPlanID"
	invalidRGName      = "invalidRGName"
	invalidRGID        = "invalidRGID"
	testCrn            = "crn:v1:bluemix:public:cloud-object-storage:global:a/0b5a00334eaf9eb9339d2ab48f20d7f5:78d88b2b-bbbb-aaaa-8888-5c26e8b6a555::"
	invalidCrn         = "crn:v1:bluemix:public:invalidcrn"
)

var tagsCache = map[string]bool{"dev": true}

// handler to mock client SDK call to global tags API
var tagsHandler = func(w http.ResponseWriter, r *http.Request) {
	var tags gtagv1.TagList
	if r.Method == http.MethodPost {
		body, _ := ioutil.ReadAll(r.Body)
		m := map[string]interface{}{}
		json.Unmarshal(body, &m)
		tags := m["tag_names"].([]interface{})

		if strings.Contains(r.URL.String(), "attach") {
			for _, t := range tags {
				tagsCache[t.(string)] = true
			}
		}
		if strings.Contains(r.URL.String(), "detach") {
			for _, t := range tags {
				delete(tagsCache, t.(string))
			}
		}
		return
	}
	_ = r.Body.Close()
	crn := r.URL.Query()["attached_to"]
	w.Header().Set("Content-Type", "application/json")
	if len(crn) > 0 && crn[0] == invalidCrn {
		tags = gtagv1.TagList{}
	} else {
		tags = gtagv1.TagList{
			Items: map2tags(tagsCache),
		}
	}
	err := json.NewEncoder(w).Encode(tags)
	if err != nil {
		klog.Errorf("%s", err)
	}
}

func map2tags(m map[string]bool) []gtagv1.Tag {
	s := []gtagv1.Tag{}
	for key := range m {
		s = append(s, gtagv1.Tag{Name: reference.ToPtrValue(dupString(key))})
	}
	return s
}

var dupString = func(s string) string {
	return string([]byte(s))
}

// handler to mock client SDK call to resource manager API
var rgHandler = func(w http.ResponseWriter, r *http.Request) {
	_ = r.Body.Close()
	w.Header().Set("Content-Type", "application/json")
	rgl := rmgrv2.ResourceGroupList{
		Resources: []rmgrv2.ResourceGroup{
			{
				ID:   reference.ToPtrValue(resourceGroupID),
				Name: reference.ToPtrValue(resourceGroupName),
			},
		},
	}
	err := json.NewEncoder(w).Encode(rgl)
	if err != nil {
		klog.Errorf("%s", err)
	}
}

// handler to mock client SDK call to global catalog API for plans
var pcatHandler = func(w http.ResponseWriter, r *http.Request) {
	_ = r.Body.Close()

	w.Header().Set("Content-Type", "application/json")
	planEntry := gcat.EntrySearchResult{
		Resources: []gcat.CatalogEntry{
			{
				ID:   reference.ToPtrValue(resourcePlanID),
				Name: reference.ToPtrValue(resourcePlanName),
			},
		},
	}
	err := json.NewEncoder(w).Encode(planEntry)
	if err != nil {
		klog.Errorf("%s", err)
	}
}

// handler to mock client SDK call to global catalog API for services
var svcatHandler = func(w http.ResponseWriter, r *http.Request) {
	_ = r.Body.Close()
	svc := r.URL.Query()["q"]
	var catEntry gcat.EntrySearchResult

	w.Header().Set("Content-Type", "application/json")
	if svc[0] == invalidServiceName {
		catEntry = gcat.EntrySearchResult{}
	} else {
		catEntry = gcat.EntrySearchResult{
			Resources: []gcat.CatalogEntry{
				{
					Metadata: &gcat.CatalogEntryMetadata{
						UI: &gcat.UIMetaData{
							PrimaryOfferingID: reference.ToPtrValue(svc[0]),
						},
					},
				},
			},
		}
	}
	err := json.NewEncoder(w).Encode(catEntry)
	if err != nil {
		klog.Errorf("%s", err)
	}
}

func TestGetResourcePlanID(t *testing.T) {
	type args struct {
		serviceName string
		planName    string
	}
	type want struct {
		planID *string
		err    error
	}
	cases := map[string]struct {
		args args
		want want
	}{
		"Found": {
			args: args{
				serviceName: serviceName,
				planName:    resourcePlanName,
			},
			want: want{planID: &resourcePlanID, err: nil},
		},
		"ServiceNotFound": {
			args: args{
				serviceName: invalidServiceName,
				planName:    resourcePlanName,
			},
			want: want{planID: nil, err: errors.Wrap(errors.New(errNotFound), errListPlanCatEntries)},
		},
		"PlanIDNotFound": {
			args: args{
				serviceName: serviceName,
				planName:    invalidPlanName,
			},
			want: want{planID: nil, err: errors.New(errPlanIDNotFound)},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			mux := http.NewServeMux()
			mux.HandleFunc("/", svcatHandler)
			mux.HandleFunc("/"+serviceName+"/", pcatHandler)
			server := httptest.NewServer(mux)
			defer server.Close()

			mClient, _ := GetTestClient(server.URL)

			planID, err := GetResourcePlanID(mClient, tc.args.serviceName, tc.args.planName)
			if tc.want.err != nil && tc.want.err.Error() != err.Error() {
				t.Errorf("GetResourcePlanID(...): want: %s\ngot: %s\n", tc.want.err, err)
			}
			if diff := cmp.Diff(tc.want.planID, planID); diff != "" {
				t.Errorf("GetResourcePlanID(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestGetResourcePlanName(t *testing.T) {
	type args struct {
		serviceName string
		planID      string
	}
	type want struct {
		planName *string
		err      error
	}
	cases := map[string]struct {
		args args
		want want
	}{
		"Found": {
			args: args{
				serviceName: serviceName,
				planID:      resourcePlanID,
			},
			want: want{planName: &resourcePlanName, err: nil},
		},
		"ServiceNotFound": {
			args: args{
				serviceName: invalidServiceName,
				planID:      resourcePlanID,
			},
			want: want{planName: nil, err: errors.Wrap(errors.New(errNotFound), errListPlanCatEntries)},
		},
		"PlanNameNotFound": {
			args: args{
				serviceName: serviceName,
				planID:      invalidPlanID,
			},
			want: want{planName: nil, err: errors.New(errPlanNameNotFound)},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			mux := http.NewServeMux()
			mux.HandleFunc("/", svcatHandler)
			mux.HandleFunc("/"+serviceName+"/", pcatHandler)
			server := httptest.NewServer(mux)
			defer server.Close()

			mClient, _ := GetTestClient(server.URL)

			planName, err := GetResourcePlanName(mClient, tc.args.serviceName, tc.args.planID)
			if tc.want.err != nil && tc.want.err.Error() != err.Error() {
				t.Errorf("GetResourcePlanName(...): want: %s\ngot: %s\n", tc.want.err, err)
			}
			if diff := cmp.Diff(tc.want.planName, planName); diff != "" {
				t.Errorf("GetResourcePlanName(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestGetResourceGroupID(t *testing.T) {
	type args struct {
		rgName *string
	}
	type want struct {
		rgID *string
		err  error
	}
	cases := map[string]struct {
		args args
		want want
	}{
		"Found": {
			args: args{
				rgName: &resourceGroupName,
			},
			want: want{rgID: &resourceGroupID, err: nil},
		},
		"Default": {
			args: args{
				rgName: nil,
			},
			want: want{rgID: &resourceGroupID, err: nil},
		},
		"NotFound": {
			args: args{
				rgName: &invalidRGName,
			},
			want: want{rgID: nil, err: errors.New(errRGIDNotFound)},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			mux := http.NewServeMux()
			mux.HandleFunc("/resource_groups/", rgHandler)
			server := httptest.NewServer(mux)
			defer server.Close()

			mClient, _ := GetTestClient(server.URL)

			rgID, err := GetResourceGroupID(mClient, tc.args.rgName)
			if tc.want.err != nil && tc.want.err.Error() != err.Error() {
				t.Errorf("GetResourceGroupID(...): want: %s\ngot: %s\n", tc.want.err, err)
			}
			if diff := cmp.Diff(tc.want.rgID, rgID); diff != "" {
				t.Errorf("GetResourceGroupID(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestGetResourceGroupName(t *testing.T) {
	type args struct {
		rgID string
	}
	type want struct {
		rgName string
		err    error
	}
	cases := map[string]struct {
		args args
		want want
	}{
		"Found": {
			args: args{
				rgID: resourceGroupID,
			},
			want: want{rgName: resourceGroupName, err: nil},
		},
		"NotFound": {
			args: args{
				rgID: invalidRGID,
			},
			want: want{rgName: "", err: errors.New(errRGNameNotFound)},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			mux := http.NewServeMux()
			mux.HandleFunc("/resource_groups/", rgHandler)
			server := httptest.NewServer(mux)
			defer server.Close()

			mClient, _ := GetTestClient(server.URL)

			rgName, err := GetResourceGroupName(mClient, tc.args.rgID)
			if tc.want.err != nil && tc.want.err.Error() != err.Error() {
				t.Errorf("GetResourceGroupName(...): want: %s\ngot: %s\n", tc.want.err, err)
			}
			if diff := cmp.Diff(tc.want.rgName, rgName); diff != "" {
				t.Errorf("GetResourceGroupName(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestGetResourceInstanceTags(t *testing.T) {
	type args struct {
		crn string
	}
	type want struct {
		tags []string
		err  error
	}
	cases := map[string]struct {
		args args
		want want
	}{
		"Found": {
			args: args{
				crn: testCrn,
			},
			want: want{tags: []string{"dev"}, err: nil},
		},
		// the API does not return 404 for not found CRN, only an empty list
		// so there is no way to distinguish between not found CRN and not found tags
		"NotFound": {
			args: args{
				crn: invalidCrn,
			},
			want: want{tags: nil, err: nil},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			mux := http.NewServeMux()
			mux.HandleFunc("/v3/tags/", tagsHandler)
			server := httptest.NewServer(mux)
			defer server.Close()

			mClient, _ := GetTestClient(server.URL)

			tags, err := GetResourceInstanceTags(mClient, tc.args.crn)
			if tc.want.err != nil && tc.want.err.Error() != err.Error() {
				t.Errorf("GetResourceInstanceTags(...): want: %s\ngot: %s\n", tc.want.err, err)
			}
			if diff := cmp.Diff(tc.want.tags, tags); diff != "" {
				t.Errorf("GetResourceInstanceTags(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestUpdateResourceInstanceTags(t *testing.T) {
	type args struct {
		crn  string
		tags []string
	}
	type want struct {
		tags []string
		err  error
	}
	cases := map[string]struct {
		args args
		want want
	}{
		"AddTags": {
			args: args{
				crn:  testCrn,
				tags: []string{"dev", "test"},
			},
			want: want{tags: []string{"dev", "test"}, err: nil},
		},
		"RemoveTags": {
			args: args{
				crn:  testCrn,
				tags: []string{},
			},
			want: want{tags: nil, err: nil},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			mux := http.NewServeMux()
			mux.HandleFunc("/v3/tags/", tagsHandler)
			server := httptest.NewServer(mux)
			defer server.Close()

			mClient, _ := GetTestClient(server.URL)

			err := UpdateResourceInstanceTags(mClient, tc.args.crn, tc.args.tags)
			if tc.want.err != nil && tc.want.err.Error() != err.Error() {
				t.Errorf("UpdateResourceInstanceTags(...): want: %s\ngot: %s\n", tc.want.err, err)
			}
			tags, err := GetResourceInstanceTags(mClient, tc.args.crn)
			if tc.want.err != nil && tc.want.err.Error() != err.Error() {
				t.Errorf("UpdateResourceInstanceTags(...): want: %s\ngot: %s\n", tc.want.err, err)
			}
			if diff := cmp.Diff(tc.want.tags, tags, cmpopts.SortSlices(func(x, y interface{}) bool {
				return fmt.Sprint("%# v", x) < fmt.Sprint("%# v", y)
			})); diff != "" {
				t.Errorf("UpdateResourceInstanceTags(...): -want, +got:\n%s", diff)
			}
		})
	}
}
