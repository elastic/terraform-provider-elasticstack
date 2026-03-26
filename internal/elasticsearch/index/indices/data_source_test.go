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
	"regexp"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/hashicorp/go-version"
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
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_indices.security_indices", "id", ".security-*"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_indices.security_indices", "target", ".security-*"),
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
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_indices.all_default", "id", "*"),
					resource.TestCheckNoResourceAttr("data.elasticstack_elasticsearch_indices.all_default", "target"),
					// indices.0.name being set proves at least one index was returned.
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_indices.all_default", "indices.0.name"),
				),
			},
			{
				// Explicit "*" wildcard — should return all non-hidden indices.
				Config: testAccIndicesDataSourceConfigStar,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_indices.all_star", "id", "*"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_indices.all_star", "target", "*"),
					// indices.0.name being set proves at least one index was returned.
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_indices.all_star", "indices.0.name"),
				),
			},
			{
				// Explicit "_all" wildcard — equivalent to "*".
				Config: testAccIndicesDataSourceConfigExplicitAll,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_indices.all_explicit", "id", "_all"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_indices.all_explicit", "target", "_all"),
					// indices.0.name being set proves at least one index was returned.
					resource.TestCheckResourceAttrSet("data.elasticstack_elasticsearch_indices.all_explicit", "indices.0.name"),
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
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_indices.wildcard", "id", prefix+"*"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_indices.wildcard", "target", prefix+"*"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_indices.wildcard", "indices.#", "2"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_indices.exact", "id", indexA),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_indices.exact", "target", indexA),
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
							"name":           aliasName,
							"filter":         `{"term":{"status":"active"}}`,
							"index_routing":  "shard-1",
							"is_hidden":      "false",
							"is_write_index": "true",
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
  deletion_protection = false

  lifecycle {
    ignore_changes = [settings_raw]
  }
}

resource "elasticstack_elasticsearch_index_alias" "test" {
  name = %q

  write_index = {
    name           = elasticstack_elasticsearch_index.test.name
    filter         = jsonencode({ term = { "status" = "active" } })
    index_routing  = "shard-1"
    is_hidden      = false
    search_routing = "shard-1"
  }
}

data "elasticstack_elasticsearch_indices" "test" {
  target     = %q
  depends_on = [elasticstack_elasticsearch_index_alias.test]
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
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_indices.test", "target", indexName),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_indices.test", "indices.0.name", indexName),
					resource.TestCheckResourceAttr(
						"data.elasticstack_elasticsearch_indices.test",
						"indices.0.mappings",
						`{"properties":{"status":{"type":"keyword"},"title":{"type":"text"}}}`,
					),
					resource.TestMatchResourceAttr(
						"data.elasticstack_elasticsearch_indices.test",
						"indices.0.settings_raw",
						regexp.MustCompile(regexp.QuoteMeta(`"index.analysis.analyzer.custom_english.type":"custom"`)),
					),
					resource.TestMatchResourceAttr(
						"data.elasticstack_elasticsearch_indices.test",
						"indices.0.settings_raw",
						regexp.MustCompile(regexp.QuoteMeta(`"index.analysis.filter.english_stop.type":"stop"`)),
					),
					resource.TestMatchResourceAttr(
						"data.elasticstack_elasticsearch_indices.test",
						"indices.0.settings_raw",
						regexp.MustCompile(regexp.QuoteMeta(`"index.number_of_replicas":"0"`)),
					),
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

// TestAccIndicesDataSource_ReadsIndexSettings_BroadCoverage creates an index with a
// wider set of scalar settings and verifies the data source returns exact values for
// high-impact settings that were previously untested.
func TestAccIndicesDataSource_ReadsIndexSettings_BroadCoverage(t *testing.T) {
	indexName := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlpha)
	pipelineName := sdkacctest.RandomWithPrefix("tf-acc-test")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccIndicesDataSourceConfigBroadSettings(indexName, pipelineName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_indices.test", "target", indexName),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_indices.test", "indices.0.name", indexName),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_indices.test", "indices.0.number_of_shards", "2"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_indices.test", "indices.0.codec", "best_compression"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_indices.test", "indices.0.mapping_coerce", "false"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_indices.test", "indices.0.max_inner_result_window", "250"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_indices.test", "indices.0.max_rescore_window", "300"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_indices.test", "indices.0.max_docvalue_fields_search", "50"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_indices.test", "indices.0.max_script_fields", "20"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_indices.test", "indices.0.max_shingle_diff", "4"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_indices.test", "indices.0.max_refresh_listeners", "150"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_indices.test", "indices.0.analyze_max_token_count", "5000"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_indices.test", "indices.0.highlight_max_analyzed_offset", "200000"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_indices.test", "indices.0.max_terms_count", "2048"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_indices.test", "indices.0.max_regex_length", "2000"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_indices.test", "indices.0.routing_rebalance_enable", "replicas"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_indices.test", "indices.0.blocks_metadata", "false"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_indices.test", "indices.0.default_pipeline", pipelineName),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_indices.test", "indices.0.final_pipeline", pipelineName),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_indices.test", "indices.0.unassigned_node_left_delayed_timeout", "45s"),
				),
			},
		},
	})
}

func testAccIndicesDataSourceConfigBroadSettings(indexName, pipelineName string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_ingest_pipeline" "test" {
  name        = %q
  description = "Acceptance test pipeline"

  processors = [
    jsonencode({ set = { field = "_pipeline_test", value = "1" } })
  ]
}

resource "elasticstack_elasticsearch_index" "test" {
  name                               = %q
  number_of_shards                   = 2
  number_of_replicas                 = 0
  codec                              = "best_compression"
  mapping_coerce                     = false
  max_inner_result_window            = 250
  max_rescore_window                 = 300
  max_docvalue_fields_search         = 50
  max_script_fields                  = 20
  max_shingle_diff                   = 4
  max_refresh_listeners              = 150
  analyze_max_token_count            = 5000
  highlight_max_analyzed_offset      = 200000
  max_terms_count                    = 2048
  max_regex_length                   = 2000
  routing_rebalance_enable           = "replicas"
  blocks_metadata                    = false
  default_pipeline                   = elasticstack_elasticsearch_ingest_pipeline.test.name
  final_pipeline                     = elasticstack_elasticsearch_ingest_pipeline.test.name
  unassigned_node_left_delayed_timeout = "45s"
  deletion_protection                = false
}

data "elasticstack_elasticsearch_indices" "test" {
  target     = %q
  depends_on = [elasticstack_elasticsearch_index.test]
}
`, pipelineName, indexName, indexName)
}

// TestAccIndicesDataSource_ReadsSlowlogSettings verifies representative search and
// indexing slowlog thresholds from the data source.
func TestAccIndicesDataSource_ReadsSlowlogSettings(t *testing.T) {
	indexName := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlpha)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccIndicesDataSourceConfigSlowlog(indexName, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_indices.test", "target", indexName),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_indices.test", "indices.0.name", indexName),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_indices.test", "indices.0.search_slowlog_threshold_query_warn", "10s"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_indices.test", "indices.0.search_slowlog_threshold_fetch_info", "800ms"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_indices.test", "indices.0.indexing_slowlog_threshold_index_debug", "10ms"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_indices.test", "indices.0.indexing_slowlog_source", "1000"),
				),
			},
		},
	})
}

var slowlogLevelVersionConstraint, _ = version.NewConstraint("< 8.0.0")

// TestAccIndicesDataSource_ReadsSlowlogLevels verifies the slowlog level fields on
// Elastic Stack versions that still expose them.
func TestAccIndicesDataSource_ReadsSlowlogLevels(t *testing.T) {
	indexName := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlpha)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				SkipFunc: versionutils.CheckIfVersionMeetsConstraints(slowlogLevelVersionConstraint),
				Config:   testAccIndicesDataSourceConfigSlowlog(indexName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_indices.test", "target", indexName),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_indices.test", "indices.0.name", indexName),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_indices.test", "indices.0.search_slowlog_level", "info"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_indices.test", "indices.0.indexing_slowlog_level", "warn"),
				),
			},
		},
	})
}

func testAccIndicesDataSourceConfigSlowlog(indexName string, includeLevels bool) string {
	slowlogLevels := ""
	if includeLevels {
		slowlogLevels = `
  search_slowlog_level                 = "info"
  indexing_slowlog_level               = "warn"
`
	}

	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_index" "test" {
  name                                 = %q
  search_slowlog_threshold_query_warn  = "10s"
  search_slowlog_threshold_fetch_info  = "800ms"
  indexing_slowlog_threshold_index_debug = "10ms"
  indexing_slowlog_source              = "1000"%s
  deletion_protection                  = false
}

data "elasticstack_elasticsearch_indices" "test" {
  target     = %q
  depends_on = [elasticstack_elasticsearch_index.test]
}
`, indexName, slowlogLevels, indexName)
}
