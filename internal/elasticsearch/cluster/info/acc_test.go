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

package info_test

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
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
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_info.test", "tagline", "You Know, for Search"),
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
					// build_snapshot is a bool; Terraform encodes it as "true" or "false".
					// Use a regex match so the test is resilient to both release and snapshot builds.
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
	endpoint := primaryESEndpoint()
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
					checkAttrAbsent("data.elasticstack_elasticsearch_info.test_conn", "elasticsearch_connection.0.headers.%"),
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
	endpoints := connectionEndpoints()
	resource.Test(t, resource.TestCase{
		PreCheck: func() { preCheckESBasicAuth(t) },
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
					checkAttrAbsent("data.elasticstack_elasticsearch_info.test_conn", "elasticsearch_connection.0.api_key"),
					checkAttrAbsent("data.elasticstack_elasticsearch_info.test_conn", "elasticsearch_connection.0.bearer_token"),
					checkAttrAbsent("data.elasticstack_elasticsearch_info.test_conn", "elasticsearch_connection.0.es_client_authentication"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_info.test_conn", "elasticsearch_connection.0.endpoints.#", "2"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_info.test_conn", "elasticsearch_connection.0.endpoints.0", endpoints[0]),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_info.test_conn", "elasticsearch_connection.0.endpoints.1", endpoints[1]),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_info.test_conn", "elasticsearch_connection.0.headers.%", "2"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_info.test_conn", "elasticsearch_connection.0.headers.XTerraformTest", "basic-auth"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_info.test_conn", "elasticsearch_connection.0.headers.XTrace", "cluster-info"),
					checkAttrAbsent("data.elasticstack_elasticsearch_info.test_conn", "elasticsearch_connection.0.insecure"),
				),
			},
		},
	})
}

func TestAccDataSourceClusterInfo_withAPIKey(t *testing.T) {
	endpoint := primaryESEndpoint()
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
					checkAttrAbsent("data.elasticstack_elasticsearch_info.test_conn", "elasticsearch_connection.0.username"),
					checkAttrAbsent("data.elasticstack_elasticsearch_info.test_conn", "elasticsearch_connection.0.password"),
					checkAttrAbsent("data.elasticstack_elasticsearch_info.test_conn", "elasticsearch_connection.0.bearer_token"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_info.test_conn", "elasticsearch_connection.0.endpoints.#", "1"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_info.test_conn", "elasticsearch_connection.0.endpoints.0", endpoint),
				),
			},
		},
	})
}

func TestAccDataSourceClusterInfo_withBearerToken(t *testing.T) {
	preCheckESBasicAuth(t)

	endpoint := primaryESEndpoint()
	bearerToken := acctest.CreateESAccessToken(t)
	resource.Test(t, resource.TestCase{
		PreCheck: func() { preCheckESBasicAuth(t) },
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
					checkAttrAbsent("data.elasticstack_elasticsearch_info.test_conn", "elasticsearch_connection.0.username"),
					checkAttrAbsent("data.elasticstack_elasticsearch_info.test_conn", "elasticsearch_connection.0.password"),
					checkAttrAbsent("data.elasticstack_elasticsearch_info.test_conn", "elasticsearch_connection.0.api_key"),
				),
			},
		},
	})
}

func TestAccDataSourceClusterInfo_withTLSInputs(t *testing.T) {
	endpoint := primaryESEndpoint()
	tlsMaterial := acctest.CreateTLSMaterial(t, "cluster-info-test")
	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("inline"),
				ConfigVariables: config.Variables{
					"ca_data":   config.StringVariable(tlsMaterial.CAPEM),
					"cert_data": config.StringVariable(tlsMaterial.CertPEM),
					"endpoint":  config.StringVariable(endpoint),
					"key_data":  config.StringVariable(tlsMaterial.KeyPEM),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_info.test_conn", "elasticsearch_connection.#", "1"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_info.test_conn", "elasticsearch_connection.0.ca_data", tlsMaterial.CAPEM),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_info.test_conn", "elasticsearch_connection.0.cert_data", tlsMaterial.CertPEM),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_info.test_conn", "elasticsearch_connection.0.key_data", tlsMaterial.KeyPEM),
					checkAttrAbsent("data.elasticstack_elasticsearch_info.test_conn", "elasticsearch_connection.0.ca_file"),
					checkAttrAbsent("data.elasticstack_elasticsearch_info.test_conn", "elasticsearch_connection.0.cert_file"),
					checkAttrAbsent("data.elasticstack_elasticsearch_info.test_conn", "elasticsearch_connection.0.key_file"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("file"),
				ConfigVariables: config.Variables{
					"ca_file":   config.StringVariable(tlsMaterial.CAFile),
					"cert_file": config.StringVariable(tlsMaterial.CertFile),
					"endpoint":  config.StringVariable(endpoint),
					"key_file":  config.StringVariable(tlsMaterial.KeyFile),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_info.test_conn", "elasticsearch_connection.#", "1"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_info.test_conn", "elasticsearch_connection.0.ca_file", tlsMaterial.CAFile),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_info.test_conn", "elasticsearch_connection.0.cert_file", tlsMaterial.CertFile),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_info.test_conn", "elasticsearch_connection.0.key_file", tlsMaterial.KeyFile),
					checkAttrAbsent("data.elasticstack_elasticsearch_info.test_conn", "elasticsearch_connection.0.ca_data"),
					checkAttrAbsent("data.elasticstack_elasticsearch_info.test_conn", "elasticsearch_connection.0.cert_data"),
					checkAttrAbsent("data.elasticstack_elasticsearch_info.test_conn", "elasticsearch_connection.0.key_data"),
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
	const standardUUIDPattern = `[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}`
	const base64UUIDPattern = `[A-Za-z0-9_-]{20,}`
	uuidRe := regexp.MustCompile(`^(` + standardUUIDPattern + `|` + base64UUIDPattern + `)$`)
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
// for the version sub-fields that were previously only asserted as "set".
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

// checkAttrAbsent asserts that the given attribute is either absent from state
// or present with an empty string value. PF blocks only write attributes that
// were explicitly configured, so unset optional fields will not appear in state.
func checkAttrAbsent(resourceName, attrName string) resource.TestCheckFunc { //nolint:unparam // resourceName is a parameter for API flexibility; all callers currently use the same resource
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found in state: %s", resourceName)
		}
		value, ok := rs.Primary.Attributes[attrName]
		if ok && value != "" {
			return fmt.Errorf("expected %s to be absent or empty, got %q", attrName, value)
		}
		return nil
	}
}

func primaryESEndpoint() string {
	for ep := range strings.SplitSeq(os.Getenv("ELASTICSEARCH_ENDPOINTS"), ",") {
		ep = strings.TrimSpace(ep)
		if ep != "" {
			return ep
		}
	}
	return "http://localhost:9200"
}

func connectionEndpoints() []string {
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

func preCheckESBasicAuth(t *testing.T) {
	t.Helper()
	acctest.PreCheck(t)
	if os.Getenv("ELASTICSEARCH_USERNAME") == "" || os.Getenv("ELASTICSEARCH_PASSWORD") == "" {
		t.Skip("ELASTICSEARCH_USERNAME and ELASTICSEARCH_PASSWORD must be set for explicit basic auth coverage")
	}
}
