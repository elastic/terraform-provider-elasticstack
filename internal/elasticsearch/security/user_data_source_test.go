package security_test

import (
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDataSourceSecurityUser(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.Providers,
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
