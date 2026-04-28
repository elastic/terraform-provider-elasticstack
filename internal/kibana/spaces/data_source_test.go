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

package spaces_test

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// testImageURL is a minimal 1×1 PNG data-URL used to verify image_url round-trip.
const testImageURL = "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg=="

// testSpacesResourceName is the Terraform resource address used across all
// spaces data source acceptance tests.
const testSpacesResourceName = "data.elasticstack_kibana_spaces.all_spaces"

// testCheckSpaceAttrByID returns a TestCheckFunc that scans the "spaces" list
// in state to find the element whose id equals spaceID, then asserts that attr
// equals value. This avoids hard-coding list indices, which can shift when the
// Kibana API returns spaces in a different order.
func testCheckSpaceAttrByID(spaceID, attr, value string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[testSpacesResourceName]
		if !ok {
			return fmt.Errorf("resource %q not found in state", testSpacesResourceName)
		}
		attrs := rs.Primary.Attributes
		countStr, ok := attrs["spaces.#"]
		if !ok {
			return fmt.Errorf("%q: spaces.# not found in state", testSpacesResourceName)
		}
		count, err := strconv.Atoi(countStr)
		if err != nil {
			return fmt.Errorf("%q: spaces.# is not a number: %w", testSpacesResourceName, err)
		}
		for i := range count {
			if attrs[fmt.Sprintf("spaces.%d.id", i)] == spaceID {
				got := attrs[fmt.Sprintf("spaces.%d.%s", i, attr)]
				if got != value {
					return fmt.Errorf("%q spaces[id=%q].%s: expected %q, got %q", testSpacesResourceName, spaceID, attr, value, got)
				}
				return nil
			}
		}
		return fmt.Errorf("%q: no space with id %q found in state", testSpacesResourceName, spaceID)
	}
}

// TestAccSpacesDataSource verifies the data source returns all expected fields
// for the pre-existing default space.
func TestAccSpacesDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(testSpacesResourceName, "id", "spaces"),
					resource.TestCheckResourceAttr(testSpacesResourceName, "spaces.0.id", "default"),
					resource.TestCheckResourceAttr(testSpacesResourceName, "spaces.0.name", "Default"),
					resource.TestCheckResourceAttr(testSpacesResourceName, "spaces.0.description", "This is your default space!"),
					resource.TestCheckResourceAttr(testSpacesResourceName, "spaces.0.disabled_features.#", "0"),
					resource.TestCheckResourceAttrSet(testSpacesResourceName, "spaces.0.color"),
					resource.TestCheckResourceAttr(testSpacesResourceName, "spaces.0.image_url", ""),
					resource.TestCheckResourceAttr(testSpacesResourceName, "spaces.0.solution", ""),
				),
			},
		},
	})
}

// TestAccSpacesDataSource_multipleSpaces verifies the data source returns
// attributes from more than one space when a second space has been created.
// The custom space ID uses "tfacc" prefix (t > d) so it reliably sorts after
// "default" in the list returned by the Kibana API.
func TestAccSpacesDataSource_multipleSpaces(t *testing.T) {
	spaceID := "tfacc" + sdkacctest.RandStringFromCharSet(17, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				ConfigVariables: config.Variables{
					"space_id": config.StringVariable(spaceID),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					// Default space is always the first element.
					resource.TestCheckResourceAttr(testSpacesResourceName, "spaces.0.id", "default"),
					resource.TestCheckResourceAttr(testSpacesResourceName, "spaces.0.name", "Default"),
					resource.TestCheckResourceAttr(testSpacesResourceName, "spaces.0.description", "This is your default space!"),
					// Data source must return at least two spaces.
					resource.TestCheckResourceAttrWith(
						testSpacesResourceName,
						"spaces.#",
						func(value string) error {
							count, err := strconv.Atoi(value)
							if err != nil {
								return err
							}
							if count < 2 {
								return fmt.Errorf("expected at least 2 spaces, got %d", count)
							}
							return nil
						},
					),
					// Custom space — looked up by ID to avoid index-ordering fragility.
					testCheckSpaceAttrByID(spaceID, "name", "Test Coverage Space"),
					testCheckSpaceAttrByID(spaceID, "description", "Test space for data source coverage"),
					testCheckSpaceAttrByID(spaceID, "disabled_features.#", "0"),
					testCheckSpaceAttrByID(spaceID, "initials", "TC"),
					testCheckSpaceAttrByID(spaceID, "color", "#E8478B"),
				),
			},
		},
	})
}

// TestAccSpacesDataSource_noDescription verifies that a space with no
// description is returned with an empty description string.
func TestAccSpacesDataSource_noDescription(t *testing.T) {
	spaceID := "tfacc" + sdkacctest.RandStringFromCharSet(17, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				ConfigVariables: config.Variables{
					"space_id": config.StringVariable(spaceID),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(testSpacesResourceName, "spaces.0.id", "default"),
					testCheckSpaceAttrByID(spaceID, "description", ""),
				),
			},
		},
	})
}

// TestAccSpacesDataSource_withImageURL verifies that a space configured with a
// custom avatar image URL returns that value from the data source.
func TestAccSpacesDataSource_withImageURL(t *testing.T) {
	spaceID := "tfacc" + sdkacctest.RandStringFromCharSet(17, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				ConfigVariables: config.Variables{
					"space_id": config.StringVariable(spaceID),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(testSpacesResourceName, "spaces.0.id", "default"),
					testCheckSpaceAttrByID(spaceID, "image_url", testImageURL),
				),
			},
		},
	})
}

// TestAccSpacesDataSource_withKibanaConnection verifies that the data source
// correctly reads spaces when an explicit kibana_connection block is provided.
func TestAccSpacesDataSource_withKibanaConnection(t *testing.T) {
	checks := []resource.TestCheckFunc{
		resource.TestCheckResourceAttr(testSpacesResourceName, "id", "spaces"),
		resource.TestCheckResourceAttr(testSpacesResourceName, "spaces.0.id", "default"),
		resource.TestCheckResourceAttr(testSpacesResourceName, "kibana_connection.#", "1"),
		resource.TestCheckResourceAttr(testSpacesResourceName, "kibana_connection.0.endpoints.#", "1"),
		resource.TestCheckResourceAttr(testSpacesResourceName, "kibana_connection.0.endpoints.0", strings.TrimSpace(os.Getenv("KIBANA_ENDPOINT"))),
		resource.TestCheckResourceAttr(testSpacesResourceName, "kibana_connection.0.insecure", "false"),
	}
	checks = append(checks, acctest.KibanaConnectionAuthChecks(testSpacesResourceName)...)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckWithExplicitKibanaEndpoint(t)
		},
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				ConfigVariables:          acctest.KibanaConnectionVariables(config.Variables{}),
				Check:                    resource.ComposeAggregateTestCheckFunc(checks...),
			},
		},
	})
}
