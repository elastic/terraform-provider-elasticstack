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
	minVersionGCPVertexAI          = version.Must(version.NewVersion("8.17.0"))
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
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "policy_id", fmt.Sprintf("%s-policy-id", policyName)),
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
					ImportState:       true,
					ImportStateVerify: true,
					ImportStateVerifyIgnore: []string{
						"vars_json",
						"space_ids",
						"inputs.aws_logs-aws-cloudwatch.defaults",
						"inputs.aws_logs-aws-s3.defaults",
					},
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

func TestAccIntegrationPolicyAzureMetrics(t *testing.T) {
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
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "description", "Azure Metrics Integration Policy"),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "integration_name", "azure_metrics"),
					resource.TestCheckResourceAttrPair("elasticstack_fleet_integration_policy.test_policy", "integration_version", "data.elasticstack_fleet_integration.test", "version"),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "vars_json", `{"client_id":"test-client-id","client_secret":"test-client-secret","subscription_id":"test-subscription-id","tenant_id":"test-tenant-id"}`),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "inputs.monitor-azure/metrics.enabled", "true"),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "inputs.monitor-azure/metrics.streams.azure.monitor.enabled", "true"),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "inputs.monitor-azure/metrics.streams.azure.monitor.vars", `{"period":"300s","resources":"- resource_query: \"resourceType eq 'Microsoft.Search/searchServices'\"\n  metrics:\n  - name: [\"DocumentsProcessedCount\", \"SearchLatency\", \"SearchQueriesPerSecond\", \"ThrottledSearchQueriesPercentage\"]\n    namespace: \"Microsoft.Search/searchServices\""}`),
				),
			},
		},
	})
}

func TestAccIntegrationPolicyInputs(t *testing.T) {
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
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "description", "Kafka Integration Policy"),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "integration_name", "kafka"),
					// Check enabled inputs
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "inputs.kafka-logfile.enabled", "true"),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "inputs.kafka-kafka/metrics.enabled", "true"),
					// Check enabled streams
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "inputs.kafka-logfile.streams.kafka.log.enabled", "true"),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "inputs.kafka-kafka/metrics.streams.kafka.broker.enabled", "true"),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "inputs.kafka-kafka/metrics.streams.kafka.consumergroup.enabled", "true"),
					// Check disabled stream
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "inputs.kafka-kafka/metrics.streams.kafka.partition.enabled", "false"),
					// Check vars
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "inputs.kafka-kafka/metrics.vars", `{"hosts":["localhost:9092"],"period":"10s","ssl.certificate_authorities":[]}`),
					// Check unspecified, disabled by default input
					resource.TestCheckNoResourceAttr("elasticstack_fleet_integration_policy.test_policy", "inputs.kafka-jolokia/metrics"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minVersionIntegrationPolicy),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update_disabled_input"),
				ConfigVariables: config.Variables{
					"policy_name": config.StringVariable(policyName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "name", policyName),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "description", "Kafka Integration Policy - Updated"),
					// Check that disabling an input works correctly
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "inputs.kafka-logfile.enabled", "false"),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "inputs.kafka-kafka/metrics.enabled", "true"),
					// Vars should remain the same since we didn't change them
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "inputs.kafka-kafka/metrics.vars", `{"hosts":["localhost:9092"],"period":"10s","ssl.certificate_authorities":[]}`),

					// Disabled input should have no vars/streams in state
					resource.TestCheckNoResourceAttr("elasticstack_fleet_integration_policy.test_policy", "inputs.kafka-logfile.vars"),
					resource.TestCheckNoResourceAttr("elasticstack_fleet_integration_policy.test_policy", "inputs.kafka-logfile.streams"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minVersionIntegrationPolicy),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update_enabled_input"),
				ConfigVariables: config.Variables{
					"policy_name": config.StringVariable(policyName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "name", policyName),
					// Check that updating an enabled input's vars triggers a change
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "inputs.kafka-kafka/metrics.streams.kafka.consumergroup.vars", `{"topics":["don't mention the war, I mentioned it once but I think I got away with it"]}`),
					// Disabled input should remain disabled
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "inputs.kafka-logfile.enabled", "false"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minVersionIntegrationPolicy),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update_reenable_input"),
				ConfigVariables: config.Variables{
					"policy_name": config.StringVariable(policyName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "name", policyName),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "description", "Kafka Integration Policy - Re-enabled"),
					// Check that the kafka-logfile input is re-enabled
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "inputs.kafka-logfile.enabled", "true"),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "inputs.kafka-logfile.streams.kafka.log.enabled", "true"),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "inputs.kafka-logfile.streams.kafka.log.vars", `{"kafka_home":"/opt/kafka*","paths":["/logs/controller.log*","/logs/server.log*","/logs/state-change.log*","/logs/kafka-*.log*"],"preserve_original_event":false,"tags":["kafka-log"]}`),
					// Check that the kafka/metrics input remains enabled
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "inputs.kafka-kafka/metrics.enabled", "true"),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "inputs.kafka-kafka/metrics.vars", `{"hosts":["localhost:9092"],"period":"10s","ssl.certificate_authorities":[]}`),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "inputs.kafka-kafka/metrics.streams.kafka.broker.enabled", "true"),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "inputs.kafka-kafka/metrics.streams.kafka.consumergroup.enabled", "true"),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "inputs.kafka-kafka/metrics.streams.kafka.consumergroup.vars", `{"topics":["don't mention the war, I mentioned it once but I think I got away with it"]}`),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "inputs.kafka-kafka/metrics.streams.kafka.partition.enabled", "false"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minVersionIntegrationPolicy),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update_logfile_tags_only"),
				ConfigVariables: config.Variables{
					"policy_name": config.StringVariable(policyName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "name", policyName),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "description", "Kafka Integration Policy - Logfile with tags only"),
					// Check that the kafka-logfile input is enabled with only tags specified
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "inputs.kafka-logfile.enabled", "true"),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "inputs.kafka-logfile.streams.kafka.log.enabled", "true"),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "inputs.kafka-logfile.streams.kafka.log.vars", `{"tags":["custom-tag-1","custom-tag-2"]}`),
					// Check that the kafka/metrics input remains enabled
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "inputs.kafka-kafka/metrics.enabled", "true"),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "inputs.kafka-kafka/metrics.vars", `{"hosts":["localhost:9092"],"period":"10s","ssl.certificate_authorities":[]}`),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "inputs.kafka-kafka/metrics.streams.kafka.broker.enabled", "true"),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "inputs.kafka-kafka/metrics.streams.kafka.consumergroup.enabled", "true"),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "inputs.kafka-kafka/metrics.streams.kafka.consumergroup.vars", `{"topics":["don't mention the war, I mentioned it once but I think I got away with it"]}`),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "inputs.kafka-kafka/metrics.streams.kafka.partition.enabled", "false"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minVersionIntegrationPolicy),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("minimal"),
				ConfigVariables: config.Variables{
					"policy_name": config.StringVariable(policyName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "name", policyName),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "description", "Kafka Integration Policy - Minimal"),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "integration_name", "kafka"),
					// Check specified inputs
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "inputs.kafka-logfile.enabled", "false"),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "inputs.kafka-kafka/metrics.enabled", "true"),
					// Check unspecified, disabled by default input
					resource.TestCheckNoResourceAttr("elasticstack_fleet_integration_policy.test_policy", "inputs.kafka-jolokia/metrics"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minVersionIntegrationPolicy),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("unset"),
				ConfigVariables: config.Variables{
					"policy_name": config.StringVariable(policyName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "name", policyName),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "description", "Kafka Integration Policy - Minimal"),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "integration_name", "kafka"),
					// Check previously specified inputs
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "inputs.kafka-logfile.enabled", "false"),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "inputs.kafka-kafka/metrics.enabled", "true"),
				),
			},
		},
	})
}

func TestAccResourceIntegrationPolicyGCPVertexAI(t *testing.T) {
	policyName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceIntegrationPolicyDestroy,
		Steps: []resource.TestStep{
			{
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minVersionGCPVertexAI),
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"policy_name": config.StringVariable(policyName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "name", policyName),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "integration_name", "gcp_vertexai"),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "integration_version", "1.4.0"),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "vars_json", `{"project_id":"my-gcp-project"}`),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "inputs.GCP Vertex AI Metrics-gcp/metrics.enabled", "true"),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "inputs.GCP Vertex AI Metrics-gcp/metrics.streams.gcp_vertexai.metrics.enabled", "true"),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "inputs.GCP Vertex AI Metrics-gcp/metrics.streams.gcp_vertexai.metrics.vars", `{"period":"60s","regions":["us-central1"]}`),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "inputs.GCP Vertex AI  Logs-gcp/metrics.enabled", "true"),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "inputs.GCP Vertex AI  Logs-gcp/metrics.streams.gcp_vertexai.prompt_response_logs.enabled", "true"),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "inputs.GCP Vertex AI  Logs-gcp/metrics.streams.gcp_vertexai.prompt_response_logs.vars", `{"exclude_labels":false,"period":"300s","table_id":"table_id","tags":["forwarded","gcp-vertexai-prompt-response-logs"],"time_lookback_hours":1}`),
				),
			},
			{
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minVersionGCPVertexAI),
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables: config.Variables{
					"policy_name": config.StringVariable(policyName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "name", policyName),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "integration_name", "gcp_vertexai"),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "integration_version", "1.4.0"),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "vars_json", `{"project_id":"my-gcp-project"}`),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "inputs.GCP Vertex AI Metrics-gcp/metrics.enabled", "true"),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "inputs.GCP Vertex AI Metrics-gcp/metrics.streams.gcp_vertexai.metrics.enabled", "true"),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "inputs.GCP Vertex AI Metrics-gcp/metrics.streams.gcp_vertexai.metrics.vars", `{"period":"60s","regions":["us-central1"]}`),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "inputs.GCP Vertex AI  Logs-gcp/metrics.enabled", "true"),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "inputs.GCP Vertex AI  Logs-gcp/metrics.streams.gcp_vertexai.prompt_response_logs.enabled", "true"),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "inputs.GCP Vertex AI  Logs-gcp/metrics.streams.gcp_vertexai.prompt_response_logs.vars", `{"exclude_labels":false,"period":"300s","table_id":"table_id","tags":["forwarded","gcp-vertexai-prompt-response-logs"],"time_lookback_hours":1}`),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "inputs.GCP Vertex AI  Logs-gcp-pubsub.enabled", "true"),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "inputs.GCP Vertex AI  Logs-gcp-pubsub.streams.gcp_vertexai.auditlogs.enabled", "true"),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "inputs.GCP Vertex AI  Logs-gcp-pubsub.streams.gcp_vertexai.auditlogs.vars", `{"preserve_original_event":false,"subscription_create":false,"subscription_name":"gcp-vertexai-audit-sub","tags":["forwarded","gcp-vertexai-audit"],"topic":"gcp-vertexai-audit"}`),
				),
			},
		},
	})
}

// TestAccResourceIntegrationPolicy_VersionUpdate tests that updating integration_version
// preserves agent_policy_id without causing inconsistent state errors.
// This is a regression test for https://github.com/elastic/terraform-provider-elasticstack/pull/1616
func TestAccResourceIntegrationPolicy_VersionUpdate(t *testing.T) {
	policyName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceIntegrationPolicyDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minVersionIntegrationPolicy),
				Config: `
resource "elasticstack_fleet_agent_policy" "test_policy" {
  name      = "` + policyName + `"
  namespace = "default"
}

resource "elasticstack_fleet_integration_policy" "test_policy" {
  name               = "` + policyName + `-integration"
  namespace          = "default"
  agent_policy_id    = elasticstack_fleet_agent_policy.test_policy.policy_id
  integration_name    = "tcp"
  integration_version = "1.16.0"
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "name", policyName+"-integration"),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "integration_name", "tcp"),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "integration_version", "1.16.0"),
					resource.TestCheckResourceAttrSet("elasticstack_fleet_integration_policy.test_policy", "agent_policy_id"),
					resource.TestCheckResourceAttrPair(
						"elasticstack_fleet_integration_policy.test_policy", "agent_policy_id",
						"elasticstack_fleet_agent_policy.test_policy", "policy_id",
					),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minVersionIntegrationPolicy),
				Config: `
resource "elasticstack_fleet_agent_policy" "test_policy" {
  name      = "` + policyName + `"
  namespace = "default"
}

resource "elasticstack_fleet_integration_policy" "test_policy" {
  name               = "` + policyName + `-integration"
  namespace          = "default"
  agent_policy_id    = elasticstack_fleet_agent_policy.test_policy.policy_id
  integration_name    = "tcp"
  integration_version = "1.17.0"  # Updated version
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "name", policyName+"-integration"),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "integration_name", "tcp"),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration_policy.test_policy", "integration_version", "1.17.0"),
					// Critical check: agent_policy_id must still be set after version update
					resource.TestCheckResourceAttrSet("elasticstack_fleet_integration_policy.test_policy", "agent_policy_id"),
					resource.TestCheckResourceAttrPair(
						"elasticstack_fleet_integration_policy.test_policy", "agent_policy_id",
						"elasticstack_fleet_agent_policy.test_policy", "policy_id",
					),
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
