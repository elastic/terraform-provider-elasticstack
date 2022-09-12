package security_test

import (
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceSecurityUser(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceSecurityUser,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_security_user.test", "username", "elastic"),
					resource.TestCheckTypeSetElemAttr("data.elasticstack_elasticsearch_security_user.test", "roles.*", "superuser"),
				),
			},
		},
	})
}

const testAccDataSourceSecurityUser = `
provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_security_user" "test" {
  username = "elastic"
}
`
