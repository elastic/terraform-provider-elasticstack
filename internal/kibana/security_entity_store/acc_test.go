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
	securityentitystore "github.com/elastic/terraform-provider-elasticstack/internal/kibana/security_entity_store"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccResourceKibanaSecurityEntityStore_basic(t *testing.T) {
	skipIfUnsupported(t)
	t.Cleanup(func() { acctest.CleanupEntityStore(t, "default") })

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("basic"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("elasticstack_kibana_security_entity_store.test", "id"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_entity_store.test", "space_id", "default"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_entity_store.test", "allow_entity_type_shrink", "false"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_entity_store.test", "started", "true"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_security_entity_store.test", "status_json"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("basic"),
				PlanOnly:                 true,
			},
		},
	})
}

func TestAccResourceKibanaSecurityEntityStore_singleType(t *testing.T) {
	skipIfUnsupported(t)
	t.Cleanup(func() { acctest.CleanupEntityStore(t, "default") })

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("single_type"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckTypeSetElemAttr("elasticstack_kibana_security_entity_store.test", "entity_types.*", "host"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("single_type"),
				PlanOnly:                 true,
			},
		},
	})
}

func TestAccResourceKibanaSecurityEntityStore_updateLogExtraction(t *testing.T) {
	skipIfUnsupported(t)
	t.Cleanup(func() { acctest.CleanupEntityStore(t, "default") })

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update_log_extraction"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_security_entity_store.test", "log_extraction.delay", "5m"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_entity_store.test", "log_extraction.frequency", "10m"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_security_entity_store.test", "log_extraction.field_history_length"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_entity_store.test", "log_extraction.lookback_period", "24h"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update_log_extraction"),
				PlanOnly:                 true,
			},
		},
	})
}

func TestAccResourceKibanaSecurityEntityStore_import(t *testing.T) {
	skipIfUnsupported(t)
	t.Cleanup(func() { acctest.CleanupEntityStore(t, "default") })

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
	skipIfUnsupported(t)
	t.Cleanup(func() { acctest.CleanupEntityStore(t, "default") })

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
	skipIfUnsupported(t)
	t.Cleanup(func() { acctest.CleanupEntityStore(t, "default") })

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("shrink_with_flag"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckTypeSetElemAttr("elasticstack_kibana_security_entity_store.test", "entity_types.*", "host"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_entity_store.test", "allow_entity_type_shrink", "true"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("shrink_with_flag"),
				PlanOnly:                 true,
			},
		},
	})
}

func TestAccResourceKibanaSecurityEntityStore_startedFalse(t *testing.T) {
	skipIfUnsupported(t)
	t.Cleanup(func() { acctest.CleanupEntityStore(t, "default") })

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("started_false"),
				Check:                    resource.TestCheckResourceAttr("elasticstack_kibana_security_entity_store.test", "started", "false"),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("started_false"),
				PlanOnly:                 true,
			},
		},
	})
}

func TestAccResourceKibanaSecurityEntityStore_historySnapshot(t *testing.T) {
	skipIfUnsupported(t)
	t.Cleanup(func() { acctest.CleanupEntityStore(t, "default") })

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("history_snapshot"),
				Check:                    resource.TestCheckResourceAttrSet("elasticstack_kibana_security_entity_store.test", "id"),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("history_snapshot"),
				PlanOnly:                 true,
			},
		},
	})
}

func TestAccDataSourceKibanaSecurityEntityStoreStatus_basic(t *testing.T) {
	skipIfUnsupported(t)
	t.Cleanup(func() { acctest.CleanupEntityStore(t, "default") })

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("default"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.elasticstack_kibana_security_entity_store_status.test", "installed", "true"),
					resource.TestMatchResourceAttr("data.elasticstack_kibana_security_entity_store_status.test", "overall_status", regexp.MustCompile(`^(running|stopped|error|installing)$`)),
					resource.TestMatchResourceAttr("data.elasticstack_kibana_security_entity_store_status.test", "engines.#", regexp.MustCompile(`^[1-9][0-9]*$`)),
					resource.TestCheckResourceAttrSet("data.elasticstack_kibana_security_entity_store_status.test", "engines.0.type"),
					resource.TestCheckResourceAttrSet("data.elasticstack_kibana_security_entity_store_status.test", "engines.0.status"),
					resource.TestCheckResourceAttr("data.elasticstack_kibana_security_entity_store_status.test", "space_id", "default"),
					resource.TestCheckResourceAttrSet("data.elasticstack_kibana_security_entity_store_status.test", "status_json"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("with_components"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.elasticstack_kibana_security_entity_store_status.test", "installed", "true"),
					resource.TestMatchResourceAttr("data.elasticstack_kibana_security_entity_store_status.test", "overall_status", regexp.MustCompile(`^(running|stopped|error|installing)$`)),
					resource.TestMatchResourceAttr("data.elasticstack_kibana_security_entity_store_status.test", "engines.#", regexp.MustCompile(`^[1-9][0-9]*$`)),
					resource.TestCheckResourceAttrSet("data.elasticstack_kibana_security_entity_store_status.test", "engines.0.type"),
					resource.TestCheckResourceAttrSet("data.elasticstack_kibana_security_entity_store_status.test", "engines.0.status"),
					resource.TestCheckResourceAttr("data.elasticstack_kibana_security_entity_store_status.test", "space_id", "default"),
					resource.TestCheckResourceAttrSet("data.elasticstack_kibana_security_entity_store_status.test", "engines.0.components.#"),
					resource.TestCheckResourceAttrSet("data.elasticstack_kibana_security_entity_store_status.test", "engines.0.components.0.id"),
					resource.TestCheckResourceAttrSet("data.elasticstack_kibana_security_entity_store_status.test", "status_json"),
				),
			},
		},
	})
}

func skipIfUnsupported(t *testing.T) {
	versionutils.SkipIfUnsupported(t, securityentitystore.MinVersion, versionutils.FlavorAny)
}
