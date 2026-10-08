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

package defaultdataview_test

import (
	"regexp"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
)

var minDataViewAPISupport = version.Must(version.NewVersion("8.1.0"))

func TestAccResourceDefaultDataView(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minDataViewAPISupport, versionutils.FlavorAny)

	indexName1 := "my-index-" + sdkacctest.RandStringFromCharSet(4, sdkacctest.CharSetAlphaNum)
	indexName2 := "my-other-index-" + sdkacctest.RandStringFromCharSet(4, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("basic"),
				ConfigVariables: config.Variables{
					"index_name": config.StringVariable(indexName1),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_default_data_view.test", "id", "default"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_default_data_view.test", "data_view_id"),
					resource.TestCheckResourceAttrPair("elasticstack_kibana_default_data_view.test", "data_view_id", "elasticstack_kibana_data_view.dv", "data_view.id"),
					resource.TestCheckResourceAttr("elasticstack_kibana_default_data_view.test", "force", "true"),
					resource.TestCheckResourceAttr("elasticstack_kibana_default_data_view.test", "skip_delete", "false"),
					resource.TestCheckResourceAttr("elasticstack_kibana_default_data_view.test", "space_id", "default"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables: config.Variables{
					"index_name1": config.StringVariable(indexName1),
					"index_name2": config.StringVariable(indexName2),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_default_data_view.test", "id", "default"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_default_data_view.test", "data_view_id"),
					resource.TestCheckResourceAttrPair("elasticstack_kibana_default_data_view.test", "data_view_id", "elasticstack_kibana_data_view.dv2", "data_view.id"),
					resource.TestCheckResourceAttr("elasticstack_kibana_default_data_view.test", "space_id", "default"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("unset"),
				ConfigVariables: config.Variables{
					"index_name1": config.StringVariable(indexName1),
					"index_name2": config.StringVariable(indexName2),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_default_data_view.test", "id", "default"),
					resource.TestCheckNoResourceAttr("elasticstack_kibana_default_data_view.test", "data_view_id"),
					resource.TestCheckResourceAttr("elasticstack_kibana_default_data_view.test", "space_id", "default"),
				),
			},
		},
	})
}

func TestAccResourceDefaultDataViewWithSkipDelete(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minDataViewAPISupport, versionutils.FlavorAny)

	indexName := "my-index-" + sdkacctest.RandStringFromCharSet(4, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				// skip_delete starts at false so the subsequent step exercises an
				// in-place false -> true update rather than only a single fixed value.
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("skip_delete_false"),
				ConfigVariables: config.Variables{
					"index_name": config.StringVariable(indexName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_default_data_view.test", "id", "default"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_default_data_view.test", "data_view_id"),
					resource.TestCheckResourceAttrPair("elasticstack_kibana_default_data_view.test", "data_view_id", "elasticstack_kibana_data_view.dv", "data_view.id"),
					resource.TestCheckResourceAttr("elasticstack_kibana_default_data_view.test", "force", "true"),
					resource.TestCheckResourceAttr("elasticstack_kibana_default_data_view.test", "skip_delete", "false"),
					resource.TestCheckResourceAttr("elasticstack_kibana_default_data_view.test", "space_id", "default"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("skip_delete"),
				ConfigVariables: config.Variables{
					"index_name": config.StringVariable(indexName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_default_data_view.test", "id", "default"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_default_data_view.test", "data_view_id"),
					resource.TestCheckResourceAttrPair("elasticstack_kibana_default_data_view.test", "data_view_id", "elasticstack_kibana_data_view.dv", "data_view.id"),
					resource.TestCheckResourceAttr("elasticstack_kibana_default_data_view.test", "force", "true"),
					resource.TestCheckResourceAttr("elasticstack_kibana_default_data_view.test", "skip_delete", "true"),
					resource.TestCheckResourceAttr("elasticstack_kibana_default_data_view.test", "space_id", "default"),
				),
			},
		},
	})
}

func TestAccResourceDefaultDataViewWithCustomSpace(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minDataViewAPISupport, versionutils.FlavorAny)

	indexName := "my-index-" + sdkacctest.RandStringFromCharSet(4, sdkacctest.CharSetAlphaNum)
	spaceID := "test-space-" + sdkacctest.RandStringFromCharSet(6, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("custom_space"),
				ConfigVariables: config.Variables{
					"index_name": config.StringVariable(indexName),
					"space_id":   config.StringVariable(spaceID),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_default_data_view.test", "id", spaceID),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_default_data_view.test", "data_view_id"),
					resource.TestCheckResourceAttr("elasticstack_kibana_default_data_view.test", "force", "true"),
					resource.TestCheckResourceAttr("elasticstack_kibana_default_data_view.test", "skip_delete", "false"),
					resource.TestCheckResourceAttr("elasticstack_kibana_default_data_view.test", "space_id", spaceID),
				),
			},
		},
	})
}

func TestAccResourceDefaultDataViewForceFalse(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minDataViewAPISupport, versionutils.FlavorAny)

	indexName := "my-index-" + sdkacctest.RandStringFromCharSet(4, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("force_true"),
				ConfigVariables: config.Variables{
					"index_name": config.StringVariable(indexName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_default_data_view.test", "id", "default"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_default_data_view.test", "data_view_id"),
					resource.TestCheckResourceAttrPair("elasticstack_kibana_default_data_view.test", "data_view_id", "elasticstack_kibana_data_view.dv", "data_view.id"),
					resource.TestCheckResourceAttr("elasticstack_kibana_default_data_view.test", "force", "true"),
					resource.TestCheckResourceAttr("elasticstack_kibana_default_data_view.test", "space_id", "default"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("force_false"),
				ConfigVariables: config.Variables{
					"index_name": config.StringVariable(indexName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_default_data_view.test", "id", "default"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_default_data_view.test", "data_view_id"),
					resource.TestCheckResourceAttrPair("elasticstack_kibana_default_data_view.test", "data_view_id", "elasticstack_kibana_data_view.dv", "data_view.id"),
					resource.TestCheckResourceAttr("elasticstack_kibana_default_data_view.test", "force", "false"),
					resource.TestCheckResourceAttr("elasticstack_kibana_default_data_view.test", "space_id", "default"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("force_nil"),
				ConfigVariables: config.Variables{
					"index_name": config.StringVariable(indexName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_default_data_view.test", "id", "default"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_default_data_view.test", "data_view_id"),
					resource.TestCheckResourceAttrPair("elasticstack_kibana_default_data_view.test", "data_view_id", "elasticstack_kibana_data_view.dv", "data_view.id"),
					resource.TestCheckNoResourceAttr("elasticstack_kibana_default_data_view.test", "force"),
					resource.TestCheckResourceAttr("elasticstack_kibana_default_data_view.test", "space_id", "default"),
				),
			},
		},
	})
}

func TestAccResourceDefaultDataViewInvalidDataViewID(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minDataViewAPISupport, versionutils.FlavorAny)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("empty_data_view_id"),
				PlanOnly:                 true,
				ExpectError:              regexp.MustCompile(`(?i)string length must be at least`),
			},
		},
	})
}

// TestAccResourceDefaultDataViewSpaceIDForcesReplace verifies that changing
// space_id on an existing default data view resource actually triggers the
// RequiresReplace plan modifier, not just that the resource defaults to and
// reflects a given space_id.
func TestAccResourceDefaultDataViewSpaceIDForcesReplace(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minDataViewAPISupport, versionutils.FlavorAny)

	spaceAID := "test-space-a-" + sdkacctest.RandStringFromCharSet(6, sdkacctest.CharSetAlphaNum)
	spaceBID := "test-space-b-" + sdkacctest.RandStringFromCharSet(6, sdkacctest.CharSetAlphaNum)
	resourceID := "elasticstack_kibana_default_data_view.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("force_replace"),
				ConfigVariables: config.Variables{
					"space_a_id":      config.StringVariable(spaceAID),
					"space_b_id":      config.StringVariable(spaceBID),
					"active_space_id": config.StringVariable(spaceAID),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceID, "id", spaceAID),
					resource.TestCheckResourceAttr(resourceID, "space_id", spaceAID),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("force_replace"),
				ConfigVariables: config.Variables{
					"space_a_id":      config.StringVariable(spaceAID),
					"space_b_id":      config.StringVariable(spaceBID),
					"active_space_id": config.StringVariable(spaceBID),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceID, plancheck.ResourceActionDestroyBeforeCreate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceID, "id", spaceBID),
					resource.TestCheckResourceAttr(resourceID, "space_id", spaceBID),
				),
			},
		},
	})
}

// TestAccResourceDefaultDataViewKibanaConnection exercises the kibana_connection
// override block, which is injected by the shared Kibana resource envelope but
// had no dedicated fixture for this resource.
func TestAccResourceDefaultDataViewKibanaConnection(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minDataViewAPISupport, versionutils.FlavorAny)

	indexName := "my-index-" + sdkacctest.RandStringFromCharSet(4, sdkacctest.CharSetAlphaNum)
	resourceID := "elasticstack_kibana_default_data_view.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckWithExplicitKibanaEndpoint(t)
		},
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("kibana_connection"),
				ConfigVariables: acctest.KibanaConnectionVariables(config.Variables{
					"index_name": config.StringVariable(indexName),
				}),
				Check: defaultDataViewKibanaConnectionChecks(resourceID),
			},
		},
	})
}

func defaultDataViewKibanaConnectionChecks(resourceID string) resource.TestCheckFunc {
	checks := []resource.TestCheckFunc{
		resource.TestCheckResourceAttr(resourceID, "id", "default"),
		resource.TestCheckResourceAttrSet(resourceID, "data_view_id"),
		resource.TestCheckResourceAttr(resourceID, "force", "true"),
		resource.TestCheckResourceAttr(resourceID, "kibana_connection.#", "1"),
		resource.TestCheckResourceAttr(resourceID, "kibana_connection.0.endpoints.0", acctest.KibanaConnectionEndpoint()),
		resource.TestCheckResourceAttr(resourceID, "kibana_connection.0.insecure", "false"),
	}
	checks = append(checks, acctest.KibanaConnectionAuthChecks(resourceID)...)

	return resource.ComposeAggregateTestCheckFunc(checks...)
}

// TestAccResourceDefaultDataViewWithTimeouts exercises the timeouts block
// injected by the shared Kibana resource envelope, which had no dedicated
// fixture for this resource.
func TestAccResourceDefaultDataViewWithTimeouts(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minDataViewAPISupport, versionutils.FlavorAny)

	indexName := "my-index-" + sdkacctest.RandStringFromCharSet(4, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("with_timeouts"),
				ConfigVariables: config.Variables{
					"index_name": config.StringVariable(indexName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_default_data_view.test", "id", "default"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_default_data_view.test", "data_view_id"),
					resource.TestCheckResourceAttr("elasticstack_kibana_default_data_view.test", "timeouts.create", "5m"),
					resource.TestCheckResourceAttr("elasticstack_kibana_default_data_view.test", "timeouts.delete", "5m"),
				),
			},
		},
	})
}
