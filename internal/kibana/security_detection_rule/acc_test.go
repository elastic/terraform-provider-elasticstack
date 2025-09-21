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
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/google/uuid"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

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

					// Check risk score mapping
					resource.TestCheckResourceAttr(resourceName, "risk_score_mapping.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "risk_score_mapping.0.field", "event.severity"),
					resource.TestCheckResourceAttr(resourceName, "risk_score_mapping.0.operator", "equals"),
					resource.TestCheckResourceAttr(resourceName, "risk_score_mapping.0.value", "high"),
					resource.TestCheckResourceAttr(resourceName, "risk_score_mapping.0.risk_score", "85"),

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

					// Check risk score mapping
					resource.TestCheckResourceAttr(resourceName, "risk_score_mapping.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "risk_score_mapping.0.field", "event.risk_level"),
					resource.TestCheckResourceAttr(resourceName, "risk_score_mapping.0.operator", "equals"),
					resource.TestCheckResourceAttr(resourceName, "risk_score_mapping.0.value", "critical"),
					resource.TestCheckResourceAttr(resourceName, "risk_score_mapping.0.risk_score", "95"),
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

					// Check risk score mapping
					resource.TestCheckResourceAttr(resourceName, "risk_score_mapping.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "risk_score_mapping.0.field", "process.executable"),
					resource.TestCheckResourceAttr(resourceName, "risk_score_mapping.0.operator", "equals"),
					resource.TestCheckResourceAttr(resourceName, "risk_score_mapping.0.value", "C:\\Windows\\System32\\cmd.exe"),
					resource.TestCheckResourceAttr(resourceName, "risk_score_mapping.0.risk_score", "75"),

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

					// Check risk score mapping
					resource.TestCheckResourceAttr(resourceName, "risk_score_mapping.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "risk_score_mapping.0.field", "process.parent.name"),
					resource.TestCheckResourceAttr(resourceName, "risk_score_mapping.0.operator", "equals"),
					resource.TestCheckResourceAttr(resourceName, "risk_score_mapping.0.value", "cmd.exe"),
					resource.TestCheckResourceAttr(resourceName, "risk_score_mapping.0.risk_score", "95"),
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

					// Check risk score mapping
					resource.TestCheckResourceAttr(resourceName, "risk_score_mapping.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "risk_score_mapping.0.field", "user.domain"),
					resource.TestCheckResourceAttr(resourceName, "risk_score_mapping.0.operator", "equals"),
					resource.TestCheckResourceAttr(resourceName, "risk_score_mapping.0.value", "admin"),
					resource.TestCheckResourceAttr(resourceName, "risk_score_mapping.0.risk_score", "80"),

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

					// Check risk score mapping
					resource.TestCheckResourceAttr(resourceName, "risk_score_mapping.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "risk_score_mapping.0.field", "event.outcome"),
					resource.TestCheckResourceAttr(resourceName, "risk_score_mapping.0.operator", "equals"),
					resource.TestCheckResourceAttr(resourceName, "risk_score_mapping.0.value", "failure"),
					resource.TestCheckResourceAttr(resourceName, "risk_score_mapping.0.risk_score", "95"),

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

					// Check risk score mapping
					resource.TestCheckResourceAttr(resourceName, "risk_score_mapping.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "risk_score_mapping.0.field", "ml.anomaly_score"),
					resource.TestCheckResourceAttr(resourceName, "risk_score_mapping.0.operator", "equals"),
					resource.TestCheckResourceAttr(resourceName, "risk_score_mapping.0.value", "critical"),
					resource.TestCheckResourceAttr(resourceName, "risk_score_mapping.0.risk_score", "100"),

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

					// Check risk score mapping
					resource.TestCheckResourceAttr(resourceName, "risk_score_mapping.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "risk_score_mapping.0.field", "ml.is_anomaly"),
					resource.TestCheckResourceAttr(resourceName, "risk_score_mapping.0.operator", "equals"),
					resource.TestCheckResourceAttr(resourceName, "risk_score_mapping.0.value", "true"),
					resource.TestCheckResourceAttr(resourceName, "risk_score_mapping.0.risk_score", "95"),

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
					resource.TestCheckResourceAttr(resourceName, "history_window_start", "now-14d"),

					// Check risk score mapping
					resource.TestCheckResourceAttr(resourceName, "risk_score_mapping.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "risk_score_mapping.0.field", "user.type"),
					resource.TestCheckResourceAttr(resourceName, "risk_score_mapping.0.operator", "equals"),
					resource.TestCheckResourceAttr(resourceName, "risk_score_mapping.0.value", "service_account"),
					resource.TestCheckResourceAttr(resourceName, "risk_score_mapping.0.risk_score", "65"),

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
					resource.TestCheckResourceAttr(resourceName, "history_window_start", "now-30d"),

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
					resource.TestCheckResourceAttr(resourceName, "index.0", "logs-*"),

					// Check risk score mapping
					resource.TestCheckResourceAttr(resourceName, "risk_score_mapping.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "risk_score_mapping.0.field", "event.category"),
					resource.TestCheckResourceAttr(resourceName, "risk_score_mapping.0.operator", "equals"),
					resource.TestCheckResourceAttr(resourceName, "risk_score_mapping.0.value", "authentication"),
					resource.TestCheckResourceAttr(resourceName, "risk_score_mapping.0.risk_score", "45"),

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
					resource.TestCheckResourceAttr(resourceName, "index.0", "logs-*"),
					resource.TestCheckResourceAttr(resourceName, "index.1", "audit-*"),

					// Check risk score mapping
					resource.TestCheckResourceAttr(resourceName, "risk_score_mapping.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "risk_score_mapping.0.field", "event.type"),
					resource.TestCheckResourceAttr(resourceName, "risk_score_mapping.0.operator", "equals"),
					resource.TestCheckResourceAttr(resourceName, "risk_score_mapping.0.value", "access"),
					resource.TestCheckResourceAttr(resourceName, "risk_score_mapping.0.risk_score", "70"),

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
					resource.TestCheckResourceAttr(resourceName, "threat_index.0", "threat-intel-*"),
					resource.TestCheckResourceAttr(resourceName, "threat_query", "threat.indicator.type:ip"),
					resource.TestCheckResourceAttr(resourceName, "threat_mapping.0.entries.0.field", "destination.ip"),
					resource.TestCheckResourceAttr(resourceName, "threat_mapping.0.entries.0.type", "mapping"),
					resource.TestCheckResourceAttr(resourceName, "threat_mapping.0.entries.0.value", "threat.indicator.ip"),

					// Check risk score mapping
					resource.TestCheckResourceAttr(resourceName, "risk_score_mapping.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "risk_score_mapping.0.field", "threat.indicator.confidence"),
					resource.TestCheckResourceAttr(resourceName, "risk_score_mapping.0.operator", "equals"),
					resource.TestCheckResourceAttr(resourceName, "risk_score_mapping.0.value", "medium"),
					resource.TestCheckResourceAttr(resourceName, "risk_score_mapping.0.risk_score", "85"),

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
					resource.TestCheckResourceAttr(resourceName, "threat_index.0", "threat-intel-*"),
					resource.TestCheckResourceAttr(resourceName, "threat_index.1", "ioc-*"),
					resource.TestCheckResourceAttr(resourceName, "threat_query", "threat.indicator.type:(ip OR domain)"),
					resource.TestCheckResourceAttr(resourceName, "threat_mapping.0.entries.0.field", "destination.ip"),
					resource.TestCheckResourceAttr(resourceName, "threat_mapping.1.entries.0.field", "source.ip"),

					// Check risk score mapping
					resource.TestCheckResourceAttr(resourceName, "risk_score_mapping.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "risk_score_mapping.0.field", "threat.indicator.confidence"),
					resource.TestCheckResourceAttr(resourceName, "risk_score_mapping.0.operator", "equals"),
					resource.TestCheckResourceAttr(resourceName, "risk_score_mapping.0.value", "high"),
					resource.TestCheckResourceAttr(resourceName, "risk_score_mapping.0.risk_score", "100"),
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
					resource.TestCheckResourceAttr(resourceName, "threshold.value", "10"),
					resource.TestCheckResourceAttr(resourceName, "threshold.field.0", "user.name"),

					// Check risk score mapping
					resource.TestCheckResourceAttr(resourceName, "risk_score_mapping.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "risk_score_mapping.0.field", "event.outcome"),
					resource.TestCheckResourceAttr(resourceName, "risk_score_mapping.0.operator", "equals"),
					resource.TestCheckResourceAttr(resourceName, "risk_score_mapping.0.value", "success"),
					resource.TestCheckResourceAttr(resourceName, "risk_score_mapping.0.risk_score", "45"),

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
					resource.TestCheckResourceAttr(resourceName, "threshold.value", "20"),
					resource.TestCheckResourceAttr(resourceName, "threshold.field.0", "user.name"),
					resource.TestCheckResourceAttr(resourceName, "threshold.field.1", "source.ip"),

					// Check risk score mapping
					resource.TestCheckResourceAttr(resourceName, "risk_score_mapping.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "risk_score_mapping.0.field", "event.outcome"),
					resource.TestCheckResourceAttr(resourceName, "risk_score_mapping.0.operator", "equals"),
					resource.TestCheckResourceAttr(resourceName, "risk_score_mapping.0.value", "failure"),
					resource.TestCheckResourceAttr(resourceName, "risk_score_mapping.0.risk_score", "90"),
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
  name        = "%s"
  type        = "query"
  query       = "*:*"
  language    = "kuery"
  enabled     = true
  description = "Test query security detection rule"
  severity    = "medium"
  risk_score  = 50
  from        = "now-6m"
  to          = "now"
  interval    = "5m"
  index       = ["logs-*"]

  risk_score_mapping = [
    {
      field      = "event.severity"
      operator   = "equals"
      value      = "high"
      risk_score = 85
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

  risk_score_mapping = [
    {
      field      = "event.risk_level"
      operator   = "equals"
      value      = "critical"
      risk_score = 95
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

  risk_score_mapping = [
    {
      field      = "process.executable"
      operator   = "equals"
      value      = "C:\\Windows\\System32\\cmd.exe"
      risk_score = 75
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

  risk_score_mapping = [
    {
      field      = "process.parent.name"
      operator   = "equals"
      value      = "cmd.exe"
      risk_score = 95
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

  risk_score_mapping = [
    {
      field      = "user.domain"
      operator   = "equals"
      value      = "admin"
      risk_score = 80
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
  
  risk_score_mapping = [
    {
      field      = "event.outcome"
      operator   = "equals"
      value      = "failure"
      risk_score = 95
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

  risk_score_mapping = [
    {
      field      = "ml.anomaly_score"
      operator   = "equals"
      value      = "critical"
      risk_score = 100
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

  risk_score_mapping = [
    {
      field      = "ml.is_anomaly"
      operator   = "equals"
      value      = "true"
      risk_score = 95
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

  risk_score_mapping = [
    {
      field      = "user.type"
      operator   = "equals"
      value      = "service_account"
      risk_score = 65
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
}
`, name)
}

func testAccSecurityDetectionRuleConfig_savedQuery(name string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_security_detection_rule" "test" {
  name        = "%s"
  type        = "saved_query"
  query       = "*:*"
  enabled     = true
  description = "Test saved query security detection rule"
  severity    = "low"
  risk_score  = 30
  from        = "now-6m"
  to          = "now"
  interval    = "5m"
  index       = ["logs-*"]
  saved_id    = "test-saved-query-id"

  risk_score_mapping = [
    {
      field      = "event.category"
      operator   = "equals"
      value      = "authentication"
      risk_score = 45
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
  name        = "%s"
  type        = "saved_query"
  query       = "event.action:*"
  enabled     = true
  description = "Updated test saved query security detection rule"
  severity    = "medium"
  risk_score  = 60
  from        = "now-6m"
  to          = "now"
  interval    = "5m"
  index       = ["logs-*", "audit-*"]
  saved_id    = "test-saved-query-id-updated"
  author      = ["Test Author"]
  tags        = ["test", "saved-query", "automation"]
  license     = "Elastic License v2"

  risk_score_mapping = [
    {
      field      = "event.type"
      operator   = "equals"
      value      = "access"
      risk_score = 70
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
  threat_index = ["threat-intel-*"]
  threat_query = "threat.indicator.type:ip"
  
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
  threat_index = ["threat-intel-*", "ioc-*"]
  threat_query = "threat.indicator.type:(ip OR domain)"
  author       = ["Test Author"]
  tags         = ["test", "threat-match", "automation"]
  license      = "Elastic License v2"
  
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
}
`, name)
}

func testAccSecurityDetectionRuleConfig_threshold(name string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_security_detection_rule" "test" {
  name        = "%s"
  type        = "threshold"
  query       = "event.action:login"
  language    = "kuery"
  enabled     = true
  description = "Test threshold security detection rule"
  severity    = "medium"
  risk_score  = 55
  from        = "now-6m"
  to          = "now"
  interval    = "5m"
  index       = ["logs-*"]
  
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
}
`, name)
}

func testAccSecurityDetectionRuleConfig_thresholdUpdate(name string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_security_detection_rule" "test" {
  name        = "%s"
  type        = "threshold"
  query       = "event.action:(login OR logout)"
  language    = "kuery"
  enabled     = true
  description = "Updated test threshold security detection rule"
  severity    = "high"
  risk_score  = 75
  from        = "now-6m"
  to          = "now"
  interval    = "5m"
  index       = ["logs-*", "audit-*"]
  author      = ["Test Author"]
  tags        = ["test", "threshold", "automation"]
  license     = "Elastic License v2"
  
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
  name        = "%s"
  description = "Test security detection rule with connector action"
  type        = "query"
  severity    = "medium"
  risk_score  = 50
  enabled     = true
  query       = "user.name:*"
  language    = "kuery"
  from        = "now-6m"
  to          = "now"
  interval    = "5m"
  index       = ["logs-*"]

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
  name        = "%s"
  description = "Updated test security detection rule with connector action"
  type        = "query"
  severity    = "high"
  risk_score  = 75
  enabled     = true
  query       = "user.name:*"
  language    = "kuery"
  from        = "now-6m"
  to          = "now"
  interval    = "5m"
  index       = ["logs-*"]
  
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
