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
        id = "duration",
        params = {
          inputFormat            = "hours",
          outputFormat           = "humanizePrecise",
          outputPrecision        = 2,
          includeSpaceWithSuffix = true,
          useShortSuffix         = true,
        }
      }
      "user.last_password_change" = {
        id     = "relative_date",
        params = {}
      },
      "user.last_login" = {
        id = "date",
        params = {
          pattern  = "MMM D, YYYY @ HH:mm:ss.SSS",
          timezone = "America/New_York"
        }
      },
      "user.is_active" = {
        id     = "boolean",
        params = {}
      },
      "user.status" = {
        id = "color",
        params = {
          fieldType = "string",
          colors = [
            {
              range      = "-Infinity:Infinity",
              regex      = "inactive*",
              text       = "#000000",
              background = "#ffffff"
            }
          ]
        }
      },
      "user.message" = {
        id = "truncate",
        params = {
          fieldLength = 10
        }
      },
      "host.name" = {
        id = "string",
        params = {
          transform = "upper"
        }
      },
      "response.code" = {
        id = "static_lookup",
        params = {
          lookupEntries = [
            {
              key   = "200",
              value = "OK"
            },
            {
              key   = "404",
              value = "Not Found"
            }
          ],
          unknownKeyValue = "Unknown"
        }
      },
      "url.original" = {
        id = "url",
        params = {
          type          = "a",
          urlTemplate   = "URL TEMPLATE",
          labelTemplate = "LABEL TEMPLATE",
        }
      },
      "user.profile_picture" = {
        id = "url",
        params = {
          type          = "img",
          urlTemplate   = "URL TEMPLATE",
          labelTemplate = "LABEL TEMPLATE",
          width         = 6,
          height        = 4
        }
      },
      "user.answering_message" = {
        id = "url",
        params = {
          type          = "audio",
          urlTemplate   = "URL TEMPLATE",
          labelTemplate = "LABEL TEMPLATE"
        }
      }
    },
    fieldAttrs = {
      "response.code" = {
        customLabel       = "Response Code",
        customDescription = "The response code from the server",
        count             = 0
      }
    }
  }
}
