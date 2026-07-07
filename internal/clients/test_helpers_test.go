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

	"github.com/elastic/go-elasticsearch/v9"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/config"
	fleetclient "github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
	kibanaoapi "github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

const (
	testElasticsearchURL = "http://localhost:9200"
	testKibanaURL        = "http://localhost:5601"
	kibanaStatusPath     = "/api/status"
)

// newTestAPIClient returns a minimal *apiClient suitable for unit tests.
// Inner typed clients are built with the same helpers as production wiring so
// factory resolution returns usable accessors.
func newTestAPIClient(t *testing.T) *apiClient {
	t.Helper()

	cfg := config.Client{
		Elasticsearch: &elasticsearch.Config{
			Addresses: []string{testElasticsearchURL},
			Username:  "elastic",
			Password:  "changeme",
		},
		KibanaOapi: &kibanaoapi.Config{
			URL:      testKibanaURL,
			Username: "elastic",
			Password: "changeme",
		},
		Fleet: &fleetclient.Config{
			URL:      testKibanaURL,
			Username: "elastic",
			Password: "changeme",
		},
	}

	esClient, err := buildEsClient(cfg)
	require.NoError(t, err)
	require.NotNil(t, esClient)

	kibOapi, err := buildKibanaOapiClient(cfg)
	require.NoError(t, err)
	require.NotNil(t, kibOapi)

	fleet, err := buildFleetClient(cfg)
	require.NoError(t, err)
	require.NotNil(t, fleet)

	return &apiClient{
		elasticsearch:  esClient,
		kibanaOapi:     kibOapi,
		fleet:          fleet,
		version:        "unit-testing",
		kibanaEndpoint: testKibanaURL,
		fleetEndpoint:  testKibanaURL,
		esEndpoints:    []string{testElasticsearchURL},
	}
}

// newMockKibanaServer returns an httptest.Server that responds to GET /api/status
// with a minimal Kibana status payload for the given version string.
// The caller must call Close() on the returned server.
func newMockKibanaServer(version string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == kibanaStatusPath {
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
