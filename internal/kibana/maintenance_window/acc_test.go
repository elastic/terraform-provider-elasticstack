package maintenance_window_test

import (
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

var minMaintenanceWindowAPISupport = version.Must(version.NewVersion("9.1.0"))

func TestAccResourceMaintenanceWindow(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minMaintenanceWindowAPISupport),
				Config:   testAccResourceMaintenanceWindowCreate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_maintenance_window.test_maintenance_window", "title", "Terraform Maintenance Window"),
					resource.TestCheckResourceAttr("elasticstack_kibana_maintenance_window.test_maintenance_window", "enabled", "true"),
					resource.TestCheckResourceAttr("elasticstack_kibana_maintenance_window.test_maintenance_window", "custom_schedule.start", "1992-01-01T05:00:00.200Z"),
					resource.TestCheckResourceAttr("elasticstack_kibana_maintenance_window.test_maintenance_window", "custom_schedule.duration", "10d"),
					resource.TestCheckResourceAttr("elasticstack_kibana_maintenance_window.test_maintenance_window", "custom_schedule.timezone", "UTC"),
					resource.TestCheckResourceAttr("elasticstack_kibana_maintenance_window.test_maintenance_window", "custom_schedule.recurring.every", "20d"),
					resource.TestCheckResourceAttr("elasticstack_kibana_maintenance_window.test_maintenance_window", "custom_schedule.recurring.end", "2029-05-17T05:05:00.000Z"),
					resource.TestCheckResourceAttr("elasticstack_kibana_maintenance_window.test_maintenance_window", "custom_schedule.recurring.on_week_day.0", "MO"),
					resource.TestCheckResourceAttr("elasticstack_kibana_maintenance_window.test_maintenance_window", "custom_schedule.recurring.on_week_day.1", "TU"),
					resource.TestCheckResourceAttr("elasticstack_kibana_maintenance_window.test_maintenance_window", "scope.alerting.kql", "_id: '1234'"),
				),
			},
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minMaintenanceWindowAPISupport),
				Config:   testAccResourceMaintenanceWindowUpdate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_maintenance_window.test_maintenance_window", "title", "Terraform Maintenance Window UPDATED"),
					resource.TestCheckResourceAttr("elasticstack_kibana_maintenance_window.test_maintenance_window", "enabled", "false"),
					resource.TestCheckResourceAttr("elasticstack_kibana_maintenance_window.test_maintenance_window", "custom_schedule.start", "1999-02-02T05:00:00.200Z"),
					resource.TestCheckResourceAttr("elasticstack_kibana_maintenance_window.test_maintenance_window", "custom_schedule.duration", "12d"),
					resource.TestCheckResourceAttr("elasticstack_kibana_maintenance_window.test_maintenance_window", "custom_schedule.timezone", "Asia/Taipei"),
					resource.TestCheckResourceAttr("elasticstack_kibana_maintenance_window.test_maintenance_window", "custom_schedule.recurring.every", "21d"),
					resource.TestCheckResourceAttr("elasticstack_kibana_maintenance_window.test_maintenance_window", "custom_schedule.recurring.on_month_day.0", "1"),
					resource.TestCheckResourceAttr("elasticstack_kibana_maintenance_window.test_maintenance_window", "custom_schedule.recurring.on_month_day.1", "2"),
					resource.TestCheckResourceAttr("elasticstack_kibana_maintenance_window.test_maintenance_window", "custom_schedule.recurring.on_month_day.2", "3"),
					resource.TestCheckResourceAttr("elasticstack_kibana_maintenance_window.test_maintenance_window", "custom_schedule.recurring.on_month.0", "4"),
					resource.TestCheckResourceAttr("elasticstack_kibana_maintenance_window.test_maintenance_window", "custom_schedule.recurring.on_month.1", "5"),
					resource.TestCheckResourceAttr("elasticstack_kibana_maintenance_window.test_maintenance_window", "scope.alerting.kql", "_id: 'foobar'"),
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

  custom_schedule = {
    start         = "1992-01-01T05:00:00.200Z"
    duration      = "10d"
	timezone      = "UTC"

    recurring = {
      every       = "20d"
      end         = "2029-05-17T05:05:00.000Z"
      on_week_day = ["MO", "TU"]
    }
  }

  scope = {
    alerting = {
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

  custom_schedule = {
    start          = "1999-02-02T05:00:00.200Z"
    duration       = "12d"
	timezone       = "Asia/Taipei"

    recurring = {
      every        = "21d"
	  on_month_day = [1, 2, 3]
	  on_month 	   = [4, 5]
    }
  }

  scope = {
    alerting = {
      kql          = "_id: 'foobar'"
    }
  }
}
`
