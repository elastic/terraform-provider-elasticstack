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
  name            = var.policy_name
  namespace       = "default"
  agent_policy_id = elasticstack_fleet_agent_policy.test.policy_id
  # description intentionally omitted so it is left unmanaged in this step
  # enabled omitted to verify it reverts to the default value of true
  force               = true
  integration_version = "8.14.0"
  preset              = "EDRComplete"

  policy = {
    windows = {
      events = {
        process             = true
        network             = true
        file                = false
        dns                 = false
        dll_and_driver_load = false
        registry            = true
        security            = true
        authentication      = true
        credential_access   = false
      }
      malware = {
        mode          = "detect"
        blocklist     = false
        notify_user   = false
        on_write_scan = false
      }
      ransomware = {
        mode = "detect"
      }
      memory_protection = {
        mode                   = "prevent"
        custom_yara_signatures = false
      }
      behavior_protection = {
        mode               = "detect"
        reputation_service = false
      }
      device_control = {
        enabled     = false
        usb_storage = "audit"
      }
      popup = {
        malware = {
          message = ""
          enabled = false
        }
        ransomware = {
          message = "Ransomware blocked"
          enabled = true
        }
        memory_protection = {
          message = "Memory alert"
          enabled = true
        }
        behavior_protection = {
          message = ""
          enabled = false
        }
        device_control = {
          message = ""
          enabled = false
        }
      }
      antivirus_registration = {
        mode    = "sync_with_malware_prevent"
        enabled = false
      }
      attack_surface_reduction = {
        credential_hardening = {
          enabled = false
        }
      }
      logging = {
        file = "error"
      }
    }
    mac = {
      events = {
        process  = true
        network  = false
        file     = false
        dns      = false
        security = true
      }
      malware = {
        mode = "detect"
      }
      ransomware = {
        mode = "detect"
      }
      memory_protection = {
        mode                   = "detect"
        custom_yara_signatures = true
      }
      behavior_protection = {
        mode               = "prevent"
        reputation_service = false
      }
      device_control = {
        enabled     = false
        usb_storage = "deny_all"
      }
      popup = {
        malware = {
          message = ""
          enabled = false
        }
        memory_protection = {
          message = "Mac memory alert"
          enabled = true
        }
        behavior_protection = {
          message = ""
          enabled = false
        }
        ransomware = {
          message = ""
          enabled = false
        }
        device_control = {
          message = "Mac device blocked"
          enabled = true
        }
      }
      logging = {
        file = "error"
      }
    }
    linux = {
      events = {
        process      = true
        network      = false
        file         = false
        session_data = false
        tty_io       = true
        dns          = false
      }
      malware = {
        mode          = "prevent"
        blocklist     = false
        on_write_scan = false
      }
      memory_protection = {
        mode                   = "detect"
        custom_yara_signatures = false
      }
      behavior_protection = {
        mode               = "prevent"
        reputation_service = false
      }
      popup = {
        malware = {
          message = ""
          enabled = false
        }
        memory_protection = {
          message = "Linux memory alert"
          enabled = true
        }
        behavior_protection = {
          message = ""
          enabled = false
        }
      }
      logging = {
        file = "error"
      }
    }
  }
}
