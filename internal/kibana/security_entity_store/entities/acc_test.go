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
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDataSourceKibanaSecurityEntityStoreEntities_basic(t *testing.T) {
	skipIfUnsupported(t)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("list"),
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

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("mixed_pagination"),
				ExpectError:              regexp.MustCompile("Mixed pagination modes"),
			},
		},
	})
}

func TestAccDataSourceKibanaSecurityEntityStoreEntities_entityIdFilterConflict(t *testing.T) {
	skipIfUnsupported(t)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("entity_id_filter_conflict"),
				ExpectError:              regexp.MustCompile("(?i)conflict|ConflictsWith|Invalid Attribute Combination"),
			},
		},
	})
}

func skipIfUnsupported(t *testing.T) {
	versionutils.SkipIfUnsupported(t, securityentitystore.MinVersion, versionutils.FlavorAny)
}
