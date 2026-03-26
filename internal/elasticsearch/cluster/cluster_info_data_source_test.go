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
	"fmt"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDataSourceClusterInfo(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceClusterInfoConfig,
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
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceClusterInfoConfig,
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
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceClusterInfoConfig,
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
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceClusterInfoConfig,
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
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceClusterInfoConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_info.test", "id"),
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_info.test", "cluster_uuid"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_info.test", "version.#", "1"),
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_info.test", "version.0.number"),
				),
			},
			// Second step with identical config: attributes must remain populated.
			{
				Config: testAccDataSourceClusterInfoConfig,
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
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceClusterInfoExplicitConnectionConfig(),
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
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceClusterInfoConfig,
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
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceClusterInfoConfig,
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

// testAccDataSourceClusterInfoExplicitConnectionConfig returns HCL that
// exercises the elasticsearch_connection block with an explicit endpoint list
// and insecure = true.  Auth credentials are drawn from the same env vars
// used by the acceptance-test harness.
func testAccDataSourceClusterInfoExplicitConnectionConfig() string {
	rawEndpoints := os.Getenv("ELASTICSEARCH_ENDPOINTS")
	apiKey := os.Getenv("ELASTICSEARCH_API_KEY")
	username := os.Getenv("ELASTICSEARCH_USERNAME")
	password := os.Getenv("ELASTICSEARCH_PASSWORD")

	parts := strings.Split(rawEndpoints, ",")
	quoted := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			quoted = append(quoted, fmt.Sprintf("%q", p))
		}
	}
	endpointList := strings.Join(quoted, ", ")

	var authLines string
	if apiKey != "" {
		authLines = fmt.Sprintf("    api_key = %q", apiKey)
	} else {
		authLines = fmt.Sprintf("    username = %q\n    password = %q", username, password)
	}

	return fmt.Sprintf(`
data "elasticstack_elasticsearch_info" "test_conn" {
  elasticsearch_connection {
    endpoints = [%s]
    insecure  = true
%s
  }
}
`, endpointList, authLines)
}

const testAccDataSourceClusterInfoConfig = `
provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_info" "test" {
}
`
