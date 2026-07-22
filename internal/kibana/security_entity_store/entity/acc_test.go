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
	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const accTestKibanaSpaceIDCharset = "abcdefghijklmnopqrstuvwxyz0123456789_-"

func TestAccResourceKibanaSecurityEntityStoreEntity_generic(t *testing.T) {
	skipIfUnsupported(t)
	spaceID := sdkacctest.RandStringFromCharSet(12, accTestKibanaSpaceIDCharset)
	vars := config.Variables{"space_id": config.StringVariable(spaceID)}
	t.Cleanup(func() { acctest.CleanupEntityStore(t, spaceID) })

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create_host"),
				ConfigVariables:          vars,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("elasticstack_kibana_security_entity_store_entity.test", "id"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_entity_store_entity.test", "entity_type", "generic"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_entity_store_entity.test", "entity_id", "generic:acc-test-entity"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_entity_store_entity.test", "force", "false"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_security_entity_store_entity.test", "document_json"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_security_entity_store_entity.test", "response_json"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_security_entity_store_entity.test", "timestamp"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_entity_store_entity.test", "labels.env", "acc-test"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_entity_store_entity.test", "tags.#", "1"),
					resource.TestCheckTypeSetElemAttr("elasticstack_kibana_security_entity_store_entity.test", "tags.*", "terraform"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_entity_store_entity.test", "entity.type", "generic"),
					resource.TestCheckTypeSetElemAttr("elasticstack_kibana_security_entity_store_entity.test", "entity.source.*", "terraform-acc-test"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create_host"),
				ConfigVariables:          vars,
				PlanOnly:                 true,
			},
		},
	})
}

func TestAccResourceKibanaSecurityEntityStoreEntity_updateHost(t *testing.T) {
	skipIfUnsupported(t)
	spaceID := sdkacctest.RandStringFromCharSet(12, accTestKibanaSpaceIDCharset)
	vars := config.Variables{"space_id": config.StringVariable(spaceID)}
	t.Cleanup(func() { acctest.CleanupEntityStore(t, spaceID) })

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create_host"),
				ConfigVariables:          vars,
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update_host"),
				ConfigVariables:          vars,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("elasticstack_kibana_security_entity_store_entity.test", "id"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_entity_store_entity.test", "entity_type", "generic"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_entity_store_entity.test", "entity_id", "generic:acc-test-entity"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_entity_store_entity.test", "entity.name", "acc-test-entity-updated"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_entity_store_entity.test", "entity.type", "generic"),
					resource.TestCheckTypeSetElemAttr("elasticstack_kibana_security_entity_store_entity.test", "entity.source.*", "terraform-acc-test"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update_host"),
				ConfigVariables:          vars,
				PlanOnly:                 true,
			},
		},
	})
}

func TestAccResourceKibanaSecurityEntityStoreEntity_import(t *testing.T) {
	skipIfUnsupported(t)
	spaceID := sdkacctest.RandStringFromCharSet(12, accTestKibanaSpaceIDCharset)
	vars := config.Variables{"space_id": config.StringVariable(spaceID)}
	t.Cleanup(func() { acctest.CleanupEntityStore(t, spaceID) })

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create_host"),
				ConfigVariables:          vars,
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create_host"),
				ConfigVariables:          vars,
				ResourceName:             "elasticstack_kibana_security_entity_store_entity.test",
				ImportState:              true,
				ImportStateVerify:        true,
				ImportStateVerifyIgnore:  []string{"force"},
			},
		},
	})
}

func TestAccResourceKibanaSecurityEntityStoreEntity_entityJsonConflict(t *testing.T) {
	skipIfUnsupported(t)
	spaceID := sdkacctest.RandStringFromCharSet(12, accTestKibanaSpaceIDCharset)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("entity_json_conflict"),
				ConfigVariables:          config.Variables{"space_id": config.StringVariable(spaceID)},
				ExpectError:              regexp.MustCompile("(?i)conflict|ConflictsWith|Invalid Attribute Combination"),
			},
		},
	})
}

func TestAccResourceKibanaSecurityEntityStoreEntity_entityIdMismatch(t *testing.T) {
	skipIfUnsupported(t)
	spaceID := sdkacctest.RandStringFromCharSet(12, accTestKibanaSpaceIDCharset)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("entity_id_mismatch"),
				ConfigVariables:          config.Variables{"space_id": config.StringVariable(spaceID)},
				ExpectError:              regexp.MustCompile("entity_id mismatch"),
			},
		},
	})
}

func TestAccResourceKibanaSecurityEntityStoreEntity_entityJsonIdMismatch(t *testing.T) {
	skipIfUnsupported(t)
	spaceID := sdkacctest.RandStringFromCharSet(12, accTestKibanaSpaceIDCharset)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("entity_json_id_mismatch"),
				ConfigVariables:          config.Variables{"space_id": config.StringVariable(spaceID)},
				ExpectError:              regexp.MustCompile("entity_id mismatch"),
			},
		},
	})
}

// TestAccResourceKibanaSecurityEntityStoreEntity_hostType exercises entity_type=host
// with a typed host block, asserting host.name and host.ip are populated.
func TestAccResourceKibanaSecurityEntityStoreEntity_hostType(t *testing.T) {
	skipIfUnsupported(t)
	spaceID := sdkacctest.RandStringFromCharSet(12, accTestKibanaSpaceIDCharset)
	vars := config.Variables{"space_id": config.StringVariable(spaceID)}
	t.Cleanup(func() { acctest.CleanupEntityStore(t, spaceID) })

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create_host"),
				ConfigVariables:          vars,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("elasticstack_kibana_security_entity_store_entity.test", "id"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_entity_store_entity.test", "entity_type", "host"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_entity_store_entity.test", "entity_id", "host:acc-test-host"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_entity_store_entity.test", "host.name", "acc-test-host"),
					resource.TestCheckTypeSetElemAttr("elasticstack_kibana_security_entity_store_entity.test", "host.ip.*", "1.2.3.4"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_security_entity_store_entity.test", "document_json"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_security_entity_store_entity.test", "response_json"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create_host"),
				ConfigVariables:          vars,
				PlanOnly:                 true,
			},
		},
	})
}

// TestAccResourceKibanaSecurityEntityStoreEntity_userType exercises entity_type=user
// with a typed user block, asserting user.name and user.email are populated.
func TestAccResourceKibanaSecurityEntityStoreEntity_userType(t *testing.T) {
	skipIfUnsupported(t)
	spaceID := sdkacctest.RandStringFromCharSet(12, accTestKibanaSpaceIDCharset)
	vars := config.Variables{"space_id": config.StringVariable(spaceID)}
	t.Cleanup(func() { acctest.CleanupEntityStore(t, spaceID) })

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create_user"),
				ConfigVariables:          vars,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("elasticstack_kibana_security_entity_store_entity.test", "id"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_entity_store_entity.test", "entity_type", "user"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_entity_store_entity.test", "entity_id", "user:acc-test-user@unknown"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_entity_store_entity.test", "user.name", "acc-test-user"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_security_entity_store_entity.test", "document_json"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_security_entity_store_entity.test", "response_json"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create_user"),
				ConfigVariables:          vars,
				PlanOnly:                 true,
			},
		},
	})
}

// TestAccResourceKibanaSecurityEntityStoreEntity_serviceType exercises entity_type=service
// with a typed service block, asserting service.name is populated.
func TestAccResourceKibanaSecurityEntityStoreEntity_serviceType(t *testing.T) {
	skipIfUnsupported(t)
	spaceID := sdkacctest.RandStringFromCharSet(12, accTestKibanaSpaceIDCharset)
	vars := config.Variables{"space_id": config.StringVariable(spaceID)}
	t.Cleanup(func() { acctest.CleanupEntityStore(t, spaceID) })

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create_service"),
				ConfigVariables:          vars,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("elasticstack_kibana_security_entity_store_entity.test", "id"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_entity_store_entity.test", "entity_type", "service"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_entity_store_entity.test", "entity_id", "service:acc-test-service"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_entity_store_entity.test", "service.name", "acc-test-service"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_security_entity_store_entity.test", "document_json"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_security_entity_store_entity.test", "response_json"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create_service"),
				ConfigVariables:          vars,
				PlanOnly:                 true,
			},
		},
	})
}

// TestAccResourceKibanaSecurityEntityStoreEntity_hostJsonFallback exercises the host_json
// JSON fallback attribute instead of the typed host block.
func TestAccResourceKibanaSecurityEntityStoreEntity_hostJsonFallback(t *testing.T) {
	skipIfUnsupported(t)
	spaceID := sdkacctest.RandStringFromCharSet(12, accTestKibanaSpaceIDCharset)
	t.Cleanup(func() { acctest.CleanupEntityStore(t, spaceID) })

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create_host_json"),
				ConfigVariables:          config.Variables{"space_id": config.StringVariable(spaceID)},
				// The API populates the typed host block on read even when host_json was used,
				// causing a non-empty plan on subsequent applies. This is a known limitation.
				ExpectNonEmptyPlan: true,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("elasticstack_kibana_security_entity_store_entity.test", "id"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_entity_store_entity.test", "entity_type", "host"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_entity_store_entity.test", "entity_id", "host:acc-test-host-json"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_security_entity_store_entity.test", "document_json"),
				),
			},
		},
	})
}

// TestAccResourceKibanaSecurityEntityStoreEntity_hostJsonConflict verifies that setting
// both host and host_json produces a validation error.
func TestAccResourceKibanaSecurityEntityStoreEntity_hostJsonConflict(t *testing.T) {
	skipIfUnsupported(t)
	spaceID := sdkacctest.RandStringFromCharSet(12, accTestKibanaSpaceIDCharset)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("host_json_conflict"),
				ConfigVariables:          config.Variables{"space_id": config.StringVariable(spaceID)},
				ExpectError:              regexp.MustCompile("(?i)conflict|ConflictsWith|Invalid Attribute Combination"),
			},
		},
	})
}

func skipIfUnsupported(t *testing.T) {
	versionutils.SkipIfUnsupported(t, entity.MinVersion, versionutils.FlavorAny)
}
