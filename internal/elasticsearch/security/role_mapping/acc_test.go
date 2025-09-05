package role_mapping_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/acctest/checks"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccResourceSecurityRoleMapping(t *testing.T) {
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
					checks.TestCheckResourceListAttr("elasticstack_elasticsearch_security_role_mapping.test", "roles", []string{"admin"}),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_security_role_mapping.test", "rules", `{"any":[{"field":{"username":"esadmin"}},{"field":{"groups":"cn=admins,dc=example,dc=com"}}]}`),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_security_role_mapping.test", "metadata", `{"version":1}`),
				),
			},
			{
				Config: testAccResourceSecurityRoleMappingUpdate(roleMappingName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_security_role_mapping.test", "name", roleMappingName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_security_role_mapping.test", "enabled", "false"),
					checks.TestCheckResourceListAttr("elasticstack_elasticsearch_security_role_mapping.test", "roles", []string{"admin", "user"}),
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

func TestAccResourceSecurityRoleMappingFromSDK(t *testing.T) {
	roleMappingName := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				// Create the role mapping with the last provider version where the role mapping resource was built on the SDK
				ExternalProviders: map[string]resource.ExternalProvider{
					"elasticstack": {
						Source:            "elastic/elasticstack",
						VersionConstraint: "0.11.17",
					},
				},
				Config: testAccResourceSecurityRoleMappingCreate(roleMappingName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_security_role_mapping.test", "name", roleMappingName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_security_role_mapping.test", "enabled", "true"),
					checks.TestCheckResourceListAttr("elasticstack_elasticsearch_security_role_mapping.test", "roles", []string{"admin"}),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				Config:                   testAccResourceSecurityRoleMappingCreate(roleMappingName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_security_role_mapping.test", "name", roleMappingName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_security_role_mapping.test", "enabled", "true"),
					checks.TestCheckResourceListAttr("elasticstack_elasticsearch_security_role_mapping.test", "roles", []string{"admin"}),
				),
			},
		},
	})
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
			return fmt.Errorf("Role mapping (%s) still exists", compId.ResourceId)
		}
	}
	return nil
}

func testAccResourceSecurityRoleMappingCreate(roleMappingName string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_security_role_mapping" "test" {
  name    = "%s"
  enabled = true
  roles   = ["admin"]
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
  name    = "%s"
  enabled = false
  roles   = ["admin", "user"]
  rules = jsonencode({
    any = [
      { field = { username = "esadmin" } },
      { field = { groups = "cn=admins,dc=example,dc=com" } },
    ]
  })

  metadata = jsonencode({})
}
	`, roleMappingName)
}

func testAccResourceSecurityRoleMappingRoleTemplates(roleMappingName string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_security_role_mapping" "test" {
  name    = "%s"
  enabled = false
  role_templates = jsonencode([
    {
      format   = "json"
      template = "{\"source\":\"{{#tojson}}groups{{/tojson}}\"}"
    }
  ])
  rules = jsonencode({
    any = [
      { field = { username = "esadmin" } },
      { field = { groups = "cn=admins,dc=example,dc=com" } },
    ]
  })

  metadata = jsonencode({})
}
	`, roleMappingName)
}
