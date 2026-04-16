// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package clients

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/disaster37/go-kibana-rest/v8"
	kibanaoapi "github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

// newTestAPIClient returns a minimal *apiClient suitable for unit tests.
// It sets up Kibana clients with a fake endpoint so that actual network
// calls are never made.
func newTestAPIClient(t *testing.T) *apiClient {
	t.Helper()
	kib, err := kibana.NewClient(kibana.Config{
		Address:  "http://localhost:5601",
		Username: "elastic",
		Password: "changeme",
	})
	require.NoError(t, err)

	kibOapi, err := kibanaoapi.NewClient(kibanaoapi.Config{
		URL:      "http://localhost:5601",
		Username: "elastic",
		Password: "changeme",
	})
	require.NoError(t, err)

	return &apiClient{
		kibana:       kib,
		kibanaOapi:   kibOapi,
		kibanaConfig: kibana.Config{Address: "http://localhost:5601", Username: "elastic", Password: "changeme"},
		version:      "unit-testing",
	}
}

// newMockKibanaServer returns an httptest.Server that responds to GET /api/status
// with a minimal Kibana status payload for the given version string.
// The caller must call Close() on the returned server.
func newMockKibanaServer(version string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == "/api/status" {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintf(w, `{"version":{"number":%q,"build_flavor":"default"}}`, version)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
}

// kibanaConnectionAttrTypes returns the attribute type map for
// config.KibanaConnection so we can build framework type values in tests.
func kibanaConnectionAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"username":     types.StringType,
		"password":     types.StringType,
		"api_key":      types.StringType,
		"bearer_token": types.StringType,
		"endpoints":    types.ListType{ElemType: types.StringType},
		"ca_certs":     types.ListType{ElemType: types.StringType},
		"insecure":     types.BoolType,
	}
}
