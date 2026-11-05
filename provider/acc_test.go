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

package provider_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"slices"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/config"
	tfconfig "github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestKibanaConfiguration_withHeaders(t *testing.T) {
	acctest.SkipIfNotAcceptanceTest(t)

	upstreamEndpoint := strings.TrimSpace(os.Getenv("KIBANA_ENDPOINT"))
	if upstreamEndpoint == "" {
		t.Fatal("KIBANA_ENDPOINT must be set")
	}

	const (
		headerName  = "X-Elasticstack-Provider-Acc-Test"
		headerValue = "provider-header-value"
	)
	proxy := acctest.NewRecordingProxy(t, upstreamEndpoint)
	t.Setenv("KIBANA_ENDPOINT", proxy.URL())
	t.Setenv(config.PreferConfiguredKibanaEndpointEnvVar, "true")

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				ConfigVariables: tfconfig.Variables{
					"kibana_endpoint": tfconfig.StringVariable(proxy.URL()),
					"header_name":     tfconfig.StringVariable(headerName),
					"header_value":    tfconfig.StringVariable(headerValue),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.elasticstack_kibana_spaces.all_spaces", "id", "spaces"),
					func(_ *terraform.State) error {
						if slices.Contains(proxy.HeaderValues(headerName), headerValue) {
							return nil
						}
						return fmt.Errorf("recording proxy received %d requests but no %s header with value %q", proxy.RequestCount(), headerName, headerValue)
					},
				),
			},
		},
	})
}

func TestFleetConfiguration_excludesKibanaHeaders(t *testing.T) {
	acctest.SkipIfNotAcceptanceTest(t)

	receivedHeaders := make(chan http.Header, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/fleet/enrollment_api_keys" {
			http.NotFound(w, r)
			return
		}
		select {
		case receivedHeaders <- r.Header.Clone():
		default:
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"items":[],"list":[],"page":1,"perPage":20,"total":0}`))
	}))
	t.Cleanup(server.Close)
	t.Setenv(config.PreferConfiguredKibanaEndpointEnvVar, "true")
	t.Setenv("FLEET_ENDPOINT", "")
	if err := os.Unsetenv("FLEET_ENDPOINT"); err != nil {
		t.Fatal(err)
	}

	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				ConfigVariables: tfconfig.Variables{
					"kibana_endpoint": tfconfig.StringVariable(server.URL),
					"fleet_endpoint":  tfconfig.StringVariable(server.URL),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_fleet_enrollment_tokens.test", "id"),
					func(_ *terraform.State) error {
						select {
						case headers := <-receivedHeaders:
							if value := headers.Get("cf-access-token"); value != "" {
								return fmt.Errorf("Fleet endpoint received cf-access-token header %q, want no header", value)
							}
							return nil
						default:
							return fmt.Errorf("Fleet endpoint received no request")
						}
					},
				),
			},
		},
	})
}

func TestKibanaConfiguration_rejectsAuthorizationHeaderBeforeRequest(t *testing.T) {
	testKibanaConfigurationInvalidHeaderSendsNoRequest(t, "Authorization", "custom", `Authorization.*reserved header name`)
}

func TestKibanaConfiguration_rejectsKbnXsrfHeaderRegardlessOfCaseBeforeRequest(t *testing.T) {
	testKibanaConfigurationInvalidHeaderSendsNoRequest(t, "KBN-XSRF", "false", `KBN-XSRF.*reserved header name`)
}

func TestKibanaConfiguration_rejectsHostHeaderBeforeRequest(t *testing.T) {
	testKibanaConfigurationInvalidHeaderSendsNoRequest(t, "Host", "kibana.example.com", `Host.*reserved header name`)
}

func TestKibanaConfiguration_rejectsContentLengthHeaderBeforeRequest(t *testing.T) {
	testKibanaConfigurationInvalidHeaderSendsNoRequest(t, "Content-Length", "42", `Content-Length.*reserved header name`)
}

func TestKibanaConfiguration_rejectsTransferEncodingHeaderBeforeRequest(t *testing.T) {
	testKibanaConfigurationInvalidHeaderSendsNoRequest(t, "Transfer-Encoding", "chunked", `Transfer-Encoding.*reserved header name`)
}

func TestKibanaConfiguration_rejectsTrailerHeaderBeforeRequest(t *testing.T) {
	testKibanaConfigurationInvalidHeaderSendsNoRequest(t, "Trailer", "X-Checksum", `Trailer.*reserved header name`)
}

func testKibanaConfigurationInvalidHeaderSendsNoRequest(t *testing.T, headerName, headerValue, diagnosticPattern string) {
	t.Helper()
	acctest.SkipIfNotAcceptanceTest(t)

	var requestCount atomic.Int64
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		requestCount.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(server.Close)
	t.Setenv("KIBANA_ENDPOINT", server.URL)

	if endpoint := strings.TrimSpace(os.Getenv("KIBANA_ENDPOINT")); endpoint != server.URL {
		t.Fatalf("KIBANA_ENDPOINT = %q, want observed server %q", endpoint, server.URL)
	}

	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				ConfigVariables: tfconfig.Variables{
					"kibana_endpoint": tfconfig.StringVariable(server.URL),
					"header_name":     tfconfig.StringVariable(headerName),
					"header_value":    tfconfig.StringVariable(headerValue),
				},
				ExpectError: regexp.MustCompile(`(?s)` + diagnosticPattern),
			},
		},
	})

	if got := requestCount.Load(); got != 0 {
		t.Fatalf("Kibana received %d requests, want 0", got)
	}
}
