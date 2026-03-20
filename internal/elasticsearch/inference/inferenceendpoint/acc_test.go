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

package inferenceendpoint_test

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	esclient "github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/inference/inferenceendpoint"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// skipValidateAndStart sets xpack.inference.skip_validate_and_start=true so
// inference endpoint tests can run without a reachable upstream service, and
// resets it at the end of the test via t.Cleanup.
func skipValidateAndStart(t *testing.T) {
	t.Helper()

	apiClient, err := clients.NewAcceptanceTestingClient()
	if err != nil {
		t.Fatalf("failed to create acceptance testing client: %v", err)
	}

	ctx := context.Background()
	if diags := esclient.PutSettings(ctx, apiClient, map[string]any{
		"persistent": map[string]any{"xpack.inference.skip_validate_and_start": true},
	}); diags.HasError() {
		t.Fatalf("failed to set xpack.inference.skip_validate_and_start: %v", diags)
	}

	t.Cleanup(func() {
		if diags := esclient.PutSettings(ctx, apiClient, map[string]any{
			"persistent": map[string]any{"xpack.inference.skip_validate_and_start": nil},
		}); diags.HasError() {
			t.Errorf("failed to reset xpack.inference.skip_validate_and_start: %v", diags)
		}
	})
}

func TestAccResourceInferenceEndpoint(t *testing.T) {
	skipFunc := versionutils.CheckIfVersionIsUnsupported(inferenceendpoint.MinSupportedVersion)
	if skip, err := skipFunc(); err != nil {
		t.Fatalf("failed to check version: %v", err)
	} else if skip {
		t.Skipf("Test requires Elasticsearch v%s or higher", inferenceendpoint.MinSupportedVersion)
	}

	inferenceID := fmt.Sprintf("test-inference-%s", sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); skipValidateAndStart(t) },
		CheckDestroy: checkInferenceEndpointDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"inference_id": config.StringVariable(inferenceID),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_inference_endpoint.test", "inference_id", inferenceID),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_inference_endpoint.test", "task_type", "completion"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_inference_endpoint.test", "service", "openai"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables: config.Variables{
					"inference_id": config.StringVariable(inferenceID),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_inference_endpoint.test", "inference_id", inferenceID),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_inference_endpoint.test", "task_type", "completion"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_inference_endpoint.test", "service", "openai"),
				),
			},
			// Re-apply with no change should produce no diff.
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables: config.Variables{
					"inference_id": config.StringVariable(inferenceID),
				},
				PlanOnly: true,
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables: config.Variables{
					"inference_id": config.StringVariable(inferenceID),
				},
				ResourceName:            "elasticstack_elasticsearch_inference_endpoint.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"service_settings"},
			},
		},
	})
}

func TestAccResourceInferenceEndpointRequiresReplace(t *testing.T) {
	skipFunc := versionutils.CheckIfVersionIsUnsupported(inferenceendpoint.MinSupportedVersion)
	if skip, err := skipFunc(); err != nil {
		t.Fatalf("failed to check version: %v", err)
	} else if skip {
		t.Skipf("Test requires Elasticsearch v%s or higher", inferenceendpoint.MinSupportedVersion)
	}

	inferenceID := fmt.Sprintf("test-inference-%s", sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); skipValidateAndStart(t) },
		CheckDestroy: checkInferenceEndpointDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"inference_id": config.StringVariable(inferenceID),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_inference_endpoint.test", "inference_id", inferenceID),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_inference_endpoint.test", "task_type", "completion"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_inference_endpoint.test", "service", "openai"),
				),
			},
			// Changing task_type must trigger a replacement, not an in-place update.
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("different_task_type"),
				ConfigVariables: config.Variables{
					"inference_id": config.StringVariable(inferenceID),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(
							"elasticstack_elasticsearch_inference_endpoint.test",
							plancheck.ResourceActionReplace,
						),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_inference_endpoint.test", "inference_id", inferenceID),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_inference_endpoint.test", "task_type", "chat_completion"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_inference_endpoint.test", "service", "openai"),
				),
			},
		},
	})
}

func checkInferenceEndpointDestroy(s *terraform.State) error {
	client, err := clients.NewAcceptanceTestingClient()
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "elasticstack_elasticsearch_inference_endpoint" {
			continue
		}

		compID, _ := clients.CompositeIDFromStr(rs.Primary.ID)

		esClient, err := client.GetESClient()
		if err != nil {
			return err
		}

		res, err := esClient.InferenceGet(
			esClient.InferenceGet.WithInferenceID(compID.ResourceID),
		)
		if err != nil {
			return err
		}
		defer res.Body.Close()

		if res.StatusCode != http.StatusNotFound {
			return fmt.Errorf("inference endpoint (%s) still exists", compID.ResourceID)
		}
	}
	return nil
}
