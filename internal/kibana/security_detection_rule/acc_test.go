package security_detection_rule_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibana_oapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/google/uuid"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// checkResourceJSONAttr compares the JSON string value of a resource attribute
func checkResourceJSONAttr(name, key, expectedJSON string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		ms := s.RootModule()
		rs, ok := ms.Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s in %s", name, ms.Path)
		}
		is := rs.Primary
		if is == nil {
			return fmt.Errorf("No primary instance: %s in %s", name, ms.Path)
		}

		actualJSON, ok := is.Attributes[key]
		if !ok {
			return fmt.Errorf("%s: Attribute '%s' not found", name, key)
		}

		if eq, err := utils.JSONBytesEqual([]byte(expectedJSON), []byte(actualJSON)); !eq {
			return fmt.Errorf(
				"%s: Attribute '%s' expected %#v, got %#v (<err>: %v)",
				name,
				key,
				expectedJSON,
				actualJSON,
				err)
		}
		return nil
	}
}

var minVersionSupport = version.Must(version.NewVersion("8.11.0"))

func TestAccResourceSecurityDetectionRule_Query(t *testing.T) {
	resourceName := "elasticstack_kibana_security_detection_rule.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.Providers,
		CheckDestroy:             testAccCheckSecurityDetectionRuleDestroy,
		Steps: []resource.TestStep{
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minVersionSupport),
				Config:   testAccSecurityDetectionRuleConfig_query("test-query-rule"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", "test-query-rule"),
					resource.TestCheckResourceAttr(resourceName, "type", "query"),
					resource.TestCheckResourceAttr(resourceName, "query", "*:*"),
					resource.TestCheckResourceAttr(resourceName, "language", "kuery"),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "description", "Test query security detection rule"),
					resource.TestCheckResourceAttr(resourceName, "severity", "medium"),
					resource.TestCheckResourceAttr(resourceName, "risk_score", "50"),
					resource.TestCheckResourceAttr(resourceName, "index.0", "logs-*"),
					resource.TestCheckResourceAttr(resourceName, "data_view_id", "test-data-view-id"),
					resource.TestCheckResourceAttr(resourceName, "namespace", "test-namespace"),
					resource.TestCheckResourceAttr(resourceName, "rule_name_override", "Custom Query Rule Name"),
					resource.TestCheckResourceAttr(resourceName, "timestamp_override", "@timestamp"),
					resource.TestCheckResourceAttr(resourceName, "timestamp_override_fallback_disabled", "true"),

					// Check risk score mapping
					resource.TestCheckResourceAttr(resourceName, "risk_score_mapping.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "risk_score_mapping.0.field", "event.severity"),
					resource.TestCheckResourceAttr(resourceName, "risk_score_mapping.0.operator", "equals"),
					resource.TestCheckResourceAttr(resourceName, "risk_score_mapping.0.value", "high"),
					resource.TestCheckResourceAttr(resourceName, "risk_score_mapping.0.risk_score", "85"),

					// Check investigation fields
					resource.TestCheckResourceAttr(resourceName, "investigation_fields.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "investigation_fields.0", "user.name"),
					resource.TestCheckResourceAttr(resourceName, "investigation_fields.1", "event.action"),

					// Check meta field
					checkResourceJSONAttr(resourceName, "meta", `{"custom_field": "test_value", "environment": "testing", "version": "1.0"}`),

					// Check related integrations
					resource.TestCheckResourceAttr(resourceName, "related_integrations.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "related_integrations.0.package", "windows"),
					resource.TestCheckResourceAttr(resourceName, "related_integrations.0.version", "1.0.0"),
					resource.TestCheckResourceAttr(resourceName, "related_integrations.0.integration", "system"),

					// Check required fields
					resource.TestCheckResourceAttr(resourceName, "required_fields.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "required_fields.0.name", "event.type"),
					resource.TestCheckResourceAttr(resourceName, "required_fields.0.type", "keyword"),
					resource.TestCheckResourceAttr(resourceName, "required_fields.0.ecs", "true"),
					resource.TestCheckResourceAttr(resourceName, "required_fields.1.name", "host.os.type"),
					resource.TestCheckResourceAttr(resourceName, "required_fields.1.type", "keyword"),
					resource.TestCheckResourceAttr(resourceName, "required_fields.1.ecs", "true"),

					// Check severity mapping
					resource.TestCheckResourceAttr(resourceName, "severity_mapping.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "severity_mapping.0.field", "event.severity_level"),
					resource.TestCheckResourceAttr(resourceName, "severity_mapping.0.operator", "equals"),
					resource.TestCheckResourceAttr(resourceName, "severity_mapping.0.value", "critical"),
					resource.TestCheckResourceAttr(resourceName, "severity_mapping.0.severity", "critical"),

					// Check response actions
					resource.TestCheckResourceAttr(resourceName, "response_actions.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "response_actions.0.action_type_id", ".osquery"),
					resource.TestCheckResourceAttr(resourceName, "response_actions.0.params.query", "SELECT * FROM processes WHERE name = 'malicious.exe';"),
					resource.TestCheckResourceAttr(resourceName, "response_actions.0.params.timeout", "300"),
					resource.TestCheckResourceAttr(resourceName, "response_actions.0.params.ecs_mapping.process.name", "name"),
					resource.TestCheckResourceAttr(resourceName, "response_actions.0.params.ecs_mapping.process.pid", "pid"),
					resource.TestCheckResourceAttr(resourceName, "response_actions.1.action_type_id", ".endpoint"),
					resource.TestCheckResourceAttr(resourceName, "response_actions.1.params.command", "isolate"),
					resource.TestCheckResourceAttr(resourceName, "response_actions.1.params.comment", "Isolate host due to suspicious activity"),

					// Verify building_block_type is not set by default
					resource.TestCheckNoResourceAttr(resourceName, "building_block_type"),

					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "rule_id"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "created_by"),
				),
			},
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minVersionSupport),
				Config:   testAccSecurityDetectionRuleConfig_queryUpdate("test-query-rule-updated"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", "test-query-rule-updated"),
					resource.TestCheckResourceAttr(resourceName, "description", "Updated test query security detection rule"),
					resource.TestCheckResourceAttr(resourceName, "severity", "high"),
					resource.TestCheckResourceAttr(resourceName, "risk_score", "75"),
					resource.TestCheckResourceAttr(resourceName, "data_view_id", "updated-data-view-id"),
					resource.TestCheckResourceAttr(resourceName, "namespace", "updated-namespace"),
					resource.TestCheckResourceAttr(resourceName, "rule_name_override", "Updated Custom Query Rule Name"),
					resource.TestCheckResourceAttr(resourceName, "timestamp_override", "event.ingested"),
					resource.TestCheckResourceAttr(resourceName, "timestamp_override_fallback_disabled", "false"),

					// Check risk score mapping
					resource.TestCheckResourceAttr(resourceName, "risk_score_mapping.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "risk_score_mapping.0.field", "event.risk_level"),
					resource.TestCheckResourceAttr(resourceName, "risk_score_mapping.0.operator", "equals"),
					resource.TestCheckResourceAttr(resourceName, "risk_score_mapping.0.value", "critical"),
					resource.TestCheckResourceAttr(resourceName, "risk_score_mapping.0.risk_score", "95"),

					// Check investigation fields
					resource.TestCheckResourceAttr(resourceName, "investigation_fields.#", "3"),
					resource.TestCheckResourceAttr(resourceName, "investigation_fields.0", "user.name"),
					resource.TestCheckResourceAttr(resourceName, "investigation_fields.1", "event.action"),
					resource.TestCheckResourceAttr(resourceName, "investigation_fields.2", "source.ip"),

					// Check meta field (updated values)
					checkResourceJSONAttr(resourceName, "meta", `{"custom_field": "updated_value", "environment": "production", "version": "2.0", "team": "security"}`),

					// Check related integrations (updated values)
					resource.TestCheckResourceAttr(resourceName, "related_integrations.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "related_integrations.0.package", "linux"),
					resource.TestCheckResourceAttr(resourceName, "related_integrations.0.version", "2.0.0"),
					resource.TestCheckResourceAttr(resourceName, "related_integrations.0.integration", "auditd"),
					resource.TestCheckResourceAttr(resourceName, "related_integrations.1.package", "network"),
					resource.TestCheckResourceAttr(resourceName, "related_integrations.1.version", "1.5.0"),
					resource.TestCheckNoResourceAttr(resourceName, "related_integrations.1.integration"),

					// Check required fields (updated values)
					resource.TestCheckResourceAttr(resourceName, "required_fields.#", "3"),
					resource.TestCheckResourceAttr(resourceName, "required_fields.0.name", "event.category"),
					resource.TestCheckResourceAttr(resourceName, "required_fields.0.type", "keyword"),
					resource.TestCheckResourceAttr(resourceName, "required_fields.0.ecs", "true"),
					resource.TestCheckResourceAttr(resourceName, "required_fields.1.name", "process.name"),
					resource.TestCheckResourceAttr(resourceName, "required_fields.1.type", "keyword"),
					resource.TestCheckResourceAttr(resourceName, "required_fields.1.ecs", "true"),
					resource.TestCheckResourceAttr(resourceName, "required_fields.2.name", "custom.field"),
					resource.TestCheckResourceAttr(resourceName, "required_fields.2.type", "text"),
					resource.TestCheckResourceAttr(resourceName, "required_fields.2.ecs", "false"),

					// Check severity mapping (updated values)
					resource.TestCheckResourceAttr(resourceName, "severity_mapping.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "severity_mapping.0.field", "alert.severity"),
					resource.TestCheckResourceAttr(resourceName, "severity_mapping.0.operator", "equals"),
					resource.TestCheckResourceAttr(resourceName, "severity_mapping.0.value", "high"),
					resource.TestCheckResourceAttr(resourceName, "severity_mapping.0.severity", "high"),
					resource.TestCheckResourceAttr(resourceName, "severity_mapping.1.field", "alert.severity"),
					resource.TestCheckResourceAttr(resourceName, "severity_mapping.1.operator", "equals"),
					resource.TestCheckResourceAttr(resourceName, "severity_mapping.1.value", "medium"),
					resource.TestCheckResourceAttr(resourceName, "severity_mapping.1.severity", "medium"),

					// Check response actions
					resource.TestCheckResourceAttr(resourceName, "response_actions.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "response_actions.0.action_type_id", ".osquery"),
					resource.TestCheckResourceAttr(resourceName, "response_actions.0.params.pack_id", "incident_response_pack"),
					resource.TestCheckResourceAttr(resourceName, "response_actions.0.params.timeout", "600"),
					resource.TestCheckResourceAttr(resourceName, "response_actions.0.params.ecs_mapping.host.name", "hostname"),
					resource.TestCheckResourceAttr(resourceName, "response_actions.0.params.ecs_mapping.user.name", "username"),
					resource.TestCheckResourceAttr(resourceName, "response_actions.0.params.ecs_mapping.process.name", "process_name"),
					resource.TestCheckResourceAttr(resourceName, "response_actions.0.params.queries.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "response_actions.0.params.queries.0.id", "query1"),
					resource.TestCheckResourceAttr(resourceName, "response_actions.0.params.queries.0.query", "SELECT * FROM logged_in_users;"),
					resource.TestCheckResourceAttr(resourceName, "response_actions.0.params.queries.0.platform", "linux"),
					resource.TestCheckResourceAttr(resourceName, "response_actions.0.params.queries.0.version", "4.6.0"),
					resource.TestCheckResourceAttr(resourceName, "response_actions.0.params.queries.1.id", "query2"),
					resource.TestCheckResourceAttr(resourceName, "response_actions.0.params.queries.1.query", "SELECT * FROM processes WHERE state = 'R';"),
					resource.TestCheckResourceAttr(resourceName, "response_actions.0.params.queries.1.platform", "linux"),
					resource.TestCheckResourceAttr(resourceName, "response_actions.0.params.queries.1.version", "4.6.0"),
					resource.TestCheckResourceAttr(resourceName, "response_actions.0.params.queries.1.ecs_mapping.process.pid", "pid"),
					resource.TestCheckResourceAttr(resourceName, "response_actions.0.params.queries.1.ecs_mapping.process.command_line", "cmdline"),
					resource.TestCheckResourceAttr(resourceName, "response_actions.1.action_type_id", ".endpoint"),
					resource.TestCheckResourceAttr(resourceName, "response_actions.1.params.command", "kill-process"),
					resource.TestCheckResourceAttr(resourceName, "response_actions.1.params.comment", "Kill suspicious process identified during investigation"),
					resource.TestCheckResourceAttr(resourceName, "response_actions.1.params.config.field", "process.entity_id"),
					resource.TestCheckResourceAttr(resourceName, "response_actions.1.params.config.overwrite", "true"),
				),
			},
		},
	})
}

func TestAccResourceSecurityDetectionRule_EQL(t *testing.T) {
	resourceName := "elasticstack_kibana_security_detection_rule.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.Providers,
		CheckDestroy:             testAccCheckSecurityDetectionRuleDestroy,
		Steps: []resource.TestStep{
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minVersionSupport),
				Config:   testAccSecurityDetectionRuleConfig_eql("test-eql-rule"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", "test-eql-rule"),
					resource.TestCheckResourceAttr(resourceName, "type", "eql"),
					resource.TestCheckResourceAttr(resourceName, "query", "process where process.name == \"cmd.exe\""),
					resource.TestCheckResourceAttr(resourceName, "language", "eql"),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "description", "Test EQL security detection rule"),
					resource.TestCheckResourceAttr(resourceName, "severity", "high"),
					resource.TestCheckResourceAttr(resourceName, "risk_score", "70"),
					resource.TestCheckResourceAttr(resourceName, "index.0", "winlogbeat-*"),
					resource.TestCheckResourceAttr(resourceName, "tiebreaker_field", "@timestamp"),
					resource.TestCheckResourceAttr(resourceName, "data_view_id", "eql-data-view-id"),
					resource.TestCheckResourceAttr(resourceName, "namespace", "eql-namespace"),
					resource.TestCheckResourceAttr(resourceName, "rule_name_override", "Custom EQL Rule Name"),
					resource.TestCheckResourceAttr(resourceName, "timestamp_override", "process.start"),
					resource.TestCheckResourceAttr(resourceName, "timestamp_override_fallback_disabled", "false"),

					// Check risk score mapping
					resource.TestCheckResourceAttr(resourceName, "risk_score_mapping.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "risk_score_mapping.0.field", "process.executable"),
					resource.TestCheckResourceAttr(resourceName, "risk_score_mapping.0.operator", "equals"),
					resource.TestCheckResourceAttr(resourceName, "risk_score_mapping.0.value", "C:\\Windows\\System32\\cmd.exe"),
					resource.TestCheckResourceAttr(resourceName, "risk_score_mapping.0.risk_score", "75"),

					// Check investigation fields
					resource.TestCheckResourceAttr(resourceName, "investigation_fields.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "investigation_fields.0", "process.name"),
					resource.TestCheckResourceAttr(resourceName, "investigation_fields.1", "process.executable"),

					// Check meta field
					checkResourceJSONAttr(resourceName, "meta", `{"rule_type": "eql", "process": "monitoring", "severity": "high"}`),

					// Check related integrations
					resource.TestCheckResourceAttr(resourceName, "related_integrations.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "related_integrations.0.package", "windows"),
					resource.TestCheckResourceAttr(resourceName, "related_integrations.0.version", "1.0.0"),
					resource.TestCheckResourceAttr(resourceName, "related_integrations.0.integration", "system"),

					// Check required fields
					resource.TestCheckResourceAttr(resourceName, "required_fields.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "required_fields.0.name", "process.name"),
					resource.TestCheckResourceAttr(resourceName, "required_fields.0.type", "keyword"),
					resource.TestCheckResourceAttr(resourceName, "required_fields.0.ecs", "true"),
					resource.TestCheckResourceAttr(resourceName, "required_fields.1.name", "event.type"),
					resource.TestCheckResourceAttr(resourceName, "required_fields.1.type", "keyword"),
					resource.TestCheckResourceAttr(resourceName, "required_fields.1.ecs", "true"),

					// Check severity mapping
					resource.TestCheckResourceAttr(resourceName, "severity_mapping.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "severity_mapping.0.field", "event.severity_level"),
					resource.TestCheckResourceAttr(resourceName, "severity_mapping.0.operator", "equals"),
					resource.TestCheckResourceAttr(resourceName, "severity_mapping.0.value", "high"),
					resource.TestCheckResourceAttr(resourceName, "severity_mapping.0.severity", "high"),

					// Check response actions
					resource.TestCheckResourceAttr(resourceName, "response_actions.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "response_actions.0.action_type_id", ".osquery"),
					resource.TestCheckResourceAttr(resourceName, "response_actions.0.params.saved_query_id", "suspicious_processes"),
					resource.TestCheckResourceAttr(resourceName, "response_actions.0.params.timeout", "300"),

					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "rule_id"),
				),
			},
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minVersionSupport),
				Config:   testAccSecurityDetectionRuleConfig_eqlUpdate("test-eql-rule-updated"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", "test-eql-rule-updated"),
					resource.TestCheckResourceAttr(resourceName, "query", "process where process.name == \"powershell.exe\""),
					resource.TestCheckResourceAttr(resourceName, "description", "Updated test EQL security detection rule"),
					resource.TestCheckResourceAttr(resourceName, "severity", "critical"),
					resource.TestCheckResourceAttr(resourceName, "risk_score", "90"),
					resource.TestCheckResourceAttr(resourceName, "rule_name_override", "Updated Custom EQL Rule Name"),
					resource.TestCheckResourceAttr(resourceName, "timestamp_override", "process.end"),
					resource.TestCheckResourceAttr(resourceName, "timestamp_override_fallback_disabled", "true"),

					// Check risk score mapping
					resource.TestCheckResourceAttr(resourceName, "risk_score_mapping.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "risk_score_mapping.0.field", "process.parent.name"),
					resource.TestCheckResourceAttr(resourceName, "risk_score_mapping.0.operator", "equals"),
					resource.TestCheckResourceAttr(resourceName, "risk_score_mapping.0.value", "cmd.exe"),
					resource.TestCheckResourceAttr(resourceName, "risk_score_mapping.0.risk_score", "95"),

					// Check investigation fields
					resource.TestCheckResourceAttr(resourceName, "investigation_fields.#", "3"),
					resource.TestCheckResourceAttr(resourceName, "investigation_fields.0", "process.name"),
					resource.TestCheckResourceAttr(resourceName, "investigation_fields.1", "process.executable"),
					resource.TestCheckResourceAttr(resourceName, "investigation_fields.2", "process.parent.name"),

					// Check meta field (updated values)
					checkResourceJSONAttr(resourceName, "meta", `{"rule_type": "eql", "process": "detection", "severity": "critical", "updated": "true"}`),

					// Check related integrations
					resource.TestCheckResourceAttr(resourceName, "related_integrations.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "related_integrations.0.package", "windows"),
					resource.TestCheckResourceAttr(resourceName, "related_integrations.0.version", "2.0.0"),
					resource.TestCheckResourceAttr(resourceName, "related_integrations.0.integration", "system"),

					// Check required fields
					resource.TestCheckResourceAttr(resourceName, "required_fields.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "required_fields.0.name", "process.parent.name"),
					resource.TestCheckResourceAttr(resourceName, "required_fields.0.type", "keyword"),
					resource.TestCheckResourceAttr(resourceName, "required_fields.0.ecs", "true"),
					resource.TestCheckResourceAttr(resourceName, "required_fields.1.name", "event.category"),
					resource.TestCheckResourceAttr(resourceName, "required_fields.1.type", "keyword"),
					resource.TestCheckResourceAttr(resourceName, "required_fields.1.ecs", "true"),

					// Check severity mapping
					resource.TestCheckResourceAttr(resourceName, "severity_mapping.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "severity_mapping.0.field", "event.severity_level"),
					resource.TestCheckResourceAttr(resourceName, "severity_mapping.0.operator", "equals"),
					resource.TestCheckResourceAttr(resourceName, "severity_mapping.0.value", "critical"),
					resource.TestCheckResourceAttr(resourceName, "severity_mapping.0.severity", "critical"),

					// Check response actions
					resource.TestCheckResourceAttr(resourceName, "response_actions.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "response_actions.0.action_type_id", ".osquery"),
					resource.TestCheckResourceAttr(resourceName, "response_actions.0.params.pack_id", "eql_response_pack"),
					resource.TestCheckResourceAttr(resourceName, "response_actions.0.params.timeout", "450"),
					resource.TestCheckResourceAttr(resourceName, "response_actions.0.params.ecs_mapping.process.executable", "executable_path"),
					resource.TestCheckResourceAttr(resourceName, "response_actions.0.params.ecs_mapping.process.parent.name", "parent_name"),
				),
			},
		},
	})
}

func TestAccResourceSecurityDetectionRule_ESQL(t *testing.T) {
	resourceName := "elasticstack_kibana_security_detection_rule.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.Providers,
		CheckDestroy:             testAccCheckSecurityDetectionRuleDestroy,
		Steps: []resource.TestStep{
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minVersionSupport),
				Config:   testAccSecurityDetectionRuleConfig_esql("test-esql-rule"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", "test-esql-rule"),
					resource.TestCheckResourceAttr(resourceName, "type", "esql"),
					resource.TestCheckResourceAttr(resourceName, "query", "FROM logs-* | WHERE event.action == \"login\" | STATS count(*) BY user.name"),
					resource.TestCheckResourceAttr(resourceName, "language", "esql"),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "description", "Test ESQL security detection rule"),
					resource.TestCheckResourceAttr(resourceName, "severity", "medium"),
					resource.TestCheckResourceAttr(resourceName, "risk_score", "60"),
					resource.TestCheckResourceAttr(resourceName, "namespace", "esql-namespace"),
					resource.TestCheckResourceAttr(resourceName, "rule_name_override", "Custom ESQL Rule Name"),
					resource.TestCheckResourceAttr(resourceName, "timestamp_override", "event.created"),
					resource.TestCheckResourceAttr(resourceName, "timestamp_override_fallback_disabled", "true"),

					// Check risk score mapping
					resource.TestCheckResourceAttr(resourceName, "risk_score_mapping.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "risk_score_mapping.0.field", "user.domain"),
					resource.TestCheckResourceAttr(resourceName, "risk_score_mapping.0.operator", "equals"),
					resource.TestCheckResourceAttr(resourceName, "risk_score_mapping.0.value", "admin"),
					resource.TestCheckResourceAttr(resourceName, "risk_score_mapping.0.risk_score", "80"),

					// Check investigation fields
					resource.TestCheckResourceAttr(resourceName, "investigation_fields.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "investigation_fields.0", "user.name"),
					resource.TestCheckResourceAttr(resourceName, "investigation_fields.1", "user.domain"),

					// Check meta field
					checkResourceJSONAttr(resourceName, "meta", `{"query_type": "esql", "analytics": "enabled", "phase": "testing"}`),

					// Check related integrations
					resource.TestCheckResourceAttr(resourceName, "related_integrations.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "related_integrations.0.package", "system"),
					resource.TestCheckResourceAttr(resourceName, "related_integrations.0.version", "1.0.0"),
					resource.TestCheckResourceAttr(resourceName, "related_integrations.0.integration", "auth"),

					// Check required fields
					resource.TestCheckResourceAttr(resourceName, "required_fields.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "required_fields.0.name", "user.name"),
					resource.TestCheckResourceAttr(resourceName, "required_fields.0.type", "keyword"),
					resource.TestCheckResourceAttr(resourceName, "required_fields.0.ecs", "true"),
					resource.TestCheckResourceAttr(resourceName, "required_fields.1.name", "event.action"),
					resource.TestCheckResourceAttr(resourceName, "required_fields.1.type", "keyword"),
					resource.TestCheckResourceAttr(resourceName, "required_fields.1.ecs", "true"),

					// Check severity mapping
					resource.TestCheckResourceAttr(resourceName, "severity_mapping.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "severity_mapping.0.field", "user.domain"),
					resource.TestCheckResourceAttr(resourceName, "severity_mapping.0.operator", "equals"),
					resource.TestCheckResourceAttr(resourceName, "severity_mapping.0.value", "admin"),
					resource.TestCheckResourceAttr(resourceName, "severity_mapping.0.severity", "high"),

					// Check response actions
					resource.TestCheckResourceAttr(resourceName, "response_actions.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "response_actions.0.action_type_id", ".osquery"),
					resource.TestCheckResourceAttr(resourceName, "response_actions.0.params.query", "SELECT * FROM users WHERE username LIKE '%admin%';"),
					resource.TestCheckResourceAttr(resourceName, "response_actions.0.params.timeout", "400"),
					resource.TestCheckResourceAttr(resourceName, "response_actions.0.params.ecs_mapping.user.name", "username"),
					resource.TestCheckResourceAttr(resourceName, "response_actions.0.params.ecs_mapping.user.domain", "domain"),
					resource.TestCheckResourceAttr(resourceName, "response_actions.1.action_type_id", ".endpoint"),
					resource.TestCheckResourceAttr(resourceName, "response_actions.1.params.command", "isolate"),
					resource.TestCheckResourceAttr(resourceName, "response_actions.1.params.comment", "Isolate host due to suspicious admin activity"),

					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "rule_id"),
				),
			},
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minVersionSupport),
				Config:   testAccSecurityDetectionRuleConfig_esqlUpdate("test-esql-rule-updated"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", "test-esql-rule-updated"),
					resource.TestCheckResourceAttr(resourceName, "query", "FROM logs-* | WHERE event.action == \"logout\" | STATS count(*) BY user.name, source.ip"),
					resource.TestCheckResourceAttr(resourceName, "description", "Updated test ESQL security detection rule"),
					resource.TestCheckResourceAttr(resourceName, "severity", "high"),
					resource.TestCheckResourceAttr(resourceName, "risk_score", "80"),
					resource.TestCheckResourceAttr(resourceName, "rule_name_override", "Updated Custom ESQL Rule Name"),
					resource.TestCheckResourceAttr(resourceName, "timestamp_override", "event.start"),
					resource.TestCheckResourceAttr(resourceName, "timestamp_override_fallback_disabled", "false"),

					// Check risk score mapping
					resource.TestCheckResourceAttr(resourceName, "risk_score_mapping.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "risk_score_mapping.0.field", "event.outcome"),
					resource.TestCheckResourceAttr(resourceName, "risk_score_mapping.0.operator", "equals"),
					resource.TestCheckResourceAttr(resourceName, "risk_score_mapping.0.value", "failure"),
					resource.TestCheckResourceAttr(resourceName, "risk_score_mapping.0.risk_score", "95"),

					// Check investigation fields
					resource.TestCheckResourceAttr(resourceName, "investigation_fields.#", "3"),
					resource.TestCheckResourceAttr(resourceName, "investigation_fields.0", "user.name"),
					resource.TestCheckResourceAttr(resourceName, "investigation_fields.1", "user.domain"),
					resource.TestCheckResourceAttr(resourceName, "investigation_fields.2", "event.outcome"),

					// Check meta field (updated values)
					checkResourceJSONAttr(resourceName, "meta", `{"query_type": "esql", "analytics": "enabled", "phase": "production", "updated": "yes"}`),

					// Check related integrations
					resource.TestCheckResourceAttr(resourceName, "related_integrations.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "related_integrations.0.package", "system"),
					resource.TestCheckResourceAttr(resourceName, "related_integrations.0.version", "2.0.0"),
					resource.TestCheckResourceAttr(resourceName, "related_integrations.0.integration", "auth"),

					// Check required fields
					resource.TestCheckResourceAttr(resourceName, "required_fields.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "required_fields.0.name", "user.name"),
					resource.TestCheckResourceAttr(resourceName, "required_fields.0.type", "keyword"),
					resource.TestCheckResourceAttr(resourceName, "required_fields.0.ecs", "true"),
					resource.TestCheckResourceAttr(resourceName, "required_fields.1.name", "event.outcome"),
					resource.TestCheckResourceAttr(resourceName, "required_fields.1.type", "keyword"),
					resource.TestCheckResourceAttr(resourceName, "required_fields.1.ecs", "true"),

					// Check severity mapping
					resource.TestCheckResourceAttr(resourceName, "severity_mapping.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "severity_mapping.0.field", "event.outcome"),
					resource.TestCheckResourceAttr(resourceName, "severity_mapping.0.operator", "equals"),
					resource.TestCheckResourceAttr(resourceName, "severity_mapping.0.value", "failure"),
					resource.TestCheckResourceAttr(resourceName, "severity_mapping.0.severity", "critical"),

					// Check response actions
					resource.TestCheckResourceAttr(resourceName, "response_actions.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "response_actions.0.action_type_id", ".osquery"),
					resource.TestCheckResourceAttr(resourceName, "response_actions.0.params.saved_query_id", "failed_login_investigation"),
					resource.TestCheckResourceAttr(resourceName, "response_actions.0.params.timeout", "500"),
					resource.TestCheckResourceAttr(resourceName, "response_actions.0.params.ecs_mapping.event.outcome", "outcome"),
					resource.TestCheckResourceAttr(resourceName, "response_actions.0.params.ecs_mapping.user.name", "username"),
					resource.TestCheckResourceAttr(resourceName, "response_actions.0.params.ecs_mapping.source.ip", "source_ip"),

					resource.TestCheckResourceAttr(resourceName, "exceptions_list.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "exceptions_list.0.id", "esql-exception-1"),
					resource.TestCheckResourceAttr(resourceName, "exceptions_list.0.list_id", "esql-rule-exceptions"),
					resource.TestCheckResourceAttr(resourceName, "exceptions_list.0.namespace_type", "single"),
					resource.TestCheckResourceAttr(resourceName, "exceptions_list.0.type", "detection"),
				),
			},
		},
	})
}

func TestAccResourceSecurityDetectionRule_MachineLearning(t *testing.T) {
	resourceName := "elasticstack_kibana_security_detection_rule.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.Providers,
		CheckDestroy:             testAccCheckSecurityDetectionRuleDestroy,
		Steps: []resource.TestStep{
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minVersionSupport),
				Config:   testAccSecurityDetectionRuleConfig_machineLearning("test-ml-rule"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", "test-ml-rule"),
					resource.TestCheckResourceAttr(resourceName, "type", "machine_learning"),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "description", "Test ML security detection rule"),
					resource.TestCheckResourceAttr(resourceName, "severity", "critical"),
					resource.TestCheckResourceAttr(resourceName, "risk_score", "90"),
					resource.TestCheckResourceAttr(resourceName, "anomaly_threshold", "75"),
					resource.TestCheckResourceAttr(resourceName, "machine_learning_job_id.0", "test-ml-job"),

					// Check meta field
					checkResourceJSONAttr(resourceName, "meta", `{"ml_type": "anomaly_detection", "custom_ml": "test_value", "threshold": "75"}`),

					resource.TestCheckResourceAttr(resourceName, "namespace", "ml-namespace"),
					resource.TestCheckResourceAttr(resourceName, "rule_name_override", "Custom ML Rule Name"),
					resource.TestCheckResourceAttr(resourceName, "timestamp_override", "ml.job_id"),
					resource.TestCheckResourceAttr(resourceName, "timestamp_override_fallback_disabled", "false"),

					// Check risk score mapping
					resource.TestCheckResourceAttr(resourceName, "risk_score_mapping.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "risk_score_mapping.0.field", "ml.anomaly_score"),
					resource.TestCheckResourceAttr(resourceName, "risk_score_mapping.0.operator", "equals"),
					resource.TestCheckResourceAttr(resourceName, "risk_score_mapping.0.value", "critical"),
					resource.TestCheckResourceAttr(resourceName, "risk_score_mapping.0.risk_score", "100"),

					// Check investigation fields
					resource.TestCheckResourceAttr(resourceName, "investigation_fields.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "investigation_fields.0", "ml.anomaly_score"),
					resource.TestCheckResourceAttr(resourceName, "investigation_fields.1", "ml.job_id"),

					// Check related integrations
					resource.TestCheckResourceAttr(resourceName, "related_integrations.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "related_integrations.0.package", "ml"),
					resource.TestCheckResourceAttr(resourceName, "related_integrations.0.version", "1.0.0"),
					resource.TestCheckResourceAttr(resourceName, "related_integrations.0.integration", "anomaly_detection"),

					// Check required fields
					resource.TestCheckResourceAttr(resourceName, "required_fields.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "required_fields.0.name", "ml.anomaly_score"),
					resource.TestCheckResourceAttr(resourceName, "required_fields.0.type", "double"),
					resource.TestCheckResourceAttr(resourceName, "required_fields.0.ecs", "false"),
					resource.TestCheckResourceAttr(resourceName, "required_fields.1.name", "ml.job_id"),
					resource.TestCheckResourceAttr(resourceName, "required_fields.1.type", "keyword"),
					resource.TestCheckResourceAttr(resourceName, "required_fields.1.ecs", "false"),

					// Check severity mapping
					resource.TestCheckResourceAttr(resourceName, "severity_mapping.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "severity_mapping.0.field", "ml.anomaly_score"),
					resource.TestCheckResourceAttr(resourceName, "severity_mapping.0.operator", "equals"),
					resource.TestCheckResourceAttr(resourceName, "severity_mapping.0.value", "critical"),
					resource.TestCheckResourceAttr(resourceName, "severity_mapping.0.severity", "critical"),

					// Check response actions
					resource.TestCheckResourceAttr(resourceName, "response_actions.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "response_actions.0.action_type_id", ".osquery"),
					resource.TestCheckResourceAttr(resourceName, "response_actions.0.params.query", "SELECT * FROM processes WHERE pid IN (SELECT DISTINCT pid FROM connections WHERE remote_address NOT LIKE '10.%' AND remote_address NOT LIKE '192.168.%' AND remote_address NOT LIKE '127.%');"),
					resource.TestCheckResourceAttr(resourceName, "response_actions.0.params.timeout", "600"),
					resource.TestCheckResourceAttr(resourceName, "response_actions.0.params.ecs_mapping.process.pid", "pid"),
					resource.TestCheckResourceAttr(resourceName, "response_actions.0.params.ecs_mapping.process.name", "name"),
					resource.TestCheckResourceAttr(resourceName, "response_actions.0.params.ecs_mapping.ml.anomaly_score", "anomaly_score"),

					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "rule_id"),
				),
			},
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minVersionSupport),
				Config:   testAccSecurityDetectionRuleConfig_machineLearningUpdate("test-ml-rule-updated"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", "test-ml-rule-updated"),
					resource.TestCheckResourceAttr(resourceName, "description", "Updated test ML security detection rule"),
					resource.TestCheckResourceAttr(resourceName, "severity", "high"),
					resource.TestCheckResourceAttr(resourceName, "risk_score", "85"),
					resource.TestCheckResourceAttr(resourceName, "anomaly_threshold", "80"),
					resource.TestCheckResourceAttr(resourceName, "machine_learning_job_id.0", "test-ml-job"),
					resource.TestCheckResourceAttr(resourceName, "machine_learning_job_id.1", "test-ml-job-2"),

					// Check meta field (updated values)
					checkResourceJSONAttr(resourceName, "meta", `{"ml_type": "anomaly_detection", "custom_ml": "updated_value", "threshold": "80", "updated": "yes"}`),

					resource.TestCheckResourceAttr(resourceName, "rule_name_override", "Updated Custom ML Rule Name"),
					resource.TestCheckResourceAttr(resourceName, "timestamp_override", "ml.anomaly_score"),
					resource.TestCheckResourceAttr(resourceName, "timestamp_override_fallback_disabled", "true"),

					// Check risk score mapping
					resource.TestCheckResourceAttr(resourceName, "risk_score_mapping.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "risk_score_mapping.0.field", "ml.is_anomaly"),
					resource.TestCheckResourceAttr(resourceName, "risk_score_mapping.0.operator", "equals"),
					resource.TestCheckResourceAttr(resourceName, "risk_score_mapping.0.value", "true"),
					resource.TestCheckResourceAttr(resourceName, "risk_score_mapping.0.risk_score", "95"),

					// Check investigation fields
					resource.TestCheckResourceAttr(resourceName, "investigation_fields.#", "3"),
					resource.TestCheckResourceAttr(resourceName, "investigation_fields.0", "ml.anomaly_score"),
					resource.TestCheckResourceAttr(resourceName, "investigation_fields.1", "ml.job_id"),
					resource.TestCheckResourceAttr(resourceName, "investigation_fields.2", "ml.is_anomaly"),

					// Check related integrations
					resource.TestCheckResourceAttr(resourceName, "related_integrations.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "related_integrations.0.package", "ml"),
					resource.TestCheckResourceAttr(resourceName, "related_integrations.0.version", "2.0.0"),
					resource.TestCheckResourceAttr(resourceName, "related_integrations.0.integration", "anomaly_detection"),

					// Check required fields
					resource.TestCheckResourceAttr(resourceName, "required_fields.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "required_fields.0.name", "ml.is_anomaly"),
					resource.TestCheckResourceAttr(resourceName, "required_fields.0.type", "boolean"),
					resource.TestCheckResourceAttr(resourceName, "required_fields.0.ecs", "false"),
					resource.TestCheckResourceAttr(resourceName, "required_fields.1.name", "ml.job_id"),
					resource.TestCheckResourceAttr(resourceName, "required_fields.1.type", "keyword"),
					resource.TestCheckResourceAttr(resourceName, "required_fields.1.ecs", "false"),

					// Check severity mapping
					resource.TestCheckResourceAttr(resourceName, "severity_mapping.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "severity_mapping.0.field", "ml.is_anomaly"),
					resource.TestCheckResourceAttr(resourceName, "severity_mapping.0.operator", "equals"),
					resource.TestCheckResourceAttr(resourceName, "severity_mapping.0.value", "true"),
					resource.TestCheckResourceAttr(resourceName, "severity_mapping.0.severity", "high"),

					// Check response actions
					resource.TestCheckResourceAttr(resourceName, "response_actions.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "response_actions.0.action_type_id", ".osquery"),
					resource.TestCheckResourceAttr(resourceName, "response_actions.0.params.pack_id", "ml_anomaly_investigation"),
					resource.TestCheckResourceAttr(resourceName, "response_actions.0.params.timeout", "700"),
					resource.TestCheckResourceAttr(resourceName, "response_actions.0.params.ecs_mapping.ml.job_id", "job_id"),
					resource.TestCheckResourceAttr(resourceName, "response_actions.0.params.ecs_mapping.ml.is_anomaly", "is_anomaly"),
					resource.TestCheckResourceAttr(resourceName, "response_actions.0.params.ecs_mapping.host.name", "hostname"),
					resource.TestCheckResourceAttr(resourceName, "response_actions.0.params.queries.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "response_actions.0.params.queries.0.id", "ml_query1"),
					resource.TestCheckResourceAttr(resourceName, "response_actions.0.params.queries.0.query", "SELECT * FROM system_info;"),
					resource.TestCheckResourceAttr(resourceName, "response_actions.0.params.queries.0.platform", "linux"),
					resource.TestCheckResourceAttr(resourceName, "response_actions.0.params.queries.0.version", "4.7.0"),
					resource.TestCheckResourceAttr(resourceName, "response_actions.1.action_type_id", ".endpoint"),
					resource.TestCheckResourceAttr(resourceName, "response_actions.1.params.command", "isolate"),
					resource.TestCheckResourceAttr(resourceName, "response_actions.1.params.comment", "Collect process tree for ML anomaly investigation"),

					resource.TestCheckResourceAttr(resourceName, "exceptions_list.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "exceptions_list.0.id", "ml-exception-1"),
					resource.TestCheckResourceAttr(resourceName, "exceptions_list.0.list_id", "ml-rule-exceptions"),
					resource.TestCheckResourceAttr(resourceName, "exceptions_list.0.namespace_type", "agnostic"),
					resource.TestCheckResourceAttr(resourceName, "exceptions_list.0.type", "detection"),
				),
			},
		},
	})
}

func TestAccResourceSecurityDetectionRule_NewTerms(t *testing.T) {
	resourceName := "elasticstack_kibana_security_detection_rule.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.Providers,
		CheckDestroy:             testAccCheckSecurityDetectionRuleDestroy,
		Steps: []resource.TestStep{
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minVersionSupport),
				Config:   testAccSecurityDetectionRuleConfig_newTerms("test-new-terms-rule"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", "test-new-terms-rule"),
					resource.TestCheckResourceAttr(resourceName, "type", "new_terms"),
					resource.TestCheckResourceAttr(resourceName, "query", "user.name:*"),
					resource.TestCheckResourceAttr(resourceName, "language", "kuery"),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "description", "Test new terms security detection rule"),
					resource.TestCheckResourceAttr(resourceName, "severity", "medium"),
					resource.TestCheckResourceAttr(resourceName, "risk_score", "50"),
					resource.TestCheckResourceAttr(resourceName, "index.0", "logs-*"),
					resource.TestCheckResourceAttr(resourceName, "new_terms_fields.0", "user.name"),

					// Check meta field
					checkResourceJSONAttr(resourceName, "meta", `{"new_terms_type": "user_behavior", "custom_field": "test_value", "detection": "anomaly"}`),

					resource.TestCheckResourceAttr(resourceName, "history_window_start", "now-14d"),
					resource.TestCheckResourceAttr(resourceName, "data_view_id", "new-terms-data-view-id"),
					resource.TestCheckResourceAttr(resourceName, "namespace", "new-terms-namespace"),
					resource.TestCheckResourceAttr(resourceName, "rule_name_override", "Custom New Terms Rule Name"),
					resource.TestCheckResourceAttr(resourceName, "timestamp_override", "user.created"),
					resource.TestCheckResourceAttr(resourceName, "timestamp_override_fallback_disabled", "true"),

					// Check risk score mapping
					resource.TestCheckResourceAttr(resourceName, "risk_score_mapping.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "risk_score_mapping.0.field", "user.type"),
					resource.TestCheckResourceAttr(resourceName, "risk_score_mapping.0.operator", "equals"),
					resource.TestCheckResourceAttr(resourceName, "risk_score_mapping.0.value", "service_account"),
					resource.TestCheckResourceAttr(resourceName, "risk_score_mapping.0.risk_score", "65"),

					// Check investigation fields
					resource.TestCheckResourceAttr(resourceName, "investigation_fields.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "investigation_fields.0", "user.name"),
					resource.TestCheckResourceAttr(resourceName, "investigation_fields.1", "user.type"),

					// Check related integrations
					resource.TestCheckResourceAttr(resourceName, "related_integrations.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "related_integrations.0.package", "security"),
					resource.TestCheckResourceAttr(resourceName, "related_integrations.0.version", "1.0.0"),
					resource.TestCheckResourceAttr(resourceName, "related_integrations.0.integration", "users"),

					// Check required fields
					resource.TestCheckResourceAttr(resourceName, "required_fields.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "required_fields.0.name", "user.name"),
					resource.TestCheckResourceAttr(resourceName, "required_fields.0.type", "keyword"),
					resource.TestCheckResourceAttr(resourceName, "required_fields.0.ecs", "true"),
					resource.TestCheckResourceAttr(resourceName, "required_fields.1.name", "user.type"),
					resource.TestCheckResourceAttr(resourceName, "required_fields.1.type", "keyword"),
					resource.TestCheckResourceAttr(resourceName, "required_fields.1.ecs", "false"),

					// Check severity mapping
					resource.TestCheckResourceAttr(resourceName, "severity_mapping.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "severity_mapping.0.field", "user.type"),
					resource.TestCheckResourceAttr(resourceName, "severity_mapping.0.operator", "equals"),
					resource.TestCheckResourceAttr(resourceName, "severity_mapping.0.value", "service_account"),
					resource.TestCheckResourceAttr(resourceName, "severity_mapping.0.severity", "medium"),

					// Check response actions
					resource.TestCheckResourceAttr(resourceName, "response_actions.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "response_actions.0.action_type_id", ".osquery"),
					resource.TestCheckResourceAttr(resourceName, "response_actions.0.params.query", "SELECT * FROM last WHERE username = '{{user.name}}';"),
					resource.TestCheckResourceAttr(resourceName, "response_actions.0.params.timeout", "350"),
					resource.TestCheckResourceAttr(resourceName, "response_actions.0.params.ecs_mapping.user.name", "username"),
					resource.TestCheckResourceAttr(resourceName, "response_actions.0.params.ecs_mapping.user.type", "user_type"),
					resource.TestCheckResourceAttr(resourceName, "response_actions.0.params.ecs_mapping.host.name", "hostname"),

					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "rule_id"),
				),
			},
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minVersionSupport),
				Config:   testAccSecurityDetectionRuleConfig_newTermsUpdate("test-new-terms-rule-updated"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", "test-new-terms-rule-updated"),
					resource.TestCheckResourceAttr(resourceName, "query", "user.name:* AND source.ip:*"),
					resource.TestCheckResourceAttr(resourceName, "description", "Updated test new terms security detection rule"),
					resource.TestCheckResourceAttr(resourceName, "severity", "high"),
					resource.TestCheckResourceAttr(resourceName, "risk_score", "75"),
					resource.TestCheckResourceAttr(resourceName, "index.0", "logs-*"),
					resource.TestCheckResourceAttr(resourceName, "index.1", "audit-*"),
					resource.TestCheckResourceAttr(resourceName, "new_terms_fields.0", "user.name"),
					resource.TestCheckResourceAttr(resourceName, "new_terms_fields.1", "source.ip"),

					// Check meta field (updated values)
					checkResourceJSONAttr(resourceName, "meta", `{"new_terms_type": "user_behavior", "custom_field": "updated_value", "detection": "anomaly", "updated": "yes"}`),

					resource.TestCheckResourceAttr(resourceName, "history_window_start", "now-30d"),
					resource.TestCheckResourceAttr(resourceName, "rule_name_override", "Updated Custom New Terms Rule Name"),
					resource.TestCheckResourceAttr(resourceName, "timestamp_override", "user.last_login"),
					resource.TestCheckResourceAttr(resourceName, "timestamp_override_fallback_disabled", "false"),

					// Check risk score mapping
					resource.TestCheckResourceAttr(resourceName, "risk_score_mapping.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "risk_score_mapping.0.field", "user.roles"),
					resource.TestCheckResourceAttr(resourceName, "risk_score_mapping.0.operator", "equals"),
					resource.TestCheckResourceAttr(resourceName, "risk_score_mapping.0.value", "admin"),
					resource.TestCheckResourceAttr(resourceName, "risk_score_mapping.0.risk_score", "95"),
					resource.TestCheckResourceAttr(resourceName, "risk_score_mapping.1.field", "source.geo.country_name"),
					resource.TestCheckResourceAttr(resourceName, "risk_score_mapping.1.operator", "equals"),
					resource.TestCheckResourceAttr(resourceName, "risk_score_mapping.1.value", "CN"),
					resource.TestCheckResourceAttr(resourceName, "risk_score_mapping.1.risk_score", "85"),

					// Check investigation fields
					resource.TestCheckResourceAttr(resourceName, "investigation_fields.#", "4"),
					resource.TestCheckResourceAttr(resourceName, "investigation_fields.0", "user.name"),
					resource.TestCheckResourceAttr(resourceName, "investigation_fields.1", "user.type"),
					resource.TestCheckResourceAttr(resourceName, "investigation_fields.2", "source.ip"),
					resource.TestCheckResourceAttr(resourceName, "investigation_fields.3", "user.roles"),

					// Check response actions
					resource.TestCheckResourceAttr(resourceName, "response_actions.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "response_actions.0.action_type_id", ".osquery"),
					resource.TestCheckResourceAttr(resourceName, "response_actions.0.params.saved_query_id", "admin_user_investigation"),
					resource.TestCheckResourceAttr(resourceName, "response_actions.0.params.timeout", "800"),
					resource.TestCheckResourceAttr(resourceName, "response_actions.0.params.ecs_mapping.user.roles", "roles"),
					resource.TestCheckResourceAttr(resourceName, "response_actions.0.params.ecs_mapping.source.ip", "source_ip"),
					resource.TestCheckResourceAttr(resourceName, "response_actions.0.params.ecs_mapping.user.name", "username"),
					resource.TestCheckResourceAttr(resourceName, "response_actions.1.action_type_id", ".endpoint"),
					resource.TestCheckResourceAttr(resourceName, "response_actions.1.params.command", "isolate"),
					resource.TestCheckResourceAttr(resourceName, "response_actions.1.params.comment", "Isolate host due to new admin user activity from suspicious IP"),
				),
			},
		},
	})
}

func TestAccResourceSecurityDetectionRule_SavedQuery(t *testing.T) {
	resourceName := "elasticstack_kibana_security_detection_rule.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.Providers,
		CheckDestroy:             testAccCheckSecurityDetectionRuleDestroy,
		Steps: []resource.TestStep{
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minVersionSupport),
				Config:   testAccSecurityDetectionRuleConfig_savedQuery("test-saved-query-rule"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", "test-saved-query-rule"),
					resource.TestCheckResourceAttr(resourceName, "type", "saved_query"),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "description", "Test saved query security detection rule"),
					resource.TestCheckResourceAttr(resourceName, "severity", "low"),
					resource.TestCheckResourceAttr(resourceName, "risk_score", "30"),
					resource.TestCheckResourceAttr(resourceName, "saved_id", "test-saved-query-id"),

					// Check meta field
					checkResourceJSONAttr(resourceName, "meta", `{"saved_query_type": "security", "custom_field": "test_value", "query_origin": "saved"}`),

					resource.TestCheckResourceAttr(resourceName, "index.0", "logs-*"),
					resource.TestCheckResourceAttr(resourceName, "data_view_id", "saved-query-data-view-id"),
					resource.TestCheckResourceAttr(resourceName, "namespace", "saved-query-namespace"),
					resource.TestCheckResourceAttr(resourceName, "rule_name_override", "Custom Saved Query Rule Name"),
					resource.TestCheckResourceAttr(resourceName, "timestamp_override", "event.start"),
					resource.TestCheckResourceAttr(resourceName, "timestamp_override_fallback_disabled", "false"),

					// Check risk score mapping
					resource.TestCheckResourceAttr(resourceName, "risk_score_mapping.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "risk_score_mapping.0.field", "event.category"),
					resource.TestCheckResourceAttr(resourceName, "risk_score_mapping.0.operator", "equals"),
					resource.TestCheckResourceAttr(resourceName, "risk_score_mapping.0.value", "authentication"),
					resource.TestCheckResourceAttr(resourceName, "risk_score_mapping.0.risk_score", "45"),

					// Check investigation fields
					resource.TestCheckResourceAttr(resourceName, "investigation_fields.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "investigation_fields.0", "event.category"),
					resource.TestCheckResourceAttr(resourceName, "investigation_fields.1", "event.action"),

					// Check related integrations
					resource.TestCheckResourceAttr(resourceName, "related_integrations.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "related_integrations.0.package", "system"),
					resource.TestCheckResourceAttr(resourceName, "related_integrations.0.version", "1.0.0"),
					resource.TestCheckResourceAttr(resourceName, "related_integrations.0.integration", "logs"),

					// Check required fields
					resource.TestCheckResourceAttr(resourceName, "required_fields.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "required_fields.0.name", "event.category"),
					resource.TestCheckResourceAttr(resourceName, "required_fields.0.type", "keyword"),
					resource.TestCheckResourceAttr(resourceName, "required_fields.0.ecs", "true"),
					resource.TestCheckResourceAttr(resourceName, "required_fields.1.name", "event.action"),
					resource.TestCheckResourceAttr(resourceName, "required_fields.1.type", "keyword"),
					resource.TestCheckResourceAttr(resourceName, "required_fields.1.ecs", "true"),

					// Check severity mapping
					resource.TestCheckResourceAttr(resourceName, "severity_mapping.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "severity_mapping.0.field", "event.category"),
					resource.TestCheckResourceAttr(resourceName, "severity_mapping.0.operator", "equals"),
					resource.TestCheckResourceAttr(resourceName, "severity_mapping.0.value", "authentication"),
					resource.TestCheckResourceAttr(resourceName, "severity_mapping.0.severity", "low"),

					// Check response actions
					resource.TestCheckResourceAttr(resourceName, "response_actions.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "response_actions.0.action_type_id", ".osquery"),
					resource.TestCheckResourceAttr(resourceName, "response_actions.0.params.query", "SELECT * FROM logged_in_users WHERE user = '{{user.name}}';"),
					resource.TestCheckResourceAttr(resourceName, "response_actions.0.params.timeout", "250"),
					resource.TestCheckResourceAttr(resourceName, "response_actions.0.params.ecs_mapping.event.category", "category"),
					resource.TestCheckResourceAttr(resourceName, "response_actions.0.params.ecs_mapping.event.action", "action"),
					resource.TestCheckResourceAttr(resourceName, "response_actions.0.params.ecs_mapping.user.name", "username"),

					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "rule_id"),
				),
			},
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minVersionSupport),
				Config:   testAccSecurityDetectionRuleConfig_savedQueryUpdate("test-saved-query-rule-updated"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", "test-saved-query-rule-updated"),
					resource.TestCheckResourceAttr(resourceName, "query", "event.action:*"),
					resource.TestCheckResourceAttr(resourceName, "description", "Updated test saved query security detection rule"),
					resource.TestCheckResourceAttr(resourceName, "severity", "medium"),
					resource.TestCheckResourceAttr(resourceName, "risk_score", "60"),
					resource.TestCheckResourceAttr(resourceName, "saved_id", "test-saved-query-id-updated"),

					// Check meta field (updated values)
					checkResourceJSONAttr(resourceName, "meta", `{"saved_query_type": "security", "custom_field": "updated_value", "query_origin": "saved", "updated": "yes"}`),

					resource.TestCheckResourceAttr(resourceName, "index.0", "logs-*"),
					resource.TestCheckResourceAttr(resourceName, "index.1", "audit-*"),
					resource.TestCheckResourceAttr(resourceName, "data_view_id", "updated-saved-query-data-view-id"),
					resource.TestCheckResourceAttr(resourceName, "namespace", "updated-saved-query-namespace"),
					resource.TestCheckResourceAttr(resourceName, "rule_name_override", "Updated Custom Saved Query Rule Name"),
					resource.TestCheckResourceAttr(resourceName, "timestamp_override", "event.end"),
					resource.TestCheckResourceAttr(resourceName, "timestamp_override_fallback_disabled", "true"),

					// Check risk score mapping
					resource.TestCheckResourceAttr(resourceName, "risk_score_mapping.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "risk_score_mapping.0.field", "event.type"),
					resource.TestCheckResourceAttr(resourceName, "risk_score_mapping.0.operator", "equals"),
					resource.TestCheckResourceAttr(resourceName, "risk_score_mapping.0.value", "access"),
					resource.TestCheckResourceAttr(resourceName, "risk_score_mapping.0.risk_score", "70"),

					// Check investigation fields
					resource.TestCheckResourceAttr(resourceName, "investigation_fields.#", "3"),
					resource.TestCheckResourceAttr(resourceName, "investigation_fields.0", "host.name"),
					resource.TestCheckResourceAttr(resourceName, "investigation_fields.1", "user.name"),
					resource.TestCheckResourceAttr(resourceName, "investigation_fields.2", "process.name"),

					// Check related integrations
					resource.TestCheckResourceAttr(resourceName, "related_integrations.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "related_integrations.0.package", "system"),
					resource.TestCheckResourceAttr(resourceName, "related_integrations.0.version", "2.0.0"),
					resource.TestCheckResourceAttr(resourceName, "related_integrations.0.integration", "logs"),

					// Check required fields
					resource.TestCheckResourceAttr(resourceName, "required_fields.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "required_fields.0.name", "event.type"),
					resource.TestCheckResourceAttr(resourceName, "required_fields.0.type", "keyword"),
					resource.TestCheckResourceAttr(resourceName, "required_fields.0.ecs", "true"),
					resource.TestCheckResourceAttr(resourceName, "required_fields.1.name", "host.name"),
					resource.TestCheckResourceAttr(resourceName, "required_fields.1.type", "keyword"),
					resource.TestCheckResourceAttr(resourceName, "required_fields.1.ecs", "true"),

					// Check severity mapping
					resource.TestCheckResourceAttr(resourceName, "severity_mapping.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "severity_mapping.0.field", "event.type"),
					resource.TestCheckResourceAttr(resourceName, "severity_mapping.0.operator", "equals"),
					resource.TestCheckResourceAttr(resourceName, "severity_mapping.0.value", "access"),
					resource.TestCheckResourceAttr(resourceName, "severity_mapping.0.severity", "medium"),

					// Check response actions
					resource.TestCheckResourceAttr(resourceName, "response_actions.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "response_actions.0.action_type_id", ".osquery"),
					resource.TestCheckResourceAttr(resourceName, "response_actions.0.params.pack_id", "access_investigation_pack"),
					resource.TestCheckResourceAttr(resourceName, "response_actions.0.params.timeout", "400"),
					resource.TestCheckResourceAttr(resourceName, "response_actions.0.params.ecs_mapping.event.type", "type"),
					resource.TestCheckResourceAttr(resourceName, "response_actions.0.params.ecs_mapping.host.name", "hostname"),
					resource.TestCheckResourceAttr(resourceName, "response_actions.0.params.ecs_mapping.user.name", "username"),
					resource.TestCheckResourceAttr(resourceName, "response_actions.0.params.queries.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "response_actions.0.params.queries.0.id", "access_query1"),
					resource.TestCheckResourceAttr(resourceName, "response_actions.0.params.queries.0.query", "SELECT * FROM users WHERE username = '{{user.name}}';"),
					resource.TestCheckResourceAttr(resourceName, "response_actions.0.params.queries.0.platform", "linux"),
					resource.TestCheckResourceAttr(resourceName, "response_actions.0.params.queries.0.version", "4.8.0"),
					resource.TestCheckResourceAttr(resourceName, "response_actions.0.params.queries.0.ecs_mapping.user.id", "uid"),

					resource.TestCheckResourceAttr(resourceName, "exceptions_list.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "exceptions_list.0.id", "saved-query-exception-1"),
					resource.TestCheckResourceAttr(resourceName, "exceptions_list.0.list_id", "saved-query-exceptions"),
					resource.TestCheckResourceAttr(resourceName, "exceptions_list.0.namespace_type", "agnostic"),
					resource.TestCheckResourceAttr(resourceName, "exceptions_list.0.type", "detection"),
				),
			},
		},
	})
}

func TestAccResourceSecurityDetectionRule_ThreatMatch(t *testing.T) {
	resourceName := "elasticstack_kibana_security_detection_rule.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.Providers,
		CheckDestroy:             testAccCheckSecurityDetectionRuleDestroy,
		Steps: []resource.TestStep{
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minVersionSupport),
				Config:   testAccSecurityDetectionRuleConfig_threatMatch("test-threat-match-rule"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", "test-threat-match-rule"),
					resource.TestCheckResourceAttr(resourceName, "type", "threat_match"),
					resource.TestCheckResourceAttr(resourceName, "query", "destination.ip:*"),
					resource.TestCheckResourceAttr(resourceName, "language", "kuery"),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "description", "Test threat match security detection rule"),
					resource.TestCheckResourceAttr(resourceName, "severity", "high"),
					resource.TestCheckResourceAttr(resourceName, "risk_score", "80"),
					resource.TestCheckResourceAttr(resourceName, "index.0", "logs-*"),
					resource.TestCheckResourceAttr(resourceName, "data_view_id", "threat-match-data-view-id"),
					resource.TestCheckResourceAttr(resourceName, "namespace", "threat-match-namespace"),
					resource.TestCheckResourceAttr(resourceName, "rule_name_override", "Custom Threat Match Rule Name"),
					resource.TestCheckResourceAttr(resourceName, "timestamp_override", "threat.indicator.first_seen"),
					resource.TestCheckResourceAttr(resourceName, "timestamp_override_fallback_disabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "threat_index.0", "threat-intel-*"),
					resource.TestCheckResourceAttr(resourceName, "threat_query", "threat.indicator.type:ip"),
					resource.TestCheckResourceAttr(resourceName, "threat_mapping.0.entries.0.field", "destination.ip"),
					resource.TestCheckResourceAttr(resourceName, "threat_mapping.0.entries.0.type", "mapping"),
					resource.TestCheckResourceAttr(resourceName, "threat_mapping.0.entries.0.value", "threat.indicator.ip"),

					// Check meta field
					checkResourceJSONAttr(resourceName, "meta", `{"threat_type": "indicator_match", "custom_field": "test_value", "intelligence": "external"}`),

					// Check investigation_fields
					resource.TestCheckResourceAttr(resourceName, "investigation_fields.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "investigation_fields.0", "destination.ip"),
					resource.TestCheckResourceAttr(resourceName, "investigation_fields.1", "source.ip"),

					// Check related integrations
					resource.TestCheckResourceAttr(resourceName, "related_integrations.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "related_integrations.0.package", "threat_intel"),
					resource.TestCheckResourceAttr(resourceName, "related_integrations.0.version", "1.0.0"),
					resource.TestCheckResourceAttr(resourceName, "related_integrations.0.integration", "indicators"),

					// Check required fields
					resource.TestCheckResourceAttr(resourceName, "required_fields.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "required_fields.0.name", "destination.ip"),
					resource.TestCheckResourceAttr(resourceName, "required_fields.0.type", "ip"),
					resource.TestCheckResourceAttr(resourceName, "required_fields.0.ecs", "true"),
					resource.TestCheckResourceAttr(resourceName, "required_fields.1.name", "threat.indicator.ip"),
					resource.TestCheckResourceAttr(resourceName, "required_fields.1.type", "ip"),
					resource.TestCheckResourceAttr(resourceName, "required_fields.1.ecs", "true"),

					// Check severity mapping
					resource.TestCheckResourceAttr(resourceName, "severity_mapping.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "severity_mapping.0.field", "threat.indicator.confidence"),
					resource.TestCheckResourceAttr(resourceName, "severity_mapping.0.operator", "equals"),
					resource.TestCheckResourceAttr(resourceName, "severity_mapping.0.value", "high"),
					resource.TestCheckResourceAttr(resourceName, "severity_mapping.0.severity", "high"),

					// Check risk score mapping
					resource.TestCheckResourceAttr(resourceName, "risk_score_mapping.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "risk_score_mapping.0.field", "threat.indicator.confidence"),
					resource.TestCheckResourceAttr(resourceName, "risk_score_mapping.0.operator", "equals"),
					resource.TestCheckResourceAttr(resourceName, "risk_score_mapping.0.value", "medium"),
					resource.TestCheckResourceAttr(resourceName, "risk_score_mapping.0.risk_score", "85"),

					// Check response actions
					resource.TestCheckResourceAttr(resourceName, "response_actions.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "response_actions.0.action_type_id", ".osquery"),
					resource.TestCheckResourceAttr(resourceName, "response_actions.0.params.query", "SELECT * FROM listening_ports WHERE address = '{{destination.ip}}';"),
					resource.TestCheckResourceAttr(resourceName, "response_actions.0.params.timeout", "300"),
					resource.TestCheckResourceAttr(resourceName, "response_actions.0.params.ecs_mapping.destination.ip", "dest_ip"),
					resource.TestCheckResourceAttr(resourceName, "response_actions.0.params.ecs_mapping.threat.indicator.ip", "threat_ip"),
					resource.TestCheckResourceAttr(resourceName, "response_actions.0.params.ecs_mapping.threat.indicator.confidence", "confidence"),
					resource.TestCheckResourceAttr(resourceName, "response_actions.1.action_type_id", ".endpoint"),
					resource.TestCheckResourceAttr(resourceName, "response_actions.1.params.command", "isolate"),
					resource.TestCheckResourceAttr(resourceName, "response_actions.1.params.comment", "Isolate host due to threat match on destination IP"),

					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "rule_id"),
				),
			},
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minVersionSupport),
				Config:   testAccSecurityDetectionRuleConfig_threatMatchUpdate("test-threat-match-rule-updated"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", "test-threat-match-rule-updated"),
					resource.TestCheckResourceAttr(resourceName, "query", "destination.ip:* OR source.ip:*"),
					resource.TestCheckResourceAttr(resourceName, "description", "Updated test threat match security detection rule"),
					resource.TestCheckResourceAttr(resourceName, "severity", "critical"),
					resource.TestCheckResourceAttr(resourceName, "risk_score", "95"),
					resource.TestCheckResourceAttr(resourceName, "index.0", "logs-*"),
					resource.TestCheckResourceAttr(resourceName, "index.1", "network-*"),
					resource.TestCheckResourceAttr(resourceName, "data_view_id", "updated-threat-match-data-view-id"),
					resource.TestCheckResourceAttr(resourceName, "namespace", "updated-threat-match-namespace"),
					resource.TestCheckResourceAttr(resourceName, "rule_name_override", "Updated Custom Threat Match Rule Name"),
					resource.TestCheckResourceAttr(resourceName, "timestamp_override", "threat.indicator.last_seen"),
					resource.TestCheckResourceAttr(resourceName, "timestamp_override_fallback_disabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "threat_index.0", "threat-intel-*"),
					resource.TestCheckResourceAttr(resourceName, "threat_index.1", "ioc-*"),
					resource.TestCheckResourceAttr(resourceName, "threat_query", "threat.indicator.type:(ip OR domain)"),
					resource.TestCheckResourceAttr(resourceName, "threat_mapping.0.entries.0.field", "destination.ip"),
					resource.TestCheckResourceAttr(resourceName, "threat_mapping.1.entries.0.field", "source.ip"),

					// Check meta field (updated values)
					checkResourceJSONAttr(resourceName, "meta", `{"threat_type": "indicator_match", "custom_field": "updated_value", "intelligence": "external", "updated": "yes"}`),

					// Check investigation_fields
					resource.TestCheckResourceAttr(resourceName, "investigation_fields.#", "3"),
					resource.TestCheckResourceAttr(resourceName, "investigation_fields.0", "destination.ip"),
					resource.TestCheckResourceAttr(resourceName, "investigation_fields.1", "source.ip"),
					resource.TestCheckResourceAttr(resourceName, "investigation_fields.2", "threat.indicator.type"),

					// Check related integrations
					resource.TestCheckResourceAttr(resourceName, "related_integrations.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "related_integrations.0.package", "threat_intel"),
					resource.TestCheckResourceAttr(resourceName, "related_integrations.0.version", "2.0.0"),
					resource.TestCheckResourceAttr(resourceName, "related_integrations.0.integration", "indicators"),

					// Check required fields
					resource.TestCheckResourceAttr(resourceName, "required_fields.#", "3"),
					resource.TestCheckResourceAttr(resourceName, "required_fields.0.name", "destination.ip"),
					resource.TestCheckResourceAttr(resourceName, "required_fields.0.type", "ip"),
					resource.TestCheckResourceAttr(resourceName, "required_fields.0.ecs", "true"),
					resource.TestCheckResourceAttr(resourceName, "required_fields.1.name", "source.ip"),
					resource.TestCheckResourceAttr(resourceName, "required_fields.1.type", "ip"),
					resource.TestCheckResourceAttr(resourceName, "required_fields.1.ecs", "true"),
					resource.TestCheckResourceAttr(resourceName, "required_fields.2.name", "threat.indicator.ip"),
					resource.TestCheckResourceAttr(resourceName, "required_fields.2.type", "ip"),
					resource.TestCheckResourceAttr(resourceName, "required_fields.2.ecs", "true"),

					// Check severity mapping
					resource.TestCheckResourceAttr(resourceName, "severity_mapping.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "severity_mapping.0.field", "threat.indicator.confidence"),
					resource.TestCheckResourceAttr(resourceName, "severity_mapping.0.operator", "equals"),
					resource.TestCheckResourceAttr(resourceName, "severity_mapping.0.value", "high"),
					resource.TestCheckResourceAttr(resourceName, "severity_mapping.0.severity", "critical"),

					// Check risk score mapping
					resource.TestCheckResourceAttr(resourceName, "risk_score_mapping.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "risk_score_mapping.0.field", "threat.indicator.confidence"),
					resource.TestCheckResourceAttr(resourceName, "risk_score_mapping.0.operator", "equals"),
					resource.TestCheckResourceAttr(resourceName, "risk_score_mapping.0.value", "high"),
					resource.TestCheckResourceAttr(resourceName, "risk_score_mapping.0.risk_score", "100"),

					// Check response actions
					resource.TestCheckResourceAttr(resourceName, "response_actions.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "response_actions.0.action_type_id", ".osquery"),
					resource.TestCheckResourceAttr(resourceName, "response_actions.0.params.saved_query_id", "threat_intel_investigation"),
					resource.TestCheckResourceAttr(resourceName, "response_actions.0.params.timeout", "450"),
					resource.TestCheckResourceAttr(resourceName, "response_actions.0.params.ecs_mapping.source.ip", "src_ip"),
					resource.TestCheckResourceAttr(resourceName, "response_actions.0.params.ecs_mapping.destination.ip", "dest_ip"),
					resource.TestCheckResourceAttr(resourceName, "response_actions.0.params.ecs_mapping.threat.indicator.type", "threat_type"),
					resource.TestCheckResourceAttr(resourceName, "response_actions.1.action_type_id", ".endpoint"),
					resource.TestCheckResourceAttr(resourceName, "response_actions.1.params.command", "kill-process"),
					resource.TestCheckResourceAttr(resourceName, "response_actions.1.params.comment", "Kill processes communicating with known threat indicators"),
					resource.TestCheckResourceAttr(resourceName, "response_actions.1.params.config.field", "process.entity_id"),
					resource.TestCheckResourceAttr(resourceName, "response_actions.1.params.config.overwrite", "true"),
				),
			},
		},
	})
}

func TestAccResourceSecurityDetectionRule_Threshold(t *testing.T) {
	resourceName := "elasticstack_kibana_security_detection_rule.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.Providers,
		CheckDestroy:             testAccCheckSecurityDetectionRuleDestroy,
		Steps: []resource.TestStep{
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minVersionSupport),
				Config:   testAccSecurityDetectionRuleConfig_threshold("test-threshold-rule"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", "test-threshold-rule"),
					resource.TestCheckResourceAttr(resourceName, "type", "threshold"),
					resource.TestCheckResourceAttr(resourceName, "query", "event.action:login"),
					resource.TestCheckResourceAttr(resourceName, "language", "kuery"),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "description", "Test threshold security detection rule"),
					resource.TestCheckResourceAttr(resourceName, "severity", "medium"),
					resource.TestCheckResourceAttr(resourceName, "risk_score", "55"),
					resource.TestCheckResourceAttr(resourceName, "index.0", "logs-*"),
					resource.TestCheckResourceAttr(resourceName, "data_view_id", "threshold-data-view-id"),
					resource.TestCheckResourceAttr(resourceName, "namespace", "threshold-namespace"),
					resource.TestCheckResourceAttr(resourceName, "rule_name_override", "Custom Threshold Rule Name"),
					resource.TestCheckResourceAttr(resourceName, "timestamp_override", "event.created"),
					resource.TestCheckResourceAttr(resourceName, "timestamp_override_fallback_disabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "threshold.value", "10"),
					resource.TestCheckResourceAttr(resourceName, "threshold.field.0", "user.name"),

					// Check meta field
					checkResourceJSONAttr(resourceName, "meta", `{"threshold_type": "count_based", "custom_field": "test_value", "monitoring": "enabled"}`),

					// Check investigation_fields
					resource.TestCheckResourceAttr(resourceName, "investigation_fields.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "investigation_fields.0", "user.name"),
					resource.TestCheckResourceAttr(resourceName, "investigation_fields.1", "event.action"),

					// Check related integrations
					resource.TestCheckResourceAttr(resourceName, "related_integrations.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "related_integrations.0.package", "system"),
					resource.TestCheckResourceAttr(resourceName, "related_integrations.0.version", "1.0.0"),
					resource.TestCheckResourceAttr(resourceName, "related_integrations.0.integration", "auth"),

					// Check required fields
					resource.TestCheckResourceAttr(resourceName, "required_fields.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "required_fields.0.name", "event.action"),
					resource.TestCheckResourceAttr(resourceName, "required_fields.0.type", "keyword"),
					resource.TestCheckResourceAttr(resourceName, "required_fields.0.ecs", "true"),
					resource.TestCheckResourceAttr(resourceName, "required_fields.1.name", "user.name"),
					resource.TestCheckResourceAttr(resourceName, "required_fields.1.type", "keyword"),
					resource.TestCheckResourceAttr(resourceName, "required_fields.1.ecs", "true"),

					// Check severity mapping
					resource.TestCheckResourceAttr(resourceName, "severity_mapping.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "severity_mapping.0.field", "event.outcome"),
					resource.TestCheckResourceAttr(resourceName, "severity_mapping.0.operator", "equals"),
					resource.TestCheckResourceAttr(resourceName, "severity_mapping.0.value", "success"),
					resource.TestCheckResourceAttr(resourceName, "severity_mapping.0.severity", "medium"),

					// Check risk score mapping
					resource.TestCheckResourceAttr(resourceName, "risk_score_mapping.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "risk_score_mapping.0.field", "event.outcome"),
					resource.TestCheckResourceAttr(resourceName, "risk_score_mapping.0.operator", "equals"),
					resource.TestCheckResourceAttr(resourceName, "risk_score_mapping.0.value", "success"),
					resource.TestCheckResourceAttr(resourceName, "risk_score_mapping.0.risk_score", "45"),

					// Check response actions
					resource.TestCheckResourceAttr(resourceName, "response_actions.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "response_actions.0.action_type_id", ".osquery"),
					resource.TestCheckResourceAttr(resourceName, "response_actions.0.params.query", "SELECT * FROM logged_in_users WHERE user = '{{user.name}}' ORDER BY time DESC LIMIT 10;"),
					resource.TestCheckResourceAttr(resourceName, "response_actions.0.params.timeout", "200"),
					resource.TestCheckResourceAttr(resourceName, "response_actions.0.params.ecs_mapping.user.name", "username"),
					resource.TestCheckResourceAttr(resourceName, "response_actions.0.params.ecs_mapping.event.action", "action"),
					resource.TestCheckResourceAttr(resourceName, "response_actions.0.params.ecs_mapping.event.outcome", "outcome"),

					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "rule_id"),
				),
			},
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minVersionSupport),
				Config:   testAccSecurityDetectionRuleConfig_thresholdUpdate("test-threshold-rule-updated"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", "test-threshold-rule-updated"),
					resource.TestCheckResourceAttr(resourceName, "query", "event.action:(login OR logout)"),
					resource.TestCheckResourceAttr(resourceName, "description", "Updated test threshold security detection rule"),
					resource.TestCheckResourceAttr(resourceName, "severity", "high"),
					resource.TestCheckResourceAttr(resourceName, "risk_score", "75"),
					resource.TestCheckResourceAttr(resourceName, "index.0", "logs-*"),
					resource.TestCheckResourceAttr(resourceName, "index.1", "audit-*"),
					resource.TestCheckResourceAttr(resourceName, "data_view_id", "updated-threshold-data-view-id"),
					resource.TestCheckResourceAttr(resourceName, "namespace", "updated-threshold-namespace"),
					resource.TestCheckResourceAttr(resourceName, "rule_name_override", "Updated Custom Threshold Rule Name"),
					resource.TestCheckResourceAttr(resourceName, "timestamp_override", "event.start"),
					resource.TestCheckResourceAttr(resourceName, "timestamp_override_fallback_disabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "threshold.value", "20"),
					resource.TestCheckResourceAttr(resourceName, "threshold.field.0", "user.name"),
					resource.TestCheckResourceAttr(resourceName, "threshold.field.1", "source.ip"),

					// Check meta field (updated values)
					checkResourceJSONAttr(resourceName, "meta", `{"threshold_type": "count_based", "custom_field": "updated_value", "monitoring": "enabled", "updated": "yes"}`),

					// Check investigation_fields
					resource.TestCheckResourceAttr(resourceName, "investigation_fields.#", "3"),
					resource.TestCheckResourceAttr(resourceName, "investigation_fields.0", "user.name"),
					resource.TestCheckResourceAttr(resourceName, "investigation_fields.1", "source.ip"),
					resource.TestCheckResourceAttr(resourceName, "investigation_fields.2", "event.outcome"),

					// Check related integrations
					resource.TestCheckResourceAttr(resourceName, "related_integrations.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "related_integrations.0.package", "system"),
					resource.TestCheckResourceAttr(resourceName, "related_integrations.0.version", "2.0.0"),
					resource.TestCheckResourceAttr(resourceName, "related_integrations.0.integration", "auth"),

					// Check required fields
					resource.TestCheckResourceAttr(resourceName, "required_fields.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "required_fields.0.name", "event.action"),
					resource.TestCheckResourceAttr(resourceName, "required_fields.0.type", "keyword"),
					resource.TestCheckResourceAttr(resourceName, "required_fields.0.ecs", "true"),
					resource.TestCheckResourceAttr(resourceName, "required_fields.1.name", "source.ip"),
					resource.TestCheckResourceAttr(resourceName, "required_fields.1.type", "ip"),
					resource.TestCheckResourceAttr(resourceName, "required_fields.1.ecs", "true"),

					// Check severity mapping
					resource.TestCheckResourceAttr(resourceName, "severity_mapping.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "severity_mapping.0.field", "event.outcome"),
					resource.TestCheckResourceAttr(resourceName, "severity_mapping.0.operator", "equals"),
					resource.TestCheckResourceAttr(resourceName, "severity_mapping.0.value", "failure"),
					resource.TestCheckResourceAttr(resourceName, "severity_mapping.0.severity", "high"),

					// Check risk score mapping
					resource.TestCheckResourceAttr(resourceName, "risk_score_mapping.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "risk_score_mapping.0.field", "event.outcome"),
					resource.TestCheckResourceAttr(resourceName, "risk_score_mapping.0.operator", "equals"),
					resource.TestCheckResourceAttr(resourceName, "risk_score_mapping.0.value", "failure"),
					resource.TestCheckResourceAttr(resourceName, "risk_score_mapping.0.risk_score", "90"),

					// Check response actions
					resource.TestCheckResourceAttr(resourceName, "response_actions.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "response_actions.0.action_type_id", ".osquery"),
					resource.TestCheckResourceAttr(resourceName, "response_actions.0.params.pack_id", "login_failure_investigation"),
					resource.TestCheckResourceAttr(resourceName, "response_actions.0.params.timeout", "350"),
					resource.TestCheckResourceAttr(resourceName, "response_actions.0.params.ecs_mapping.event.outcome", "outcome"),
					resource.TestCheckResourceAttr(resourceName, "response_actions.0.params.ecs_mapping.source.ip", "source_ip"),
					resource.TestCheckResourceAttr(resourceName, "response_actions.0.params.ecs_mapping.user.name", "username"),
					resource.TestCheckResourceAttr(resourceName, "response_actions.0.params.queries.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "response_actions.0.params.queries.0.id", "failed_login_query"),
					resource.TestCheckResourceAttr(resourceName, "response_actions.0.params.queries.0.query", "SELECT * FROM last WHERE type = 7 AND username = '{{user.name}}';"),
					resource.TestCheckResourceAttr(resourceName, "response_actions.0.params.queries.0.platform", "linux"),
					resource.TestCheckResourceAttr(resourceName, "response_actions.0.params.queries.0.version", "4.9.0"),
					resource.TestCheckResourceAttr(resourceName, "response_actions.1.action_type_id", ".endpoint"),
					resource.TestCheckResourceAttr(resourceName, "response_actions.1.params.command", "isolate"),
					resource.TestCheckResourceAttr(resourceName, "response_actions.1.params.comment", "Isolate host due to multiple failed login attempts"),
				),
			},
		},
	})
}

func testAccCheckSecurityDetectionRuleDestroy(s *terraform.State) error {
	client, err := clients.NewAcceptanceTestingClient()
	if err != nil {
		return err
	}

	kbClient, err := client.GetKibanaOapiClient()
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		switch rs.Type {
		case "elasticstack_kibana_security_detection_rule":
			// Parse ID to get space_id and rule_id
			parts := strings.Split(rs.Primary.ID, "/")
			if len(parts) != 2 {
				return fmt.Errorf("invalid resource ID format: %s", rs.Primary.ID)
			}
			ruleId := parts[1]

			// Check if the rule still exists
			ruleObjectId := kbapi.SecurityDetectionsAPIRuleObjectId(uuid.MustParse(ruleId))
			params := &kbapi.ReadRuleParams{
				Id: &ruleObjectId,
			}

			response, err := kbClient.API.ReadRuleWithResponse(context.Background(), params)
			if err != nil {
				return fmt.Errorf("failed to read security detection rule: %v", err)
			}

			// If the rule still exists (status 200), it means destroy failed
			if response.StatusCode() == 200 {
				return fmt.Errorf("security detection rule (%s) still exists", ruleId)
			}

			// If we get a 404, that's expected - the rule was properly destroyed
			// Any other status code indicates an error
			if response.StatusCode() != 404 {
				return fmt.Errorf("unexpected status code when checking security detection rule: %d", response.StatusCode())
			}

		case "elasticstack_kibana_action_connector":
			// Parse ID to get space_id and connector_id
			compId, _ := clients.CompositeIdFromStr(rs.Primary.ID)

			// Get connector client from the Kibana OAPI client
			oapiClient, err := client.GetKibanaOapiClient()
			if err != nil {
				return err
			}

			connector, diags := kibana_oapi.GetConnector(context.Background(), oapiClient, compId.ResourceId, compId.ClusterId)
			if diags.HasError() {
				return fmt.Errorf("failed to get connector: %v", diags)
			}

			if connector != nil {
				return fmt.Errorf("action connector (%s) still exists", compId.ResourceId)
			}
		}
	}

	return nil
}

func testAccSecurityDetectionRuleConfig_query(name string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_security_detection_rule" "test" {
  name         = "%s"
  type         = "query"
  query        = "*:*"
  language     = "kuery"
  enabled      = true
  description  = "Test query security detection rule"
  severity     = "medium"
  risk_score   = 50
  from         = "now-6m"
  to           = "now"
  interval     = "5m"
  index        = ["logs-*"]
  data_view_id = "test-data-view-id"
  namespace    = "test-namespace"
  rule_name_override = "Custom Query Rule Name"
  timestamp_override = "@timestamp"
  timestamp_override_fallback_disabled = true

  meta = jsonencode({
    "custom_field" = "test_value"
    "environment"  = "testing"
    "version"      = "1.0"
  })

  investigation_fields = ["user.name", "event.action"]

  risk_score_mapping = [
    {
      field      = "event.severity"
      operator   = "equals"
      value      = "high"
      risk_score = 85
    }
  ]

  related_integrations = [
    {
      package     = "windows"
      version     = "1.0.0"
      integration = "system"
    }
  ]

  required_fields = [
    {
      name = "event.type"
      type = "keyword"
    },
    {
      name = "host.os.type"
      type = "keyword"
    }
  ]

  severity_mapping = [
    {
      field    = "event.severity_level"
      operator = "equals"
      value    = "critical"
      severity = "critical"
    }
  ]

  response_actions = [
    {
      action_type_id = ".osquery"
      params = {
        query   = "SELECT * FROM processes WHERE name = 'malicious.exe';"
        timeout = 300
        ecs_mapping = {
          "process.name" = "name"
          "process.pid"  = "pid"
        }
      }
    },
    {
      action_type_id = ".endpoint"
      params = {
        command = "isolate"
        comment = "Isolate host due to suspicious activity"
      }
    }
  ]
}
`, name)
}

func testAccSecurityDetectionRuleConfig_queryUpdate(name string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_security_detection_rule" "test" {
  name        = "%s"
  type        = "query"
  query       = "*:*"
  language    = "kuery"
  enabled     = true
  description = "Updated test query security detection rule"
  severity    = "high"
  risk_score  = 75
  from        = "now-6m"
  to          = "now"
  interval    = "5m"
  index       = ["logs-*"]
  author      = ["Test Author"]
  tags        = ["test", "automation"]
  license     = "Elastic License v2"
  data_view_id = "updated-data-view-id"
  namespace    = "updated-namespace"
  rule_name_override = "Updated Custom Query Rule Name"
  timestamp_override = "event.ingested"
  timestamp_override_fallback_disabled = false

  meta = jsonencode({
    "custom_field" = "updated_value"
    "environment"  = "production"
    "version"      = "2.0"
    "team"         = "security"
  })

  investigation_fields = ["user.name", "event.action", "source.ip"]

  risk_score_mapping = [
    {
      field      = "event.risk_level"
      operator   = "equals"
      value      = "critical"
      risk_score = 95
    }
  ]

  related_integrations = [
    {
      package     = "linux"
      version     = "2.0.0"
      integration = "auditd"
    },
    {
      package     = "network"
      version     = "1.5.0"
    }
  ]

  required_fields = [
    {
      name = "event.category"
      type = "keyword"
    },
    {
      name = "process.name"
      type = "keyword"
    },
    {
      name = "custom.field"
      type = "text"
    }
  ]

  severity_mapping = [
    {
      field    = "alert.severity"
      operator = "equals"
      value    = "high"
      severity = "high"
    },
    {
      field    = "alert.severity"
      operator = "equals"
      value    = "medium"
      severity = "medium"
    }
  ]

  response_actions = [
    {
      action_type_id = ".osquery"
      params = {
        pack_id = "incident_response_pack"
        timeout = 600
        ecs_mapping = {
          "host.name"    = "hostname"
          "user.name"    = "username"
          "process.name" = "process_name"
        }
        queries = [
          {
            id       = "query1"
            query    = "SELECT * FROM logged_in_users;"
            platform = "linux"
            version  = "4.6.0"
          },
          {
            id       = "query2"
            query    = "SELECT * FROM processes WHERE state = 'R';"
            platform = "linux"
            version  = "4.6.0"
            ecs_mapping = {
              "process.pid" = "pid"
              "process.command_line" = "cmdline"
            }
          }
        ]
      }
    },
    {
      action_type_id = ".endpoint"
      params = {
        command = "kill-process"
        comment = "Kill suspicious process identified during investigation"
        config = {
          field     = "process.entity_id"
          overwrite = true
        }
      }
    }
  ]
}
`, name)
}

func testAccSecurityDetectionRuleConfig_eql(name string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_security_detection_rule" "test" {
  name             = "%s"
  type             = "eql"
  query            = "process where process.name == \"cmd.exe\""
  language         = "eql"
  enabled          = true
  description      = "Test EQL security detection rule"
  severity         = "high"
  risk_score       = 70
  from             = "now-6m"
  to               = "now"
  interval         = "5m"
  index            = ["winlogbeat-*"]
  tiebreaker_field = "@timestamp"
  data_view_id     = "eql-data-view-id"
  namespace        = "eql-namespace"
  rule_name_override = "Custom EQL Rule Name"
  timestamp_override = "process.start"
  timestamp_override_fallback_disabled = false

  meta = jsonencode({
    "rule_type"   = "eql"
    "process"     = "monitoring"
    "severity"    = "high"
  })

  investigation_fields = ["process.name", "process.executable"]

  risk_score_mapping = [
    {
      field      = "process.executable"
      operator   = "equals"
      value      = "C:\\Windows\\System32\\cmd.exe"
      risk_score = 75
    }
  ]

  related_integrations = [
    {
      package     = "windows"
      version     = "1.0.0"
      integration = "system"
    }
  ]

  required_fields = [
    {
      name = "process.name"
      type = "keyword"
    },
    {
      name = "event.type"
      type = "keyword"
    }
  ]

  severity_mapping = [
    {
      field    = "event.severity_level"
      operator = "equals"
      value    = "high"
      severity = "high"
    }
  ]

  response_actions = [
    {
      action_type_id = ".osquery"
      params = {
        saved_query_id = "suspicious_processes"
        timeout        = 300
      }
    }
  ]
}
`, name)
}

func testAccSecurityDetectionRuleConfig_eqlUpdate(name string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_security_detection_rule" "test" {
  name             = "%s"
  type             = "eql"
  query            = "process where process.name == \"powershell.exe\""
  language         = "eql"
  enabled          = true
  description      = "Updated test EQL security detection rule"
  severity         = "critical"
  risk_score       = 90
  from             = "now-6m"
  to               = "now"
  interval         = "5m"
  index            = ["winlogbeat-*"]
  tiebreaker_field = "@timestamp"
  author           = ["Test Author"]
  tags             = ["test", "eql", "automation"]
  license          = "Elastic License v2"
  rule_name_override = "Updated Custom EQL Rule Name"
  timestamp_override = "process.end"
  timestamp_override_fallback_disabled = true

  meta = jsonencode({
    "rule_type"   = "eql"
    "process"     = "detection"
    "severity"    = "critical"
    "updated"     = "true"
  })

  investigation_fields = ["process.name", "process.executable", "process.parent.name"]

  risk_score_mapping = [
    {
      field      = "process.parent.name"
      operator   = "equals"
      value      = "cmd.exe"
      risk_score = 95
    }
  ]

  related_integrations = [
    {
      package     = "windows"
      version     = "2.0.0"
      integration = "system"
    }
  ]

  required_fields = [
    {
      name = "process.parent.name"
      type = "keyword"
    },
    {
      name = "event.category"
      type = "keyword"
    }
  ]

  severity_mapping = [
    {
      field    = "event.severity_level"
      operator = "equals"
      value    = "critical"
      severity = "critical"
    }
  ]

  response_actions = [
    {
      action_type_id = ".osquery"
      params = {
        pack_id = "eql_response_pack"
        timeout = 450
        ecs_mapping = {
          "process.executable" = "executable_path"
          "process.parent.name" = "parent_name"
        }
      }
    }
  ]
}
`, name)
}

func testAccSecurityDetectionRuleConfig_esql(name string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_security_detection_rule" "test" {
  name        = "%s"
  type        = "esql"
  query       = "FROM logs-* | WHERE event.action == \"login\" | STATS count(*) BY user.name"
  language    = "esql"
  enabled     = true
  description = "Test ESQL security detection rule"
  severity    = "medium"
  risk_score  = 60
  from        = "now-6m"
  to          = "now"
  interval    = "5m"
  namespace   = "esql-namespace"
  rule_name_override = "Custom ESQL Rule Name"
  timestamp_override = "event.created"
  timestamp_override_fallback_disabled = true

  meta = jsonencode({
    "query_type" = "esql"
    "analytics"  = "enabled"
    "phase"      = "testing"
  })

  investigation_fields = ["user.name", "user.domain"]

  risk_score_mapping = [
    {
      field      = "user.domain"
      operator   = "equals"
      value      = "admin"
      risk_score = 80
    }
  ]

  related_integrations = [
    {
      package     = "system"
      version     = "1.0.0"
      integration = "auth"
    }
  ]

  required_fields = [
    {
      name = "user.name"
      type = "keyword"
    },
    {
      name = "event.action"
      type = "keyword"
    }
  ]

  severity_mapping = [
    {
      field    = "user.domain"
      operator = "equals"
      value    = "admin"
      severity = "high"
    }
  ]

  response_actions = [
    {
      action_type_id = ".osquery"
      params = {
        query   = "SELECT * FROM users WHERE username LIKE '%%admin%%';"
        timeout = 400
        ecs_mapping = {
          "user.name"   = "username"
          "user.domain" = "domain"
        }
      }
    },
    {
      action_type_id = ".endpoint"
      params = {
        command = "isolate"
        comment = "Isolate host due to suspicious admin activity"
      }
    }
  ]
}
`, name)
}

func testAccSecurityDetectionRuleConfig_esqlUpdate(name string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_security_detection_rule" "test" {
  name        = "%s"
  type        = "esql"
  query       = "FROM logs-* | WHERE event.action == \"logout\" | STATS count(*) BY user.name, source.ip"
  language    = "esql"
  enabled     = true
  description = "Updated test ESQL security detection rule"
  severity    = "high"
  risk_score  = 80
  from        = "now-6m"
  to          = "now"
  interval    = "5m"
  author      = ["Test Author"]
  tags        = ["test", "esql", "automation"]
  license     = "Elastic License v2"
  rule_name_override = "Updated Custom ESQL Rule Name"
  timestamp_override = "event.start"
  timestamp_override_fallback_disabled = false

  meta = jsonencode({
    "query_type" = "esql"
    "analytics"  = "enabled"
    "phase"      = "production"
    "updated"    = "yes"
  })
  
  investigation_fields = ["user.name", "user.domain", "event.outcome"]
  
  risk_score_mapping = [
    {
      field      = "event.outcome"
      operator   = "equals"
      value      = "failure"
      risk_score = 95
    }
  ]
  
  related_integrations = [
    {
      package     = "system"
      version     = "2.0.0"
      integration = "auth"
    }
  ]

  required_fields = [
    {
      name = "user.name"
      type = "keyword"
    },
    {
      name = "event.outcome"
      type = "keyword"
    }
  ]

  severity_mapping = [
    {
      field    = "event.outcome"
      operator = "equals"
      value    = "failure"
      severity = "critical"
    }
  ]
  
  response_actions = [
    {
      action_type_id = ".osquery"
      params = {
        saved_query_id = "failed_login_investigation"
        timeout        = 500
        ecs_mapping = {
          "event.outcome" = "outcome"
          "user.name"     = "username"
          "source.ip"     = "source_ip"
        }
      }
    }
  ]
  
  exceptions_list = [
    {
      id             = "esql-exception-1"
      list_id        = "esql-rule-exceptions"
      namespace_type = "single"
      type           = "detection"
    }
  ]
}
`, name)
}

func testAccSecurityDetectionRuleConfig_machineLearning(name string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_security_detection_rule" "test" {
  name                     = "%s"
  type                     = "machine_learning"
  enabled                  = true
  description              = "Test ML security detection rule"
  severity                 = "critical"
  risk_score               = 90
  from                     = "now-6m"
  to                       = "now"
  interval                 = "5m"
  anomaly_threshold        = 75
  machine_learning_job_id  = ["test-ml-job"]
  namespace                = "ml-namespace"
  rule_name_override = "Custom ML Rule Name"
  timestamp_override = "ml.job_id"
  timestamp_override_fallback_disabled = false

  meta = jsonencode({
    "ml_type"    = "anomaly_detection"
    "custom_ml"  = "test_value"
    "threshold"  = "75"
  })

  investigation_fields = ["ml.anomaly_score", "ml.job_id"]

  risk_score_mapping = [
    {
      field      = "ml.anomaly_score"
      operator   = "equals"
      value      = "critical"
      risk_score = 100
    }
  ]

  related_integrations = [
    {
      package     = "ml"
      version     = "1.0.0"
      integration = "anomaly_detection"
    }
  ]

  required_fields = [
    {
      name = "ml.anomaly_score"
      type = "double"
    },
    {
      name = "ml.job_id"
      type = "keyword"
    }
  ]

  severity_mapping = [
    {
      field    = "ml.anomaly_score"
      operator = "equals"
      value    = "critical"
      severity = "critical"
    }
  ]

  response_actions = [
    {
      action_type_id = ".osquery"
      params = {
        query   = "SELECT * FROM processes WHERE pid IN (SELECT DISTINCT pid FROM connections WHERE remote_address NOT LIKE '10.%%' AND remote_address NOT LIKE '192.168.%%' AND remote_address NOT LIKE '127.%%');"
        timeout = 600
        ecs_mapping = {
          "process.pid"        = "pid"
          "process.name"       = "name"
          "ml.anomaly_score"   = "anomaly_score"
        }
      }
    }
  ]
}
`, name)
}

func testAccSecurityDetectionRuleConfig_machineLearningUpdate(name string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_security_detection_rule" "test" {
  name                     = "%s"
  type                     = "machine_learning"
  enabled                  = true
  description              = "Updated test ML security detection rule"
  severity                 = "high"
  risk_score               = 85
  from                     = "now-6m"
  to                       = "now"
  interval                 = "5m"
  anomaly_threshold        = 80
  machine_learning_job_id  = ["test-ml-job", "test-ml-job-2"]
  author                   = ["Test Author"]
  tags                     = ["test", "ml", "automation"]
  license                  = "Elastic License v2"
  rule_name_override = "Updated Custom ML Rule Name"
  timestamp_override = "ml.anomaly_score"
  timestamp_override_fallback_disabled = true

  meta = jsonencode({
    "ml_type"    = "anomaly_detection"
    "custom_ml"  = "updated_value"
    "threshold"  = "80"
    "updated"    = "yes"
  })

  investigation_fields = ["ml.anomaly_score", "ml.job_id", "ml.is_anomaly"]

  risk_score_mapping = [
    {
      field      = "ml.is_anomaly"
      operator   = "equals"
      value      = "true"
      risk_score = 95
    }
  ]
  
  related_integrations = [
    {
      package     = "ml"
      version     = "2.0.0"
      integration = "anomaly_detection"
    }
  ]

  required_fields = [
    {
      name = "ml.is_anomaly"
      type = "boolean"
    },
    {
      name = "ml.job_id"
      type = "keyword"
    }
  ]

  severity_mapping = [
    {
      field    = "ml.is_anomaly"
      operator = "equals"
      value    = "true"
      severity = "high"
    }
  ]
  
  response_actions = [
    {
      action_type_id = ".osquery"
      params = {
        pack_id = "ml_anomaly_investigation"
        timeout = 700
        ecs_mapping = {
          "ml.job_id"        = "job_id"
          "ml.is_anomaly"    = "is_anomaly"
          "host.name"        = "hostname"
        }
        queries = [
          {
            id       = "ml_query1"
            query    = "SELECT * FROM system_info;"
            platform = "linux"
            version  = "4.7.0"
          }
        ]
      }
    },
    {
      action_type_id = ".endpoint"
      params = {
        command = "isolate"
        comment = "Collect process tree for ML anomaly investigation"
      }
    }
  ]
  
  exceptions_list = [
    {
      id             = "ml-exception-1"
      list_id        = "ml-rule-exceptions"
      namespace_type = "agnostic"
      type           = "detection"
    }
  ]
}
`, name)
}

func testAccSecurityDetectionRuleConfig_newTerms(name string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_security_detection_rule" "test" {
  name                = "%s"
  type                = "new_terms"
  query               = "user.name:*"
  language            = "kuery"
  enabled             = true
  description         = "Test new terms security detection rule"
  severity            = "medium"
  risk_score          = 50
  from                = "now-6m"
  to                  = "now"
  interval            = "5m"
  index               = ["logs-*"]
  new_terms_fields    = ["user.name"]
  history_window_start = "now-14d"
  data_view_id        = "new-terms-data-view-id"
  namespace           = "new-terms-namespace"
  rule_name_override = "Custom New Terms Rule Name"
  timestamp_override = "user.created"
  timestamp_override_fallback_disabled = true

  meta = jsonencode({
    "new_terms_type" = "user_behavior"
    "custom_field"   = "test_value"
    "detection"      = "anomaly"
  })

  investigation_fields = ["user.name", "user.type"]

  risk_score_mapping = [
    {
      field      = "user.type"
      operator   = "equals"
      value      = "service_account"
      risk_score = 65
    }
  ]

  related_integrations = [
    {
      package     = "security"
      version     = "1.0.0"
      integration = "users"
    }
  ]

  required_fields = [
    {
      name = "user.name"
      type = "keyword"
    },
    {
      name = "user.type"
      type = "keyword"
    }
  ]

  severity_mapping = [
    {
      field    = "user.type"
      operator = "equals"
      value    = "service_account"
      severity = "medium"
    }
  ]

  response_actions = [
    {
      action_type_id = ".osquery"
      params = {
        query   = "SELECT * FROM last WHERE username = '{{user.name}}';"
        timeout = 350
        ecs_mapping = {
          "user.name" = "username"
          "user.type" = "user_type"
          "host.name" = "hostname"
        }
      }
    }
  ]
}
`, name)
}

func testAccSecurityDetectionRuleConfig_newTermsUpdate(name string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_security_detection_rule" "test" {
  name                = "%s"
  type                = "new_terms"
  query               = "user.name:* AND source.ip:*"
  language            = "kuery"
  enabled             = true
  description         = "Updated test new terms security detection rule"
  severity            = "high"
  risk_score          = 75
  from                = "now-6m"
  to                  = "now"
  interval            = "5m"
  index               = ["logs-*", "audit-*"]
  new_terms_fields    = ["user.name", "source.ip"]
  history_window_start = "now-30d"
  author              = ["Test Author"]
  tags                = ["test", "new-terms", "automation"]
  license             = "Elastic License v2"
  rule_name_override = "Updated Custom New Terms Rule Name"
  timestamp_override = "user.last_login"
  timestamp_override_fallback_disabled = false

  meta = jsonencode({
    "new_terms_type" = "user_behavior"
    "custom_field"   = "updated_value"
    "detection"      = "anomaly"
    "updated"        = "yes"
  })

  investigation_fields = ["user.name", "user.type", "source.ip", "user.roles"]

  risk_score_mapping = [
    {
      field      = "user.roles"
      operator   = "equals"
      value      = "admin"
      risk_score = 95
    },
    {
      field      = "source.geo.country_name"
      operator   = "equals"
      value      = "CN"
      risk_score = 85
    }
  ]

  related_integrations = [
    {
      package     = "security"
      version     = "2.0.0"
      integration = "users"
    }
  ]

  required_fields = [
    {
      name = "user.name"
      type = "keyword"
    },
    {
      name = "source.ip"
      type = "ip"
    },
    {
      name = "user.roles"
      type = "keyword"
    }
  ]

  severity_mapping = [
    {
      field    = "user.roles"
      operator = "equals"
      value    = "admin"
      severity = "high"
    }
  ]

  response_actions = [
    {
      action_type_id = ".osquery"
      params = {
        saved_query_id = "admin_user_investigation"
        timeout        = 800
        ecs_mapping = {
          "user.roles"     = "roles"
          "source.ip"      = "source_ip"
          "user.name"      = "username"
        }
      }
    },
    {
      action_type_id = ".endpoint"
      params = {
        command = "isolate"
        comment = "Isolate host due to new admin user activity from suspicious IP"
      }
    }
  ]
}
`, name)
}

func testAccSecurityDetectionRuleConfig_savedQuery(name string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_security_detection_rule" "test" {
  name         = "%s"
  type         = "saved_query"
  query        = "*:*"
  enabled      = true
  description  = "Test saved query security detection rule"
  severity     = "low"
  risk_score   = 30
  from         = "now-6m"
  to           = "now"
  interval     = "5m"
  index        = ["logs-*"]
  saved_id     = "test-saved-query-id"
  data_view_id = "saved-query-data-view-id"
  namespace    = "saved-query-namespace"
  rule_name_override = "Custom Saved Query Rule Name"
  timestamp_override = "event.start"
  timestamp_override_fallback_disabled = false

  meta = jsonencode({
    "saved_query_type" = "security"
    "custom_field"     = "test_value"
    "query_origin"     = "saved"
  })

  investigation_fields = ["event.category", "event.action"]

  risk_score_mapping = [
    {
      field      = "event.category"
      operator   = "equals"
      value      = "authentication"
      risk_score = 45
    }
  ]

  related_integrations = [
    {
      package     = "system"
      version     = "1.0.0"
      integration = "logs"
    }
  ]

  required_fields = [
    {
      name = "event.category"
      type = "keyword"
    },
    {
      name = "event.action"
      type = "keyword"
    }
  ]

  severity_mapping = [
    {
      field    = "event.category"
      operator = "equals"
      value    = "authentication"
      severity = "low"
    }
  ]

  response_actions = [
    {
      action_type_id = ".osquery"
      params = {
        query   = "SELECT * FROM logged_in_users WHERE user = '{{user.name}}';"
        timeout = 250
        ecs_mapping = {
          "event.category" = "category"
          "event.action"   = "action"
          "user.name"      = "username"
        }
      }
    }
  ]
}
`, name)
}

func testAccSecurityDetectionRuleConfig_savedQueryUpdate(name string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_security_detection_rule" "test" {
  name         = "%s"
  type         = "saved_query"
  query        = "event.action:*"
  enabled      = true
  description  = "Updated test saved query security detection rule"
  severity     = "medium"
  risk_score   = 60
  from         = "now-6m"
  to           = "now"
  interval     = "5m"
  index        = ["logs-*", "audit-*"]
  saved_id     = "test-saved-query-id-updated"
  data_view_id = "updated-saved-query-data-view-id"
  namespace    = "updated-saved-query-namespace"
  author       = ["Test Author"]
  tags        = ["test", "saved-query", "automation"]
  license     = "Elastic License v2"
  rule_name_override = "Updated Custom Saved Query Rule Name"
  timestamp_override = "event.end"
  timestamp_override_fallback_disabled = true

  meta = jsonencode({
    "saved_query_type" = "security"
    "custom_field"     = "updated_value"
    "query_origin"     = "saved"
    "updated"          = "yes"
  })

  investigation_fields = ["host.name", "user.name", "process.name"]

  risk_score_mapping = [
    {
      field      = "event.type"
      operator   = "equals"
      value      = "access"
      risk_score = 70
    }
  ]
  
  related_integrations = [
    {
      package     = "system"
      version     = "2.0.0"
      integration = "logs"
    }
  ]

  required_fields = [
    {
      name = "event.type"
      type = "keyword"
    },
    {
      name = "host.name"
      type = "keyword"
    }
  ]

  severity_mapping = [
    {
      field    = "event.type"
      operator = "equals"
      value    = "access"
      severity = "medium"
    }
  ]
  
  response_actions = [
    {
      action_type_id = ".osquery"
      params = {
        pack_id = "access_investigation_pack"
        timeout = 400
        ecs_mapping = {
          "event.type" = "type"
          "host.name"  = "hostname"
          "user.name"  = "username"
        }
        queries = [
          {
            id       = "access_query1"
            query    = "SELECT * FROM users WHERE username = '{{user.name}}';"
            platform = "linux"
            version  = "4.8.0"
            ecs_mapping = {
              "user.id" = "uid"
            }
          }
        ]
      }
    }
  ]
  
  exceptions_list = [
    {
      id             = "saved-query-exception-1"
      list_id        = "saved-query-exceptions"
      namespace_type = "agnostic"
      type           = "detection"
    }
  ]
}
`, name)
}

func testAccSecurityDetectionRuleConfig_threatMatch(name string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_security_detection_rule" "test" {
  name         = "%s"
  type         = "threat_match"
  query        = "destination.ip:*"
  language     = "kuery"
  enabled      = true
  description  = "Test threat match security detection rule"
  severity     = "high"
  risk_score   = 80
  from         = "now-6m"
  to           = "now"
  interval     = "5m"
  index        = ["logs-*"]
  data_view_id = "threat-match-data-view-id"
  namespace    = "threat-match-namespace"
  rule_name_override = "Custom Threat Match Rule Name"
  timestamp_override = "threat.indicator.first_seen"
  timestamp_override_fallback_disabled = true
  threat_index = ["threat-intel-*"]
  threat_query = "threat.indicator.type:ip"
  
  meta = jsonencode({
    "threat_type"    = "indicator_match"
    "custom_field"   = "test_value"
    "intelligence"   = "external"
  })

  investigation_fields = ["destination.ip", "source.ip"]

  threat_mapping = [
    {
      entries = [
        {
          field = "destination.ip"
          type  = "mapping"
          value = "threat.indicator.ip"
        }
      ]
    }
  ]

  risk_score_mapping = [
    {
      field      = "threat.indicator.confidence"
      operator   = "equals"
      value      = "medium"
      risk_score = 85
    }
  ]

  related_integrations = [
    {
      package     = "threat_intel"
      version     = "1.0.0"
      integration = "indicators"
    }
  ]

  required_fields = [
    {
      name = "destination.ip"
      type = "ip"
    },
    {
      name = "threat.indicator.ip"
      type = "ip"
    }
  ]

  severity_mapping = [
    {
      field    = "threat.indicator.confidence"
      operator = "equals"
      value    = "high"
      severity = "high"
    }
  ]

  response_actions = [
    {
      action_type_id = ".osquery"
      params = {
        query   = "SELECT * FROM listening_ports WHERE address = '{{destination.ip}}';"
        timeout = 300
        ecs_mapping = {
          "destination.ip"            = "dest_ip"
          "threat.indicator.ip"       = "threat_ip"
          "threat.indicator.confidence" = "confidence"
        }
      }
    },
    {
      action_type_id = ".endpoint"
      params = {
        command = "isolate"
        comment = "Isolate host due to threat match on destination IP"
      }
    }
  ]
}
`, name)
}

func testAccSecurityDetectionRuleConfig_threatMatchUpdate(name string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_security_detection_rule" "test" {
  name         = "%s"
  type         = "threat_match"
  query        = "destination.ip:* OR source.ip:*"
  language     = "kuery"
  enabled      = true
  description  = "Updated test threat match security detection rule"
  severity     = "critical"
  risk_score   = 95
  from         = "now-6m"
  to           = "now"
  interval     = "5m"
  index        = ["logs-*", "network-*"]
  data_view_id = "updated-threat-match-data-view-id"
  namespace    = "updated-threat-match-namespace"
  threat_index = ["threat-intel-*", "ioc-*"]
  threat_query = "threat.indicator.type:(ip OR domain)"
  author       = ["Test Author"]
  tags         = ["test", "threat-match", "automation"]
  license      = "Elastic License v2"
  rule_name_override = "Updated Custom Threat Match Rule Name"
  timestamp_override = "threat.indicator.last_seen"
  timestamp_override_fallback_disabled = false
  
  meta = jsonencode({
    "threat_type"    = "indicator_match"
    "custom_field"   = "updated_value"
    "intelligence"   = "external"
    "updated"        = "yes"
  })

  investigation_fields = ["destination.ip", "source.ip", "threat.indicator.type"]

  threat_mapping = [
    {
      entries = [
        {
          field = "destination.ip"
          type  = "mapping"
          value = "threat.indicator.ip"
        }
      ]
    },
    {
      entries = [
        {
          field = "source.ip"
          type  = "mapping"
          value = "threat.indicator.ip"
        }
      ]
    }
  ]

  risk_score_mapping = [
    {
      field      = "threat.indicator.confidence"
      operator   = "equals"
      value      = "high"
      risk_score = 100
    }
  ]

  related_integrations = [
    {
      package     = "threat_intel"
      version     = "2.0.0"
      integration = "indicators"
    }
  ]

  required_fields = [
    {
      name = "destination.ip"
      type = "ip"
    },
    {
      name = "source.ip"
      type = "ip"
    },
    {
      name = "threat.indicator.ip"
      type = "ip"
    }
  ]

  severity_mapping = [
    {
      field    = "threat.indicator.confidence"
      operator = "equals"
      value    = "high"
      severity = "critical"
    }
  ]

  response_actions = [
    {
      action_type_id = ".osquery"
      params = {
        saved_query_id = "threat_intel_investigation"
        timeout        = 450
        ecs_mapping = {
          "source.ip"                 = "src_ip"
          "destination.ip"            = "dest_ip"
          "threat.indicator.type"     = "threat_type"
        }
      }
    },
    {
      action_type_id = ".endpoint"
      params = {
        command = "kill-process"
        comment = "Kill processes communicating with known threat indicators"
        config = {
          field     = "process.entity_id"
          overwrite = true
        }
      }
    }
  ]
}
`, name)
}

func testAccSecurityDetectionRuleConfig_threshold(name string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_security_detection_rule" "test" {
  name         = "%s"
  type         = "threshold"
  query        = "event.action:login"
  language     = "kuery"
  enabled      = true
  description  = "Test threshold security detection rule"
  severity     = "medium"
  risk_score   = 55
  from         = "now-6m"
  to           = "now"
  interval     = "5m"
  index        = ["logs-*"]
  data_view_id = "threshold-data-view-id"
  namespace    = "threshold-namespace"
  rule_name_override = "Custom Threshold Rule Name"
  timestamp_override = "event.created"
  timestamp_override_fallback_disabled = false
  
  meta = jsonencode({
    "threshold_type" = "count_based"
    "custom_field"   = "test_value"
    "monitoring"     = "enabled"
  })

  investigation_fields = ["user.name", "event.action"]

  threshold = {
    value = 10
    field = ["user.name"]
  }

  risk_score_mapping = [
    {
      field      = "event.outcome"
      operator   = "equals"
      value      = "success"
      risk_score = 45
    }
  ]

  related_integrations = [
    {
      package     = "system"
      version     = "1.0.0"
      integration = "auth"
    }
  ]

  required_fields = [
    {
      name = "event.action"
      type = "keyword"
    },
    {
      name = "user.name"
      type = "keyword"
    }
  ]

  severity_mapping = [
    {
      field    = "event.outcome"
      operator = "equals"
      value    = "success"
      severity = "medium"
    }
  ]

  response_actions = [
    {
      action_type_id = ".osquery"
      params = {
        query   = "SELECT * FROM logged_in_users WHERE user = '{{user.name}}' ORDER BY time DESC LIMIT 10;"
        timeout = 200
        ecs_mapping = {
          "user.name"     = "username"
          "event.action"  = "action"
          "event.outcome" = "outcome"
        }
      }
    }
  ]
}
`, name)
}

func testAccSecurityDetectionRuleConfig_thresholdUpdate(name string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_security_detection_rule" "test" {
  name         = "%s"
  type         = "threshold"
  query        = "event.action:(login OR logout)"
  language     = "kuery"
  enabled      = true
  description  = "Updated test threshold security detection rule"
  severity     = "high"
  risk_score   = 75
  from         = "now-6m"
  to           = "now"
  interval     = "5m"
  index        = ["logs-*", "audit-*"]
  data_view_id = "updated-threshold-data-view-id"
  namespace    = "updated-threshold-namespace"
  author       = ["Test Author"]
  tags        = ["test", "threshold", "automation"]
  license     = "Elastic License v2"
  rule_name_override = "Updated Custom Threshold Rule Name"
  timestamp_override = "event.start"
  timestamp_override_fallback_disabled = true
  
  meta = jsonencode({
    "threshold_type" = "count_based"
    "custom_field"   = "updated_value"
    "monitoring"     = "enabled"
    "updated"        = "yes"
  })

  investigation_fields = ["user.name", "source.ip", "event.outcome"]

  threshold = {
    value = 20
    field = ["user.name", "source.ip"]
  }

  risk_score_mapping = [
    {
      field      = "event.outcome"
      operator   = "equals"
      value      = "failure"
      risk_score = 90
    }
  ]

  related_integrations = [
    {
      package     = "system"
      version     = "2.0.0"
      integration = "auth"
    }
  ]

  required_fields = [
    {
      name = "event.action"
      type = "keyword"
    },
    {
      name = "source.ip"
      type = "ip"
    }
  ]

  severity_mapping = [
    {
      field    = "event.outcome"
      operator = "equals"
      value    = "failure"
      severity = "high"
    }
  ]

  response_actions = [
    {
      action_type_id = ".osquery"
      params = {
        pack_id = "login_failure_investigation"
        timeout = 350
        ecs_mapping = {
          "event.outcome" = "outcome"
          "source.ip"     = "source_ip"
          "user.name"     = "username"
        }
        queries = [
          {
            id       = "failed_login_query"
            query    = "SELECT * FROM last WHERE type = 7 AND username = '{{user.name}}';"
            platform = "linux"
            version  = "4.9.0"
          }
        ]
      }
    },
    {
      action_type_id = ".endpoint"
      params = {
        command = "isolate"
        comment = "Isolate host due to multiple failed login attempts"
      }
    }
  ]
}
`, name)
}

func TestAccResourceSecurityDetectionRule_WithConnectorAction(t *testing.T) {
	resourceName := "elasticstack_kibana_security_detection_rule.test"
	connectorResourceName := "elasticstack_kibana_action_connector.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.Providers,
		CheckDestroy:             testAccCheckSecurityDetectionRuleDestroy,
		Steps: []resource.TestStep{
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minVersionSupport),
				Config:   testAccSecurityDetectionRuleConfig_withConnectorAction("test-rule-with-action"),
				Check: resource.ComposeTestCheckFunc(
					// Check connector attributes
					resource.TestCheckResourceAttr(connectorResourceName, "name", "test connector 1"),
					resource.TestCheckResourceAttr(connectorResourceName, "connector_id", "1d30b67b-f90b-4e28-87c2-137cba361509"),
					resource.TestCheckResourceAttr(connectorResourceName, "connector_type_id", ".cases-webhook"),
					resource.TestCheckResourceAttrSet(connectorResourceName, "config"),
					resource.TestCheckResourceAttrSet(connectorResourceName, "secrets"),

					// Check security detection rule attributes
					resource.TestCheckResourceAttr(resourceName, "name", "test-rule-with-action"),
					resource.TestCheckResourceAttr(resourceName, "type", "query"),
					resource.TestCheckResourceAttr(resourceName, "query", "user.name:*"),
					resource.TestCheckResourceAttr(resourceName, "language", "kuery"),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "description", "Test security detection rule with connector action"),
					resource.TestCheckResourceAttr(resourceName, "severity", "medium"),
					resource.TestCheckResourceAttr(resourceName, "risk_score", "50"),
					resource.TestCheckResourceAttr(resourceName, "index.0", "logs-*"),
					resource.TestCheckResourceAttr(resourceName, "data_view_id", "connector-action-data-view-id"),
					resource.TestCheckResourceAttr(resourceName, "namespace", "connector-action-namespace"),

					// Check risk score mapping
					resource.TestCheckResourceAttr(resourceName, "risk_score_mapping.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "risk_score_mapping.0.field", "user.privileged"),
					resource.TestCheckResourceAttr(resourceName, "risk_score_mapping.0.operator", "equals"),
					resource.TestCheckResourceAttr(resourceName, "risk_score_mapping.0.value", "true"),
					resource.TestCheckResourceAttr(resourceName, "risk_score_mapping.0.risk_score", "75"),

					// Check action attributes
					resource.TestCheckResourceAttr(resourceName, "actions.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "actions.0.action_type_id", ".cases-webhook"),
					resource.TestCheckResourceAttr(resourceName, "actions.0.id", "1d30b67b-f90b-4e28-87c2-137cba361509"),
					resource.TestCheckResourceAttr(resourceName, "actions.0.group", "default"),
					resource.TestCheckResourceAttr(resourceName, "actions.0.params.message", "CRITICAL EQL Alert: PowerShell process detected"),
					resource.TestCheckResourceAttr(resourceName, "actions.0.frequency.notify_when", "onActiveAlert"),
					resource.TestCheckResourceAttr(resourceName, "actions.0.frequency.summary", "true"),
					resource.TestCheckResourceAttr(resourceName, "actions.0.frequency.throttle", "10m"),

					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "rule_id"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "created_by"),
				),
			},
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minVersionSupport),
				Config:   testAccSecurityDetectionRuleConfig_withConnectorActionUpdate("test-rule-with-action-updated"),
				Check: resource.ComposeTestCheckFunc(
					// Check updated rule attributes
					resource.TestCheckResourceAttr(resourceName, "name", "test-rule-with-action-updated"),
					resource.TestCheckResourceAttr(resourceName, "description", "Updated test security detection rule with connector action"),
					resource.TestCheckResourceAttr(resourceName, "severity", "high"),
					resource.TestCheckResourceAttr(resourceName, "risk_score", "75"),
					resource.TestCheckResourceAttr(resourceName, "data_view_id", "updated-connector-action-data-view-id"),
					resource.TestCheckResourceAttr(resourceName, "namespace", "updated-connector-action-namespace"),
					resource.TestCheckResourceAttr(resourceName, "tags.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.0", "test"),
					resource.TestCheckResourceAttr(resourceName, "tags.1", "terraform"),

					// Check risk score mapping
					resource.TestCheckResourceAttr(resourceName, "risk_score_mapping.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "risk_score_mapping.0.field", "user.privileged"),
					resource.TestCheckResourceAttr(resourceName, "risk_score_mapping.0.operator", "equals"),
					resource.TestCheckResourceAttr(resourceName, "risk_score_mapping.0.value", "true"),
					resource.TestCheckResourceAttr(resourceName, "risk_score_mapping.0.risk_score", "95"),

					// Check updated action attributes
					resource.TestCheckResourceAttr(resourceName, "actions.0.params.message", "UPDATED CRITICAL Alert: Security event detected"),
					resource.TestCheckResourceAttr(resourceName, "actions.0.frequency.throttle", "5m"),

					// Check exceptions list attributes
					resource.TestCheckResourceAttr(resourceName, "exceptions_list.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "exceptions_list.0.id", "test-action-exception"),
					resource.TestCheckResourceAttr(resourceName, "exceptions_list.0.list_id", "action-rule-exceptions"),
					resource.TestCheckResourceAttr(resourceName, "exceptions_list.0.namespace_type", "single"),
					resource.TestCheckResourceAttr(resourceName, "exceptions_list.0.type", "detection"),
				),
			},
		},
	})
}

func testAccSecurityDetectionRuleConfig_withConnectorAction(name string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_action_connector" "test" {
  name         = "test connector 1"
  connector_id = "1d30b67b-f90b-4e28-87c2-137cba361509"
  config = jsonencode({
    createIncidentJson = "{}"
    createIncidentResponseKey = "key"
    createIncidentUrl = "https://www.elastic.co/"
    getIncidentResponseExternalTitleKey = "title"
    getIncidentUrl = "https://www.elastic.co/"
    updateIncidentJson = "{}"
    updateIncidentUrl = "https://elasticsearch.com/"
    viewIncidentUrl = "https://www.elastic.co/"
    createIncidentMethod = "put"
  })
  secrets = jsonencode({
    user = "user2"
    password = "password2"
  })
  connector_type_id = ".cases-webhook"
}

resource "elasticstack_kibana_security_detection_rule" "test" {
  name         = "%s"
  description  = "Test security detection rule with connector action"
  type         = "query"
  severity     = "medium"
  risk_score   = 50
  enabled      = true
  query        = "user.name:*"
  language     = "kuery"
  from         = "now-6m"
  to           = "now"
  interval     = "5m"
  index        = ["logs-*"]
  data_view_id = "connector-action-data-view-id"
  namespace    = "connector-action-namespace"

  risk_score_mapping = [
    {
      field      = "user.privileged"
      operator   = "equals"
      value      = "true"
      risk_score = 75
    }
  ]

  actions = [
    {
      action_type_id = ".cases-webhook"
      id             = "${elasticstack_kibana_action_connector.test.connector_id}"
      params = {
        message = "CRITICAL EQL Alert: PowerShell process detected"
      }
      group = "default"
      frequency = {
        notify_when = "onActiveAlert"
        summary     = true
        throttle    = "10m"
      }
    }
  ]
}
`, name)
}

func testAccSecurityDetectionRuleConfig_withConnectorActionUpdate(name string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_action_connector" "test" {
  name         = "test connector 1"
  connector_id = "1d30b67b-f90b-4e28-87c2-137cba361509"
  config = jsonencode({
    createIncidentJson = "{}"
    createIncidentResponseKey = "key"
    createIncidentUrl = "https://www.elastic.co/"
    getIncidentResponseExternalTitleKey = "title"
    getIncidentUrl = "https://www.elastic.co/"
    updateIncidentJson = "{}"
    updateIncidentUrl = "https://elasticsearch.com/"
    viewIncidentUrl = "https://www.elastic.co/"
    createIncidentMethod = "put"
  })
  secrets = jsonencode({
    user = "user2"
    password = "password2"
  })
  connector_type_id = ".cases-webhook"
}

resource "elasticstack_kibana_security_detection_rule" "test" {
  name         = "%s"
  description  = "Updated test security detection rule with connector action"
  type         = "query"
  severity     = "high"
  risk_score   = 75
  enabled      = true
  query        = "user.name:*"
  language     = "kuery"
  from         = "now-6m"
  to           = "now"
  interval     = "5m"
  index        = ["logs-*"]
  data_view_id = "updated-connector-action-data-view-id"
  namespace    = "updated-connector-action-namespace"
  
  tags = ["test", "terraform"]

  risk_score_mapping = [
    {
      field      = "user.privileged"
      operator   = "equals"
      value      = "true"
      risk_score = 95
    }
  ]

  actions = [
    {
      action_type_id = ".cases-webhook"
      id             = "${elasticstack_kibana_action_connector.test.connector_id}"
      params = {
        message = "UPDATED CRITICAL Alert: Security event detected"
      }
      group = "default"
      frequency = {
        notify_when = "onActiveAlert"
        summary     = true
        throttle    = "5m"
      }
    }
  ]
  
  exceptions_list = [
    {
      id             = "test-action-exception"
      list_id        = "action-rule-exceptions"
      namespace_type = "single"
      type           = "detection"
    }
  ]
}
`, name)
}

func TestAccResourceSecurityDetectionRule_BuildingBlockType(t *testing.T) {
	resourceName := "elasticstack_kibana_security_detection_rule.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.Providers,
		CheckDestroy:             testAccCheckSecurityDetectionRuleDestroy,
		Steps: []resource.TestStep{
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minVersionSupport),
				Config:   testAccSecurityDetectionRuleConfig_buildingBlockType("test-building-block-rule"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", "test-building-block-rule"),
					resource.TestCheckResourceAttr(resourceName, "type", "query"),
					resource.TestCheckResourceAttr(resourceName, "query", "process.name:*"),
					resource.TestCheckResourceAttr(resourceName, "language", "kuery"),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "description", "Test building block security detection rule"),
					resource.TestCheckResourceAttr(resourceName, "severity", "low"),
					resource.TestCheckResourceAttr(resourceName, "risk_score", "21"),
					resource.TestCheckResourceAttr(resourceName, "index.0", "logs-*"),
					resource.TestCheckResourceAttr(resourceName, "data_view_id", "building-block-data-view-id"),
					resource.TestCheckResourceAttr(resourceName, "namespace", "building-block-namespace"),
					resource.TestCheckResourceAttr(resourceName, "building_block_type", "default"),

					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "rule_id"),
				),
			},
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minVersionSupport),
				Config:   testAccSecurityDetectionRuleConfig_buildingBlockTypeUpdate("test-building-block-rule-updated"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", "test-building-block-rule-updated"),
					resource.TestCheckResourceAttr(resourceName, "query", "process.name:* AND user.name:*"),
					resource.TestCheckResourceAttr(resourceName, "description", "Updated test building block security detection rule"),
					resource.TestCheckResourceAttr(resourceName, "severity", "medium"),
					resource.TestCheckResourceAttr(resourceName, "risk_score", "40"),
					resource.TestCheckResourceAttr(resourceName, "data_view_id", "updated-building-block-data-view-id"),
					resource.TestCheckResourceAttr(resourceName, "namespace", "updated-building-block-namespace"),
					resource.TestCheckResourceAttr(resourceName, "building_block_type", "default"),
					resource.TestCheckResourceAttr(resourceName, "tags.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.0", "building-block"),
					resource.TestCheckResourceAttr(resourceName, "tags.1", "test"),
				),
			},
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minVersionSupport),
				Config:   testAccSecurityDetectionRuleConfig_buildingBlockTypeRemoved("test-building-block-rule-no-type"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", "test-building-block-rule-no-type"),
					resource.TestCheckResourceAttr(resourceName, "description", "Test rule without building block type"),
					resource.TestCheckResourceAttr(resourceName, "data_view_id", "no-building-block-data-view-id"),
					resource.TestCheckResourceAttr(resourceName, "namespace", "no-building-block-namespace"),
					resource.TestCheckNoResourceAttr(resourceName, "building_block_type"),
				),
			},
		},
	})
}

func testAccSecurityDetectionRuleConfig_buildingBlockType(name string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_security_detection_rule" "test" {
  name                = "%s"
  type                = "query"
  query               = "process.name:*"
  language            = "kuery"
  enabled             = true
  description         = "Test building block security detection rule"
  severity            = "low"
  risk_score          = 21
  from                = "now-6m"
  to                  = "now"
  interval            = "5m"
  index               = ["logs-*"]
  data_view_id        = "building-block-data-view-id"
  namespace           = "building-block-namespace"
  building_block_type = "default"
}
`, name)
}

func testAccSecurityDetectionRuleConfig_buildingBlockTypeUpdate(name string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_security_detection_rule" "test" {
  name                = "%s"
  type                = "query"
  query               = "process.name:* AND user.name:*"
  language            = "kuery"
  enabled             = true
  description         = "Updated test building block security detection rule"
  severity            = "medium"
  risk_score          = 40
  from                = "now-6m"
  to                  = "now"
  interval            = "5m"
  index               = ["logs-*"]
  data_view_id        = "updated-building-block-data-view-id"
  namespace           = "updated-building-block-namespace"
  building_block_type = "default"
  author              = ["Test Author"]
  tags                = ["building-block", "test"]
  license             = "Elastic License v2"
}
`, name)
}

func testAccSecurityDetectionRuleConfig_buildingBlockTypeRemoved(name string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_security_detection_rule" "test" {
  name         = "%s"
  type         = "query"
  query        = "process.name:*"
  language     = "kuery"
  enabled      = true
  description  = "Test rule without building block type"
  severity     = "medium"
  risk_score   = 50
  from         = "now-6m"
  to           = "now"
  interval     = "5m"
  index        = ["logs-*"]
  data_view_id = "no-building-block-data-view-id"
  namespace    = "no-building-block-namespace"
}
`, name)
}

func TestAccResourceSecurityDetectionRule_Meta(t *testing.T) {
	resourceName := "elasticstack_kibana_security_detection_rule.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.Providers,
		CheckDestroy:             testAccCheckSecurityDetectionRuleDestroy,
		Steps: []resource.TestStep{
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minVersionSupport),
				Config:   testAccSecurityDetectionRuleConfig_meta("test-meta-rule"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", "test-meta-rule"),
					resource.TestCheckResourceAttr(resourceName, "type", "query"),
					checkResourceJSONAttr(resourceName, "meta", `{"test_key": "test_value", "author": "terraform-provider", "version": "1.0"}`),
				),
			},
		},
	})
}

func testAccSecurityDetectionRuleConfig_meta(name string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_security_detection_rule" "test" {
  name         = "%s"
  type         = "query"
  query        = "*:*"
  language     = "kuery"
  enabled      = true
  description  = "Test query security detection rule with meta field"
  severity     = "medium"
  risk_score   = 50
  from         = "now-6m"
  to           = "now"
  interval     = "5m"
  index        = ["logs-*"]

  meta = jsonencode({
    test_key = "test_value"
    author   = "terraform-provider"
    version  = "1.0"
  })
}
`, name)
}

func TestAccResourceSecurityDetectionRule_MetaMixedTypes(t *testing.T) {
	resourceName := "elasticstack_kibana_security_detection_rule.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.Providers,
		CheckDestroy:             testAccCheckSecurityDetectionRuleDestroy,
		Steps: []resource.TestStep{
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minVersionSupport),
				Config:   testAccSecurityDetectionRuleConfig_metaMixedTypes("test-meta-mixed-types-rule"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", "test-meta-mixed-types-rule"),
					resource.TestCheckResourceAttr(resourceName, "type", "query"),
					// Check that the meta field contains all the mixed types as a JSON string
					checkResourceJSONAttr(resourceName, "meta", `{
						"string_field": "test_value",
						"number_field": 42,
						"float_field": 3.14,
						"boolean_field": true,
						"array_field": ["item1", "item2", "item3"],
						"object_field": {
							"nested_string": "nested_value",
							"nested_number": 100,
							"nested_boolean": false
						},
						"null_field": null
					}`),
				),
			},
		},
	})
}

func testAccSecurityDetectionRuleConfig_metaMixedTypes(name string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_security_detection_rule" "test" {
  name         = "%s"
  type         = "query"
  query        = "*:*"
  language     = "kuery"
  enabled      = true
  description  = "Test query security detection rule with mixed type meta field"
  severity     = "medium"
  risk_score   = 50
  from         = "now-6m"
  to           = "now"
  interval     = "5m"
  index        = ["logs-*"]

  meta = jsonencode({
    string_field  = "test_value"
    number_field  = 42
    float_field   = 3.14
    boolean_field = true
    array_field   = ["item1", "item2", "item3"]
    object_field = {
      nested_string  = "nested_value"
      nested_number  = 100
      nested_boolean = false
    }
    null_field = null
  })
}
`, name)
}
