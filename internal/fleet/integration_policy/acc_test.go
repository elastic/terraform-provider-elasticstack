package integration_policy_test

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"regexp"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/stretchr/testify/require"
)

var (
	minVersionIntegrationPolicy    = version.Must(version.NewVersion("8.10.0"))
	minVersionIntegrationPolicyIds = version.Must(version.NewVersion("8.15.0"))
	minVersionOutputId             = version.Must(version.NewVersion("8.16.0"))
	minVersionSqlIntegration       = version.Must(version.NewVersion("9.1.0"))
)

func TestJsonTypes(t *testing.T) {
	mapBytes, err := json.Marshal(map[string]string{})
	require.NoError(t, err)
	equal, diags := jsontypes.NewNormalizedValue(`{"a": "b"}`).StringSemanticEquals(context.Background(), jsontypes.NewNormalizedValue(string(mapBytes)))
	require.Empty(t, diags)
	require.False(t, equal)
}

func TestAccResourceIntegrationPolicyMultipleAgentPolicies(t *testing.T) {
	policyName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceIntegrationPolicyDestroy,
		Steps: []resource.TestStep{
			{
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minVersionIntegrationPolicyIds),
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"policy_name": config.StringVariable(policyName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "name", policyName),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "description", "IntegrationPolicyTest Policy"),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "integration_name", "tcp"),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "integration_version", "1.16.0"),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "agent_policy_ids.#", "2"),
				),
			},
		},
	})
}

func TestAccResourceIntegrationPolicyWithOutput(t *testing.T) {
	policyName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceIntegrationPolicyDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minVersionOutputId),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"policy_name": config.StringVariable(policyName),
					"output_name": config.StringVariable(fmt.Sprintf("Test Output %s", policyName)),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "name", policyName),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "description", "IntegrationPolicyTest Policy with Output"),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "integration_name", "tcp"),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "integration_version", "1.16.0"),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "output_id", fmt.Sprintf("%s-test-output", policyName)),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "inputs.tcp-tcp.enabled", "true"),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "inputs.tcp-tcp.streams.tcp.generic.enabled", "true"),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "inputs.tcp-tcp.streams.tcp.generic.vars", `{"custom":"","data_stream.dataset":"tcp.generic","listen_address":"localhost","listen_port":8080,"ssl":"","syslog_options":"field: message","tags":[]}`),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minVersionOutputId),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables: config.Variables{
					"policy_name":         config.StringVariable(policyName),
					"output_name":         config.StringVariable(fmt.Sprintf("Test Output %s", policyName)),
					"updated_output_name": config.StringVariable(fmt.Sprintf("Updated Test Output %s", policyName)),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "name", policyName),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "description", "Updated Integration Policy with Output"),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "integration_name", "tcp"),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "integration_version", "1.16.0"),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "output_id", fmt.Sprintf("%s-updated-output", policyName)),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "inputs.tcp-tcp.enabled", "true"),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "inputs.tcp-tcp.streams.tcp.generic.enabled", "true"),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "inputs.tcp-tcp.streams.tcp.generic.vars", `{"custom":"","data_stream.dataset":"tcp.generic","listen_address":"localhost","listen_port":8080,"ssl":"","syslog_options":"field: message","tags":[]}`),
				),
			},
		},
	})
}

func TestAccResourceIntegrationPolicy(t *testing.T) {
	policyName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceIntegrationPolicyDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minVersionIntegrationPolicy),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"policy_name": config.StringVariable(policyName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "name", policyName),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "description", "IntegrationPolicyTest Policy"),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "integration_name", "tcp"),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "integration_version", "1.16.0"),
					resource.TestCheckNoResourceAttr("elasticstack_fleet_integration_policy.test_policy", "vars_json"),
					resource.TestCheckNoResourceAttr("elasticstack_fleet_integration_policy.test_policy", "output_id"),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "inputs.tcp-tcp.enabled", "true"),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "inputs.tcp-tcp.streams.tcp.generic.enabled", "true"),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "inputs.tcp-tcp.streams.tcp.generic.vars", `{"custom":"","data_stream.dataset":"tcp.generic","listen_address":"localhost","listen_port":8080,"ssl":"","syslog_options":"field: message","tags":[]}`),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minVersionIntegrationPolicy),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables: config.Variables{
					"policy_name": config.StringVariable(policyName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "name", policyName),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "description", "Updated Integration Policy"),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "integration_name", "tcp"),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "integration_version", "1.16.0"),
					resource.TestCheckNoResourceAttr("elasticstack_fleet_integration_policy.test_policy", "vars_json"),
					resource.TestCheckNoResourceAttr("elasticstack_fleet_integration_policy.test_policy", "output_id"),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "inputs.tcp-tcp.enabled", "false"),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "inputs.tcp-tcp.streams.tcp.generic.enabled", "false"),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "inputs.tcp-tcp.streams.tcp.generic.vars", `{"custom":"","data_stream.dataset":"tcp.generic","listen_address":"localhost","listen_port":8085,"ssl":"","syslog_options":"field: message","tags":[]}`),
				),
			},
		},
	})
}

//go:embed testdata/TestAccResourceIntegrationPolicySecretsFromSDK/legacy/integration_policy.tf
var sdkCreateTestConfig string

func TestAccResourceIntegrationPolicySecretsFromSDK(t *testing.T) {
	policyName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)

	sdkConstrains, err := version.NewConstraint(">=8.10.0,<8.16.0")
	require.NoError(t, err)

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
				SkipFunc: versionutils.CheckIfVersionMeetsConstraints(sdkConstrains),
				Config:   sdkCreateTestConfig,
				ConfigVariables: config.Variables{
					"policy_name": config.StringVariable(policyName),
					"secret_key":  config.StringVariable("created"),
				},
				ExpectNonEmptyPlan: true, // secret churn
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "name", policyName),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "description", "IntegrationPolicyTest Policy"),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "integration_name", "aws_logs"),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "integration_version", "1.4.0"),
					resource.TestMatchResourceAttr("elasticstack_fleet_integration_policy.test_policy", "vars_json", regexp.MustCompile(`{"access_key_id":{"id":"\S+","isSecretRef":true},"default_region":"us-east-1","endpoint":"endpoint","secret_access_key":{"id":"\S+","isSecretRef":true},"session_token":{"id":"\S+","isSecretRef":true}}`)),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "input.0.streams_json", `{"aws_logs.generic":{"enabled":true,"vars":{"api_sleep":"200ms","api_timeput":"120s","custom":"","data_stream.dataset":"aws_logs.generic","log_streams":[],"number_of_workers":1,"preserve_original_event":false,"scan_frequency":"1m","start_position":"beginning","tags":["forwarded"]}}}`),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionMeetsConstraints(sdkConstrains),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("current"),
				ConfigVariables: config.Variables{
					"policy_name": config.StringVariable(policyName),
					"secret_key":  config.StringVariable("created"),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "name", policyName),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "description", "IntegrationPolicyTest Policy"),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "integration_name", "aws_logs"),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "integration_version", "1.4.0"),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "vars_json", fmt.Sprintf(`{"access_key_id":"placeholder","default_region":"us-east-1","endpoint":"endpoint","secret_access_key":"created %s","session_token":"placeholder"}`, policyName)),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "inputs.aws_logs-aws-cloudwatch.enabled", "true"),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "inputs.aws_logs-aws-cloudwatch.streams.aws_logs.generic.enabled", "true"),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "inputs.aws_logs-aws-cloudwatch.streams.aws_logs.generic.vars", `{"api_sleep":"200ms","api_timeput":"120s","custom":"","data_stream.dataset":"aws_logs.generic","log_streams":[],"number_of_workers":1,"preserve_original_event":false,"scan_frequency":"1m","start_position":"beginning","tags":["forwarded"]}`),
				),
			},
		},
	})
}

func TestAccResourceIntegrationPolicySecrets(t *testing.T) {
	policyName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)

	t.Run("single valued secrets", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			PreCheck:     func() { acctest.PreCheck(t) },
			CheckDestroy: checkResourceIntegrationPolicyDestroy,
			Steps: []resource.TestStep{
				{
					ProtoV6ProviderFactories: acctest.Providers,
					SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minVersionIntegrationPolicy),
					ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
					ConfigVariables: config.Variables{
						"policy_name": config.StringVariable(policyName),
						"secret_key":  config.StringVariable("created"),
					},
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "name", policyName),
						resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "description", "IntegrationPolicyTest Policy"),
						resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "integration_name", "aws_logs"),
						resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "integration_version", "1.4.0"),
						resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "vars_json", fmt.Sprintf(`{"access_key_id":"placeholder","default_region":"us-east-1","endpoint":"endpoint","secret_access_key":"created %s","session_token":"placeholder"}`, policyName)),
						resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "inputs.aws_logs-aws-cloudwatch.enabled", "true"),
						resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "inputs.aws_logs-aws-cloudwatch.streams.aws_logs.generic.enabled", "true"),
						resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "inputs.aws_logs-aws-cloudwatch.streams.aws_logs.generic.vars", `{"api_sleep":"200ms","api_timeput":"120s","custom":"","data_stream.dataset":"aws_logs.generic","log_streams":[],"number_of_workers":1,"preserve_original_event":false,"scan_frequency":"1m","start_position":"beginning","tags":["forwarded"]}`),
					),
				},
				{
					ProtoV6ProviderFactories: acctest.Providers,
					SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minVersionIntegrationPolicy),
					ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
					ConfigVariables: config.Variables{
						"policy_name": config.StringVariable(policyName),
						"secret_key":  config.StringVariable("updated"),
					},
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "name", policyName),
						resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "description", "Updated Integration Policy"),
						resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "integration_name", "aws_logs"),
						resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "integration_version", "1.4.0"),
						resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "vars_json", fmt.Sprintf(`{"access_key_id":"placeholder","default_region":"us-east-2","endpoint":"endpoint","secret_access_key":"updated %s","session_token":"placeholder"}`, policyName)),
						resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "inputs.aws_logs-aws-cloudwatch.enabled", "false"),
						resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "inputs.aws_logs-aws-cloudwatch.streams.aws_logs.generic.enabled", "false"),
						resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "inputs.aws_logs-aws-cloudwatch.streams.aws_logs.generic.vars", `{"api_sleep":"200ms","api_timeput":"120s","custom":"","data_stream.dataset":"aws_logs.generic","log_streams":[],"number_of_workers":1,"preserve_original_event":false,"scan_frequency":"2m","start_position":"beginning","tags":["forwarded"]}`),
					),
				},
				{
					ProtoV6ProviderFactories: acctest.Providers,
					SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minVersionIntegrationPolicy),
					ResourceName:             "elasticstack_fleet_integration_policy.test_policy",
					ConfigDirectory:          acctest.NamedTestCaseDirectory("import_test"),
					ConfigVariables: config.Variables{
						"policy_name": config.StringVariable(policyName),
						"secret_key":  config.StringVariable("updated"),
					},
					ImportState:             true,
					ImportStateVerify:       true,
					ImportStateVerifyIgnore: []string{"vars_json", "space_ids"},
					Check: resource.ComposeTestCheckFunc(
						resource.TestMatchResourceAttr("elasticstack_fleet_integration_policy.test_policy", "vars_json", regexp.MustCompile(`{"access_key_id":{"id":"\S+","isSecretRef":true},"default_region":"us-east-2","endpoint":"endpoint","secret_access_key":{"id":"\S+","isSecretRef":true},"session_token":{"id":"\S+","isSecretRef":true}}`)),
					),
				},
			},
		})
	})

	t.Run("multi-valued secrets", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			PreCheck:     func() { acctest.PreCheck(t) },
			CheckDestroy: checkResourceIntegrationPolicyDestroy,
			Steps: []resource.TestStep{
				{
					ProtoV6ProviderFactories: acctest.Providers,
					SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minVersionSqlIntegration),
					ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
					ConfigVariables: config.Variables{
						"policy_name": config.StringVariable(policyName),
						"secret_key":  config.StringVariable("created"),
					},
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "name", policyName),
						resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "description", "SQL Integration Policy"),
						resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "integration_name", "sql"),
						resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "integration_version", "1.1.0"),
						resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "inputs.sql-sql/metrics.enabled", "true"),
						resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "inputs.sql-sql/metrics.streams.sql.sql.enabled", "true"),
						resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "inputs.sql-sql/metrics.streams.sql.sql.vars", `{"data_stream.dataset":"sql","driver":"mysql","hosts":["root:test@tcp(127.0.0.1:3306)/"],"merge_results":false,"period":"1m","processors":"","sql_queries":"- query: SHOW GLOBAL STATUS LIKE 'Innodb_system%'\n  response_format: variables\n        \n","ssl":""}`),
					),
				},
				{
					ProtoV6ProviderFactories: acctest.Providers,
					SkipFunc: func() (bool, error) {
						return versionutils.CheckIfVersionIsUnsupported(minVersionSqlIntegration)()
					},
					ResourceName:    "elasticstack_fleet_integration_policy.test_policy",
					ConfigDirectory: acctest.NamedTestCaseDirectory("import_test"),
					ConfigVariables: config.Variables{
						"policy_name": config.StringVariable(policyName),
						"secret_key":  config.StringVariable("created"),
					},
					ImportState:             true,
					ImportStateVerify:       true,
					ImportStateVerifyIgnore: []string{"inputs.sql-sql/metrics.streams.sql.sql.vars", "space_ids"},
					Check: resource.ComposeTestCheckFunc(
						resource.TestMatchResourceAttr("elasticstack_fleet_integration_policy.test_policy", "inputs.sql-sql/metrics.streams.sql.sql.vars", regexp.MustCompile(`"hosts":{"ids":["\S+"],"isSecretRef":true}`)),
					),
				},
			},
		})
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
			policy, diags := fleet.GetAgentPolicy(context.Background(), fleetClient, rs.Primary.ID, "")
			if diags.HasError() {
				return diagutil.FwDiagsAsError(diags)
			}
			if policy != nil {
				return fmt.Errorf("agent policy id=%v still exists, but it should have been removed", rs.Primary.ID)
			}
		case "elasticstack_fleet_integration_policy":
			policy, diags := fleet.GetPackagePolicy(context.Background(), fleetClient, rs.Primary.ID, "")
			if diags.HasError() {
				return diagutil.FwDiagsAsError(diags)
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
