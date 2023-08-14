package kibana_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibana"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/hashicorp/go-version"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccResourceSlo(t *testing.T) {
	sloName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)
	getIndicator := func(indicatorType string) string {
		var indicator string

		switch indicatorType {
		case "apm_latency_indicator":
			indicator = `
		apm_latency_indicator {
			  environment     = "production"
			  service         = "my-service"
			  transaction_type = "request"
			  transaction_name = "GET /sup/dawg"
			  index           = "my-index"
			  threshold       = 500
		  }
		  `

		case "apm_availability_indicator":
			indicator = `
		apm_availability_indicator {
			  environment     = "production"
			  service         = "my-service"
			  transaction_type = "request"
			  transaction_name = "GET /sup/dawg"
			  index           = "my-index"
		  }
		  `

		case "kql_custom_indicator":
			indicator = `
		kql_custom_indicator {
			index = "my-index"
			good = "latency < 300"
			total = "*"
			filter = "labels.groupId: group-0"
			timestamp_field = "custom_timestamp"
		  }
		  `

		case "histogram_custom_indicator":
			indicator = `
		histogram_custom_indicator {
			index = "my-index"
			good {
				field = "test"
				aggregation = "value_count"
				filter = "latency < 300"
			}
			total {
				field = "test"
				aggregation = "value_count"
			}
			filter = "labels.groupId: group-0"
			timestamp_field = "custom_timestamp"
		  }
		  `

		case "metric_custom_indicator":
			indicator = `
		metric_custom_indicator {
			index = "my-index"
			good {
				metrics {
						name = "A"
						aggregation = "sum"
						field = "processor.processed"
				}
				equation = "A"
			}

			total {
				metrics {
						name = "A"
						aggregation = "sum"
						field = "processor.accepted"
				}
				equation = "A"
			}
		  }
		  `
		}
		return indicator
	}

	getTFConfig := func(name string, indicatorType string, settingsEnabled bool) string {
		var settings string
		if settingsEnabled {
			settings = `
			settings {
				sync_delay = "5m"
				frequency = "5m"
			}
			`
		} else {
			settings = ""
		}

		config := fmt.Sprintf(`
	provider "elasticstack" {
	  elasticsearch {}
	  kibana {}
	}
	
	resource "elasticstack_elasticsearch_index" "my_index" {
		name = "my-index"
		deletion_protection = false
	}  
	
	resource "elasticstack_kibana_slo" "test_slo" {
		name        = "%s"
		description = "fully sick SLO"
	
	%s
	  
		time_window {
		  duration   = "7d"
		  type = "rolling"
		}
	  
		budgeting_method = "timeslices"
	  
		objective {
		  target          = 0.999
		  timeslice_target = 0.95
		  timeslice_window = "5m"
		}
	  
	%s
	
		group_by = "some.field"
	
		depends_on = [elasticstack_elasticsearch_index.my_index]
	  
	  }
	  
	`, name, getIndicator(indicatorType), settings)
		return config
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		CheckDestroy:             checkResourceSloDestroy,
		ProtoV5ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(version.Must(version.NewSemver("8.9.0"))),
				Config:   getTFConfig(sloName, "apm_latency_indicator", false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "name", sloName),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "description", "fully sick SLO"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "apm_latency_indicator.0.environment", "production"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "apm_latency_indicator.0.service", "my-service"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "apm_latency_indicator.0.transaction_type", "request"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "apm_latency_indicator.0.transaction_name", "GET /sup/dawg"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "apm_latency_indicator.0.index", "my-index"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "apm_latency_indicator.0.threshold", "500"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "time_window.0.duration", "7d"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "time_window.0.type", "rolling"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "budgeting_method", "timeslices"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "objective.0.target", "0.999"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "objective.0.timeslice_target", "0.95"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "objective.0.timeslice_window", "5m"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "settings.0.sync_delay", "1m"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "settings.0.frequency", "1m"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "space_id", "default"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "group_by", "some.field"),
				),
			},
			{ //check that name can be updated
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(version.Must(version.NewSemver("8.9.0"))),
				Config:   getTFConfig(fmt.Sprintf("Updated %s", sloName), "apm_latency_indicator", false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "name", fmt.Sprintf("Updated %s", sloName)),
				),
			},
			{ //check that settings can be updated from api-computed defaults
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(version.Must(version.NewSemver("8.9.0"))),
				Config:   getTFConfig(sloName, "apm_latency_indicator", true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "settings.0.sync_delay", "5m"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "settings.0.frequency", "5m"),
				),
			},
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(version.Must(version.NewSemver("8.9.0"))),
				Config:   getTFConfig(sloName, "apm_availability_indicator", true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "apm_availability_indicator.0.environment", "production"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "apm_availability_indicator.0.service", "my-service"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "apm_availability_indicator.0.transaction_type", "request"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "apm_availability_indicator.0.transaction_name", "GET /sup/dawg"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "apm_availability_indicator.0.index", "my-index"),
				),
			},
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(version.Must(version.NewSemver("8.9.0"))),
				Config:   getTFConfig(sloName, "kql_custom_indicator", true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "kql_custom_indicator.0.index", "my-index"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "kql_custom_indicator.0.good", "latency < 300"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "kql_custom_indicator.0.total", "*"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "kql_custom_indicator.0.filter", "labels.groupId: group-0"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "kql_custom_indicator.0.timestamp_field", "custom_timestamp"),
				),
			},
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(version.Must(version.NewSemver("8.10.0-SNAPSHOT"))),
				Config:   getTFConfig(sloName, "histogram_custom_indicator", true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "histogram_custom_indicator.0.index", "my-index"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "histogram_custom_indicator.0.good.0.field", "test"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "histogram_custom_indicator.0.good.0.aggregation", "value_count"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "histogram_custom_indicator.0.good.0.filter", "latency < 300"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "histogram_custom_indicator.0.total.0.field", "test"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "histogram_custom_indicator.0.total.0.aggregation", "value_count"),
				),
			},
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(version.Must(version.NewSemver("8.10.0-SNAPSHOT"))),
				Config:   getTFConfig(sloName, "metric_custom_indicator", true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "metric_custom_indicator.0.index", "my-index"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "metric_custom_indicator.0.good.0.metrics.0.name", "A"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "metric_custom_indicator.0.good.0.metrics.0.aggregation", "sum"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "metric_custom_indicator.0.good.0.metrics.0.field", "processor.processed"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "metric_custom_indicator.0.total.0.metrics.0.name", "A"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "metric_custom_indicator.0.total.0.metrics.0.aggregation", "sum"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "metric_custom_indicator.0.total.0.metrics.0.field", "processor.accepted"),
				),
			},
		},
	})
}

func checkResourceSloDestroy(s *terraform.State) error {
	client, err := clients.NewAcceptanceTestingClient()
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "elasticstack_kibana_slo" {
			continue
		}
		compId, _ := clients.CompositeIdFromStr(rs.Primary.ID)

		slo, diags := kibana.GetSlo(context.Background(), client, compId.ResourceId, compId.ClusterId)
		if diags.HasError() {
			if len(diags) > 1 || diags[0].Summary != "404 Not Found" {
				return fmt.Errorf("Failed to check if SLO was destroyed: %v", diags)
			}
		}

		if slo != nil {
			return fmt.Errorf("SLO (%s) still exists", compId.ResourceId)
		}
	}
	return nil
}
