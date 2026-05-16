variable "policy_name" {
  description = "The integration policy name"
  type        = string
}

provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_fleet_agent_policy" "test" {
  name      = "${var.policy_name}-agent-policy"
  namespace = "default"
}

resource "elasticstack_fleet_elastic_defend_integration_policy" "test" {
  name                = var.policy_name
  namespace           = "default"
  agent_policy_id     = elasticstack_fleet_agent_policy.test.policy_id
  enabled             = true
  integration_version = "8.14.0"
  preset              = "EDRComplete"

  policy = {
    windows = {
      events = {
        process             = true
        network             = true
        file                = true
        dns                 = true
        dll_and_driver_load = true
        registry            = false
        security            = false
        authentication      = false
      }
      malware = {
        mode          = "prevent"
        blocklist     = true
        notify_user   = true
        on_write_scan = true
      }
      ransomware = {
        mode = "prevent"
      }
      memory_protection = {
        mode = "detect"
      }
      behavior_protection = {
        mode               = "prevent"
        reputation_service = true
      }
      popup = {
        malware = {
          message = "Malware detected"
          enabled = true
        }
        ransomware = {
          message = ""
          enabled = false
        }
        memory_protection = {
          message = ""
          enabled = false
        }
        behavior_protection = {
          message = ""
          enabled = false
        }
      }
      antivirus_registration = {
        mode    = "enabled"
        enabled = true
      }
      attack_surface_reduction = {
        credential_hardening = {
          enabled = true
        }
      }
      logging = {
        file = "info"
      }
    }
    mac = {
      events = {
        process = true
        file    = true
      }
      malware = {
        mode          = "prevent"
        blocklist     = true
        on_write_scan = true
        notify_user   = true
      }
      memory_protection = {
        mode = "prevent"
      }
      behavior_protection = {
        mode               = "detect"
        reputation_service = true
      }
      popup = {
        malware = {
          message = "Mac malware alert"
          enabled = true
        }
        memory_protection = {
          message = ""
          enabled = false
        }
        behavior_protection = {
          message = ""
          enabled = false
        }
      }
      logging = {
        file = "warning"
      }
    }
    linux = {
      events = {
        process      = true
        network      = true
        file         = true
        session_data = true
        tty_io       = false
      }
      malware = {
        mode      = "detect"
        blocklist = true
      }
      memory_protection = {
        mode = "prevent"
      }
      behavior_protection = {
        mode               = "detect"
        reputation_service = true
      }
      popup = {
        malware = {
          message = "Linux malware alert"
          enabled = true
        }
        memory_protection = {
          message = ""
          enabled = false
        }
        behavior_protection = {
          message = ""
          enabled = false
        }
      }
      logging = {
        file = "warning"
      }
    }
  }
}
