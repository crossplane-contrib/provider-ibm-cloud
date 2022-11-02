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
	crossplaneClient "github.com/crossplane-contrib/provider-ibm-cloud/pkg/clients/vpcv1/subnet"

	"github.com/crossplane-contrib/provider-ibm-cloud/pkg/controller/tstutil"
)

var booleanCombSubnet []bool

// Interface to a function that takes as argument a Subnet create request, and modifies it
type subnetModifier func(*crossplaneApi.Subnet)

// Sets the external name of a Subnet. If necessary, sets its CRN, too
func withExternalSubnetName() subnetModifier {
	return func(c *crossplaneApi.Subnet) {
		if c.Status.AtProvider.CRN == "" {
			c.Status.AtProvider.CRN = crossplaneClient.CrnVal
		}

		meta.SetExternalName(c, c.Status.AtProvider.CRN)
	}
}

// Sets the name in the spec part of the Subnet
//
// Params
//
//	newName - the new name. If nil, it will be ignored, unless the second parameter is set
//	acceptNilAsNewName - will set the name to nil
func withSubnetSpecName(newName *string, acceptNilAsNewName bool) subnetModifier {
	return func(c *crossplaneApi.Subnet) {
		if newName != nil {
			if c.Spec.ForProvider.ByTocalCount != nil {
				c.Spec.ForProvider.ByTocalCount.Name = newName
			} else {
				c.Spec.ForProvider.ByCIDR.Name = newName
			}
		} else {
			if acceptNilAsNewName {
				if c.Spec.ForProvider.ByTocalCount != nil {
					c.Spec.ForProvider.ByTocalCount.Name = newName
				} else {
					c.Spec.ForProvider.ByCIDR.Name = newName
				}
			} else {
				// Make sure we use the same one everywhere, at the same time...
				obs := crossplaneClient.GetDummyObservation(
					booleanCombSubnet[0], booleanCombSubnet[1], false, booleanCombSubnet[3], booleanCombSubnet[4],
					booleanCombSubnet[5], booleanCombSubnet[6], booleanCombSubnet[7], booleanCombSubnet[8], booleanCombSubnet[9],
					booleanCombSubnet[10], booleanCombSubnet[11], booleanCombSubnet[12], booleanCombSubnet[13], booleanCombSubnet[14],
					booleanCombSubnet[15], booleanCombSubnet[16], booleanCombSubnet[17], booleanCombSubnet[18], booleanCombSubnet[19],
					booleanCombSubnet[20], booleanCombSubnet[21], booleanCombSubnet[22], booleanCombSubnet[23], booleanCombSubnet[24],
					booleanCombSubnet[25], booleanCombSubnet[26], booleanCombSubnet[27], booleanCombSubnet[28], booleanCombSubnet[29],
					booleanCombSubnet[30], booleanCombSubnet[31], booleanCombSubnet[32], booleanCombSubnet[33], booleanCombSubnet[34],
					booleanCombSubnet[35], booleanCombSubnet[36], booleanCombSubnet[37], booleanCombSubnet[38], booleanCombSubnet[39],
					booleanCombSubnet[40], booleanCombSubnet[41], booleanCombSubnet[42], booleanCombSubnet[43], booleanCombSubnet[44],
					booleanCombSubnet[45])

				if c.Spec.ForProvider.ByTocalCount != nil {
					c.Spec.ForProvider.ByTocalCount.Name = obs.Name
				} else {
					c.Spec.ForProvider.ByCIDR.Name = obs.Name
				}
			}
		}
	}
}

// Sets the resource group in the spec part of the Subnet
func withSubnetResourceGroup() subnetModifier {
	return func(c *crossplaneApi.Subnet) {
		// Make sure we use the same one everywhere, at the same time...
		obs := crossplaneClient.GetDummyObservation(
			booleanCombSubnet[0], booleanCombSubnet[1], false, booleanCombSubnet[3], booleanCombSubnet[4],
			booleanCombSubnet[5], booleanCombSubnet[6], booleanCombSubnet[7], booleanCombSubnet[8], booleanCombSubnet[9],
			booleanCombSubnet[10], booleanCombSubnet[11], booleanCombSubnet[12], booleanCombSubnet[13], booleanCombSubnet[14],
			booleanCombSubnet[15], booleanCombSubnet[16], booleanCombSubnet[17], booleanCombSubnet[18], booleanCombSubnet[19],
			booleanCombSubnet[20], booleanCombSubnet[21], booleanCombSubnet[22], booleanCombSubnet[23], booleanCombSubnet[24],
			booleanCombSubnet[25], booleanCombSubnet[26], booleanCombSubnet[27], booleanCombSubnet[28], booleanCombSubnet[29],
			booleanCombSubnet[30], booleanCombSubnet[31], booleanCombSubnet[32], booleanCombSubnet[33], booleanCombSubnet[34],
			booleanCombSubnet[35], booleanCombSubnet[36], booleanCombSubnet[37], booleanCombSubnet[38], booleanCombSubnet[39],
			booleanCombSubnet[40], booleanCombSubnet[41], booleanCombSubnet[42], booleanCombSubnet[43], booleanCombSubnet[44],
			booleanCombSubnet[45])

		if obs.ResourceGroup != nil {
			if booleanCombSubnet[0] {
				c.Spec.ForProvider.ByTocalCount.ResourceGroup = &v1alpha1.ResourceGroupIdentity{
					ID: *obs.ResourceGroup.ID,
				}
			} else {
				c.Spec.ForProvider.ByCIDR.ResourceGroup = &v1alpha1.ResourceGroupIdentity{
					ID: *obs.ResourceGroup.ID,
				}
			}
		}
	}
}

// Returns a function that sets the cluster conditions
func withSubnetConditions(c ...cpv1alpha1.Condition) subnetModifier {
	return func(i *crossplaneApi.Subnet) {
		i.Status.SetConditions(c...)
	}
}

// Sets the status part of a Subnet
func withSubnetStatus() subnetModifier {
	return func(c *crossplaneApi.Subnet) {
		obs := crossplaneClient.GetDummyObservation(
			booleanCombSubnet[0], booleanCombSubnet[1], false, booleanCombSubnet[3], booleanCombSubnet[4],
			booleanCombSubnet[5], booleanCombSubnet[6], booleanCombSubnet[7], booleanCombSubnet[8], booleanCombSubnet[9],
			booleanCombSubnet[10], booleanCombSubnet[11], booleanCombSubnet[12], booleanCombSubnet[13], booleanCombSubnet[14],
			booleanCombSubnet[15], booleanCombSubnet[16], booleanCombSubnet[17], booleanCombSubnet[18], booleanCombSubnet[19],
			booleanCombSubnet[20], booleanCombSubnet[21], booleanCombSubnet[22], booleanCombSubnet[23], booleanCombSubnet[24],
			booleanCombSubnet[25], booleanCombSubnet[26], booleanCombSubnet[27], booleanCombSubnet[28], booleanCombSubnet[29],
			booleanCombSubnet[30], booleanCombSubnet[31], booleanCombSubnet[32], booleanCombSubnet[33], booleanCombSubnet[34],
			booleanCombSubnet[35], booleanCombSubnet[36], booleanCombSubnet[37], booleanCombSubnet[38], booleanCombSubnet[39],
			booleanCombSubnet[40], booleanCombSubnet[41], booleanCombSubnet[42], booleanCombSubnet[43], booleanCombSubnet[44],
			booleanCombSubnet[45])

		c.Status.AtProvider, _ = crossplaneClient.GenerateObservation(&obs)
	}
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
func setupServerAndGetUnitTestExternalSubnet(testingObj *testing.T, handlers *[]tstutil.Handler, kube *client.Client) (*subnetExternal, *httptest.Server, error) {
	mClient, tstServer, err := tstutil.SetupTestServerClient(testingObj, handlers)
	if err != nil || mClient == nil || tstServer == nil {
		return nil, nil, err
	}

	return &subnetExternal{
			kube:   *kube,
			client: *mClient,
			logger: logging.NewNopLogger(),
		},
		tstServer,
		nil
}

// Creates a Subnet, by creating a generic one + applying a list of modifiers to the Spec part (which is the only one populated)
//
// Params
//
//	     byTotalCount - whether the returned object is "ByTotalCount" or "ByCIDR"
//			ipVersionNil - whether to set the 'IPVersion' member to nil
//			nameNil - whether to set the 'Name' member to nil
//	     networkACLNil - whether to set the 'NetworkACL' member to nil
//	     publicGatewayNil - whether to set the 'PublicGateway' member to nil..
//			resourceGroupNil - whether to set the 'ResourceGroup' member to nil
//			routingTableNil - whether to set the 'RoutingTable' member to nil
//	     totalIpv4AddressCountNil - whether to set the 'TotalIpv4AddressCount' member to nil (does not apply if the returned object is 'ByTotalCount')
//	     zoneNil - whether to set the 'Zone' member to nil  (does not apply if the returned object is 'ByTotalCount')
//	     ipv4CIDRBlockNil - whether to set the 'Ipv4CIDRBlockNil' member to nil (does not apply if the returned object is 'ByCIDR')
//	     modifiers... - well, a list thereof
//
// Returns
//
//	a Subnet
func createCrossplaneSubnet(byTotalCount bool, ipVersionNil bool, nameNil bool, networkACLNil bool, publicGatewayNil bool,
	resourceGroupNil bool, routingTableNil bool, totalIpv4AddressCountNil bool, zoneNil bool, ipv4CIDRBlockNil bool,
	modifiers ...subnetModifier) *crossplaneApi.Subnet {
	result := &crossplaneApi.Subnet{
		Spec: crossplaneApi.SubnetSpec{
			ForProvider: crossplaneClient.GetDummyCreateParams(byTotalCount,
				ipVersionNil,
				nameNil,
				networkACLNil,
				publicGatewayNil,
				resourceGroupNil,
				routingTableNil,
				totalIpv4AddressCountNil,
				zoneNil,
				ipv4CIDRBlockNil),
		},
		Status: crossplaneApi.SubnetStatus{
			AtProvider: crossplaneApi.SubnetObservation{},
		},
	}

	for _, modifier := range modifiers {
		modifier(result)
	}

	return result
}

// Writes a an ibmVPC.SubnetCollection to the writer
func sendBackCollection(w http.ResponseWriter) {
	collection := ibmVPC.SubnetCollection{
		Subnets: make([]ibmVPC.Subnet, 2),
	}

	collection.Subnets[0] = crossplaneClient.GetDummyObservation(
		booleanCombSubnet[0], booleanCombSubnet[1], false, booleanCombSubnet[3], booleanCombSubnet[4],
		booleanCombSubnet[5], booleanCombSubnet[6], booleanCombSubnet[7], booleanCombSubnet[8], booleanCombSubnet[9],
		booleanCombSubnet[10], booleanCombSubnet[11], booleanCombSubnet[12], booleanCombSubnet[13], booleanCombSubnet[14],
		booleanCombSubnet[15], booleanCombSubnet[16], booleanCombSubnet[17], booleanCombSubnet[18], booleanCombSubnet[19],
		booleanCombSubnet[20], booleanCombSubnet[21], booleanCombSubnet[22], booleanCombSubnet[23], booleanCombSubnet[24],
		booleanCombSubnet[25], booleanCombSubnet[26], booleanCombSubnet[27], booleanCombSubnet[28], booleanCombSubnet[29],
		booleanCombSubnet[30], booleanCombSubnet[31], booleanCombSubnet[32], booleanCombSubnet[33], booleanCombSubnet[34],
		booleanCombSubnet[35], booleanCombSubnet[36], booleanCombSubnet[37], booleanCombSubnet[38], booleanCombSubnet[39],
		booleanCombSubnet[40], booleanCombSubnet[41], booleanCombSubnet[42], booleanCombSubnet[43], booleanCombSubnet[44],
		booleanCombSubnet[45])

	collection.Subnets[1] = crossplaneClient.GetDummyObservation(
		booleanCombSubnet[0], booleanCombSubnet[1], false, booleanCombSubnet[3], booleanCombSubnet[4],
		booleanCombSubnet[5], booleanCombSubnet[6], booleanCombSubnet[7], booleanCombSubnet[8], booleanCombSubnet[9],
		booleanCombSubnet[10], booleanCombSubnet[11], booleanCombSubnet[12], booleanCombSubnet[13], booleanCombSubnet[14],
		booleanCombSubnet[15], booleanCombSubnet[16], booleanCombSubnet[17], booleanCombSubnet[18], booleanCombSubnet[19],
		booleanCombSubnet[20], booleanCombSubnet[21], booleanCombSubnet[22], booleanCombSubnet[23], booleanCombSubnet[24],
		booleanCombSubnet[25], booleanCombSubnet[26], booleanCombSubnet[27], booleanCombSubnet[28], booleanCombSubnet[29],
		booleanCombSubnet[30], booleanCombSubnet[31], booleanCombSubnet[32], booleanCombSubnet[33], booleanCombSubnet[34],
		booleanCombSubnet[35], booleanCombSubnet[36], booleanCombSubnet[37], booleanCombSubnet[38], booleanCombSubnet[39],
		booleanCombSubnet[40], booleanCombSubnet[41], booleanCombSubnet[42], booleanCombSubnet[43], booleanCombSubnet[44],
		booleanCombSubnet[45])

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	err := json.NewEncoder(w).Encode(collection)
	if err != nil {
		klog.Errorf("%s", err)
	}
}

// Tests the Subnet "Create" method
func testCreateSubnet(t *testing.T) {
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

						cloudInfo := crossplaneClient.GetDummyObservation(
							booleanCombSubnet[0], booleanCombSubnet[1], false, booleanCombSubnet[3], booleanCombSubnet[4],
							booleanCombSubnet[5], booleanCombSubnet[6], booleanCombSubnet[7], booleanCombSubnet[8], booleanCombSubnet[9],
							booleanCombSubnet[10], booleanCombSubnet[11], booleanCombSubnet[12], booleanCombSubnet[13], booleanCombSubnet[14],
							booleanCombSubnet[15], booleanCombSubnet[16], booleanCombSubnet[17], booleanCombSubnet[18], booleanCombSubnet[19],
							booleanCombSubnet[20], booleanCombSubnet[21], booleanCombSubnet[22], booleanCombSubnet[23], booleanCombSubnet[24],
							booleanCombSubnet[25], booleanCombSubnet[26], booleanCombSubnet[27], booleanCombSubnet[28], booleanCombSubnet[29],
							booleanCombSubnet[30], booleanCombSubnet[31], booleanCombSubnet[32], booleanCombSubnet[33], booleanCombSubnet[34],
							booleanCombSubnet[35], booleanCombSubnet[36], booleanCombSubnet[37], booleanCombSubnet[38], booleanCombSubnet[39],
							booleanCombSubnet[40], booleanCombSubnet[41], booleanCombSubnet[42], booleanCombSubnet[43], booleanCombSubnet[44],
							booleanCombSubnet[45])

						err := json.NewEncoder(w).Encode(cloudInfo)
						if err != nil {
							klog.Errorf("%s", err)
						}
					},
				},
			},
			args: tstutil.Args{
				Managed: createCrossplaneSubnet(booleanCombSubnet[0], booleanCombSubnet[1], booleanCombSubnet[2], booleanCombSubnet[3], booleanCombSubnet[4],
					booleanCombSubnet[5], booleanCombSubnet[6], booleanCombSubnet[7], booleanCombSubnet[8], booleanCombSubnet[9]),
			},
			want: want{
				mg: createCrossplaneSubnet(booleanCombSubnet[0], booleanCombSubnet[1], booleanCombSubnet[2], booleanCombSubnet[3], booleanCombSubnet[4],
					booleanCombSubnet[5], booleanCombSubnet[6], booleanCombSubnet[7], booleanCombSubnet[8], booleanCombSubnet[9],
					withSubnetConditions(cpv1alpha1.Creating())),
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
				Managed: createCrossplaneSubnet(booleanCombSubnet[0], booleanCombSubnet[1], booleanCombSubnet[2], booleanCombSubnet[3], booleanCombSubnet[4],
					booleanCombSubnet[5], booleanCombSubnet[6], booleanCombSubnet[7], booleanCombSubnet[8], booleanCombSubnet[9]),
			},
			want: want{
				mg: createCrossplaneSubnet(booleanCombSubnet[0], booleanCombSubnet[1], booleanCombSubnet[2], booleanCombSubnet[3], booleanCombSubnet[4],
					booleanCombSubnet[5], booleanCombSubnet[6], booleanCombSubnet[7], booleanCombSubnet[8], booleanCombSubnet[9],
					withSubnetConditions(cpv1alpha1.Creating())),
				cre: managed.ExternalCreation{ExternalNameAssigned: false},
				err: errors.Wrap(errors.New(http.StatusText(http.StatusBadRequest)), errCreateSubnet),
			},
			kube: &test.MockClient{
				MockUpdate: test.NewMockUpdateFn(nil),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e, server, err := setupServerAndGetUnitTestExternalSubnet(t, &tc.handlers, &tc.kube)
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

// Tests the Subnet "Delete" method
func testDeleteSubnet(t *testing.T) {
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
				Managed: createCrossplaneSubnet(booleanCombSubnet[0], booleanCombSubnet[1], booleanCombSubnet[2], booleanCombSubnet[3], booleanCombSubnet[4],
					booleanCombSubnet[5], booleanCombSubnet[6], booleanCombSubnet[7], booleanCombSubnet[8], booleanCombSubnet[9], withSubnetStatus()),
			},
			want: want{
				mg: createCrossplaneSubnet(booleanCombSubnet[0], booleanCombSubnet[1], booleanCombSubnet[2], booleanCombSubnet[3], booleanCombSubnet[4],
					booleanCombSubnet[5], booleanCombSubnet[6], booleanCombSubnet[7], booleanCombSubnet[8], booleanCombSubnet[9], withSubnetStatus(),
					withSubnetConditions(cpv1alpha1.Deleting())),
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
				Managed: createCrossplaneSubnet(booleanCombSubnet[0], booleanCombSubnet[1], booleanCombSubnet[2], booleanCombSubnet[3], booleanCombSubnet[4],
					booleanCombSubnet[5], booleanCombSubnet[6], booleanCombSubnet[7], booleanCombSubnet[8], booleanCombSubnet[9], withSubnetStatus()),
			},
			want: want{
				mg: createCrossplaneSubnet(booleanCombSubnet[0], booleanCombSubnet[1], booleanCombSubnet[2], booleanCombSubnet[3], booleanCombSubnet[4],
					booleanCombSubnet[5], booleanCombSubnet[6], booleanCombSubnet[7], booleanCombSubnet[8], booleanCombSubnet[9], withSubnetStatus(),
					withSubnetConditions(cpv1alpha1.Deleting())),
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
				Managed: createCrossplaneSubnet(booleanCombSubnet[0], booleanCombSubnet[1], booleanCombSubnet[2], booleanCombSubnet[3], booleanCombSubnet[4],
					booleanCombSubnet[5], booleanCombSubnet[6], booleanCombSubnet[7], booleanCombSubnet[8], booleanCombSubnet[9], withSubnetStatus()),
			},
			want: want{
				mg: createCrossplaneSubnet(booleanCombSubnet[0], booleanCombSubnet[1], booleanCombSubnet[2], booleanCombSubnet[3], booleanCombSubnet[4],
					booleanCombSubnet[5], booleanCombSubnet[6], booleanCombSubnet[7], booleanCombSubnet[8], booleanCombSubnet[9], withSubnetStatus(),
					withSubnetConditions(cpv1alpha1.Deleting())),
				err: errors.Wrap(errors.New(http.StatusText(http.StatusBadRequest)), errDeleteSubnet),
			},
			kube: &test.MockClient{
				MockUpdate: test.NewMockUpdateFn(nil),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e, server, setupErr := setupServerAndGetUnitTestExternalSubnet(t, &tc.handlers, &tc.kube)
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

// Tests the Subnet "Observe" method
func testObserveSubnet(t *testing.T) {
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
				Managed: createCrossplaneSubnet(booleanCombSubnet[0], booleanCombSubnet[1], booleanCombSubnet[2], booleanCombSubnet[3], booleanCombSubnet[4],
					booleanCombSubnet[5], booleanCombSubnet[6], booleanCombSubnet[7], booleanCombSubnet[8], booleanCombSubnet[9],
					withSubnetStatus(), withExternalSubnetName()),
			},
			want: want{
				mg: createCrossplaneSubnet(booleanCombSubnet[0], booleanCombSubnet[1], booleanCombSubnet[2], booleanCombSubnet[3], booleanCombSubnet[4],
					booleanCombSubnet[5], booleanCombSubnet[6], booleanCombSubnet[7], booleanCombSubnet[8], booleanCombSubnet[9],
					withSubnetStatus(), withExternalSubnetName()),
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
				Managed: createCrossplaneSubnet(booleanCombSubnet[0], booleanCombSubnet[1], booleanCombSubnet[2], booleanCombSubnet[3], booleanCombSubnet[4],
					booleanCombSubnet[5], booleanCombSubnet[6], booleanCombSubnet[7], booleanCombSubnet[8], booleanCombSubnet[9],
					withSubnetStatus(), withExternalSubnetName()),
			},
			want: want{
				mg: createCrossplaneSubnet(booleanCombSubnet[0], booleanCombSubnet[1], booleanCombSubnet[2], booleanCombSubnet[3], booleanCombSubnet[4],
					booleanCombSubnet[5], booleanCombSubnet[6], booleanCombSubnet[7], booleanCombSubnet[8], booleanCombSubnet[9],
					withSubnetStatus(), withExternalSubnetName()),
				obs: managed.ExternalObservation{},
				errInfo: &errInfo{
					err:     errors.Wrap(errors.New(http.StatusText(http.StatusBadRequest)), errGetFailedSubnet),
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
				Managed: createCrossplaneSubnet(booleanCombSubnet[0], booleanCombSubnet[1], booleanCombSubnet[2], booleanCombSubnet[3], booleanCombSubnet[4],
					booleanCombSubnet[5], booleanCombSubnet[6], booleanCombSubnet[7], booleanCombSubnet[8], booleanCombSubnet[9],
					withSubnetStatus(), withExternalSubnetName()),
			},
			want: want{
				mg: createCrossplaneSubnet(booleanCombSubnet[0], booleanCombSubnet[1], booleanCombSubnet[2], booleanCombSubnet[3], booleanCombSubnet[4],
					booleanCombSubnet[5], booleanCombSubnet[6], booleanCombSubnet[7], booleanCombSubnet[8], booleanCombSubnet[9],
					withSubnetStatus(), withExternalSubnetName()),
				obs: managed.ExternalObservation{},
				errInfo: &errInfo{
					err:     errors.Wrap(errors.New(http.StatusText(http.StatusForbidden)), errGetFailedSubnet),
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

						sendBackCollection(w)
					},
				},
			},
			args: tstutil.Args{
				Managed: createCrossplaneSubnet(booleanCombSubnet[0], booleanCombSubnet[1], booleanCombSubnet[2], booleanCombSubnet[3], booleanCombSubnet[4],
					booleanCombSubnet[5], booleanCombSubnet[6], booleanCombSubnet[7], booleanCombSubnet[8], booleanCombSubnet[9],
					withSubnetStatus(), withExternalSubnetName()),
			},
			want: want{
				mg: createCrossplaneSubnet(booleanCombSubnet[0], booleanCombSubnet[1], booleanCombSubnet[2], booleanCombSubnet[3], booleanCombSubnet[4],
					booleanCombSubnet[5], booleanCombSubnet[6], booleanCombSubnet[7], booleanCombSubnet[8], booleanCombSubnet[9],
					withSubnetStatus(), withExternalSubnetName(), withSubnetSpecName(nil, false), withSubnetResourceGroup()),
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
			e, server, errCr := setupServerAndGetUnitTestExternalSubnet(t, &tc.handlers, &tc.kube)
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

// Tests the Subnet "Create" method several times (with various fields set to nil).
//
// The # of times/combinations is the value of variable 'numTests'
func TestCreateSubnet(t *testing.T) {
	for tstNum, booleanCombSubnet = range crossplaneClientUtil.GenerateSomePermutations(numTests, 46, true) {
		varCombinationLogging = crossplaneClientUtil.GetBinaryRep(tstNum, numTests)

		testCreateSubnet(t)
	}
}

// Tests the Subnet "Delete" method several times (with various fields set to nil).
//
// The # of times/combinations is the value of variable 'numTests'
func TestDeleteSubnet(t *testing.T) {
	for tstNum, booleanCombSubnet = range crossplaneClientUtil.GenerateSomePermutations(numTests, 46, true) {
		varCombinationLogging = crossplaneClientUtil.GetBinaryRep(tstNum, numTests)

		testDeleteSubnet(t)
	}
}

// Tests the Subnet "Observe" method several times (with various fields set to nil).
//
// The # of times/combinations is the value of variable 'numTests'
func TestObserveSubnet(t *testing.T) {
	for tstNum, booleanCombSubnet = range crossplaneClientUtil.GenerateSomePermutations(numTests, 46, true) {
		varCombinationLogging = crossplaneClientUtil.GetBinaryRep(tstNum, numTests)

		testObserveSubnet(t)
	}
}

func TestUpdateSubnet(t *testing.T) {
	booleanCombSubnet = crossplaneClientUtil.GenerateSomePermutations(1, 46, true)[0]

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

						if diff := cmp.Diff(http.MethodGet, r.Method); diff == "" {
							sendBackCollection(w)
						} else {
							if diff := cmp.Diff(http.MethodPatch, r.Method); diff != "" {
								t.Errorf("r: -want, +got:\n%s", diff)
							}

							// content type should always set before writeHeader()
							w.Header().Set("Content-Type", "application/json")
							w.WriteHeader(http.StatusOK)
						}
					},
				},
			},
			args: tstutil.Args{
				Managed: createCrossplaneSubnet(booleanCombSubnet[0], true, false, true, true, true, true, true, true, true,
					withSubnetSpecName(reference.ToPtrValue("new name"), false),
					withExternalSubnetName(),
					withSubnetStatus()),
			},
			want: want{
				mg: createCrossplaneSubnet(booleanCombSubnet[0], true, false, true, true, true, true, true, true, true,
					withSubnetSpecName(reference.ToPtrValue("new name"), false),
					withExternalSubnetName(),
					withSubnetStatus()),
				upd: managed.ExternalUpdate{},
			},
		},
		"Successful-2": {
			handlers: []tstutil.Handler{
				{
					Path: "/",
					HandlerFunc: func(w http.ResponseWriter, r *http.Request) {
						_ = r.Body.Close()

						if diff := cmp.Diff(http.MethodGet, r.Method); diff == "" {
							sendBackCollection(w)
						} else {
							if diff := cmp.Diff(http.MethodPatch, r.Method); diff != "" {
								t.Errorf("r: -want, +got:\n%s", diff)
							}

							// content type should always set before writeHeader()
							w.Header().Set("Content-Type", "application/json")
							w.WriteHeader(http.StatusOK)
						}
					},
				},
			},
			args: tstutil.Args{
				Managed: createCrossplaneSubnet(booleanCombSubnet[0], true, false, true, true, true, true, true, true, true,
					withSubnetSpecName(nil, true),
					withExternalSubnetName(),
					withSubnetStatus()),
			},
			want: want{
				mg: createCrossplaneSubnet(booleanCombSubnet[0], true, false, true, true, true, true, true, true, true,
					withSubnetSpecName(nil, true),
					withExternalSubnetName(),
					withSubnetStatus()),
				upd: managed.ExternalUpdate{},
			},
		},
		"Failed-1": {
			handlers: []tstutil.Handler{
				{
					Path: "/",
					HandlerFunc: func(w http.ResponseWriter, r *http.Request) {
						_ = r.Body.Close()

						if diff := cmp.Diff(http.MethodGet, r.Method); diff == "" {
							sendBackCollection(w)
						} else {
							if diff := cmp.Diff(http.MethodPatch, r.Method); diff != "" {
								t.Errorf("r: -want, +got:\n%s", diff)
							}

							// content type should always set before writeHeader()
							w.Header().Set("Content-Type", "application/json")
							w.WriteHeader(http.StatusBadRequest)
						}
					},
				},
			},
			args: tstutil.Args{
				Managed: createCrossplaneSubnet(booleanCombSubnet[0], true, false, true, true, true, true, true, true, true,
					withSubnetSpecName(nil, true),
					withExternalSubnetName(),
					withSubnetStatus()),
			},
			want: want{
				mg: createCrossplaneSubnet(booleanCombSubnet[0], true, false, true, true, true, true, true, true, true,
					withSubnetSpecName(nil, true),
					withExternalSubnetName(),
					withSubnetStatus()),
				upd: managed.ExternalUpdate{},
				err: errors.Wrap(errors.New(http.StatusText(http.StatusBadRequest)), errUpdateSubnet),
			},
		},
	}

	for name, tc := range cases {
		tc.name = name

		t.Run(name, func(t *testing.T) {
			e, server, errCr := setupServerAndGetUnitTestExternalSubnet(t, &tc.handlers, &tc.kube)
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
