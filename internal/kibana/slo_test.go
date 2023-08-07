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

var minSupportedVersion = version.Must(version.NewSemver("8.9.0"))

func TestAccResourceSlo(t *testing.T) {
	sloName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		CheckDestroy:             checkResourceSloDestroy,
		ProtoV5ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minSupportedVersion),
				Config:   getTFConfig(sloName, "sli.apm.transactionDuration", true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "name", sloName),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "description", "fully sick SLO"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "indicator.0.type", "sli.apm.transactionDuration"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "indicator.0.params.0.environment", "production"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "indicator.0.params.0.service", "my-service"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "indicator.0.params.0.transaction_type", "request"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "indicator.0.params.0.transaction_name", "GET /sup/dawg"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "indicator.0.params.0.index", "my-index"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "indicator.0.params.0.threshold", "500"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "time_window.0.duration", "7d"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "time_window.0.type", "rolling"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "budgeting_method", "timeslices"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "objective.0.target", "0.999"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "objective.0.timeslice_target", "0.95"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "objective.0.timeslice_window", "5m"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "settings.0.sync_delay", "5m"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "settings.0.frequency", "1m"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "space_id", "default"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "group_by", "some.field"),
				),
			},
			{ //check that name can be updated
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minSupportedVersion),
				Config:   getTFConfig(fmt.Sprintf("Updated %s", sloName), "sli.apm.transactionDuration", true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "name", fmt.Sprintf("Updated %s", sloName)),
				),
			},
			{ //check that settings get reset to defauts when omitted from tf definition
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minSupportedVersion),
				Config:   getTFConfig(sloName, "sli.apm.transactionDuration", false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "settings.0.sync_delay", "1m"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "settings.0.frequency", "1m"),
				),
			},
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minSupportedVersion),
				Config:   getTFConfig(sloName, "sli.apm.transactionErrorRate", true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "indicator.0.type", "sli.apm.transactionErrorRate"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "indicator.0.params.0.environment", "production"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "indicator.0.params.0.service", "my-service"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "indicator.0.params.0.transaction_type", "request"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "indicator.0.params.0.transaction_name", "GET /sup/dawg"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "indicator.0.params.0.index", "my-index"),
				),
			},
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minSupportedVersion),
				Config:   getTFConfig(sloName, "sli.kql.custom", true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "indicator.0.type", "sli.kql.custom"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "indicator.0.params.0.index", "my-index"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "indicator.0.params.0.good", "latency < 300"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "indicator.0.params.0.total", "*"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "indicator.0.params.0.filter", "labels.groupId: group-0"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "indicator.0.params.0.timestamp_field", "custom_timestamp"),
				),
			},
		},
	})
}

func getTFConfig(name string, indicatorType string, settingsEnabled bool) string {
	var settings string
	if settingsEnabled {
		settings = `
		settings {
			sync_delay = "5m"
			frequency = "1m"
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

	fmt.Println("applying config: ", config)

	return config
}

func getIndicator(indicatorType string) string {
	var indicator string

	switch indicatorType {
	case "sli.apm.transactionDuration":
		indicator = `
	indicator {
		type = "sli.apm.transactionDuration"
		params {
		  environment     = "production"
		  service         = "my-service"
		  transaction_type = "request"
		  transaction_name = "GET /sup/dawg"
		  index           = "my-index"
		  threshold       = 500
		}
	  }
	  `

	case "sli.apm.transactionErrorRate":
		indicator = `
	indicator {
		type = "sli.apm.transactionErrorRate"
		params {
		  environment     = "production"
		  service         = "my-service"
		  transaction_type = "request"
		  transaction_name = "GET /sup/dawg"
		  index           = "my-index"
		}
	  }
	  `

	case "sli.kql.custom":
		indicator = `
	indicator {
		type = "sli.kql.custom"
		params {
		  index = "my-index"
		  good = "latency < 300"
		  total = "*"
		  filter = "labels.groupId: group-0"
		  timestamp_field = "custom_timestamp"
		}
	  }
	  `
	}

	return indicator
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
		fmt.Printf("Checking for SLO (%s)\n", compId.ResourceId)

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
