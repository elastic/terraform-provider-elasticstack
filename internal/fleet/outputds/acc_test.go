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

package outputds_test

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

var minVersionOutput = version.Must(version.NewVersion("8.6.0"))

func TestAccDataSourceOutputDefault(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minVersionOutput),
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("data"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.elasticstack_fleet_output.test", "id", "outputs"),
					testCheckResourceOutputsMinCount("data.elasticstack_fleet_output.test", 1),
					testCheckResourceHasOutput("data.elasticstack_fleet_output.test", map[string]string{
						"id":   "fleet-default-output",
						"name": "default",
					}),
				),
			},
		},
	})
}

func TestAccDataSourceOutputCustomSpace(t *testing.T) {
	spaceName := "test-" + sdkacctest.RandStringFromCharSet(8, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minVersionOutput),
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"space_name": config.StringVariable(spaceName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_space.test", "name", spaceName),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test", "name", "test"),
				),
			},
			{
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minVersionOutput),
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("data"),
				ConfigVariables: config.Variables{
					"space_name": config.StringVariable(spaceName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_space.test", "name", spaceName),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test", "name", "test"),
					resource.TestCheckResourceAttr("data.elasticstack_fleet_output.test", "id", "outputs"),
					testCheckResourceOutputsMinCount("data.elasticstack_fleet_output.test", 2),
					testCheckResourceHasOutput("data.elasticstack_fleet_output.test", map[string]string{
						"id":   "fleet-default-output",
						"name": "default",
					}),
					testCheckResourceHasOutputAttrPair("data.elasticstack_fleet_output.test", "elasticstack_fleet_output.test", map[string]string{
						"name": "test",
						"id":   "id",
					}),
				),
			},
		},
	})
}

func TestAccDataSourceOutputMissingSpace(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minVersionOutput),
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("data"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.elasticstack_fleet_output.test", "id", "outputs"),
					testCheckResourceOutputsMinCount("data.elasticstack_fleet_output.test", 1),
					testCheckResourceHasOutput("data.elasticstack_fleet_output.test", map[string]string{
						"id":   "fleet-default-output",
						"name": "default",
					}),
				),
			},
		},
	})
}

func testCheckResourceOutputsMinCount(resourceName string, minCount int) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource %q not found in state", resourceName)
		}

		rawCount, ok := rs.Primary.Attributes["outputs.#"]
		if !ok {
			return fmt.Errorf("resource %q has no outputs count in state", resourceName)
		}

		count, err := strconv.Atoi(rawCount)
		if err != nil {
			return fmt.Errorf("resource %q has invalid outputs count %q: %w", resourceName, rawCount, err)
		}

		if count < minCount {
			return fmt.Errorf("resource %q expected at least %d outputs, got %d", resourceName, minCount, count)
		}

		return nil
	}
}

func testCheckResourceHasOutput(resourceName string, expected map[string]string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		return hasMatchingOutput(state, resourceName, expected)
	}
}

func testCheckResourceHasOutputAttrPair(dataResourceName, sourceResourceName string, expected map[string]string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		sourceResource, ok := state.RootModule().Resources[sourceResourceName]
		if !ok {
			return fmt.Errorf("resource %q not found in state", sourceResourceName)
		}

		matched := make(map[string]string, len(expected))
		for key, value := range expected {
			if key == "id" {
				sourceValue, ok := sourceResource.Primary.Attributes[value]
				if !ok {
					return fmt.Errorf("resource %q attribute %q not found in state", sourceResourceName, value)
				}
				matched[key] = sourceValue
				continue
			}
			matched[key] = value
		}

		return hasMatchingOutput(state, dataResourceName, matched)
	}
}

func hasMatchingOutput(state *terraform.State, resourceName string, expected map[string]string) error {
	rs, ok := state.RootModule().Resources[resourceName]
	if !ok {
		return fmt.Errorf("resource %q not found in state", resourceName)
	}

	rawCount, ok := rs.Primary.Attributes["outputs.#"]
	if !ok {
		return fmt.Errorf("resource %q has no outputs count in state", resourceName)
	}

	count, err := strconv.Atoi(rawCount)
	if err != nil {
		return fmt.Errorf("resource %q has invalid outputs count %q: %w", resourceName, rawCount, err)
	}

	for i := range count {
		matches := true
		for key, value := range expected {
			attrKey := fmt.Sprintf("outputs.%d.%s", i, key)
			if rs.Primary.Attributes[attrKey] != value {
				matches = false
				break
			}
		}
		if matches {
			return nil
		}
	}

	return fmt.Errorf("resource %q has no output matching %v", resourceName, expected)
}
