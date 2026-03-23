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
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const (
	providerConfig = `
provider "elasticstack" {
	kibana {}
}
`
)

var (
	minKibanaParameterAPIVersion = version.Must(version.NewVersion("8.12.0"))
)

func TestSyntheticParameterResource(t *testing.T) {
	resourceID := "elasticstack_kibana_synthetics_parameter.test"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minKibanaParameterAPIVersion),
				Config: providerConfig + `
resource "elasticstack_kibana_synthetics_parameter" "test" {
	key = "test-key"
	value = "test-value"
	description = "Test description"
	tags = ["a", "b"]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceID, "key", "test-key"),
					resource.TestCheckResourceAttr(resourceID, "value", "test-value"),
					resource.TestCheckResourceAttr(resourceID, "description", "Test description"),
					resource.TestCheckResourceAttr(resourceID, "tags.#", "2"),
					resource.TestCheckResourceAttr(resourceID, "tags.0", "a"),
					resource.TestCheckResourceAttr(resourceID, "tags.1", "b"),
				),
			},
			// ImportState testing
			{
				SkipFunc:          versionutils.CheckIfVersionIsUnsupported(minKibanaParameterAPIVersion),
				ResourceName:      resourceID,
				ImportState:       true,
				ImportStateVerify: true,
				Config: providerConfig + `
resource "elasticstack_kibana_synthetics_parameter" "test" {
	key = "test-key"
	value = "test-value"
	description = "Test description"
	tags = ["a", "b"]
}
`,
			},
			// Update and Read testing
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minKibanaParameterAPIVersion),
				Config: providerConfig + `
resource "elasticstack_kibana_synthetics_parameter" "test" {
	key = "test-key-2"
	value = "test-value-2"
	description = "Test description 2"
	tags = ["c", "d", "e"]
}
`,
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
