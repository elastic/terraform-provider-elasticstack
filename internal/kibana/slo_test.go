package kibana_test

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibana"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/hashicorp/go-version"
	"github.com/stretchr/testify/require"

	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccResourceSlo(t *testing.T) {
	// This test exposes a bug in Kibana present in 8.11.x
	slo8_9Constraints, err := version.NewConstraint(">=8.9.0,!=8.11.0,!=8.11.1,!=8.11.2,!=8.11.3,!=8.11.4")
	require.NoError(t, err)

	slo8_10Constraints, err := version.NewConstraint(">=8.10.0,!=8.11.0,!=8.11.1,!=8.11.2,!=8.11.3,!=8.11.4")
	require.NoError(t, err)

	sloName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		CheckDestroy:             checkResourceSloDestroy,
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				SkipFunc: versionutils.CheckIfVersionMeetsConstraints(slo8_9Constraints),
				Config:   getSLOConfig(sloName, "apm_latency_indicator", false, []string{}, ""),
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
				),
			},
			{ //check that name can be updated
				SkipFunc: versionutils.CheckIfVersionMeetsConstraints(slo8_9Constraints),
				Config:   getSLOConfig(fmt.Sprintf("Updated %s", sloName), "apm_latency_indicator", false, []string{}, ""),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "name", fmt.Sprintf("Updated %s", sloName)),
				),
			},
			{ //check that settings can be updated from api-computed defaults
				SkipFunc: versionutils.CheckIfVersionMeetsConstraints(slo8_9Constraints),
				Config:   getSLOConfig(sloName, "apm_latency_indicator", true, []string{}, ""),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "settings.0.sync_delay", "5m"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "settings.0.frequency", "5m"),
				),
			},
			{
				SkipFunc: versionutils.CheckIfVersionMeetsConstraints(slo8_9Constraints),
				Config:   getSLOConfig(sloName, "apm_availability_indicator", true, []string{}, ""),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "apm_availability_indicator.0.environment", "production"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "apm_availability_indicator.0.service", "my-service"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "apm_availability_indicator.0.transaction_type", "request"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "apm_availability_indicator.0.transaction_name", "GET /sup/dawg"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "apm_availability_indicator.0.index", "my-index"),
				),
			},
			{
				SkipFunc: versionutils.CheckIfVersionMeetsConstraints(slo8_9Constraints),
				Config:   getSLOConfig(sloName, "kql_custom_indicator", true, []string{}, ""),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "kql_custom_indicator.0.index", "my-index"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "kql_custom_indicator.0.good", "latency < 300"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "kql_custom_indicator.0.total", "*"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "kql_custom_indicator.0.filter", "labels.groupId: group-0"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "kql_custom_indicator.0.timestamp_field", "custom_timestamp"),
				),
			},
			{
				SkipFunc: versionutils.CheckIfVersionMeetsConstraints(slo8_10Constraints),
				Config:   getSLOConfig(sloName, "histogram_custom_indicator", true, []string{}, ""),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "histogram_custom_indicator.0.index", "my-index"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "histogram_custom_indicator.0.good.0.field", "test"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "histogram_custom_indicator.0.good.0.aggregation", "value_count"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "histogram_custom_indicator.0.good.0.filter", "latency < 300"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_slo.test_slo", "histogram_custom_indicator.0.good.0.from"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_slo.test_slo", "histogram_custom_indicator.0.good.0.to"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "histogram_custom_indicator.0.total.0.field", "test"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "histogram_custom_indicator.0.total.0.aggregation", "value_count"),
				),
			},
			{
				SkipFunc: versionutils.CheckIfVersionMeetsConstraints(slo8_10Constraints),
				Config:   getSLOConfig(sloName, "metric_custom_indicator", true, []string{}, "some.field"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "metric_custom_indicator.0.index", "my-index"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "metric_custom_indicator.0.good.0.metrics.0.name", "A"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "metric_custom_indicator.0.good.0.metrics.0.aggregation", "sum"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "metric_custom_indicator.0.good.0.metrics.0.field", "processor.processed"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "metric_custom_indicator.0.good.0.metrics.1.name", "B"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "metric_custom_indicator.0.good.0.metrics.1.aggregation", "sum"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "metric_custom_indicator.0.good.0.metrics.1.field", "processor.processed"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "metric_custom_indicator.0.good.0.equation", "A + B"),

					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "metric_custom_indicator.0.total.0.metrics.0.name", "A"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "metric_custom_indicator.0.total.0.metrics.0.aggregation", "sum"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "metric_custom_indicator.0.total.0.metrics.0.field", "processor.accepted"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "metric_custom_indicator.0.total.0.metrics.1.name", "B"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "metric_custom_indicator.0.total.0.metrics.1.aggregation", "sum"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "metric_custom_indicator.0.total.0.metrics.1.field", "processor.accepted"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "metric_custom_indicator.0.total.0.equation", "A + B"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "group_by", "some.field"),
				),
			},
			{
				SkipFunc: versionutils.CheckIfVersionMeetsConstraints(slo8_10Constraints),
				Config:   getSLOConfig(sloName, "metric_custom_indicator", true, []string{"tag-1", "another_tag"}, ""),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "tags.0", "tag-1"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "tags.1", "another_tag"),
				),
			},
		},
	})
}

func TestAccResourceSloErrors(t *testing.T) {
	multipleIndicatorsConfig := `
	provider "elasticstack" {
		elasticsearch {}
		kibana {}
	  }

	  resource "elasticstack_elasticsearch_index" "my_index" {
		  name = "my-index"
		  deletion_protection = false
	  }

	  resource "elasticstack_kibana_slo" "test_slo" {
		  name        = "fail"
		  description = "multiple indicator fail"

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

		kql_custom_indicator {
			index = "my-index"
			good = "latency < 300"
			total = "*"
			filter = "labels.groupId: group-0"
			timestamp_field = "custom_timestamp"
		}

		  time_window {
			duration   = "7d"
			type = "rolling"
		  }

		  budgeting_method = "supdawg"

		  objective {
			target          = 0.999
			timeslice_target = 0.95
			timeslice_window = "5m"
		  }

		  depends_on = [elasticstack_elasticsearch_index.my_index]

	}`

	budgetingMethodFailConfig := getSLOConfig("budgetingMethodFail", "apm_latency_indicator", false, []string{}, "")
	budgetingMethodFailConfig = strings.Replace(budgetingMethodFailConfig, "budgeting_method = \"timeslices\"", "budgeting_method = \"supdawg\"", -1)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				SkipFunc:    versionutils.CheckIfVersionIsUnsupported(version.Must(version.NewSemver("8.9.0"))),
				Config:      multipleIndicatorsConfig,
				ExpectError: regexp.MustCompile("Invalid combination of arguments"),
			},
			{
				SkipFunc:    versionutils.CheckIfVersionIsUnsupported(version.Must(version.NewSemver("8.10.0-SNAPSHOT"))),
				Config:      getSLOConfig("failwhale", "histogram_custom_indicator_agg_fail", false, []string{}, ""),
				ExpectError: regexp.MustCompile(`expected histogram_custom_indicator.0.good.0.aggregation to be one of \["?value_count"? "?range"?\], got supdawg`),
			},
			{
				SkipFunc:    versionutils.CheckIfVersionIsUnsupported(version.Must(version.NewSemver("8.9.0"))),
				Config:      budgetingMethodFailConfig,
				ExpectError: regexp.MustCompile(`expected budgeting_method to be one of \["occurrences" "timeslices"\], got supdawg`),
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

func getSLOConfig(name string, indicatorType string, settingsEnabled bool, tags []string, group_by string) string {
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

	var tagsOption string
	if len(tags) != 0 {
		tagsJson, _ := json.Marshal(tags)
		tagsOption = "tags = " + string(tagsJson)
	} else {
		tagsOption = ""
	}

	var groupByOption string
	if len(group_by) != 0 {
		groupByOption = "group_by = \"" + group_by + "\""
	} else {
		groupByOption = ""
	}

	configTemplate := `
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

			%s

			depends_on = [elasticstack_elasticsearch_index.my_index]

			%s
		}
	`

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

		case "histogram_custom_indicator_agg_fail":
			indicator = `
		histogram_custom_indicator {
			index = "my-index"
			good {
				field = "test"
				aggregation = "supdawg"
				filter = "latency < 300"
				from = 0
				to = 10
			}
			total {
				field = "test"
				aggregation = "supdawg"
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
				metrics {
						name = "B"
						aggregation = "sum"
						field = "processor.processed"
				}
				equation = "A + B"
			}

			total {
				metrics {
						name = "A"
						aggregation = "sum"
						field = "processor.accepted"
				}
				metrics {
						name = "B"
						aggregation = "sum"
						field = "processor.accepted"
				}
				equation = "A + B"
			}
		}
		  `
		}
		return indicator
	}

	config := fmt.Sprintf(configTemplate, name, getIndicator(indicatorType), settings, groupByOption, tagsOption)
	return config
}
