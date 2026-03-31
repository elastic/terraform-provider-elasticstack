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
