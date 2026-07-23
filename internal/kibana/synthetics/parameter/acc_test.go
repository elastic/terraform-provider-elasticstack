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
	"context"
	"fmt"
	"net/http"
	"regexp"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanautil"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

var (
	minKibanaParameterAPIVersion = version.Must(version.NewVersion("8.12.0"))
)

// accTestKibanaSpaceIDCharset matches elasticstack_kibana_space space_id validation (^[a-z0-9_-]+$).
const accTestKibanaSpaceIDCharset = "abcdefghijklmnopqrstuvwxyz0123456789_-"

func TestSyntheticParameterResource(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minKibanaParameterAPIVersion, versionutils.FlavorAny)

	resourceID := "elasticstack_kibana_synthetics_parameter.test"
	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceID, "key", "test-key"),
					resource.TestCheckResourceAttr(resourceID, "value", "test-value"),
					resource.TestCheckResourceAttr(resourceID, "description", "Test description"),
					resource.TestCheckResourceAttr(resourceID, "tags.#", "2"),
					resource.TestCheckResourceAttr(resourceID, "tags.0", "a"),
					resource.TestCheckResourceAttr(resourceID, "tags.1", "b"),
					resource.TestCheckResourceAttr(resourceID, "space_id", clients.DefaultSpaceID),
					testAccCheckCompositeIDFormat(resourceID, clients.DefaultSpaceID),
					testAccCheckParameterExistsInKibanaSpace(resourceID, clients.DefaultSpaceID),
				),
			},
			// Import by bare parameter UUID (default space).
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ResourceName:             resourceID,
				ImportState:              true,
				ImportStateVerify:        true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					return testAccParameterResourceUUIDFromState(s, resourceID)
				},
				ConfigDirectory: acctest.NamedTestCaseDirectory("import"),
			},
			// Import by composite `<space_id>/<uuid>`.
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ResourceName:             resourceID,
				ImportState:              true,
				ImportStateVerify:        true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources[resourceID]
					if !ok {
						return "", fmt.Errorf("resource not found: %s", resourceID)
					}
					return rs.Primary.ID, nil
				},
				ConfigDirectory: acctest.NamedTestCaseDirectory("import"),
			},
			// Defaults: omit description and tags to verify empty defaults are applied.
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("defaults"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceID, "key", "test-key"),
					resource.TestCheckResourceAttr(resourceID, "value", "test-value"),
					resource.TestCheckResourceAttr(resourceID, "description", ""),
					resource.TestCheckResourceAttr(resourceID, "tags.#", "0"),
					resource.TestCheckResourceAttr(resourceID, "space_id", clients.DefaultSpaceID),
					testAccCheckCompositeIDFormat(resourceID, clients.DefaultSpaceID),
				),
			},
			// Update and Read testing
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
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

func TestSyntheticParameterResource_SharedAcrossSpaces(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minKibanaParameterAPIVersion, versionutils.FlavorAny)

	resourceID := "elasticstack_kibana_synthetics_parameter.test"
	var firstID string
	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			// Create with share_across_spaces = true and assert the attribute.
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("shared"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceID, "key", "test-key-shared"),
					resource.TestCheckResourceAttr(resourceID, "value", "test-value-shared"),
					resource.TestCheckResourceAttr(resourceID, "share_across_spaces", "true"),
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources[resourceID]
						if !ok {
							return fmt.Errorf("resource not found: %s", resourceID)
						}
						firstID = rs.Primary.ID
						return nil
					},
				),
			},
			// Change share_across_spaces to false — triggers RequiresReplace; new resource should have a different ID.
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("not_shared"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceID, "share_across_spaces", "false"),
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources[resourceID]
						if !ok {
							return fmt.Errorf("resource not found: %s", resourceID)
						}
						if rs.Primary.ID == firstID {
							return fmt.Errorf("expected new resource id after share_across_spaces change (RequiresReplace), got same id: %s", firstID)
						}
						return nil
					},
				),
			},
		},
	})
}

func TestSyntheticParameterResource_nonDefaultSpace(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minKibanaParameterAPIVersion, versionutils.FlavorAny)

	resourceID := "elasticstack_kibana_synthetics_parameter.test"
	suffix := sdkacctest.RandStringFromCharSet(8, sdkacctest.CharSetAlphaNum)
	spaceID := sdkacctest.RandStringFromCharSet(12, accTestKibanaSpaceIDCharset)
	vars := config.Variables{
		"suffix":   config.StringVariable(suffix),
		"space_id": config.StringVariable(spaceID),
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          vars,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceID, "space_id", spaceID),
					resource.TestCheckResourceAttr(resourceID, "key", fmt.Sprintf("test-key-space-%s", suffix)),
					resource.TestCheckResourceAttr(resourceID, "value", "test-value-space"),
					resource.TestMatchResourceAttr(resourceID, "id", regexp.MustCompile("^"+regexp.QuoteMeta(spaceID)+"/")),
					testAccCheckCompositeIDFormat(resourceID, spaceID),
					testAccCheckParameterExistsInKibanaSpace(resourceID, spaceID),
					testAccCheckParameterAbsentFromKibanaSpace(resourceID, clients.DefaultSpaceID),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables:          vars,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceID, "space_id", spaceID),
					resource.TestCheckResourceAttr(resourceID, "key", fmt.Sprintf("test-key-space-updated-%s", suffix)),
					resource.TestCheckResourceAttr(resourceID, "value", "test-value-space-updated"),
					testAccCheckParameterExistsInKibanaSpace(resourceID, spaceID),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ResourceName:             resourceID,
				ImportState:              true,
				ImportStateVerify:        true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources[resourceID]
					if !ok {
						return "", fmt.Errorf("resource not found: %s", resourceID)
					}
					return rs.Primary.ID, nil
				},
				ConfigDirectory: acctest.NamedTestCaseDirectory("import"),
				ConfigVariables: vars,
			},
		},
	})
}

func testAccParameterResourceUUIDFromState(s *terraform.State, resourceAddr string) (string, error) {
	rs, ok := s.RootModule().Resources[resourceAddr]
	if !ok {
		return "", fmt.Errorf("resource not found: %s", resourceAddr)
	}
	compID, diags := clients.CompositeIDFromStr(rs.Primary.ID)
	if diags.HasError() {
		return "", fmt.Errorf("parse composite id %q: %v", rs.Primary.ID, diags)
	}
	return compID.ResourceID, nil
}

func testAccCheckCompositeIDFormat(resourceAddr, spaceID string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceAddr]
		if !ok {
			return fmt.Errorf("resource %q not found in state", resourceAddr)
		}

		compID, diags := clients.CompositeIDFromStr(rs.Primary.ID)
		if diags.HasError() {
			return fmt.Errorf("parse composite id %q: %v", rs.Primary.ID, diags)
		}
		if compID.ClusterID != spaceID {
			return fmt.Errorf("resource %q space segment %q, want %q", resourceAddr, compID.ClusterID, spaceID)
		}
		if rs.Primary.Attributes["space_id"] != spaceID {
			return fmt.Errorf("resource %q space_id %q, want %q", resourceAddr, rs.Primary.Attributes["space_id"], spaceID)
		}
		return nil
	}
}

func testAccCheckParameterExistsInKibanaSpace(resourceAddr, spaceID string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		paramUUID, err := testAccParameterResourceUUIDFromState(s, resourceAddr)
		if err != nil {
			return err
		}
		return testAccGetParameter(spaceID, paramUUID, true)
	}
}

func testAccCheckParameterAbsentFromKibanaSpace(resourceAddr, spaceID string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		paramUUID, err := testAccParameterResourceUUIDFromState(s, resourceAddr)
		if err != nil {
			return err
		}
		return testAccGetParameter(spaceID, paramUUID, false)
	}
}

func testAccGetParameter(spaceID, paramUUID string, wantExists bool) error {
	apiClient, err := clients.NewAcceptanceTestingKibanaScopedClient()
	if err != nil {
		return err
	}
	kibanaClient := apiClient.GetKibanaOapiClient()
	resp, err := kibanaClient.API.GetParameterWithResponse(
		context.Background(),
		paramUUID,
		kibanautil.SpaceAwarePathRequestEditor(spaceID),
	)
	if err != nil {
		return fmt.Errorf("get parameter %q in space %q: %w", paramUUID, spaceID, err)
	}
	status := http.StatusInternalServerError
	if resp.HTTPResponse != nil {
		status = resp.HTTPResponse.StatusCode
	}
	exists := resp.JSON200 != nil && status == http.StatusOK
	if wantExists && !exists {
		return fmt.Errorf("expected parameter %q to exist in Kibana space %q (status %d)", paramUUID, spaceID, status)
	}
	if !wantExists && exists {
		return fmt.Errorf("expected parameter %q to be absent from Kibana space %q", paramUUID, spaceID)
	}
	return nil
}
