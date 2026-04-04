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

package privatelocation_test

// this test is in synthetics_test package, because of https://github.com/elastic/kibana/issues/190801
// having both tests in same package allows to use mutex in kibana API client and workaround the issue

import (
	"fmt"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

var (
	minKibanaPrivateLocationAPIVersion = version.Must(version.NewVersion("8.12.0"))
)

// accTestKibanaSpaceIDCharset matches elasticstack_kibana_space space_id validation (^[a-z0-9_-]+$).
const accTestKibanaSpaceIDCharset = "abcdefghijklmnopqrstuvwxyz0123456789_-"

func TestSyntheticPrivateLocationResource(t *testing.T) {
	resourceID := "elasticstack_kibana_synthetics_private_location.test"
	randomSuffix := sdkacctest.RandStringFromCharSet(4, sdkacctest.CharSetAlphaNum)
	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minKibanaPrivateLocationAPIVersion),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"suffix": config.StringVariable(randomSuffix),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceID, "space_id", ""),
					resource.TestCheckResourceAttr(resourceID, "label", fmt.Sprintf("pl-test-label-%s", randomSuffix)),
					resource.TestCheckResourceAttrSet(resourceID, "agent_policy_id"),
					resource.TestCheckResourceAttr(resourceID, "tags.#", "2"),
					resource.TestCheckResourceAttr(resourceID, "tags.0", "a"),
					resource.TestCheckResourceAttr(resourceID, "tags.1", "b"),
					resource.TestCheckResourceAttr(resourceID, "geo.lat", "42.42"),
					resource.TestCheckResourceAttr(resourceID, "geo.lon", "-42.42"),
				),
			},
			// ImportState testing
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minKibanaPrivateLocationAPIVersion),
				ResourceName:             resourceID,
				ImportState:              true,
				ImportStateVerify:        true,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("import"),
				ConfigVariables: config.Variables{
					"suffix": config.StringVariable(randomSuffix),
				},
			},
			// Update and Read testing
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minKibanaPrivateLocationAPIVersion),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update_with_geo"),
				ConfigVariables: config.Variables{
					"suffix": config.StringVariable(randomSuffix),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceID, "label", fmt.Sprintf("pl-test-label-2-%s", randomSuffix)),
					resource.TestCheckResourceAttrSet(resourceID, "agent_policy_id"),
					resource.TestCheckResourceAttr(resourceID, "tags.#", "3"),
					resource.TestCheckResourceAttr(resourceID, "tags.0", "c"),
					resource.TestCheckResourceAttr(resourceID, "tags.1", "d"),
					resource.TestCheckResourceAttr(resourceID, "tags.2", "e"),
					resource.TestCheckResourceAttr(resourceID, "geo.lat", "-33.21"),
					resource.TestCheckResourceAttr(resourceID, "geo.lon", "-33.21"),
				),
			},
			// Update and Read testing
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minKibanaPrivateLocationAPIVersion),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update_no_optional"),
				ConfigVariables: config.Variables{
					"suffix": config.StringVariable(randomSuffix),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceID, "label", fmt.Sprintf("pl-test-label-2-%s", randomSuffix)),
					resource.TestCheckResourceAttrSet(resourceID, "agent_policy_id"),
					resource.TestCheckNoResourceAttr(resourceID, "tags"),
					resource.TestCheckNoResourceAttr(resourceID, "geo"),
				),
			},
			// Update and Read testing
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minKibanaPrivateLocationAPIVersion),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update_tags_only"),
				ConfigVariables: config.Variables{
					"suffix": config.StringVariable(randomSuffix),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceID, "label", fmt.Sprintf("pl-test-label-2-%s", randomSuffix)),
					resource.TestCheckResourceAttrSet(resourceID, "agent_policy_id"),
					resource.TestCheckResourceAttr(resourceID, "tags.#", "3"),
					resource.TestCheckResourceAttr(resourceID, "tags.0", "c"),
					resource.TestCheckResourceAttr(resourceID, "tags.1", "d"),
					resource.TestCheckResourceAttr(resourceID, "tags.2", "e"),
					resource.TestCheckNoResourceAttr(resourceID, "geo"),
				),
			},
			// Update and Read testing
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minKibanaPrivateLocationAPIVersion),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update_geo_only"),
				ConfigVariables: config.Variables{
					"suffix": config.StringVariable(randomSuffix),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceID, "label", fmt.Sprintf("pl-test-label-2-%s", randomSuffix)),
					resource.TestCheckResourceAttrSet(resourceID, "agent_policy_id"),
					resource.TestCheckNoResourceAttr(resourceID, "tags"),
					resource.TestCheckResourceAttr(resourceID, "geo.lat", "-33.21"),
					resource.TestCheckResourceAttr(resourceID, "geo.lon", "-33.21"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func TestSyntheticPrivateLocationResource_nonDefaultSpace(t *testing.T) {
	resourceID := "elasticstack_kibana_synthetics_private_location.test"
	randomSuffix := sdkacctest.RandStringFromCharSet(4, sdkacctest.CharSetAlphaNum)
	spaceID := sdkacctest.RandStringFromCharSet(12, accTestKibanaSpaceIDCharset)
	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minKibanaPrivateLocationAPIVersion),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create_in_space"),
				ConfigVariables: config.Variables{
					"suffix":   config.StringVariable(randomSuffix),
					"space_id": config.StringVariable(spaceID),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceID, "space_id", spaceID),
					resource.TestCheckResourceAttr(resourceID, "label", fmt.Sprintf("pl-test-label-space-%s", randomSuffix)),
					resource.TestCheckResourceAttrSet(resourceID, "agent_policy_id"),
					resource.TestCheckResourceAttr(resourceID, "tags.#", "2"),
					resource.TestCheckResourceAttr(resourceID, "tags.0", "a"),
					resource.TestCheckResourceAttr(resourceID, "tags.1", "b"),
					resource.TestCheckResourceAttr(resourceID, "geo.lat", "42.42"),
					resource.TestCheckResourceAttr(resourceID, "geo.lon", "-42.42"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minKibanaPrivateLocationAPIVersion),
				ResourceName:             resourceID,
				ImportState:              true,
				ImportStateVerify:        true,
				ImportStateVerifyIgnore:  []string{"id"},
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources[resourceID]
					if !ok {
						return "", fmt.Errorf("resource not found: %s", resourceID)
					}
					id := rs.Primary.Attributes["id"]
					return fmt.Sprintf("%s/%s", spaceID, id), nil
				},
				ConfigDirectory: acctest.NamedTestCaseDirectory("create_in_space"),
				ConfigVariables: config.Variables{
					"suffix":   config.StringVariable(randomSuffix),
					"space_id": config.StringVariable(spaceID),
				},
			},
		},
	})
}
