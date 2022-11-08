package cluster_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccResourceClusterSettings(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		CheckDestroy:      checkResourceClusterSettingsDestroy,
		ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceClusterSettingsCreate(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckTypeSetElemNestedAttrs("elasticstack_elasticsearch_cluster_settings.test", "persistent.0.setting.*",
						map[string]string{
							"name":  "indices.lifecycle.poll_interval",
							"value": "10m",
						}),
					resource.TestCheckTypeSetElemNestedAttrs("elasticstack_elasticsearch_cluster_settings.test", "persistent.0.setting.*",
						map[string]string{
							"name":  "indices.recovery.max_bytes_per_sec",
							"value": "50mb",
						}),
					resource.TestCheckTypeSetElemNestedAttrs("elasticstack_elasticsearch_cluster_settings.test", "persistent.0.setting.*",
						map[string]string{
							"name":  "indices.breaker.total.limit",
							"value": "65%",
						}),
					resource.TestCheckTypeSetElemNestedAttrs("elasticstack_elasticsearch_cluster_settings.test", "transient.0.setting.*",
						map[string]string{
							"name":  "indices.breaker.total.limit",
							"value": "60%",
						}),
				),
			},
			{
				Config: testAccResourceClusterSettingsUpdate(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckTypeSetElemNestedAttrs("elasticstack_elasticsearch_cluster_settings.test", "persistent.0.setting.*",
						map[string]string{
							"name":  "indices.lifecycle.poll_interval",
							"value": "15m",
						}),
					resource.TestCheckTypeSetElemNestedAttrs("elasticstack_elasticsearch_cluster_settings.test", "persistent.0.setting.*",
						map[string]string{
							"name":  "indices.recovery.max_bytes_per_sec",
							"value": "40mb",
						}),
					resource.TestCheckTypeSetElemNestedAttrs("elasticstack_elasticsearch_cluster_settings.test", "persistent.0.setting.*",
						map[string]string{
							"name":  "indices.breaker.total.limit",
							"value": "60%",
						}),
					resource.TestCheckTypeSetElemAttr("elasticstack_elasticsearch_cluster_settings.test", "persistent.0.setting.*.value_list.*", "ACCESS_DENIED"),
					resource.TestCheckTypeSetElemAttr("elasticstack_elasticsearch_cluster_settings.test", "persistent.0.setting.*.value_list.*", "ACCESS_GRANTED"),
					resource.TestCheckNoResourceAttr("elasticstack_elasticsearch_cluster_settings.test", "transient.#"),
				),
			},
		},
	})
}

func testAccResourceClusterSettingsCreate() string {
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
      value = "60%"
    }
  }
}
`
}

func testAccResourceClusterSettingsUpdate() string {
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
	client := acctest.Provider.Meta().(*clients.ApiClient)

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

		req := client.GetESClient().Cluster.GetSettings.WithFlatSettings(true)
		res, err := client.GetESClient().Cluster.GetSettings(req)
		if err != nil {
			return err
		}
		defer res.Body.Close()

		clusterSettings := make(map[string]interface{})
		if err := json.NewDecoder(res.Body).Decode(&clusterSettings); err != nil {
			return err
		}

		if clusterSettings["persistent"] != 0 {
			settings := clusterSettings["persistent"].(map[string]interface{})
			for _, s := range listOfSettings {
				if v, ok := settings[s]; ok {
					return fmt.Errorf(`Setting "%s=%s" still in the cluster, but it should be removed`, s, v)
				}
			}
		}
	}
	return nil
}
