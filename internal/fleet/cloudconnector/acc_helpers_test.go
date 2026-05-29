// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package cloudconnector_test

import (
	"context"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	fleetclient "github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/hashicorp/go-version"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

const resourceName = "elasticstack_fleet_cloud_connector.test"

var minCloudConnectorVersion = version.Must(version.NewVersion("9.2.0"))

func accRandSuffix() string {
	return sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
}

// accExternalID returns a 20-character value compatible with Fleet cloud
// connector external ID secret reference validation.
func accExternalID() string {
	return sdkacctest.RandStringFromCharSet(20, sdkacctest.CharSetAlphaNum)
}

func checkCloudConnectorDestroy(s *terraform.State) error {
	client, err := clients.NewAcceptanceTestingKibanaScopedClient()
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "elasticstack_fleet_cloud_connector" {
			continue
		}

		connectorID := rs.Primary.Attributes["cloud_connector_id"]
		spaceID := rs.Primary.Attributes["space_id"]
		if spaceID == "" {
			spaceID = "default"
		}

		item, diags := fleetclient.GetCloudConnector(context.Background(), client.GetFleetClient(), spaceID, connectorID)
		if diags.HasError() {
			return diagutil.FwDiagsAsError(diags)
		}
		if item != nil {
			return fmt.Errorf("cloud connector id=%q still exists in space %q", connectorID, spaceID)
		}
	}

	return nil
}

func testCheckCloudConnectorHasTypedAWS() resource.TestCheckFunc {
	return resource.ComposeTestCheckFunc(
		resource.TestCheckResourceAttrSet(resourceName, "aws.role_arn"),
		resource.TestCheckResourceAttrSet(resourceName, "vars.role_arn.type"),
		resource.TestCheckResourceAttrSet(resourceName, "vars.external_id.type"),
		resource.TestCheckResourceAttrSet(resourceName, "aws.external_id_secret_ref.id"),
	)
}

func testCheckCloudConnectorHasTypedAzure() resource.TestCheckFunc {
	return resource.ComposeTestCheckFunc(
		resource.TestCheckResourceAttrSet(resourceName, "azure.cloud_connector_id"),
		resource.TestCheckResourceAttrSet(resourceName, "azure.tenant_id_secret_ref.id"),
		resource.TestCheckResourceAttrSet(resourceName, "azure.client_id_secret_ref.id"),
		resource.TestCheckResourceAttrSet(resourceName, "vars.tenant_id.type"),
		resource.TestCheckResourceAttrSet(resourceName, "vars.client_id.type"),
		resource.TestCheckResourceAttrSet(resourceName, fmt.Sprintf("vars.%s.type", "azure_credentials_cloud_connector_id")),
	)
}

func testCheckCaptureCloudConnectorID(target *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok || rs.Primary == nil {
			return fmt.Errorf("resource %s not found in state", resourceName)
		}
		id := rs.Primary.Attributes["cloud_connector_id"]
		if id == "" {
			return fmt.Errorf("cloud_connector_id not set in state")
		}
		*target = id
		return nil
	}
}

func testCheckCloudConnectorHasVarKeys(keys ...string) resource.TestCheckFunc {
	checks := make([]resource.TestCheckFunc, 0, len(keys))
	for _, key := range keys {
		checks = append(checks, resource.TestCheckResourceAttrSet(resourceName, fmt.Sprintf("vars.%s.type", key)))
	}
	return resource.ComposeTestCheckFunc(checks...)
}

type expectWriteOnlyDriftPlanCheck struct {
	resourceAddress string
	attributePath   string
}

func (c expectWriteOnlyDriftPlanCheck) CheckPlan(_ context.Context, req plancheck.CheckPlanRequest, resp *plancheck.CheckPlanResponse) {
	for _, rc := range req.Plan.ResourceChanges {
		if rc.Address != c.resourceAddress {
			continue
		}
		if rc.Change.Actions.Update() {
			return
		}
	}
	resp.Error = fmt.Errorf("expected update for %s after write-only drift on %s", c.resourceAddress, c.attributePath)
}

func expectWriteOnlyDriftPlanChecks(attributePath string) resource.ConfigPlanChecks {
	checks := []plancheck.PlanCheck{
		plancheck.ExpectNonEmptyPlan(),
		plancheck.ExpectUnknownValue(resourceName, tfjsonpath.New("updated_at")),
		expectWriteOnlyDriftPlanCheck{
			resourceAddress: resourceName,
			attributePath:   attributePath,
		},
	}
	return resource.ConfigPlanChecks{
		PostApplyPreRefresh: checks,
	}
}
