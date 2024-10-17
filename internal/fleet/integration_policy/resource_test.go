package integration_policy_test

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/require"
)

var minVersionIntegrationPolicy = version.Must(version.NewVersion("8.10.0"))

func TestJsonTypes(t *testing.T) {
	mapBytes, err := json.Marshal(map[string]string{})
	require.NoError(t, err)
	equal, diags := jsontypes.NewNormalizedValue(`{"a": "b"}`).StringSemanticEquals(context.Background(), jsontypes.NewNormalizedValue(string(mapBytes)))
	require.Empty(t, diags)
	require.False(t, equal)
}

func TestAccResourceIntegrationPolicy(t *testing.T) {
	policyName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		CheckDestroy:             checkResourceIntegrationPolicyDestroy,
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minVersionIntegrationPolicy),
				Config:   testAccResourceIntegrationPolicyCreate(policyName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "name", policyName),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "description", "IntegrationPolicyTest Policy"),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "integration_name", "tcp"),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "integration_version", "1.16.0"),
					resource.TestCheckNoResourceAttr("elasticstack_fleet_integration_policy.test_policy", "vars_json"),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "input.0.input_id", "tcp-tcp"),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "input.0.enabled", "true"),
					resource.TestCheckNoResourceAttr("elasticstack_fleet_integration_policy.test_policy", "input.0.vars_json"),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "input.0.streams_json", `{"tcp.generic":{"enabled":true,"vars":{"custom":"","data_stream.dataset":"tcp.generic","listen_address":"localhost","listen_port":8080,"ssl":"","syslog_options":"field: message","tags":[]}}}`),
				),
			},
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minVersionIntegrationPolicy),
				Config:   testAccResourceIntegrationPolicyUpdate(policyName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "name", policyName),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "description", "Updated Integration Policy"),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "integration_name", "tcp"),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "integration_version", "1.16.0"),
					resource.TestCheckNoResourceAttr("elasticstack_fleet_integration_policy.test_policy", "vars_json"),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "input.0.input_id", "tcp-tcp"),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "input.0.enabled", "false"),
					resource.TestCheckNoResourceAttr("elasticstack_fleet_integration_policy.test_policy", "input.0.vars_json"),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "input.0.streams_json", `{"tcp.generic":{"enabled":false,"vars":{"custom":"","data_stream.dataset":"tcp.generic","listen_address":"localhost","listen_port":8085,"ssl":"","syslog_options":"field: message","tags":[]}}}`),
				),
			},
		},
	})
}

func TestAccResourceIntegrationPolicySecretsFromSDK(t *testing.T) {
	policyName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceIntegrationPolicyDestroy,
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"elasticstack": {
						Source:            "elastic/elasticstack",
						VersionConstraint: "0.11.7",
					},
				},
				SkipFunc:           versionutils.CheckIfVersionIsUnsupported(minVersionIntegrationPolicy),
				Config:             testAccResourceIntegrationPolicySecretsCreate(policyName, "created"),
				ExpectNonEmptyPlan: true, // secret churn
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "name", policyName),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "description", "IntegrationPolicyTest Policy"),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "integration_name", "aws_logs"),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "integration_version", "1.4.0"),
					resource.TestMatchResourceAttr("elasticstack_fleet_integration_policy.test_policy", "vars_json", regexp.MustCompile(`{"access_key_id":{"id":"\S+","isSecretRef":true},"default_region":"us-east-1","endpoint":"endpoint","secret_access_key":{"id":"\S+","isSecretRef":true},"session_token":{"id":"\S+","isSecretRef":true}}`)),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "input.0.input_id", "aws_logs-aws-cloudwatch"),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "input.0.enabled", "true"),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "input.0.vars_json", ""),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "input.0.streams_json", `{"aws_logs.generic":{"enabled":true,"vars":{"api_sleep":"200ms","api_timeput":"120s","custom":"","data_stream.dataset":"aws_logs.generic","log_streams":[],"number_of_workers":1,"preserve_original_event":false,"scan_frequency":"1m","start_position":"beginning","tags":["forwarded"]}}}`),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minVersionIntegrationPolicy),
				Config:                   testAccResourceIntegrationPolicySecretsCreate(policyName, "created"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "name", policyName),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "description", "IntegrationPolicyTest Policy"),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "integration_name", "aws_logs"),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "integration_version", "1.4.0"),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "vars_json", fmt.Sprintf(`{"access_key_id":"placeholder","default_region":"us-east-1","endpoint":"endpoint","secret_access_key":"created %s","session_token":"placeholder"}`, policyName)),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "input.0.input_id", "aws_logs-aws-cloudwatch"),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "input.0.enabled", "true"),
					resource.TestCheckNoResourceAttr("elasticstack_fleet_integration_policy.test_policy", "input.0.vars_json"),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "input.0.streams_json", `{"aws_logs.generic":{"enabled":true,"vars":{"api_sleep":"200ms","api_timeput":"120s","custom":"","data_stream.dataset":"aws_logs.generic","log_streams":[],"number_of_workers":1,"preserve_original_event":false,"scan_frequency":"1m","start_position":"beginning","tags":["forwarded"]}}}`),
				),
			},
		},
	})
}

func TestAccResourceIntegrationPolicySecrets(t *testing.T) {
	policyName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		CheckDestroy:             checkResourceIntegrationPolicyDestroy,
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minVersionIntegrationPolicy),
				Config:   testAccResourceIntegrationPolicySecretsCreate(policyName, "created"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "name", policyName),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "description", "IntegrationPolicyTest Policy"),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "integration_name", "aws_logs"),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "integration_version", "1.4.0"),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "vars_json", fmt.Sprintf(`{"access_key_id":"placeholder","default_region":"us-east-1","endpoint":"endpoint","secret_access_key":"created %s","session_token":"placeholder"}`, policyName)),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "input.0.input_id", "aws_logs-aws-cloudwatch"),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "input.0.enabled", "true"),
					resource.TestCheckNoResourceAttr("elasticstack_fleet_integration_policy.test_policy", "input.0.vars_json"),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "input.0.streams_json", `{"aws_logs.generic":{"enabled":true,"vars":{"api_sleep":"200ms","api_timeput":"120s","custom":"","data_stream.dataset":"aws_logs.generic","log_streams":[],"number_of_workers":1,"preserve_original_event":false,"scan_frequency":"1m","start_position":"beginning","tags":["forwarded"]}}}`),
				),
			},
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minVersionIntegrationPolicy),
				Config:   testAccResourceIntegrationPolicySecretsUpdate(policyName, "updated"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "name", policyName),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "description", "Updated Integration Policy"),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "integration_name", "aws_logs"),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "integration_version", "1.4.0"),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "vars_json", fmt.Sprintf(`{"access_key_id":"placeholder","default_region":"us-east-2","endpoint":"endpoint","secret_access_key":"updated %s","session_token":"placeholder"}`, policyName)),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "input.0.input_id", "aws_logs-aws-cloudwatch"),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "input.0.enabled", "false"),
					resource.TestCheckNoResourceAttr("elasticstack_fleet_integration_policy.test_policy", "input.0.vars_json"),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "input.0.streams_json", `{"aws_logs.generic":{"enabled":false,"vars":{"api_sleep":"200ms","api_timeput":"120s","custom":"","data_stream.dataset":"aws_logs.generic","log_streams":[],"number_of_workers":1,"preserve_original_event":false,"scan_frequency":"2m","start_position":"beginning","tags":["forwarded"]}}}`),
				),
			},
			{
				SkipFunc:                versionutils.CheckIfVersionIsUnsupported(minVersionIntegrationPolicy),
				ResourceName:            "elasticstack_fleet_integration_policy.test_policy",
				Config:                  testAccResourceIntegrationPolicyUpdate(policyName),
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"vars_json"},
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr("elasticstack_fleet_integration_policy.test_policy", "vars_json", regexp.MustCompile(`{"access_key_id":{"id":"\S+","isSecretRef":true},"default_region":"us-east-2","endpoint":"endpoint","secret_access_key":{"id":"\S+","isSecretRef":true},"session_token":{"id":"\S+","isSecretRef":true}}`)),
				),
			},
		},
	})
}

func checkResourceIntegrationPolicyDestroy(s *terraform.State) error {
	client, err := clients.NewAcceptanceTestingClient()
	if err != nil {
		return err
	}

	fleetClient, err := client.GetFleetClient()
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		switch rs.Type {
		case "elasticstack_fleet_agent_policy":
			policy, diags := fleet.GetAgentPolicy(context.Background(), fleetClient, rs.Primary.ID)
			if diags.HasError() {
				return utils.FwDiagsAsError(diags)
			}
			if policy != nil {
				return fmt.Errorf("agent policy id=%v still exists, but it should have been removed", rs.Primary.ID)
			}
		case "elasticstack_fleet_integration_policy":
			policy, diags := fleet.GetPackagePolicy(context.Background(), fleetClient, rs.Primary.ID)
			if diags.HasError() {
				return utils.FwDiagsAsError(diags)
			}
			if policy != nil {
				return fmt.Errorf("integration policy id=%v still exists, but it should have been removed", rs.Primary.ID)
			}
		default:
			continue
		}

	}
	return nil
}

func testAccResourceIntegrationPolicyCommon(name string, integrationName string, integrationVersion string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_fleet_integration" "test_policy" {
  name    = "%s"
  version = "%s"
  force   = true
}

resource "elasticstack_fleet_agent_policy" "test_policy" {
  name            = "%s Agent Policy"
  namespace       = "default"
  description     = "IntegrationPolicyTest Agent Policy"
  monitor_logs    = true
  monitor_metrics = true
  skip_destroy    = false
}

data "elasticstack_fleet_enrollment_tokens" "test_policy" {
  policy_id = elasticstack_fleet_agent_policy.test_policy.policy_id
}
`, integrationName, integrationVersion, name)
}

func testAccResourceIntegrationPolicyCreate(id string) string {
	common := testAccResourceIntegrationPolicyCommon(id, "tcp", "1.16.0")
	return fmt.Sprintf(`
%s

resource "elasticstack_fleet_integration_policy" "test_policy" {
  name            = "%s"
  namespace       = "default"
  description     = "IntegrationPolicyTest Policy"
  agent_policy_id = elasticstack_fleet_agent_policy.test_policy.policy_id
  integration_name    = elasticstack_fleet_integration.test_policy.name
  integration_version = elasticstack_fleet_integration.test_policy.version

  input {
    input_id = "tcp-tcp"
    enabled = true
    streams_json = jsonencode({
      "tcp.generic": {
        "enabled": true
        "vars": {
          "listen_address": "localhost"
          "listen_port": 8080
          "data_stream.dataset": "tcp.generic"
          "tags": []
          "syslog_options": "field: message"
          "ssl": ""
          "custom": ""
        }
      }
    })
  }
}
`, common, id)
}

func testAccResourceIntegrationPolicyUpdate(id string) string {
	common := testAccResourceIntegrationPolicyCommon(id, "tcp", "1.16.0")
	return fmt.Sprintf(`
%s

resource "elasticstack_fleet_integration_policy" "test_policy" {
  name            = "%s"
  namespace       = "default"
  description     = "Updated Integration Policy"
  agent_policy_id = elasticstack_fleet_agent_policy.test_policy.policy_id
  integration_name    = elasticstack_fleet_integration.test_policy.name
  integration_version = elasticstack_fleet_integration.test_policy.version

  input {
    input_id = "tcp-tcp"
    enabled  = false
    streams_json = jsonencode({
      "tcp.generic": {
        "enabled": false
        "vars": {
          "listen_address": "localhost"
          "listen_port": 8085
          "data_stream.dataset": "tcp.generic"
          "tags": []
          "syslog_options": "field: message"
          "ssl": ""
          "custom": ""
        }
      }
    })
  }
}
`, common, id)
}

func testAccResourceIntegrationPolicySecretsCreate(id string, key string) string {
	common := testAccResourceIntegrationPolicyCommon(id, "aws_logs", "1.4.0")
	return fmt.Sprintf(`
%s

resource "elasticstack_fleet_integration_policy" "test_policy" {
  name            = "%s"
  namespace       = "default"
  description     = "IntegrationPolicyTest Policy"
  agent_policy_id = elasticstack_fleet_agent_policy.test_policy.policy_id
  integration_name    = elasticstack_fleet_integration.test_policy.name
  integration_version = elasticstack_fleet_integration.test_policy.version

  vars_json = jsonencode({
    "access_key_id": "placeholder"
    "secret_access_key": "%s %s"
    "session_token": "placeholder"
    "endpoint": "endpoint"
    "default_region": "us-east-1"
  })
  input {
    input_id = "aws_logs-aws-cloudwatch"
    enabled  = true
    streams_json = jsonencode({
      "aws_logs.generic" = {
        enabled = true
        vars = {
          "number_of_workers": 1
          "log_streams": []
          "start_position": "beginning"
          "scan_frequency": "1m"
          "api_timeput": "120s"
          "api_sleep": "200ms"
          "tags": ["forwarded"]
          "preserve_original_event": false
          "data_stream.dataset": "aws_logs.generic"
          "custom": ""
        }
      }
    })
  }
  input {
    input_id = "aws_logs-aws-s3"
    enabled = true
    streams_json = jsonencode({
      "aws_logs.generic" = {
        enabled = true
        vars = {
          "number_of_workers": 1
          "bucket_list_interval": "120s"
          "file_selectors": ""
          "fips_enabled": false
          "include_s3_metadata": []
          "max_bytes": "10MiB"
          "max_number_of_messages": 5
          "parsers": ""
          "sqs.max_receive_count": 5
          "sqs.wait_time": "20s"
          "tags": ["forwarded"]
          "preserve_original_event": false
          "data_stream.dataset": "aws_logs.generic"
          "custom": ""
        }
      }
    })
  }
}
`, common, id, key, id)
}

func testAccResourceIntegrationPolicySecretsUpdate(id string, key string) string {
	common := testAccResourceIntegrationPolicyCommon(id, "aws_logs", "1.4.0")
	return fmt.Sprintf(`
%s

resource "elasticstack_fleet_integration_policy" "test_policy" {
  name            = "%s"
  namespace       = "default"
  description     = "Updated Integration Policy"
  agent_policy_id = elasticstack_fleet_agent_policy.test_policy.policy_id
  integration_name    = elasticstack_fleet_integration.test_policy.name
  integration_version = elasticstack_fleet_integration.test_policy.version

  vars_json = jsonencode({
    "access_key_id": "placeholder"
    "secret_access_key": "%s %s"
    "session_token": "placeholder"
    "endpoint": "endpoint"
    "default_region": "us-east-2"
  })
  input {
    input_id = "aws_logs-aws-cloudwatch"
    enabled  = false
    streams_json = jsonencode({
      "aws_logs.generic" = {
        enabled = false
        vars = {
          "number_of_workers": 1,
          "log_streams": [],
          "start_position": "beginning",
          "scan_frequency": "2m",
          "api_timeput": "120s",
          "api_sleep": "200ms",
          "tags": ["forwarded"],
          "preserve_original_event": false,
          "data_stream.dataset": "aws_logs.generic",
          "custom": "",
        }
      }
    })
  }
  input {
    input_id = "aws_logs-aws-s3"
    enabled = false
    streams_json = jsonencode({
      "aws_logs.generic" = {
        enabled = false
        vars = {
          "number_of_workers": 1,
          "bucket_list_interval": "120s",
          "file_selectors": "",
          "fips_enabled": false,
          "include_s3_metadata": [],
          "max_bytes": "20MiB",
          "max_number_of_messages": 5,
          "parsers": "",
          "sqs.max_receive_count": 5,
          "sqs.wait_time": "20s",
          "tags": ["forwarded"],
          "preserve_original_event": false,
          "data_stream.dataset": "aws_logs.generic",
          "custom": "",
        }
      }
    })
  }
}
`, common, id, key, id)
}
