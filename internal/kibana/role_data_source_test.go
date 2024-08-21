package kibana_test

import (
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceKibanaSecurityRole(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceSecurityRole,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.elasticstack_kibana_security_role.test", "name", "data_source_test"),
					resource.TestCheckNoResourceAttr("data.elasticstack_kibana_security_role.test", "kibana.0.feature.#"),
					resource.TestCheckNoResourceAttr("data.elasticstack_kibana_security_role.test", "elasticsearch.0.indices.0.field_security.#"),
					utils.TestCheckResourceListAttr("data.elasticstack_kibana_security_role.test", "elasticsearch.0.run_as", []string{"elastic", "kibana"}),
					utils.TestCheckResourceListAttr("data.elasticstack_kibana_security_role.test", "kibana.0.base", []string{"all"}),
					utils.TestCheckResourceListAttr("data.elasticstack_kibana_security_role.test", "kibana.0.spaces", []string{"default"}),
				),
			},
		},
	})
}

const testAccDataSourceSecurityRole = `
provider "elasticstack" {
  elasticsearch {}
  kibana {}
}


resource "elasticstack_kibana_security_role" "test" {
	name    = "data_source_test"
	elasticsearch {
	  cluster = [ "create_snapshot" ]
	  indices {
		names = ["sample"]
		privileges = ["create", "read", "write"]
	  }
	  remote_indices {
	    clusters = ["test-cluster"]
	    field_security {
	      grant = ["sample"]
	      except = []
	    }
	    names = ["sample"]
	    privileges = ["create", "read", "write"]
	  }
	  run_as = ["kibana", "elastic"]
	}
	kibana {
	  base = [ "all" ]
	  spaces = ["default"]
	}
}

data "elasticstack_kibana_security_role" "test" {
  name = elasticstack_kibana_security_role.test.name
}
`
