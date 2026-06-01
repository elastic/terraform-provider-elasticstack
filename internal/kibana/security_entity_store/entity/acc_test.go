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

package entity_test

import (
	"regexp"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/security_entity_store/entity"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccResourceKibanaSecurityEntityStoreEntity_host(t *testing.T) {
	skipIfUnsupported(t)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create_host"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("elasticstack_kibana_security_entity_store_entity.test", "id"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_entity_store_entity.test", "entity_type", "host"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_entity_store_entity.test", "entity_id", "host:acc-test-host-01"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_security_entity_store_entity.test", "document_json"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create_host"),
				PlanOnly:                 true,
			},
		},
	})
}

func TestAccResourceKibanaSecurityEntityStoreEntity_updateHost(t *testing.T) {
	skipIfUnsupported(t)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create_host"),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update_host"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("elasticstack_kibana_security_entity_store_entity.test", "id"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_entity_store_entity.test", "entity_type", "host"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_entity_store_entity.test", "entity_id", "host:acc-test-host-01"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update_host"),
				PlanOnly:                 true,
			},
		},
	})
}

func TestAccResourceKibanaSecurityEntityStoreEntity_import(t *testing.T) {
	skipIfUnsupported(t)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create_host"),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create_host"),
				ResourceName:             "elasticstack_kibana_security_entity_store_entity.test",
				ImportState:              true,
				ImportStateVerify:        true,
				ImportStateVerifyIgnore:  []string{"force"},
			},
		},
	})
}

func TestAccResourceKibanaSecurityEntityStoreEntity_jsonFallback(t *testing.T) {
	skipIfUnsupported(t)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("entity_json"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("elasticstack_kibana_security_entity_store_entity.test", "id"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_security_entity_store_entity.test", "document_json"),
				),
			},
		},
	})
}

func TestAccResourceKibanaSecurityEntityStoreEntity_entityJsonConflict(t *testing.T) {
	skipIfUnsupported(t)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("entity_json_conflict"),
				ExpectError:              regexp.MustCompile("conflict|ConflictsWith"),
			},
		},
	})
}

func skipIfUnsupported(t *testing.T) {
	versionutils.SkipIfUnsupported(t, entity.MinVersion, versionutils.FlavorAny)
}
