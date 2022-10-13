package logstash_test

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestResourceLogstashPipeline(t *testing.T) {
	pipelineID := strings.ToLower(sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum))
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		CheckDestroy:      checkResourceLogstashPipelineDestroy,
		ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceLogstashPipelineCreate(pipelineID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_logstash_pipeline.test", "pipeline_id", pipelineID),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_logstash_pipeline.test", "description", "Description of Logstash Pipeline"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_logstash_pipeline.test", "pipeline", "input{} filter{} output{}"),
				),
			},
			// {
			// 	Config: testAccResourceLogstashPipelineUpdate(pipelineID),
			// 	Check:  resource.ComposeTestCheckFunc(),
			// },
		},
	})
}

func testAccResourceLogstashPipelineCreate(pipelineID string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_logstash_pipeline" "test" {
  pipeline_id = "%s"
	description = "Description of Logstash Pipeline"
	pipeline = "input{} filter{} output{}"
}
	`, pipelineID)
}

// func testAccResourceLogstashPipelineUpdate(pipelineID string) string {

// }

func checkResourceLogstashPipelineDestroy(s *terraform.State) error {

	client := acctest.Provider.Meta().(*clients.ApiClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "elasticstack_elasticsearch_logstash_pipeline" {
			continue
		}
		compId, _ := clients.CompositeIdFromStr(rs.Primary.ID)

		res, err := client.GetESClient().LogstashGetPipeline(compId.ResourceId)
		if err != nil {
			return err
		}

		if res.StatusCode != http.StatusNotFound {
			return fmt.Errorf("logstash pipeline (%s) still exists", compId.ResourceId)
		}
	}
	return nil
}
