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

package parameter_test

import (
	"fmt"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

var (
	minKibanaParameterAPIVersion = version.Must(version.NewVersion("8.12.0"))
)

func TestSyntheticParameterResource(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minKibanaParameterAPIVersion, versionutils.FlavorAny)

	resourceID := "elasticstack_kibana_synthetics_parameter.test"
	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceID, "key", "test-key"),
					resource.TestCheckResourceAttr(resourceID, "value", "test-value"),
					resource.TestCheckResourceAttr(resourceID, "description", "Test description"),
					resource.TestCheckResourceAttr(resourceID, "tags.#", "2"),
					resource.TestCheckResourceAttr(resourceID, "tags.0", "a"),
					resource.TestCheckResourceAttr(resourceID, "tags.1", "b"),
				),
			},
			// ImportState testing — also verifies value round-trips correctly from the API.
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ResourceName:             resourceID,
				ImportState:              true,
				ImportStateVerify:        true,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("import"),
			},
			// Defaults: omit description and tags to verify empty defaults are applied.
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("defaults"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceID, "key", "test-key"),
					resource.TestCheckResourceAttr(resourceID, "value", "test-value"),
					resource.TestCheckResourceAttr(resourceID, "description", ""),
					resource.TestCheckResourceAttr(resourceID, "tags.#", "0"),
				),
			},
			// Update and Read testing
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceID, "key", "test-key-2"),
					resource.TestCheckResourceAttr(resourceID, "value", "test-value-2"),
					resource.TestCheckResourceAttr(resourceID, "description", "Test description 2"),
					resource.TestCheckResourceAttr(resourceID, "tags.#", "3"),
					resource.TestCheckResourceAttr(resourceID, "tags.0", "c"),
					resource.TestCheckResourceAttr(resourceID, "tags.1", "d"),
					resource.TestCheckResourceAttr(resourceID, "tags.2", "e"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func TestSyntheticParameterResource_SharedAcrossSpaces(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minKibanaParameterAPIVersion, versionutils.FlavorAny)

	resourceID := "elasticstack_kibana_synthetics_parameter.test"
	var firstID string
	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			// Create with share_across_spaces = true and assert the attribute.
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("shared"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceID, "key", "test-key-shared"),
					resource.TestCheckResourceAttr(resourceID, "value", "test-value-shared"),
					resource.TestCheckResourceAttr(resourceID, "share_across_spaces", "true"),
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources[resourceID]
						if !ok {
							return fmt.Errorf("resource not found: %s", resourceID)
						}
						firstID = rs.Primary.ID
						return nil
					},
				),
			},
			// Change share_across_spaces to false — triggers RequiresReplace; new resource should have a different ID.
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("not_shared"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceID, "share_across_spaces", "false"),
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources[resourceID]
						if !ok {
							return fmt.Errorf("resource not found: %s", resourceID)
						}
						if rs.Primary.ID == firstID {
							return fmt.Errorf("expected new resource id after share_across_spaces change (RequiresReplace), got same id: %s", firstID)
						}
						return nil
					},
				),
			},
		},
	})
}
