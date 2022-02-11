provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_security_role_mapping" "role_mapping" {
  name    = "testrolemapping"
  enabled = true
  roles = [
    "user",
  ]

  # role_templates can be defined in different ways
  role_templates = [
    # using the jsonencode function, which is the recommended way if you want to provide JSON object yourself
    jsonencode({
      template = {
        source = "saml_user"
      }
    }),

    // or using HEREDOC construct to provide the role_template definition
    <<-EOT
    {
      "template": {
        "source": "_user_{{username}}"
      }
    }
    EOT
  ]

  metadata = jsonencode({
    version = 1
  })
}

output "role_mapping" {
  value = elasticstack_elasticsearch_security_role_mapping.name
}
