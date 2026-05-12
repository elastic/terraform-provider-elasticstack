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

package settings_test

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

//go:embed testdata/TestAccResourceClusterSettings/create/main.tf
var fromSDKCreateConfig string

func TestAccResourceClusterSettings(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceClusterSettingsDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_cluster_settings.test", "id"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_cluster_settings.test", "persistent.setting.#", "3"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_cluster_settings.test", "transient.setting.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs("elasticstack_elasticsearch_cluster_settings.test", "persistent.setting.*",
						map[string]string{
							"name":  "indices.lifecycle.poll_interval",
							"value": "10m",
						}),
					resource.TestCheckTypeSetElemNestedAttrs("elasticstack_elasticsearch_cluster_settings.test", "persistent.setting.*",
						map[string]string{
							"name":  "indices.recovery.max_bytes_per_sec",
							"value": "50mb",
						}),
					resource.TestCheckTypeSetElemNestedAttrs("elasticstack_elasticsearch_cluster_settings.test", "persistent.setting.*",
						map[string]string{
							"name":  "indices.breaker.total.limit",
							"value": "65%",
						}),
					resource.TestCheckTypeSetElemNestedAttrs("elasticstack_elasticsearch_cluster_settings.test", "transient.setting.*",
						map[string]string{
							"name":  "indices.breaker.total.limit",
							"value": "60%",
						}),
					resource.TestCheckTypeSetElemNestedAttrs("elasticstack_elasticsearch_cluster_settings.test", "transient.setting.*",
						map[string]string{
							"name":         "xpack.security.audit.logfile.events.include",
							"value_list.#": "2",
						}),
					resource.TestCheckTypeSetElemAttr("elasticstack_elasticsearch_cluster_settings.test", "transient.setting.*.value_list.*", "ACCESS_DENIED"),
					resource.TestCheckTypeSetElemAttr("elasticstack_elasticsearch_cluster_settings.test", "transient.setting.*.value_list.*", "ACCESS_GRANTED"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("transient_update"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_cluster_settings.test", "id"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_cluster_settings.test", "persistent.setting.#", "3"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_cluster_settings.test", "transient.setting.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs("elasticstack_elasticsearch_cluster_settings.test", "persistent.setting.*",
						map[string]string{
							"name":  "indices.lifecycle.poll_interval",
							"value": "10m",
						}),
					resource.TestCheckTypeSetElemNestedAttrs("elasticstack_elasticsearch_cluster_settings.test", "persistent.setting.*",
						map[string]string{
							"name":  "indices.recovery.max_bytes_per_sec",
							"value": "50mb",
						}),
					resource.TestCheckTypeSetElemNestedAttrs("elasticstack_elasticsearch_cluster_settings.test", "persistent.setting.*",
						map[string]string{
							"name":  "indices.breaker.total.limit",
							"value": "65%",
						}),
					resource.TestCheckTypeSetElemNestedAttrs("elasticstack_elasticsearch_cluster_settings.test", "transient.setting.*",
						map[string]string{
							"name":  "indices.breaker.total.limit",
							"value": "70%",
						}),
					resource.TestCheckTypeSetElemNestedAttrs("elasticstack_elasticsearch_cluster_settings.test", "transient.setting.*",
						map[string]string{
							"name":         "xpack.security.audit.logfile.events.include",
							"value_list.#": "2",
						}),
					resource.TestCheckTypeSetElemAttr("elasticstack_elasticsearch_cluster_settings.test", "transient.setting.*.value_list.*", "ACCESS_DENIED"),
					resource.TestCheckTypeSetElemAttr("elasticstack_elasticsearch_cluster_settings.test", "transient.setting.*.value_list.*", "AUTHENTICATION_SUCCESS"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_cluster_settings.test", "id"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_cluster_settings.test", "persistent.setting.#", "4"),
					resource.TestCheckTypeSetElemNestedAttrs("elasticstack_elasticsearch_cluster_settings.test", "persistent.setting.*",
						map[string]string{
							"name":  "indices.lifecycle.poll_interval",
							"value": "15m",
						}),
					resource.TestCheckTypeSetElemNestedAttrs("elasticstack_elasticsearch_cluster_settings.test", "persistent.setting.*",
						map[string]string{
							"name":  "indices.recovery.max_bytes_per_sec",
							"value": "40mb",
						}),
					resource.TestCheckTypeSetElemNestedAttrs("elasticstack_elasticsearch_cluster_settings.test", "persistent.setting.*",
						map[string]string{
							"name":  "indices.breaker.total.limit",
							"value": "60%",
						}),
					resource.TestCheckTypeSetElemNestedAttrs("elasticstack_elasticsearch_cluster_settings.test", "persistent.setting.*",
						map[string]string{
							"name":         "xpack.security.audit.logfile.events.include",
							"value_list.#": "2",
						}),
					resource.TestCheckTypeSetElemAttr("elasticstack_elasticsearch_cluster_settings.test", "persistent.setting.*.value_list.*", "ACCESS_DENIED"),
					resource.TestCheckTypeSetElemAttr("elasticstack_elasticsearch_cluster_settings.test", "persistent.setting.*.value_list.*", "AUTHENTICATION_SUCCESS"),
					resource.TestCheckNoResourceAttr("elasticstack_elasticsearch_cluster_settings.test", "transient.setting.#"),
					// The previous step had a transient "indices.breaker.total.limit"
					// setting; this step removes the transient block entirely. The
					// remote API state must reflect that removal.
					checkRemoteSettingAbsent("transient", "indices.breaker.total.limit"),
					checkRemoteSettingAbsent("transient", "xpack.security.audit.logfile.events.include"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("persistent_value_list_update"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_cluster_settings.test", "id"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_cluster_settings.test", "persistent.setting.#", "4"),
					resource.TestCheckTypeSetElemNestedAttrs("elasticstack_elasticsearch_cluster_settings.test", "persistent.setting.*",
						map[string]string{
							"name":  "indices.lifecycle.poll_interval",
							"value": "15m",
						}),
					resource.TestCheckTypeSetElemNestedAttrs("elasticstack_elasticsearch_cluster_settings.test", "persistent.setting.*",
						map[string]string{
							"name":  "indices.recovery.max_bytes_per_sec",
							"value": "40mb",
						}),
					resource.TestCheckTypeSetElemNestedAttrs("elasticstack_elasticsearch_cluster_settings.test", "persistent.setting.*",
						map[string]string{
							"name":  "indices.breaker.total.limit",
							"value": "60%",
						}),
					resource.TestCheckTypeSetElemNestedAttrs("elasticstack_elasticsearch_cluster_settings.test", "persistent.setting.*",
						map[string]string{
							"name":         "xpack.security.audit.logfile.events.include",
							"value_list.#": "2",
						}),
					resource.TestCheckTypeSetElemAttr("elasticstack_elasticsearch_cluster_settings.test", "persistent.setting.*.value_list.*", "ACCESS_DENIED"),
					resource.TestCheckTypeSetElemAttr("elasticstack_elasticsearch_cluster_settings.test", "persistent.setting.*.value_list.*", "ACCESS_GRANTED"),
					resource.TestCheckNoResourceAttr("elasticstack_elasticsearch_cluster_settings.test", "transient.setting.#"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("import"),
				ResourceName:             "elasticstack_elasticsearch_cluster_settings.test",
				ImportState:              true,
				// ImportStateVerify is intentionally omitted: Read only reflects
				// settings already tracked in state, so a freshly-imported state
				// (which has no persistent/transient blocks) cannot reproduce the
				// prior plan's setting values. The custom ImportStateCheck below
				// verifies the imported ID is well-formed instead.
				ImportStateCheck: func(is []*terraform.InstanceState) error {
					if len(is) != 1 {
						return fmt.Errorf("expected 1 imported instance state, got %d", len(is))
					}

					importedID := is[0].ID
					if importedID == "" {
						return fmt.Errorf("expected imported resource ID to be set")
					}

					if !strings.HasSuffix(importedID, "/cluster-settings") {
						return fmt.Errorf("expected imported resource ID [%s] to end with /cluster-settings", importedID)
					}

					if is[0].Attributes["id"] != importedID {
						return fmt.Errorf("expected imported id attribute [%s] to equal resource ID [%s]", is[0].Attributes["id"], importedID)
					}

					return nil
				},
			},
		},
	})
}

func TestAccResourceClusterSettingsPersistentOnly(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceClusterSettingsDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_cluster_settings.test_persistent", "id"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_cluster_settings.test_persistent", "persistent.setting.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("elasticstack_elasticsearch_cluster_settings.test_persistent", "persistent.setting.*",
						map[string]string{
							"name":  "indices.lifecycle.poll_interval",
							"value": "10m",
						}),
					resource.TestCheckNoResourceAttr("elasticstack_elasticsearch_cluster_settings.test_persistent", "transient.setting.#"),
				),
			},
		},
	})
}

// TestAccResourceClusterSettingsFromSDK verifies that state created by the last
// released SDKv2-based provider (v0.14.5) re-applies cleanly under the Plugin
// Framework migration with no plan diffs and no state mutation.
func TestAccResourceClusterSettingsFromSDK(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceClusterSettingsDestroy,
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"elasticstack": {
						Source:            "elastic/elasticstack",
						VersionConstraint: "0.14.5",
					},
				},
				Config: fromSDKCreateConfig,
				Check: resource.ComposeTestCheckFunc(
					// Under the SDK v0.14.5 schema persistent / transient were
					// list-of-one blocks, so attribute paths use ".0.setting".
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_cluster_settings.test", "id"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_cluster_settings.test", "persistent.0.setting.#", "3"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_cluster_settings.test", "transient.0.setting.#", "2"),
				),
			},
			{
				// Under the new PF schema (version 1) the same logical state
				// must produce no plan diff thanks to the v0->v1 state upgrader.
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("from-sdk"),
				PlanOnly:                 true,
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("from-sdk"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_cluster_settings.test", "id"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_cluster_settings.test", "persistent.setting.#", "3"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_cluster_settings.test", "transient.setting.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs("elasticstack_elasticsearch_cluster_settings.test", "transient.setting.*",
						map[string]string{
							"name":         "xpack.security.audit.logfile.events.include",
							"value_list.#": "2",
						}),
				),
			},
		},
	})
}

func checkResourceClusterSettingsDestroy(s *terraform.State) error {
	client, err := clients.NewAcceptanceTestingElasticsearchScopedClient()
	if err != nil {
		return err
	}

	listOfSettings := []string{
		"indices.lifecycle.poll_interval",
		"indices.recovery.max_bytes_per_sec",
		"indices.breaker.total.limit",
		"xpack.security.audit.logfile.events.include",
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "elasticstack_elasticsearch_cluster_settings" {
			continue
		}

		typedClient, err := client.GetESClient()
		if err != nil {
			return err
		}
		res, err := typedClient.Cluster.GetSettings().FlatSettings(true).Do(context.Background())
		if err != nil {
			return err
		}

		for _, setting := range listOfSettings {
			if v, ok := res.Persistent[setting]; ok {
				return fmt.Errorf(`Setting "%s=%s" still in the persistent cluster settings, but it should be removed`, setting, string(v))
			}
			if v, ok := res.Transient[setting]; ok {
				return fmt.Errorf(`Setting "%s=%s" still in the transient cluster settings, but it should be removed`, setting, string(v))
			}
		}
	}
	return nil
}

// checkRemoteSettingAbsent asserts via the live ES API that the named setting
// is no longer present in the given category (persistent/transient). This
// catches regressions in updateRemovedSettings / Update where state may say a
// setting was removed but the PUT didn't actually null it out remotely.
func checkRemoteSettingAbsent(category, setting string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		client, err := clients.NewAcceptanceTestingElasticsearchScopedClient()
		if err != nil {
			return err
		}
		typedClient, err := client.GetESClient()
		if err != nil {
			return err
		}
		res, err := typedClient.Cluster.GetSettings().FlatSettings(true).Do(context.Background())
		if err != nil {
			return err
		}

		var present map[string]json.RawMessage
		switch category {
		case "persistent":
			present = res.Persistent
		case "transient":
			present = res.Transient
		default:
			return fmt.Errorf("unknown category %q", category)
		}

		if v, ok := present[setting]; ok {
			return fmt.Errorf(`expected %s setting %q to be removed, but it is still set to %s`, category, setting, string(v))
		}
		return nil
	}
}
