package security_test

import (
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceSecurityRoleMapping(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceSecurityRoleMapping,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_security_role_mapping.test", "name", "data_source_test"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_security_role_mapping.test", "enabled", "true"),
					utils.TestCheckResourceListAttr("data.elasticstack_elasticsearch_security_role_mapping.test", "roles", []string{"admin"}),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_security_role_mapping.test", "rules", `{"any":[{"field":{"username":"esadmin"}},{"field":{"groups":"cn=admins,dc=example,dc=com"}}]}`),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_security_role_mapping.test", "metadata", `{"version":1}`),
				),
			},
		},
	})
}

const testAccDataSourceSecurityRoleMapping = `
provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_security_role_mapping" "test" {
  name = "data_source_test"
  enabled = true
  roles = [
    "admin"
  ]
  rules = jsonencode({
    any = [
      { field = { username = "esadmin" } },
      { field = { groups = "cn=admins,dc=example,dc=com" } },
    ]
  })

  metadata = jsonencode({ version = 1 })
}

data "elasticstack_elasticsearch_security_role_mapping" "test" {
  name = elasticstack_elasticsearch_security_role_mapping.test.name
}
`
