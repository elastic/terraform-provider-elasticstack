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

package cluster_test

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"io"
	"math/big"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDataSourceClusterInfo(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_info.test", "tagline", "You Know, for Search"),
				),
			},
		},
	})
}

// TestAccDataSourceClusterInfo_topLevelAttributes verifies that the top-level
// metadata attributes (cluster_uuid, cluster_name, name) are populated and
// that the resource id is set to cluster_uuid.
func TestAccDataSourceClusterInfo_topLevelAttributes(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_info.test", "id"),
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_info.test", "cluster_uuid"),
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_info.test", "cluster_name"),
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_info.test", "name"),
					// id must equal cluster_uuid
					resource.TestCheckResourceAttrPair(
						"data.elasticstack_elasticsearch_info.test", "id",
						"data.elasticstack_elasticsearch_info.test", "cluster_uuid",
					),
				),
			},
		},
	})
}

// TestAccDataSourceClusterInfo_versionBlock verifies that the version block is
// present (exactly one element) and that all nested fields are populated.
func TestAccDataSourceClusterInfo_versionBlock(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_info.test", "version.#", "1"),
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_info.test", "version.0.number"),
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_info.test", "version.0.build_date"),
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_info.test", "version.0.build_flavor"),
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_info.test", "version.0.build_hash"),
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_info.test", "version.0.build_type"),
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_info.test", "version.0.lucene_version"),
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_info.test", "version.0.minimum_index_compatibility_version"),
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_info.test", "version.0.minimum_wire_compatibility_version"),
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_info.test", "version.0.build_snapshot"),
				),
			},
		},
	})
}

// TestAccDataSourceClusterInfo_versionFieldFormats verifies that key version
// fields match expected formats: semantic version numbers for number and
// compatibility fields, and a boolean string for build_snapshot.
func TestAccDataSourceClusterInfo_versionFieldFormats(t *testing.T) {
	// Matches x.y.z and x.y.z-TAG (e.g. 8.14.0-SNAPSHOT), anchored at both ends.
	semverRe := regexp.MustCompile(`^\d+\.\d+\.\d+(-[a-zA-Z0-9.]+)?$`)
	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr("data.elasticstack_elasticsearch_info.test", "version.0.number", semverRe),
					resource.TestMatchResourceAttr("data.elasticstack_elasticsearch_info.test", "version.0.minimum_index_compatibility_version", semverRe),
					resource.TestMatchResourceAttr("data.elasticstack_elasticsearch_info.test", "version.0.minimum_wire_compatibility_version", semverRe),
					// build_snapshot is a bool; Terraform encodes it as "true" or "false"
					resource.TestMatchResourceAttr("data.elasticstack_elasticsearch_info.test", "version.0.build_snapshot", regexp.MustCompile(`^(true|false)$`)),
				),
			},
		},
	})
}

// TestAccDataSourceClusterInfo_refreshStability verifies that a second plan
// step (with no configuration changes) returns consistent, non-empty values
// for the key identity and version attributes.
func TestAccDataSourceClusterInfo_refreshStability(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_info.test", "id"),
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_info.test", "cluster_uuid"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_info.test", "version.#", "1"),
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_info.test", "version.0.number"),
				),
			},
			// Second step with identical config: attributes must remain populated.
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read_again"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_info.test", "id"),
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_info.test", "cluster_uuid"),
					resource.TestCheckResourceAttrPair(
						"data.elasticstack_elasticsearch_info.test", "id",
						"data.elasticstack_elasticsearch_info.test", "cluster_uuid",
					),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_info.test", "version.#", "1"),
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_info.test", "version.0.number"),
				),
			},
		},
	})
}

// TestAccDataSourceClusterInfo_withExplicitConnection verifies that the data
// source works correctly when an explicit elasticsearch_connection block is
// provided with endpoints and insecure = true.  The test also confirms that
// the connection block attributes are reflected back in the state.
func TestAccDataSourceClusterInfo_withExplicitConnection(t *testing.T) {
	endpoint := clusterInfoPrimaryESEndpoint()
	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				ConfigVariables: config.Variables{
					"endpoints": config.ListVariable(config.StringVariable(endpoint)),
					"api_key":   config.StringVariable(os.Getenv("ELASTICSEARCH_API_KEY")),
					"username":  config.StringVariable(os.Getenv("ELASTICSEARCH_USERNAME")),
					"password":  config.StringVariable(os.Getenv("ELASTICSEARCH_PASSWORD")),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_info.test_conn", "id"),
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_info.test_conn", "cluster_uuid"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_info.test_conn", "tagline", "You Know, for Search"),
					// Connection block attributes must be stored in state.
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_info.test_conn", "elasticsearch_connection.#", "1"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_info.test_conn", "elasticsearch_connection.0.endpoints.#", "1"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_info.test_conn", "elasticsearch_connection.0.endpoints.0", endpoint),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_info.test_conn", "elasticsearch_connection.0.insecure", "true"),
				),
			},
		},
	})
}

func TestAccDataSourceClusterInfo_withoutExplicitConnection(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_info.test", "elasticsearch_connection.#", "0"),
				),
			},
		},
	})
}

func TestAccDataSourceClusterInfo_withBasicAuthHeadersAndMultiEndpoints(t *testing.T) {
	endpoints := clusterInfoConnectionEndpoints()
	resource.Test(t, resource.TestCase{
		PreCheck: func() { preCheckClusterInfoESBasicAuth(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				ConfigVariables: config.Variables{
					"endpoints": config.ListVariable(
						config.StringVariable(endpoints[0]),
						config.StringVariable(endpoints[1]),
					),
					"headers": config.MapVariable(map[string]config.Variable{
						"XTerraformTest": config.StringVariable("basic-auth"),
						"XTrace":         config.StringVariable("cluster-info"),
					}),
					"password": config.StringVariable(os.Getenv("ELASTICSEARCH_PASSWORD")),
					"username": config.StringVariable(os.Getenv("ELASTICSEARCH_USERNAME")),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_info.test_conn", "id"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_info.test_conn", "elasticsearch_connection.#", "1"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_info.test_conn", "elasticsearch_connection.0.username", os.Getenv("ELASTICSEARCH_USERNAME")),
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_info.test_conn", "elasticsearch_connection.0.password"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_info.test_conn", "elasticsearch_connection.0.api_key", ""),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_info.test_conn", "elasticsearch_connection.0.bearer_token", ""),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_info.test_conn", "elasticsearch_connection.0.es_client_authentication", ""),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_info.test_conn", "elasticsearch_connection.0.endpoints.#", "2"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_info.test_conn", "elasticsearch_connection.0.endpoints.0", endpoints[0]),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_info.test_conn", "elasticsearch_connection.0.endpoints.1", endpoints[1]),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_info.test_conn", "elasticsearch_connection.0.headers.%", "2"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_info.test_conn", "elasticsearch_connection.0.headers.XTerraformTest", "basic-auth"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_info.test_conn", "elasticsearch_connection.0.headers.XTrace", "cluster-info"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_info.test_conn", "elasticsearch_connection.0.insecure", "false"),
				),
			},
		},
	})
}

func TestAccDataSourceClusterInfo_withAPIKey(t *testing.T) {
	endpoint := clusterInfoPrimaryESEndpoint()
	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				ConfigVariables: config.Variables{
					"endpoint": config.StringVariable(endpoint),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_info.test_conn", "elasticsearch_connection.#", "1"),
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_info.test_conn", "elasticsearch_connection.0.api_key"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_info.test_conn", "elasticsearch_connection.0.username", ""),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_info.test_conn", "elasticsearch_connection.0.password", ""),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_info.test_conn", "elasticsearch_connection.0.bearer_token", ""),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_info.test_conn", "elasticsearch_connection.0.endpoints.#", "1"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_info.test_conn", "elasticsearch_connection.0.endpoints.0", endpoint),
				),
			},
		},
	})
}

func TestAccDataSourceClusterInfo_withBearerToken(t *testing.T) {
	endpoint := clusterInfoPrimaryESEndpoint()
	var bearerToken string
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			preCheckClusterInfoESBasicAuth(t)
			bearerToken = createClusterInfoESAccessToken(t)
		},
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				ConfigVariables: config.Variables{
					"bearer_token": config.StringVariable(bearerToken),
					"endpoint":     config.StringVariable(endpoint),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_info.test_conn", "elasticsearch_connection.#", "1"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_info.test_conn", "elasticsearch_connection.0.bearer_token", bearerToken),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_info.test_conn", "elasticsearch_connection.0.es_client_authentication", "Authorization"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_info.test_conn", "elasticsearch_connection.0.username", ""),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_info.test_conn", "elasticsearch_connection.0.password", ""),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_info.test_conn", "elasticsearch_connection.0.api_key", ""),
				),
			},
		},
	})
}

func TestAccDataSourceClusterInfo_withTLSInputs(t *testing.T) {
	endpoint := clusterInfoPrimaryESEndpoint()
	tlsMaterial := createClusterInfoTLSMaterial(t)
	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("inline"),
				ConfigVariables: config.Variables{
					"ca_data":   config.StringVariable(tlsMaterial.caPEM),
					"cert_data": config.StringVariable(tlsMaterial.certPEM),
					"endpoint":  config.StringVariable(endpoint),
					"key_data":  config.StringVariable(tlsMaterial.keyPEM),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_info.test_conn", "elasticsearch_connection.#", "1"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_info.test_conn", "elasticsearch_connection.0.ca_data", tlsMaterial.caPEM),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_info.test_conn", "elasticsearch_connection.0.cert_data", tlsMaterial.certPEM),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_info.test_conn", "elasticsearch_connection.0.key_data", tlsMaterial.keyPEM),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_info.test_conn", "elasticsearch_connection.0.ca_file", ""),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_info.test_conn", "elasticsearch_connection.0.cert_file", ""),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_info.test_conn", "elasticsearch_connection.0.key_file", ""),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("file"),
				ConfigVariables: config.Variables{
					"ca_file":   config.StringVariable(tlsMaterial.caFile),
					"cert_file": config.StringVariable(tlsMaterial.certFile),
					"endpoint":  config.StringVariable(endpoint),
					"key_file":  config.StringVariable(tlsMaterial.keyFile),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_info.test_conn", "elasticsearch_connection.#", "1"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_info.test_conn", "elasticsearch_connection.0.ca_file", tlsMaterial.caFile),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_info.test_conn", "elasticsearch_connection.0.cert_file", tlsMaterial.certFile),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_info.test_conn", "elasticsearch_connection.0.key_file", tlsMaterial.keyFile),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_info.test_conn", "elasticsearch_connection.0.ca_data", ""),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_info.test_conn", "elasticsearch_connection.0.cert_data", ""),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_info.test_conn", "elasticsearch_connection.0.key_data", ""),
				),
			},
		},
	})
}

// TestAccDataSourceClusterInfo_clusterUUIDAndNameFormats upgrades the
// set-only assertions for cluster_uuid, cluster_name and name to regex
// checks so we catch obviously-wrong values (e.g. empty string).
// Note: Elasticsearch returns cluster_uuid either in the standard hyphenated
// UUID format (8-4-4-4-12 hex) or as a URL-safe base64-encoded identifier
// (≥20 URL-safe base64 chars).  Both variants are accepted by the regex.
func TestAccDataSourceClusterInfo_clusterUUIDAndNameFormats(t *testing.T) {
	// standardUUIDPattern matches the canonical hyphenated 8-4-4-4-12 hex UUID.
	const standardUUIDPattern = `[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}`
	// base64UUIDPattern matches Elasticsearch's compact URL-safe base64 cluster ID (≥20 chars).
	const base64UUIDPattern = `[A-Za-z0-9_-]{20,}`
	uuidRe := regexp.MustCompile(`^(` + standardUUIDPattern + `|` + base64UUIDPattern + `)$`)
	// cluster_name and node name must be at least one non-whitespace character.
	nonEmptyRe := regexp.MustCompile(`\S`)
	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr("data.elasticstack_elasticsearch_info.test", "cluster_uuid", uuidRe),
					resource.TestMatchResourceAttr("data.elasticstack_elasticsearch_info.test", "cluster_name", nonEmptyRe),
					resource.TestMatchResourceAttr("data.elasticstack_elasticsearch_info.test", "name", nonEmptyRe),
					// id must still equal cluster_uuid (regression guard).
					resource.TestCheckResourceAttrPair(
						"data.elasticstack_elasticsearch_info.test", "id",
						"data.elasticstack_elasticsearch_info.test", "cluster_uuid",
					),
				),
			},
		},
	})
}

// TestAccDataSourceClusterInfo_versionBuildFormats adds format-level checks
// for the version sub-fields that were previously only asserted as "set":
// build_hash (hex git ref), build_date (date-prefixed string), build_flavor,
// build_type, and lucene_version (semver).
// Note: build_date is stored via Go's time.Time.String() which produces the
// format "YYYY-MM-DD HH:MM:SS.NNNNNNNNN +0000 UTC" rather than strict ISO-8601.
func TestAccDataSourceClusterInfo_versionBuildFormats(t *testing.T) {
	buildHashRe := regexp.MustCompile(`^[0-9a-f]{7,40}$`)
	// Matches both "YYYY-MM-DDTHH:MM:SS" (ISO-8601) and "YYYY-MM-DD HH:MM:SS" (Go time.String).
	buildDateRe := regexp.MustCompile(`^\d{4}-\d{2}-\d{2}[T ]`)
	buildFlavorRe := regexp.MustCompile(`^(default|oss|serverless)$`)
	buildTypeRe := regexp.MustCompile(`^(tar|zip|docker|rpm|deb|pkg)$`)
	luceneVersionRe := regexp.MustCompile(`^\d+\.\d+\.\d+`)
	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr("data.elasticstack_elasticsearch_info.test", "version.0.build_hash", buildHashRe),
					resource.TestMatchResourceAttr("data.elasticstack_elasticsearch_info.test", "version.0.build_date", buildDateRe),
					resource.TestMatchResourceAttr("data.elasticstack_elasticsearch_info.test", "version.0.build_flavor", buildFlavorRe),
					resource.TestMatchResourceAttr("data.elasticstack_elasticsearch_info.test", "version.0.build_type", buildTypeRe),
					resource.TestMatchResourceAttr("data.elasticstack_elasticsearch_info.test", "version.0.lucene_version", luceneVersionRe),
				),
			},
		},
	})
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// clusterInfoPrimaryESEndpoint returns the first endpoint from the
// ELASTICSEARCH_ENDPOINTS env var, falling back to http://localhost:9200.
func clusterInfoPrimaryESEndpoint() string {
	for ep := range strings.SplitSeq(os.Getenv("ELASTICSEARCH_ENDPOINTS"), ",") {
		ep = strings.TrimSpace(ep)
		if ep != "" {
			return ep
		}
	}
	return "http://localhost:9200"
}

func clusterInfoConnectionEndpoints() []string {
	endpoints := make([]string, 0, 2)
	for endpoint := range strings.SplitSeq(os.Getenv("ELASTICSEARCH_ENDPOINTS"), ",") {
		endpoint = strings.TrimSpace(endpoint)
		if endpoint != "" {
			endpoints = append(endpoints, endpoint)
		}
	}

	if len(endpoints) == 0 {
		endpoints = append(endpoints, "http://localhost:9200")
	}
	if len(endpoints) == 1 {
		endpoints = append(endpoints, endpoints[0])
	}

	return endpoints[:2]
}

func preCheckClusterInfoESBasicAuth(t *testing.T) {
	t.Helper()
	acctest.PreCheck(t)
	if os.Getenv("ELASTICSEARCH_USERNAME") == "" || os.Getenv("ELASTICSEARCH_PASSWORD") == "" {
		t.Skip("ELASTICSEARCH_USERNAME and ELASTICSEARCH_PASSWORD must be set for explicit basic auth coverage")
	}
}

func createClusterInfoESAccessToken(t *testing.T) string {
	t.Helper()

	client, err := clients.NewAcceptanceTestingClient()
	if err != nil {
		t.Fatalf("failed to create acceptance testing client: %v", err)
	}
	esClient, err := client.GetESClient()
	if err != nil {
		t.Fatalf("failed to get Elasticsearch client: %v", err)
	}

	payload, err := json.Marshal(map[string]string{
		"grant_type": "password",
		"username":   os.Getenv("ELASTICSEARCH_USERNAME"),
		"password":   os.Getenv("ELASTICSEARCH_PASSWORD"),
	})
	if err != nil {
		t.Fatalf("failed to marshal token request: %v", err)
	}

	resp, err := esClient.Security.GetToken(
		bytes.NewReader(payload),
		esClient.Security.GetToken.WithContext(context.Background()),
	)
	if err != nil {
		t.Fatalf("failed to create Elasticsearch access token: %v", err)
	}
	defer resp.Body.Close()

	if resp.IsError() {
		body, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			t.Fatalf("failed to create Elasticsearch access token: status %d (additionally failed to read error response: %v)", resp.StatusCode, readErr)
		}
		t.Fatalf("failed to create Elasticsearch access token: status %d: %s", resp.StatusCode, string(body))
	}

	var tokenResponse struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tokenResponse); err != nil {
		t.Fatalf("failed to decode token response: %v", err)
	}
	if tokenResponse.AccessToken == "" {
		t.Fatalf("token response did not include an access_token")
	}

	return tokenResponse.AccessToken
}

type clusterInfoTLSMaterial struct {
	caPEM    string
	certPEM  string
	keyPEM   string
	caFile   string
	certFile string
	keyFile  string
}

func createClusterInfoTLSMaterial(t *testing.T) clusterInfoTLSMaterial {
	t.Helper()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("failed to generate private key: %v", err)
	}

	certificateDER, err := x509.CreateCertificate(rand.Reader, &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName: "cluster-info-test",
		},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IsCA:                  true,
	}, &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName: "cluster-info-test",
		},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IsCA:                  true,
	}, &privateKey.PublicKey, privateKey)
	if err != nil {
		t.Fatalf("failed to generate certificate: %v", err)
	}

	certPEM := string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certificateDER}))
	keyPEM := string(pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)}))

	tempDir := t.TempDir()
	caFile := filepath.Join(tempDir, "ca.pem")
	certFile := filepath.Join(tempDir, "cert.pem")
	keyFile := filepath.Join(tempDir, "key.pem")

	for path, contents := range map[string]string{
		caFile:   certPEM,
		certFile: certPEM,
		keyFile:  keyPEM,
	} {
		if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
			t.Fatalf("failed to write TLS test file %s: %v", path, err)
		}
	}

	return clusterInfoTLSMaterial{
		caPEM:    certPEM,
		certPEM:  certPEM,
		keyPEM:   keyPEM,
		caFile:   caFile,
		certFile: certFile,
		keyFile:  keyFile,
	}
}
