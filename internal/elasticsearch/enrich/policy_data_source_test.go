package enrich_test

import (
	"fmt"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceEnrichPolicy(t *testing.T) {
	name := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV5ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccEnrichPolicyDataSource(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_enrich_policy.test", "name", name),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_enrich_policy.test", "policy_type", "match"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_enrich_policy.test", "match_field", "email"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_enrich_policy.test", "indices.0", name),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_enrich_policy.test", "enrich_fields.0", "first_name"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_enrich_policy.test", "enrich_fields.1", "last_name"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_enrich_policy.test", "query", "{\"match_all\":{}}"),
				),
			},
		},
	})
}

func testAccEnrichPolicyDataSource(name string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_index" "my_index" {
  name = "%s"

  mappings = jsonencode({
    properties = {
      email      = { type = "text" }
      first_name = { type = "text" }
      last_name  = { type = "text" }
    }
  })
  deletion_protection = false
}

resource "elasticstack_elasticsearch_enrich_policy" "policy" {
  name          = "%s"
  policy_type   = "match"
  indices       = [elasticstack_elasticsearch_index.my_index.name]
  match_field   = "email"
  enrich_fields = ["first_name", "last_name"]
	query = <<-EOD
	{"match_all": {}}
	EOD
}

data "elasticstack_elasticsearch_enrich_policy" "test" {
	name = elasticstack_elasticsearch_enrich_policy.policy.name
}
	`, name, name)
}
