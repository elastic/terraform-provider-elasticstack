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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func CheckMinVersionAndServerless() (bool, error) {
	minSupportedVersion := version.Must(version.NewSemver("9.1.0"))
	versionIsUnsupported, err := versionutils.CheckIfVersionIsUnsupported(minSupportedVersion)()
	if err != nil {
		return false, err
	}

	isServerless, err := versionutils.CheckIfServerless()()
	if err != nil {
		return false, err
	}

	return versionIsUnsupported && !isServerless, err
}

func TestAccResourceMaintenanceWindow(t *testing.T) {
	t.Setenv("KIBANA_API_KEY", "")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		CheckDestroy:             checkResourceMaintenanceWindowDestroy,
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				SkipFunc: CheckMinVersionAndServerless,
				Config:   testAccResourceMaintenanceWindowCreate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_maintenance_window.test_maintenance_window", "title", "Terraform Maintenance Window"),
					resource.TestCheckResourceAttr("elasticstack_kibana_maintenance_window.test_maintenance_window", "enabled", "true"),
					resource.TestCheckResourceAttr("elasticstack_kibana_maintenance_window.test_maintenance_window", "custom_schedule.0.start", "1992-01-01T05:00:00.200Z"),
					resource.TestCheckResourceAttr("elasticstack_kibana_maintenance_window.test_maintenance_window", "custom_schedule.0.duration", "10d"),
					resource.TestCheckResourceAttr("elasticstack_kibana_maintenance_window.test_maintenance_window", "custom_schedule.0.timezone", "UTC"),
					resource.TestCheckResourceAttr("elasticstack_kibana_maintenance_window.test_maintenance_window", "custom_schedule.0.recurring.0.every", "20d"),
					resource.TestCheckResourceAttr("elasticstack_kibana_maintenance_window.test_maintenance_window", "custom_schedule.0.recurring.0.end", "2029-05-17T05:05:00.000Z"),
					resource.TestCheckResourceAttr("elasticstack_kibana_maintenance_window.test_maintenance_window", "custom_schedule.0.recurring.0.on_week_day.0", "MO"),
					resource.TestCheckResourceAttr("elasticstack_kibana_maintenance_window.test_maintenance_window", "custom_schedule.0.recurring.0.on_week_day.1", "TU"),
					resource.TestCheckResourceAttr("elasticstack_kibana_maintenance_window.test_maintenance_window", "scope.0.alerting.0.kql", "_id: '1234'"),
				),
			},
			{
				SkipFunc: CheckMinVersionAndServerless,
				Config:   testAccResourceMaintenanceWindowUpdate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_maintenance_window.test_maintenance_window", "title", "Terraform Maintenance Window UPDATED"),
					resource.TestCheckResourceAttr("elasticstack_kibana_maintenance_window.test_maintenance_window", "enabled", "false"),
					resource.TestCheckResourceAttr("elasticstack_kibana_maintenance_window.test_maintenance_window", "custom_schedule.0.start", "1999-02-02T05:00:00.200Z"),
					resource.TestCheckResourceAttr("elasticstack_kibana_maintenance_window.test_maintenance_window", "custom_schedule.0.duration", "12d"),
					resource.TestCheckResourceAttr("elasticstack_kibana_maintenance_window.test_maintenance_window", "custom_schedule.0.timezone", "Asia/Taipei"),
					resource.TestCheckResourceAttr("elasticstack_kibana_maintenance_window.test_maintenance_window", "custom_schedule.0.recurring.0.every", "21d"),
					resource.TestCheckResourceAttr("elasticstack_kibana_maintenance_window.test_maintenance_window", "custom_schedule.0.recurring.0.end", ""),
					resource.TestCheckResourceAttr("elasticstack_kibana_maintenance_window.test_maintenance_window", "custom_schedule.0.recurring.0.on_month_day.0", "1"),
					resource.TestCheckResourceAttr("elasticstack_kibana_maintenance_window.test_maintenance_window", "custom_schedule.0.recurring.0.on_month_day.1", "2"),
					resource.TestCheckResourceAttr("elasticstack_kibana_maintenance_window.test_maintenance_window", "custom_schedule.0.recurring.0.on_month_day.2", "3"),
					resource.TestCheckResourceAttr("elasticstack_kibana_maintenance_window.test_maintenance_window", "custom_schedule.0.recurring.0.on_month.0", "4"),
					resource.TestCheckResourceAttr("elasticstack_kibana_maintenance_window.test_maintenance_window", "custom_schedule.0.recurring.0.on_month.1", "5"),
					resource.TestCheckResourceAttr("elasticstack_kibana_maintenance_window.test_maintenance_window", "scope.0.alerting.0.kql", "_id: 'foobar'"),
				),
			},
		},
	})
}

const testAccResourceMaintenanceWindowCreate = `
provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_kibana_maintenance_window" "test_maintenance_window" {
  title   	      = "Terraform Maintenance Window"
  enabled 	      = true
  custom_schedule {
    start         = "1992-01-01T05:00:00.200Z"
    duration      = "10d"
	timezone      = "UTC"

    recurring {
      every       = "20d"
      end         = "2029-05-17T05:05:00.000Z"
      on_week_day = ["MO", "TU"]
    }
  }

  scope {
    alerting {
      kql         = "_id: '1234'"
    }
  }
}
`

const testAccResourceMaintenanceWindowUpdate = `
provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_kibana_maintenance_window" "test_maintenance_window" {
  title   		   = "Terraform Maintenance Window UPDATED"
  enabled 		   = false
  custom_schedule {
    start          = "1999-02-02T05:00:00.200Z"
    duration       = "12d"
	timezone       = "Asia/Taipei"

    recurring {
      every        = "21d"
	  on_month_day = [1, 2, 3]
	  on_month 	   = [4, 5]
    }
  }

  scope {
    alerting {
      kql          = "_id: 'foobar'"
    }
  }
}
`

func checkResourceMaintenanceWindowDestroy(s *terraform.State) error {
	client, err := clients.NewAcceptanceTestingClient()
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "elasticstack_kibana_maintenance_window" {
			continue
		}
		compId, _ := clients.CompositeIdFromStr(rs.Primary.ID)

		maintenanceWindow, diags := kibana.GetMaintenanceWindow(context.Background(), client, compId.ResourceId, compId.ClusterId)

		if diags.HasError() {
			return fmt.Errorf("Failed to get maintenance window: %v", diags)
		}

		if maintenanceWindow != nil {
			return fmt.Errorf("Maintenance window (%s) still exists", compId.ResourceId)
		}
	}

	return nil
}
