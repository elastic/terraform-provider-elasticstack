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

package security_entity_store_test

import (
	"regexp"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccResourceKibanaSecurityEntityStore_basic(t *testing.T) {
	testAccEntityStoreApplyAndPlan(t, basicConfig(), resource.TestCheckResourceAttrSet("elasticstack_kibana_security_entity_store.test", "id"))
}

func TestAccResourceKibanaSecurityEntityStore_singleType(t *testing.T) {
	testAccEntityStoreApplyAndPlan(t, singleTypeConfig(),
		resource.TestCheckResourceAttr("elasticstack_kibana_security_entity_store.test", "entity_types.#", "1"),
		resource.TestCheckTypeSetElemAttr("elasticstack_kibana_security_entity_store.test", "entity_types.*", "host"),
	)
}

func TestAccResourceKibanaSecurityEntityStore_updateLogExtraction(t *testing.T) {
	testAccEntityStoreApplyAndPlan(t, updateLogExtractionConfig(),
		resource.TestCheckResourceAttr("elasticstack_kibana_security_entity_store.test", "log_extraction.delay", "5m"),
		resource.TestCheckResourceAttr("elasticstack_kibana_security_entity_store.test", "log_extraction.frequency", "10m"),
	)
}

func TestAccResourceKibanaSecurityEntityStore_import(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ResourceName:             "elasticstack_kibana_security_entity_store.test",
				ImportState:              true,
				ImportStateVerify:        true,
				ImportStateVerifyIgnore:  []string{"allow_entity_type_shrink", "history_snapshot"},
			},
		},
	})
}

func TestAccResourceKibanaSecurityEntityStore_shrinkGuardFails(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("shrink"),
				ExpectError:              regexp.MustCompile("Entity type shrink blocked"),
			},
		},
	})
}

func TestAccResourceKibanaSecurityEntityStore_shrinkWithFlag(t *testing.T) {
	testAccEntityStoreApplyAndPlan(t, shrinkWithFlagConfig(),
		resource.TestCheckResourceAttr("elasticstack_kibana_security_entity_store.test", "entity_types.#", "1"),
		resource.TestCheckTypeSetElemAttr("elasticstack_kibana_security_entity_store.test", "entity_types.*", "host"),
		resource.TestCheckResourceAttr("elasticstack_kibana_security_entity_store.test", "allow_entity_type_shrink", "true"),
	)
}

func TestAccResourceKibanaSecurityEntityStore_startedFalse(t *testing.T) {
	testAccEntityStoreApplyAndPlan(t, startedFalseConfig(),
		resource.TestCheckResourceAttr("elasticstack_kibana_security_entity_store.test", "started", "false"),
	)
}

func TestAccDataSourceKibanaSecurityEntityStoreStatus_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("default"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_kibana_security_entity_store_status.test", "installed"),
					resource.TestCheckResourceAttrSet("data.elasticstack_kibana_security_entity_store_status.test", "overall_status"),
					resource.TestCheckResourceAttrSet("data.elasticstack_kibana_security_entity_store_status.test", "engines_json"),
					resource.TestCheckResourceAttrSet("data.elasticstack_kibana_security_entity_store_status.test", "status_json"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("with_components"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_kibana_security_entity_store_status.test", "installed"),
					resource.TestCheckResourceAttrSet("data.elasticstack_kibana_security_entity_store_status.test", "overall_status"),
					resource.TestCheckResourceAttrSet("data.elasticstack_kibana_security_entity_store_status.test", "engines_json"),
					resource.TestCheckResourceAttrSet("data.elasticstack_kibana_security_entity_store_status.test", "status_json"),
				),
			},
		},
	})
}

func testAccEntityStoreApplyAndPlan(t *testing.T, cfg string, checks ...resource.TestCheckFunc) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: cfg,
				Check:  resource.ComposeTestCheckFunc(checks...),
			},
			{Config: cfg, PlanOnly: true},
		},
	})
}

func basicConfig() string {
	return `resource "elasticstack_kibana_security_entity_store" "test" {}
`
}

func singleTypeConfig() string {
	return `resource "elasticstack_kibana_security_entity_store" "test" {
  entity_types = ["host"]
}
`
}

func updateLogExtractionConfig() string {
	return `resource "elasticstack_kibana_security_entity_store" "test" {
  log_extraction {
    delay     = "5m"
    frequency = "10m"
  }
}
`
}

func shrinkWithFlagConfig() string {
	return `resource "elasticstack_kibana_security_entity_store" "test" {
  entity_types             = ["host"]
  allow_entity_type_shrink = true
}
`
}

func startedFalseConfig() string {
	return `resource "elasticstack_kibana_security_entity_store" "test" {
  started = false
}
`
}
