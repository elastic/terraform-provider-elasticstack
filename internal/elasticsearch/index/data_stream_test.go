package index_test

import (
	"fmt"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccResourceDataStream(t *testing.T) {
	// generate renadom name
	dsName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlpha)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		CheckDestroy:      checkResourceDataStreamDestroy,
		ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceDataStreamCreate(dsName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_data_stream.test_ds", "name", dsName),
					// check some computed fields
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_data_stream.test_ds", "indices.#", "1"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_data_stream.test_ds", "template", dsName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_data_stream.test_ds", "ilm_policy", dsName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_data_stream.test_ds", "hidden", "false"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_data_stream.test_ds", "system", "false"),
				),
			},
		},
	})
}

func testAccResourceDataStreamCreate(name string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_index_lifecycle" "test_ilm" {
  name = "%s"

  hot {
    min_age = "1h"
    set_priority {
      priority = 10
    }
    rollover {
      max_age = "1d"
    }
    readonly {}
  }

  delete {
    min_age = "2d"
    delete {}
  }
}

resource "elasticstack_elasticsearch_index_template" "test_ds_template" {
  name = "%s"

  index_patterns = ["%s*"]

  template {
    // make sure our template uses prepared ILM policy
    settings = jsonencode({
      "lifecycle.name" = elasticstack_elasticsearch_index_lifecycle.test_ilm.name
    })
  }

  data_stream {}
}

// and now we can create data stream based on the index template
resource "elasticstack_elasticsearch_data_stream" "test_ds" {
  name = "%s"

  // make sure that template is created before the data stream
  depends_on = [
    elasticstack_elasticsearch_index_template.test_ds_template
  ]
}
	`, name, name, name, name)
}

func checkResourceDataStreamDestroy(s *terraform.State) error {
	client := acctest.Provider.Meta().(*clients.ApiClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "elasticstack_elasticsearch_data_stream" {
			continue
		}
		compId, _ := clients.CompositeIdFromStr(rs.Primary.ID)

		req := client.GetESClient().Indices.GetDataStream.WithName(compId.ResourceId)
		res, err := client.GetESClient().Indices.GetDataStream(req)
		if err != nil {
			return err
		}

		if res.StatusCode != 404 {
			return fmt.Errorf("Data Stream (%s) still exists", compId.ResourceId)
		}
	}
	return nil
}
