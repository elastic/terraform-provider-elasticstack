package kibana_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibana"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/hashicorp/go-version"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccResourceSlo(t *testing.T) {
	minSupportedVersion := version.Must(version.NewSemver("8.8.0"))

	sloName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		CheckDestroy:             checkResourceSloDestroy,
		ProtoV5ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minSupportedVersion),
				Config:   testAccResourceSloCreate(sloName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "name", sloName),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "description", "fully sick SLO"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "indicator.0.type", "sli.apm.transactionDuration"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "space_id", "default"),
				),
			},
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minSupportedVersion),
				Config:   testAccResourceSloUpdate(sloName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "name", fmt.Sprintf("Updated %s", sloName)),
				),
			},
		},
	})
}

func testAccResourceSloCreate(name string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_elasticsearch_index" "my_index" {
	name = "my-index"
	deletion_protection = false
}  

resource "elasticstack_kibana_slo" "test_slo" {
	name        = "%s"
	description = "fully sick SLO"
	indicator {
	  type = "sli.apm.transactionDuration"
	  params {
		environment     = "production"
		service         = "my-service"
		transaction_type = "request"
		transaction_name = "GET /sup/dawg"
		index           = "my-index"
		threshold       = 500
	  }
	}
  
	time_window {
	  duration   = "1w"
	  type = "rolling"
	}
  
	budgeting_method = "timeslices"
  
	objective {
	  target          = 0.999
	  timeslice_target = 0.95
	  timeslice_window = "5m"
	}
  
	settings {
	  sync_delay = "5m"
	  frequency = "1m"
	}

	depends_on = [elasticstack_elasticsearch_index.my_index]
  
  }
  
`, name)
}

func testAccResourceSloUpdate(name string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
  kibana {}
}


resource "elasticstack_elasticsearch_index" "my_index" {
	name = "my-index"
	deletion_protection = false
} 

resource "elasticstack_kibana_slo" "test_slo" {
	name        = "Updated %s"
	description = "fully sick SLO"
	indicator {
	  type = "sli.apm.transactionDuration"
	  params {
		environment     = "production"
		service         = "my-service"
		transaction_type = "request"
		transaction_name = "GET /sup/dawg"
		index           = "my-index"
		threshold       = 500
	  }
	}
  
	time_window {
	  duration   = "1w"
	  type = "rolling"
	}
  
	budgeting_method = "timeslices"
  
	objective {
	  target          = 0.999
	  timeslice_target = 0.95
	  timeslice_window = "5m"
	}
  
	settings {
	  sync_delay = "5m"
	  frequency = "1m"
	}
  
	depends_on = [elasticstack_elasticsearch_index.my_index]


  }
  
`, name)
}

func checkResourceSloDestroy(s *terraform.State) error {
	client, err := clients.NewAcceptanceTestingClient()
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "elasticstack_kibana_slo" {
			continue
		}
		compId, _ := clients.CompositeIdFromStr(rs.Primary.ID)
		fmt.Printf("Checking for SLO (%s)\n", compId.ResourceId)

		slo, diags := kibana.GetSlo(context.Background(), client, compId.ResourceId, compId.ClusterId)
		if diags.HasError() {
			if len(diags) > 1 || diags[0].Summary != "404 Not Found" {
				return fmt.Errorf("Failed to check if SLO was destroyed: %v", diags)
			}
		}

		if slo != nil {
			return fmt.Errorf("SLO (%s) still exists", compId.ResourceId)
		}
	}
	return nil
}
