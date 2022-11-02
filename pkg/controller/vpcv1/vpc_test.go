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

package vpcv1

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	ibmVPC "github.com/IBM/vpc-go-sdk/vpcv1"
	"k8s.io/klog"

	cpv1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/reference"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane/crossplane-runtime/pkg/test"

	"github.com/crossplane-contrib/provider-ibm-cloud/apis/vpcv1/v1alpha1"

	crossplaneApi "github.com/crossplane-contrib/provider-ibm-cloud/apis/vpcv1/v1alpha1"
	crossplaneClientUtil "github.com/crossplane-contrib/provider-ibm-cloud/pkg/clients/vpcv1"
	crossplaneClient "github.com/crossplane-contrib/provider-ibm-cloud/pkg/clients/vpcv1/vpc"

	"github.com/crossplane-contrib/provider-ibm-cloud/pkg/controller/tstutil"
)

const (
	numTests = 3 // decent size (we want to avoid timing out, too)
)

var (
	tstNum                int
	booleanCombVPC        []bool
	varCombinationLogging string
)

// Interface to a function that takes as argument a VPC create request, and modifies it
type vpcModifier func(*crossplaneApi.VPC)

// Sets the external name of a VPC
func withExternalVPCName() vpcModifier {
	return func(c *crossplaneApi.VPC) {
		meta.SetExternalName(c, c.Status.AtProvider.CRN)
	}
}

// Sets the name in the spec part of the VPC
//
// Params
//
//	newName - the new name. If nil, it will be ignored, unless the second parameter is set
//	acceptNilAsNewName - will set the name to nil
func withVPCSpecName(newName *string, acceptNilAsNewName bool) vpcModifier {
	return func(c *crossplaneApi.VPC) {
		if newName != nil {
			c.Spec.ForProvider.Name = newName
		} else {
			if acceptNilAsNewName {
				c.Spec.ForProvider.Name = nil
			} else {
				// Make sure we use the same one everywhere, at the same time...
				vpcObs := crossplaneClient.GetDummyObservation(
					booleanCombVPC[0], booleanCombVPC[1], booleanCombVPC[2], booleanCombVPC[4], booleanCombVPC[5],
					booleanCombVPC[6], booleanCombVPC[7], booleanCombVPC[8], booleanCombVPC[9], booleanCombVPC[10], booleanCombVPC[11],
					booleanCombVPC[12], booleanCombVPC[13], booleanCombVPC[14], booleanCombVPC[15], booleanCombVPC[16], booleanCombVPC[17],
					booleanCombVPC[18], booleanCombVPC[19], booleanCombVPC[20], booleanCombVPC[21], booleanCombVPC[22], booleanCombVPC[23],
					booleanCombVPC[24], booleanCombVPC[25], booleanCombVPC[26], booleanCombVPC[27], booleanCombVPC[28], booleanCombVPC[29],
					booleanCombVPC[30], booleanCombVPC[31], booleanCombVPC[32])

				c.Spec.ForProvider.Name = vpcObs.Name
			}
		}
	}
}

// Sets the resource group in the spec part of the VPC
func withVPCResourceGroup() vpcModifier {
	return func(c *crossplaneApi.VPC) {
		// Make sure we use the same one everywhere, at the same time...
		vpcObs := crossplaneClient.GetDummyObservation(
			booleanCombVPC[0], booleanCombVPC[1], booleanCombVPC[2], booleanCombVPC[4], booleanCombVPC[5],
			booleanCombVPC[6], booleanCombVPC[7], booleanCombVPC[8], booleanCombVPC[9], booleanCombVPC[10], booleanCombVPC[11],
			booleanCombVPC[12], booleanCombVPC[13], booleanCombVPC[14], booleanCombVPC[15], booleanCombVPC[16], booleanCombVPC[17],
			booleanCombVPC[18], booleanCombVPC[19], booleanCombVPC[20], booleanCombVPC[21], booleanCombVPC[22], booleanCombVPC[23],
			booleanCombVPC[24], booleanCombVPC[25], booleanCombVPC[26], booleanCombVPC[27], booleanCombVPC[28], booleanCombVPC[29],
			booleanCombVPC[30], booleanCombVPC[31], booleanCombVPC[32])

		if vpcObs.ResourceGroup != nil {
			c.Spec.ForProvider.ResourceGroup = &v1alpha1.ResourceGroupIdentity{
				ID: *vpcObs.ResourceGroup.ID,
			}
		}
	}
}

// Returns a function that sets the cluster conditions
func withVPCConditions(c ...cpv1alpha1.Condition) vpcModifier {
	return func(i *crossplaneApi.VPC) {
		i.Status.SetConditions(c...)
	}
}

// Sets the status part of a VPC
func withVPCStatus() vpcModifier {
	return func(c *crossplaneApi.VPC) {
		vpcObs := crossplaneClient.GetDummyObservation(
			booleanCombVPC[0], booleanCombVPC[1], booleanCombVPC[2], booleanCombVPC[4], booleanCombVPC[5],
			booleanCombVPC[6], booleanCombVPC[7], booleanCombVPC[8], booleanCombVPC[9], booleanCombVPC[10], booleanCombVPC[11],
			booleanCombVPC[12], booleanCombVPC[13], booleanCombVPC[14], booleanCombVPC[15], booleanCombVPC[16], booleanCombVPC[17],
			booleanCombVPC[18], booleanCombVPC[19], booleanCombVPC[20], booleanCombVPC[21], booleanCombVPC[22], booleanCombVPC[23],
			booleanCombVPC[24], booleanCombVPC[25], booleanCombVPC[26], booleanCombVPC[27], booleanCombVPC[28], booleanCombVPC[29],
			booleanCombVPC[30], booleanCombVPC[31], booleanCombVPC[32])

		c.Status.AtProvider, _ = crossplaneClient.GenerateObservation(&vpcObs)
	}
}

// Creates a VPC, by creating a generic one + applying a list of modifiers to the Spec part (which is the only one populated)
//
// Params
//
//			addressNil - whether to set the 'AddressPrefixManagement' member to nil
//			nameNil - whether to set the 'Name' member to nil
//			resourceGroupIDNil - whether to set the 'resourceGroupIDNil' member to nil
//	     modifiers... - well, a list thereof
//
// Returns
//
//	a VPC
func createCrossplaneVPC(addressNil bool, nameNil bool, resourceGroupIDNil bool, modifiers ...vpcModifier) *crossplaneApi.VPC {
	result := &crossplaneApi.VPC{
		Spec: crossplaneApi.VPCSpec{
			ForProvider: crossplaneClient.GetDummyCreateParams(addressNil, nameNil, resourceGroupIDNil),
		},
		Status: crossplaneApi.VPCStatus{
			AtProvider: crossplaneApi.VPCObservation{},
		},
	}

	for _, modifier := range modifiers {
		modifier(result)
	}

	return result
}

// Sets up a unit test http server, and creates an external cluster structure appropriate for unit test.
//
// Params
//
//	testingObj - the test object
//	handlers - the handlers that create the responses
//	client - the controller runtime client
//
// Returns
//   - the external object, ready for unit test
//   - the test http server, on which the caller should call 'defer ....Close()' (reason for this is we need to keep it around to prevent
//     garbage collection)
//     -- an error (if...)
func setupServerAndGetUnitTestExternalVPC(testingObj *testing.T, handlers *[]tstutil.Handler, kube *client.Client) (*vpcExternal, *httptest.Server, error) {
	mClient, tstServer, err := tstutil.SetupTestServerClient(testingObj, handlers)
	if err != nil || mClient == nil || tstServer == nil {
		return nil, nil, err
	}

	return &vpcExternal{
			kube:   *kube,
			client: *mClient,
			logger: logging.NewNopLogger(),
		},
		tstServer,
		nil
}

// Tests the VPC "Create" method
func testCreateVPC(t *testing.T) {
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
						_ = r.Body.Close()

						if diff := cmp.Diff(http.MethodPost, r.Method); diff != "" {
							t.Errorf("Test: "+varCombinationLogging+", r: -want, +got:\n%s", diff)
						}
						w.Header().Set("Content-Type", "application/json")
						w.WriteHeader(http.StatusCreated)

						ibmVPCInfo := crossplaneClient.GetDummyObservation(
							booleanCombVPC[0], booleanCombVPC[1], booleanCombVPC[2], booleanCombVPC[3], booleanCombVPC[4],
							booleanCombVPC[5], booleanCombVPC[6], booleanCombVPC[7], booleanCombVPC[8], booleanCombVPC[9],
							booleanCombVPC[10], booleanCombVPC[11], booleanCombVPC[12], booleanCombVPC[13], booleanCombVPC[14],
							booleanCombVPC[15], booleanCombVPC[16], booleanCombVPC[17], booleanCombVPC[18], booleanCombVPC[19],
							booleanCombVPC[20], booleanCombVPC[21], booleanCombVPC[22], booleanCombVPC[23], booleanCombVPC[24],
							booleanCombVPC[25], booleanCombVPC[26], booleanCombVPC[27], booleanCombVPC[28], booleanCombVPC[29],
							booleanCombVPC[30], booleanCombVPC[31])

						err := json.NewEncoder(w).Encode(ibmVPCInfo)
						if err != nil {
							klog.Errorf("%s", err)
						}
					},
				},
			},
			args: tstutil.Args{
				Managed: createCrossplaneVPC(booleanCombVPC[0], booleanCombVPC[1], booleanCombVPC[2]),
			},
			want: want{
				mg: createCrossplaneVPC(booleanCombVPC[0], booleanCombVPC[1], booleanCombVPC[2],
					withVPCConditions(cpv1alpha1.Creating())),
				cre: managed.ExternalCreation{ExternalNameAssigned: true},
				err: nil,
			},
			kube: &test.MockClient{
				MockUpdate: test.NewMockUpdateFn(nil),
			},
		},
		"Failed": {
			handlers: []tstutil.Handler{
				{
					Path: "/",
					HandlerFunc: func(w http.ResponseWriter, r *http.Request) {
						_ = r.Body.Close()

						if diff := cmp.Diff(http.MethodPost, r.Method); diff != "" {
							t.Errorf("Test: "+varCombinationLogging+", r: -want, +got:\n%s", diff)
						}
						w.Header().Set("Content-Type", "application/json")
						w.WriteHeader(http.StatusBadRequest)
					},
				},
			},
			args: tstutil.Args{
				Managed: createCrossplaneVPC(booleanCombVPC[0], booleanCombVPC[1], booleanCombVPC[2]),
			},
			want: want{
				mg: createCrossplaneVPC(booleanCombVPC[0], booleanCombVPC[1], booleanCombVPC[2],
					withVPCConditions(cpv1alpha1.Creating())),
				cre: managed.ExternalCreation{ExternalNameAssigned: false},
				err: errors.Wrap(errors.New(http.StatusText(http.StatusBadRequest)), errCreateVPC),
			},
			kube: &test.MockClient{
				MockUpdate: test.NewMockUpdateFn(nil),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e, server, err := setupServerAndGetUnitTestExternalVPC(t, &tc.handlers, &tc.kube)
			if err != nil {
				t.Errorf("Test: "+varCombinationLogging+", Create(...): problem setting up the test server %s", err)
			}

			defer server.Close()

			cre, err := e.Create(context.Background(), tc.args.Managed)
			if tc.want.err != nil && err != nil {
				wantedPrefix := strings.Split(tc.want.err.Error(), ":")[0]
				actualPrefix := strings.Split(err.Error(), ":")[0]
				if diff := cmp.Diff(wantedPrefix, actualPrefix); diff != "" {
					t.Errorf("Test: "+varCombinationLogging+", Create(...): -want, +got:\n%s", diff)
				}
			} else if diff := cmp.Diff(tc.want.err, err); diff != "" {
				t.Errorf("Test: "+varCombinationLogging+", Create(...): -want, +got:\n%s", diff)
			}

			if diff := cmp.Diff(tc.want.cre, cre); diff != "" {
				t.Errorf("Test: "+varCombinationLogging+", Create(...): -want, +got:\n%s", diff)
			}
		})
	}
}

// Tests the VPC "Delete" method
func testDeleteVPC(t *testing.T) {
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
						_ = r.Body.Close()

						if diff := cmp.Diff(http.MethodDelete, r.Method); diff != "" {
							t.Errorf("Test: "+varCombinationLogging+", r: -want, +got:\n%s", diff)
						}

						w.Header().Set("Content-Type", "application/json")
						w.WriteHeader(http.StatusAccepted)
					},
				},
			},
			args: tstutil.Args{
				Managed: createCrossplaneVPC(booleanCombVPC[0], booleanCombVPC[1], booleanCombVPC[2], withVPCStatus()),
			},
			want: want{
				mg: createCrossplaneVPC(booleanCombVPC[0], booleanCombVPC[1], booleanCombVPC[2], withVPCStatus(),
					withVPCConditions(cpv1alpha1.Deleting())),
				err: nil,
			},
			kube: &test.MockClient{
				MockUpdate: test.NewMockUpdateFn(nil),
			},
		},
		"AlreadyGone": {
			handlers: []tstutil.Handler{
				{
					Path: "/",
					HandlerFunc: func(w http.ResponseWriter, r *http.Request) {
						_ = r.Body.Close()

						if diff := cmp.Diff(http.MethodDelete, r.Method); diff != "" {
							t.Errorf("Test: "+varCombinationLogging+", r: -want, +got:\n%s", diff)
						}

						w.Header().Set("Content-Type", "application/json")
						w.WriteHeader(http.StatusNotFound)
					},
				},
			},
			args: tstutil.Args{
				Managed: createCrossplaneVPC(booleanCombVPC[0], booleanCombVPC[1], booleanCombVPC[2], withVPCStatus()),
			},
			want: want{
				mg: createCrossplaneVPC(booleanCombVPC[0], booleanCombVPC[1], booleanCombVPC[2], withVPCStatus(),
					withVPCConditions(cpv1alpha1.Deleting())),
				err: nil,
			},
			kube: &test.MockClient{
				MockUpdate: test.NewMockUpdateFn(nil),
			},
		},
		"Failed": {
			handlers: []tstutil.Handler{
				{
					Path: "/",
					HandlerFunc: func(w http.ResponseWriter, r *http.Request) {
						_ = r.Body.Close()

						if diff := cmp.Diff(http.MethodDelete, r.Method); diff != "" {
							t.Errorf("Test: "+varCombinationLogging+", r: -want, +got:\n%s", diff)
						}

						w.Header().Set("Content-Type", "application/json")
						w.WriteHeader(http.StatusBadRequest)
					},
				},
			},
			args: tstutil.Args{
				Managed: createCrossplaneVPC(booleanCombVPC[0], booleanCombVPC[1], booleanCombVPC[2], withVPCStatus()),
			},
			want: want{
				mg: createCrossplaneVPC(booleanCombVPC[0], booleanCombVPC[1], booleanCombVPC[2], withVPCStatus(),
					withVPCConditions(cpv1alpha1.Deleting())),
				err: errors.Wrap(errors.New(http.StatusText(http.StatusBadRequest)), errDeleteVPC),
			},
			kube: &test.MockClient{
				MockUpdate: test.NewMockUpdateFn(nil),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e, server, setupErr := setupServerAndGetUnitTestExternalVPC(t, &tc.handlers, &tc.kube)
			if setupErr != nil {
				t.Errorf("Test: "+varCombinationLogging+", Delete(...): problem setting up the test server %s", setupErr)
			}

			defer server.Close()

			err := e.Delete(context.Background(), tc.args.Managed)
			if tc.want.err != nil && err != nil {
				// the case where our mock server returns error, is tricky, as the returned error string is long/spans multiple lines
				wantedPrefix := strings.Split(tc.want.err.Error(), ":")[0]
				actualPrefix := strings.Split(err.Error(), ":")[0]
				if diff := cmp.Diff(wantedPrefix, actualPrefix); diff != "" {
					t.Errorf("Test: "+varCombinationLogging+", Delete(...): -want, +got:\n%s", diff)
				}
			} else if diff := cmp.Diff(tc.want.err, err); diff != "" {
				t.Errorf("Test: "+varCombinationLogging+", Delete(...): -want, +got:\n%s", diff)
			}

			if diff := cmp.Diff(tc.want.mg, tc.args.Managed); diff != "" {
				t.Errorf("Test: "+varCombinationLogging+", Delete(...): -want, +got:\n%s", diff)
			}
		})
	}
}

// Tests the VPC "Observe" method
func testObserveVPC(t *testing.T) {
	type errInfo struct {
		err     error
		errCode int
	}

	type want struct {
		mg      resource.Managed
		obs     managed.ExternalObservation
		errInfo *errInfo
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
							t.Errorf("Test: "+varCombinationLogging+", r: -want, +got:\n%s", diff)
						}

						// content type should always set before writeHeader()
						w.Header().Set("Content-Type", "application/json")
						w.WriteHeader(http.StatusNotFound)
					},
				},
			},
			args: tstutil.Args{
				Managed: createCrossplaneVPC(booleanCombVPC[0], booleanCombVPC[1], booleanCombVPC[2], withVPCStatus(), withExternalVPCName()),
			},
			want: want{
				mg:  createCrossplaneVPC(booleanCombVPC[0], booleanCombVPC[1], booleanCombVPC[2], withVPCStatus(), withExternalVPCName()),
				obs: managed.ExternalObservation{ResourceExists: false},
			},
			kube: &test.MockClient{
				MockUpdate: test.NewMockUpdateFn(nil),
			},
		},
		"GetFailed": {
			handlers: []tstutil.Handler{
				{
					Path: "/",
					HandlerFunc: func(w http.ResponseWriter, r *http.Request) {
						_ = r.Body.Close()

						if diff := cmp.Diff(http.MethodGet, r.Method); diff != "" {
							t.Errorf("Test: "+varCombinationLogging+", r: -want, +got:\n%s", diff)
						}

						w.Header().Set("Content-Type", "application/json")
						w.WriteHeader(http.StatusBadRequest)
					},
				},
			},
			args: tstutil.Args{
				Managed: createCrossplaneVPC(booleanCombVPC[0], booleanCombVPC[1], booleanCombVPC[2], withVPCStatus(), withExternalVPCName()),
			},
			want: want{
				mg:  createCrossplaneVPC(booleanCombVPC[0], booleanCombVPC[1], booleanCombVPC[2], withVPCStatus(), withExternalVPCName()),
				obs: managed.ExternalObservation{},
				errInfo: &errInfo{
					err:     errors.Wrap(errors.New(http.StatusText(http.StatusBadRequest)), errGetFailedVPC),
					errCode: http.StatusBadRequest,
				},
			},
			kube: &test.MockClient{
				MockUpdate: test.NewMockUpdateFn(nil),
			},
		},
		"GetForbidden": {
			handlers: []tstutil.Handler{
				{
					Path: "/",
					HandlerFunc: func(w http.ResponseWriter, r *http.Request) {
						_ = r.Body.Close()

						if diff := cmp.Diff(http.MethodGet, r.Method); diff != "" {
							t.Errorf("Test: "+varCombinationLogging+", r: -want, +got:\n%s", diff)
						}

						w.Header().Set("Content-Type", "application/json")
						w.WriteHeader(http.StatusForbidden)
					},
				},
			},
			args: tstutil.Args{
				Managed: createCrossplaneVPC(booleanCombVPC[0], booleanCombVPC[1], booleanCombVPC[2], withVPCStatus(), withExternalVPCName()),
			},
			want: want{
				mg:  createCrossplaneVPC(booleanCombVPC[0], booleanCombVPC[1], booleanCombVPC[2], withVPCStatus(), withExternalVPCName()),
				obs: managed.ExternalObservation{},
				errInfo: &errInfo{
					err:     errors.Wrap(errors.New(http.StatusText(http.StatusForbidden)), errGetFailedVPC),
					errCode: http.StatusForbidden,
				},
			},
			kube: &test.MockClient{
				MockUpdate: test.NewMockUpdateFn(nil),
			},
		},
		"UpToDate": {
			handlers: []tstutil.Handler{
				{
					Path: "/",
					HandlerFunc: func(w http.ResponseWriter, r *http.Request) {
						_ = r.Body.Close()
						if diff := cmp.Diff(http.MethodGet, r.Method); diff != "" {
							t.Errorf("r: -want, +got:\n%s", diff)
						}

						w.Header().Set("Content-Type", "application/json")
						w.WriteHeader(http.StatusOK)

						collection := ibmVPC.VPCCollection{
							Vpcs: make([]ibmVPC.VPC, 3),
						}

						collection.Vpcs[0] = crossplaneClient.GetDummyObservation(
							booleanCombVPC[0], booleanCombVPC[1], booleanCombVPC[2], booleanCombVPC[4],
							booleanCombVPC[5], booleanCombVPC[6], booleanCombVPC[7], booleanCombVPC[8], booleanCombVPC[9],
							booleanCombVPC[10], booleanCombVPC[11], booleanCombVPC[12], booleanCombVPC[13], booleanCombVPC[14],
							booleanCombVPC[15], booleanCombVPC[16], booleanCombVPC[17], booleanCombVPC[18], booleanCombVPC[19],
							booleanCombVPC[20], booleanCombVPC[21], booleanCombVPC[22], booleanCombVPC[23], booleanCombVPC[24],
							booleanCombVPC[25], booleanCombVPC[26], booleanCombVPC[27], booleanCombVPC[28], booleanCombVPC[29],
							booleanCombVPC[30], booleanCombVPC[31], booleanCombVPC[32])

						collection.Vpcs[1] = crossplaneClient.GetDummyObservation(
							booleanCombVPC[0], booleanCombVPC[1], booleanCombVPC[2], booleanCombVPC[4],
							booleanCombVPC[5], booleanCombVPC[6], booleanCombVPC[7], booleanCombVPC[8], booleanCombVPC[9],
							booleanCombVPC[10], booleanCombVPC[11], booleanCombVPC[12], booleanCombVPC[13], booleanCombVPC[14],
							booleanCombVPC[15], booleanCombVPC[16], booleanCombVPC[17], booleanCombVPC[18], booleanCombVPC[19],
							booleanCombVPC[20], booleanCombVPC[21], booleanCombVPC[22], booleanCombVPC[23], booleanCombVPC[24],
							booleanCombVPC[25], booleanCombVPC[26], booleanCombVPC[27], booleanCombVPC[28], booleanCombVPC[29],
							booleanCombVPC[30], !booleanCombVPC[31], !booleanCombVPC[32])

						collection.Vpcs[2] = crossplaneClient.GetDummyObservation(
							!booleanCombVPC[0], booleanCombVPC[1], booleanCombVPC[2], booleanCombVPC[4],
							booleanCombVPC[5], booleanCombVPC[6], booleanCombVPC[7], booleanCombVPC[8], booleanCombVPC[9],
							booleanCombVPC[10], booleanCombVPC[11], booleanCombVPC[12], booleanCombVPC[13], booleanCombVPC[14],
							booleanCombVPC[15], !booleanCombVPC[16], booleanCombVPC[17], booleanCombVPC[18], booleanCombVPC[19],
							!booleanCombVPC[20], booleanCombVPC[21], booleanCombVPC[22], booleanCombVPC[23], booleanCombVPC[24],
							booleanCombVPC[25], !booleanCombVPC[26], booleanCombVPC[27], booleanCombVPC[28], booleanCombVPC[29],
							booleanCombVPC[30], !booleanCombVPC[31], !booleanCombVPC[32])

						err := json.NewEncoder(w).Encode(collection)
						if err != nil {
							klog.Errorf("%s", err)
						}
					},
				},
			},
			args: tstutil.Args{
				Managed: createCrossplaneVPC(booleanCombVPC[0], booleanCombVPC[1], booleanCombVPC[2], withVPCStatus(), withExternalVPCName()),
			},
			want: want{
				mg: createCrossplaneVPC(booleanCombVPC[0], booleanCombVPC[1], booleanCombVPC[2], withVPCStatus(), withExternalVPCName(),
					withVPCSpecName(nil, false), withVPCResourceGroup()),
				obs: managed.ExternalObservation{
					ResourceExists:    true,
					ResourceUpToDate:  true,
					ConnectionDetails: nil,
				},
			},
			kube: &test.MockClient{
				MockUpdate: test.NewMockUpdateFn(nil),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e, server, errCr := setupServerAndGetUnitTestExternalVPC(t, &tc.handlers, &tc.kube)
			if errCr != nil {
				t.Errorf("Test: "+varCombinationLogging+", Observe(...): problem setting up the test server %s", errCr)
			}

			defer server.Close()

			obs, err := e.Observe(context.Background(), tc.args.Managed)
			if tc.want.errInfo != nil && err != nil {
				if diff := cmp.Diff(tc.want.errInfo.err.Error(), err.Error()); diff != "" {
					t.Errorf("Test: "+varCombinationLogging+", Observe(...): -want, +got:\n%s", diff)
				}
			} else if tc.want.errInfo != nil {
				if diff := cmp.Diff(tc.want.errInfo.err, err); diff != "" {
					t.Errorf("Test: "+varCombinationLogging+", Observe(...): want error != got error:\n%s", diff)
				}
			}

			if diff := cmp.Diff(tc.want.obs, obs); diff != "" {
				t.Errorf("Test: "+varCombinationLogging+", Observe(...): -want, +got:\n%s", diff)
			}

			if diff := cmp.Diff(tc.want.mg, tc.args.Managed); diff != "" {
				t.Errorf("Test: "+varCombinationLogging+", Observe(...): -want, +got:\n%s", diff)
			}
		})
	}
}

// Tests the VPC "Create" method several times (with various fields set to nil).
//
// The # of times/combinations is the value of variable 'numTests'
func TestCreateVPC(t *testing.T) {
	for tstNum, booleanCombVPC = range crossplaneClientUtil.GenerateSomePermutations(numTests, 33, true) {
		varCombinationLogging = crossplaneClientUtil.GetBinaryRep(tstNum, numTests)

		testCreateVPC(t)
	}
}

// Tests the VPC "Delete" method several times (with various fields set to nil).
//
// The # of times/combinations is the value of variable 'numTests'
func TestDeleteVPC(t *testing.T) {
	for tstNum, booleanCombVPC = range crossplaneClientUtil.GenerateSomePermutations(numTests, 33, true) {
		varCombinationLogging = crossplaneClientUtil.GetBinaryRep(tstNum, numTests)

		testDeleteVPC(t)
	}
}

// Tests the VPC "Observe" method several times (with various fields set to nil).
//
// The # of times/combinations is the value of variable 'numTests'
func TestObserveVPC(t *testing.T) {
	for tstNum, booleanCombVPC = range crossplaneClientUtil.GenerateSomePermutations(numTests, 33, true) {
		varCombinationLogging = crossplaneClientUtil.GetBinaryRep(tstNum, numTests)

		testObserveVPC(t)
	}
}

func TestUpdateVPC(t *testing.T) {
	booleanCombVPC = crossplaneClientUtil.GenerateSomePermutations(1, 35, true)[0]

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
		name     string // used for debugging convenience
	}{
		"Successful-1": {
			handlers: []tstutil.Handler{
				{
					Path: "/",
					HandlerFunc: func(w http.ResponseWriter, r *http.Request) {
						_ = r.Body.Close()

						if diff := cmp.Diff(http.MethodPatch, r.Method); diff != "" {
							t.Errorf("r: -want, +got:\n%s", diff)
						}

						// content type should always set before writeHeader()
						w.Header().Set("Content-Type", "application/json")
						w.WriteHeader(http.StatusOK)
					},
				},
			},
			args: tstutil.Args{
				Managed: createCrossplaneVPC(true, false, true, withVPCSpecName(reference.ToPtrValue("new name"), false),
					withVPCStatus()),
			},
			want: want{
				mg: createCrossplaneVPC(true, false, true, withVPCSpecName(reference.ToPtrValue("new name"), false),
					withVPCStatus()),
				upd: managed.ExternalUpdate{},
			},
		},
		"Successful-2": {
			handlers: []tstutil.Handler{
				{
					Path: "/",
					HandlerFunc: func(w http.ResponseWriter, r *http.Request) {
						_ = r.Body.Close()

						if diff := cmp.Diff(http.MethodPatch, r.Method); diff != "" {
							t.Errorf("r: -want, +got:\n%s", diff)
						}

						// content type should always set before writeHeader()
						w.Header().Set("Content-Type", "application/json")
						w.WriteHeader(http.StatusOK)
					},
				},
			},
			args: tstutil.Args{
				Managed: createCrossplaneVPC(true, false, true, withVPCSpecName(nil, true),
					withVPCStatus()),
			},
			want: want{
				mg: createCrossplaneVPC(true, false, true, withVPCSpecName(nil, true),
					withVPCStatus()),
				upd: managed.ExternalUpdate{},
			},
		},
		"Failed-1": {
			handlers: []tstutil.Handler{
				{
					Path: "/",
					HandlerFunc: func(w http.ResponseWriter, r *http.Request) {
						_ = r.Body.Close()

						if diff := cmp.Diff(http.MethodPatch, r.Method); diff != "" {
							t.Errorf("r: -want, +got:\n%s", diff)
						}

						// content type should always set before writeHeader()
						w.Header().Set("Content-Type", "application/json")
						w.WriteHeader(http.StatusBadRequest)
					},
				},
			},
			args: tstutil.Args{
				Managed: createCrossplaneVPC(true, false, true, withVPCSpecName(nil, true),
					withVPCStatus()),
			},
			want: want{
				mg: createCrossplaneVPC(true, false, true, withVPCSpecName(nil, true),
					withVPCStatus()),
				upd: managed.ExternalUpdate{},
				err: errors.Wrap(errors.New(http.StatusText(http.StatusBadRequest)), errUpdateVPC),
			},
		},
	}

	for name, tc := range cases {
		tc.name = name

		t.Run(name, func(t *testing.T) {
			e, server, errCr := setupServerAndGetUnitTestExternalVPC(t, &tc.handlers, &tc.kube)
			if errCr != nil {
				t.Errorf(tc.name+", Update(...): problem setting up the test server %s", errCr)
			}

			defer server.Close()

			xu, err := e.Update(context.Background(), tc.args.Managed)
			if tc.want.err != nil && err != nil {
				if diff := cmp.Diff(tc.want.err.Error(), err.Error()); diff != "" {
					t.Errorf(tc.name+", Update(...): want error string != got error string:\n%s", diff)
				}
			} else if diff := cmp.Diff(tc.want.err, err); diff != "" {
				t.Errorf(tc.name+", Update(...): want error != got error:\n%s", diff)
			}

			if diff := cmp.Diff(tc.want.upd, xu); diff != "" {
				t.Errorf(tc.name+", Update(...): -want, +got:\n%s", diff)
			}

			if tc.want.mg != nil {
				if diff := cmp.Diff(tc.want.mg, tc.args.Managed); diff != "" {
					t.Errorf(tc.name+", Update(...): -want, +got:\n%s", diff)
				}
			}
		})
	}
}
