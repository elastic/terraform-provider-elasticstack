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
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
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
					resource.TestCheckResourceAttr("data.elasticstack_fleet_output.test", "outputs.#", "1"),
					resource.TestCheckResourceAttr("data.elasticstack_fleet_output.test", "outputs.0.id", "fleet-default-output"),
					resource.TestCheckResourceAttr("data.elasticstack_fleet_output.test", "outputs.0.name", "default"),
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
					resource.TestCheckResourceAttr("data.elasticstack_fleet_output.test", "outputs.#", "2"),
					resource.TestCheckResourceAttr("data.elasticstack_fleet_output.test", "outputs.0.name", "default"),
					resource.TestCheckResourceAttr("data.elasticstack_fleet_output.test", "outputs.1.name", "test"),
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
					resource.TestCheckResourceAttr("data.elasticstack_fleet_output.test", "outputs.#", "1"),
					resource.TestCheckResourceAttr("data.elasticstack_fleet_output.test", "outputs.0.id", "fleet-default-output"),
					resource.TestCheckResourceAttr("data.elasticstack_fleet_output.test", "outputs.0.name", "default"),
				),
			},
		},
	})
}
