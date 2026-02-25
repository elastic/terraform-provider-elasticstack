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

package kibana_test

import (
	"fmt"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

var minSelfManagedVersionForSpaceSolution = version.Must(version.NewVersion("8.18.0"))

func TestAccResourceSpace(t *testing.T) {
	spaceID := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceSpaceDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"space_id": config.StringVariable(spaceID),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_space.test_space", "space_id", spaceID),
					resource.TestCheckResourceAttr("elasticstack_kibana_space.test_space", "name", fmt.Sprintf("Name %s", spaceID)),
					resource.TestCheckResourceAttr("elasticstack_kibana_space.test_space", "description", "Test Space"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables: config.Variables{
					"space_id": config.StringVariable(spaceID),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_space.test_space", "space_id", spaceID),
					resource.TestCheckResourceAttr("elasticstack_kibana_space.test_space", "name", fmt.Sprintf("Updated %s", spaceID)),
					resource.TestCheckResourceAttr("elasticstack_kibana_space.test_space", "description", "Updated space description"),
					resource.TestCheckTypeSetElemAttr("elasticstack_kibana_space.test_space", "disabled_features.*", "ingestManager"),
					resource.TestCheckTypeSetElemAttr("elasticstack_kibana_space.test_space", "disabled_features.*", "enterpriseSearch"),
					resource.TestCheckResourceAttr("elasticstack_kibana_space.test_space", "color", "#FFFFFF"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_space.test_space", "image_url"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("with_solution"),
				ConfigVariables: config.Variables{
					"space_id": config.StringVariable(spaceID),
				},
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minSelfManagedVersionForSpaceSolution),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_space.test_space", "space_id", spaceID),
					resource.TestCheckResourceAttr("elasticstack_kibana_space.test_space", "name", fmt.Sprintf("Solution %s", spaceID)),
					resource.TestCheckResourceAttr("elasticstack_kibana_space.test_space", "description", "Test Space with Solution"),
					resource.TestCheckResourceAttr("elasticstack_kibana_space.test_space", "solution", "security"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"space_id": config.StringVariable(spaceID),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_space.test_space", "space_id", spaceID),
					resource.TestCheckResourceAttr("elasticstack_kibana_space.test_space", "name", fmt.Sprintf("Name %s", spaceID)),
					resource.TestCheckResourceAttr("elasticstack_kibana_space.test_space", "description", "Test Space"),
					resource.TestCheckResourceAttr("elasticstack_kibana_space.test_space", "color", "#FFFFFF"),
				),
			},
		},
	})
}

func checkResourceSpaceDestroy(s *terraform.State) error {
	client, err := clients.NewAcceptanceTestingClient()
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "elasticstack_kibana_space" {
			continue
		}

		kibanaClient, err := client.GetKibanaClient()
		if err != nil {
			return err
		}
		res, err := kibanaClient.KibanaSpaces.Get(rs.Primary.ID)
		if err != nil {
			return err
		}

		if res != nil {
			return fmt.Errorf("Space (%s) still exists", rs.Primary.ID)
		}
	}
	return nil
}
