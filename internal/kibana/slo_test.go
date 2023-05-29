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
	minSupportedVersion := version.Must(version.NewSemver("7.14.0"))

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
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "descripion", "fully sick SLO"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "indicator.type", "sli.apm.transactionDuration"),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "space_id", "default"),
				),
			},
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minSupportedVersion),
				Config:   testAccResourceSloUpdate(sloName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "name", fmt.Sprintf("Updated %s", sloName)),
					resource.TestCheckResourceAttr("elasticstack_kibana_slo.test_slo", "indicator.params.index", "newindex"),
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

resource "elasticstack_kibana_slo" "test_slo" {
	name        = "%s"
	description = "fully sick SLO"
	indicator {
	  type = "sli.apm.transactionDuration"
	  params = {
		environment     = "production"
		service         = "my-service"
		transactionType = "request"
		transactionName = "GET /sup/dawg"
		index           = "my-index"
		threshold       = 500
	  }
	}
  
	time_window {
	  duration   = "1w"
	  isCalendar = true
	}
  
	budgetingMethod = "timeslices"
  
	objective {
	  target          = 0.999
	  timesliceTarget = 0.95
	  timesliceWindow = "5m"
	}
  
	settings {
	  syncDelay = "5m"
	  frequency = "1m"
	}
  
  }
  
`, name)
}

func testAccResourceSloUpdate(name string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_kibana_slo" "test_slo" {
	name        = "Updated %s"
	description = "fully sick SLO"
	indicator {
	  type = "sli.apm.transactionDuration"
	  params = {
		environment     = "production"
		service         = "my-service"
		transactionType = "request"
		transactionName = "GET /sup/dawg"
		index           = "newindex"
		threshold       = 500
	  }
	}
  
	time_window {
	  duration   = "1w"
	  isCalendar = true
	}
  
	budgetingMethod = "timeslices"
  
	objective {
	  target          = 0.999
	  timesliceTarget = 0.95
	  timesliceWindow = "5m"
	}
  
	settings {
	  syncDelay = "5m"
	  frequency = "1m"
	}
  
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

		rule, diags := kibana.GetSlo(context.Background(), client, compId.ResourceId, compId.ClusterId)
		if diags.HasError() {
			return fmt.Errorf("Failed to get slo: %v", diags)
		}

		if rule != nil {
			return fmt.Errorf("SLO (%s) still exists", compId.ResourceId)
		}
	}
	return nil
}
