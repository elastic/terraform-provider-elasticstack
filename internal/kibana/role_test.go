package kibana_test

import (
	"fmt"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/acctest/checks"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/hashicorp/go-version"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccResourceKibanaSecurityRole(t *testing.T) {
	// generate a random role name
	roleName := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	roleNameRemoteIndices := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	minSupportedRemoteIndicesVersion := version.Must(version.NewSemver("8.10.0"))
	minSupportedDescriptionVersion := version.Must(version.NewVersion("8.15.0"))

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
					checks.TestCheckResourceListAttr("elasticstack_kibana_security_role.test", "elasticsearch.0.indices.0.names", []string{"sample"}),
					checks.TestCheckResourceListAttr("elasticstack_kibana_security_role.test", "elasticsearch.0.indices.0.field_security.0.grant", []string{"sample"}),
					checks.TestCheckResourceListAttr("elasticstack_kibana_security_role.test", "kibana.0.feature.2.privileges", []string{"minimal_read", "store_search_session", "url_create"}),
					checks.TestCheckResourceListAttr("elasticstack_kibana_security_role.test", "kibana.0.spaces", []string{"default"}),
				),
			},
			{
				Config: testAccResourceSecurityRoleUpdate(roleName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_security_role.test", "name", roleName),
					resource.TestCheckNoResourceAttr("elasticstack_kibana_security_role.test", "kibana.0.feature.#"),
					resource.TestCheckNoResourceAttr("elasticstack_kibana_security_role.test", "elasticsearch.0.indices.0.field_security.#"),
					checks.TestCheckResourceListAttr("elasticstack_kibana_security_role.test", "elasticsearch.0.run_as", []string{"elastic", "kibana"}),
					checks.TestCheckResourceListAttr("elasticstack_kibana_security_role.test", "kibana.0.base", []string{"all"}),
					checks.TestCheckResourceListAttr("elasticstack_kibana_security_role.test", "kibana.0.spaces", []string{"default"}),
				),
			},
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minSupportedDescriptionVersion),
				Config:   testAccResourceSecurityRoleWithDescription(roleName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_security_role.test", "name", roleName),
					resource.TestCheckNoResourceAttr("elasticstack_kibana_security_role.test", "kibana.0.feature.#"),
					resource.TestCheckNoResourceAttr("elasticstack_kibana_security_role.test", "elasticsearch.0.indices.0.field_security.#"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_role.test", "description", "Role description"),
				),
			},
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minSupportedRemoteIndicesVersion),
				Config:   testAccResourceSecurityRoleRemoteIndicesCreate(roleNameRemoteIndices),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_security_role.test", "name", roleNameRemoteIndices),
					resource.TestCheckNoResourceAttr("elasticstack_kibana_security_role.test", "kibana.0.base.#"),
					resource.TestCheckNoResourceAttr("elasticstack_kibana_security_role.test", "elasticsearch.0.run_as.#"),
					checks.TestCheckResourceListAttr("elasticstack_kibana_security_role.test", "elasticsearch.0.indices.0.names", []string{"sample"}),
					checks.TestCheckResourceListAttr("elasticstack_kibana_security_role.test", "elasticsearch.0.indices.0.field_security.0.grant", []string{"sample"}),
					checks.TestCheckResourceListAttr("elasticstack_kibana_security_role.test", "kibana.0.feature.2.privileges", []string{"minimal_read", "store_search_session", "url_create"}),
					checks.TestCheckResourceListAttr("elasticstack_kibana_security_role.test", "kibana.0.spaces", []string{"default"}),
					checks.TestCheckResourceListAttr("elasticstack_kibana_security_role.test", "elasticsearch.0.remote_indices.0.clusters", []string{"test-cluster"}),
					checks.TestCheckResourceListAttr("elasticstack_kibana_security_role.test", "elasticsearch.0.remote_indices.0.field_security.0.grant", []string{"sample"}),
					checks.TestCheckResourceListAttr("elasticstack_kibana_security_role.test", "elasticsearch.0.remote_indices.0.names", []string{"sample"}),
					checks.TestCheckResourceListAttr("elasticstack_kibana_security_role.test", "elasticsearch.0.remote_indices.0.privileges", []string{"create", "read", "write"}),
				),
			},
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minSupportedRemoteIndicesVersion),
				Config:   testAccResourceSecurityRoleRemoteIndicesUpdate(roleNameRemoteIndices),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_security_role.test", "name", roleNameRemoteIndices),
					resource.TestCheckNoResourceAttr("elasticstack_kibana_security_role.test", "kibana.0.feature.#"),
					resource.TestCheckNoResourceAttr("elasticstack_kibana_security_role.test", "elasticsearch.0.indices.0.field_security.#"),
					checks.TestCheckResourceListAttr("elasticstack_kibana_security_role.test", "elasticsearch.0.run_as", []string{"elastic", "kibana"}),
					checks.TestCheckResourceListAttr("elasticstack_kibana_security_role.test", "kibana.0.base", []string{"all"}),
					checks.TestCheckResourceListAttr("elasticstack_kibana_security_role.test", "kibana.0.spaces", []string{"default"}),
					checks.TestCheckResourceListAttr("elasticstack_kibana_security_role.test", "elasticsearch.0.remote_indices.0.clusters", []string{"test-cluster2"}),
					checks.TestCheckResourceListAttr("elasticstack_kibana_security_role.test", "elasticsearch.0.remote_indices.0.field_security.0.grant", []string{"sample2"}),
					checks.TestCheckResourceListAttr("elasticstack_kibana_security_role.test", "elasticsearch.0.remote_indices.0.names", []string{"sample2"}),
					checks.TestCheckResourceListAttr("elasticstack_kibana_security_role.test", "elasticsearch.0.remote_indices.0.privileges", []string{"create", "read", "write"}),
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

func testAccResourceSecurityRoleWithDescription(roleName string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_kibana_security_role" "test" {
	name    = "%s"
	description = "Role description"
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

func testAccResourceSecurityRoleRemoteIndicesCreate(roleName string) string {
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
    remote_indices {
	  clusters = ["test-cluster"]
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

func testAccResourceSecurityRoleRemoteIndicesUpdate(roleName string) string {
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
      remote_indices {
	    clusters = ["test-cluster2"]
        field_security {
          grant = ["sample2"]
          except = []
        }
        names = ["sample2"]
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
