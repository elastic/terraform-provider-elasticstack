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

package trainedmodeldeployment_test

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

const testResourceName = "elasticstack_elasticsearch_ml_trained_model_deployment.test"

// preCheckMLTrainedModelDeployment ensures a trained model exists in the test
// cluster and that ML nodes have enough capacity to deploy it. The acceptance
// test helper creates a minimal tree_ensemble model that uses negligible memory.
func preCheckMLTrainedModelDeployment(t *testing.T) {
	t.Helper()
	acctest.PreCheck(t)
	acctest.EnsureTrainedModel(t)

	ctx := context.Background()
	client, err := clients.NewAcceptanceTestingElasticsearchScopedClient()
	if err != nil {
		t.Fatal(err)
	}
	es := client.GetESClient()

	_, startErr := es.Ml.StartTrainedModelDeployment(acctest.AccTestTrainedModelID).Do(ctx)
	if startErr != nil {
		errStr := strings.ToLower(startErr.Error())
		if strings.Contains(errStr, "429") ||
			strings.Contains(errStr, "too_many_requests") ||
			strings.Contains(errStr, "no ml nodes") ||
			strings.Contains(errStr, "insufficient memory") ||
			strings.Contains(errStr, "insufficient capacity") ||
			strings.Contains(errStr, "status_exception") ||
			strings.Contains(errStr, "not supported") {
			t.Skipf("skipping test: ML cluster cannot deploy test model: %v", startErr)
		}
		// Other errors (e.g. model already started) are fine.
	}
	// Clean up the probe deployment so the test can manage lifecycle.
	stopDiags := elasticsearch.StopTrainedModelDeployment(ctx, client, acctest.AccTestTrainedModelID, false)
	if stopDiags.HasError() {
		t.Fatalf("failed to stop probe deployment: %v", stopDiags)
	}
}

func testAccCheckMLTrainedModelDeploymentDestroy(s *terraform.State) error {
	ctx := context.Background()
	client, err := clients.NewAcceptanceTestingElasticsearchScopedClient()
	if err != nil {
		return err
	}

	rs, ok := s.RootModule().Resources[testResourceName]
	if !ok {
		return fmt.Errorf("not found: %s", testResourceName)
	}

	modelID := rs.Primary.Attributes["model_id"]
	deploymentID := rs.Primary.Attributes["deployment_id"]
	if deploymentID == "" {
		deploymentID = modelID
	}

	stats, diags := elasticsearch.GetTrainedModelStats(ctx, client, modelID, deploymentID)
	if diags.HasError() {
		return fmt.Errorf("error checking deployment destroy: %v", diags)
	}
	if stats != nil {
		return fmt.Errorf("trained model deployment %s for model %s still exists", deploymentID, modelID)
	}

	return nil
}

func TestAccResourceMLTrainedModelDeployment_basic(t *testing.T) {
	modelID := acctest.AccTestTrainedModelID

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { preCheckMLTrainedModelDeployment(t) },
		CheckDestroy: testAccCheckMLTrainedModelDeploymentDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("basic"),
				ConfigVariables: config.Variables{
					"model_id": config.StringVariable(modelID),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(testResourceName, "model_id", modelID),
					resource.TestCheckResourceAttrSet(testResourceName, "deployment_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "id"),
					resource.TestCheckResourceAttrSet(testResourceName, "state"),
					resource.TestCheckResourceAttrSet(testResourceName, "allocation_status"),
					resource.TestCheckResourceAttrSet(testResourceName, "stats_json"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("basic"),
				ConfigVariables: config.Variables{
					"model_id": config.StringVariable(modelID),
				},
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update_allocations"),
				ConfigVariables: config.Variables{
					"model_id": config.StringVariable(modelID),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(testResourceName, "model_id", modelID),
					resource.TestCheckResourceAttr(testResourceName, "number_of_allocations", "2"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update_adaptive"),
				ConfigVariables: config.Variables{
					"model_id": config.StringVariable(modelID),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(testResourceName, "model_id", modelID),
					resource.TestCheckResourceAttr(testResourceName, "adaptive_allocations.enabled", "true"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update_adaptive"),
				ConfigVariables: config.Variables{
					"model_id": config.StringVariable(modelID),
				},
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("basic"),
				ConfigVariables: config.Variables{
					"model_id": config.StringVariable(modelID),
				},
				ResourceName:      testResourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"force_stop",
					"wait_for",
					"timeouts",
				},
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources[testResourceName]
					if !ok {
						return "", fmt.Errorf("not found: %s", testResourceName)
					}
					return rs.Primary.Attributes["id"], nil
				},
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("basic"),
				ConfigVariables: config.Variables{
					"model_id": config.StringVariable(modelID),
				},
				Destroy: true,
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("force_stop"),
				ConfigVariables: config.Variables{
					"model_id": config.StringVariable(modelID),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(testResourceName, "force_stop", "true"),
				),
			},
		},
	})
}

func TestAccResourceMLTrainedModelDeployment_nonExistentModel(t *testing.T) {
	modelID := fmt.Sprintf("non-existent-model-%s", sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum))

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("non_existent"),
				ConfigVariables: config.Variables{
					"model_id": config.StringVariable(modelID),
				},
				ExpectError: regexp.MustCompile(`(?s)Failed to start trained model deployment`),
			},
		},
	})
}
