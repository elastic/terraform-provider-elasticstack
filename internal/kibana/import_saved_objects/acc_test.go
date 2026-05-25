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

package importsavedobjects_test

import (
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

var minVersionCompatibilityMode = version.Must(version.NewVersion("8.8.0"))

func TestAccResourceImportSavedObjects(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_import_saved_objects.settings", "success", "true"),
					resource.TestCheckResourceAttr("elasticstack_kibana_import_saved_objects.settings", "success_count", "1"),
					resource.TestCheckResourceAttr("elasticstack_kibana_import_saved_objects.settings", "success_results.#", "1"),
					resource.TestCheckResourceAttr("elasticstack_kibana_import_saved_objects.settings", "errors.#", "0"),
					resource.TestCheckResourceAttr("elasticstack_kibana_import_saved_objects.settings", "overwrite", "true"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_import_saved_objects.settings", "file_contents"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_import_saved_objects.settings", "success_results.0.id"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_import_saved_objects.settings", "success_results.0.type"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_import_saved_objects.settings", "success", "true"),
					resource.TestCheckResourceAttr("elasticstack_kibana_import_saved_objects.settings", "success_count", "1"),
					resource.TestCheckResourceAttr("elasticstack_kibana_import_saved_objects.settings", "success_results.#", "1"),
					resource.TestCheckResourceAttr("elasticstack_kibana_import_saved_objects.settings", "errors.#", "0"),
					resource.TestCheckResourceAttr("elasticstack_kibana_import_saved_objects.settings", "overwrite", "true"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_import_saved_objects.settings", "file_contents"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_import_saved_objects.settings", "success_results.0.id"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_import_saved_objects.settings", "success_results.0.type"),
				),
			},
			{
				// Ensure a partially successful import doesn't throw a provider error
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("missing_ref"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_import_saved_objects.settings", "success", "false"),
					resource.TestCheckResourceAttr("elasticstack_kibana_import_saved_objects.settings", "success_count", "1"),
					resource.TestCheckResourceAttr("elasticstack_kibana_import_saved_objects.settings", "success_results.#", "1"),
					resource.TestCheckResourceAttr("elasticstack_kibana_import_saved_objects.settings", "errors.#", "1"),
					resource.TestCheckResourceAttr("elasticstack_kibana_import_saved_objects.settings", "overwrite", "true"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_import_saved_objects.settings", "errors.0.id"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_import_saved_objects.settings", "errors.0.type"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_import_saved_objects.settings", "errors.0.error.type"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_import_saved_objects.settings", "success_results.0.id"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_import_saved_objects.settings", "success_results.0.type"),
				),
			},
			{
				// Ensure compatibility_mode flag is accepted and import succeeds (requires Kibana 8.8+)
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minVersionCompatibilityMode),
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("compatibility_mode"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_import_saved_objects.settings", "success", "true"),
					resource.TestCheckResourceAttr("elasticstack_kibana_import_saved_objects.settings", "success_count", "1"),
					resource.TestCheckResourceAttr("elasticstack_kibana_import_saved_objects.settings", "errors.#", "0"),
					resource.TestCheckResourceAttr("elasticstack_kibana_import_saved_objects.settings", "compatibility_mode", "true"),
					resource.TestCheckResourceAttr("elasticstack_kibana_import_saved_objects.settings", "overwrite", "true"),
					resource.TestCheckResourceAttr("elasticstack_kibana_import_saved_objects.settings", "success_results.#", "1"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_import_saved_objects.settings", "success_results.0.id"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_import_saved_objects.settings", "success_results.0.type"),
				),
			},
		},
	})
}

func TestAccResourceImportSavedObjects_CreateNewCopies(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_import_saved_objects.settings", "create_new_copies", "true"),
					resource.TestCheckResourceAttr("elasticstack_kibana_import_saved_objects.settings", "success", "true"),
					resource.TestCheckResourceAttr("elasticstack_kibana_import_saved_objects.settings", "success_count", "1"),
					resource.TestCheckResourceAttr("elasticstack_kibana_import_saved_objects.settings", "success_results.#", "1"),
					resource.TestCheckResourceAttr("elasticstack_kibana_import_saved_objects.settings", "errors.#", "0"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_import_saved_objects.settings", "success_results.0.id"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_import_saved_objects.settings", "success_results.0.type"),
				),
			},
		},
	})
}

func TestAccResourceImportSavedObjects_IgnoreImportErrors(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				// Establish the object so the next step triggers a conflict
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("setup"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_import_saved_objects.settings", "success", "true"),
				),
			},
			{
				// Re-import without overwrite: Kibana returns a conflict error.
				// With ignore_import_errors=true the provider must not return a TF error.
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("conflict"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_import_saved_objects.settings", "ignore_import_errors", "true"),
					resource.TestCheckResourceAttr("elasticstack_kibana_import_saved_objects.settings", "success", "false"),
					resource.TestCheckResourceAttr("elasticstack_kibana_import_saved_objects.settings", "errors.#", "1"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_import_saved_objects.settings", "errors.0.id"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_import_saved_objects.settings", "errors.0.type"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_import_saved_objects.settings", "errors.0.error.type"),
				),
			},
		},
	})
}

func TestAccResourceImportSavedObjects_SpaceID(t *testing.T) {
	spaceID := "tf-iso-" + sdkacctest.RandStringFromCharSet(4, "abcdefghijklmnopqrstuvwxyz0123456789")

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("import_in_space"),
				ConfigVariables: config.Variables{
					"space_id": config.StringVariable(spaceID),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_import_saved_objects.settings", "space_id", spaceID),
					resource.TestCheckResourceAttr("elasticstack_kibana_import_saved_objects.settings", "success", "true"),
					resource.TestCheckResourceAttr("elasticstack_kibana_import_saved_objects.settings", "success_count", "1"),
					resource.TestCheckResourceAttr("elasticstack_kibana_import_saved_objects.settings", "success_results.#", "1"),
					resource.TestCheckResourceAttr("elasticstack_kibana_import_saved_objects.settings", "errors.#", "0"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_import_saved_objects.settings", "success_results.0.id"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_import_saved_objects.settings", "success_results.0.type"),
				),
			},
		},
	})
}
