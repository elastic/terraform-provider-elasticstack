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
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	esclient "github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/inference/inferenceendpoint"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

const (
	defaultInferenceEndpointAPIKey = "test-openai-api-key"
	inferenceEndpointAPIKeyEnvVar  = "TF_ELASTICSTACK_TEST_OPENAI_API_KEY"
)

func inferenceEndpointTestAPIKey() (string, bool) {
	apiKey, ok := os.LookupEnv(inferenceEndpointAPIKeyEnvVar)
	if ok && apiKey != "" {
		return apiKey, false
	}

	return defaultInferenceEndpointAPIKey, true
}

func inferenceEndpointConfigVariables(inferenceID, apiKey string) config.Variables {
	return config.Variables{
		"inference_id": config.StringVariable(inferenceID),
		"api_key":      config.StringVariable(apiKey),
	}
}

func skipWhenUsingFakeInferenceEndpointAPIKey(usingFakeAPIKey bool) func() (bool, error) {
	return func() (bool, error) {
		return usingFakeAPIKey, nil
	}
}

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
	apiKey, usingFakeAPIKey := inferenceEndpointTestAPIKey()
	skipIfUsingFakeAPIKey := skipWhenUsingFakeInferenceEndpointAPIKey(usingFakeAPIKey)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); skipValidateAndStart(t) },
		CheckDestroy: checkInferenceEndpointDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          inferenceEndpointConfigVariables(inferenceID, apiKey),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_inference_endpoint.test", "inference_id", inferenceID),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_inference_endpoint.test", "task_type", "completion"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_inference_endpoint.test", "service", "openai"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables:          inferenceEndpointConfigVariables(inferenceID, apiKey),
				SkipFunc:                 skipIfUsingFakeAPIKey,
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
				ConfigVariables:          inferenceEndpointConfigVariables(inferenceID, apiKey),
				SkipFunc:                 skipIfUsingFakeAPIKey,
				PlanOnly:                 true,
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables:          inferenceEndpointConfigVariables(inferenceID, apiKey),
				SkipFunc:                 skipIfUsingFakeAPIKey,
				ResourceName:             "elasticstack_elasticsearch_inference_endpoint.test",
				ImportState:              true,
				ImportStateVerify:        true,
				ImportStateVerifyIgnore:  []string{"service_settings"},
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
	apiKey, _ := inferenceEndpointTestAPIKey()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); skipValidateAndStart(t) },
		CheckDestroy: checkInferenceEndpointDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          inferenceEndpointConfigVariables(inferenceID, apiKey),
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
				ConfigVariables:          inferenceEndpointConfigVariables(inferenceID, apiKey),
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

// TestAccResourceInferenceEndpointTaskSettingsNoDrift verifies that server-applied
// defaults returned in task_settings do not cause perpetual plan drift.
func TestAccResourceInferenceEndpointTaskSettingsNoDrift(t *testing.T) {
	skipFunc := versionutils.CheckIfVersionIsUnsupported(inferenceendpoint.MinSupportedVersion)
	if skip, err := skipFunc(); err != nil {
		t.Fatalf("failed to check version: %v", err)
	} else if skip {
		t.Skipf("Test requires Elasticsearch v%s or higher", inferenceendpoint.MinSupportedVersion)
	}

	inferenceID := fmt.Sprintf("test-inference-%s", sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum))
	apiKey, _ := inferenceEndpointTestAPIKey()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); skipValidateAndStart(t) },
		CheckDestroy: checkInferenceEndpointDestroy,
		Steps: []resource.TestStep{
			// Create with task_settings containing a subset of keys.
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("with_task_settings"),
				ConfigVariables:          inferenceEndpointConfigVariables(inferenceID, apiKey),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_inference_endpoint.test", "inference_id", inferenceID),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_inference_endpoint.test", "task_type", "completion"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_inference_endpoint.test", "service", "openai"),
				),
			},
			// Re-plan with the same config must produce no diff — server-applied defaults
			// in the API response must not leak into state and cause perpetual drift.
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("with_task_settings"),
				ConfigVariables:          inferenceEndpointConfigVariables(inferenceID, apiKey),
				PlanOnly:                 true,
				ExpectNonEmptyPlan:       false,
			},
			// Update to remove task_settings entirely — must not drift either.
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("without_task_settings"),
				ConfigVariables:          inferenceEndpointConfigVariables(inferenceID, apiKey),
				Check:                    resource.TestCheckNoResourceAttr("elasticstack_elasticsearch_inference_endpoint.test", "task_settings"),
			},
			// Re-plan after removing task_settings — still no drift.
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("without_task_settings"),
				ConfigVariables:          inferenceEndpointConfigVariables(inferenceID, apiKey),
				PlanOnly:                 true,
				ExpectNonEmptyPlan:       false,
			},
		},
	})
}

// TestAccResourceInferenceEndpointTaskSettingsDrift verifies two complementary behaviours:
//  1. When task_settings is omitted, server-applied defaults in the API response do not cause drift.
//  2. When task_settings is explicitly set to a value that differs from the server default,
//     Terraform correctly surfaces that as a real change.
func TestAccResourceInferenceEndpointTaskSettingsDrift(t *testing.T) {
	skipFunc := versionutils.CheckIfVersionIsUnsupported(inferenceendpoint.MinSupportedVersion)
	if skip, err := skipFunc(); err != nil {
		t.Fatalf("failed to check version: %v", err)
	} else if skip {
		t.Skipf("Test requires Elasticsearch v%s or higher", inferenceendpoint.MinSupportedVersion)
	}

	inferenceID := fmt.Sprintf("test-inference-%s", sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum))
	apiKey, usingFakeAPIKey := inferenceEndpointTestAPIKey()
	skipIfUsingFakeAPIKey := skipWhenUsingFakeInferenceEndpointAPIKey(usingFakeAPIKey)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); skipValidateAndStart(t) },
		CheckDestroy: checkInferenceEndpointDestroy,
		Steps: []resource.TestStep{
			// Create without task_settings. The azureaistudio service returns
			// {max_new_tokens: 64} by default — that must not cause drift.
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("no_task_settings"),
				ConfigVariables:          inferenceEndpointConfigVariables(inferenceID, apiKey),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_inference_endpoint.test", "service", "azureaistudio"),
					resource.TestCheckNoResourceAttr("elasticstack_elasticsearch_inference_endpoint.test", "task_settings"),
				),
			},
			// Re-plan with no task_settings — must be empty (server default not leaked into state).
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("no_task_settings"),
				ConfigVariables:          inferenceEndpointConfigVariables(inferenceID, apiKey),
				PlanOnly:                 true,
				ExpectNonEmptyPlan:       false,
			},
			// Now explicitly set max_new_tokens to a non-default value.
			// This is a real user-driven change and must be applied.
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("with_task_settings"),
				ConfigVariables:          inferenceEndpointConfigVariables(inferenceID, apiKey),
				SkipFunc:                 skipIfUsingFakeAPIKey,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrWith(
						"elasticstack_elasticsearch_inference_endpoint.test",
						"task_settings",
						func(value string) error {
							var ts map[string]any
							if err := json.Unmarshal([]byte(value), &ts); err != nil {
								return err
							}
							if ts["max_new_tokens"] != float64(32) {
								return fmt.Errorf("expected max_new_tokens=32, got %v", ts["max_new_tokens"])
							}
							return nil
						},
					),
				),
			},
			// Re-plan after applying explicit task_settings — no drift.
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("with_task_settings"),
				ConfigVariables:          inferenceEndpointConfigVariables(inferenceID, apiKey),
				SkipFunc:                 skipIfUsingFakeAPIKey,
				PlanOnly:                 true,
				ExpectNonEmptyPlan:       false,
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
