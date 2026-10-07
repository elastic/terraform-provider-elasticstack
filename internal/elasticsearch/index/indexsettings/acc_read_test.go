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

package indexsettings_test

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	esclient "github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

const indexSettingsResourceName = "elasticstack_elasticsearch_index_settings.test"

func indexSettingsIDRegexpForIndex(indexName string) *regexp.Regexp {
	return regexp.MustCompile(`^[A-Za-z0-9_-]+/` + regexp.QuoteMeta(indexName) + `$`)
}

// REQ-004: settings Elasticsearch reports that the configuration does not
// declare (e.g. index.number_of_shards) must not drift into state; a
// re-apply of the same configuration plans no changes.
func TestAccResourceIndexSettings_unrelatedSettingsDoNotDrift(t *testing.T) {
	indexName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceIndexSettingsDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("settings_json"),
				ConfigVariables: config.Variables{
					"index_name": config.StringVariable(indexName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(indexSettingsResourceName, "id", indexSettingsIDRegexpForIndex(indexName)),
					resource.TestCheckResourceAttr(indexSettingsResourceName, "index", indexName),
					resource.TestCheckNoResourceAttr(indexSettingsResourceName, "number_of_replicas"),
					resource.TestCheckResourceAttr(indexSettingsResourceName, "settings_json", `{"index.refresh_interval":"10s"}`),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("settings_json"),
				ConfigVariables: config.Variables{
					"index_name": config.StringVariable(indexName),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
				},
			},
		},
	})
}

// REQ-004: a declared setting changed out of band surfaces as drift.
func TestAccResourceIndexSettings_outOfBandChangeSurfacesDrift(t *testing.T) {
	indexName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceIndexSettingsDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("settings_json"),
				ConfigVariables: config.Variables{
					"index_name": config.StringVariable(indexName),
				},
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("settings_json"),
				ConfigVariables: config.Variables{
					"index_name": config.StringVariable(indexName),
				},
				PreConfig: func() {
					updateIndexSettingsOutOfBand(t, indexName, map[string]any{"index.refresh_interval": "20s"})
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{plancheck.ExpectNonEmptyPlan()},
				},
				// Applying the declared configuration reverts the out-of-band
				// drift (settings_json was refreshed to "20s" on read).
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(indexSettingsResourceName, "settings_json", `{"index.refresh_interval":"10s"}`),
				),
			},
			{
				// Re-apply the declared value to converge.
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("settings_json"),
				ConfigVariables: config.Variables{
					"index_name": config.StringVariable(indexName),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
				},
			},
		},
	})
}

// REQ-006: terraform import hydrates the declared typed attributes once
// (settings_json stays unset) and a later refresh adopts nothing.
func TestAccResourceIndexSettings_importHydratesOnce(t *testing.T) {
	indexName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceIndexSettingsDestroy,
		Steps: []resource.TestStep{
			{
				// The index exists; the settings are out of band for the
				// index_settings resource (declared by no configuration).
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("index_only"),
				ConfigVariables: config.Variables{
					"index_name": config.StringVariable(indexName),
				},
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("settings_json"),
				ConfigVariables: config.Variables{
					"index_name": config.StringVariable(indexName),
				},
				PreConfig: func() {
					updateIndexSettingsOutOfBand(t, indexName, map[string]any{
						"index.number_of_replicas": 2,
						"index.refresh_interval":   "10s",
					})
				},
				ResourceName:       indexSettingsResourceName,
				ImportState:        true,
				ImportStatePersist: true,
				ImportStateIdFunc:  importStateIDForIndexName(indexName),
				// The first import hydrates the declared typed attributes and
				// leaves settings_json unset so nothing is adopted silently.
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(indexSettingsResourceName, "id", indexSettingsIDRegexpForIndex(indexName)),
					resource.TestCheckResourceAttr(indexSettingsResourceName, "index", indexName),
					resource.TestCheckResourceAttr(indexSettingsResourceName, "number_of_replicas", "2"),
					resource.TestCheckResourceAttr(indexSettingsResourceName, "refresh_interval", "10s"),
					resource.TestCheckNoResourceAttr(indexSettingsResourceName, "settings_json"),
				),
			},
			{
				// Applying the declared configuration converges the resource:
				// the hydrated typed attributes the configuration does not
				// declare stop being tracked and settings_json is set.
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("settings_json"),
				ConfigVariables: config.Variables{
					"index_name": config.StringVariable(indexName),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{plancheck.ExpectNonEmptyPlan()},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(indexSettingsResourceName, "settings_json", `{"index.refresh_interval":"10s"}`),
				),
			},
			{
				// A refresh after the import flag was cleared adopts nothing.
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("settings_json"),
				ConfigVariables: config.Variables{
					"index_name": config.StringVariable(indexName),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
				},
			},
		},
	})
}

// REQ-005: destroy is a no-op; the settings remain on the index.
func TestAccResourceIndexSettings_destroyIsNoop(t *testing.T) {
	indexName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceIndexSettingsDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("settings_json"),
				ConfigVariables: config.Variables{
					"index_name": config.StringVariable(indexName),
				},
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("index_only"),
				ConfigVariables: config.Variables{
					"index_name": config.StringVariable(indexName),
				},
				Check: resource.ComposeTestCheckFunc(
					checkIndexSettingValue(indexName, "index.refresh_interval", "10s"),
				),
			},
		},
	})
}

// REQ-004: a read of an index deleted out of band reports not-found so the
// resource is removed from state.
func TestAccResourceIndexSettings_indexDeletedOutOfBand(t *testing.T) {
	indexName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("settings_json"),
				ConfigVariables: config.Variables{
					"index_name": config.StringVariable(indexName),
				},
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("settings_json"),
				ConfigVariables: config.Variables{
					"index_name": config.StringVariable(indexName),
				},
				PreConfig: func() {
					deleteIndexOutOfBand(t, indexName)
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{plancheck.ExpectNonEmptyPlan()},
				},
			},
		},
	})
}

func importStateIDForIndexName(indexName string) resource.ImportStateIdFunc {
	return func(_ *terraform.State) (string, error) {
		client, err := clients.NewAcceptanceTestingElasticsearchScopedClient()
		if err != nil {
			return "", err
		}

		id, diags := client.ID(context.Background(), indexName)
		if diags.HasError() {
			return "", fmt.Errorf("failed to build import ID for index %q: %v", indexName, diags)
		}
		return id.String(), nil
	}
}

func checkResourceIndexSettingsDestroy(s *terraform.State) error {
	client, err := clients.NewAcceptanceTestingElasticsearchScopedClient()
	if err != nil {
		return err
	}

	indexNames := make(map[string]struct{})
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "elasticstack_elasticsearch_index" {
			continue
		}
		compID, compDiags := clients.CompositeIDFromStr(rs.Primary.ID)
		if compDiags.HasError() {
			return fmt.Errorf("failed to parse composite ID %q: %v", rs.Primary.ID, compDiags)
		}
		indexNames[compID.ResourceID] = struct{}{}
	}

	for indexName := range indexNames {
		indexState, diags := esclient.GetIndex(context.Background(), client, indexName)
		if diags.HasError() {
			return fmt.Errorf("failed to get index %q: %v", indexName, diags)
		}
		if indexState != nil {
			return fmt.Errorf("index %q still exists", indexName)
		}
	}
	return nil
}

func updateIndexSettingsOutOfBand(t *testing.T, indexName string, settings map[string]any) {
	t.Helper()

	client, err := clients.NewAcceptanceTestingElasticsearchScopedClient()
	if err != nil {
		t.Fatalf("failed to create Elasticsearch client: %s", err)
	}

	diags := esclient.UpdateIndexSettings(context.Background(), client, indexName, settings)
	if diags.HasError() {
		t.Fatalf("failed to update index %q settings out of band: %v", indexName, diags)
	}
}

func deleteIndexOutOfBand(t *testing.T, indexName string) {
	t.Helper()

	client, err := clients.NewAcceptanceTestingElasticsearchScopedClient()
	if err != nil {
		t.Fatalf("failed to create Elasticsearch client: %s", err)
	}

	diags := esclient.DeleteIndex(context.Background(), client, indexName)
	if diags.HasError() {
		t.Fatalf("failed to delete index %q out of band: %v", indexName, diags)
	}
}

func checkIndexSettingValue(indexName, settingKey, expected string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		client, err := clients.NewAcceptanceTestingElasticsearchScopedClient()
		if err != nil {
			return err
		}

		indexState, diags := esclient.GetIndex(context.Background(), client, indexName)
		if diags.HasError() {
			return fmt.Errorf("failed to get index %q: %v", indexName, diags)
		}
		if indexState == nil {
			return fmt.Errorf("index %q not found", indexName)
		}

		settingsBytes, err := json.Marshal(indexState.Settings)
		if err != nil {
			return err
		}

		var flat map[string]any
		if err := json.Unmarshal(settingsBytes, &flat); err != nil {
			return err
		}
		if fmt.Sprint(flat[settingKey]) != expected {
			return fmt.Errorf("index %q setting %q = %v, want %q", indexName, settingKey, flat[settingKey], expected)
		}
		return nil
	}
}
