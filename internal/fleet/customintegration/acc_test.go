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
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// buildMinimalIntegrationZip creates a minimal valid Elastic custom integration
// zip archive at a temp path. The zip contains the required top-level directory
// <pkgName>-<pkgVersion>/ with manifest.yml and docs/README.md.
func buildMinimalIntegrationZip(t *testing.T, pkgName, pkgVersion string) string {
	t.Helper()

	dir := t.TempDir()
	zipPath := filepath.Join(dir, fmt.Sprintf("%s-%s.zip", pkgName, pkgVersion))

	manifest := fmt.Sprintf(`format_version: "3.0.0"
name: %s
version: %s
title: "Test Integration %s"
description: "Minimal custom integration for acceptance testing"
type: integration
categories:
  - custom
conditions:
  kibana:
    version: "^8.0.0"
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

func TestAccFleetCustomIntegration(t *testing.T) {
	pkgName := "testcustompkg"
	pkgVersion := "1.0.0"
	zipPath := buildMinimalIntegrationZip(t, pkgName, pkgVersion)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			// Step 1: Create — verify all computed attributes are set.
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"package_path": config.StringVariable(zipPath),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_custom_integration.test", "package_name", pkgName),
					resource.TestCheckResourceAttr("elasticstack_fleet_custom_integration.test", "package_version", pkgVersion),
					resource.TestCheckResourceAttrSet("elasticstack_fleet_custom_integration.test", "checksum"),
					resource.TestCheckResourceAttrSet("elasticstack_fleet_custom_integration.test", "id"),
				),
			},
			// Step 2: Plan-only step — second apply must produce no changes (plan is clean).
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"package_path": config.StringVariable(zipPath),
				},
				PlanOnly: true,
			},
		},
	})
}
