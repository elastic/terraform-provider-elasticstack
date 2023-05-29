package kibana_test

import (
	"fmt"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccResourceSpace(t *testing.T) {
	spaceId := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		CheckDestroy:             checkResourceSpaceDestroy,
		ProtoV5ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceSpaceCreate(spaceId),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_space.test_space", "space_id", spaceId),
					resource.TestCheckResourceAttr("elasticstack_kibana_space.test_space", "name", fmt.Sprintf("Name %s", spaceId)),
					resource.TestCheckResourceAttr("elasticstack_kibana_space.test_space", "description", "Test Space"),
				),
			},
			{
				Config: testAccResourceSpaceUpdate(spaceId),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_space.test_space", "space_id", spaceId),
					resource.TestCheckResourceAttr("elasticstack_kibana_space.test_space", "name", fmt.Sprintf("Updated %s", spaceId)),
					resource.TestCheckResourceAttr("elasticstack_kibana_space.test_space", "description", "Updated space description"),
					resource.TestCheckTypeSetElemAttr("elasticstack_kibana_space.test_space", "disabled_features.*", "ingestManager"),
					resource.TestCheckTypeSetElemAttr("elasticstack_kibana_space.test_space", "disabled_features.*", "enterpriseSearch"),
				),
			},
		},
	})
}

func testAccResourceSpaceCreate(id string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_kibana_space" "test_space" {
  space_id    = "%s"
  name        = "%s"
  description = "Test Space"
}
	`, id, fmt.Sprintf("Name %s", id))
}

func testAccResourceSpaceUpdate(id string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_kibana_space" "test_space" {
  space_id          = "%s"
  name              = "%s"
  description       = "Updated space description"
	disabled_features = ["ingestManager", "enterpriseSearch"]
}
	`, id, fmt.Sprintf("Updated %s", id))
}

func checkResourceSpaceDestroy(s *terraform.State) error {
	client, err := clients.NewAcceptanceTestingClient()
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "elasticstack_kibana_space" {
			continue
		}
		compId, _ := clients.CompositeIdFromStr(rs.Primary.ID)

		kibanaClient, err := client.GetKibanaClient()
		if err != nil {
			return err
		}
		res, err := kibanaClient.KibanaSpaces.Get(compId.ResourceId)
		if err != nil {
			return err
		}

		if res != nil {
			return fmt.Errorf("Space (%s) still exists", compId.ResourceId)
		}
	}
	return nil
}
