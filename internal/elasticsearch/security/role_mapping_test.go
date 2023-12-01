package security_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestResourceRoleMapping(t *testing.T) {
	roleMappingName := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		CheckDestroy:             checkResourceSecurityRoleMappingDestroy,
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceSecurityRoleMappingCreate(roleMappingName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_security_role_mapping.test", "name", roleMappingName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_security_role_mapping.test", "enabled", "true"),
					utils.TestCheckResourceListAttr("elasticstack_elasticsearch_security_role_mapping.test", "roles", []string{"admin"}),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_security_role_mapping.test", "rules", `{"any":[{"field":{"username":"esadmin"}},{"field":{"groups":"cn=admins,dc=example,dc=com"}}]}`),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_security_role_mapping.test", "metadata", `{"version":1}`),
				),
			},
			{
				Config: testAccResourceSecurityRoleMappingUpdate(roleMappingName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_security_role_mapping.test", "name", roleMappingName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_security_role_mapping.test", "enabled", "false"),
					utils.TestCheckResourceListAttr("elasticstack_elasticsearch_security_role_mapping.test", "roles", []string{"admin", "user"}),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_security_role_mapping.test", "rules", `{"any":[{"field":{"username":"esadmin"}},{"field":{"groups":"cn=admins,dc=example,dc=com"}}]}`),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_security_role_mapping.test", "metadata", `{}`),
				),
			},
			{
				Config: testAccResourceSecurityRoleMappingRoleTemplates(roleMappingName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_security_role_mapping.test", "name", roleMappingName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_security_role_mapping.test", "enabled", "false"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_security_role_mapping.test", "role_templates", `[{"format":"json","template":"{\"source\":\"{{#tojson}}groups{{/tojson}}\"}"}]`),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_security_role_mapping.test", "rules", `{"any":[{"field":{"username":"esadmin"}},{"field":{"groups":"cn=admins,dc=example,dc=com"}}]}`),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_security_role_mapping.test", "metadata", `{}`),
				),
			},
		},
	})
}

func testAccResourceSecurityRoleMappingCreate(roleMappingName string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_security_role_mapping" "test" {
  name = "%s"
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
	`, roleMappingName)
}

func testAccResourceSecurityRoleMappingUpdate(roleMappingName string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_security_role_mapping" "test" {
  name = "%s"
  enabled = false
  roles = [
    "admin",
    "user"
  ]
  rules = jsonencode({
    any = [
      { field = { username = "esadmin" } },
      { field = { groups = "cn=admins,dc=example,dc=com" } },
    ]
  })
}
	`, roleMappingName)
}

func testAccResourceSecurityRoleMappingRoleTemplates(roleMappingName string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_security_role_mapping" "test" {
  name = "%s"
  enabled = false
  role_templates = jsonencode([
    {
      template = jsonencode({ source = "{{#tojson}}groups{{/tojson}}" }),
      format = "json"
    }
  ])
  rules = jsonencode({
    any = [
      { field = { username = "esadmin" } },
      { field = { groups = "cn=admins,dc=example,dc=com" } },
    ]
  })
}
	`, roleMappingName)
}

func checkResourceSecurityRoleMappingDestroy(s *terraform.State) error {
	client, err := clients.NewAcceptanceTestingClient()
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "elasticstack_elasticsearch_security_role_mapping" {
			continue
		}
		compId, _ := clients.CompositeIdFromStr(rs.Primary.ID)

		esClient, err := client.GetESClient()
		if err != nil {
			return err
		}
		req := esClient.Security.GetRoleMapping.WithName(compId.ResourceId)
		res, err := esClient.Security.GetRoleMapping(req)
		if err != nil {
			return err
		}

		if res.StatusCode != http.StatusNotFound {
			return fmt.Errorf("role mapping (%s) still exists", compId.ResourceId)
		}
	}
	return nil
}
