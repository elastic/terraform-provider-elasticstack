terraform {
  required_providers {
    elasticstack = {
      source = "elastic/elasticstack"
    }
  }
}

provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_dashboard" "getting_started" {
  title       = "Getting started: Web server logs"
  description = "A step-by-step web server log monitoring dashboard built with Terraform using Kibana sample logs data."

  time_range = {
    from = "now-7d"
    to   = "now"
  }

  refresh_interval = {
    pause = true
    value = 0
  }

  query = {
    language = "kql"
    text     = ""
  }

  panels = [
    {
      type = "markdown"
      grid = { x = 0, y = 0, w = 48, h = 6 }
      markdown_config = {
        by_value = {
          title   = "About this dashboard"
          content = <<-EOT
            # Welcome

            This dashboard monitors **web server logs** from the Kibana sample logs dataset.
            Explore request volume, top URLs, and HTTP response codes over the last 7 days.

            Learn more in the [Kibana dashboards guide](https://www.elastic.co/docs/explore-analyze/dashboards).
          EOT
          settings = {
            open_links_in_new_tab = true
          }
        }
      }
    },
  ]
}
