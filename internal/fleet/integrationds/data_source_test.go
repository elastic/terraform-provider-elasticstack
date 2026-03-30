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

package integrationds_test

import (
	"fmt"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/hashicorp/go-version"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

var (
	minVersionIntegrationDataSource        = version.Must(version.NewVersion("8.6.0"))
	minVersionIntegrationDataSourceSpaceID = version.Must(version.NewVersion("9.1.0"))
)

func TestAccDataSourceIntegration(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minVersionIntegrationDataSource),
				Config:   testAccDataSourceIntegration,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.elasticstack_fleet_integration.test", "name", "tcp"),
					checkResourceAttrStringNotEmpty("data.elasticstack_fleet_integration.test", "version"),
				),
			},
		},
	})
}

const testAccDataSourceIntegration = `
provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

data "elasticstack_fleet_integration" "test" {
  name = "tcp"
}
`

func TestAccDataSourceIntegrationWithSpaceID(t *testing.T) {
	spaceName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)
	spaceID := fmt.Sprintf("space-%s", spaceName)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minVersionIntegrationDataSourceSpaceID),
				Config:   testAccDataSourceIntegrationWithSpaceID(spaceID, spaceName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.elasticstack_fleet_integration.test", "name", "tcp"),
					resource.TestCheckResourceAttr("data.elasticstack_fleet_integration.test", "space_id", spaceID),
					checkResourceAttrStringNotEmpty("data.elasticstack_fleet_integration.test", "version"),
				),
			},
		},
	})
}

func testAccDataSourceIntegrationWithSpaceID(spaceID, spaceName string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_kibana_space" "test_space" {
  space_id    = %q
  name        = %q
  description = "Test space for Fleet integration data source space_id test"
}

data "elasticstack_fleet_integration" "test" {
  name     = "tcp"
  space_id = elasticstack_kibana_space.test_space.space_id

  depends_on = [elasticstack_kibana_space.test_space]
}
`, spaceID, spaceName)
}

// checkResourceAttrStringNotEmpty verifies that the string value at key
// is not empty.
func checkResourceAttrStringNotEmpty(name, key string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		ms := s.RootModule()
		rs, ok := ms.Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s in %s", name, ms.Path)
		}
		is := rs.Primary
		if is == nil {
			return fmt.Errorf("no primary instance: %s in %s", name, ms.Path)
		}

		v, ok := is.Attributes[key]
		if !ok {
			return fmt.Errorf("%s: Attribute '%s' not found", name, key)
		}
		if v == "" {
			return fmt.Errorf("%s: Attribute '%s' expected non-empty string", name, key)
		}

		return nil
	}
}
