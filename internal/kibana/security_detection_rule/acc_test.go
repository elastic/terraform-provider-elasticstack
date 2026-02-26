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

package securitydetectionrule_test

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	kibanaoapi "github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	schemautil "github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/google/uuid"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

const osqueryResponseActionQuery = "SELECT * FROM processes WHERE pid IN (SELECT DISTINCT pid FROM connections WHERE remote_address NOT LIKE '10.%'" +
	" AND remote_address NOT LIKE '192.168.%' AND remote_address NOT LIKE '127.%');"

// checkResourceJSONAttrKey compares the JSON string value of a resource attribute key.
// The attribute value is expected to be a JSON-encoded string.
func checkResourceJSONAttrKey(key, expectedJSON string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		name := securityDetectionRuleResourceName
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

		if eq, jsonErr := schemautil.JSONBytesEqual([]byte(expectedJSON), []byte(actualJSON)); !eq {
			if jsonErr != nil {
				return fmt.Errorf(
					"%s: Attribute '%s' expected %#v, got %#v: %w",
					name,
					key,
					expectedJSON,
					actualJSON,
					jsonErr,
				)
			}

			return fmt.Errorf(
				"%s: Attribute '%s' expected %#v, got %#v",
				name,
				key,
				expectedJSON,
				actualJSON,
			)
		}
		return nil
	}
}

var minVersionSupport = version.Must(version.NewVersion("8.11.0"))
var minResponseActionVersionSupport = version.Must(version.NewVersion("8.16.0"))

const securityDetectionRuleResourceName = "elasticstack_kibana_security_detection_rule.test"

func TestAccResourceSecurityDetectionRule_Query(t *testing.T) {
	resourceName := securityDetectionRuleResourceName

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: testAccCheckSecurityDetectionRuleDestroy,
		Steps: []resource.TestStep{
			{
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minResponseActionVersionSupport),
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"name": config.StringVariable("test-query-rule"),
				},
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

					// Check filters field
					checkResourceJSONAttrKey("filters", `[{"bool": {"must": [{"term": {"event.category": "authentication"}}], "must_not": [{"term": {"event.outcome": "success"}}]}}]`),

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

					// Check alert suppression
					resource.TestCheckResourceAttr(resourceName, "alert_suppression.group_by.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "alert_suppression.group_by.0", "user.name"),
					resource.TestCheckResourceAttr(resourceName, "alert_suppression.group_by.1", "host.name"),
					resource.TestCheckResourceAttr(resourceName, "alert_suppression.duration", "5m"),
					resource.TestCheckResourceAttr(resourceName, "alert_suppression.missing_fields_strategy", "suppress"),

					// Verify building_block_type is not set by default
					resource.TestCheckNoResourceAttr(resourceName, "building_block_type"),

					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "rule_id"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "created_by"),
				),
			},
			{
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minResponseActionVersionSupport),
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables: config.Variables{
					"name": config.StringVariable("test-query-rule-updated"),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", "test-query-rule-updated"),
					resource.TestCheckResourceAttr(resourceName, "description", "Updated test query security detection rule"),
					resource.TestCheckResourceAttr(resourceName, "severity", "high"),
					resource.TestCheckResourceAttr(resourceName, "risk_score", "75"),
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

					// Check filters field (updated values)
					checkResourceJSONAttrKey("filters", `[{"range": {"@timestamp": {"gte": "now-1h", "lte": "now"}}}, {"terms": {"event.action": ["login", "logout", "access"]}}]`),

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
					resource.TestCheckResourceAttr(resourceName, "response_actions.1.params.config.overwrite", "false"),
				),
			},
			{
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minResponseActionVersionSupport),
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("remove_filters"),
				ConfigVariables: config.Variables{
					"name": config.StringVariable("test-query-rule-no-filters"),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", "test-query-rule-no-filters"),
					resource.TestCheckResourceAttr(resourceName, "description", "Test query rule with filters removed"),
					resource.TestCheckResourceAttr(resourceName, "severity", "medium"),
					resource.TestCheckResourceAttr(resourceName, "risk_score", "55"),

					// Verify filters field is not present when not specified
					resource.TestCheckNoResourceAttr(resourceName, "filters"),
				),
			},
			{
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minResponseActionVersionSupport),
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("remove_filters"),
				ConfigVariables: config.Variables{
					"name": config.StringVariable("test-query-rule-no-filters"),
				},
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccResourceSecurityDetectionRule_EQL(t *testing.T) {
	resourceName := securityDetectionRuleResourceName

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: testAccCheckSecurityDetectionRuleDestroy,
		Steps: []resource.TestStep{
			{
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minResponseActionVersionSupport),
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"name": config.StringVariable("test-eql-rule"),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", "test-eql-rule"),
					resource.TestCheckResourceAttr(resourceName, "type", "eql"),
					resource.TestCheckResourceAttr(resourceName, "query", "process where process.name == \"cmd.exe\""),
					resource.TestCheckResourceAttr(resourceName, "language", "eql"),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "description", "Test EQL security detection rule"),
					resource.TestCheckResourceAttr(resourceName, "severity", "high"),
					resource.TestCheckResourceAttr(resourceName, "risk_score", "70"),
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

					// Check filters field
					checkResourceJSONAttrKey("filters", `[{"bool": {"filter": [{"term": {"process.parent.name": "explorer.exe"}}]}}]`),

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
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minResponseActionVersionSupport),
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables: config.Variables{
					"name": config.StringVariable("test-eql-rule-updated"),
				},
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

					// Check filters field (updated values)
					checkResourceJSONAttrKey("filters", `[{"exists": {"field": "process.code_signature.trusted"}}, {"term": {"host.os.family": "windows"}}]`),

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

					// Check alert suppression (updated values)
					resource.TestCheckResourceAttr(resourceName, "alert_suppression.group_by.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "alert_suppression.group_by.0", "process.parent.name"),
					resource.TestCheckResourceAttr(resourceName, "alert_suppression.group_by.1", "host.name"),
					resource.TestCheckResourceAttr(resourceName, "alert_suppression.duration", "45m"),
					resource.TestCheckResourceAttr(resourceName, "alert_suppression.missing_fields_strategy", "doNotSuppress"),
				),
			},
		},
	})
}

func TestAccResourceSecurityDetectionRule_ESQL(t *testing.T) {
	resourceName := securityDetectionRuleResourceName

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: testAccCheckSecurityDetectionRuleDestroy,
		Steps: []resource.TestStep{
			{
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minResponseActionVersionSupport),
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"name": config.StringVariable("test-esql-rule"),
				},
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

					// Check alert suppression
					resource.TestCheckResourceAttr(resourceName, "alert_suppression.group_by.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "alert_suppression.group_by.0", "user.name"),
					resource.TestCheckResourceAttr(resourceName, "alert_suppression.group_by.1", "user.domain"),
					resource.TestCheckResourceAttr(resourceName, "alert_suppression.duration", "15m"),
					resource.TestCheckResourceAttr(resourceName, "alert_suppression.missing_fields_strategy", "doNotSuppress"),

					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "rule_id"),
				),
			},
			{
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minResponseActionVersionSupport),
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables: config.Variables{
					"name": config.StringVariable("test-esql-rule-updated"),
				},
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
	resourceName := securityDetectionRuleResourceName

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: testAccCheckSecurityDetectionRuleDestroy,
		Steps: []resource.TestStep{
			{
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minResponseActionVersionSupport),
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"name": config.StringVariable("test-ml-rule"),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", "test-ml-rule"),
					resource.TestCheckResourceAttr(resourceName, "type", "machine_learning"),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "description", "Test ML security detection rule"),
					resource.TestCheckResourceAttr(resourceName, "severity", "critical"),
					resource.TestCheckResourceAttr(resourceName, "risk_score", "90"),
					resource.TestCheckResourceAttr(resourceName, "anomaly_threshold", "75"),
					resource.TestCheckResourceAttr(resourceName, "machine_learning_job_id.0", "test-ml-job"),

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
					resource.TestCheckResourceAttr(resourceName, "response_actions.0.params.query", osqueryResponseActionQuery),
					resource.TestCheckResourceAttr(resourceName, "response_actions.0.params.timeout", "600"),
					resource.TestCheckResourceAttr(resourceName, "response_actions.0.params.ecs_mapping.process.pid", "pid"),
					resource.TestCheckResourceAttr(resourceName, "response_actions.0.params.ecs_mapping.process.name", "name"),
					resource.TestCheckResourceAttr(resourceName, "response_actions.0.params.ecs_mapping.ml.anomaly_score", "anomaly_score"),

					// Check alert suppression
					resource.TestCheckResourceAttr(resourceName, "alert_suppression.group_by.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "alert_suppression.group_by.0", "ml.job_id"),
					resource.TestCheckResourceAttr(resourceName, "alert_suppression.duration", "30m"),
					resource.TestCheckResourceAttr(resourceName, "alert_suppression.missing_fields_strategy", "suppress"),

					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "rule_id"),
				),
			},
			{
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minResponseActionVersionSupport),
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables: config.Variables{
					"name": config.StringVariable("test-ml-rule-updated"),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", "test-ml-rule-updated"),
					resource.TestCheckResourceAttr(resourceName, "description", "Updated test ML security detection rule"),
					resource.TestCheckResourceAttr(resourceName, "severity", "high"),
					resource.TestCheckResourceAttr(resourceName, "risk_score", "85"),
					resource.TestCheckResourceAttr(resourceName, "anomaly_threshold", "80"),
					resource.TestCheckResourceAttr(resourceName, "machine_learning_job_id.0", "test-ml-job"),
					resource.TestCheckResourceAttr(resourceName, "machine_learning_job_id.1", "test-ml-job-2"),

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
	resourceName := securityDetectionRuleResourceName

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: testAccCheckSecurityDetectionRuleDestroy,
		Steps: []resource.TestStep{
			{
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minResponseActionVersionSupport),
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"name": config.StringVariable("test-new-terms-rule"),
				},
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

					// Check filters field
					checkResourceJSONAttrKey("filters", `[{"bool": {"should": [{"wildcard": {"user.domain": "*.internal"}}, {"term": {"user.type": "service_account"}}]}}]`),

					resource.TestCheckResourceAttr(resourceName, "history_window_start", "now-14d"),
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

					// Check alert suppression
					resource.TestCheckResourceAttr(resourceName, "alert_suppression.group_by.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "alert_suppression.group_by.0", "user.name"),
					resource.TestCheckResourceAttr(resourceName, "alert_suppression.group_by.1", "user.type"),
					resource.TestCheckResourceAttr(resourceName, "alert_suppression.duration", "20m"),
					resource.TestCheckResourceAttr(resourceName, "alert_suppression.missing_fields_strategy", "doNotSuppress"),

					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "rule_id"),
				),
			},
			{
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minResponseActionVersionSupport),
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables: config.Variables{
					"name": config.StringVariable("test-new-terms-rule-updated"),
				},
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

					// Check filters field (updated values)
					checkResourceJSONAttrKey("filters", `[{"geo_distance": {"distance": "1000km", "source.geo.location": {"lat": 40.12, "lon": -71.34}}}]`),

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
	resourceName := securityDetectionRuleResourceName

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: testAccCheckSecurityDetectionRuleDestroy,
		Steps: []resource.TestStep{
			{
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minResponseActionVersionSupport),
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"name": config.StringVariable("test-saved-query-rule"),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", "test-saved-query-rule"),
					resource.TestCheckResourceAttr(resourceName, "type", "saved_query"),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "description", "Test saved query security detection rule"),
					resource.TestCheckResourceAttr(resourceName, "severity", "low"),
					resource.TestCheckResourceAttr(resourceName, "risk_score", "30"),
					resource.TestCheckResourceAttr(resourceName, "saved_id", "test-saved-query-id"),

					// Check filters field
					checkResourceJSONAttrKey("filters", `[{"prefix": {"event.action": "user_"}}]`),

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

					// Check alert suppression
					resource.TestCheckResourceAttr(resourceName, "alert_suppression.group_by.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "alert_suppression.group_by.0", "event.category"),
					resource.TestCheckResourceAttr(resourceName, "alert_suppression.group_by.1", "event.action"),
					resource.TestCheckResourceAttr(resourceName, "alert_suppression.duration", "8h"),
					resource.TestCheckResourceAttr(resourceName, "alert_suppression.missing_fields_strategy", "suppress"),

					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "rule_id"),
				),
			},
			{
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minResponseActionVersionSupport),
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables: config.Variables{
					"name": config.StringVariable("test-saved-query-rule-updated"),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", "test-saved-query-rule-updated"),
					resource.TestCheckResourceAttr(resourceName, "query", "event.action:*"),
					resource.TestCheckResourceAttr(resourceName, "description", "Updated test saved query security detection rule"),
					resource.TestCheckResourceAttr(resourceName, "severity", "medium"),
					resource.TestCheckResourceAttr(resourceName, "risk_score", "60"),
					resource.TestCheckResourceAttr(resourceName, "saved_id", "test-saved-query-id-updated"),

					// Check filters field (updated values)
					checkResourceJSONAttrKey("filters", `[{"script": {"script": {"source": "doc['event.severity'].value > 2"}}}]`),

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
	resourceName := securityDetectionRuleResourceName

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: testAccCheckSecurityDetectionRuleDestroy,
		Steps: []resource.TestStep{
			{
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minResponseActionVersionSupport),
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"name": config.StringVariable("test-threat-match-rule"),
				},
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
					resource.TestCheckResourceAttr(resourceName, "namespace", "threat-match-namespace"),
					resource.TestCheckResourceAttr(resourceName, "rule_name_override", "Custom Threat Match Rule Name"),
					resource.TestCheckResourceAttr(resourceName, "timestamp_override", "threat.indicator.first_seen"),
					resource.TestCheckResourceAttr(resourceName, "timestamp_override_fallback_disabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "threat_index.0", "threat-intel-*"),
					resource.TestCheckResourceAttr(resourceName, "threat_query", "threat.indicator.type:ip"),
					resource.TestCheckResourceAttr(resourceName, "threat_mapping.0.entries.0.field", "destination.ip"),
					resource.TestCheckResourceAttr(resourceName, "threat_mapping.0.entries.0.type", "mapping"),
					resource.TestCheckResourceAttr(resourceName, "threat_mapping.0.entries.0.value", "threat.indicator.ip"),

					// Check filters field
					checkResourceJSONAttrKey("filters", `[{"bool": {"must_not": [{"term": {"destination.ip": "127.0.0.1"}}]}}]`),

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

					// Check alert suppression
					resource.TestCheckResourceAttr(resourceName, "alert_suppression.group_by.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "alert_suppression.group_by.0", "destination.ip"),
					resource.TestCheckResourceAttr(resourceName, "alert_suppression.group_by.1", "source.ip"),
					resource.TestCheckResourceAttr(resourceName, "alert_suppression.duration", "1h"),
					resource.TestCheckResourceAttr(resourceName, "alert_suppression.missing_fields_strategy", "doNotSuppress"),

					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "rule_id"),
				),
			},
			{
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minResponseActionVersionSupport),
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables: config.Variables{
					"name": config.StringVariable("test-threat-match-rule-updated"),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", "test-threat-match-rule-updated"),
					resource.TestCheckResourceAttr(resourceName, "query", "destination.ip:* OR source.ip:*"),
					resource.TestCheckResourceAttr(resourceName, "description", "Updated test threat match security detection rule"),
					resource.TestCheckResourceAttr(resourceName, "severity", "critical"),
					resource.TestCheckResourceAttr(resourceName, "risk_score", "95"),
					resource.TestCheckResourceAttr(resourceName, "index.0", "logs-*"),
					resource.TestCheckResourceAttr(resourceName, "index.1", "network-*"),
					resource.TestCheckResourceAttr(resourceName, "namespace", "updated-threat-match-namespace"),
					resource.TestCheckResourceAttr(resourceName, "rule_name_override", "Updated Custom Threat Match Rule Name"),
					resource.TestCheckResourceAttr(resourceName, "timestamp_override", "threat.indicator.last_seen"),
					resource.TestCheckResourceAttr(resourceName, "timestamp_override_fallback_disabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "threat_index.0", "threat-intel-*"),
					resource.TestCheckResourceAttr(resourceName, "threat_index.1", "ioc-*"),
					resource.TestCheckResourceAttr(resourceName, "threat_query", "threat.indicator.type:(ip OR domain)"),
					resource.TestCheckResourceAttr(resourceName, "threat_mapping.0.entries.0.field", "destination.ip"),
					resource.TestCheckResourceAttr(resourceName, "threat_mapping.1.entries.0.field", "source.ip"),

					// Check filters field (updated values)
					checkResourceJSONAttrKey("filters", `[{"regexp": {"destination.domain": ".*\\.suspicious\\.com"}}]`),

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
					resource.TestCheckResourceAttr(resourceName, "response_actions.1.params.config.overwrite", "false"),
				),
			},
		},
	})
}

func TestAccResourceSecurityDetectionRule_ThreatMatch_ThreatFilters(t *testing.T) {
	resourceName := securityDetectionRuleResourceName
	ruleID := "threat-filters-" + sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: testAccCheckSecurityDetectionRuleDestroy,
		Steps: []resource.TestStep{
			{
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minResponseActionVersionSupport),
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("repro"),
				ConfigVariables: config.Variables{
					"name":    config.StringVariable("test-threat-filters-repro"),
					"rule_id": config.StringVariable(ruleID),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", "test-threat-filters-repro"),
					resource.TestCheckResourceAttr(resourceName, "rule_id", ruleID),
					resource.TestCheckResourceAttr(resourceName, "type", "threat_match"),
					resource.TestCheckResourceAttr(resourceName, "language", "kuery"),
					resource.TestCheckResourceAttr(resourceName, "threat_filters.#", "3"),
					checkResourceJSONAttrKey(
						"threat_filters.0",
						`{"$state":{"store":"appState"},"meta":{"disabled":false,"key":"event.category","negate":false,`+
							`"params":{"query":"threat"},"type":"phrase"},"query":{"match_phrase":{"event.category":"threat"}}}`,
					),
					checkResourceJSONAttrKey(
						"threat_filters.1",
						`{"$state":{"store":"appState"},"meta":{"disabled":false,"key":"event.kind","negate":false,`+
							`"params":{"query":"enrichment"},"type":"phrase"},"query":{"match_phrase":{"event.kind":"enrichment"}}}`,
					),
					checkResourceJSONAttrKey(
						"threat_filters.2",
						`{"$state":{"store":"appState"},"meta":{"disabled":false,"key":"event.type","negate":false,"params":{"query":"indicator"},"type":"phrase"},"query":{"match_phrase":{"event.type":"indicator"}}}`,
					),
				),
			},
		},
	})
}

func TestAccResourceSecurityDetectionRule_Threshold(t *testing.T) {
	resourceName := securityDetectionRuleResourceName

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: testAccCheckSecurityDetectionRuleDestroy,
		Steps: []resource.TestStep{
			{
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minResponseActionVersionSupport),
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"name": config.StringVariable("test-threshold-rule"),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", "test-threshold-rule"),
					resource.TestCheckResourceAttr(resourceName, "type", "threshold"),
					resource.TestCheckResourceAttr(resourceName, "query", "event.action:login"),
					resource.TestCheckResourceAttr(resourceName, "language", "kuery"),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "description", "Test threshold security detection rule"),
					resource.TestCheckResourceAttr(resourceName, "severity", "medium"),
					resource.TestCheckResourceAttr(resourceName, "risk_score", "55"),
					resource.TestCheckResourceAttr(resourceName, "data_view_id", "threshold-data-view-id"),
					resource.TestCheckResourceAttr(resourceName, "namespace", "threshold-namespace"),
					resource.TestCheckResourceAttr(resourceName, "rule_name_override", "Custom Threshold Rule Name"),
					resource.TestCheckResourceAttr(resourceName, "timestamp_override", "event.created"),
					resource.TestCheckResourceAttr(resourceName, "timestamp_override_fallback_disabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "threshold.value", "10"),
					resource.TestCheckResourceAttr(resourceName, "threshold.field.0", "user.name"),

					// Check filters field
					checkResourceJSONAttrKey("filters", `[{"bool": {"filter": [{"range": {"event.ingested": {"gte": "now-24h"}}}]}}]`),

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

					// Check alert suppression (threshold rules only support duration)
					resource.TestCheckResourceAttr(resourceName, "alert_suppression.duration", "30m"),

					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "rule_id"),
				),
			},
			{
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minResponseActionVersionSupport),
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables: config.Variables{
					"name": config.StringVariable("test-threshold-rule-updated"),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", "test-threshold-rule-updated"),
					resource.TestCheckResourceAttr(resourceName, "query", "event.action:(login OR logout)"),
					resource.TestCheckResourceAttr(resourceName, "description", "Updated test threshold security detection rule"),
					resource.TestCheckResourceAttr(resourceName, "severity", "high"),
					resource.TestCheckResourceAttr(resourceName, "risk_score", "75"),
					resource.TestCheckResourceAttr(resourceName, "data_view_id", "updated-threshold-data-view-id"),
					resource.TestCheckResourceAttr(resourceName, "namespace", "updated-threshold-namespace"),
					resource.TestCheckResourceAttr(resourceName, "rule_name_override", "Updated Custom Threshold Rule Name"),
					resource.TestCheckResourceAttr(resourceName, "timestamp_override", "event.start"),
					resource.TestCheckResourceAttr(resourceName, "timestamp_override_fallback_disabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "threshold.value", "20"),
					resource.TestCheckResourceAttr(resourceName, "threshold.field.0", "user.name"),
					resource.TestCheckResourceAttr(resourceName, "threshold.field.1", "source.ip"),

					// Check filters field (updated values)
					checkResourceJSONAttrKey("filters", `[{"bool": {"should": [{"match": {"user.roles": "admin"}}, {"term": {"event.severity": "high"}}], "minimum_should_match": 1}}]`),

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

					// Check updated alert suppression (threshold rules only support duration)
					resource.TestCheckResourceAttr(resourceName, "alert_suppression.duration", "45h"),
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
			ruleID := parts[1]

			// Check if the rule still exists
			ruleObjectID := uuid.MustParse(ruleID)
			params := &kbapi.ReadRuleParams{
				Id: &ruleObjectID,
			}

			response, err := kbClient.API.ReadRuleWithResponse(context.Background(), parts[0], params)
			if err != nil {
				return fmt.Errorf("failed to read security detection rule: %w", err)
			}

			// If the rule still exists (status 200), it means destroy failed
			if response.StatusCode() == 200 {
				return fmt.Errorf("security detection rule (%s) still exists", ruleID)
			}

			// If we get a 404, that's expected - the rule was properly destroyed
			// Any other status code indicates an error
			if response.StatusCode() != 404 {
				return fmt.Errorf("unexpected status code when checking security detection rule: %d", response.StatusCode())
			}

		case "elasticstack_kibana_action_connector":
			// Parse ID to get space_id and connector_id
			compID, _ := clients.CompositeIDFromStr(rs.Primary.ID)

			// Get connector client from the Kibana OAPI client
			oapiClient, err := client.GetKibanaOapiClient()
			if err != nil {
				return err
			}

			connector, diags := kibanaoapi.GetConnector(context.Background(), oapiClient, compID.ResourceID, compID.ClusterID)
			if diags.HasError() {
				return fmt.Errorf("failed to get connector: %v", diags)
			}

			if connector != nil {
				return fmt.Errorf("action connector (%s) still exists", compID.ResourceID)
			}
		}
	}

	return nil
}

func TestAccResourceSecurityDetectionRule_WithConnectorAction(t *testing.T) {
	resourceName := securityDetectionRuleResourceName
	connectorResourceName := "elasticstack_kibana_action_connector.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: testAccCheckSecurityDetectionRuleDestroy,
		Steps: []resource.TestStep{
			{
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minResponseActionVersionSupport),
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"name": config.StringVariable("test-rule-with-action"),
				},
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
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minResponseActionVersionSupport),
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables: config.Variables{
					"name": config.StringVariable("test-rule-with-action-updated"),
				},
				Check: resource.ComposeTestCheckFunc(
					// Check updated rule attributes
					resource.TestCheckResourceAttr(resourceName, "name", "test-rule-with-action-updated"),
					resource.TestCheckResourceAttr(resourceName, "description", "Updated test security detection rule with connector action"),
					resource.TestCheckResourceAttr(resourceName, "severity", "high"),
					resource.TestCheckResourceAttr(resourceName, "risk_score", "75"),
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

func TestAccResourceSecurityDetectionRule_BuildingBlockType(t *testing.T) {
	resourceName := securityDetectionRuleResourceName

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: testAccCheckSecurityDetectionRuleDestroy,
		Steps: []resource.TestStep{
			{
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minResponseActionVersionSupport),
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"name": config.StringVariable("test-building-block-rule"),
				},
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
					resource.TestCheckResourceAttr(resourceName, "namespace", "building-block-namespace"),
					resource.TestCheckResourceAttr(resourceName, "building_block_type", "default"),

					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "rule_id"),
				),
			},
			{
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minResponseActionVersionSupport),
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables: config.Variables{
					"name": config.StringVariable("test-building-block-rule-updated"),
				},
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
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minResponseActionVersionSupport),
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("removed"),
				ConfigVariables: config.Variables{
					"name": config.StringVariable("test-building-block-rule-no-type"),
				},
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

func TestAccResourceSecurityDetectionRule_QueryMinimal(t *testing.T) {
	resourceName := securityDetectionRuleResourceName

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: testAccCheckSecurityDetectionRuleDestroy,
		Steps: []resource.TestStep{
			{
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minVersionSupport),
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"name": config.StringVariable("test-query-rule-minimal"),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", "test-query-rule-minimal"),
					resource.TestCheckResourceAttr(resourceName, "type", "query"),
					resource.TestCheckResourceAttr(resourceName, "query", "*:*"),
					resource.TestCheckResourceAttr(resourceName, "language", "kuery"),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "description", "Minimal test query security detection rule"),
					resource.TestCheckResourceAttr(resourceName, "severity", "low"),
					resource.TestCheckResourceAttr(resourceName, "risk_score", "21"),
					resource.TestCheckResourceAttr(resourceName, "index.0", "logs-*"),

					// Verify only required fields are set
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "rule_id"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "created_by"),

					// Verify optional fields are not set
					resource.TestCheckNoResourceAttr(resourceName, "data_view_id"),
					resource.TestCheckNoResourceAttr(resourceName, "namespace"),
					resource.TestCheckNoResourceAttr(resourceName, "rule_name_override"),
					resource.TestCheckNoResourceAttr(resourceName, "timestamp_override"),
					resource.TestCheckNoResourceAttr(resourceName, "timestamp_override_fallback_disabled"),
					resource.TestCheckNoResourceAttr(resourceName, "filters"),
					resource.TestCheckNoResourceAttr(resourceName, "investigation_fields"),
					resource.TestCheckNoResourceAttr(resourceName, "risk_score_mapping"),
					resource.TestCheckNoResourceAttr(resourceName, "related_integrations"),
					resource.TestCheckNoResourceAttr(resourceName, "required_fields"),
					resource.TestCheckNoResourceAttr(resourceName, "severity_mapping"),
					resource.TestCheckNoResourceAttr(resourceName, "response_actions"),
					resource.TestCheckNoResourceAttr(resourceName, "alert_suppression"),
					resource.TestCheckNoResourceAttr(resourceName, "building_block_type"),
				),
			},
			{
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minVersionSupport),
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables: config.Variables{
					"name": config.StringVariable("test-query-rule-minimal-updated"),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", "test-query-rule-minimal-updated"),
					resource.TestCheckResourceAttr(resourceName, "type", "query"),
					resource.TestCheckResourceAttr(resourceName, "query", "event.category:authentication"),
					resource.TestCheckResourceAttr(resourceName, "language", "kuery"),
					resource.TestCheckResourceAttr(resourceName, "enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "description", "Updated minimal test query security detection rule"),
					resource.TestCheckResourceAttr(resourceName, "severity", "medium"),
					resource.TestCheckResourceAttr(resourceName, "risk_score", "55"),
					resource.TestCheckResourceAttr(resourceName, "index.0", "logs-*"),
					resource.TestCheckResourceAttr(resourceName, "index.1", "winlogbeat-*"),

					// Verify required fields are still set
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "rule_id"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "created_by"),

					// Verify optional fields are still not set
					resource.TestCheckNoResourceAttr(resourceName, "data_view_id"),
					resource.TestCheckNoResourceAttr(resourceName, "namespace"),
					resource.TestCheckNoResourceAttr(resourceName, "rule_name_override"),
					resource.TestCheckNoResourceAttr(resourceName, "timestamp_override"),
					resource.TestCheckNoResourceAttr(resourceName, "timestamp_override_fallback_disabled"),
					resource.TestCheckNoResourceAttr(resourceName, "filters"),
					resource.TestCheckNoResourceAttr(resourceName, "investigation_fields"),
					resource.TestCheckNoResourceAttr(resourceName, "risk_score_mapping"),
					resource.TestCheckNoResourceAttr(resourceName, "related_integrations"),
					resource.TestCheckNoResourceAttr(resourceName, "required_fields"),
					resource.TestCheckNoResourceAttr(resourceName, "severity_mapping"),
					resource.TestCheckNoResourceAttr(resourceName, "response_actions"),
					resource.TestCheckNoResourceAttr(resourceName, "alert_suppression"),
					resource.TestCheckNoResourceAttr(resourceName, "building_block_type"),
				),
			},
		},
	})
}

func TestAccResourceSecurityDetectionRule_QueryMinimalWithSpace(t *testing.T) {
	resourceName := securityDetectionRuleResourceName
	spaceResourceName := "elasticstack_kibana_space.test"
	spaceID := fmt.Sprintf("test-space-%s", uuid.New().String()[:8])

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: testAccCheckSecurityDetectionRuleDestroy,
		Steps: []resource.TestStep{
			{
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minVersionSupport),
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"name":     config.StringVariable("test-query-rule-with-space"),
					"space_id": config.StringVariable(spaceID),
				},
				Check: resource.ComposeTestCheckFunc(
					// Check space attributes
					resource.TestCheckResourceAttr(spaceResourceName, "space_id", spaceID),
					resource.TestCheckResourceAttr(spaceResourceName, "name", "Test Space for Detection Rules"),

					// Check detection rule attributes
					resource.TestCheckResourceAttr(resourceName, "name", "test-query-rule-with-space"),
					resource.TestCheckResourceAttr(resourceName, "type", "query"),
					resource.TestCheckResourceAttr(resourceName, "query", "*:*"),
					resource.TestCheckResourceAttr(resourceName, "language", "kuery"),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "description", "Minimal test query security detection rule in custom space"),
					resource.TestCheckResourceAttr(resourceName, "severity", "low"),
					resource.TestCheckResourceAttr(resourceName, "risk_score", "21"),
					resource.TestCheckResourceAttr(resourceName, "index.0", "logs-*"),
					resource.TestCheckResourceAttr(resourceName, "space_id", spaceID),

					// Verify required fields are set
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "rule_id"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "created_by"),

					// Verify optional fields are not set
					resource.TestCheckNoResourceAttr(resourceName, "data_view_id"),
					resource.TestCheckNoResourceAttr(resourceName, "namespace"),
					resource.TestCheckNoResourceAttr(resourceName, "rule_name_override"),
					resource.TestCheckNoResourceAttr(resourceName, "timestamp_override"),
					resource.TestCheckNoResourceAttr(resourceName, "timestamp_override_fallback_disabled"),
					resource.TestCheckNoResourceAttr(resourceName, "filters"),
					resource.TestCheckNoResourceAttr(resourceName, "investigation_fields"),
					resource.TestCheckNoResourceAttr(resourceName, "risk_score_mapping"),
					resource.TestCheckNoResourceAttr(resourceName, "related_integrations"),
					resource.TestCheckNoResourceAttr(resourceName, "required_fields"),
					resource.TestCheckNoResourceAttr(resourceName, "severity_mapping"),
					resource.TestCheckNoResourceAttr(resourceName, "response_actions"),
					resource.TestCheckNoResourceAttr(resourceName, "alert_suppression"),
					resource.TestCheckNoResourceAttr(resourceName, "building_block_type"),
				),
			},
			{
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minVersionSupport),
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables: config.Variables{
					"name":     config.StringVariable("test-query-rule-with-space-updated"),
					"space_id": config.StringVariable(spaceID),
				},
				Check: resource.ComposeTestCheckFunc(
					// Check space attributes remain the same
					resource.TestCheckResourceAttr(spaceResourceName, "space_id", spaceID),
					resource.TestCheckResourceAttr(spaceResourceName, "name", "Test Space for Detection Rules"),

					// Check updated detection rule attributes
					resource.TestCheckResourceAttr(resourceName, "name", "test-query-rule-with-space-updated"),
					resource.TestCheckResourceAttr(resourceName, "type", "query"),
					resource.TestCheckResourceAttr(resourceName, "query", "event.category:authentication"),
					resource.TestCheckResourceAttr(resourceName, "language", "kuery"),
					resource.TestCheckResourceAttr(resourceName, "enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "description", "Updated minimal test query security detection rule in custom space"),
					resource.TestCheckResourceAttr(resourceName, "severity", "medium"),
					resource.TestCheckResourceAttr(resourceName, "risk_score", "55"),
					resource.TestCheckResourceAttr(resourceName, "index.0", "logs-*"),
					resource.TestCheckResourceAttr(resourceName, "index.1", "winlogbeat-*"),
					resource.TestCheckResourceAttr(resourceName, "space_id", spaceID),

					// Verify required fields are still set
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "rule_id"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "created_by"),

					// Verify optional fields are still not set
					resource.TestCheckNoResourceAttr(resourceName, "data_view_id"),
					resource.TestCheckNoResourceAttr(resourceName, "namespace"),
					resource.TestCheckNoResourceAttr(resourceName, "rule_name_override"),
					resource.TestCheckNoResourceAttr(resourceName, "timestamp_override"),
					resource.TestCheckNoResourceAttr(resourceName, "timestamp_override_fallback_disabled"),
					resource.TestCheckNoResourceAttr(resourceName, "filters"),
					resource.TestCheckNoResourceAttr(resourceName, "investigation_fields"),
					resource.TestCheckNoResourceAttr(resourceName, "risk_score_mapping"),
					resource.TestCheckNoResourceAttr(resourceName, "related_integrations"),
					resource.TestCheckNoResourceAttr(resourceName, "required_fields"),
					resource.TestCheckNoResourceAttr(resourceName, "severity_mapping"),
					resource.TestCheckNoResourceAttr(resourceName, "response_actions"),
					resource.TestCheckNoResourceAttr(resourceName, "alert_suppression"),
					resource.TestCheckNoResourceAttr(resourceName, "building_block_type"),
				),
			},
			{
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minVersionSupport),
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables: config.Variables{
					"name":     config.StringVariable("test-query-rule-with-space-updated"),
					"space_id": config.StringVariable(spaceID),
				},
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccResourceSecurityDetectionRule_EQLMinimal(t *testing.T) {
	resourceName := securityDetectionRuleResourceName

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: testAccCheckSecurityDetectionRuleDestroy,
		Steps: []resource.TestStep{
			{
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minVersionSupport),
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"name": config.StringVariable("test-eql-rule-minimal"),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", "test-eql-rule-minimal"),
					resource.TestCheckResourceAttr(resourceName, "type", "eql"),
					resource.TestCheckResourceAttr(resourceName, "query", "process where process.name == \"cmd.exe\""),
					resource.TestCheckResourceAttr(resourceName, "language", "eql"),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "description", "Minimal test EQL security detection rule"),
					resource.TestCheckResourceAttr(resourceName, "severity", "low"),
					resource.TestCheckResourceAttr(resourceName, "risk_score", "21"),
					resource.TestCheckResourceAttr(resourceName, "index.0", "winlogbeat-*"),

					// Verify only required fields are set
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "rule_id"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "created_by"),

					// Verify optional fields are not set
					resource.TestCheckNoResourceAttr(resourceName, "data_view_id"),
					resource.TestCheckNoResourceAttr(resourceName, "namespace"),
					resource.TestCheckNoResourceAttr(resourceName, "rule_name_override"),
					resource.TestCheckNoResourceAttr(resourceName, "timestamp_override"),
					resource.TestCheckNoResourceAttr(resourceName, "timestamp_override_fallback_disabled"),
					resource.TestCheckNoResourceAttr(resourceName, "filters"),
					resource.TestCheckNoResourceAttr(resourceName, "investigation_fields"),
					resource.TestCheckNoResourceAttr(resourceName, "risk_score_mapping"),
					resource.TestCheckNoResourceAttr(resourceName, "related_integrations"),
					resource.TestCheckNoResourceAttr(resourceName, "required_fields"),
					resource.TestCheckNoResourceAttr(resourceName, "severity_mapping"),
					resource.TestCheckNoResourceAttr(resourceName, "response_actions"),
					resource.TestCheckNoResourceAttr(resourceName, "alert_suppression"),
					resource.TestCheckNoResourceAttr(resourceName, "building_block_type"),
					resource.TestCheckNoResourceAttr(resourceName, "tiebreaker_field"),
				),
			},
			{
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minVersionSupport),
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables: config.Variables{
					"name": config.StringVariable("test-eql-rule-minimal-updated"),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", "test-eql-rule-minimal-updated"),
					resource.TestCheckResourceAttr(resourceName, "query", "process where process.name == \"powershell.exe\""),
					resource.TestCheckResourceAttr(resourceName, "description", "Updated minimal test EQL security detection rule"),
					resource.TestCheckResourceAttr(resourceName, "severity", "medium"),
					resource.TestCheckResourceAttr(resourceName, "risk_score", "55"),
				),
			},
		},
	})
}

func TestAccResourceSecurityDetectionRule_ESQLMinimal(t *testing.T) {
	resourceName := securityDetectionRuleResourceName

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: testAccCheckSecurityDetectionRuleDestroy,
		Steps: []resource.TestStep{
			{
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minVersionSupport),
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"name": config.StringVariable("test-esql-rule-minimal"),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", "test-esql-rule-minimal"),
					resource.TestCheckResourceAttr(resourceName, "type", "esql"),
					resource.TestCheckResourceAttr(resourceName, "query", "FROM logs-* | WHERE event.action == \"login\" | STATS count(*) BY user.name"),
					resource.TestCheckResourceAttr(resourceName, "language", "esql"),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "description", "Minimal test ESQL security detection rule"),
					resource.TestCheckResourceAttr(resourceName, "severity", "low"),
					resource.TestCheckResourceAttr(resourceName, "risk_score", "21"),

					// Verify only required fields are set
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "rule_id"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "created_by"),

					// Verify optional fields are not set
					resource.TestCheckNoResourceAttr(resourceName, "data_view_id"),
					resource.TestCheckNoResourceAttr(resourceName, "namespace"),
					resource.TestCheckNoResourceAttr(resourceName, "rule_name_override"),
					resource.TestCheckNoResourceAttr(resourceName, "timestamp_override"),
					resource.TestCheckNoResourceAttr(resourceName, "timestamp_override_fallback_disabled"),
					resource.TestCheckNoResourceAttr(resourceName, "filters"),
					resource.TestCheckNoResourceAttr(resourceName, "investigation_fields"),
					resource.TestCheckNoResourceAttr(resourceName, "risk_score_mapping"),
					resource.TestCheckNoResourceAttr(resourceName, "related_integrations"),
					resource.TestCheckNoResourceAttr(resourceName, "required_fields"),
					resource.TestCheckNoResourceAttr(resourceName, "severity_mapping"),
					resource.TestCheckNoResourceAttr(resourceName, "response_actions"),
					resource.TestCheckNoResourceAttr(resourceName, "alert_suppression"),
					resource.TestCheckNoResourceAttr(resourceName, "building_block_type"),
					// Note: index is not checked for ESQL as it doesn't use index patterns
				),
			},
			{
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minVersionSupport),
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables: config.Variables{
					"name": config.StringVariable("test-esql-rule-minimal-updated"),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", "test-esql-rule-minimal-updated"),
					resource.TestCheckResourceAttr(resourceName, "query", "FROM logs-* | WHERE event.action == \"logout\" | STATS count(*) BY user.name, source.ip"),
					resource.TestCheckResourceAttr(resourceName, "description", "Updated minimal test ESQL security detection rule"),
					resource.TestCheckResourceAttr(resourceName, "severity", "medium"),
					resource.TestCheckResourceAttr(resourceName, "risk_score", "55"),
				),
			},
		},
	})
}

func TestAccResourceSecurityDetectionRule_MachineLearningMinimal(t *testing.T) {
	resourceName := securityDetectionRuleResourceName

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: testAccCheckSecurityDetectionRuleDestroy,
		Steps: []resource.TestStep{
			{
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minVersionSupport),
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"name": config.StringVariable("test-ml-rule-minimal"),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", "test-ml-rule-minimal"),
					resource.TestCheckResourceAttr(resourceName, "type", "machine_learning"),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "description", "Minimal test ML security detection rule"),
					resource.TestCheckResourceAttr(resourceName, "severity", "low"),
					resource.TestCheckResourceAttr(resourceName, "risk_score", "21"),
					resource.TestCheckResourceAttr(resourceName, "anomaly_threshold", "75"),
					resource.TestCheckResourceAttr(resourceName, "machine_learning_job_id.0", "test-ml-job"),

					// Verify only required fields are set
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "rule_id"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "created_by"),

					// Verify optional fields are not set
					resource.TestCheckNoResourceAttr(resourceName, "data_view_id"),
					resource.TestCheckNoResourceAttr(resourceName, "namespace"),
					resource.TestCheckNoResourceAttr(resourceName, "rule_name_override"),
					resource.TestCheckNoResourceAttr(resourceName, "timestamp_override"),
					resource.TestCheckNoResourceAttr(resourceName, "timestamp_override_fallback_disabled"),
					resource.TestCheckNoResourceAttr(resourceName, "filters"),
					resource.TestCheckNoResourceAttr(resourceName, "investigation_fields"),
					resource.TestCheckNoResourceAttr(resourceName, "risk_score_mapping"),
					resource.TestCheckNoResourceAttr(resourceName, "related_integrations"),
					resource.TestCheckNoResourceAttr(resourceName, "required_fields"),
					resource.TestCheckNoResourceAttr(resourceName, "severity_mapping"),
					resource.TestCheckNoResourceAttr(resourceName, "response_actions"),
					resource.TestCheckNoResourceAttr(resourceName, "alert_suppression"),
					resource.TestCheckNoResourceAttr(resourceName, "building_block_type"),
				),
			},
			{
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minVersionSupport),
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables: config.Variables{
					"name": config.StringVariable("test-ml-rule-minimal-updated"),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", "test-ml-rule-minimal-updated"),
					resource.TestCheckResourceAttr(resourceName, "description", "Updated minimal test ML security detection rule"),
					resource.TestCheckResourceAttr(resourceName, "severity", "medium"),
					resource.TestCheckResourceAttr(resourceName, "risk_score", "55"),
					resource.TestCheckResourceAttr(resourceName, "anomaly_threshold", "80"),
					resource.TestCheckResourceAttr(resourceName, "machine_learning_job_id.0", "test-ml-job"),
					resource.TestCheckResourceAttr(resourceName, "machine_learning_job_id.1", "test-ml-job-2"),
				),
			},
		},
	})
}

func TestAccResourceSecurityDetectionRule_NewTermsMinimal(t *testing.T) {
	resourceName := securityDetectionRuleResourceName

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: testAccCheckSecurityDetectionRuleDestroy,
		Steps: []resource.TestStep{
			{
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minVersionSupport),
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"name": config.StringVariable("test-new-terms-rule-minimal"),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", "test-new-terms-rule-minimal"),
					resource.TestCheckResourceAttr(resourceName, "type", "new_terms"),
					resource.TestCheckResourceAttr(resourceName, "query", "user.name:*"),
					resource.TestCheckResourceAttr(resourceName, "language", "kuery"),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "description", "Minimal test new terms security detection rule"),
					resource.TestCheckResourceAttr(resourceName, "severity", "low"),
					resource.TestCheckResourceAttr(resourceName, "risk_score", "21"),
					resource.TestCheckResourceAttr(resourceName, "index.0", "logs-*"),
					resource.TestCheckResourceAttr(resourceName, "new_terms_fields.0", "user.name"),
					resource.TestCheckResourceAttr(resourceName, "history_window_start", "now-14d"),

					// Verify only required fields are set
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "rule_id"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "created_by"),

					// Verify optional fields are not set
					resource.TestCheckNoResourceAttr(resourceName, "data_view_id"),
					resource.TestCheckNoResourceAttr(resourceName, "namespace"),
					resource.TestCheckNoResourceAttr(resourceName, "rule_name_override"),
					resource.TestCheckNoResourceAttr(resourceName, "timestamp_override"),
					resource.TestCheckNoResourceAttr(resourceName, "timestamp_override_fallback_disabled"),
					resource.TestCheckNoResourceAttr(resourceName, "filters"),
					resource.TestCheckNoResourceAttr(resourceName, "investigation_fields"),
					resource.TestCheckNoResourceAttr(resourceName, "risk_score_mapping"),
					resource.TestCheckNoResourceAttr(resourceName, "related_integrations"),
					resource.TestCheckNoResourceAttr(resourceName, "required_fields"),
					resource.TestCheckNoResourceAttr(resourceName, "severity_mapping"),
					resource.TestCheckNoResourceAttr(resourceName, "response_actions"),
					resource.TestCheckNoResourceAttr(resourceName, "alert_suppression"),
					resource.TestCheckNoResourceAttr(resourceName, "building_block_type"),
				),
			},
			{
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minVersionSupport),
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables: config.Variables{
					"name": config.StringVariable("test-new-terms-rule-minimal-updated"),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", "test-new-terms-rule-minimal-updated"),
					resource.TestCheckResourceAttr(resourceName, "query", "host.name:*"),
					resource.TestCheckResourceAttr(resourceName, "description", "Updated minimal test new terms security detection rule"),
					resource.TestCheckResourceAttr(resourceName, "severity", "medium"),
					resource.TestCheckResourceAttr(resourceName, "risk_score", "55"),
					resource.TestCheckResourceAttr(resourceName, "new_terms_fields.0", "host.name"),
					resource.TestCheckResourceAttr(resourceName, "history_window_start", "now-7d"),
				),
			},
		},
	})
}

func TestAccResourceSecurityDetectionRule_SavedQueryMinimal(t *testing.T) {
	resourceName := securityDetectionRuleResourceName

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: testAccCheckSecurityDetectionRuleDestroy,
		Steps: []resource.TestStep{
			{
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minVersionSupport),
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"name": config.StringVariable("test-saved-query-rule-minimal"),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", "test-saved-query-rule-minimal"),
					resource.TestCheckResourceAttr(resourceName, "type", "saved_query"),
					resource.TestCheckResourceAttr(resourceName, "query", "*:*"),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "description", "Minimal test saved query security detection rule"),
					resource.TestCheckResourceAttr(resourceName, "severity", "low"),
					resource.TestCheckResourceAttr(resourceName, "risk_score", "21"),
					resource.TestCheckResourceAttr(resourceName, "index.0", "logs-*"),
					resource.TestCheckResourceAttr(resourceName, "saved_id", "test-saved-query-id"),

					// Verify only required fields are set
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "rule_id"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "created_by"),

					// Verify optional fields are not set
					resource.TestCheckNoResourceAttr(resourceName, "data_view_id"),
					resource.TestCheckNoResourceAttr(resourceName, "namespace"),
					resource.TestCheckNoResourceAttr(resourceName, "rule_name_override"),
					resource.TestCheckNoResourceAttr(resourceName, "timestamp_override"),
					resource.TestCheckNoResourceAttr(resourceName, "timestamp_override_fallback_disabled"),
					resource.TestCheckNoResourceAttr(resourceName, "filters"),
					resource.TestCheckNoResourceAttr(resourceName, "investigation_fields"),
					resource.TestCheckNoResourceAttr(resourceName, "risk_score_mapping"),
					resource.TestCheckNoResourceAttr(resourceName, "related_integrations"),
					resource.TestCheckNoResourceAttr(resourceName, "required_fields"),
					resource.TestCheckNoResourceAttr(resourceName, "severity_mapping"),
					resource.TestCheckNoResourceAttr(resourceName, "response_actions"),
					resource.TestCheckNoResourceAttr(resourceName, "alert_suppression"),
					resource.TestCheckNoResourceAttr(resourceName, "building_block_type"),
				),
			},
			{
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minVersionSupport),
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables: config.Variables{
					"name": config.StringVariable("test-saved-query-rule-minimal-updated"),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", "test-saved-query-rule-minimal-updated"),
					resource.TestCheckResourceAttr(resourceName, "query", "event.category:authentication"),
					resource.TestCheckResourceAttr(resourceName, "description", "Updated minimal test saved query security detection rule"),
					resource.TestCheckResourceAttr(resourceName, "severity", "medium"),
					resource.TestCheckResourceAttr(resourceName, "risk_score", "55"),
					resource.TestCheckResourceAttr(resourceName, "saved_id", "test-saved-query-id-updated"),
				),
			},
		},
	})
}

func TestAccResourceSecurityDetectionRule_ThreatMatchMinimal(t *testing.T) {
	resourceName := securityDetectionRuleResourceName

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: testAccCheckSecurityDetectionRuleDestroy,
		Steps: []resource.TestStep{
			{
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minVersionSupport),
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"name": config.StringVariable("test-threat-match-rule-minimal"),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", "test-threat-match-rule-minimal"),
					resource.TestCheckResourceAttr(resourceName, "type", "threat_match"),
					resource.TestCheckResourceAttr(resourceName, "query", "destination.ip:*"),
					resource.TestCheckResourceAttr(resourceName, "language", "kuery"),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "description", "Minimal test threat match security detection rule"),
					resource.TestCheckResourceAttr(resourceName, "severity", "low"),
					resource.TestCheckResourceAttr(resourceName, "risk_score", "21"),
					resource.TestCheckResourceAttr(resourceName, "index.0", "logs-*"),
					resource.TestCheckResourceAttr(resourceName, "threat_index.0", "threat-intel-*"),
					resource.TestCheckResourceAttr(resourceName, "threat_query", "threat.indicator.type:ip"),
					resource.TestCheckResourceAttr(resourceName, "threat_mapping.0.entries.0.field", "destination.ip"),
					resource.TestCheckResourceAttr(resourceName, "threat_mapping.0.entries.0.type", "mapping"),
					resource.TestCheckResourceAttr(resourceName, "threat_mapping.0.entries.0.value", "threat.indicator.ip"),

					// Verify only required fields are set
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "rule_id"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "created_by"),

					// Verify optional fields are not set
					resource.TestCheckNoResourceAttr(resourceName, "data_view_id"),
					resource.TestCheckNoResourceAttr(resourceName, "namespace"),
					resource.TestCheckNoResourceAttr(resourceName, "rule_name_override"),
					resource.TestCheckNoResourceAttr(resourceName, "timestamp_override"),
					resource.TestCheckNoResourceAttr(resourceName, "timestamp_override_fallback_disabled"),
					resource.TestCheckNoResourceAttr(resourceName, "filters"),
					resource.TestCheckNoResourceAttr(resourceName, "investigation_fields"),
					resource.TestCheckNoResourceAttr(resourceName, "risk_score_mapping"),
					resource.TestCheckNoResourceAttr(resourceName, "related_integrations"),
					resource.TestCheckNoResourceAttr(resourceName, "required_fields"),
					resource.TestCheckNoResourceAttr(resourceName, "severity_mapping"),
					resource.TestCheckNoResourceAttr(resourceName, "response_actions"),
					resource.TestCheckNoResourceAttr(resourceName, "alert_suppression"),
					resource.TestCheckNoResourceAttr(resourceName, "building_block_type"),
				),
			},
			{
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minVersionSupport),
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables: config.Variables{
					"name": config.StringVariable("test-threat-match-rule-minimal-updated"),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", "test-threat-match-rule-minimal-updated"),
					resource.TestCheckResourceAttr(resourceName, "query", "source.ip:*"),
					resource.TestCheckResourceAttr(resourceName, "description", "Updated minimal test threat match security detection rule"),
					resource.TestCheckResourceAttr(resourceName, "severity", "medium"),
					resource.TestCheckResourceAttr(resourceName, "risk_score", "55"),
					resource.TestCheckResourceAttr(resourceName, "threat_query", "threat.indicator.type:domain"),
					resource.TestCheckResourceAttr(resourceName, "threat_mapping.0.entries.0.field", "source.ip"),
					resource.TestCheckResourceAttr(resourceName, "threat_mapping.0.entries.0.value", "threat.indicator.domain"),
				),
			},
		},
	})
}

func TestAccResourceSecurityDetectionRule_ThresholdMinimal(t *testing.T) {
	resourceName := securityDetectionRuleResourceName

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: testAccCheckSecurityDetectionRuleDestroy,
		Steps: []resource.TestStep{
			{
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minVersionSupport),
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"name": config.StringVariable("test-threshold-rule-minimal"),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", "test-threshold-rule-minimal"),
					resource.TestCheckResourceAttr(resourceName, "type", "threshold"),
					resource.TestCheckResourceAttr(resourceName, "query", "event.action:login"),
					resource.TestCheckResourceAttr(resourceName, "language", "kuery"),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "description", "Minimal test threshold security detection rule"),
					resource.TestCheckResourceAttr(resourceName, "severity", "low"),
					resource.TestCheckResourceAttr(resourceName, "risk_score", "21"),
					resource.TestCheckResourceAttr(resourceName, "index.0", "logs-*"),
					resource.TestCheckResourceAttr(resourceName, "threshold.value", "10"),
					resource.TestCheckResourceAttr(resourceName, "threshold.field.0", "user.name"),

					// Verify only required fields are set
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "rule_id"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "created_by"),

					// Verify optional fields are not set
					resource.TestCheckNoResourceAttr(resourceName, "data_view_id"),
					resource.TestCheckNoResourceAttr(resourceName, "namespace"),
					resource.TestCheckNoResourceAttr(resourceName, "rule_name_override"),
					resource.TestCheckNoResourceAttr(resourceName, "timestamp_override"),
					resource.TestCheckNoResourceAttr(resourceName, "timestamp_override_fallback_disabled"),
					resource.TestCheckNoResourceAttr(resourceName, "filters"),
					resource.TestCheckNoResourceAttr(resourceName, "investigation_fields"),
					resource.TestCheckNoResourceAttr(resourceName, "risk_score_mapping"),
					resource.TestCheckNoResourceAttr(resourceName, "related_integrations"),
					resource.TestCheckNoResourceAttr(resourceName, "required_fields"),
					resource.TestCheckNoResourceAttr(resourceName, "severity_mapping"),
					resource.TestCheckNoResourceAttr(resourceName, "response_actions"),
					resource.TestCheckNoResourceAttr(resourceName, "alert_suppression"),
					resource.TestCheckNoResourceAttr(resourceName, "building_block_type"),
				),
			},
			{
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minVersionSupport),
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables: config.Variables{
					"name": config.StringVariable("test-threshold-rule-minimal-updated"),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", "test-threshold-rule-minimal-updated"),
					resource.TestCheckResourceAttr(resourceName, "query", "event.action:logout"),
					resource.TestCheckResourceAttr(resourceName, "description", "Updated minimal test threshold security detection rule"),
					resource.TestCheckResourceAttr(resourceName, "severity", "medium"),
					resource.TestCheckResourceAttr(resourceName, "risk_score", "55"),
					resource.TestCheckResourceAttr(resourceName, "threshold.value", "20"),
					resource.TestCheckResourceAttr(resourceName, "threshold.field.0", "host.name"),
				),
			},
		},
	})
}

func TestAccResourceSecurityDetectionRule_QueryWithMitreThreat(t *testing.T) {
	resourceName := securityDetectionRuleResourceName

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: testAccCheckSecurityDetectionRuleDestroy,
		Steps: []resource.TestStep{
			{
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minVersionSupport),
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"name": config.StringVariable("test-query-mitre-rule"),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", "test-query-mitre-rule"),
					resource.TestCheckResourceAttr(resourceName, "type", "query"),
					resource.TestCheckResourceAttr(resourceName, "query", "process.parent.name:(EXCEL.EXE OR WINWORD.EXE OR POWERPNT.EXE OR OUTLOOK.EXE)"),
					resource.TestCheckResourceAttr(resourceName, "language", "kuery"),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "description", "Detects processes started by MS Office programs"),
					resource.TestCheckResourceAttr(resourceName, "severity", "low"),
					resource.TestCheckResourceAttr(resourceName, "risk_score", "50"),
					resource.TestCheckResourceAttr(resourceName, "from", "now-70m"),
					resource.TestCheckResourceAttr(resourceName, "to", "now"),
					resource.TestCheckResourceAttr(resourceName, "interval", "1h"),
					resource.TestCheckResourceAttr(resourceName, "index.0", "logs-*"),
					resource.TestCheckResourceAttr(resourceName, "index.1", "winlogbeat-*"),
					resource.TestCheckResourceAttr(resourceName, "max_signals", "100"),

					// Check tags
					resource.TestCheckResourceAttr(resourceName, "tags.#", "3"),
					resource.TestCheckResourceAttr(resourceName, "tags.0", "child process"),
					resource.TestCheckResourceAttr(resourceName, "tags.1", "ms office"),
					resource.TestCheckResourceAttr(resourceName, "tags.2", "terraform-test"),

					// Check references
					resource.TestCheckResourceAttr(resourceName, "references.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "references.0", "https://attack.mitre.org/techniques/T1566/001/"),

					// Check false positives
					resource.TestCheckResourceAttr(resourceName, "false_positives.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "false_positives.0", "Legitimate corporate macros"),

					// Check author
					resource.TestCheckResourceAttr(resourceName, "author.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "author.0", "Security Team"),

					// Check license
					resource.TestCheckResourceAttr(resourceName, "license", "Elastic License v2"),

					// Check note
					resource.TestCheckResourceAttr(resourceName, "note", "Investigate parent process and command line"),

					// Check threat (MITRE ATT&CK)
					resource.TestCheckResourceAttr(resourceName, "threat.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "threat.0.framework", "MITRE ATT&CK"),
					resource.TestCheckResourceAttr(resourceName, "threat.0.tactic.id", "TA0009"),
					resource.TestCheckResourceAttr(resourceName, "threat.0.tactic.name", "Collection"),
					resource.TestCheckResourceAttr(resourceName, "threat.0.tactic.reference", "https://attack.mitre.org/tactics/TA0009"),
					resource.TestCheckResourceAttr(resourceName, "threat.0.technique.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "threat.0.technique.0.id", "T1123"),
					resource.TestCheckResourceAttr(resourceName, "threat.0.technique.0.name", "Audio Capture"),
					resource.TestCheckResourceAttr(resourceName, "threat.0.technique.0.reference", "https://attack.mitre.org/techniques/T1123"),

					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "rule_id"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "created_by"),
				),
			},
			{
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minVersionSupport),
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables: config.Variables{
					"name": config.StringVariable("test-query-mitre-rule-updated"),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", "test-query-mitre-rule-updated"),
					resource.TestCheckResourceAttr(resourceName, "description", "Updated detection rule for processes started by MS Office programs"),
					resource.TestCheckResourceAttr(resourceName, "severity", "medium"),
					resource.TestCheckResourceAttr(resourceName, "risk_score", "75"),
					resource.TestCheckResourceAttr(resourceName, "from", "now-2h"),
					resource.TestCheckResourceAttr(resourceName, "interval", "30m"),
					resource.TestCheckResourceAttr(resourceName, "max_signals", "200"),

					// Check updated tags
					resource.TestCheckResourceAttr(resourceName, "tags.#", "4"),
					resource.TestCheckResourceAttr(resourceName, "tags.0", "child process"),
					resource.TestCheckResourceAttr(resourceName, "tags.1", "ms office"),
					resource.TestCheckResourceAttr(resourceName, "tags.2", "terraform-test"),
					resource.TestCheckResourceAttr(resourceName, "tags.3", "updated"),

					// Check updated references
					resource.TestCheckResourceAttr(resourceName, "references.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "references.0", "https://attack.mitre.org/techniques/T1566/001/"),
					resource.TestCheckResourceAttr(resourceName, "references.1", "https://attack.mitre.org/techniques/T1204/002/"),

					// Check updated false positives
					resource.TestCheckResourceAttr(resourceName, "false_positives.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "false_positives.0", "Legitimate corporate macros"),
					resource.TestCheckResourceAttr(resourceName, "false_positives.1", "Authorized office automation"),

					// Check updated author
					resource.TestCheckResourceAttr(resourceName, "author.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "author.0", "Security Team"),
					resource.TestCheckResourceAttr(resourceName, "author.1", "SOC Team"),

					// Check updated note
					resource.TestCheckResourceAttr(resourceName, "note", "Investigate parent process and command line. Check for malicious documents."),

					// Check updated threat - multiple techniques
					resource.TestCheckResourceAttr(resourceName, "threat.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "threat.0.framework", "MITRE ATT&CK"),
					resource.TestCheckResourceAttr(resourceName, "threat.0.tactic.id", "TA0002"),
					resource.TestCheckResourceAttr(resourceName, "threat.0.tactic.name", "Execution"),
					resource.TestCheckResourceAttr(resourceName, "threat.0.tactic.reference", "https://attack.mitre.org/tactics/TA0002"),
					resource.TestCheckResourceAttr(resourceName, "threat.0.technique.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "threat.0.technique.0.id", "T1566"),
					resource.TestCheckResourceAttr(resourceName, "threat.0.technique.0.name", "Phishing"),
					resource.TestCheckResourceAttr(resourceName, "threat.0.technique.0.reference", "https://attack.mitre.org/techniques/T1566"),
					resource.TestCheckResourceAttr(resourceName, "threat.0.technique.0.subtechnique.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "threat.0.technique.0.subtechnique.0.id", "T1566.001"),
					resource.TestCheckResourceAttr(resourceName, "threat.0.technique.0.subtechnique.0.name", "Spearphishing Attachment"),
					resource.TestCheckResourceAttr(resourceName, "threat.0.technique.0.subtechnique.0.reference", "https://attack.mitre.org/techniques/T1566/001"),
					resource.TestCheckResourceAttr(resourceName, "threat.0.technique.1.id", "T1204"),
					resource.TestCheckResourceAttr(resourceName, "threat.0.technique.1.name", "User Execution"),
					resource.TestCheckResourceAttr(resourceName, "threat.0.technique.1.reference", "https://attack.mitre.org/techniques/T1204"),
					resource.TestCheckResourceAttr(resourceName, "threat.0.technique.1.subtechnique.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "threat.0.technique.1.subtechnique.0.id", "T1204.002"),
					resource.TestCheckResourceAttr(resourceName, "threat.0.technique.1.subtechnique.0.name", "Malicious File"),
					resource.TestCheckResourceAttr(resourceName, "threat.0.technique.1.subtechnique.0.reference", "https://attack.mitre.org/techniques/T1204/002"),
				),
			},
		},
	})
}

// TestAccResourceSecurityDetectionRule_ValidateConfig tests the ValidateConfig method
// to ensure proper validation of index vs data_view_id configuration
func TestAccResourceSecurityDetectionRule_ValidateConfig(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			// Test 1: Valid config with only index (should succeed)
			{
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minVersionSupport),
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("index_only"),
				ConfigVariables: config.Variables{
					"name": config.StringVariable("test-validation-index-only"),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(securityDetectionRuleResourceName, "name", "test-validation-index-only"),
					resource.TestCheckResourceAttr(securityDetectionRuleResourceName, "index.0", "logs-*"),
					resource.TestCheckNoResourceAttr(securityDetectionRuleResourceName, "data_view_id"),
				),
			},
			// Test 2: Valid config with only data_view_id (should succeed)
			{
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minVersionSupport),
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("dataview_only"),
				ConfigVariables: config.Variables{
					"name": config.StringVariable("test-validation-dataview-only"),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(securityDetectionRuleResourceName, "name", "test-validation-dataview-only"),
					resource.TestCheckResourceAttr(securityDetectionRuleResourceName, "data_view_id", "test-data-view-id"),
					resource.TestCheckNoResourceAttr(securityDetectionRuleResourceName, "index.0"),
				),
			},
			// Test 3: Invalid config with both index and data_view_id (should fail)
			{
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minVersionSupport),
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("both"),
				ConfigVariables: config.Variables{
					"name": config.StringVariable("test-validation-both"),
				},
				ExpectError: regexp.MustCompile("Both 'index' and 'data_view_id' cannot be set at the same time"),
				PlanOnly:    true,
			},
			// Test 4: Invalid config with neither index nor data_view_id (should fail)
			{
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minVersionSupport),
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("neither"),
				ConfigVariables: config.Variables{
					"name": config.StringVariable("test-validation-neither"),
				},
				ExpectError: regexp.MustCompile("One of 'index' or 'data_view_id' must be set"),
				PlanOnly:    true,
			},
			// Test 5: ESQL rule type should skip validation (both index and data_view_id allowed to be unset)
			{
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minVersionSupport),
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("esql_type"),
				ConfigVariables: config.Variables{
					"name": config.StringVariable("test-validation-esql"),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(securityDetectionRuleResourceName, "name", "test-validation-esql"),
					resource.TestCheckResourceAttr(securityDetectionRuleResourceName, "type", "esql"),
					resource.TestCheckNoResourceAttr(securityDetectionRuleResourceName, "index.0"),
					resource.TestCheckNoResourceAttr(securityDetectionRuleResourceName, "data_view_id"),
				),
			},
			// Test 6: Machine learning rule type should skip validation (both index and data_view_id allowed to be unset)
			{
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minVersionSupport),
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("ml_type"),
				ConfigVariables: config.Variables{
					"name": config.StringVariable("test-validation-ml"),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(securityDetectionRuleResourceName, "name", "test-validation-ml"),
					resource.TestCheckResourceAttr(securityDetectionRuleResourceName, "type", "machine_learning"),
					resource.TestCheckNoResourceAttr(securityDetectionRuleResourceName, "index.0"),
					resource.TestCheckNoResourceAttr(securityDetectionRuleResourceName, "data_view_id"),
				),
			},
		},
	})
}
