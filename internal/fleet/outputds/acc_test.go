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
	"regexp"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

var minVersionOutput = version.Must(version.NewVersion("8.6.0"))

func TestAccDataSourceOutput(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minVersionOutput),
				Config:   testAccDataSourceOutput,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.elasticstack_fleet_output.test", "name", "default"),
					resource.TestCheckResourceAttrSet("data.elasticstack_fleet_output.test", "id"),
				),
			},
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minVersionOutput),
				Config:   testAccDataSourceOutputSpace,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.elasticstack_fleet_output.test", "name", "default"),
					resource.TestCheckResourceAttrSet("data.elasticstack_fleet_output.test", "id"),
				),
			},
			{
				SkipFunc:    versionutils.CheckIfVersionIsUnsupported(minVersionOutput),
				Config:      testAccDataSourceOutputMissing,
				ExpectError: regexp.MustCompile("Output not found"),
			},
		},
	})
}

const testAccDataSourceOutput = `
provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

data "elasticstack_fleet_output" "test" {
  name = "default"
}
`

const testAccDataSourceOutputSpace = `
provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

data "elasticstack_fleet_output" "test" {
  name = "default"
  space_id = "default"
}
`

const testAccDataSourceOutputMissing = `
provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

data "elasticstack_fleet_output" "test" {
  name = "missing"
}
`
