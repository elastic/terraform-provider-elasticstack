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

package cluster_test

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccResourceClusterSettings(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		CheckDestroy:             checkResourceClusterSettingsDestroy,
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				ConfigDirectory: acctest.NamedTestCaseDirectory("create"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_cluster_settings.test", "id"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_cluster_settings.test", "persistent.0.setting.#", "3"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_cluster_settings.test", "transient.0.setting.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs("elasticstack_elasticsearch_cluster_settings.test", "persistent.0.setting.*",
						map[string]string{
							"name":         "indices.lifecycle.poll_interval",
							"value":        "10m",
							"value_list.#": "0",
						}),
					resource.TestCheckTypeSetElemNestedAttrs("elasticstack_elasticsearch_cluster_settings.test", "persistent.0.setting.*",
						map[string]string{
							"name":         "indices.recovery.max_bytes_per_sec",
							"value":        "50mb",
							"value_list.#": "0",
						}),
					resource.TestCheckTypeSetElemNestedAttrs("elasticstack_elasticsearch_cluster_settings.test", "persistent.0.setting.*",
						map[string]string{
							"name":         "indices.breaker.total.limit",
							"value":        "65%",
							"value_list.#": "0",
						}),
					resource.TestCheckTypeSetElemNestedAttrs("elasticstack_elasticsearch_cluster_settings.test", "transient.0.setting.*",
						map[string]string{
							"name":         "indices.breaker.total.limit",
							"value":        "60%",
							"value_list.#": "0",
						}),
					resource.TestCheckTypeSetElemNestedAttrs("elasticstack_elasticsearch_cluster_settings.test", "transient.0.setting.*",
						map[string]string{
							"name":         "xpack.security.audit.logfile.events.include",
							"value":        "",
							"value_list.#": "2",
						}),
					resource.TestCheckTypeSetElemAttr("elasticstack_elasticsearch_cluster_settings.test", "transient.0.setting.*.value_list.*", "ACCESS_DENIED"),
					resource.TestCheckTypeSetElemAttr("elasticstack_elasticsearch_cluster_settings.test", "transient.0.setting.*.value_list.*", "ACCESS_GRANTED"),
				),
			},
			{
				ConfigDirectory: acctest.NamedTestCaseDirectory("transient_update"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_cluster_settings.test", "id"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_cluster_settings.test", "persistent.0.setting.#", "3"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_cluster_settings.test", "transient.0.setting.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs("elasticstack_elasticsearch_cluster_settings.test", "persistent.0.setting.*",
						map[string]string{
							"name":         "indices.lifecycle.poll_interval",
							"value":        "10m",
							"value_list.#": "0",
						}),
					resource.TestCheckTypeSetElemNestedAttrs("elasticstack_elasticsearch_cluster_settings.test", "persistent.0.setting.*",
						map[string]string{
							"name":         "indices.recovery.max_bytes_per_sec",
							"value":        "50mb",
							"value_list.#": "0",
						}),
					resource.TestCheckTypeSetElemNestedAttrs("elasticstack_elasticsearch_cluster_settings.test", "persistent.0.setting.*",
						map[string]string{
							"name":         "indices.breaker.total.limit",
							"value":        "65%",
							"value_list.#": "0",
						}),
					resource.TestCheckTypeSetElemNestedAttrs("elasticstack_elasticsearch_cluster_settings.test", "transient.0.setting.*",
						map[string]string{
							"name":         "indices.breaker.total.limit",
							"value":        "70%",
							"value_list.#": "0",
						}),
					resource.TestCheckTypeSetElemNestedAttrs("elasticstack_elasticsearch_cluster_settings.test", "transient.0.setting.*",
						map[string]string{
							"name":         "xpack.security.audit.logfile.events.include",
							"value":        "",
							"value_list.#": "2",
						}),
					resource.TestCheckTypeSetElemAttr("elasticstack_elasticsearch_cluster_settings.test", "transient.0.setting.*.value_list.*", "ACCESS_DENIED"),
					resource.TestCheckTypeSetElemAttr("elasticstack_elasticsearch_cluster_settings.test", "transient.0.setting.*.value_list.*", "AUTHENTICATION_SUCCESS"),
				),
			},
			{
				ConfigDirectory: acctest.NamedTestCaseDirectory("update"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_cluster_settings.test", "id"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_cluster_settings.test", "persistent.0.setting.#", "4"),
					resource.TestCheckTypeSetElemNestedAttrs("elasticstack_elasticsearch_cluster_settings.test", "persistent.0.setting.*",
						map[string]string{
							"name":         "indices.lifecycle.poll_interval",
							"value":        "15m",
							"value_list.#": "0",
						}),
					resource.TestCheckTypeSetElemNestedAttrs("elasticstack_elasticsearch_cluster_settings.test", "persistent.0.setting.*",
						map[string]string{
							"name":         "indices.recovery.max_bytes_per_sec",
							"value":        "40mb",
							"value_list.#": "0",
						}),
					resource.TestCheckTypeSetElemNestedAttrs("elasticstack_elasticsearch_cluster_settings.test", "persistent.0.setting.*",
						map[string]string{
							"name":         "indices.breaker.total.limit",
							"value":        "60%",
							"value_list.#": "0",
						}),
					resource.TestCheckTypeSetElemNestedAttrs("elasticstack_elasticsearch_cluster_settings.test", "persistent.0.setting.*",
						map[string]string{
							"name":         "xpack.security.audit.logfile.events.include",
							"value":        "",
							"value_list.#": "2",
						}),
					resource.TestCheckTypeSetElemAttr("elasticstack_elasticsearch_cluster_settings.test", "persistent.0.setting.*.value_list.*", "ACCESS_DENIED"),
					resource.TestCheckTypeSetElemAttr("elasticstack_elasticsearch_cluster_settings.test", "persistent.0.setting.*.value_list.*", "AUTHENTICATION_SUCCESS"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_cluster_settings.test", "transient.#", "0"),
				),
			},
			{
				ConfigDirectory: acctest.NamedTestCaseDirectory("persistent_value_list_update"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_cluster_settings.test", "id"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_cluster_settings.test", "persistent.0.setting.#", "4"),
					resource.TestCheckTypeSetElemNestedAttrs("elasticstack_elasticsearch_cluster_settings.test", "persistent.0.setting.*",
						map[string]string{
							"name":         "indices.lifecycle.poll_interval",
							"value":        "15m",
							"value_list.#": "0",
						}),
					resource.TestCheckTypeSetElemNestedAttrs("elasticstack_elasticsearch_cluster_settings.test", "persistent.0.setting.*",
						map[string]string{
							"name":         "indices.recovery.max_bytes_per_sec",
							"value":        "40mb",
							"value_list.#": "0",
						}),
					resource.TestCheckTypeSetElemNestedAttrs("elasticstack_elasticsearch_cluster_settings.test", "persistent.0.setting.*",
						map[string]string{
							"name":         "indices.breaker.total.limit",
							"value":        "60%",
							"value_list.#": "0",
						}),
					resource.TestCheckTypeSetElemNestedAttrs("elasticstack_elasticsearch_cluster_settings.test", "persistent.0.setting.*",
						map[string]string{
							"name":         "xpack.security.audit.logfile.events.include",
							"value":        "",
							"value_list.#": "2",
						}),
					resource.TestCheckTypeSetElemAttr("elasticstack_elasticsearch_cluster_settings.test", "persistent.0.setting.*.value_list.*", "ACCESS_DENIED"),
					resource.TestCheckTypeSetElemAttr("elasticstack_elasticsearch_cluster_settings.test", "persistent.0.setting.*.value_list.*", "ACCESS_GRANTED"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_cluster_settings.test", "transient.#", "0"),
				),
			},
			{
				ResourceName: "elasticstack_elasticsearch_cluster_settings.test",
				ImportState:  true,
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
		PreCheck:                 func() { acctest.PreCheck(t) },
		CheckDestroy:             checkResourceClusterSettingsDestroy,
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				ConfigDirectory: acctest.NamedTestCaseDirectory("read"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_cluster_settings.test_persistent", "id"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_cluster_settings.test_persistent", "persistent.0.setting.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("elasticstack_elasticsearch_cluster_settings.test_persistent", "persistent.0.setting.*",
						map[string]string{
							"name":         "indices.lifecycle.poll_interval",
							"value":        "10m",
							"value_list.#": "0",
						}),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_cluster_settings.test_persistent", "transient.#", "0"),
				),
			},
		},
	})
}

func testAccResourceClusterSettingsTransientUpdate() string {
	return `
provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_cluster_settings" "test" {
  persistent {
    setting {
      name  = "indices.lifecycle.poll_interval"
      value = "10m"
    }
    setting {
      name  = "indices.recovery.max_bytes_per_sec"
      value = "50mb"
    }
    setting {
      name  = "indices.breaker.total.limit"
      value = "65%"
    }
  }

  transient {
    setting {
      name  = "indices.breaker.total.limit"
      value = "70%"
    }
    setting {
      name       = "xpack.security.audit.logfile.events.include"
      value_list = ["ACCESS_DENIED", "AUTHENTICATION_SUCCESS"]
    }
  }
}
`
}

func testAccResourceClusterSettingsPersistentValueListUpdate() string {
	return `
provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_cluster_settings" "test" {
  persistent {
    setting {
      name  = "indices.lifecycle.poll_interval"
      value = "15m"
    }
    setting {
      name  = "indices.recovery.max_bytes_per_sec"
      value = "40mb"
    }
    setting {
      name  = "indices.breaker.total.limit"
      value = "60%"
    }
    setting {
      name       = "xpack.security.audit.logfile.events.include"
      value_list = ["ACCESS_DENIED", "ACCESS_GRANTED"]
    }
  }
}
`
}

func checkResourceClusterSettingsDestroy(s *terraform.State) error {
	client, err := clients.NewAcceptanceTestingClient()
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

		esClient, err := client.GetESClient()
		if err != nil {
			return err
		}
		req := esClient.Cluster.GetSettings.WithFlatSettings(true)
		res, err := esClient.Cluster.GetSettings(req)
		if err != nil {
			return err
		}
		defer res.Body.Close()

		clusterSettings := make(map[string]any)
		if err := json.NewDecoder(res.Body).Decode(&clusterSettings); err != nil {
			return err
		}

		if clusterSettings["persistent"] != 0 {
			settings := clusterSettings["persistent"].(map[string]any)
			for _, s := range listOfSettings {
				if v, ok := settings[s]; ok {
					return fmt.Errorf(`Setting "%s=%s" still in the cluster, but it should be removed`, s, v)
				}
			}
		}
	}
	return nil
}
