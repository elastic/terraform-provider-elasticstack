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

package advancedsettings_test

import (
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"reflect"
	"regexp"
	"strings"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/stretchr/testify/require"
)

const advancedSettingsResource = "elasticstack_kibana_advanced_settings.test"

func checkKibanaSettings(client *kibanaoapi.Client, spaceID string, want map[string]string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		settings, diags := kibanaoapi.GetAdvancedSettings(context.Background(), client, spaceID)
		if diags.HasError() {
			return fmt.Errorf("reading settings of space %q: %v", spaceID, diags)
		}
		for key, wantJSON := range want {
			got, found := settings[key]
			if !found {
				return fmt.Errorf("space %q holds no value for %q, want %s", spaceID, key, wantJSON)
			}
			if !jsonEqual(string(got), wantJSON) {
				return fmt.Errorf("space %q holds %s for %q, want %s", spaceID, got, key, wantJSON)
			}
		}
		return nil
	}
}

func checkKibanaSettingsAtDefault(client *kibanaoapi.Client, spaceID string, keys ...string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		settings, diags := kibanaoapi.GetAdvancedSettings(context.Background(), client, spaceID)
		if diags.HasError() {
			return fmt.Errorf("reading settings of space %q: %v", spaceID, diags)
		}
		for _, key := range keys {
			if got, found := settings[key]; found {
				return fmt.Errorf("space %q holds %s for %q, want it reset to its Kibana default", spaceID, got, key)
			}
		}
		return nil
	}
}

func setKibanaSettings(t *testing.T, client *kibanaoapi.Client, spaceID string, settings map[string]any) {
	t.Helper()
	_, diags := kibanaoapi.ApplyAdvancedSettings(context.Background(), client, spaceID, settings)
	require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)
}

func snapshotKibanaSettings(t *testing.T, client *kibanaoapi.Client, spaceID string) map[string]json.RawMessage {
	t.Helper()
	settings, diags := kibanaoapi.GetAdvancedSettings(context.Background(), client, spaceID)
	require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)
	return settings
}

func checkKibanaSettingsUnchanged(client *kibanaoapi.Client, spaceID string, before *map[string]json.RawMessage, managedKeys ...string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		after, diags := kibanaoapi.GetAdvancedSettings(context.Background(), client, spaceID)
		if diags.HasError() {
			return fmt.Errorf("reading settings of space %q: %v", spaceID, diags)
		}
		unmanagedBefore, unmanagedAfter := withoutKeys(*before, managedKeys), withoutKeys(after, managedKeys)
		if !maps.EqualFunc(unmanagedBefore, unmanagedAfter, func(a, b json.RawMessage) bool { return jsonEqual(string(a), string(b)) }) {
			return fmt.Errorf("settings of space %q changed from %s to %s", spaceID, rawSettingsText(unmanagedBefore), rawSettingsText(unmanagedAfter))
		}
		return nil
	}
}

func withoutKeys(settings map[string]json.RawMessage, keys []string) map[string]json.RawMessage {
	remaining := maps.Clone(settings)
	for _, key := range keys {
		delete(remaining, key)
	}
	return remaining
}

func requireKibanaSettingsUnchanged(t *testing.T, client *kibanaoapi.Client, spaceID string, before map[string]json.RawMessage) {
	t.Helper()
	require.NoError(t, checkKibanaSettingsUnchanged(client, spaceID, &before)(nil))
}

func rawSettingsText(settings map[string]json.RawMessage) string {
	text, _ := json.Marshal(settings)
	return string(text)
}

func jsonEqual(a, b string) bool {
	var left, right any
	return json.Unmarshal([]byte(a), &left) == nil && json.Unmarshal([]byte(b), &right) == nil && reflect.DeepEqual(left, right)
}

func deleteKibanaSpace(t *testing.T, client *kibanaoapi.Client, spaceID string) {
	t.Helper()
	diags := kibanaoapi.DeleteSpace(context.Background(), client, spaceID)
	require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)
}

func checkResourceRemovedFromState(resourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		if _, found := state.RootModule().Resources[resourceName]; found {
			return fmt.Errorf("%s is still in the state", resourceName)
		}
		return nil
	}
}

func expectStringSettingUpdatedTo(key, value string) resource.ConfigPlanChecks {
	return resource.ConfigPlanChecks{
		PreApply: []plancheck.PlanCheck{
			plancheck.ExpectResourceAction(advancedSettingsResource, plancheck.ResourceActionUpdate),
			plancheck.ExpectKnownValue(advancedSettingsResource,
				tfjsonpath.New("string_settings").AtMapKey(key), knownvalue.StringExact(value)),
		},
	}
}

func spaceVariables(spaceID string) config.Variables {
	return config.Variables{"space_id": config.StringVariable(spaceID)}
}

func TestAccResourceKibanaAdvancedSettings_create(t *testing.T) {
	client := acctest.AdvancedSettingsKibanaClient(t)
	spaceID := acctest.NewAdvancedSettingsTestSpace(t, client)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          spaceVariables(spaceID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(advancedSettingsResource, "id", spaceID),
					resource.TestCheckResourceAttr(advancedSettingsResource, "space_id", spaceID),
					resource.TestCheckResourceAttr(advancedSettingsResource, "string_settings.%", "3"),
					resource.TestCheckResourceAttr(advancedSettingsResource, "string_settings.dateFormat:tz", "Europe/Rome"),
					resource.TestCheckResourceAttr(advancedSettingsResource, "string_settings.custom:jsonText", `["a","b"]`),
					resource.TestCheckResourceAttr(advancedSettingsResource, "string_settings.custom:teamNote", "hello"),
					resource.TestCheckResourceAttr(advancedSettingsResource, "json_settings.%", "3"),
					resource.TestCheckResourceAttr(advancedSettingsResource, "json_settings.csv:quoteValues", "false"),
					resource.TestCheckResourceAttr(advancedSettingsResource, "json_settings.truncate:maxHeight", "0"),
					resource.TestCheckResourceAttr(advancedSettingsResource, "json_settings.defaultColumns", `["@timestamp","message"]`),
					checkKibanaSettings(client, spaceID, map[string]string{
						"dateFormat:tz":      `"Europe/Rome"`,
						"custom:jsonText":    `"[\"a\",\"b\"]"`,
						"custom:teamNote":    `"hello"`,
						"csv:quoteValues":    `false`,
						"truncate:maxHeight": `0`,
						"defaultColumns":     `["@timestamp","message"]`,
					}),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          spaceVariables(spaceID),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
				},
			},
		},
	})
}

func TestAccResourceKibanaAdvancedSettings_createLeavesUnmanagedSettingsUntouched(t *testing.T) {
	client := acctest.AdvancedSettingsKibanaClient(t)
	spaceID := acctest.NewAdvancedSettingsTestSpace(t, client)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					setKibanaSettings(t, client, spaceID, map[string]any{"defaultRoute": "/app/dashboards"})
				},
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          spaceVariables(spaceID),
				Check: checkKibanaSettings(client, spaceID, map[string]string{
					"custom:teamNote": `"hello"`,
					"defaultRoute":    `"/app/dashboards"`,
				}),
			},
		},
	})
}

func TestAccResourceKibanaAdvancedSettings_createWithOmittedMaps(t *testing.T) {
	client := acctest.AdvancedSettingsKibanaClient(t)
	spaceID := acctest.NewAdvancedSettingsTestSpace(t, client)
	var settingsBefore map[string]json.RawMessage

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					setKibanaSettings(t, client, spaceID, map[string]any{"custom:existing": "x"})
					settingsBefore = snapshotKibanaSettings(t, client, spaceID)
				},
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          spaceVariables(spaceID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(advancedSettingsResource, "id", spaceID),
					resource.TestCheckNoResourceAttr(advancedSettingsResource, "string_settings"),
					resource.TestCheckNoResourceAttr(advancedSettingsResource, "json_settings"),
					checkKibanaSettingsUnchanged(client, spaceID, &settingsBefore),
				),
			},
		},
	})
}

func TestAccResourceKibanaAdvancedSettings_createWithEmptyMaps(t *testing.T) {
	client := acctest.AdvancedSettingsKibanaClient(t)
	spaceID := acctest.NewAdvancedSettingsTestSpace(t, client)
	var settingsBefore map[string]json.RawMessage

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					setKibanaSettings(t, client, spaceID, map[string]any{"custom:existing": "x"})
					settingsBefore = snapshotKibanaSettings(t, client, spaceID)
				},
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          spaceVariables(spaceID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(advancedSettingsResource, "id", spaceID),
					resource.TestCheckResourceAttr(advancedSettingsResource, "string_settings.%", "0"),
					resource.TestCheckResourceAttr(advancedSettingsResource, "json_settings.%", "0"),
					checkKibanaSettingsUnchanged(client, spaceID, &settingsBefore),
				),
			},
		},
	})
}

func TestAccResourceKibanaAdvancedSettings_jsonFormattingDoesNotCauseDiff(t *testing.T) {
	client := acctest.AdvancedSettingsKibanaClient(t)
	spaceID := acctest.NewAdvancedSettingsTestSpace(t, client)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          spaceVariables(spaceID),
				Check: checkKibanaSettings(client, spaceID, map[string]string{
					"defaultColumns": `["@timestamp","message"]`,
					"custom:object":  `{"a":{"c":true,"d":[1,2]},"b":1}`,
				}),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          spaceVariables(spaceID),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
				},
			},
		},
	})
}

func TestAccResourceKibanaAdvancedSettings_spacesAreIndependent(t *testing.T) {
	client := acctest.AdvancedSettingsKibanaClient(t)
	firstSpaceID := acctest.NewAdvancedSettingsTestSpace(t, client)
	secondSpaceID := acctest.NewAdvancedSettingsTestSpace(t, client)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"first_space_id":  config.StringVariable(firstSpaceID),
					"second_space_id": config.StringVariable(secondSpaceID),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_advanced_settings.first", "id", firstSpaceID),
					resource.TestCheckResourceAttr("elasticstack_kibana_advanced_settings.second", "id", secondSpaceID),
					checkKibanaSettings(client, firstSpaceID, map[string]string{"dateFormat:tz": `"Europe/Rome"`}),
					checkKibanaSettings(client, secondSpaceID, map[string]string{"dateFormat:tz": `"Europe/London"`}),
				),
			},
		},
	})
}

func TestAccResourceKibanaAdvancedSettings_changingSpaceIDForcesReplacement(t *testing.T) {
	client := acctest.AdvancedSettingsKibanaClient(t)
	opsSpaceID := acctest.NewAdvancedSettingsTestSpace(t, client)
	stagingSpaceID := acctest.NewAdvancedSettingsTestSpace(t, client)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("replace"),
				ConfigVariables:          config.Variables{"active_space_id": config.StringVariable(opsSpaceID)},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(advancedSettingsResource, "id", opsSpaceID),
					checkKibanaSettings(client, opsSpaceID, map[string]string{"dateFormat:tz": `"Europe/Rome"`}),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("replace"),
				ConfigVariables:          config.Variables{"active_space_id": config.StringVariable(stagingSpaceID)},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(advancedSettingsResource, plancheck.ResourceActionDestroyBeforeCreate),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(advancedSettingsResource, "id", stagingSpaceID),
					resource.TestCheckResourceAttr(advancedSettingsResource, "space_id", stagingSpaceID),
					checkKibanaSettings(client, stagingSpaceID, map[string]string{"dateFormat:tz": `"Europe/Rome"`}),
				),
			},
		},
	})
}

func TestAccResourceKibanaAdvancedSettings_createFailsWhenTheSpaceDoesNotExist(t *testing.T) {
	client := acctest.AdvancedSettingsKibanaClient(t)
	missingSpaceID := "advset-missing-" + sdkacctest.RandStringFromCharSet(8, sdkacctest.CharSetAlpha)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		CheckDestroy: func(_ *terraform.State) error {
			settings, diags := kibanaoapi.GetAdvancedSettings(context.Background(), client, missingSpaceID)
			if diags.HasError() {
				return fmt.Errorf("reading settings of space %q: %v", missingSpaceID, diags)
			}
			if _, written := settings["custom:teamNote"]; written {
				return fmt.Errorf("a setting was written to the nonexistent space %q", missingSpaceID)
			}
			return nil
		},
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          spaceVariables(missingSpaceID),
				ExpectError:              regexp.MustCompile(`(?s)must\s+exist`),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          spaceVariables(missingSpaceID),
				PlanOnly:                 true,
				ExpectNonEmptyPlan:       true,
			},
		},
	})
}

func TestAccResourceKibanaAdvancedSettings_kibanaConnection(t *testing.T) {
	client := acctest.AdvancedSettingsKibanaClient(t)
	spaceID := acctest.NewAdvancedSettingsTestSpace(t, client)
	acctest.PreCheckWithExplicitKibanaEndpoint(t)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          acctest.KibanaConnectionVariables(spaceVariables(spaceID)),
				Check:                    advancedSettingsKibanaConnectionChecks(client, spaceID),
			},
		},
	})
}

func advancedSettingsKibanaConnectionChecks(client *kibanaoapi.Client, spaceID string) resource.TestCheckFunc {
	checks := []resource.TestCheckFunc{
		resource.TestCheckResourceAttr(advancedSettingsResource, "id", spaceID),
		resource.TestCheckResourceAttr(advancedSettingsResource, "kibana_connection.#", "1"),
		resource.TestCheckResourceAttr(advancedSettingsResource, "kibana_connection.0.endpoints.0", acctest.KibanaConnectionEndpoint()),
		resource.TestCheckResourceAttr(advancedSettingsResource, "kibana_connection.0.insecure", "false"),
		checkKibanaSettings(client, spaceID, map[string]string{"dateFormat:tz": `"Europe/Rome"`}),
	}
	checks = append(checks, acctest.KibanaConnectionAuthChecks(advancedSettingsResource)...)

	return resource.ComposeAggregateTestCheckFunc(checks...)
}

func TestAccResourceKibanaAdvancedSettings_rejectsAnEmptyKey(t *testing.T) {
	client := acctest.AdvancedSettingsKibanaClient(t)
	spaceID := acctest.NewAdvancedSettingsTestSpace(t, client)
	settingsBefore := snapshotKibanaSettings(t, client, spaceID)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("empty_key"),
				ConfigVariables:          spaceVariables(spaceID),
				PlanOnly:                 true,
				ExpectError:              regexp.MustCompile(`(?s)character\s+count\s+must\s+be\s+between\s+1\s+and\s+1024`),
			},
		},
	})

	requireKibanaSettingsUnchanged(t, client, spaceID, settingsBefore)
}

func TestAccResourceKibanaAdvancedSettings_rejectsAKeyLongerThan1024Characters(t *testing.T) {
	client := acctest.AdvancedSettingsKibanaClient(t)
	spaceID := acctest.NewAdvancedSettingsTestSpace(t, client)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("long_key"),
				ConfigVariables: config.Variables{
					"space_id": config.StringVariable(spaceID),
					"long_key": config.StringVariable(strings.Repeat("k", 1025)),
				},
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`(?s)character\s+count\s+must\s+be\s+between\s+1\s+and\s+1024`),
			},
		},
	})
}

func TestAccResourceKibanaAdvancedSettings_rejectsAKeyDeclaredInBothMaps(t *testing.T) {
	client := acctest.AdvancedSettingsKibanaClient(t)
	spaceID := acctest.NewAdvancedSettingsTestSpace(t, client)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("duplicate_key"),
				ConfigVariables:          spaceVariables(spaceID),
				PlanOnly:                 true,
				ExpectError:              regexp.MustCompile(`(?s)"dateFormat:tz"\s+is\s+declared\s+in\s+both`),
			},
		},
	})
}

func TestAccResourceKibanaAdvancedSettings_sameKeyInDifferentResourcesIsNotChecked(t *testing.T) {
	client := acctest.AdvancedSettingsKibanaClient(t)
	spaceID := acctest.NewAdvancedSettingsTestSpace(t, client)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("two_resources"),
				ConfigVariables:          spaceVariables(spaceID),
				PlanOnly:                 true,
				ExpectNonEmptyPlan:       true,
			},
		},
	})
}

func TestAccResourceKibanaAdvancedSettings_rejectsInvalidJSON(t *testing.T) {
	client := acctest.AdvancedSettingsKibanaClient(t)
	spaceID := acctest.NewAdvancedSettingsTestSpace(t, client)
	settingsBefore := snapshotKibanaSettings(t, client, spaceID)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("invalid_json"),
				ConfigVariables:          spaceVariables(spaceID),
				PlanOnly:                 true,
				ExpectError:              regexp.MustCompile(`(?s)"truncate:maxHeight"\s+is\s+not\s+valid\s+JSON`),
			},
		},
	})

	requireKibanaSettingsUnchanged(t, client, spaceID, settingsBefore)
}

func TestAccResourceKibanaAdvancedSettings_rejectsTheJSONValueNull(t *testing.T) {
	client := acctest.AdvancedSettingsKibanaClient(t)
	spaceID := acctest.NewAdvancedSettingsTestSpace(t, client)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("json_null"),
				ConfigVariables:          spaceVariables(spaceID),
				PlanOnly:                 true,
				ExpectError:              regexp.MustCompile(`(?s)"defaultRoute".*remove\s+the\s+key`),
			},
		},
	})
}

func TestAccResourceKibanaAdvancedSettings_rejectsANullStringEntry(t *testing.T) {
	client := acctest.AdvancedSettingsKibanaClient(t)
	spaceID := acctest.NewAdvancedSettingsTestSpace(t, client)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("null_entry"),
				ConfigVariables:          spaceVariables(spaceID),
				PlanOnly:                 true,
				ExpectError:              regexp.MustCompile(`(?s)"defaultRoute".*remove\s+the\s+key`),
			},
		},
	})
}

func TestAccResourceKibanaAdvancedSettings_rejectsAnUnknownValueThatResolvesToInvalidJSON(t *testing.T) {
	client := acctest.AdvancedSettingsKibanaClient(t)
	spaceID := acctest.NewAdvancedSettingsTestSpace(t, client)
	settingsBefore := snapshotKibanaSettings(t, client, spaceID)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("unknown_value"),
				ConfigVariables:          spaceVariables(spaceID),
				ExpectError:              regexp.MustCompile(`(?s)"truncate:maxHeight"\s+is\s+not\s+valid\s+JSON`),
			},
		},
	})

	requireKibanaSettingsUnchanged(t, client, spaceID, settingsBefore)
}

func TestAccResourceKibanaAdvancedSettings_acceptsAJSONStringValue(t *testing.T) {
	client := acctest.AdvancedSettingsKibanaClient(t)
	spaceID := acctest.NewAdvancedSettingsTestSpace(t, client)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("json_string"),
				ConfigVariables:          spaceVariables(spaceID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(advancedSettingsResource, "json_settings.defaultRoute", `"/app/dashboards"`),
					checkKibanaSettings(client, spaceID, map[string]string{"defaultRoute": `"/app/dashboards"`}),
				),
			},
		},
	})
}

func TestAccResourceKibanaAdvancedSettings_updateAppliesAChangedValue(t *testing.T) {
	client := acctest.AdvancedSettingsKibanaClient(t)
	spaceID := acctest.NewAdvancedSettingsTestSpace(t, client)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("initial"),
				ConfigVariables:          spaceVariables(spaceID),
				Check:                    checkKibanaSettings(client, spaceID, map[string]string{"dateFormat:tz": `"Europe/Rome"`}),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("changed"),
				ConfigVariables:          spaceVariables(spaceID),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(advancedSettingsResource, plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(advancedSettingsResource, "string_settings.dateFormat:tz", "America/New_York"),
					checkKibanaSettings(client, spaceID, map[string]string{"dateFormat:tz": `"America/New_York"`}),
				),
			},
		},
	})
}

func TestAccResourceKibanaAdvancedSettings_updateAppliesAnAddedKey(t *testing.T) {
	client := acctest.AdvancedSettingsKibanaClient(t)
	spaceID := acctest.NewAdvancedSettingsTestSpace(t, client)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("initial"),
				ConfigVariables:          spaceVariables(spaceID),
				Check:                    checkKibanaSettings(client, spaceID, map[string]string{"dateFormat:tz": `"Europe/Rome"`}),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("added"),
				ConfigVariables:          spaceVariables(spaceID),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(advancedSettingsResource, plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(advancedSettingsResource, "string_settings.%", "2"),
					checkKibanaSettings(client, spaceID, map[string]string{
						"dateFormat:tz": `"Europe/Rome"`,
						"defaultRoute":  `"/app/dashboards"`,
					}),
				),
			},
		},
	})
}

func TestAccResourceKibanaAdvancedSettings_updateResetsARemovedKey(t *testing.T) {
	client := acctest.AdvancedSettingsKibanaClient(t)
	spaceID := acctest.NewAdvancedSettingsTestSpace(t, client)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("initial"),
				ConfigVariables:          spaceVariables(spaceID),
				Check: checkKibanaSettings(client, spaceID, map[string]string{
					"dateFormat:tz": `"Europe/Rome"`,
					"defaultRoute":  `"/app/dashboards"`,
				}),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("removed"),
				ConfigVariables:          spaceVariables(spaceID),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(advancedSettingsResource, plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(advancedSettingsResource, "string_settings.%", "1"),
					checkKibanaSettings(client, spaceID, map[string]string{"dateFormat:tz": `"Europe/Rome"`}),
					checkKibanaSettingsAtDefault(client, spaceID, "defaultRoute"),
				),
			},
		},
	})
}

func TestAccResourceKibanaAdvancedSettings_updateMovesAKeyBetweenMaps(t *testing.T) {
	client := acctest.AdvancedSettingsKibanaClient(t)
	spaceID := acctest.NewAdvancedSettingsTestSpace(t, client)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("initial"),
				ConfigVariables:          spaceVariables(spaceID),
				Check:                    checkKibanaSettings(client, spaceID, map[string]string{"custom:teamNote": `"5"`}),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("moved"),
				ConfigVariables:          spaceVariables(spaceID),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(advancedSettingsResource, plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckNoResourceAttr(advancedSettingsResource, "string_settings"),
					resource.TestCheckResourceAttr(advancedSettingsResource, "json_settings.custom:teamNote", "5"),
					checkKibanaSettings(client, spaceID, map[string]string{"custom:teamNote": `5`}),
				),
			},
		},
	})
}

func TestAccResourceKibanaAdvancedSettings_updateLeavesUnmanagedSettingsUntouched(t *testing.T) {
	client := acctest.AdvancedSettingsKibanaClient(t)
	spaceID := acctest.NewAdvancedSettingsTestSpace(t, client)
	var settingsBefore map[string]json.RawMessage

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					setKibanaSettings(t, client, spaceID, map[string]any{"defaultRoute": "/app/dashboards", "custom:existing": "x"})
				},
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("initial"),
				ConfigVariables:          spaceVariables(spaceID),
			},
			{
				PreConfig:                func() { settingsBefore = snapshotKibanaSettings(t, client, spaceID) },
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("changed"),
				ConfigVariables:          spaceVariables(spaceID),
				Check: resource.ComposeAggregateTestCheckFunc(
					checkKibanaSettings(client, spaceID, map[string]string{
						"custom:teamNote": `"changed"`,
						"custom:added":    `"new"`,
						"defaultRoute":    `"/app/dashboards"`,
						"custom:existing": `"x"`,
					}),
					checkKibanaSettingsAtDefault(client, spaceID, "custom:removed"),
					checkKibanaSettingsUnchanged(client, spaceID, &settingsBefore, "custom:teamNote", "custom:removed", "custom:added"),
				),
			},
		},
	})
}

func TestAccResourceKibanaAdvancedSettings_refreshRemovesTheResourceWhenTheSpaceIsDeletedOutOfBand(t *testing.T) {
	client := acctest.AdvancedSettingsKibanaClient(t)
	spaceID := acctest.NewAdvancedSettingsTestSpace(t, client)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          spaceVariables(spaceID),
				Check:                    checkKibanaSettings(client, spaceID, map[string]string{"dateFormat:tz": `"Europe/Rome"`}),
			},
			{
				PreConfig:                func() { deleteKibanaSpace(t, client, spaceID) },
				ProtoV6ProviderFactories: acctest.Providers,
				RefreshState:             true,
				Check: resource.ComposeAggregateTestCheckFunc(
					checkResourceRemovedFromState(advancedSettingsResource),
					checkKibanaSettings(client, spaceID, map[string]string{"dateFormat:tz": `"Europe/Rome"`}),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          spaceVariables(spaceID),
				PlanOnly:                 true,
				ExpectNonEmptyPlan:       true,
			},
		},
	})
}
func TestAccResourceKibanaAdvancedSettings_refreshDetectsAnOutOfBandChangeAndApplyCorrectsIt(t *testing.T) {
	client := acctest.AdvancedSettingsKibanaClient(t)
	spaceID := acctest.NewAdvancedSettingsTestSpace(t, client)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          spaceVariables(spaceID),
				Check:                    checkKibanaSettings(client, spaceID, map[string]string{"dateFormat:tz": `"Europe/Rome"`}),
			},
			{
				PreConfig: func() {
					setKibanaSettings(t, client, spaceID, map[string]any{"dateFormat:tz": "UTC"})
				},
				ProtoV6ProviderFactories: acctest.Providers,
				RefreshState:             true,
				ExpectNonEmptyPlan:       true,
				Check:                    resource.TestCheckResourceAttr(advancedSettingsResource, "string_settings.dateFormat:tz", "UTC"),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          spaceVariables(spaceID),
				ConfigPlanChecks:         expectStringSettingUpdatedTo("dateFormat:tz", "Europe/Rome"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(advancedSettingsResource, "string_settings.dateFormat:tz", "Europe/Rome"),
					checkKibanaSettings(client, spaceID, map[string]string{"dateFormat:tz": `"Europe/Rome"`}),
				),
			},
		},
	})
}

func TestAccResourceKibanaAdvancedSettings_refreshDetectsAnOutOfBandResetAndApplyReappliesTheSetting(t *testing.T) {
	client := acctest.AdvancedSettingsKibanaClient(t)
	spaceID := acctest.NewAdvancedSettingsTestSpace(t, client)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          spaceVariables(spaceID),
				Check:                    checkKibanaSettings(client, spaceID, map[string]string{"defaultRoute": `"/app/dashboards"`}),
			},
			{
				PreConfig: func() {
					setKibanaSettings(t, client, spaceID, map[string]any{"defaultRoute": nil})
				},
				ProtoV6ProviderFactories: acctest.Providers,
				RefreshState:             true,
				ExpectNonEmptyPlan:       true,
				Check: resource.ComposeAggregateTestCheckFunc(
					checkKibanaSettingsAtDefault(client, spaceID, "defaultRoute"),
					resource.TestCheckNoResourceAttr(advancedSettingsResource, "string_settings.defaultRoute"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          spaceVariables(spaceID),
				ConfigPlanChecks:         expectStringSettingUpdatedTo("defaultRoute", "/app/dashboards"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(advancedSettingsResource, "string_settings.defaultRoute", "/app/dashboards"),
					checkKibanaSettings(client, spaceID, map[string]string{"defaultRoute": `"/app/dashboards"`}),
				),
			},
		},
	})
}

func TestAccResourceKibanaAdvancedSettings_refreshIgnoresAnOutOfBandChangeToAnUnmanagedSetting(t *testing.T) {
	client := acctest.AdvancedSettingsKibanaClient(t)
	spaceID := acctest.NewAdvancedSettingsTestSpace(t, client)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          spaceVariables(spaceID),
				Check:                    checkKibanaSettings(client, spaceID, map[string]string{"dateFormat:tz": `"Europe/Rome"`}),
			},
			{
				PreConfig: func() {
					setKibanaSettings(t, client, spaceID, map[string]any{"truncate:maxHeight": 99})
				},
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          spaceVariables(spaceID),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckNoResourceAttr(advancedSettingsResource, "json_settings"),
					checkKibanaSettings(client, spaceID, map[string]string{
						"dateFormat:tz":      `"Europe/Rome"`,
						"truncate:maxHeight": `99`,
					}),
				),
			},
		},
	})
}

func TestAccResourceKibanaAdvancedSettings_refreshShowsAStringSettingThatBecameANonStringAsDrift(t *testing.T) {
	client := acctest.AdvancedSettingsKibanaClient(t)
	spaceID := acctest.NewAdvancedSettingsTestSpace(t, client)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          spaceVariables(spaceID),
				Check:                    checkKibanaSettings(client, spaceID, map[string]string{"custom:teamNote": `"hello"`}),
			},
			{
				PreConfig: func() {
					setKibanaSettings(t, client, spaceID, map[string]any{"custom:teamNote": 5})
				},
				ProtoV6ProviderFactories: acctest.Providers,
				RefreshState:             true,
				ExpectNonEmptyPlan:       true,
				Check:                    resource.TestCheckResourceAttr(advancedSettingsResource, "string_settings.custom:teamNote", "5"),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          spaceVariables(spaceID),
				ConfigPlanChecks:         expectStringSettingUpdatedTo("custom:teamNote", "hello"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(advancedSettingsResource, "string_settings.custom:teamNote", "hello"),
					checkKibanaSettings(client, spaceID, map[string]string{"custom:teamNote": `"hello"`}),
				),
			},
		},
	})
}

func TestAccResourceKibanaAdvancedSettings_destroyResetsOnlyTheManagedKeys(t *testing.T) {
	client := acctest.AdvancedSettingsKibanaClient(t)
	spaceID := acctest.NewAdvancedSettingsTestSpace(t, client)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		CheckDestroy: resource.ComposeAggregateTestCheckFunc(
			checkKibanaSettingsAtDefault(client, spaceID, "dateFormat:tz"),
			checkKibanaSettings(client, spaceID, map[string]string{"defaultRoute": `"/app/dashboards"`}),
		),
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					setKibanaSettings(t, client, spaceID, map[string]any{"defaultRoute": "/app/dashboards"})
				},
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          spaceVariables(spaceID),
				Check: checkKibanaSettings(client, spaceID, map[string]string{
					"dateFormat:tz": `"Europe/Rome"`,
					"defaultRoute":  `"/app/dashboards"`,
				}),
			},
		},
	})
}

func TestAccResourceKibanaAdvancedSettings_destroySucceedsWhenTheSpaceIsDeletedOutOfBand(t *testing.T) {
	client := acctest.AdvancedSettingsKibanaClient(t)
	spaceID := acctest.NewAdvancedSettingsTestSpace(t, client)
	var destroyReached bool

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		CheckDestroy: func(state *terraform.State) error {
			destroyReached = true
			return checkKibanaSettings(client, spaceID, map[string]string{"dateFormat:tz": `"Europe/Rome"`})(state)
		},
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          spaceVariables(spaceID),
				Check:                    checkKibanaSettings(client, spaceID, map[string]string{"dateFormat:tz": `"Europe/Rome"`}),
			},
			{
				// A plan-only step keeps the prior state and the test's final destroy does not refresh,
				// so the destroy runs against the deleted space instead of the resource being dropped first.
				// CheckDestroy is skipped when the state is empty, hence the destroyReached guard.
				PreConfig:                func() { deleteKibanaSpace(t, client, spaceID) },
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          spaceVariables(spaceID),
				PlanOnly:                 true,
				ExpectNonEmptyPlan:       true,
			},
		},
	})

	require.True(t, destroyReached, "the destroy did not run against the state that still held the resource")
}

func TestAccResourceKibanaAdvancedSettings_destroyWithNoManagedKeysChangesNothing(t *testing.T) {
	client := acctest.AdvancedSettingsKibanaClient(t)
	spaceID := acctest.NewAdvancedSettingsTestSpace(t, client)
	var settingsBefore map[string]json.RawMessage

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkKibanaSettingsUnchanged(client, spaceID, &settingsBefore),
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					setKibanaSettings(t, client, spaceID, map[string]any{"defaultRoute": "/app/dashboards", "custom:existing": "x"})
					settingsBefore = snapshotKibanaSettings(t, client, spaceID)
				},
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          spaceVariables(spaceID),
				Check:                    resource.TestCheckNoResourceAttr(advancedSettingsResource, "string_settings"),
			},
		},
	})
}

func TestAccResourceKibanaAdvancedSettings_destroySucceedsWhenAManagedKeyIsAlreadyAtItsDefault(t *testing.T) {
	client := acctest.AdvancedSettingsKibanaClient(t)
	spaceID := acctest.NewAdvancedSettingsTestSpace(t, client)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		CheckDestroy: resource.ComposeAggregateTestCheckFunc(
			resource.TestCheckResourceAttr(advancedSettingsResource, "string_settings.dateFormat:tz", "Europe/Rome"),
			checkKibanaSettingsAtDefault(client, spaceID, "dateFormat:tz", "defaultRoute"),
		),
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          spaceVariables(spaceID),
				Check: checkKibanaSettings(client, spaceID, map[string]string{
					"dateFormat:tz": `"Europe/Rome"`,
					"defaultRoute":  `"/app/dashboards"`,
				}),
			},
			{
				// A plan-only step keeps the prior state and the test's final destroy does not refresh,
				// so the destroy still manages the key that Kibana no longer holds.
				PreConfig: func() {
					setKibanaSettings(t, client, spaceID, map[string]any{"dateFormat:tz": nil})
				},
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          spaceVariables(spaceID),
				PlanOnly:                 true,
				ExpectNonEmptyPlan:       true,
			},
		},
	})
}

func TestAccResourceKibanaAdvancedSettings_jsonNumberFormattingDoesNotCauseDiff(t *testing.T) {
	client := acctest.AdvancedSettingsKibanaClient(t)
	spaceID := acctest.NewAdvancedSettingsTestSpace(t, client)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          spaceVariables(spaceID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(advancedSettingsResource, "json_settings.custom:ratio", "1.50"),
					resource.TestCheckResourceAttr(advancedSettingsResource, "json_settings.custom:whole", "1.0"),
					resource.TestCheckResourceAttr(advancedSettingsResource, "json_settings.custom:scaled", "1e2"),
					resource.TestCheckResourceAttr(advancedSettingsResource, "json_settings.custom:nested", `{"weights": [1.0, 2.50], "limit": 1e2}`),
					checkKibanaSettings(client, spaceID, map[string]string{
						"custom:ratio":  `1.5`,
						"custom:whole":  `1`,
						"custom:scaled": `100`,
						"custom:nested": `{"weights":[1,2.5],"limit":100}`,
					}),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          spaceVariables(spaceID),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
				},
			},
		},
	})
}

func TestAccResourceKibanaAdvancedSettings_createFailsWithAHintWhenAStringSettingHoldsANonStringValue(t *testing.T) {
	client := acctest.AdvancedSettingsKibanaClient(t)
	spaceID := acctest.NewAdvancedSettingsTestSpace(t, client)
	settingsBefore := snapshotKibanaSettings(t, client, spaceID)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          spaceVariables(spaceID),
				ExpectError:              regexp.MustCompile(`(?s)expected\s+value\s+of\s+type\s+\[boolean\]\s+but\s+got\s+\[string\].*non-string\s+values.*belong\s+in\s+json_settings`),
			},
		},
	})

	requireKibanaSettingsUnchanged(t, client, spaceID, settingsBefore)
}

func TestAccResourceKibanaAdvancedSettings_createFailsWhenKibanaRejectsAValueAndRecordsNothing(t *testing.T) {
	client := acctest.AdvancedSettingsKibanaClient(t)
	spaceID := acctest.NewAdvancedSettingsTestSpace(t, client)
	settingsBefore := snapshotKibanaSettings(t, client, spaceID)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          spaceVariables(spaceID),
				ExpectError:              regexp.MustCompile(`(?s)expected\s+value\s+of\s+type\s+\[number\]\s+but\s+got\s+\[string\]`),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          spaceVariables(spaceID),
				PlanOnly:                 true,
				ExpectNonEmptyPlan:       true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPreRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(advancedSettingsResource, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})

	requireKibanaSettingsUnchanged(t, client, spaceID, settingsBefore)
}

// The stack pins observability:streamsEnableQueryStreams through uiSettings.overrides in kibana-9.4.yml.
func TestAccResourceKibanaAdvancedSettings_createFailsWhenASettingIsPinnedByKibanaServerConfiguration(t *testing.T) {
	client := acctest.AdvancedSettingsKibanaClient(t)
	spaceID := acctest.NewAdvancedSettingsTestSpace(t, client)
	settingsBefore := snapshotKibanaSettings(t, client, spaceID)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          spaceVariables(spaceID),
				ExpectError:              regexp.MustCompile(`(?s)Unable\s+to\s+update\s+"observability:streamsEnableQueryStreams"\s+because\s+it\s+is\s+overridden`),
			},
		},
	})

	requireKibanaSettingsUnchanged(t, client, spaceID, settingsBefore)
}

func TestAccResourceKibanaAdvancedSettings_updateFailsWhenKibanaRejectsTheValueAndKeepsThePriorState(t *testing.T) {
	client := acctest.AdvancedSettingsKibanaClient(t)
	spaceID := acctest.NewAdvancedSettingsTestSpace(t, client)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("initial"),
				ConfigVariables:          spaceVariables(spaceID),
				Check:                    resource.TestCheckResourceAttr(advancedSettingsResource, "string_settings.dateFormat:tz", "Europe/Rome"),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("rejected"),
				ConfigVariables:          spaceVariables(spaceID),
				ExpectError:              regexp.MustCompile(`(?s)Invalid\s+timezone:\s+Not/AZone`),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("initial"),
				ConfigVariables:          spaceVariables(spaceID),
				PlanOnly:                 true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPreRefresh: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
				},
			},
		},
	})
}

func TestAccResourceKibanaAdvancedSettings_reformattingAJSONValueIsAnInPlaceUpdateThatChangesNothingInKibana(t *testing.T) {
	client := acctest.AdvancedSettingsKibanaClient(t)
	spaceID := acctest.NewAdvancedSettingsTestSpace(t, client)
	kibanaSettings := map[string]string{
		"custom:ratio":  `1.5`,
		"custom:list":   `["a","b"]`,
		"custom:object": `{"a":1,"b":{"c":2,"d":3}}`,
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("initial"),
				ConfigVariables:          spaceVariables(spaceID),
				Check:                    checkKibanaSettings(client, spaceID, kibanaSettings),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("reformatted"),
				ConfigVariables:          spaceVariables(spaceID),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(advancedSettingsResource, plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(advancedSettingsResource, "json_settings.custom:ratio", "1.50"),
					resource.TestCheckResourceAttr(advancedSettingsResource, "json_settings.custom:list", `[ "a",   "b" ]`),
					resource.TestCheckResourceAttr(advancedSettingsResource, "json_settings.custom:object", `{"b":{"d":3,"c":2},"a":1}`),
					checkKibanaSettings(client, spaceID, kibanaSettings),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("reformatted"),
				ConfigVariables:          spaceVariables(spaceID),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
				},
			},
		},
	})
}
