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
	"regexp"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

var minVersionCompatibilityMode = version.Must(version.NewVersion("8.8.0"))

const resourceName = "elasticstack_kibana_import_saved_objects.settings"

func TestAccResourceImportSavedObjects(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("elasticstack_kibana_import_saved_objects.settings", "id"),
					resource.TestCheckResourceAttr("elasticstack_kibana_import_saved_objects.settings", "space_id", "default"),
					resource.TestCheckResourceAttr("elasticstack_kibana_import_saved_objects.settings", "success", "true"),
					resource.TestCheckResourceAttr("elasticstack_kibana_import_saved_objects.settings", "success_count", "1"),
					resource.TestCheckResourceAttr("elasticstack_kibana_import_saved_objects.settings", "success_results.#", "1"),
					resource.TestCheckResourceAttr("elasticstack_kibana_import_saved_objects.settings", "errors.#", "0"),
					resource.TestCheckResourceAttr("elasticstack_kibana_import_saved_objects.settings", "overwrite", "true"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_import_saved_objects.settings", "file_contents"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_import_saved_objects.settings", "success_results.0.id"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_import_saved_objects.settings", "success_results.0.type"),
					resource.TestCheckResourceAttr("elasticstack_kibana_import_saved_objects.settings", "success_results.0.destination_id", ""),
					resource.TestCheckResourceAttr("elasticstack_kibana_import_saved_objects.settings", "success_results.0.meta.title", "Advanced Settings [7.14.0]"),
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
					// Kibana's missing_references error does not populate the top-level
					// "title" field (only errors.0.meta.title is set for this error type).
					resource.TestCheckResourceAttr("elasticstack_kibana_import_saved_objects.settings", "errors.0.title", ""),
					resource.TestCheckResourceAttr("elasticstack_kibana_import_saved_objects.settings", "errors.0.meta.icon", "visualizeApp"),
					resource.TestCheckResourceAttr("elasticstack_kibana_import_saved_objects.settings", "errors.0.meta.title", "healthchecks"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_import_saved_objects.settings", "success_results.0.id"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_import_saved_objects.settings", "success_results.0.type"),
					resource.TestCheckResourceAttr("elasticstack_kibana_import_saved_objects.settings", "success_results.0.destination_id", ""),
					resource.TestCheckResourceAttr("elasticstack_kibana_import_saved_objects.settings", "success_results.0.meta.icon", ""),
					resource.TestCheckResourceAttr("elasticstack_kibana_import_saved_objects.settings", "success_results.0.meta.title", "Advanced Settings [7.14.0]"),
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
					// create_new_copies regenerates the object ID, so Kibana always
					// reports the newly assigned destination_id for the copy.
					resource.TestCheckResourceAttrSet("elasticstack_kibana_import_saved_objects.settings", "success_results.0.destination_id"),
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

// TestAccResourceImportSavedObjects_ConfigValidators exercises the negative path of the
// resource's two ConfigValidators at the acceptance-test level: create_new_copies cannot be
// combined with overwrite, nor with compatibility_mode, when both are explicitly true.
func TestAccResourceImportSavedObjects_ConfigValidators(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create_new_copies_and_overwrite"),
				ExpectError:              regexp.MustCompile(`create_new_copies and overwrite cannot both be set to true`),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create_new_copies_and_compatibility_mode"),
				ExpectError:              regexp.MustCompile(`create_new_copies and compatibility_mode cannot both be set to true`),
			},
		},
	})
}

// TestAccResourceImportSavedObjects_BoolDefaults covers the unset (null, since none of these
// Optional-only attributes has a schema default) state of every optional boolean flag, then
// sets "overwrite" to true in a second step to exercise update coverage for a boolean attribute.
func TestAccResourceImportSavedObjects_BoolDefaults(t *testing.T) {
	objectID := "tf-iso-" + sdkacctest.RandStringFromCharSet(8, "abcdefghijklmnopqrstuvwxyz0123456789")

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("defaults"),
				ConfigVariables: config.Variables{
					"object_id": config.StringVariable(objectID),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckNoResourceAttr(resourceName, "ignore_import_errors"),
					resource.TestCheckNoResourceAttr(resourceName, "create_new_copies"),
					resource.TestCheckNoResourceAttr(resourceName, "overwrite"),
					resource.TestCheckNoResourceAttr(resourceName, "compatibility_mode"),
					resource.TestCheckResourceAttr(resourceName, "success", "true"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("overwrite_true"),
				ConfigVariables: config.Variables{
					"object_id": config.StringVariable(objectID),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "overwrite", "true"),
					resource.TestCheckNoResourceAttr(resourceName, "ignore_import_errors"),
					resource.TestCheckNoResourceAttr(resourceName, "create_new_copies"),
					resource.TestCheckNoResourceAttr(resourceName, "compatibility_mode"),
					resource.TestCheckResourceAttr(resourceName, "success", "true"),
				),
			},
		},
	})
}

// TestAccResourceImportSavedObjects_KibanaConnection exercises the per-resource
// kibana_connection block, which overrides the provider-level Kibana connection.
func TestAccResourceImportSavedObjects_KibanaConnection(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckWithExplicitKibanaEndpoint(t)
		},
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          acctest.KibanaConnectionVariables(),
				Check: resource.ComposeAggregateTestCheckFunc(
					append([]resource.TestCheckFunc{
						resource.TestCheckResourceAttr(resourceName, "success", "true"),
						resource.TestCheckResourceAttr(resourceName, "kibana_connection.#", "1"),
						resource.TestCheckResourceAttrSet(resourceName, "kibana_connection.0.endpoints.0"),
						resource.TestCheckResourceAttr(resourceName, "kibana_connection.0.insecure", "false"),
					}, acctest.KibanaConnectionAuthChecks(resourceName)...)...,
				),
			},
		},
	})
}
