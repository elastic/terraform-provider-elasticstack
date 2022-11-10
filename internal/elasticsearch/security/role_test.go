package security_test

import (
	"fmt"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccResourceSecurityRole(t *testing.T) {
	// generate a random username
	roleName := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		CheckDestroy:      checkResourceSecurityRoleDestroy,
		ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceSecurityRoleCreate(roleName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_security_role.test", "name", roleName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_security_role.test", "indices.0.allow_restricted_indices", "true"),
					resource.TestCheckTypeSetElemAttr("elasticstack_elasticsearch_security_role.test", "indices.*.names.*", "index1"),
					resource.TestCheckTypeSetElemAttr("elasticstack_elasticsearch_security_role.test", "indices.*.names.*", "index2"),
					resource.TestCheckTypeSetElemAttr("elasticstack_elasticsearch_security_role.test", "cluster.*", "all"),
					resource.TestCheckTypeSetElemAttr("elasticstack_elasticsearch_security_role.test", "run_as.*", "other_user"),
					resource.TestCheckNoResourceAttr("elasticstack_elasticsearch_security_role.test", "global"),
				),
			},
			{
				Config: testAccResourceSecurityRoleUpdate(roleName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_security_role.test", "name", roleName),
					resource.TestCheckTypeSetElemAttr("elasticstack_elasticsearch_security_role.test", "indices.*.names.*", "index1"),
					resource.TestCheckTypeSetElemAttr("elasticstack_elasticsearch_security_role.test", "indices.*.names.*", "index2"),
					resource.TestCheckTypeSetElemAttr("elasticstack_elasticsearch_security_role.test", "cluster.*", "all"),
					resource.TestCheckNoResourceAttr("elasticstack_elasticsearch_security_role.test", "run_as.#"),
					resource.TestCheckNoResourceAttr("elasticstack_elasticsearch_security_role.test", "global"),
					resource.TestCheckNoResourceAttr("elasticstack_elasticsearch_security_role.test", "applications.#"),
					resource.TestCheckNoResourceAttr("elasticstack_elasticsearch_security_role.test", "indices.0.allow_restricted_indices"),
				),
			},
		},
	})
}

func testAccResourceSecurityRoleCreate(roleName string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_security_role" "test" {
  name    = "%s"
  cluster = ["all"]

  indices {
    names                    = ["index1", "index2"]
    privileges               = ["all"]
    allow_restricted_indices = true
  }

  applications {
    application = "myapp"
    privileges  = ["admin", "read"]
    resources   = ["*"]
  }

  run_as = ["other_user"]

  metadata = jsonencode({
    version = 1
  })
}
	`, roleName)
}

func testAccResourceSecurityRoleUpdate(roleName string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_security_role" "test" {
  name    = "%s"
  cluster = ["all"]

  indices {
    names      = ["index1", "index2"]
    privileges = ["all"]
  }

  metadata = jsonencode({
    version = 1
  })
}
	`, roleName)
}

func checkResourceSecurityRoleDestroy(s *terraform.State) error {
	client := acctest.Provider.Meta().(*clients.ApiClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "elasticstack_elasticsearch_security_role" {
			continue
		}
		compId, _ := clients.CompositeIdFromStr(rs.Primary.ID)

		req := client.GetESClient().Security.GetRole.WithName(compId.ResourceId)
		res, err := client.GetESClient().Security.GetRole(req)
		if err != nil {
			return err
		}

		if res.StatusCode != 404 {
			return fmt.Errorf("Role (%s) still exists", compId.ResourceId)
		}
	}
	return nil
}
