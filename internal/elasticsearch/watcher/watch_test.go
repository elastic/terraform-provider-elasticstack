package watcher_test

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

func TestResourceWatch(t *testing.T) {
	watchID := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		CheckDestroy:      checkResourceWatchDestroy,
		ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceWatchCreate(watchID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_watcher_watch.test", "watch_id", watchID),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_watcher_watch.test", "active", "false"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_watcher_watch.test", "body", `"json":"true"`),
				),
			},
		},
	})
}

func testAccResourceWatchCreate(watchID string) string {
	return fmt.Sprintf(`
 provider "elasticstack" {
   elasticsearch {}
 }
 resource "elasticstack_elasticsearch_watcher_watch" "test" {
   watch_id = "%s"
 	active = false
 	body = jsonencode({
		"json" = "true"
	})
 }
 	`, watchID)
}

func checkResourceWatchDestroy(s *terraform.State) error {

	client := acctest.Provider.Meta().(*clients.ApiClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "elasticstack_elasticsearch_watcher_watch" {
			continue
		}
		compId, _ := clients.CompositeIdFromStr(rs.Primary.ID)

		res, err := client.GetESClient().Watcher.GetWatch(compId.ResourceId)
		if err != nil {
			return err
		}

		if res.StatusCode != http.StatusNotFound {
			return fmt.Errorf("watch (%s) still exists", compId.ResourceId)
		}
	}
	return nil
}
