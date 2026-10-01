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
	"fmt"
	"reflect"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	kibanaoapi "github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

const resourceName = "elasticstack_kibana_advanced_settings.test"

// The global settings API exists since Kibana 8.7.0, but the first global
// settings (custom branding) are registered in 8.8.0.
var minGlobalSettingsTestVersion = version.Must(version.NewVersion("8.8.0"))

const globalSettingKey = "xpackCustomBranding:pageTitle"

func TestAccResourceKibanaAdvancedSettings(t *testing.T) {
	spaceID := "advanced-settings-" + sdkacctest.RandStringFromCharSet(6, sdkacctest.CharSetAlphaNum)
	vars := config.Variables{"space_id": config.StringVariable(spaceID)}

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          vars,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "id", spaceID),
					resource.TestCheckResourceAttr(resourceName, "space_id", spaceID),
					resource.TestCheckResourceAttr(resourceName, "global", "false"),
					resource.TestCheckResourceAttr(resourceName, "settings.%", "5"),
					resource.TestCheckResourceAttr(resourceName, "settings.dateFormat:tz", `"Europe/Berlin"`),
					resource.TestCheckResourceAttr(resourceName, "settings.discover:sampleSize", `321`),
					resource.TestCheckResourceAttr(resourceName, "settings.courier:ignoreFilterIfFieldNotInIndex", `true`),
					resource.TestCheckResourceAttr(resourceName, "settings.defaultColumns", `["host.name","message"]`),
					resource.TestCheckResourceAttr(resourceName, "settings.timepicker:timeDefaults", `"{\"from\":\"now-30m\",\"to\":\"now\"}"`),
					checkKibanaSettings(spaceID, false, map[string]any{
						"dateFormat:tz":                         "Europe/Berlin",
						"discover:sampleSize":                   float64(321),
						"courier:ignoreFilterIfFieldNotInIndex": true,
						"defaultColumns":                        []any{"host.name", "message"},
						"timepicker:timeDefaults":               `{"from":"now-30m","to":"now"}`,
					}),
				),
			},
			{
				// Changes and adds settings, and resets the settings removed
				// from the configuration to their defaults.
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables:          vars,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "id", spaceID),
					resource.TestCheckResourceAttr(resourceName, "settings.%", "4"),
					resource.TestCheckResourceAttr(resourceName, "settings.dateFormat:tz", `"UTC"`),
					resource.TestCheckResourceAttr(resourceName, "settings.discover:sampleSize", `321`),
					resource.TestCheckResourceAttr(resourceName, "settings.defaultColumns", `["message"]`),
					resource.TestCheckResourceAttr(resourceName, "settings.csv:separator", `";"`),
					resource.TestCheckNoResourceAttr(resourceName, "settings.courier:ignoreFilterIfFieldNotInIndex"),
					checkKibanaSettings(spaceID, false, map[string]any{
						"dateFormat:tz":                         "UTC",
						"discover:sampleSize":                   float64(321),
						"defaultColumns":                        []any{"message"},
						"csv:separator":                         ";",
						"courier:ignoreFilterIfFieldNotInIndex": nil,
						"timepicker:timeDefaults":               nil,
					}),
				),
			},
			{
				// Settings changed or reset outside Terraform are detected as
				// drift and restored.
				PreConfig: func() {
					updateKibanaSettings(t, spaceID, false, map[string]any{
						"dateFormat:tz": "America/New_York",
						"csv:separator": nil,
					})
				},
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables:          vars,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "settings.dateFormat:tz", `"UTC"`),
					resource.TestCheckResourceAttr(resourceName, "settings.csv:separator", `";"`),
					checkKibanaSettings(spaceID, false, map[string]any{
						"dateFormat:tz": "UTC",
						"csv:separator": ";",
					}),
				),
			},
			{
				// Deleting the space outside Terraform removes the settings
				// from state, so they are created again with the space.
				PreConfig: func() {
					deleteSpace(t, spaceID)
				},
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables:          vars,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "id", spaceID),
					checkKibanaSettings(spaceID, false, map[string]any{
						"dateFormat:tz": "UTC",
						"csv:separator": ";",
					}),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("import"),
				ConfigVariables:          vars,
				ResourceName:             resourceName,
				ImportState:              true,
				ImportStateId:            spaceID,
				ImportStateVerify:        true,
				// Settings are not imported; the configuration declares
				// which settings the resource manages.
				ImportStateVerifyIgnore: []string{"settings"},
			},
		},
	})
}

func TestAccResourceKibanaAdvancedSettingsGlobal(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minGlobalSettingsTestVersion, versionutils.FlavorAny)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkKibanaSettingsReset(clients.DefaultSpaceID, true, globalSettingKey),
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "id", "global"),
					resource.TestCheckResourceAttr(resourceName, "global", "true"),
					resource.TestCheckResourceAttr(resourceName, "settings."+globalSettingKey, `"Terraform"`),
					checkKibanaSettings(clients.DefaultSpaceID, true, map[string]any{globalSettingKey: "Terraform"}),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "id", "global"),
					resource.TestCheckResourceAttr(resourceName, "settings."+globalSettingKey, `"Terraform updated"`),
					checkKibanaSettings(clients.DefaultSpaceID, true, map[string]any{globalSettingKey: "Terraform updated"}),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ResourceName:             resourceName,
				ImportState:              true,
				ImportStateId:            "global",
				ImportStateVerify:        true,
				ImportStateVerifyIgnore:  []string{"settings"},
			},
		},
	})
}

func getKibanaSettings(spaceID string, global bool) (map[string]kibanaoapi.AdvancedSetting, error) {
	client, err := clients.NewAcceptanceTestingKibanaScopedClient()
	if err != nil {
		return nil, err
	}
	settings, diags := kibanaoapi.GetAdvancedSettings(context.Background(), client.GetKibanaOapiClient(), spaceID, global)
	if diags.HasError() {
		return nil, fmt.Errorf("failed to read advanced settings: %v", diags)
	}
	return settings, nil
}

func updateKibanaSettings(t *testing.T, spaceID string, global bool, changes map[string]any) {
	t.Helper()
	client, err := clients.NewAcceptanceTestingKibanaScopedClient()
	if err != nil {
		t.Fatalf("failed to create Kibana client: %v", err)
	}
	if _, diags := kibanaoapi.UpdateAdvancedSettings(context.Background(), client.GetKibanaOapiClient(), spaceID, global, changes); diags.HasError() {
		t.Fatalf("failed to update advanced settings: %v", diags)
	}
}

func deleteSpace(t *testing.T, spaceID string) {
	t.Helper()
	client, err := clients.NewAcceptanceTestingKibanaScopedClient()
	if err != nil {
		t.Fatalf("failed to create Kibana client: %v", err)
	}
	if diags := kibanaoapi.DeleteSpace(context.Background(), client.GetKibanaOapiClient(), spaceID); diags.HasError() {
		t.Fatalf("failed to delete space %q: %v", spaceID, diags)
	}
}

// checkKibanaSettings verifies the values Kibana reports for the given
// settings. A nil expected value asserts that the setting is at its default.
func checkKibanaSettings(spaceID string, global bool, expected map[string]any) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		settings, err := getKibanaSettings(spaceID, global)
		if err != nil {
			return err
		}
		for key, want := range expected {
			setting, ok := settings[key]
			if want == nil {
				if ok {
					return fmt.Errorf("expected setting %q to be reset, got %v", key, setting.Value)
				}
				continue
			}
			if !ok {
				return fmt.Errorf("expected setting %q to be %v, but it is not set", key, want)
			}
			if !reflect.DeepEqual(setting.Value, want) {
				return fmt.Errorf("expected setting %q to be %v, got %v", key, want, setting.Value)
			}
		}
		return nil
	}
}

func checkKibanaSettingsReset(spaceID string, global bool, keys ...string) resource.TestCheckFunc {
	expected := make(map[string]any, len(keys))
	for _, key := range keys {
		expected[key] = nil
	}
	return checkKibanaSettings(spaceID, global, expected)
}
