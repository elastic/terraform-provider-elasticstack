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

package customintegration_test

import (
	"archive/zip"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// buildMinimalIntegrationZip creates a minimal valid Elastic custom integration
// zip archive at a temp path. The zip contains the required top-level directory
// <pkgName>-<pkgVersion>/ with manifest.yml and docs/README.md.
func buildMinimalIntegrationZip(t *testing.T, pkgName, pkgVersion string) string {
	t.Helper()

	dir := t.TempDir()
	zipPath := filepath.Join(dir, fmt.Sprintf("%s-%s.zip", pkgName, pkgVersion))

	// format_version 1.0.0 is supported across all tested Kibana versions
	// (7.17.x through 9.x). It requires the `release` field and uses flat
	// condition syntax (kibana.version rather than kibana: version:).
	manifest := fmt.Sprintf(`format_version: "1.0.0"
name: %s
version: %s
title: "Test Integration %s"
description: "Minimal custom integration for acceptance testing"
type: integration
release: ga
categories:
  - custom
conditions:
  kibana.version: "^7.17.0 || ^8.0.0 || ^9.0.0"
owner:
  github: elastic
`, pkgName, pkgVersion, pkgName)

	readme := fmt.Sprintf("# %s\n\nMinimal custom integration for acceptance testing.\n", pkgName)

	f, err := os.Create(zipPath)
	if err != nil {
		t.Fatalf("failed to create zip file: %v", err)
	}
	defer f.Close()

	w := zip.NewWriter(f)
	defer w.Close()

	topDir := fmt.Sprintf("%s-%s/", pkgName, pkgVersion)

	// manifest.yml
	manifestWriter, err := w.Create(topDir + "manifest.yml")
	if err != nil {
		t.Fatalf("failed to create manifest.yml entry: %v", err)
	}
	if _, err := fmt.Fprint(manifestWriter, manifest); err != nil {
		t.Fatalf("failed to write manifest.yml: %v", err)
	}

	// docs/README.md
	readmeWriter, err := w.Create(topDir + "docs/README.md")
	if err != nil {
		t.Fatalf("failed to create README.md entry: %v", err)
	}
	if _, err := fmt.Fprint(readmeWriter, readme); err != nil {
		t.Fatalf("failed to write README.md: %v", err)
	}

	return zipPath
}

// checkCustomIntegrationDestroy verifies that the custom integration package
// is no longer installed after the resource is destroyed.
func checkCustomIntegrationDestroy(s *terraform.State) error {
	client, err := clients.NewAcceptanceTestingKibanaScopedClient()
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "elasticstack_fleet_custom_integration" {
			continue
		}

		fleetClient, err := client.GetFleetClient()
		if err != nil {
			return err
		}

		pkgName := rs.Primary.Attributes["package_name"]
		pkgVersion := rs.Primary.Attributes["package_version"]
		spaceID := rs.Primary.Attributes["space_id"]

		pkg, diags := fleet.GetPackage(context.Background(), fleetClient, pkgName, pkgVersion, spaceID)
		if diags.HasError() {
			return diagutil.FwDiagsAsError(diags)
		}

		if pkg != nil && pkg.Status != nil && *pkg.Status == "installed" {
			return fmt.Errorf("custom integration package %s/%s still exists and is installed, but it should have been removed", pkgName, pkgVersion)
		}
	}

	return nil
}

func TestAccFleetCustomIntegration(t *testing.T) {
	pkgName := "testcustompkg"

	zipPathV100 := buildMinimalIntegrationZip(t, pkgName, "1.0.0")
	zipPathV101 := buildMinimalIntegrationZip(t, pkgName, "1.0.1")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkCustomIntegrationDestroy,
		Steps: []resource.TestStep{
			// Step 1: Create — verify all computed attributes are set.
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"package_path": config.StringVariable(zipPathV100),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_custom_integration.test", "package_name", pkgName),
					resource.TestCheckResourceAttr("elasticstack_fleet_custom_integration.test", "package_version", "1.0.0"),
					resource.TestCheckResourceAttrSet("elasticstack_fleet_custom_integration.test", "checksum"),
					resource.TestCheckResourceAttrSet("elasticstack_fleet_custom_integration.test", "id"),
				),
			},
			// Step 2: Plan-only step — second apply must produce no changes (plan is clean).
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"package_path": config.StringVariable(zipPathV100),
				},
				PlanOnly: true,
			},
			// Step 3: Update — point package_path at a new zip with version 1.0.1.
			// ModifyPlan detects the checksum change and marks computed fields Unknown,
			// triggering Update to re-upload and set new values.
			// PreConfig waits for Fleet's upload rate limit (10s) to reset.
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables: config.Variables{
					"package_path": config.StringVariable(zipPathV101),
				},
				PreConfig: func() {
					time.Sleep(15 * time.Second)
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_custom_integration.test", "package_name", pkgName),
					resource.TestCheckResourceAttr("elasticstack_fleet_custom_integration.test", "package_version", "1.0.1"),
					resource.TestCheckResourceAttrSet("elasticstack_fleet_custom_integration.test", "checksum"),
				),
			},
		},
	})
}
