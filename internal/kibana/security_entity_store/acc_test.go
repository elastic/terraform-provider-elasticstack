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
	"context"
	"fmt"
	"regexp"
	"strconv"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	securityentitystore "github.com/elastic/terraform-provider-elasticstack/internal/kibana/security_entity_store"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

const accTestKibanaSpaceIDCharset = "abcdefghijklmnopqrstuvwxyz0123456789_-"

func TestAccResourceKibanaSecurityEntityStore_basic(t *testing.T) {
	skipIfUnsupported(t)
	spaceID := sdkacctest.RandStringFromCharSet(12, accTestKibanaSpaceIDCharset)
	vars := config.Variables{"space_id": config.StringVariable(spaceID)}
	t.Cleanup(func() { acctest.CleanupEntityStore(t, spaceID) })

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("basic"),
				ConfigVariables:          vars,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("elasticstack_kibana_security_entity_store.test", "id"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_entity_store.test", "space_id", spaceID),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_entity_store.test", "allow_entity_type_shrink", "false"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_entity_store.test", "started", "true"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_security_entity_store.test", "status_json"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("basic"),
				ConfigVariables:          vars,
				PlanOnly:                 true,
			},
		},
	})
}

func TestAccResourceKibanaSecurityEntityStore_singleType(t *testing.T) {
	skipIfUnsupported(t)
	spaceID := sdkacctest.RandStringFromCharSet(12, accTestKibanaSpaceIDCharset)
	vars := config.Variables{"space_id": config.StringVariable(spaceID)}
	t.Cleanup(func() { acctest.CleanupEntityStore(t, spaceID) })

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("single_type"),
				ConfigVariables:          vars,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckTypeSetElemAttr("elasticstack_kibana_security_entity_store.test", "entity_types.*", "host"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("single_type"),
				ConfigVariables:          vars,
				PlanOnly:                 true,
			},
		},
	})
}

func TestAccResourceKibanaSecurityEntityStore_updateLogExtraction(t *testing.T) {
	skipIfUnsupported(t)
	spaceID := sdkacctest.RandStringFromCharSet(12, accTestKibanaSpaceIDCharset)
	vars := config.Variables{"space_id": config.StringVariable(spaceID)}
	t.Cleanup(func() { acctest.CleanupEntityStore(t, spaceID) })

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update_log_extraction"),
				ConfigVariables:          vars,
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
				ConfigVariables:          vars,
				PlanOnly:                 true,
			},
		},
	})
}

func TestAccResourceKibanaSecurityEntityStore_import(t *testing.T) {
	skipIfUnsupported(t)
	spaceID := sdkacctest.RandStringFromCharSet(12, accTestKibanaSpaceIDCharset)
	vars := config.Variables{"space_id": config.StringVariable(spaceID)}
	t.Cleanup(func() { acctest.CleanupEntityStore(t, spaceID) })

	const resName = "elasticstack_kibana_security_entity_store.test"

	// ImportStateVerify compares the imported state against the pre-import
	// state byte-for-byte, so the raw status_json (whose engine array order is
	// arbitrary per API response) is excluded there and re-verified below with
	// the StatusJSON custom type's semantic equality instead.
	var preImportStatusJSON string

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          vars,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resName, "id"),
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources[resName]
						if !ok || rs == nil {
							return fmt.Errorf("resource %s not found in state to capture pre-import status_json", resName)
						}
						if rs.Primary == nil {
							return fmt.Errorf("resource %s has no primary instance state to capture pre-import status_json", resName)
						}
						preImportStatusJSON = rs.Primary.Attributes["status_json"]
						if preImportStatusJSON == "" {
							return fmt.Errorf("resource %s has no status_json to capture", resName)
						}
						return nil
					},
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          vars,
				ResourceName:             resName,
				ImportState:              true,
				ImportStateVerify:        true,
				// allow_entity_type_shrink and history_snapshot are Terraform-only
				// inputs the API never echoes back; status_json is excluded from the
				// byte comparison and verified semantically in ImportStateCheck.
				ImportStateVerifyIgnore: []string{"allow_entity_type_shrink", "history_snapshot", "status_json"},
				ImportStateCheck: func(is []*terraform.InstanceState) error {
					if len(is) != 1 {
						return fmt.Errorf("expected exactly 1 imported instance state, got %d", len(is))
					}
					if is[0] == nil {
						return fmt.Errorf("imported instance state for %s is nil; no state to read status_json from", resName)
					}
					imported := is[0].Attributes["status_json"]
					if imported == "" {
						return fmt.Errorf("imported state for %s has no status_json attribute to compare", resName)
					}
					equal, diags := securityentitystore.NewStatusJSONValue(preImportStatusJSON).
						SemanticallyEqual(context.Background(), securityentitystore.NewStatusJSONValue(imported))
					if diags.HasError() {
						return fmt.Errorf("status_json semantic comparison failed: %v", diags)
					}
					if !equal {
						return fmt.Errorf("imported status_json %s is not semantically equal to pre-import status_json %s", imported, preImportStatusJSON)
					}
					return nil
				},
			},
		},
	})
}

func TestAccResourceKibanaSecurityEntityStore_shrinkGuardFails(t *testing.T) {
	skipIfUnsupported(t)
	spaceID := sdkacctest.RandStringFromCharSet(12, accTestKibanaSpaceIDCharset)
	vars := config.Variables{"space_id": config.StringVariable(spaceID)}
	t.Cleanup(func() { acctest.CleanupEntityStore(t, spaceID) })

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          vars,
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("shrink"),
				ConfigVariables:          vars,
				ExpectError:              regexp.MustCompile("Entity type shrink blocked"),
			},
		},
	})
}

func TestAccResourceKibanaSecurityEntityStore_shrinkWithFlag(t *testing.T) {
	skipIfUnsupported(t)
	spaceID := sdkacctest.RandStringFromCharSet(12, accTestKibanaSpaceIDCharset)
	vars := config.Variables{"space_id": config.StringVariable(spaceID)}
	t.Cleanup(func() { acctest.CleanupEntityStore(t, spaceID) })

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("shrink_with_flag"),
				ConfigVariables:          vars,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckTypeSetElemAttr("elasticstack_kibana_security_entity_store.test", "entity_types.*", "host"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_entity_store.test", "allow_entity_type_shrink", "true"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("shrink_with_flag"),
				ConfigVariables:          vars,
				PlanOnly:                 true,
			},
		},
	})
}

func TestAccResourceKibanaSecurityEntityStore_startedFalse(t *testing.T) {
	skipIfUnsupported(t)
	spaceID := sdkacctest.RandStringFromCharSet(12, accTestKibanaSpaceIDCharset)
	vars := config.Variables{"space_id": config.StringVariable(spaceID)}
	t.Cleanup(func() { acctest.CleanupEntityStore(t, spaceID) })

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("started_false"),
				ConfigVariables:          vars,
				Check:                    resource.TestCheckResourceAttr("elasticstack_kibana_security_entity_store.test", "started", "false"),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("started_false"),
				ConfigVariables:          vars,
				PlanOnly:                 true,
			},
		},
	})
}

func TestAccResourceKibanaSecurityEntityStore_historySnapshot(t *testing.T) {
	skipIfUnsupported(t)
	spaceID := sdkacctest.RandStringFromCharSet(12, accTestKibanaSpaceIDCharset)
	vars := config.Variables{"space_id": config.StringVariable(spaceID)}
	t.Cleanup(func() { acctest.CleanupEntityStore(t, spaceID) })

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("history_snapshot"),
				ConfigVariables:          vars,
				Check:                    resource.TestCheckResourceAttrSet("elasticstack_kibana_security_entity_store.test", "id"),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("history_snapshot"),
				ConfigVariables:          vars,
				PlanOnly:                 true,
			},
		},
	})
}

// entityStoreEngineTypePattern matches the entity types the entity store
// resource can install (see entity_types validator in schema.go). Engine
// order in the API response is arbitrary, so assertions key off this enum
// rather than a fixed index-to-type mapping.
var entityStoreEngineTypePattern = regexp.MustCompile(`^(user|host|service|generic)$`)

const statusDS = "data.elasticstack_kibana_security_entity_store_status.test"

func TestAccDataSourceKibanaSecurityEntityStoreStatus_basic(t *testing.T) {
	skipIfUnsupported(t)
	spaceID := sdkacctest.RandStringFromCharSet(12, accTestKibanaSpaceIDCharset)
	vars := config.Variables{"space_id": config.StringVariable(spaceID)}
	t.Cleanup(func() { acctest.CleanupEntityStore(t, spaceID) })

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("default"),
				ConfigVariables:          vars,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(statusDS, "installed", "true"),
					resource.TestCheckResourceAttr(statusDS, "overall_status", "running"),
					resource.TestMatchResourceAttr(statusDS, "engines.#", regexp.MustCompile(`^[1-9][0-9]*$`)),
					resource.TestMatchResourceAttr(statusDS, "engines.0.type", entityStoreEngineTypePattern),
					resource.TestCheckResourceAttrSet(statusDS, "engines.0.status"),
					// A freshly installed engine with no explicit log_extraction
					// config reports a fixed set of server-side defaults before any
					// data has been processed; index_pattern is empty until then.
					resource.TestCheckResourceAttr(statusDS, "engines.0.index_pattern", ""),
					resource.TestCheckResourceAttr(statusDS, "engines.0.field_history_length", "10"),
					resource.TestCheckResourceAttr(statusDS, "engines.0.delay", "1m"),
					resource.TestCheckResourceAttr(statusDS, "engines.0.frequency", "1m"),
					resource.TestCheckResourceAttr(statusDS, "engines.0.lookback_period", "3h"),
					resource.TestCheckNoResourceAttr(statusDS, "engines.0.error_action"),
					resource.TestCheckNoResourceAttr(statusDS, "engines.0.error_message"),
					checkEnginesHaveDistinctTypes(statusDS),
					resource.TestCheckNoResourceAttr(statusDS, "include_components"),
					resource.TestCheckNoResourceAttr(statusDS, "engines.0.components.#"),
					resource.TestCheckResourceAttr(statusDS, "space_id", spaceID),
					resource.TestCheckResourceAttrSet(statusDS, "status_json"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("with_components"),
				ConfigVariables:          vars,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(statusDS, "installed", "true"),
					resource.TestCheckResourceAttr(statusDS, "overall_status", "running"),
					resource.TestMatchResourceAttr(statusDS, "engines.#", regexp.MustCompile(`^[1-9][0-9]*$`)),
					resource.TestMatchResourceAttr(statusDS, "engines.0.type", entityStoreEngineTypePattern),
					resource.TestCheckResourceAttrSet(statusDS, "engines.0.status"),
					checkEnginesHaveDistinctTypes(statusDS),
					resource.TestCheckResourceAttr(statusDS, "include_components", "true"),
					resource.TestMatchResourceAttr(statusDS, "engines.0.components.#", regexp.MustCompile(`^[1-9][0-9]*$`)),
					resource.TestCheckResourceAttrSet(statusDS, "engines.0.components.0.id"),
					resource.TestCheckResourceAttrSet(statusDS, "engines.0.components.0.installed"),
					resource.TestCheckResourceAttrSet(statusDS, "engines.0.components.0.resource"),
					resource.TestCheckResourceAttr(statusDS, "space_id", spaceID),
					resource.TestCheckResourceAttrSet(statusDS, "status_json"),
				),
			},
		},
	})
}

// checkEnginesHaveDistinctTypes asserts that when at least two engines are
// present, their types differ. It guards against a regression where every
// engine reported the same type (e.g. a copy/flatten bug), which a per-index
// enum match alone would not catch.
func checkEnginesHaveDistinctTypes(resName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resName]
		if !ok || rs == nil || rs.Primary == nil {
			return fmt.Errorf("resource %s not found in state", resName)
		}
		countStr, ok := rs.Primary.Attributes["engines.#"]
		if !ok {
			return fmt.Errorf("resource %s has no engines.# attribute", resName)
		}
		count, err := strconv.Atoi(countStr)
		if err != nil {
			return fmt.Errorf("resource %s has non-numeric engines.# attribute %q: %w", resName, countStr, err)
		}
		if count < 2 {
			return nil
		}
		type0 := rs.Primary.Attributes["engines.0.type"]
		type1 := rs.Primary.Attributes["engines.1.type"]
		if type0 == type1 {
			return fmt.Errorf("expected engines.0.type (%q) and engines.1.type (%q) to differ across %d engines", type0, type1, count)
		}
		return nil
	}
}

// TestAccDataSourceKibanaSecurityEntityStoreStatus_defaultSpace exercises the
// space_id optional-unset resolution path: with space_id omitted on both the
// resource and the data source, both should resolve to the Kibana default
// space rather than erroring or defaulting to an empty string.
func TestAccDataSourceKibanaSecurityEntityStoreStatus_defaultSpace(t *testing.T) {
	skipIfUnsupported(t)
	t.Cleanup(func() { acctest.CleanupEntityStore(t, clients.DefaultSpaceID) })

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("default_space"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(statusDS, "space_id", clients.DefaultSpaceID),
					resource.TestCheckResourceAttr(statusDS, "installed", "true"),
				),
			},
		},
	})
}

// TestAccDataSourceKibanaSecurityEntityStoreStatus_notInstalled reads the
// status data source for a space where the entity store resource was never
// created, confirming the not-yet-installed scenario resolves to
// installed = false / overall_status = not_installed instead of erroring.
func TestAccDataSourceKibanaSecurityEntityStoreStatus_notInstalled(t *testing.T) {
	skipIfUnsupported(t)
	spaceID := sdkacctest.RandStringFromCharSet(12, accTestKibanaSpaceIDCharset)
	vars := config.Variables{"space_id": config.StringVariable(spaceID)}
	t.Cleanup(func() { acctest.CleanupEntityStore(t, spaceID) })

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("not_installed"),
				ConfigVariables:          vars,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(statusDS, "installed", "false"),
					resource.TestCheckResourceAttr(statusDS, "overall_status", "not_installed"),
					resource.TestCheckResourceAttr(statusDS, "engines.#", "0"),
					resource.TestCheckResourceAttr(statusDS, "space_id", spaceID),
				),
			},
		},
	})
}

func skipIfUnsupported(t *testing.T) {
	versionutils.SkipIfUnsupported(t, securityentitystore.MinVersion, versionutils.FlavorAny)
}
