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

package entities_test

import (
	"regexp"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	securityentitystore "github.com/elastic/terraform-provider-elasticstack/internal/kibana/security_entity_store"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const accTestKibanaSpaceIDCharset = "abcdefghijklmnopqrstuvwxyz0123456789_-"

func TestAccDataSourceKibanaSecurityEntityStoreEntities_basic(t *testing.T) {
	skipIfUnsupported(t)
	spaceID := sdkacctest.RandStringFromCharSet(12, accTestKibanaSpaceIDCharset)
	t.Cleanup(func() { acctest.CleanupEntityStore(t, spaceID) })

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("list"),
				ConfigVariables:          config.Variables{"space_id": config.StringVariable(spaceID)},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_kibana_security_entity_store_entities.test", "id"),
					resource.TestCheckResourceAttrSet("data.elasticstack_kibana_security_entity_store_entities.test", "results_json"),
					resource.TestCheckResourceAttrSet("data.elasticstack_kibana_security_entity_store_entities.test", "items.#"),
				),
			},
		},
	})
}

func TestAccDataSourceKibanaSecurityEntityStoreEntities_mixedPaginationError(t *testing.T) {
	skipIfUnsupported(t)
	spaceID := sdkacctest.RandStringFromCharSet(12, accTestKibanaSpaceIDCharset)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("mixed_pagination"),
				ConfigVariables:          config.Variables{"space_id": config.StringVariable(spaceID)},
				ExpectError:              regexp.MustCompile("Mixed pagination modes"),
			},
		},
	})
}

func TestAccDataSourceKibanaSecurityEntityStoreEntities_entityIdFilterConflict(t *testing.T) {
	skipIfUnsupported(t)
	spaceID := sdkacctest.RandStringFromCharSet(12, accTestKibanaSpaceIDCharset)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("entity_id_filter_conflict"),
				ConfigVariables:          config.Variables{"space_id": config.StringVariable(spaceID)},
				ExpectError:              regexp.MustCompile("(?i)conflict|ConflictsWith|Invalid Attribute Combination"),
			},
		},
	})
}

func TestAccDataSourceKibanaSecurityEntityStoreEntities_filter(t *testing.T) {
	skipIfUnsupported(t)
	spaceID := sdkacctest.RandStringFromCharSet(12, accTestKibanaSpaceIDCharset)
	t.Cleanup(func() { acctest.CleanupEntityStore(t, spaceID) })

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("filter"),
				ConfigVariables:          config.Variables{"space_id": config.StringVariable(spaceID)},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_kibana_security_entity_store_entities.test", "id"),
					resource.TestCheckResourceAttrSet("data.elasticstack_kibana_security_entity_store_entities.test", "items.#"),
					resource.TestCheckResourceAttr("data.elasticstack_kibana_security_entity_store_entities.test", "space_id", spaceID),
				),
			},
		},
	})
}

func TestAccDataSourceKibanaSecurityEntityStoreEntities_pageMode(t *testing.T) {
	skipIfUnsupported(t)
	spaceID := sdkacctest.RandStringFromCharSet(12, accTestKibanaSpaceIDCharset)
	t.Cleanup(func() { acctest.CleanupEntityStore(t, spaceID) })

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("page_mode"),
				ConfigVariables:          config.Variables{"space_id": config.StringVariable(spaceID)},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_kibana_security_entity_store_entities.test", "id"),
					resource.TestCheckResourceAttr("data.elasticstack_kibana_security_entity_store_entities.test", "page", "1"),
					resource.TestCheckResourceAttr("data.elasticstack_kibana_security_entity_store_entities.test", "per_page", "10"),
					resource.TestCheckResourceAttr("data.elasticstack_kibana_security_entity_store_entities.test", "sort_field", "entity.id"),
					resource.TestCheckResourceAttr("data.elasticstack_kibana_security_entity_store_entities.test", "sort_order", "asc"),
					resource.TestCheckResourceAttrSet("data.elasticstack_kibana_security_entity_store_entities.test", "items.#"),
				),
			},
		},
	})
}

func TestAccDataSourceKibanaSecurityEntityStoreEntities_spaceId(t *testing.T) {
	skipIfUnsupported(t)
	spaceID := sdkacctest.RandStringFromCharSet(12, accTestKibanaSpaceIDCharset)
	t.Cleanup(func() { acctest.CleanupEntityStore(t, spaceID) })

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("space_id"),
				ConfigVariables:          config.Variables{"space_id": config.StringVariable(spaceID)},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_kibana_security_entity_store_entities.test", "id"),
					resource.TestCheckResourceAttr("data.elasticstack_kibana_security_entity_store_entities.test", "space_id", spaceID),
					resource.TestCheckResourceAttrSet("data.elasticstack_kibana_security_entity_store_entities.test", "items.#"),
				),
			},
		},
	})
}

func TestAccDataSourceKibanaSecurityEntityStoreEntities_entityId(t *testing.T) {
	skipIfUnsupported(t)
	spaceID := sdkacctest.RandStringFromCharSet(12, accTestKibanaSpaceIDCharset)
	t.Cleanup(func() { acctest.CleanupEntityStore(t, spaceID) })

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("entity_id"),
				ConfigVariables:          config.Variables{"space_id": config.StringVariable(spaceID)},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_kibana_security_entity_store_entities.test", "id"),
					resource.TestCheckResourceAttrSet("data.elasticstack_kibana_security_entity_store_entities.test", "items.#"),
				),
			},
		},
	})
}

func TestAccDataSourceKibanaSecurityEntityStoreEntities_filterQuery(t *testing.T) {
	skipIfUnsupported(t)
	spaceID := sdkacctest.RandStringFromCharSet(12, accTestKibanaSpaceIDCharset)
	t.Cleanup(func() { acctest.CleanupEntityStore(t, spaceID) })

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("filter_query"),
				ConfigVariables:          config.Variables{"space_id": config.StringVariable(spaceID)},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_kibana_security_entity_store_entities.test", "id"),
					resource.TestCheckResourceAttrSet("data.elasticstack_kibana_security_entity_store_entities.test", "results_json"),
					resource.TestCheckResourceAttr("data.elasticstack_kibana_security_entity_store_entities.test", "per_page", "5"),
					resource.TestCheckResourceAttr("data.elasticstack_kibana_security_entity_store_entities.test", "page", "1"),
				),
			},
		},
	})
}

func skipIfUnsupported(t *testing.T) {
	versionutils.SkipIfUnsupported(t, securityentitystore.MinVersion, versionutils.FlavorAny)
}
