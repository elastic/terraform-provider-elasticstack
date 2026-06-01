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

package security_entity_store_entity_link_test

import (
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

var minVersionEntityStoreResolution = version.Must(version.NewVersion("9.1.0"))

func TestAccResourceSecurityEntityStoreEntityLink(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minVersionEntityStoreResolution, versionutils.FlavorAny)

	// Skipping until entity store fixture data can be created in CI.
	// The resolution APIs require pre-existing entities with enterprise license.
	t.Skip("Skipping: requires entity store with pre-existing entities and enterprise license")

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"target_id":  config.StringVariable("user:target@example.com"),
					"entity_ids": config.ListVariable(config.StringVariable("user:alias1@example.com"), config.StringVariable("user:alias2@example.com")),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_security_entity_store_entity_link.test", "target_id", "user:target@example.com"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_entity_store_entity_link.test", "space_id", "default"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_entity_store_entity_link.test", "id", "default/user:target@example.com"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_security_entity_store_entity_link.test", "resolution_group_json"),
					resource.TestCheckTypeSetElemAttr("elasticstack_kibana_security_entity_store_entity_link.test", "entity_ids.*", "user:alias1@example.com"),
					resource.TestCheckTypeSetElemAttr("elasticstack_kibana_security_entity_store_entity_link.test", "entity_ids.*", "user:alias2@example.com"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables: config.Variables{
					"target_id":  config.StringVariable("user:target@example.com"),
					"entity_ids": config.ListVariable(config.StringVariable("user:alias1@example.com"), config.StringVariable("user:alias3@example.com")),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_security_entity_store_entity_link.test", "target_id", "user:target@example.com"),
					resource.TestCheckTypeSetElemAttr("elasticstack_kibana_security_entity_store_entity_link.test", "entity_ids.*", "user:alias1@example.com"),
					resource.TestCheckTypeSetElemAttr("elasticstack_kibana_security_entity_store_entity_link.test", "entity_ids.*", "user:alias3@example.com"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"target_id":  config.StringVariable("user:target@example.com"),
					"entity_ids": config.ListVariable(config.StringVariable("user:alias1@example.com"), config.StringVariable("user:alias2@example.com")),
				},
				ResourceName:      "elasticstack_kibana_security_entity_store_entity_link.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
