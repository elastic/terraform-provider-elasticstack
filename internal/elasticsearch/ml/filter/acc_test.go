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

package filter_test

import (
	"fmt"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccResourceMLFilter(t *testing.T) {
	filterID := fmt.Sprintf("test-filter-%s", sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum))

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"filter_id": config.StringVariable(filterID),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_filter.test", "filter_id", filterID),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_filter.test", "description", "Safe domains filter"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_filter.test", "items.#", "2"),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_ml_filter.test", "id"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables: config.Variables{
					"filter_id": config.StringVariable(filterID),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_filter.test", "filter_id", filterID),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_filter.test", "description", "Updated safe domains filter"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_filter.test", "items.#", "3"),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_ml_filter.test", "id"),
				),
			},
		},
	})
}

func TestAccResourceMLFilterNoItems(t *testing.T) {
	filterID := fmt.Sprintf("test-filter-noitems-%s", sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum))

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"filter_id": config.StringVariable(filterID),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_filter.test", "filter_id", filterID),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_filter.test", "description", "Empty filter"),
					resource.TestCheckNoResourceAttr("elasticstack_elasticsearch_ml_filter.test", "items.#"),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_ml_filter.test", "id"),
				),
			},
		},
	})
}

func TestAccResourceMLFilterImport(t *testing.T) {
	filterID := fmt.Sprintf("test-filter-import-%s", sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum))

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"filter_id": config.StringVariable(filterID),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_filter.test", "filter_id", filterID),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_filter.test", "description", "Filter for import test"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ResourceName:             "elasticstack_elasticsearch_ml_filter.test",
				ImportState:              true,
				ImportStateVerify:        true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs := s.RootModule().Resources["elasticstack_elasticsearch_ml_filter.test"]
					return rs.Primary.ID, nil
				},
				ConfigVariables: config.Variables{
					"filter_id": config.StringVariable(filterID),
				},
			},
		},
	})
}
