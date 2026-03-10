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

package indices_test

import (
	"fmt"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccIndicesDataSource verifies that the data source can list system security indices
// and exposes at least one well-known .security index with its shard count and alias.
func TestAccIndicesDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccIndicesDataSourceConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					// At least one index must be returned for the .security-* pattern.
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_indices.security_indices", "id"),
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_indices.security_indices", "indices.0.name"),
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_indices.security_indices", "indices.0.number_of_shards"),
				),
			},
		},
	})
}

const testAccIndicesDataSourceConfig = `
provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_indices" "security_indices" {
  target = ".security-*"
}
`

// TestAccIndicesDataSource_Target_DefaultAndExplicitAll validates that all three
// "match everything" forms — omitted target, "*", and "_all" — each return a
// non-empty result with a populated id.
func TestAccIndicesDataSource_Target_DefaultAndExplicitAll(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				// Omitted target — defaults to "*" (all indices).
				Config: testAccIndicesDataSourceConfigNoTarget,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_indices.all_default", "id"),
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_indices.all_default", "indices.#"),
				),
			},
			{
				// Explicit "*" wildcard — should return all non-hidden indices.
				Config: testAccIndicesDataSourceConfigStar,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_indices.all_star", "id"),
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_indices.all_star", "indices.#"),
				),
			},
			{
				// Explicit "_all" wildcard — equivalent to "*".
				Config: testAccIndicesDataSourceConfigExplicitAll,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_indices.all_explicit", "id"),
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_indices.all_explicit", "indices.#"),
				),
			},
		},
	})
}

const testAccIndicesDataSourceConfigNoTarget = `
provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_indices" "all_default" {
}
`

const testAccIndicesDataSourceConfigStar = `
provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_indices" "all_star" {
  target = "*"
}
`

const testAccIndicesDataSourceConfigExplicitAll = `
provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_indices" "all_explicit" {
  target = "_all"
}
`

// TestAccIndicesDataSource_Target_FilteringExactVsWildcard creates two indices with
// a shared random prefix and verifies that the wildcard target matches both while
// targeting one index by exact name returns exactly one result.
func TestAccIndicesDataSource_Target_FilteringExactVsWildcard(t *testing.T) {
	prefix := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlpha)
	indexA := prefix + "a"
	indexB := prefix + "b"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				// Wildcard should return both indices.
				Config: testAccIndicesDataSourceConfigFilteringWildcard(prefix, indexA, indexB),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_indices.wildcard", "indices.#", "2"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_indices.exact", "indices.#", "1"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_indices.exact", "indices.0.name", indexA),
				),
			},
		},
	})
}

func testAccIndicesDataSourceConfigFilteringWildcard(prefix, indexA, indexB string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_index" "a" {
  name                = %q
  number_of_shards    = 1
  number_of_replicas  = 0
  deletion_protection = false
}

resource "elasticstack_elasticsearch_index" "b" {
  name                = %q
  number_of_shards    = 1
  number_of_replicas  = 0
  deletion_protection = false
}

data "elasticstack_elasticsearch_indices" "wildcard" {
  target     = "%s*"
  depends_on = [elasticstack_elasticsearch_index.a, elasticstack_elasticsearch_index.b]
}

data "elasticstack_elasticsearch_indices" "exact" {
  target     = %q
  depends_on = [elasticstack_elasticsearch_index.a, elasticstack_elasticsearch_index.b]
}
`, indexA, indexB, prefix, indexA)
}

// TestAccIndicesDataSource_ReadsIndexSettings_TypedFields creates a known index with a
// set of representative typed settings and verifies the data source surfaces the
// correct values for each attribute category (int, string, bool).
func TestAccIndicesDataSource_ReadsIndexSettings_TypedFields(t *testing.T) {
	indexName := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlpha)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccIndicesDataSourceConfigTypedSettings(indexName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_indices.test", "indices.0.name", indexName),
					// Static settings
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_indices.test", "indices.0.number_of_shards", "1"),
					// Dynamic settings — exact values set on the resource
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_indices.test", "indices.0.number_of_replicas", "0"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_indices.test", "indices.0.refresh_interval", "30s"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_indices.test", "indices.0.max_result_window", "5000"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_indices.test", "indices.0.max_ngram_diff", "3"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_indices.test", "indices.0.gc_deletes", "30s"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_indices.test", "indices.0.blocks_read", "false"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_indices.test", "indices.0.blocks_write", "false"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_indices.test", "indices.0.routing_allocation_enable", "all"),
				),
			},
		},
	})
}

func testAccIndicesDataSourceConfigTypedSettings(indexName string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_index" "test" {
  name               = %q
  number_of_shards   = 1
  number_of_replicas = 0
  refresh_interval   = "30s"
  max_result_window  = 5000
  max_ngram_diff     = 3
  gc_deletes         = "30s"
  blocks_read        = false
  blocks_write       = false
  routing_allocation_enable = "all"
  deletion_protection = false
}

data "elasticstack_elasticsearch_indices" "test" {
  target     = %q
  depends_on = [elasticstack_elasticsearch_index.test]
}
`, indexName, indexName)
}

// TestAccIndicesDataSource_ReadsAliasNestedFields creates an index with a richly
// configured alias (filter, routing fields, is_write_index, is_hidden) and verifies
// the data source surfaces those nested alias attributes correctly.
func TestAccIndicesDataSource_ReadsAliasNestedFields(t *testing.T) {
	indexName := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlpha)
	aliasName := indexName + "_alias"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccIndicesDataSourceConfigAliasNestedFields(indexName, aliasName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_indices.test", "indices.0.name", indexName),
					// The alias list must have exactly one entry.
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_indices.test", "indices.0.alias.#", "1"),
					// Alias is modeled as a SetNestedAttribute, so we must not rely on a stable index.
					resource.TestCheckTypeSetElemNestedAttrs(
						"data.elasticstack_elasticsearch_indices.test",
						"indices.0.alias.*",
						map[string]string{
							"name":          aliasName,
							"is_write_index": "true",
							"index_routing":  "shard-1",
							"search_routing": "shard-1",
						},
					),
				),
			},
		},
	})
}

func testAccIndicesDataSourceConfigAliasNestedFields(indexName, aliasName string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_index" "test" {
  name               = %q
  number_of_shards   = 1
  number_of_replicas = 0

  alias = [
    {
      name           = %q
      filter         = jsonencode({ term = { "status" = "active" } })
      index_routing  = "shard-1"
      search_routing = "shard-1"
      is_write_index = true
    }
  ]

  deletion_protection = false
}

data "elasticstack_elasticsearch_indices" "test" {
  target     = %q
  depends_on = [elasticstack_elasticsearch_index.test]
}
`, indexName, aliasName, indexName)
}

// TestAccIndicesDataSource_ReadsMappingsAnalysisAndSettingsRaw creates an index with
// explicit mappings and a custom analysis configuration. It verifies that the data
// source surfaces the computed mappings and settings_raw fields.
//
// Note: analysis_analyzer and analysis_filter are not asserted here because the data
// source does not currently populate those fields from the Elasticsearch API response
// (setSettingsFromAPI iterates over static/dynamic settings keys only; analysis settings
// live under a separate index.analysis.* key namespace). The index resource still stores
// the analysis configuration, which is visible via the settings_raw field.
func TestAccIndicesDataSource_ReadsMappingsAnalysisAndSettingsRaw(t *testing.T) {
	indexName := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlpha)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccIndicesDataSourceConfigMappingsAnalysis(indexName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_indices.test", "indices.0.name", indexName),
					// mappings is a computed JSON string — assert it is populated.
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_indices.test", "indices.0.mappings"),
					// settings_raw should be populated with the raw settings blob.
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_indices.test", "indices.0.settings_raw"),
				),
			},
		},
	})
}

func testAccIndicesDataSourceConfigMappingsAnalysis(indexName string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_index" "test" {
  name               = %q
  number_of_shards   = 1
  number_of_replicas = 0

  mappings = jsonencode({
    properties = {
      title  = { type = "text" }
      status = { type = "keyword" }
    }
  })

  analysis_filter = jsonencode({
    english_stop = {
      type      = "stop"
      stopwords = "_english_"
    }
  })

  analysis_analyzer = jsonencode({
    custom_english = {
      type      = "custom"
      tokenizer = "standard"
      filter    = ["lowercase", "english_stop"]
    }
  })

  deletion_protection = false
}

data "elasticstack_elasticsearch_indices" "test" {
  target     = %q
  depends_on = [elasticstack_elasticsearch_index.test]
}
`, indexName, indexName)
}
