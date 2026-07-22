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
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

const accTestKibanaSpaceIDCharset = "abcdefghijklmnopqrstuvwxyz0123456789_-"

var minVersionEntityStoreResolution = version.Must(version.NewVersion("9.4.0"))

func TestAccResourceSecurityEntityStoreEntityLink(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minVersionEntityStoreResolution, versionutils.FlavorAny)
	spaceID := sdkacctest.RandStringFromCharSet(12, accTestKibanaSpaceIDCharset)
	t.Cleanup(func() { acctest.CleanupEntityStore(t, spaceID) })

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          config.Variables{"space_id": config.StringVariable(spaceID)},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_security_entity_store_entity_link.test", "target_id", "generic:acc-test-target"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_entity_store_entity_link.test", "space_id", spaceID),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_entity_store_entity_link.test", "id", spaceID+"/generic:acc-test-target"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_security_entity_store_entity_link.test", "resolution_group_json"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_entity_store_entity_link.test", "entity_ids.#", "2"),
					resource.TestCheckTypeSetElemAttr("elasticstack_kibana_security_entity_store_entity_link.test", "entity_ids.*", "generic:acc-test-alias1"),
					resource.TestCheckTypeSetElemAttr("elasticstack_kibana_security_entity_store_entity_link.test", "entity_ids.*", "generic:acc-test-alias2"),
					resource.TestCheckResourceAttrWith("elasticstack_kibana_security_entity_store_entity_link.test", "resolution_group_json", checkResolutionGroupJSON(
						"generic:acc-test-target",
						[]string{"generic:acc-test-alias1", "generic:acc-test-alias2"},
					)),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables:          config.Variables{"space_id": config.StringVariable(spaceID)},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_security_entity_store_entity_link.test", "target_id", "generic:acc-test-target"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_entity_store_entity_link.test", "space_id", spaceID),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_entity_store_entity_link.test", "id", spaceID+"/generic:acc-test-target"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_security_entity_store_entity_link.test", "resolution_group_json"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_entity_store_entity_link.test", "entity_ids.#", "2"),
					resource.TestCheckTypeSetElemAttr("elasticstack_kibana_security_entity_store_entity_link.test", "entity_ids.*", "generic:acc-test-alias1"),
					resource.TestCheckTypeSetElemAttr("elasticstack_kibana_security_entity_store_entity_link.test", "entity_ids.*", "generic:acc-test-alias3"),
					checkSetElemAbsent("elasticstack_kibana_security_entity_store_entity_link.test", "entity_ids", "generic:acc-test-alias2"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          config.Variables{"space_id": config.StringVariable(spaceID)},
				ResourceName:             "elasticstack_kibana_security_entity_store_entity_link.test",
				ImportState:              true,
				ImportStateVerify:        true,
			},
		},
	})
}

func TestAccResourceSecurityEntityStoreEntityLink_SingleElement(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minVersionEntityStoreResolution, versionutils.FlavorAny)
	spaceID := sdkacctest.RandStringFromCharSet(12, accTestKibanaSpaceIDCharset)
	t.Cleanup(func() { acctest.CleanupEntityStore(t, spaceID) })

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("single_element"),
				ConfigVariables:          config.Variables{"space_id": config.StringVariable(spaceID)},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_security_entity_store_entity_link.test", "target_id", "generic:acc-test-target"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_entity_store_entity_link.test", "space_id", spaceID),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_entity_store_entity_link.test", "entity_ids.#", "1"),
					resource.TestCheckTypeSetElemAttr("elasticstack_kibana_security_entity_store_entity_link.test", "entity_ids.*", "generic:acc-test-alias1"),
				),
			},
		},
	})
}

// checkResolutionGroupJSON returns a check function that parses the JSON value
// and asserts that the alias IDs appear within the payload's aliases list.
func checkResolutionGroupJSON(_ string, aliasIDs []string) func(string) error {
	return func(v string) error {
		var payload map[string]any
		if err := json.Unmarshal([]byte(v), &payload); err != nil {
			return fmt.Errorf("resolution_group_json is not valid JSON: %w", err)
		}

		aliases, _ := payload["aliases"].([]any)
		seen := make(map[string]bool)
		for _, a := range aliases {
			aliasMap, ok := a.(map[string]any)
			if !ok {
				continue
			}
			entityMap, _ := aliasMap["entity"].(map[string]any)
			if id, ok := entityMap["id"].(string); ok {
				seen[id] = true
			}
		}

		for _, id := range aliasIDs {
			if !seen[id] {
				return fmt.Errorf("resolution_group_json: alias %q not found in aliases list", id)
			}
		}

		return nil
	}
}

// checkSetElemAbsent asserts that value is not present in the set attribute of
// the named resource.
func checkSetElemAbsent(name, attr, value string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("resource %q not found in state", name)
		}

		prefix := attr + "."
		for k, v := range rs.Primary.Attributes {
			if strings.HasPrefix(k, prefix) && v == value {
				return fmt.Errorf("expected %q to be absent from %s but found it (key %s)", value, attr, k)
			}
		}

		return nil
	}
}
