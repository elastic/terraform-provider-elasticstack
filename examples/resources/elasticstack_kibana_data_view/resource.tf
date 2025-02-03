provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_kibana_data_view" "my_data_view" {
  data_view = {
    name            = "logs-*"
    title           = "logs-*"
    time_field_name = "@timestamp"
    namespaces      = ["backend"]
  }
}

resource "elasticstack_kibana_data_view" "custom_fields_data_view" {
  data_view = {
    name            = "custom-data-view"
    id              = "custom-data-view"
    title           = "logs-*"
    time_field_name = "@timestamp"
    namespaces      = ["default"]
    field_formats = {
      "host.uptime" = {
        id = "duration"
        params = {
          input_format              = "hours"
          output_format             = "humanizePrecise"
          output_precision          = 2
          include_space_with_suffix = true
          use_short_suffix          = true
        }
      }
      "user.last_password_change" = {
        id     = "relative_date"
        params = {}
      }
      "user.last_login" = {
        id = "date"
        params = {
          pattern  = "MMM D, YYYY @ HH:mm:ss.SSS"
          timezone = "America/New_York"
        }
      }
      "user.is_active" = {
        id     = "boolean"
        params = {}
      }
      "user.status" = {
        id = "color"
        params = {
          field_type = "string"
          colors = [
            {
              range      = "-Infinity:Infinity"
              regex      = "inactive*"
              text       = "#000000"
              background = "#ffffff"
            }
          ]
        }
      }
      "user.message" = {
        id = "truncate"
        params = {
          field_length = 10
        }
      }
      "host.name" = {
        id = "string"
        params = {
          transform = "upper"
        }
      }
      "response.code" = {
        id = "static_lookup"
        params = {
          lookup_entries = [
            {
              key   = "200"
              value = "OK"
            }
            {
              key   = "404"
              value = "Not Found"
            }
          ]
          unknown_key_value = "Unknown"
        }
      }
      "url.original" = {
        id = "url"
        params = {
          type          = "a"
          urltemplate   = "https://test.com/{{value}}"
          labeltemplate = "{{value}}"
        }
      }
      "user.profile_picture" = {
        id = "url"
        params = {
          type          = "img"
          urltemplate   = "https://test.com/{{value}}"
          labeltemplate = "{{value}}"
          width         = 6
          height        = 4
        }
      }
      "user.answering_message" = {
        id = "url"
        params = {
          type           = "audio"
          urltemplate    = "https://test.com/{{value}}"
          labeltemplate  = "{{value}}"
        }
      }
    }
    field_attrs = {
      "response.code" = {
        custom_label = "Response Code"
        count        = 0
      }
    }
  }
}
