package security_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccResourceSecurityRoleMapping(t *testing.T) {
	roleMappingName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)

	resource.UnitTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		CheckDestroy:      checkResourceSecurityRoleMappingDestroy,
		ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceSecurityRoleMappingCreate(roleMappingName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_security_role_mapping.test", "name", roleMappingName),
					resource.TestCheckTypeSetElemAttr("elasticstack_elasticsearch_security_role_mapping.test", "enabled", "true"),
					resource.TestCheckTypeSetElemAttr("elasticstack_elasticsearch_security_role_mapping.test", "roles.*", "apm_system"),
					// TODO: Check attributes
				),
			},
			{
				Config: testAccResourceSecurityRoleMappingUpdate(roleMappingName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_security_role_mapping.test", "name", roleMappingName),
					resource.TestCheckTypeSetElemAttr("elasticstack_elasticsearch_security_role_mapping.test", "enabled", "false"),
					resource.TestCheckTypeSetElemAttr("elasticstack_elasticsearch_security_role_mapping.test", "roles.*", "apm_system"),
					resource.TestCheckTypeSetElemAttr("elasticstack_elasticsearch_security_role_mapping.test", "roles.*", "beats_system"),
					// TODO: Check attributes
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
  name    = "%s"
  enabled = true

  roles = [
    "apm_system"
  ]

  rules = jsonencode({
    any = [
      {
        field = {
          username = "*"
        }
      }
    ]
  })
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

  roles = [
    "apm_system",
	"beats_system"
  ]
}
`, roleMappingName)
}

func checkResourceSecurityRoleMappingDestroy(s *terraform.State) error {
	client := acctest.Provider.Meta().(*clients.ApiClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "elasticstack_elasticsearch_security_role_mapping" {
			continue
		}
		compId, _ := clients.CompositeIdFromStr(rs.Primary.ID)

		req := client.GetESClient().Security.GetRoleMapping.WithName(compId.ResourceId)
		res, err := client.GetESClient().Security.GetRoleMapping(req)
		if err != nil {
			return err
		}

		if res.StatusCode != http.StatusNotFound {
			return fmt.Errorf("role mapping (%s) still exists", compId.ResourceId)
		}
	}
	return nil
}
