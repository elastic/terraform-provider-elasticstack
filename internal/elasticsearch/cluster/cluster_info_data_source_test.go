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
	"regexp"
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

const testAccDataSourceClusterInfoConfig = `
provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_info" "test" {
}
`
