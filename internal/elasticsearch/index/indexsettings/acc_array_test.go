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
	"fmt"
	"slices"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	tfjsonpath "github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func checkIndexSettingAbsent(indexName, settingKey string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		flat, err := flatIndexSettings(indexName)
		if err != nil {
			return err
		}
		if value, ok := flat[settingKey]; ok {
			return fmt.Errorf("index %q setting %q = %v, want unset", indexName, settingKey, value)
		}
		return nil
	}
}

// Elasticsearch list settings do not guarantee element order.
func checkIndexSettingJSONArray(indexName string, expected []any) resource.TestCheckFunc {
	const settingKey = "index.query.default_field"

	return func(_ *terraform.State) error {
		flat, err := flatIndexSettings(indexName)
		if err != nil {
			return err
		}

		elements, ok := flat[settingKey].([]any)
		if !ok {
			return fmt.Errorf("index %q setting %q = %v (%T), want JSON array %v", indexName, settingKey, flat[settingKey], flat[settingKey], expected)
		}

		sortedJSON := func(values []any) []string {
			encoded := make([]string, 0, len(values))
			for _, value := range values {
				encoded = append(encoded, fmt.Sprintf("%v", value))
			}
			slices.Sort(encoded)
			return encoded
		}

		got := sortedJSON(elements)
		want := sortedJSON(expected)
		if !slices.Equal(got, want) {
			return fmt.Errorf("index %q setting %q = %v, want array elements %v", indexName, settingKey, elements, expected)
		}
		return nil
	}
}

// REQ-004: the typed query_default_field Set attribute round-trips as a real
// array against live Elasticsearch across the lifecycle: create, update,
// no-drift refresh, out-of-band drift detection with convergence, and removal
// via null reset.
func TestAccResourceIndexSettings_typedArraySettingsLifecycle(t *testing.T) {
	indexName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)

	twoFieldSet := statecheck.ExpectKnownValue(
		indexSettingsResourceName,
		tfjsonpath.New("query_default_field"),
		knownvalue.SetExact([]knownvalue.Check{
			knownvalue.StringExact("title"),
			knownvalue.StringExact("body"),
		}),
	)
	threeFieldSet := statecheck.ExpectKnownValue(
		indexSettingsResourceName,
		tfjsonpath.New("query_default_field"),
		knownvalue.SetExact([]knownvalue.Check{
			knownvalue.StringExact("title"),
			knownvalue.StringExact("body"),
			knownvalue.StringExact("summary"),
		}),
	)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceIndexSettingsDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("qdf_two"),
				ConfigVariables:          config.Variables{"index_name": config.StringVariable(indexName)},
				Check: resource.ComposeTestCheckFunc(
					checkIndexSettingJSONArray(indexName, []any{"title", "body"}),
				),
				ConfigStateChecks: []statecheck.StateCheck{twoFieldSet},
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("qdf_three"),
				ConfigVariables:          config.Variables{"index_name": config.StringVariable(indexName)},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{plancheck.ExpectNonEmptyPlan()},
				},
				Check: resource.ComposeTestCheckFunc(
					checkIndexSettingJSONArray(indexName, []any{"title", "body", "summary"}),
				),
				ConfigStateChecks: []statecheck.StateCheck{threeFieldSet},
			},
			{
				// Unchanged configuration: no drift after the previous read.
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("qdf_three"),
				ConfigVariables:          config.Variables{"index_name": config.StringVariable(indexName)},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
				},
				ConfigStateChecks: []statecheck.StateCheck{threeFieldSet},
			},
			{
				// The array is changed out of band; the plan shows drift and
				// the apply converges back to the declared element set.
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("qdf_three"),
				ConfigVariables:          config.Variables{"index_name": config.StringVariable(indexName)},
				PreConfig: func() {
					updateIndexSettingsOutOfBand(t, indexName, map[string]any{
						"index.query.default_field": []any{"body"},
					})
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{plancheck.ExpectNonEmptyPlan()},
				},
				Check: resource.ComposeTestCheckFunc(
					checkIndexSettingJSONArray(indexName, []any{"title", "body", "summary"}),
				),
				ConfigStateChecks: []statecheck.StateCheck{threeFieldSet},
			},
			{
				// Removing the declared setting resets it via null; the key
				// no longer appears in state or in Elasticsearch settings.
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("index_only"),
				ConfigVariables:          config.Variables{"index_name": config.StringVariable(indexName)},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckNoResourceAttr(indexSettingsResourceName, "query_default_field"),
					checkIndexSettingAbsent(indexName, "index.query.default_field"),
				),
			},
		},
	})
}

// REQ-006: terraform import hydrates the typed query_default_field Set from
// the Elasticsearch array response (settings_json stays unset), a follow-up
// apply converges the non-declared typed attributes, and the persisted state
// matches the declared configuration with no drift.
func TestAccResourceIndexSettings_typedArrayImport(t *testing.T) {
	indexName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)

	twoFieldSet := statecheck.ExpectKnownValue(
		indexSettingsResourceName,
		tfjsonpath.New("query_default_field"),
		knownvalue.SetExact([]knownvalue.Check{
			knownvalue.StringExact("title"),
			knownvalue.StringExact("body"),
		}),
	)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceIndexSettingsDestroy,
		Steps: []resource.TestStep{
			{
				// The index exists; the settings are out of band for the
				// index_settings resource (declared by no configuration).
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("index_only"),
				ConfigVariables:          config.Variables{"index_name": config.StringVariable(indexName)},
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("qdf_two"),
				ConfigVariables:          config.Variables{"index_name": config.StringVariable(indexName)},
				PreConfig: func() {
					updateIndexSettingsOutOfBand(t, indexName, map[string]any{
						"index.query.default_field": []any{"title", "body"},
					})
				},
				ResourceName:       indexSettingsResourceName,
				ImportState:        true,
				ImportStatePersist: true,
				ImportStateIdFunc:  importStateIDForIndexName(indexName),
				// The import hydrates the typed set from the API array and
				// leaves settings_json unset so nothing is adopted silently.
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(indexSettingsResourceName, "index", indexName),
					checkIndexSettingJSONArray(indexName, []any{"title", "body"}),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					twoFieldSet,
					statecheck.ExpectKnownValue(
						indexSettingsResourceName,
						tfjsonpath.New("settings_json"),
						knownvalue.Null(),
					),
				},
			},
			{
				// Applying the declared configuration converges the resource:
				// typed attributes the configuration does not declare stop
				// being tracked, and the declared query_default_field is
				// re-asserted.
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("qdf_two"),
				ConfigVariables:          config.Variables{"index_name": config.StringVariable(indexName)},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{plancheck.ExpectNonEmptyPlan()},
				},
				ConfigStateChecks: []statecheck.StateCheck{twoFieldSet},
			},
			{
				// A refresh after the convergence plans no changes.
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("qdf_two"),
				ConfigVariables:          config.Variables{"index_name": config.StringVariable(indexName)},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
				},
				ConfigStateChecks: []statecheck.StateCheck{twoFieldSet},
			},
		},
	})
}

// REQ-004: array-valued settings_json keys round-trip as real arrays against
// live Elasticsearch: create, update, no-drift refresh, out-of-band drift
// detection with convergence, and removal via null reset.
func TestAccResourceIndexSettings_settingsJSONArrayLifecycle(t *testing.T) {
	indexName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceIndexSettingsDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("array"),
				ConfigVariables:          config.Variables{"index_name": config.StringVariable(indexName)},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(indexSettingsResourceName, "settings_json", `{"index.query.default_field":["title","body"]}`),
					checkIndexSettingJSONArray(indexName, []any{"title", "body"}),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("array_updated"),
				ConfigVariables:          config.Variables{"index_name": config.StringVariable(indexName)},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{plancheck.ExpectNonEmptyPlan()},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(indexSettingsResourceName, "settings_json", `{"index.query.default_field":["title","body","summary"]}`),
					checkIndexSettingJSONArray(indexName, []any{"title", "body", "summary"}),
				),
			},
			{
				// Unchanged configuration: no drift after the previous read,
				// with element order preserved.
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("array_updated"),
				ConfigVariables:          config.Variables{"index_name": config.StringVariable(indexName)},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
				},
			},
			{
				// The array is changed out of band; the plan shows drift and
				// the apply converges back to the declared array.
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("array_updated"),
				ConfigVariables:          config.Variables{"index_name": config.StringVariable(indexName)},
				PreConfig: func() {
					updateIndexSettingsOutOfBand(t, indexName, map[string]any{
						"index.query.default_field": []any{"body"},
					})
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{plancheck.ExpectNonEmptyPlan()},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(indexSettingsResourceName, "settings_json", `{"index.query.default_field":["title","body","summary"]}`),
					checkIndexSettingJSONArray(indexName, []any{"title", "body", "summary"}),
				),
			},
			{
				// Removing the settings_json key resets it via null; the
				// attribute no longer appears in state or in Elasticsearch
				// settings.
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("index_only"),
				ConfigVariables:          config.Variables{"index_name": config.StringVariable(indexName)},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckNoResourceAttr(indexSettingsResourceName, "settings_json"),
					checkIndexSettingAbsent(indexName, "index.query.default_field"),
				),
			},
		},
	})
}
