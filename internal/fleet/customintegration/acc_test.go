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
	"archive/tar"
	"archive/zip"
	"compress/gzip"
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
	goversion "github.com/hashicorp/go-version"
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

// preCheckMinKibanaVersion skips the test if the connected Kibana version is
// older than 8.2.0. elasticstack_fleet_custom_integration requires 8.2+
// because GET /api/fleet/epm/packages/{name}/{version} is unreliable for
// custom packages on older versions.
func preCheckMinKibanaVersion(t *testing.T) {
	t.Helper()
	client, err := clients.NewAcceptanceTestingKibanaScopedClient()
	if err != nil {
		t.Fatalf("failed to create Kibana client for version check: %v", err)
	}
	minVer := goversion.Must(goversion.NewVersion("8.2.0"))
	meets, verDiags := client.EnforceMinVersion(context.Background(), minVer)
	if verDiags.HasError() {
		t.Fatalf("failed to check Kibana version: %v", verDiags)
	}
	if !meets {
		t.Skip("skipping: elasticstack_fleet_custom_integration requires Kibana 8.2.0 or later")
	}
}

// checkCustomIntegrationDestroy verifies that the custom integration package
// is no longer installed after the resource is destroyed.
func checkCustomIntegrationDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "elasticstack_fleet_custom_integration" {
			continue
		}

		pkgName := rs.Primary.Attributes["package_name"]
		pkgVersion := rs.Primary.Attributes["package_version"]
		spaceID := rs.Primary.Attributes["space_id"]

		installed, err := fleetPackageInstalled(context.Background(), pkgName, pkgVersion, spaceID)
		if err != nil {
			return err
		}

		if installed {
			return fmt.Errorf("custom integration package %s/%s still exists and is installed, but it should have been removed", pkgName, pkgVersion)
		}
	}

	return nil
}

func fleetPackageInstalled(ctx context.Context, pkgName, pkgVersion, spaceID string) (bool, error) {
	client, err := clients.NewAcceptanceTestingKibanaScopedClient()
	if err != nil {
		return false, err
	}
	fleetClient, err := client.GetFleetClient()
	if err != nil {
		return false, err
	}
	pkg, diags := fleet.GetPackage(ctx, fleetClient, pkgName, pkgVersion, spaceID)
	if diags.HasError() {
		return false, diagutil.FwDiagsAsError(diags)
	}
	return pkg != nil && pkg.Status != nil && *pkg.Status == "installed", nil
}

// buildMinimalIntegrationTarGz creates a minimal valid Elastic custom integration
// tar.gz archive at a temp path. The archive contains the required top-level
// directory <pkgName>-<pkgVersion>/ with manifest.yml and docs/README.md.
func buildMinimalIntegrationTarGz(t *testing.T, pkgName, pkgVersion string) string {
	t.Helper()

	dir := t.TempDir()
	tgzPath := filepath.Join(dir, fmt.Sprintf("%s-%s.tar.gz", pkgName, pkgVersion))

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

	f, err := os.Create(tgzPath)
	if err != nil {
		t.Fatalf("failed to create tar.gz file: %v", err)
	}
	defer f.Close()

	gw := gzip.NewWriter(f)
	defer gw.Close()

	tw := tar.NewWriter(gw)
	defer tw.Close()

	topDir := fmt.Sprintf("%s-%s/", pkgName, pkgVersion)

	addEntry := func(name, content string) {
		hdr := &tar.Header{
			Name: topDir + name,
			Mode: 0o644,
			Size: int64(len(content)),
		}
		if err := tw.WriteHeader(hdr); err != nil {
			t.Fatalf("failed to write tar header for %s: %v", name, err)
		}
		if _, err := fmt.Fprint(tw, content); err != nil {
			t.Fatalf("failed to write tar entry %s: %v", name, err)
		}
	}

	addEntry("manifest.yml", manifest)
	addEntry("docs/README.md", readme)

	return tgzPath
}

func TestAccFleetCustomIntegration(t *testing.T) {
	pkgNameV100 := "testcustompkg"
	pkgNameV101 := "testcustompkgnext"

	zipPathV100 := buildMinimalIntegrationZip(t, pkgNameV100, "1.0.0")
	zipPathV101 := buildMinimalIntegrationZip(t, pkgNameV101, "1.0.1")

	// step1Checksum captures the checksum from step 1 so we can assert that it
	// changes after the package is updated in step 3.
	var step1Checksum string

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t); preCheckMinKibanaVersion(t) },
		Steps: []resource.TestStep{
			// Step 1: Create — verify all computed attributes are set.
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"package_path": config.StringVariable(zipPathV100),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_custom_integration.test", "package_name", pkgNameV100),
					resource.TestCheckResourceAttr("elasticstack_fleet_custom_integration.test", "package_version", "1.0.0"),
					resource.TestCheckResourceAttrSet("elasticstack_fleet_custom_integration.test", "checksum"),
					resource.TestCheckResourceAttrSet("elasticstack_fleet_custom_integration.test", "id"),
					resource.TestCheckResourceAttr("elasticstack_fleet_custom_integration.test", "ignore_mapping_update_errors", "false"),
					resource.TestCheckResourceAttr("elasticstack_fleet_custom_integration.test", "skip_data_stream_rollover", "false"),
					resource.TestCheckResourceAttr("elasticstack_fleet_custom_integration.test", "skip_destroy", "false"),
					resource.TestCheckNoResourceAttr("elasticstack_fleet_custom_integration.test", "space_id"),
					func(s *terraform.State) error {
						rs := s.RootModule().Resources["elasticstack_fleet_custom_integration.test"]
						if rs == nil {
							return fmt.Errorf("resource elasticstack_fleet_custom_integration.test not found in state")
						}
						step1Checksum = rs.Primary.Attributes["checksum"]
						return nil
					},
				),
			},
			// Step 2: Plan-only step — second apply must produce no changes (plan is clean).
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"package_path": config.StringVariable(zipPathV100),
				},
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
			// Step 3: Update — point package_path at a new zip with a new package name
			// and version. ModifyPlan detects the checksum change and marks computed
			// fields Unknown, triggering Update to re-upload, adopt the new package,
			// and uninstall the old one.
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
					resource.TestCheckResourceAttr("elasticstack_fleet_custom_integration.test", "package_name", pkgNameV101),
					resource.TestCheckResourceAttr("elasticstack_fleet_custom_integration.test", "package_version", "1.0.1"),
					resource.TestCheckResourceAttrSet("elasticstack_fleet_custom_integration.test", "checksum"),
					func(_ *terraform.State) error {
						return checkPackageNotInstalledInFleet(pkgNameV100, "1.0.0", "")
					},
					// Assert the checksum changed after switching to a different archive.
					func(s *terraform.State) error {
						rs := s.RootModule().Resources["elasticstack_fleet_custom_integration.test"]
						if rs == nil {
							return fmt.Errorf("resource elasticstack_fleet_custom_integration.test not found in state")
						}
						step3Checksum := rs.Primary.Attributes["checksum"]
						if step3Checksum == step1Checksum {
							return fmt.Errorf("expected checksum to change after package update, got unchanged value: %s", step3Checksum)
						}
						return nil
					},
				),
			},
			// Step 4: Verify skip_data_stream_rollover=true uploads successfully.
			// PreConfig waits for Fleet's upload rate limit (10s) to reset.
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("skip_rollover"),
				ConfigVariables: config.Variables{
					"package_path": config.StringVariable(zipPathV101),
				},
				PreConfig: func() {
					time.Sleep(15 * time.Second)
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_custom_integration.test", "package_name", pkgNameV101),
					resource.TestCheckResourceAttr("elasticstack_fleet_custom_integration.test", "skip_data_stream_rollover", "true"),
					resource.TestCheckResourceAttrSet("elasticstack_fleet_custom_integration.test", "checksum"),
				),
			},
			// Step 5: Reset skip_data_stream_rollover back to false and assert "false".
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
					resource.TestCheckResourceAttr("elasticstack_fleet_custom_integration.test", "package_name", pkgNameV101),
					resource.TestCheckResourceAttr("elasticstack_fleet_custom_integration.test", "skip_data_stream_rollover", "false"),
					resource.TestCheckResourceAttrSet("elasticstack_fleet_custom_integration.test", "checksum"),
				),
			},
			// Step 6: Verify ignore_mapping_update_errors=true uploads successfully.
			// PreConfig waits for Fleet's upload rate limit (10s) to reset.
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("ignore_mapping"),
				ConfigVariables: config.Variables{
					"package_path": config.StringVariable(zipPathV101),
				},
				PreConfig: func() {
					time.Sleep(15 * time.Second)
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_custom_integration.test", "package_name", pkgNameV101),
					resource.TestCheckResourceAttr("elasticstack_fleet_custom_integration.test", "ignore_mapping_update_errors", "true"),
					resource.TestCheckResourceAttrSet("elasticstack_fleet_custom_integration.test", "checksum"),
				),
			},
			// Step 7: Reset ignore_mapping_update_errors back to false and assert "false".
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
					resource.TestCheckResourceAttr("elasticstack_fleet_custom_integration.test", "package_name", pkgNameV101),
					resource.TestCheckResourceAttr("elasticstack_fleet_custom_integration.test", "ignore_mapping_update_errors", "false"),
					resource.TestCheckResourceAttrSet("elasticstack_fleet_custom_integration.test", "checksum"),
				),
			},
		},
	})
}

func TestAccFleetCustomIntegration_Gzip(t *testing.T) {
	pkgName := "testcustomgzpkg"
	tgzPath := buildMinimalIntegrationTarGz(t, pkgName, "1.0.0")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); preCheckMinKibanaVersion(t) },
		CheckDestroy: checkCustomIntegrationDestroy,
		Steps: []resource.TestStep{
			// Step 1: Upload a tar.gz archive and verify computed attributes are set.
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"package_path": config.StringVariable(tgzPath),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_custom_integration.test", "package_name", pkgName),
					resource.TestCheckResourceAttr("elasticstack_fleet_custom_integration.test", "package_version", "1.0.0"),
					resource.TestCheckResourceAttrSet("elasticstack_fleet_custom_integration.test", "checksum"),
				),
			},
		},
	})
}

// TestAccFleetCustomIntegration_KibanaConnection exercises the kibana_connection
// block by configuring an explicit endpoint, auth, and insecure=true on the
// resource-scoped Kibana client. The test verifies that resource creation succeeds
// and all core computed attributes are returned when the connection block is present.
func TestAccFleetCustomIntegration_KibanaConnection(t *testing.T) {
	pkgName := "testcustomkbconpkg"
	zipPath := buildMinimalIntegrationZip(t, pkgName, "1.0.0")

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			preCheckMinKibanaVersion(t)
			acctest.PreCheckWithExplicitKibanaEndpoint(t)
		},
		CheckDestroy: checkCustomIntegrationDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: acctest.KibanaConnectionVariables(config.Variables{
					"package_path": config.StringVariable(zipPath),
				}),
				Check: resource.ComposeTestCheckFunc(append([]resource.TestCheckFunc{
					resource.TestCheckResourceAttr("elasticstack_fleet_custom_integration.test", "package_name", pkgName),
					resource.TestCheckResourceAttr("elasticstack_fleet_custom_integration.test", "package_version", "1.0.0"),
					resource.TestCheckResourceAttrSet("elasticstack_fleet_custom_integration.test", "checksum"),
					resource.TestCheckResourceAttrSet("elasticstack_fleet_custom_integration.test", "id"),
					resource.TestCheckResourceAttr("elasticstack_fleet_custom_integration.test", "kibana_connection.#", "1"),
					resource.TestCheckResourceAttr("elasticstack_fleet_custom_integration.test", "kibana_connection.0.insecure", "true"),
					resource.TestCheckResourceAttr("elasticstack_fleet_custom_integration.test", "kibana_connection.0.endpoints.#", "1"),
					resource.TestCheckResourceAttrSet("elasticstack_fleet_custom_integration.test", "kibana_connection.0.endpoints.0"),
				}, acctest.KibanaConnectionAuthChecks("elasticstack_fleet_custom_integration.test")...)...),
			},
		},
	})
}

func TestAccFleetCustomIntegration_SkipDestroy(t *testing.T) {
	pkgName := "testcustomskippkg"
	pkgVersion := "1.0.0"
	zipPath := buildMinimalIntegrationZip(t, pkgName, pkgVersion)

	t.Cleanup(func() {
		cleanupPackageInFleet(t, pkgName, pkgVersion, "")
	})

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); preCheckMinKibanaVersion(t) },
		CheckDestroy: checkCustomIntegrationDestroy,
		Steps: []resource.TestStep{
			// Step 1: Create with skip_destroy=true.
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("skip_destroy_on"),
				ConfigVariables: config.Variables{
					"package_path": config.StringVariable(zipPath),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_custom_integration.test", "package_name", pkgName),
					resource.TestCheckResourceAttr("elasticstack_fleet_custom_integration.test", "skip_destroy", "true"),
					resource.TestCheckResourceAttrSet("elasticstack_fleet_custom_integration.test", "checksum"),
				),
			},
			// Step 2: Explicit destroy while skip_destroy=true is active. The resource
			// is removed from Terraform state but the Fleet package must remain installed.
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("skip_destroy_on"),
				ConfigVariables: config.Variables{
					"package_path": config.StringVariable(zipPath),
				},
				Destroy: true,
				Check: func(_ *terraform.State) error {
					return checkPackageStillInstalledInFleet(pkgName, pkgVersion, "")
				},
			},
		},
	})
}

// checkPackageStillInstalledInFleet verifies that a custom integration package
// is still installed in Fleet after a skip_destroy=true resource destroy.
func checkPackageStillInstalledInFleet(pkgName, pkgVersion, spaceID string) error {
	installed, err := fleetPackageInstalled(context.Background(), pkgName, pkgVersion, spaceID)
	if err != nil {
		return err
	}
	if !installed {
		return fmt.Errorf(
			"expected package %s/%s to remain installed after skip_destroy=true destroy",
			pkgName, pkgVersion,
		)
	}
	return nil
}

func checkPackageNotInstalledInFleet(pkgName, pkgVersion, spaceID string) error {
	installed, err := fleetPackageInstalled(context.Background(), pkgName, pkgVersion, spaceID)
	if err != nil {
		return err
	}
	if installed {
		return fmt.Errorf(
			"expected package %s/%s to be uninstalled after update",
			pkgName, pkgVersion,
		)
	}
	return nil
}

func cleanupPackageInFleet(t *testing.T, pkgName, pkgVersion, spaceID string) {
	t.Helper()

	client, err := clients.NewAcceptanceTestingKibanaScopedClient()
	if err != nil {
		t.Logf("skipping cleanup for %s/%s: %v", pkgName, pkgVersion, err)
		return
	}
	fleetClient, err := client.GetFleetClient()
	if err != nil {
		t.Logf("skipping cleanup for %s/%s: %v", pkgName, pkgVersion, err)
		return
	}
	diags := fleet.Uninstall(context.Background(), fleetClient, pkgName, pkgVersion, spaceID, false)
	if diags.HasError() {
		t.Errorf("failed to uninstall package during cleanup: %v", diagutil.FwDiagsAsError(diags))
	}
}

func TestAccFleetCustomIntegration_SpaceID(t *testing.T) {
	pkgName := "testcustomspacepkg"
	spaceID := "acc-test-space-customintegration"
	zipPath := buildMinimalIntegrationZip(t, pkgName, "1.0.0")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); preCheckMinKibanaVersion(t) },
		CheckDestroy: checkCustomIntegrationDestroy,
		Steps: []resource.TestStep{
			// Step 1: Upload the package into a non-default Kibana space and verify
			// all Fleet API calls are routed to /s/{space_id}/api/fleet/epm/packages.
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("space_id"),
				ConfigVariables: config.Variables{
					"package_path": config.StringVariable(zipPath),
					"space_id":     config.StringVariable(spaceID),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_custom_integration.test", "package_name", pkgName),
					resource.TestCheckResourceAttr("elasticstack_fleet_custom_integration.test", "space_id", spaceID),
					resource.TestCheckResourceAttrSet("elasticstack_fleet_custom_integration.test", "checksum"),
				),
			},
			// Step 2: Plan-only with unchanged config — verify the plan is empty (idempotency).
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("space_id"),
				ConfigVariables: config.Variables{
					"package_path": config.StringVariable(zipPath),
					"space_id":     config.StringVariable(spaceID),
				},
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
			// Step 3: Plan-only with a different space_id on the custom integration only —
			// the elasticstack_kibana_space resource stays unchanged (same space_id as
			// steps 1-2), so the non-empty plan is caused exclusively by the custom
			// integration's RequiresReplace() modifier firing on the changed space_id.
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("space_id_replace"),
				ConfigVariables: config.Variables{
					"package_path": config.StringVariable(zipPath),
					"space_id":     config.StringVariable(spaceID),
					"new_space_id": config.StringVariable("acc-test-space-customintegration-b"),
				},
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccFleetCustomIntegration_Timeouts(t *testing.T) {
	pkgName := "testcustomtimeoutpkg"
	zipPath := buildMinimalIntegrationZip(t, pkgName, "1.0.0")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); preCheckMinKibanaVersion(t) },
		CheckDestroy: checkCustomIntegrationDestroy,
		Steps: []resource.TestStep{
			// Step 1: Create with distinct create (15m) and update (25m) timeouts.
			// Verify the resource is created successfully with the timeout block accepted.
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("timeouts"),
				ConfigVariables: config.Variables{
					"package_path": config.StringVariable(zipPath),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_custom_integration.test", "package_name", pkgName),
					resource.TestCheckResourceAttr("elasticstack_fleet_custom_integration.test", "package_version", "1.0.0"),
					resource.TestCheckResourceAttrSet("elasticstack_fleet_custom_integration.test", "checksum"),
					resource.TestCheckResourceAttr("elasticstack_fleet_custom_integration.test", "timeouts.create", "15m"),
					resource.TestCheckResourceAttr("elasticstack_fleet_custom_integration.test", "timeouts.update", "25m"),
				),
			},
			// Step 2: Update — change skip_data_stream_rollover to trigger a re-upload,
			// exercising the update timeout (25m) path.
			// PreConfig waits for Fleet's upload rate limit (10s) to reset.
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("timeouts_update"),
				ConfigVariables: config.Variables{
					"package_path": config.StringVariable(zipPath),
				},
				PreConfig: func() {
					time.Sleep(15 * time.Second)
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_custom_integration.test", "package_name", pkgName),
					resource.TestCheckResourceAttr("elasticstack_fleet_custom_integration.test", "package_version", "1.0.0"),
					resource.TestCheckResourceAttr("elasticstack_fleet_custom_integration.test", "skip_data_stream_rollover", "true"),
					resource.TestCheckResourceAttrSet("elasticstack_fleet_custom_integration.test", "checksum"),
				),
			},
		},
	})
}
