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

package exportsavedobjects_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

const dataSourceName = "data.elasticstack_kibana_export_saved_objects.test"

func TestAccDataSourceKibanaExportSavedObjects(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "id"),
					resource.TestCheckResourceAttr(dataSourceName, "space_id", "default"),
					resource.TestCheckResourceAttr(dataSourceName, "exclude_export_details", "true"),
					resource.TestCheckResourceAttr(dataSourceName, "include_references_deep", "true"),
					resource.TestCheckResourceAttr(dataSourceName, "objects.#", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "objects.0.type", "action"),
					resource.TestCheckResourceAttrPair(
						dataSourceName, "objects.0.id",
						"elasticstack_kibana_action_connector.test", "connector_id",
					),
					checkExportedObjectsContains(dataSourceName, "exported_objects", "action"),
				),
			},
		},
	})
}

func TestAccDataSourceKibanaExportSavedObjects_boolOptions(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				ConfigVariables: config.Variables{
					"exclude_export_details":  config.BoolVariable(false),
					"include_references_deep": config.BoolVariable(true),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "id"),
					resource.TestCheckResourceAttr(dataSourceName, "exclude_export_details", "false"),
					resource.TestCheckResourceAttr(dataSourceName, "include_references_deep", "true"),
					resource.TestCheckResourceAttrSet(dataSourceName, "exported_objects"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				ConfigVariables: config.Variables{
					"exclude_export_details":  config.BoolVariable(true),
					"include_references_deep": config.BoolVariable(false),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "id"),
					resource.TestCheckResourceAttr(dataSourceName, "exclude_export_details", "true"),
					resource.TestCheckResourceAttr(dataSourceName, "include_references_deep", "false"),
					resource.TestCheckResourceAttrSet(dataSourceName, "exported_objects"),
				),
			},
		},
	})
}

func TestAccDataSourceKibanaExportSavedObjects_defaultSpaceID(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "space_id", "default"),
					resource.TestCheckResourceAttrSet(dataSourceName, "exported_objects"),
				),
			},
		},
	})
}

func TestAccDataSourceKibanaExportSavedObjects_kibanaConnection(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckWithExplicitKibanaEndpoint(t)
		},
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				ConfigVariables:          acctest.KibanaConnectionVariables(),
				Check: resource.ComposeAggregateTestCheckFunc(
					append([]resource.TestCheckFunc{
						resource.TestCheckResourceAttrSet(dataSourceName, "id"),
						resource.TestCheckResourceAttrSet(dataSourceName, "exported_objects"),
						resource.TestCheckResourceAttr(dataSourceName, "kibana_connection.#", "1"),
						resource.TestCheckResourceAttr(dataSourceName, "kibana_connection.0.endpoints.#", "1"),
						resource.TestCheckResourceAttr(dataSourceName, "kibana_connection.0.insecure", "false"),
					}, acctest.KibanaConnectionAuthChecks(dataSourceName)...)...,
				),
			},
		},
	})
}

func checkExportedObjectsContains(resourceName, attr, substr string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}
		val, ok := rs.Primary.Attributes[attr]
		if !ok {
			return fmt.Errorf("attribute %q not found on %s", attr, resourceName)
		}
		if !strings.Contains(val, substr) {
			return fmt.Errorf("expected %s.%s to contain %q, got: %s", resourceName, attr, substr, val)
		}
		return nil
	}
}
