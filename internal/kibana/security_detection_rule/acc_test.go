package security_detection_rule_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccResourceSecurityDetectionRule_Query(t *testing.T) {
	resourceName := "elasticstack_kibana_security_detection_rule.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.Providers,
		CheckDestroy:             testAccCheckSecurityDetectionRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSecurityDetectionRuleConfig_query("test-query-rule"),
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
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "rule_id"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "created_by"),
				),
			},
			{
				Config: testAccSecurityDetectionRuleConfig_queryUpdate("test-query-rule-updated"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", "test-query-rule-updated"),
					resource.TestCheckResourceAttr(resourceName, "description", "Updated test query security detection rule"),
					resource.TestCheckResourceAttr(resourceName, "severity", "high"),
					resource.TestCheckResourceAttr(resourceName, "risk_score", "75"),
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
				Config: testAccSecurityDetectionRuleConfig_eql("test-eql-rule"),
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
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "rule_id"),
				),
			},
			{
				Config: testAccSecurityDetectionRuleConfig_eqlUpdate("test-eql-rule-updated"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", "test-eql-rule-updated"),
					resource.TestCheckResourceAttr(resourceName, "query", "process where process.name == \"powershell.exe\""),
					resource.TestCheckResourceAttr(resourceName, "description", "Updated test EQL security detection rule"),
					resource.TestCheckResourceAttr(resourceName, "severity", "critical"),
					resource.TestCheckResourceAttr(resourceName, "risk_score", "90"),
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
				Config: testAccSecurityDetectionRuleConfig_esql("test-esql-rule"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", "test-esql-rule"),
					resource.TestCheckResourceAttr(resourceName, "type", "esql"),
					resource.TestCheckResourceAttr(resourceName, "query", "FROM logs-* | WHERE event.action == \"login\" | STATS count(*) BY user.name"),
					resource.TestCheckResourceAttr(resourceName, "language", "esql"),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "description", "Test ESQL security detection rule"),
					resource.TestCheckResourceAttr(resourceName, "severity", "medium"),
					resource.TestCheckResourceAttr(resourceName, "risk_score", "60"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "rule_id"),
				),
			},
			{
				Config: testAccSecurityDetectionRuleConfig_esqlUpdate("test-esql-rule-updated"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", "test-esql-rule-updated"),
					resource.TestCheckResourceAttr(resourceName, "query", "FROM logs-* | WHERE event.action == \"logout\" | STATS count(*) BY user.name, source.ip"),
					resource.TestCheckResourceAttr(resourceName, "description", "Updated test ESQL security detection rule"),
					resource.TestCheckResourceAttr(resourceName, "severity", "high"),
					resource.TestCheckResourceAttr(resourceName, "risk_score", "80"),
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
				Config: testAccSecurityDetectionRuleConfig_machineLearning("test-ml-rule"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", "test-ml-rule"),
					resource.TestCheckResourceAttr(resourceName, "type", "machine_learning"),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "description", "Test ML security detection rule"),
					resource.TestCheckResourceAttr(resourceName, "severity", "critical"),
					resource.TestCheckResourceAttr(resourceName, "risk_score", "90"),
					resource.TestCheckResourceAttr(resourceName, "anomaly_threshold", "75"),
					resource.TestCheckResourceAttr(resourceName, "machine_learning_job_id.0", "test-ml-job"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "rule_id"),
				),
			},
			{
				Config: testAccSecurityDetectionRuleConfig_machineLearningUpdate("test-ml-rule-updated"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", "test-ml-rule-updated"),
					resource.TestCheckResourceAttr(resourceName, "description", "Updated test ML security detection rule"),
					resource.TestCheckResourceAttr(resourceName, "severity", "high"),
					resource.TestCheckResourceAttr(resourceName, "risk_score", "85"),
					resource.TestCheckResourceAttr(resourceName, "anomaly_threshold", "80"),
					resource.TestCheckResourceAttr(resourceName, "machine_learning_job_id.0", "test-ml-job"),
					resource.TestCheckResourceAttr(resourceName, "machine_learning_job_id.1", "test-ml-job-2"),
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
				Config: testAccSecurityDetectionRuleConfig_newTerms("test-new-terms-rule"),
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
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "rule_id"),
				),
			},
			{
				Config: testAccSecurityDetectionRuleConfig_newTermsUpdate("test-new-terms-rule-updated"),
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
				Config: testAccSecurityDetectionRuleConfig_savedQuery("test-saved-query-rule"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", "test-saved-query-rule"),
					resource.TestCheckResourceAttr(resourceName, "type", "saved_query"),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "description", "Test saved query security detection rule"),
					resource.TestCheckResourceAttr(resourceName, "severity", "low"),
					resource.TestCheckResourceAttr(resourceName, "risk_score", "30"),
					resource.TestCheckResourceAttr(resourceName, "saved_id", "test-saved-query-id"),
					resource.TestCheckResourceAttr(resourceName, "index.0", "logs-*"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "rule_id"),
				),
			},
			{
				Config: testAccSecurityDetectionRuleConfig_savedQueryUpdate("test-saved-query-rule-updated"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", "test-saved-query-rule-updated"),
					resource.TestCheckResourceAttr(resourceName, "query", "event.action:*"),
					resource.TestCheckResourceAttr(resourceName, "description", "Updated test saved query security detection rule"),
					resource.TestCheckResourceAttr(resourceName, "severity", "medium"),
					resource.TestCheckResourceAttr(resourceName, "risk_score", "60"),
					resource.TestCheckResourceAttr(resourceName, "saved_id", "test-saved-query-id-updated"),
					resource.TestCheckResourceAttr(resourceName, "index.0", "logs-*"),
					resource.TestCheckResourceAttr(resourceName, "index.1", "audit-*"),
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
				Config: testAccSecurityDetectionRuleConfig_threatMatch("test-threat-match-rule"),
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
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "rule_id"),
				),
			},
			{
				Config: testAccSecurityDetectionRuleConfig_threatMatchUpdate("test-threat-match-rule-updated"),
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
				Config: testAccSecurityDetectionRuleConfig_threshold("test-threshold-rule"),
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
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "rule_id"),
				),
			},
			{
				Config: testAccSecurityDetectionRuleConfig_thresholdUpdate("test-threshold-rule-updated"),
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
		if rs.Type != "elasticstack_kibana_security_detection_rule" {
			continue
		}

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
}
`, name)
}
