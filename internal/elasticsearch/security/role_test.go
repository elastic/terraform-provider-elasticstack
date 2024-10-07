package security_test

import (
	"fmt"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/hashicorp/go-version"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

var minSupportedRemoteIndicesVersion = version.Must(version.NewSemver("8.10.0"))
var minSupportedDescriptionVersion = version.Must(version.NewVersion("8.15.0"))

func TestAccResourceSecurityRole(t *testing.T) {
	// generate a random username
	roleName := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	roleNameRemoteIndices := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	roleNameDescription := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		CheckDestroy:             checkResourceSecurityRoleDestroy,
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceSecurityRoleCreate(roleName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_security_role.test", "name", roleName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_security_role.test", "indices.0.allow_restricted_indices", "true"),
					resource.TestCheckTypeSetElemAttr("elasticstack_elasticsearch_security_role.test", "indices.*.names.*", "index1"),
					resource.TestCheckTypeSetElemAttr("elasticstack_elasticsearch_security_role.test", "indices.*.names.*", "index2"),
					resource.TestCheckTypeSetElemAttr("elasticstack_elasticsearch_security_role.test", "cluster.*", "all"),
					resource.TestCheckTypeSetElemAttr("elasticstack_elasticsearch_security_role.test", "run_as.*", "other_user"),
					resource.TestCheckNoResourceAttr("elasticstack_elasticsearch_security_role.test", "global"),
				),
			},
			{
				Config: testAccResourceSecurityRoleUpdate(roleName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_security_role.test", "name", roleName),
					resource.TestCheckTypeSetElemAttr("elasticstack_elasticsearch_security_role.test", "indices.*.names.*", "index1"),
					resource.TestCheckTypeSetElemAttr("elasticstack_elasticsearch_security_role.test", "indices.*.names.*", "index2"),
					resource.TestCheckTypeSetElemAttr("elasticstack_elasticsearch_security_role.test", "cluster.*", "all"),
					resource.TestCheckNoResourceAttr("elasticstack_elasticsearch_security_role.test", "run_as.#"),
					resource.TestCheckNoResourceAttr("elasticstack_elasticsearch_security_role.test", "global"),
					resource.TestCheckNoResourceAttr("elasticstack_elasticsearch_security_role.test", "applications.#"),
					resource.TestCheckNoResourceAttr("elasticstack_elasticsearch_security_role.test", "indices.0.allow_restricted_indices"),
				),
			},
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minSupportedRemoteIndicesVersion),
				Config:   testAccResourceSecurityRoleRemoteIndicesCreate(roleNameRemoteIndices),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_security_role.test", "name", roleNameRemoteIndices),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_security_role.test", "indices.0.allow_restricted_indices", "true"),
					resource.TestCheckTypeSetElemAttr("elasticstack_elasticsearch_security_role.test", "indices.*.names.*", "index1"),
					resource.TestCheckTypeSetElemAttr("elasticstack_elasticsearch_security_role.test", "indices.*.names.*", "index2"),
					resource.TestCheckTypeSetElemAttr("elasticstack_elasticsearch_security_role.test", "cluster.*", "all"),
					resource.TestCheckTypeSetElemAttr("elasticstack_elasticsearch_security_role.test", "run_as.*", "other_user"),
					resource.TestCheckNoResourceAttr("elasticstack_elasticsearch_security_role.test", "global"),
					resource.TestCheckTypeSetElemAttr("elasticstack_elasticsearch_security_role.test", "remote_indices.*.clusters.*", "test-cluster"),
					resource.TestCheckTypeSetElemAttr("elasticstack_elasticsearch_security_role.test", "remote_indices.*.names.*", "sample"),
				),
			},
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minSupportedRemoteIndicesVersion),
				Config:   testAccResourceSecurityRoleRemoteIndicesUpdate(roleNameRemoteIndices),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_security_role.test", "name", roleNameRemoteIndices),
					resource.TestCheckTypeSetElemAttr("elasticstack_elasticsearch_security_role.test", "indices.*.names.*", "index1"),
					resource.TestCheckTypeSetElemAttr("elasticstack_elasticsearch_security_role.test", "indices.*.names.*", "index2"),
					resource.TestCheckTypeSetElemAttr("elasticstack_elasticsearch_security_role.test", "cluster.*", "all"),
					resource.TestCheckNoResourceAttr("elasticstack_elasticsearch_security_role.test", "run_as.#"),
					resource.TestCheckNoResourceAttr("elasticstack_elasticsearch_security_role.test", "global"),
					resource.TestCheckNoResourceAttr("elasticstack_elasticsearch_security_role.test", "applications.#"),
					resource.TestCheckNoResourceAttr("elasticstack_elasticsearch_security_role.test", "indices.0.allow_restricted_indices"),
					resource.TestCheckTypeSetElemAttr("elasticstack_elasticsearch_security_role.test", "remote_indices.*.clusters.*", "test-cluster2"),
					resource.TestCheckTypeSetElemAttr("elasticstack_elasticsearch_security_role.test", "remote_indices.*.names.*", "sample2"),
				),
			},
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minSupportedDescriptionVersion),
				Config:   testAccResourceSecurityRoleDescriptionCreate(roleNameDescription),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_security_role.test", "name", roleNameDescription),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_security_role.test", "description", "test description"),
				),
			},
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minSupportedDescriptionVersion),
				Config:   testAccResourceSecurityRoleDescriptionUpdate(roleNameDescription),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_security_role.test", "name", roleNameDescription),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_security_role.test", "description", "updated test description"),
				),
			},
		},
	})
}

func testAccResourceSecurityRoleCreate(roleName string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_security_role" "test" {
  name        = "%s"

  cluster = ["all"]

  indices {
    names                    = ["index1", "index2"]
    privileges               = ["all"]
    allow_restricted_indices = true
  }

  applications {
    application = "myapp"
    privileges  = ["admin", "read"]
    resources   = ["*"]
  }

  run_as = ["other_user"]

  metadata = jsonencode({
    version = 1
  })
}
	`, roleName)
}

func testAccResourceSecurityRoleUpdate(roleName string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_security_role" "test" {
  name        = "%s"

  cluster = ["all"]

  indices {
    names      = ["index1", "index2"]
    privileges = ["all"]
  }

  metadata = jsonencode({
    version = 1
  })
}
	`, roleName)
}

func testAccResourceSecurityRoleRemoteIndicesCreate(roleName string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_security_role" "test" {
  name    = "%s"
  cluster = ["all"]

  indices {
    names                    = ["index1", "index2"]
    privileges               = ["all"]
    allow_restricted_indices = true
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

  applications {
    application = "myapp"
    privileges  = ["admin", "read"]
    resources   = ["*"]
  }

  run_as = ["other_user"]

  metadata = jsonencode({
    version = 1
  })
}
	`, roleName)
}

func testAccResourceSecurityRoleRemoteIndicesUpdate(roleName string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_security_role" "test" {
  name    = "%s"
  cluster = ["all"]

  indices {
    names      = ["index1", "index2"]
    privileges = ["all"]
  }

  remote_indices {
	clusters = ["test-cluster2"]
	field_security {
	  grant = ["sample"]
	  except = []
	}
	names = ["sample2"]
	privileges = ["create", "read", "write"]
  }

  metadata = jsonencode({
    version = 1
  })
}
	`, roleName)
}

func testAccResourceSecurityRoleDescriptionCreate(roleName string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_security_role" "test" {
  name        = "%s"
  description = "test description"
}
	`, roleName)
}

func testAccResourceSecurityRoleDescriptionUpdate(roleName string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_security_role" "test" {
  name        = "%s"
  description = "updated test description"
}
	`, roleName)
}

func checkResourceSecurityRoleDestroy(s *terraform.State) error {
	client, err := clients.NewAcceptanceTestingClient()
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "elasticstack_elasticsearch_security_role" {
			continue
		}
		compId, _ := clients.CompositeIdFromStr(rs.Primary.ID)

		esClient, err := client.GetESClient()
		if err != nil {
			return err
		}
		req := esClient.Security.GetRole.WithName(compId.ResourceId)
		res, err := esClient.Security.GetRole(req)
		if err != nil {
			return err
		}

		if res.StatusCode != 404 {
			return fmt.Errorf("role (%s) still exists", compId.ResourceId)
		}
	}
	return nil
}
