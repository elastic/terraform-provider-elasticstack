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
	"context"
	"fmt"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/acctest/checks"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
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

// TestAccResourceSpace_ClearEmptyFields pins the fix for issue #1881:
// without the configuredString helper SDKv2's d.GetOk treats `description = ""`
// as "attribute unset" and omits it from the PUT body, so Kibana silently
// retains the previous description. Terraform state would then read "" while
// Kibana still held the original value, which surfaced as drift on every
// subsequent plan and an apparent inability to clear the field at all.
//
// This test creates a space with a non-empty description, updates the config
// to set description = "", and asserts via the Kibana Spaces API that the
// stored description is actually empty. State-only assertions wouldn't catch
// the regression because the bug was specifically that state and Kibana
// disagreed.
func TestAccResourceSpace_ClearEmptyFields(t *testing.T) {
	spaceID := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceSpaceDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("with_description"),
				ConfigVariables: config.Variables{
					"space_id": config.StringVariable(spaceID),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_space.test_space", "description", "initial description"),
					checkResourceSpaceAPIDescription(spaceID, "initial description"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("empty_description"),
				ConfigVariables: config.Variables{
					"space_id": config.StringVariable(spaceID),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_space.test_space", "description", ""),
					checkResourceSpaceAPIDescription(spaceID, ""),
				),
			},
		},
	})
}

// checkResourceSpaceAPIDescription queries the Kibana Spaces API directly and
// asserts the stored description matches expected. Used by
// TestAccResourceSpace_ClearEmptyFields to prove the empty-string update
// actually reached Kibana, not just terraform state.
func checkResourceSpaceAPIDescription(spaceID, expected string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		scopedClient, err := clients.NewAcceptanceTestingKibanaScopedClient()
		if err != nil {
			return err
		}
		oapiClient, err := scopedClient.GetKibanaOapiClient()
		if err != nil {
			return err
		}
		space, diags := kibanaoapi.GetSpace(context.Background(), oapiClient, spaceID)
		if diags.HasError() {
			return fmt.Errorf("error fetching space %q: %s", spaceID, diags[0].Detail())
		}
		if space == nil {
			return fmt.Errorf("space %q not found via Kibana Spaces API", spaceID)
		}
		got := ""
		if space.Description != nil {
			got = *space.Description
		}
		if got != expected {
			return fmt.Errorf("Kibana API description for space %q: got %q, want %q", spaceID, got, expected)
		}
		return nil
	}
}

var checkResourceSpaceDestroy = checks.KibanaResourceDestroyCheck(
	"elasticstack_kibana_space",
	func(ctx context.Context, client *kibanaoapi.Client, id string) (bool, error) {
		space, diags := kibanaoapi.GetSpace(ctx, client, id)
		if diags.HasError() {
			return false, fmt.Errorf("error checking space destroy: %s", diags[0].Detail())
		}
		return space != nil, nil
	},
)
