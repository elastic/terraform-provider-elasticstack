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

func TestAccResourceSecurityUser(t *testing.T) {
	// generate a random username
	username := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)

	resource.UnitTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		CheckDestroy:      checkResourceSecurityUserDestroy,
		ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceSecurityUserCreate(username),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_security_user.test", "username", username),
					resource.TestCheckTypeSetElemAttr("elasticstack_elasticsearch_security_user.test", "roles.*", "kibana_user"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_security_user.test", "email", ""),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_security_user.test", "metadata", `{"test":"abc"}`),
				),
				ExpectNonEmptyPlan: true,
			},
			{
				Config: testAccResourceSecurityUpdate(username),
				Check:  resource.TestCheckResourceAttr("elasticstack_elasticsearch_security_user.test", "email", "test@example.com"),
			},
		},
	})
}

func testAccResourceSecurityUserCreate(username string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_security_user" "test" {
  username  = "%s"
  roles     = ["kibana_user"]
  full_name = "Test User"
  password  = "qwerty123"
  metadata = jsonencode({
    test     = "abc"
    _ignored = true
  })
}
	`, username)
}

func testAccResourceSecurityUpdate(username string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_security_user" "test" {
  username  = "%s"
  roles     = ["kibana_user"]
  full_name = "Test User"
  email     = "test@example.com"
  password  = "qwerty123"
}
	`, username)
}

func checkResourceSecurityUserDestroy(s *terraform.State) error {
	client := acctest.Provider.Meta().(*clients.ApiClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "elasticstack_elasticsearch_security_user" {
			continue
		}
		compId, _ := clients.CompositeIdFromStr(rs.Primary.ID)

		req := client.GetESClient().Security.GetUser.WithUsername(compId.ResourceId)
		res, err := client.GetESClient().Security.GetUser(req)
		if err != nil {
			return err
		}

		if res.StatusCode != 404 {
			return fmt.Errorf("User (%s) still exists", compId.ResourceId)
		}
	}
	return nil
}
