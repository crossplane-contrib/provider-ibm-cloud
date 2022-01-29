/*
Copyright 2022 The Crossplane Authors.

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

package tstutil

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/IBM/go-sdk-core/core"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	ibmc "github.com/crossplane-contrib/provider-ibm-cloud/pkg/clients"
)

// Args is used in testing, to hold the arguments passed to the crossplane functions
type Args struct {
	Managed resource.Managed
}

// Handler is used in unit testing
type Handler struct {
	// http path to handle a request
	Path string

	// Function that will deal with the request, and write the response in the writer.
	HandlerFunc func(w http.ResponseWriter, r *http.Request)
}

// SetupTestServerClient sets up a unit test http server, and creates a client to be used in unit testing.
//
// Params
//	   testingObj - the test object
//	   handlers - the handlers that create the responses
//
// Returns
//		- a configured "provider" client (ready to talk to the IBM cloud). nil of there is a problem
//		- the (started) test http server, on which the caller should call 'defer ....Close()' (reason for this is we need to keep it around to prevent
//		  garbage collection)
//      - an error (if there is one)
func SetupTestServerClient(testingObj *testing.T, handlers *[]Handler) (*ibmc.ClientSession, *httptest.Server, error) {
	mux := http.NewServeMux()
	for _, h := range *handlers {
		mux.HandleFunc(h.Path, h.HandlerFunc)
	}

	tstServer := httptest.NewServer(mux)

	mClient, errNC := ibmc.NewClient(
		ibmc.ClientOptions{
			URL: tstServer.URL,
			Authenticator: &core.BearerTokenAuthenticator{
				BearerToken: ibmc.FakeBearerToken,
			},

			BearerToken:  ibmc.FakeBearerToken,
			RefreshToken: "does format matter?",
		})

	if errNC != nil {
		return nil, tstServer, errNC
	}

	return &mClient, tstServer, nil
}
