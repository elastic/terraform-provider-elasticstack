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
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

var minVersionFleetAgentDownloadSource = version.Must(version.NewVersion("8.13.0"))

func TestAccResourceFleetAgentDownloadSource(t *testing.T) {
	random := sdkacctest.RandString(8)
	var idBeforeReplacement string

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		CheckDestroy:             checkResourceFleetAgentDownloadSourceDestroy,
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minVersionFleetAgentDownloadSource),
				Config:   testAccResourceFleetAgentDownloadSourceCreate(random),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_download_source.test", "name", fmt.Sprintf("Agent Download Source %s", random)),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_download_source.test", "default", "false"),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_download_source.test", "host", "https://artifacts.elastic.co/downloads/elastic-agent"),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_download_source.test", "proxy_id", "proxy-123"),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_download_source.test", "source_id", fmt.Sprintf("agent-download-source-%s", random)),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_download_source.test", "space_ids.#", "1"),
					resource.TestCheckTypeSetElemAttr("elasticstack_fleet_agent_download_source.test", "space_ids.*", "default"),
				),
			},
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minVersionFleetAgentDownloadSource),
				Config:   testAccResourceFleetAgentDownloadSourceUpdate(random),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_download_source.test", "name", fmt.Sprintf("Updated Agent Download Source %s", random)),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_download_source.test", "default", "true"),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_download_source.test", "host", "https://artifacts.elastic.co/downloads/elastic-agent-updated"),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_download_source.test", "proxy_id", "proxy-456"),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_download_source.test", "source_id", fmt.Sprintf("agent-download-source-%s", random)),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_download_source.test", "space_ids.#", "1"),
					resource.TestCheckTypeSetElemAttr("elasticstack_fleet_agent_download_source.test", "space_ids.*", "default"),
				),
			},
			{
				SkipFunc:          versionutils.CheckIfVersionIsUnsupported(minVersionFleetAgentDownloadSource),
				Config:            testAccResourceFleetAgentDownloadSourceUpdate(random),
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
				SkipFunc:          versionutils.CheckIfVersionIsUnsupported(minVersionFleetAgentDownloadSource),
				Config:            testAccResourceFleetAgentDownloadSourceUpdate(random),
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
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minVersionFleetAgentDownloadSource),
				Config:   testAccResourceFleetAgentDownloadSourceOmitOptionals(random),
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
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minVersionFleetAgentDownloadSource),
				Config:   testAccResourceFleetAgentDownloadSourceEmptySpaceIDs(random),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_download_source.test", "name", fmt.Sprintf("Empty Space IDs Agent Download Source %s", random)),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_download_source.test", "space_ids.#", "0"),
				),
			},
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minVersionFleetAgentDownloadSource),
				Config:   testAccResourceFleetAgentDownloadSourceReplaceSourceID(random),
				Check: resource.ComposeTestCheckFunc(
					testCheckFleetAgentDownloadSourceIDChanged("elasticstack_fleet_agent_download_source.test", &idBeforeReplacement),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_download_source.test", "source_id", fmt.Sprintf("agent-download-source-replaced-%s", random)),
				),
			},
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minVersionFleetAgentDownloadSource),
				Config:   testAccResourceFleetAgentDownloadSourceNonDefaultSpace(random),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_download_source.test", "space_ids.#", "1"),
					testCheckFleetAgentDownloadSourceSpaceContains("elasticstack_fleet_agent_download_source.test", fmt.Sprintf("fleet-agent-download-source-%s", random)),
				),
			},
		},
	})
}

func testAccResourceFleetAgentDownloadSourceCreate(suffix string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  kibana {}
}

resource "elasticstack_fleet_agent_download_source" "test" {
  name      = "Agent Download Source %s"
  source_id = "agent-download-source-%s"
  default   = false
  host      = "https://artifacts.elastic.co/downloads/elastic-agent"
  proxy_id  = "proxy-123"
  space_ids = ["default"]
}
`, suffix, suffix)
}

func testAccResourceFleetAgentDownloadSourceUpdate(suffix string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  kibana {}
}

resource "elasticstack_fleet_agent_download_source" "test" {
  name      = "Updated Agent Download Source %s"
  source_id = "agent-download-source-%s"
  default   = true
  host      = "https://artifacts.elastic.co/downloads/elastic-agent-updated"
  proxy_id  = "proxy-456"
  space_ids = ["default"]
}
`, suffix, suffix)
}

func testAccResourceFleetAgentDownloadSourceOmitOptionals(suffix string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  kibana {}
}

resource "elasticstack_fleet_agent_download_source" "test" {
  name      = "No Optionals Agent Download Source %s"
  host      = "https://artifacts.elastic.co/downloads/elastic-agent-no-optionals"
}
`, suffix)
}

func testAccResourceFleetAgentDownloadSourceEmptySpaceIDs(suffix string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  kibana {}
}

resource "elasticstack_fleet_agent_download_source" "test" {
  name      = "Empty Space IDs Agent Download Source %s"
  source_id = "agent-download-source-empty-space-ids-%s"
  default   = false
  host      = "https://artifacts.elastic.co/downloads/elastic-agent-empty-space-ids"
  space_ids = []
}
`, suffix, suffix)
}

func testAccResourceFleetAgentDownloadSourceReplaceSourceID(suffix string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  kibana {}
}

resource "elasticstack_fleet_agent_download_source" "test" {
  name      = "Replace Source ID Agent Download Source %s"
  source_id = "agent-download-source-replaced-%s"
  default   = false
  host      = "https://artifacts.elastic.co/downloads/elastic-agent-replaced"
  space_ids = ["default"]
}
`, suffix, suffix)
}

func testAccResourceFleetAgentDownloadSourceNonDefaultSpace(suffix string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_space" "test" {
  space_id = "fleet-agent-download-source-%s"
  name     = "Fleet Agent Download Source %s"
}

resource "elasticstack_fleet_agent_download_source" "test" {
  name      = "Non Default Space Agent Download Source %s"
  source_id = "agent-download-source-space-%s"
  default   = false
  host      = "https://artifacts.elastic.co/downloads/elastic-agent-space"
  space_ids = [elasticstack_kibana_space.test.space_id]
}
`, suffix, suffix, suffix, suffix)
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
