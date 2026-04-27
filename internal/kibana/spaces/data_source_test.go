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
)

// testImageURL is a minimal 1×1 PNG data-URL used to verify image_url round-trip.
const testImageURL = "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg=="

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
					resource.TestCheckResourceAttr("data.elasticstack_kibana_spaces.all_spaces", "id", "spaces"),
					resource.TestCheckResourceAttr("data.elasticstack_kibana_spaces.all_spaces", "spaces.0.id", "default"),
					resource.TestCheckResourceAttr("data.elasticstack_kibana_spaces.all_spaces", "spaces.0.name", "Default"),
					resource.TestCheckResourceAttr("data.elasticstack_kibana_spaces.all_spaces", "spaces.0.description", "This is your default space!"),
					resource.TestCheckResourceAttr("data.elasticstack_kibana_spaces.all_spaces", "spaces.0.disabled_features.#", "0"),
					resource.TestCheckResourceAttrSet("data.elasticstack_kibana_spaces.all_spaces", "spaces.0.initials"),
					resource.TestCheckResourceAttrSet("data.elasticstack_kibana_spaces.all_spaces", "spaces.0.color"),
					resource.TestCheckResourceAttr("data.elasticstack_kibana_spaces.all_spaces", "spaces.0.image_url", ""),
					resource.TestCheckResourceAttr("data.elasticstack_kibana_spaces.all_spaces", "spaces.0.solution", ""),
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
					resource.TestCheckResourceAttr("data.elasticstack_kibana_spaces.all_spaces", "spaces.0.id", "default"),
					resource.TestCheckResourceAttr("data.elasticstack_kibana_spaces.all_spaces", "spaces.0.name", "Default"),
					resource.TestCheckResourceAttr("data.elasticstack_kibana_spaces.all_spaces", "spaces.0.description", "This is your default space!"),
					// Data source must return at least two spaces.
					resource.TestCheckResourceAttrWith(
						"data.elasticstack_kibana_spaces.all_spaces",
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
					// Custom space (sorts after "default") is at index 1.
					resource.TestCheckResourceAttr("data.elasticstack_kibana_spaces.all_spaces", "spaces.1.id", spaceID),
					resource.TestCheckResourceAttr("data.elasticstack_kibana_spaces.all_spaces", "spaces.1.name", "Test Coverage Space"),
					resource.TestCheckResourceAttr("data.elasticstack_kibana_spaces.all_spaces", "spaces.1.description", "Test space for data source coverage"),
					resource.TestCheckResourceAttr("data.elasticstack_kibana_spaces.all_spaces", "spaces.1.disabled_features.#", "0"),
					resource.TestCheckResourceAttrSet("data.elasticstack_kibana_spaces.all_spaces", "spaces.1.initials"),
					resource.TestCheckResourceAttrSet("data.elasticstack_kibana_spaces.all_spaces", "spaces.1.color"),
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
					resource.TestCheckResourceAttr("data.elasticstack_kibana_spaces.all_spaces", "spaces.0.id", "default"),
					resource.TestCheckResourceAttr("data.elasticstack_kibana_spaces.all_spaces", "spaces.1.id", spaceID),
					resource.TestCheckResourceAttr("data.elasticstack_kibana_spaces.all_spaces", "spaces.1.description", ""),
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
					resource.TestCheckResourceAttr("data.elasticstack_kibana_spaces.all_spaces", "spaces.0.id", "default"),
					resource.TestCheckResourceAttr("data.elasticstack_kibana_spaces.all_spaces", "spaces.1.id", spaceID),
					resource.TestCheckResourceAttr("data.elasticstack_kibana_spaces.all_spaces", "spaces.1.image_url", testImageURL),
				),
			},
		},
	})
}

// TestAccSpacesDataSource_withKibanaConnection verifies that the data source
// correctly reads spaces when an explicit kibana_connection block is provided.
func TestAccSpacesDataSource_withKibanaConnection(t *testing.T) {
	checks := []resource.TestCheckFunc{
		resource.TestCheckResourceAttr("data.elasticstack_kibana_spaces.all_spaces", "id", "spaces"),
		resource.TestCheckResourceAttr("data.elasticstack_kibana_spaces.all_spaces", "spaces.0.id", "default"),
		resource.TestCheckResourceAttr("data.elasticstack_kibana_spaces.all_spaces", "kibana_connection.#", "1"),
		resource.TestCheckResourceAttr("data.elasticstack_kibana_spaces.all_spaces", "kibana_connection.0.endpoints.#", "1"),
		resource.TestCheckResourceAttr("data.elasticstack_kibana_spaces.all_spaces", "kibana_connection.0.endpoints.0", strings.TrimSpace(os.Getenv("KIBANA_ENDPOINT"))),
		resource.TestCheckResourceAttr("data.elasticstack_kibana_spaces.all_spaces", "kibana_connection.0.insecure", "false"),
	}
	checks = append(checks, acctest.KibanaConnectionAuthChecks("data.elasticstack_kibana_spaces.all_spaces")...)

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
