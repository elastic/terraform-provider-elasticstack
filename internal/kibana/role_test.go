package kibana_test

import (
	"fmt"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccResourceKibanaSecurityRole(t *testing.T) {
	// generate a random role name
	roleName := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		CheckDestroy:             checkResourceSecurityRoleDestroy,
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceSecurityRoleCreate(roleName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_security_role.test", "name", roleName),
					resource.TestCheckNoResourceAttr("elasticstack_kibana_security_role.test", "kibana.0.base.#"),
					resource.TestCheckNoResourceAttr("elasticstack_kibana_security_role.test", "elasticsearch.0.run_as.#"),
					utils.TestCheckResourceListAttr("elasticstack_kibana_security_role.test", "elasticsearch.0.indices.0.names", []string{"sample"}),
					utils.TestCheckResourceListAttr("elasticstack_kibana_security_role.test", "elasticsearch.0.indices.0.field_security.0.grant", []string{"sample"}),
					utils.TestCheckResourceListAttr("elasticstack_kibana_security_role.test", "kibana.0.feature.2.privileges", []string{"minimal_read", "store_search_session", "url_create"}),
					utils.TestCheckResourceListAttr("elasticstack_kibana_security_role.test", "kibana.0.spaces", []string{"default"}),
				),
			},
			{
				Config: testAccResourceSecurityRoleUpdate(roleName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_security_role.test", "name", roleName),
					resource.TestCheckNoResourceAttr("elasticstack_kibana_security_role.test", "kibana.0.feature.#"),
					resource.TestCheckNoResourceAttr("elasticstack_kibana_security_role.test", "elasticsearch.0.indices.0.field_security.#"),
					utils.TestCheckResourceListAttr("elasticstack_kibana_security_role.test", "elasticsearch.0.run_as", []string{"elastic", "kibana"}),
					utils.TestCheckResourceListAttr("elasticstack_kibana_security_role.test", "kibana.0.base", []string{"all"}),
					utils.TestCheckResourceListAttr("elasticstack_kibana_security_role.test", "kibana.0.spaces", []string{"default"}),
				),
			},
		},
	})
}

func testAccResourceSecurityRoleCreate(roleName string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_kibana_security_role" "test" {
  name    = "%s"
  elasticsearch {
    cluster = [ "create_snapshot" ]
    indices {
      field_security {
        grant = ["sample"]
        except = []
      }
      names = ["sample"]
      privileges = ["create", "read", "write"]
    }
  }
  kibana {
    feature {
      name = "actions"
      privileges = ["read"]
    }
    feature {
      name = "advancedSettings"
      privileges = ["read"]
    }
    feature {
      name = "discover"
      privileges = ["minimal_read", "url_create", "store_search_session"]
    }
    feature {
      name = "generalCases"
      privileges = ["minimal_read", "cases_delete"]
    }
    feature {
      name = "observabilityCases"
      privileges = ["minimal_read", "cases_delete"]
    }
    feature {
      name = "osquery"
      privileges = ["minimal_read", "live_queries_all", "run_saved_queries", "saved_queries_read", "packs_all"]
    }
    feature {
      name = "rulesSettings"
      privileges = ["minimal_read", "readFlappingSettings"]
    }
    feature {
      name = "securitySolutionCases"
      privileges = ["minimal_read", "cases_delete"]
    }
    feature {
      name = "visualize"
      privileges = ["minimal_read", "url_create"]
    }

    spaces = ["default"]
  }
}
	`, roleName)
}

func testAccResourceSecurityRoleUpdate(roleName string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_kibana_security_role" "test" {
	name    = "%s"
	elasticsearch {
	  cluster = [ "create_snapshot" ]
	  indices {
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
	`, roleName)
}

func checkResourceSecurityRoleDestroy(s *terraform.State) error {
	client, err := clients.NewAcceptanceTestingClient()
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "elasticstack_kibana_security_role" {
			continue
		}
		compId := rs.Primary.ID

		kibanaClient, err := client.GetKibanaClient()
		if err != nil {
			return err
		}
		res, err := kibanaClient.KibanaRoleManagement.Get(compId)
		if err != nil || res != nil {
			return fmt.Errorf("Role (%s) still exists", compId)
		}
	}
	return nil
}
