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

package integrationpolicy_test

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

const (
	integrationPolicyResourceAddress = "elasticstack_fleet_integration_policy.test_policy"
	outOfBandConfigKey               = "tf_acc_out_of_band"
)

// readTypedPolicyInputs reads the integration policy referenced by the given
// resource address in the typed (non-simplified) format and returns its inputs
// as generic maps, so that input and stream config can be inspected.
func readTypedPolicyInputs(s *terraform.State, resourceAddress string) (*kbapi.PackagePolicy, []map[string]any, error) {
	rs, ok := s.RootModule().Resources[resourceAddress]
	if !ok {
		return nil, nil, fmt.Errorf("resource %q not found in state", resourceAddress)
	}
	apiClient, err := clients.NewAcceptanceTestingKibanaScopedClient()
	if err != nil {
		return nil, nil, err
	}
	policyID := rs.Primary.Attributes["policy_id"]
	policy, diags := fleet.GetPackagePolicyTyped(context.Background(), apiClient.GetFleetClient(), policyID, "")
	if diags.HasError() || policy == nil {
		return nil, nil, fmt.Errorf("error reading policy %q: %v", policyID, diags)
	}
	data, err := json.Marshal(policy.Inputs)
	if err != nil {
		return nil, nil, err
	}
	var inputs []map[string]any
	if err := json.Unmarshal(data, &inputs); err != nil {
		return nil, nil, err
	}
	return policy, inputs, nil
}

// setInputConfigOutOfBand writes a marker into the input-level config and the
// first stream's config of every input through the typed Fleet API, the way
// Kibana features such as APM or Osquery store data the resource does not
// manage.
func setInputConfigOutOfBand(resourceAddress string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		policy, inputs, err := readTypedPolicyInputs(s, resourceAddress)
		if err != nil {
			return err
		}
		for _, input := range inputs {
			delete(input, "compiled_input")
			cfg, _ := input["config"].(map[string]any)
			if cfg == nil {
				cfg = map[string]any{}
			}
			cfg[outOfBandConfigKey] = map[string]any{"value": map[string]any{"nested": map[string]any{"secret": "input-marker"}}}
			input["config"] = cfg
			streams, _ := input["streams"].([]any)
			for i, st := range streams {
				stream, _ := st.(map[string]any)
				delete(stream, "compiled_stream")
				if i == 0 {
					stream["config"] = map[string]any{outOfBandConfigKey: map[string]any{"value": "stream-marker"}}
				}
			}
		}

		data, err := json.Marshal(inputs)
		if err != nil {
			return err
		}
		var requestInputs []kbapi.PackagePolicyRequestTypedInput
		if err := json.Unmarshal(data, &requestInputs); err != nil {
			return err
		}
		apiClient, err := clients.NewAcceptanceTestingKibanaScopedClient()
		if err != nil {
			return err
		}
		req := kbapi.PackagePolicyUpdateRequestTypedInputs{Inputs: &requestInputs, Version: policy.Version}
		if _, diags := fleet.UpdatePackagePolicyTyped(context.Background(), apiClient.GetFleetClient(), policy.Id, "", req); diags.HasError() {
			return fmt.Errorf("error writing input config out of band: %v", diags)
		}
		return nil
	}
}

// checkOutOfBandInputConfig asserts the marker written by
// setInputConfigOutOfBand is still present on the server.
func checkOutOfBandInputConfig(resourceAddress string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, inputs, err := readTypedPolicyInputs(s, resourceAddress)
		if err != nil {
			return err
		}
		if len(inputs) == 0 {
			return fmt.Errorf("policy has no inputs")
		}
		for _, input := range inputs {
			cfg, _ := input["config"].(map[string]any)
			marker, _ := cfg[outOfBandConfigKey].(map[string]any)
			value, _ := marker["value"].(map[string]any)
			nested, _ := value["nested"].(map[string]any)
			if nested["secret"] != "input-marker" {
				return fmt.Errorf("input %v lost its out-of-band config key %q", input["type"], outOfBandConfigKey)
			}
			streams, _ := input["streams"].([]any)
			if len(streams) == 0 {
				continue
			}
			stream, _ := streams[0].(map[string]any)
			streamCfg, _ := stream["config"].(map[string]any)
			streamMarker, _ := streamCfg[outOfBandConfigKey].(map[string]any)
			if streamMarker["value"] != "stream-marker" {
				return fmt.Errorf("stream %v lost its out-of-band config key %q", stream["data_stream"], outOfBandConfigKey)
			}
		}
		return nil
	}
}

// TestAccResourceIntegrationPolicyPreservesInputConfig verifies that input and
// stream config written outside Terraform survives an update of the policy.
func TestAccResourceIntegrationPolicyPreservesInputConfig(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minVersionIntegrationPolicy, versionutils.FlavorAny)

	policyName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)
	vars := func() config.Variables {
		return config.Variables{"policy_name": config.StringVariable(policyName)}
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceIntegrationPolicyDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          vars(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(integrationPolicyResourceAddress, "id"),
					setInputConfigOutOfBand(integrationPolicyResourceAddress),
					checkOutOfBandInputConfig(integrationPolicyResourceAddress),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables:          vars(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(integrationPolicyResourceAddress, "description", "IntegrationPolicyTest Policy updated"),
					resource.TestCheckResourceAttr(integrationPolicyResourceAddress, "inputs.tcp-tcp.streams.tcp.generic.vars", tcpGenericVarsExpected8085),
					checkOutOfBandInputConfig(integrationPolicyResourceAddress),
				),
			},
		},
	})
}

// checkAPMAPIKeysInConfig asserts the API keys Kibana's APM plugin stores in
// the APM input config are present on the server.
func checkAPMAPIKeysInConfig(resourceAddress string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, inputs, err := readTypedPolicyInputs(s, resourceAddress)
		if err != nil {
			return err
		}
		for _, input := range inputs {
			if input["type"] != "apm" {
				continue
			}
			cfg, _ := input["config"].(map[string]any)
			apmServer, _ := cfg["apm-server"].(map[string]any)
			value, _ := apmServer["value"].(map[string]any)
			for _, path := range [][]string{{"rum", "source_mapping", "elasticsearch"}, {"agent", "config", "elasticsearch"}} {
				node := value
				for _, key := range path {
					node, _ = node[key].(map[string]any)
				}
				// Never include config values in the message: they hold API keys.
				if key, _ := node["api_key"].(string); key == "" {
					return fmt.Errorf("APM input config is missing apm-server.value.%s.api_key", strings.Join(path, "."))
				}
			}
			return nil
		}
		return fmt.Errorf("policy has no apm input")
	}
}

// TestAccResourceIntegrationPolicyAPMPreservesAPIKeys covers the reported
// impact: Kibana's APM plugin stores the source map and agent configuration
// API keys in the APM input config, and an update of the policy must not
// remove them.
func TestAccResourceIntegrationPolicyAPMPreservesAPIKeys(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minVersionIntegrationPolicy, versionutils.FlavorAny)

	policyName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)
	vars := func() config.Variables {
		return config.Variables{"policy_name": config.StringVariable(policyName)}
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceIntegrationPolicyDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          vars(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(integrationPolicyResourceAddress, "id"),
					checkAPMAPIKeysInConfig(integrationPolicyResourceAddress),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables:          vars(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(integrationPolicyResourceAddress, "id"),
					checkAPMAPIKeysInConfig(integrationPolicyResourceAddress),
				),
			},
		},
	})
}
