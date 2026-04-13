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

package agentdownloadsource_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

var minVersionFleetAgentDownloadSource = version.Must(version.NewVersion("8.13.0"))

func TestAccResourceFleetAgentDownloadSource(t *testing.T) {
	random := sdkacctest.RandString(8)
	var idBeforeReplacement string

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceFleetAgentDownloadSourceDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minVersionFleetAgentDownloadSource),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"suffix": config.StringVariable(random),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_download_source.test", "name", fmt.Sprintf("Agent Download Source %s", random)),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_download_source.test", "default", "false"),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_download_source.test", "host", "https://artifacts.elastic.co/downloads/elastic-agent"),
					resource.TestCheckNoResourceAttr("elasticstack_fleet_agent_download_source.test", "proxy_id"),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_download_source.test", "source_id", fmt.Sprintf("agent-download-source-%s", random)),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_download_source.test", "space_ids.#", "1"),
					resource.TestCheckTypeSetElemAttr("elasticstack_fleet_agent_download_source.test", "space_ids.*", "default"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minVersionFleetAgentDownloadSource),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables: config.Variables{
					"suffix": config.StringVariable(random),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_download_source.test", "name", fmt.Sprintf("Updated Agent Download Source %s", random)),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_download_source.test", "default", "true"),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_download_source.test", "host", "https://artifacts.elastic.co/downloads/elastic-agent-updated"),
					resource.TestCheckNoResourceAttr("elasticstack_fleet_agent_download_source.test", "proxy_id"),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_download_source.test", "source_id", fmt.Sprintf("agent-download-source-%s", random)),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_download_source.test", "space_ids.#", "1"),
					resource.TestCheckTypeSetElemAttr("elasticstack_fleet_agent_download_source.test", "space_ids.*", "default"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minVersionFleetAgentDownloadSource),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables: config.Variables{
					"suffix": config.StringVariable(random),
				},
				ResourceName:      "elasticstack_fleet_agent_download_source.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					res := s.RootModule().Resources["elasticstack_fleet_agent_download_source.test"]
					if res == nil || res.Primary == nil {
						return "", fmt.Errorf("resource elasticstack_fleet_agent_download_source.test not found in state")
					}
					return fmt.Sprintf("default/%s", res.Primary.ID), nil
				},
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minVersionFleetAgentDownloadSource),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables: config.Variables{
					"suffix": config.StringVariable(random),
				},
				ResourceName:      "elasticstack_fleet_agent_download_source.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					res := s.RootModule().Resources["elasticstack_fleet_agent_download_source.test"]
					if res == nil || res.Primary == nil {
						return "", fmt.Errorf("resource elasticstack_fleet_agent_download_source.test not found in state")
					}
					return res.Primary.ID, nil
				},
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minVersionFleetAgentDownloadSource),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("omit_optionals"),
				ConfigVariables: config.Variables{
					"suffix": config.StringVariable(random),
				},
				Check: resource.ComposeTestCheckFunc(
					testCheckFleetAgentDownloadSourceCaptureID("elasticstack_fleet_agent_download_source.test", &idBeforeReplacement),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_download_source.test", "name", fmt.Sprintf("No Optionals Agent Download Source %s", random)),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_download_source.test", "default", "false"),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_download_source.test", "host", "https://artifacts.elastic.co/downloads/elastic-agent-no-optionals"),
					resource.TestCheckNoResourceAttr("elasticstack_fleet_agent_download_source.test", "proxy_id"),
					resource.TestCheckResourceAttrSet("elasticstack_fleet_agent_download_source.test", "source_id"),
					resource.TestCheckResourceAttrPair("elasticstack_fleet_agent_download_source.test", "id", "elasticstack_fleet_agent_download_source.test", "source_id"),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_download_source.test", "space_ids.#", "1"),
					resource.TestCheckTypeSetElemAttr("elasticstack_fleet_agent_download_source.test", "space_ids.*", "default"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minVersionFleetAgentDownloadSource),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("empty_space_ids"),
				ConfigVariables: config.Variables{
					"suffix": config.StringVariable(random),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_download_source.test", "name", fmt.Sprintf("Empty Space IDs Agent Download Source %s", random)),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_download_source.test", "space_ids.#", "0"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minVersionFleetAgentDownloadSource),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("replace_source_id"),
				ConfigVariables: config.Variables{
					"suffix": config.StringVariable(random),
				},
				Check: resource.ComposeTestCheckFunc(
					testCheckFleetAgentDownloadSourceIDChanged("elasticstack_fleet_agent_download_source.test", &idBeforeReplacement),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_download_source.test", "source_id", fmt.Sprintf("agent-download-source-replaced-%s", random)),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minVersionFleetAgentDownloadSource),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("non_default_space"),
				ConfigVariables: config.Variables{
					"suffix":               config.StringVariable(random),
					"non_default_space_id": config.StringVariable(fmt.Sprintf("fleet-agent-download-source-%s", random)),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_download_source.test", "space_ids.#", "1"),
					testCheckFleetAgentDownloadSourceSpaceContains("elasticstack_fleet_agent_download_source.test", fmt.Sprintf("fleet-agent-download-source-%s", random)),
				),
			},
		},
	})
}

func testCheckFleetAgentDownloadSourceCaptureID(resourceName string, target *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		res, ok := s.RootModule().Resources[resourceName]
		if !ok || res.Primary == nil {
			return fmt.Errorf("resource %s not found in state", resourceName)
		}
		*target = res.Primary.ID
		return nil
	}
}

func testCheckFleetAgentDownloadSourceIDChanged(resourceName string, previousID *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		res, ok := s.RootModule().Resources[resourceName]
		if !ok || res.Primary == nil {
			return fmt.Errorf("resource %s not found in state", resourceName)
		}
		if *previousID == "" {
			return fmt.Errorf("previous ID was not captured")
		}
		if res.Primary.ID == *previousID {
			return fmt.Errorf("expected resource ID to change after source_id replacement, but remained %q", res.Primary.ID)
		}
		return nil
	}
}

func testCheckFleetAgentDownloadSourceSpaceContains(resourceName, spaceID string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		res, ok := s.RootModule().Resources[resourceName]
		if !ok || res.Primary == nil {
			return fmt.Errorf("resource %s not found in state", resourceName)
		}
		for k, v := range res.Primary.Attributes {
			if k == "space_ids.#" {
				continue
			}
			if strings.HasPrefix(k, "space_ids.") && v == spaceID {
				return nil
			}
		}
		return fmt.Errorf("expected space_ids to contain %q", spaceID)
	}
}

func checkResourceFleetAgentDownloadSourceDestroy(s *terraform.State) error {
	client, err := clients.NewAcceptanceTestingClient()
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "elasticstack_fleet_agent_download_source" {
			continue
		}

		fleetClient, err := client.GetFleetClient()
		if err != nil {
			return err
		}
		spaceID := getOperationalSpaceFromResourceState(rs)
		resp, diags := fleet.GetAgentDownloadSource(context.Background(), fleetClient, rs.Primary.ID, spaceID)
		if diags.HasError() {
			return diagutil.FwDiagsAsError(diags)
		}
		if resp != nil && resp.JSON200 != nil {
			return fmt.Errorf("fleet agent download source id=%v still exists, but it should have been removed", rs.Primary.ID)
		}
	}
	return nil
}

func getOperationalSpaceFromResourceState(rs *terraform.ResourceState) string {
	for k, v := range rs.Primary.Attributes {
		if strings.HasPrefix(k, "space_ids.") && k != "space_ids.#" && v != "" {
			return v
		}
	}
	return ""
}
