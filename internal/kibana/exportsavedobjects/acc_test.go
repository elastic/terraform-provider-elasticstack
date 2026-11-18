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
	"regexp"
	"strings"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
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
					resource.TestCheckResourceAttr(dataSourceName, "id", "default/export"),
					resource.TestCheckResourceAttr(dataSourceName, "space_id", "default"),
					resource.TestCheckResourceAttr(dataSourceName, "exclude_export_details", "true"),
					resource.TestCheckResourceAttr(dataSourceName, "include_references_deep", "true"),
					resource.TestCheckResourceAttr(dataSourceName, "objects.#", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "objects.0.type", "action"),
					resource.TestCheckResourceAttrPair(
						dataSourceName, "objects.0.id",
						"elasticstack_kibana_action_connector.test", "connector_id",
					),
					checkExportedObjectsContains("action"),
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
					checkExportedObjectsContains("exportedCount"),
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
					checkExportedObjectsNotContains("exportedCount"),
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

func TestAccDataSourceKibanaExportSavedObjects_boolOptionsDefault(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "id"),
					resource.TestCheckResourceAttr(dataSourceName, "exclude_export_details", "true"),
					resource.TestCheckResourceAttr(dataSourceName, "include_references_deep", "true"),
					resource.TestCheckResourceAttrSet(dataSourceName, "exported_objects"),
				),
			},
		},
	})
}

func TestAccDataSourceKibanaExportSavedObjects_emptyObjects(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				ExpectError:              regexp.MustCompile(`(?s)at least 1`),
			},
		},
	})
}

func TestAccDataSourceKibanaExportSavedObjects_multipleObjects(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "id"),
					resource.TestCheckResourceAttr(dataSourceName, "objects.#", "2"),
					resource.TestCheckResourceAttr(dataSourceName, "objects.0.type", "action"),
					resource.TestCheckResourceAttrPair(
						dataSourceName, "objects.0.id",
						"elasticstack_kibana_action_connector.test", "connector_id",
					),
					resource.TestCheckResourceAttr(dataSourceName, "objects.1.type", "action"),
					resource.TestCheckResourceAttrPair(
						dataSourceName, "objects.1.id",
						"elasticstack_kibana_action_connector.test2", "connector_id",
					),
					checkExportedObjectsContains(".slack"),
					checkExportedObjectsContains(".server-log"),
				),
			},
		},
	})
}

func TestAccDataSourceKibanaExportSavedObjects_customSpace(t *testing.T) {
	spaceID := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)

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
					resource.TestCheckResourceAttr(dataSourceName, "id", fmt.Sprintf("%s/export", spaceID)),
					resource.TestCheckResourceAttr(dataSourceName, "space_id", spaceID),
					resource.TestCheckResourceAttr(dataSourceName, "objects.#", "1"),
					resource.TestCheckResourceAttrPair(
						dataSourceName, "objects.0.id",
						"elasticstack_kibana_action_connector.test", "connector_id",
					),
					checkExportedObjectsContains("action"),
				),
			},
		},
	})
}

const exportedObjectsAttr = "exported_objects"

func checkExportedObjectsContains(substr string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[dataSourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", dataSourceName)
		}
		val, ok := rs.Primary.Attributes[exportedObjectsAttr]
		if !ok {
			return fmt.Errorf("attribute %q not found on %s", exportedObjectsAttr, dataSourceName)
		}
		if !strings.Contains(val, substr) {
			return fmt.Errorf("expected %s.%s to contain %q, got: %s", dataSourceName, exportedObjectsAttr, substr, val)
		}
		return nil
	}
}

func checkExportedObjectsNotContains(substr string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[dataSourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", dataSourceName)
		}
		val, ok := rs.Primary.Attributes[exportedObjectsAttr]
		if !ok {
			return fmt.Errorf("attribute %q not found on %s", exportedObjectsAttr, dataSourceName)
		}
		if strings.Contains(val, substr) {
			return fmt.Errorf("expected %s.%s to not contain %q, got: %s", dataSourceName, exportedObjectsAttr, substr, val)
		}
		return nil
	}
}
