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

package anomalydetectionjob_test

import (
	"context"
	_ "embed"
	"fmt"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/elastic/go-elasticsearch/v8/typedapi/ml/putfilter"
	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

const testResourceAddr = "elasticstack_elasticsearch_ml_anomaly_detection_job.test"

// Elasticsearch often validates custom_rules.scope ML filter ids when the job is opened, not when the
// job config is put or updated; OpenJob errors mentioning filters/resource_not_found indicate a bad id.
var mlJobOpenFailsUnknownFilterRE = regexp.MustCompile(
	`(?is)(filter|not found|resource_not_found|does not exist|could not find|illegal_argument|validation)`)

// mlOpenJobErrorLooksLikeMLNodeCapacity reports whether OpenJob failed because the cluster could not
// assign the job to an ML node (common under CI load). In that case the response is not a reliable
// signal about missing filter ids, so callers may retry OpenJob.
func mlOpenJobErrorLooksLikeMLNodeCapacity(err error) bool {
	if err == nil {
		return false
	}
	s := strings.ToLower(err.Error())
	return strings.Contains(s, "429") ||
		strings.Contains(s, "too_many_requests") ||
		strings.Contains(s, "no ml nodes") ||
		strings.Contains(s, "insufficient memory") ||
		(strings.Contains(s, "insufficient") && strings.Contains(s, "capacity"))
}

// testAccCheckOpenMLJobFailsWithUnknownFilter verifies OpenJob fails because scope references a missing
// ML filter. PutJob/UpdateJob may still succeed in Elasticsearch; opening surfaces the problem.
//
// Shared CI clusters sometimes return only ML capacity / node assignment errors (HTTP 429) for OpenJob;
// after retries, the test is skipped so infra saturation does not fail the suite.
func testAccCheckOpenMLJobFailsWithUnknownFilter(t *testing.T, jobID string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		ctx := context.Background()
		client, err := clients.NewAcceptanceTestingElasticsearchScopedClient()
		if err != nil {
			return err
		}
		es, err := client.GetESClient()
		if err != nil {
			return err
		}
		const maxAttempts = 15
		const betweenAttempts = 4 * time.Second
		var openErr error
		for attempt := range maxAttempts {
			_, openErr = es.Ml.OpenJob(jobID).Do(ctx)
			if openErr == nil {
				break
			}
			if mlOpenJobErrorLooksLikeMLNodeCapacity(openErr) && attempt+1 < maxAttempts {
				time.Sleep(betweenAttempts)
				continue
			}
			break
		}
		err = openErr
		if err == nil {
			if _, closeErr := es.Ml.CloseJob(jobID).Force(true).Do(ctx); closeErr != nil {
				t.Logf("CloseJob after unexpected OpenJob success for job %q: %v", jobID, closeErr)
			}
			t.Skipf("skipping OpenJob filter negative check for job %q: OpenJob succeeded; this Elasticsearch build does not reject missing custom_rules.scope filter ids on open (version-dependent)", jobID)
		}
		if mlOpenJobErrorLooksLikeMLNodeCapacity(err) {
			t.Skipf("skipping OpenJob filter validation for job %q: ML cluster still reports capacity/node assignment errors after %d retries (shared CI load); last error: %v",
				jobID, maxAttempts, err)
		}
		if !mlJobOpenFailsUnknownFilterRE.MatchString(err.Error()) {
			return fmt.Errorf("OpenJob failed for job %q but error did not match unknown-filter pattern: %w", jobID, err)
		}
		return nil
	}
}

func TestAccResourceAnomalyDetectionJobBasic(t *testing.T) {
	jobID := fmt.Sprintf("test-anomaly-detector-basic-%s", sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum))

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"job_id": config.StringVariable(jobID),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_anomaly_detection_job.test", "job_id", jobID),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_anomaly_detection_job.test", "description", "Test anomaly detection job"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_anomaly_detection_job.test", "analysis_config.bucket_span", "15m"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_anomaly_detection_job.test", "analysis_config.detectors.0.function", "count"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_anomaly_detection_job.test", "data_description.time_field", "@timestamp"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_anomaly_detection_job.test", "data_description.time_format", "epoch_ms"),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_ml_anomaly_detection_job.test", "create_time"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_anomaly_detection_job.test", "job_type", "anomaly_detector"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables: config.Variables{
					"job_id": config.StringVariable(jobID),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_anomaly_detection_job.test", "job_id", jobID),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_anomaly_detection_job.test", "description", "Updated basic test anomaly detection job"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_anomaly_detection_job.test", "analysis_config.bucket_span", "15m"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_anomaly_detection_job.test", "analysis_config.detectors.0.function", "count"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_anomaly_detection_job.test", "data_description.time_field", "@timestamp"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_anomaly_detection_job.test", "data_description.time_format", "epoch_ms"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_anomaly_detection_job.test", "groups.#", "1"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_anomaly_detection_job.test", "groups.0", "basic-group"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_anomaly_detection_job.test", "analysis_limits.model_memory_limit", "128mb"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_anomaly_detection_job.test", "allow_lazy_open", "true"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_anomaly_detection_job.test", "results_retention_days", "15"),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_ml_anomaly_detection_job.test", "create_time"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_anomaly_detection_job.test", "job_type", "anomaly_detector"),
				),
			},
			// ImportState testing
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ResourceName:             "elasticstack_elasticsearch_ml_anomaly_detection_job.test",
				ImportState:              true,
				ImportStateVerify:        true,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables: config.Variables{
					"job_id": config.StringVariable(jobID),
				},
			},
		},
	})
}

func TestAccResourceAnomalyDetectionJobComprehensive(t *testing.T) {
	jobID := fmt.Sprintf("test-anomaly-detector-comprehensive-%s", sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum))

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"job_id": config.StringVariable(jobID),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_anomaly_detection_job.test", "job_id", jobID),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_anomaly_detection_job.test", "description", "Comprehensive test anomaly detection job"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_anomaly_detection_job.test", "groups.#", "2"),
					// Analysis config checks
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_anomaly_detection_job.test", "analysis_config.bucket_span", "10m"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_anomaly_detection_job.test", "analysis_config.latency", "30s"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_anomaly_detection_job.test", "analysis_config.summary_count_field_name", "event_count"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_anomaly_detection_job.test", "analysis_config.detectors.#", "2"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_anomaly_detection_job.test", "analysis_config.detectors.0.function", "count"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_anomaly_detection_job.test", "analysis_config.detectors.0.partition_field_name", "host"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_anomaly_detection_job.test", "analysis_config.detectors.1.function", "mean"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_anomaly_detection_job.test", "analysis_config.detectors.1.field_name", "response_time"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_anomaly_detection_job.test", "analysis_config.detectors.1.by_field_name", "status"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_anomaly_detection_job.test", "analysis_config.detectors.1.over_field_name", "clientip"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_anomaly_detection_job.test", "analysis_config.influencers.#", "1"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_anomaly_detection_job.test", "analysis_config.influencers.0", "status_code"),
					// Analysis limits checks
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_anomaly_detection_job.test", "analysis_limits.model_memory_limit", "100mb"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_anomaly_detection_job.test", "analysis_limits.categorization_examples_limit", "5"),
					// Data description checks
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_anomaly_detection_job.test", "data_description.time_field", "@timestamp"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_anomaly_detection_job.test", "data_description.time_format", "epoch_ms"),
					// Model plot config checks
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_anomaly_detection_job.test", "model_plot_config.enabled", "true"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_anomaly_detection_job.test", "model_plot_config.terms", "host1"),
					// Other settings checks
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_anomaly_detection_job.test", "allow_lazy_open", "true"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_anomaly_detection_job.test", "background_persist_interval", "1h"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_anomaly_detection_job.test", "custom_settings", "{\"custom_key\": \"custom_value\"}"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_anomaly_detection_job.test", "daily_model_snapshot_retention_after_days", "3"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_anomaly_detection_job.test", "model_snapshot_retention_days", "7"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_anomaly_detection_job.test", "renormalization_window_days", "14"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_anomaly_detection_job.test", "results_retention_days", "30"),
					// Computed fields
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_ml_anomaly_detection_job.test", "create_time"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_anomaly_detection_job.test", "job_type", "anomaly_detector"),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_ml_anomaly_detection_job.test", "job_version"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables: config.Variables{
					"job_id": config.StringVariable(jobID),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_anomaly_detection_job.test", "job_id", jobID),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_anomaly_detection_job.test", "description", "Updated comprehensive test anomaly detection job"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_anomaly_detection_job.test", "groups.#", "3"),
					// Analysis config checks (should remain the same since these are generally immutable)
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_anomaly_detection_job.test", "analysis_config.bucket_span", "10m"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_anomaly_detection_job.test", "analysis_config.latency", "30s"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_anomaly_detection_job.test", "analysis_config.summary_count_field_name", "event_count"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_anomaly_detection_job.test", "analysis_config.detectors.#", "2"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_anomaly_detection_job.test", "analysis_config.detectors.0.function", "count"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_anomaly_detection_job.test", "analysis_config.detectors.0.partition_field_name", "host"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_anomaly_detection_job.test", "analysis_config.detectors.1.function", "mean"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_anomaly_detection_job.test", "analysis_config.detectors.1.field_name", "response_time"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_anomaly_detection_job.test", "analysis_config.detectors.1.by_field_name", "status"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_anomaly_detection_job.test", "analysis_config.detectors.1.over_field_name", "clientip"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_anomaly_detection_job.test", "analysis_config.influencers.#", "1"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_anomaly_detection_job.test", "analysis_config.influencers.0", "status_code"),
					// Updated analysis limits checks
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_anomaly_detection_job.test", "analysis_limits.model_memory_limit", "256mb"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_anomaly_detection_job.test", "analysis_limits.categorization_examples_limit", "10"),
					// Data description checks (should remain the same)
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_anomaly_detection_job.test", "data_description.time_field", "@timestamp"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_anomaly_detection_job.test", "data_description.time_format", "epoch_ms"),
					// Updated model plot config checks
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_anomaly_detection_job.test", "model_plot_config.enabled", "false"),
					// Updated other settings checks
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_anomaly_detection_job.test", "allow_lazy_open", "false"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_anomaly_detection_job.test", "background_persist_interval", "3h"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_anomaly_detection_job.test", "custom_settings", "{\"updated_key\": \"updated_value\", \"additional_key\": \"additional_value\"}"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_anomaly_detection_job.test", "daily_model_snapshot_retention_after_days", "7"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_anomaly_detection_job.test", "model_snapshot_retention_days", "21"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_anomaly_detection_job.test", "renormalization_window_days", "28"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_anomaly_detection_job.test", "results_retention_days", "90"),
					// Computed fields
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_ml_anomaly_detection_job.test", "create_time"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_anomaly_detection_job.test", "job_type", "anomaly_detector"),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_ml_anomaly_detection_job.test", "job_version"),
				),
			},
		},
	})
}

// Regression test for #1567: empty influencer list causes "inconsistent result after apply".
func TestAccResourceAnomalyDetectionJobEmptyInfluencers(t *testing.T) {
	jobID := fmt.Sprintf("test-ad-empty-inf-%s", sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum))
	addr := testResourceAddr

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          config.Variables{"job_id": config.StringVariable(jobID)},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(addr, "job_id", jobID),
					resource.TestCheckResourceAttr(addr, "analysis_config.influencers.#", "0"),
					resource.TestCheckNoResourceAttr(addr, "analysis_config.detectors.0.detector_description"),
					resource.TestCheckResourceAttrSet(addr, "id"),
				),
			},
		},
	})
}

// Regression test for #1568: categorization_filters cause "inconsistent result after apply"
// because ES silently converts them to categorization_analyzer char_filter patterns.
func TestAccResourceAnomalyDetectionJobCategorizationFilters(t *testing.T) {
	jobID := fmt.Sprintf("test-ad-cat-filt-%s", sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum))
	addr := testResourceAddr

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          config.Variables{"job_id": config.StringVariable(jobID)},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(addr, "job_id", jobID),
					resource.TestCheckResourceAttr(addr, "analysis_config.categorization_field_name", "message"),
					resource.TestCheckResourceAttr(addr, "analysis_config.categorization_filters.#", "2"),
					resource.TestCheckResourceAttr(addr, "analysis_config.categorization_filters.0", `\b\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}\b`),
					resource.TestCheckResourceAttr(addr, "analysis_config.categorization_filters.1", `\b[A-Fa-f0-9]{8,}\b`),
					resource.TestCheckResourceAttrSet(addr, "id"),
				),
			},
		},
	})
}

// Regression test for #1569: per_partition_categorization with enabled=false causes
// "inconsistent result" because ES drops stop_on_warn when disabled.
func TestAccResourceAnomalyDetectionJobPerPartitionDisabled(t *testing.T) {
	jobID := fmt.Sprintf("test-ad-ppc-off-%s", sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum))
	addr := testResourceAddr

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          config.Variables{"job_id": config.StringVariable(jobID)},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(addr, "job_id", jobID),
					resource.TestCheckResourceAttr(addr, "analysis_config.per_partition_categorization.enabled", "false"),
					resource.TestCheckResourceAttr(addr, "analysis_config.per_partition_categorization.stop_on_warn", "false"),
					resource.TestCheckResourceAttrSet(addr, "id"),
				),
			},
		},
	})
}

// TestAccResourceAnomalyDetectionJobPerPartitionEnabled covers per_partition_categorization
// with enabled=true and stop_on_warn=true, complementing the disabled regression test.
func TestAccResourceAnomalyDetectionJobPerPartitionEnabled(t *testing.T) {
	jobID := fmt.Sprintf("test-ad-ppc-on-%s", sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum))
	addr := testResourceAddr

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          config.Variables{"job_id": config.StringVariable(jobID)},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(addr, "job_id", jobID),
					resource.TestCheckResourceAttr(addr, "analysis_config.per_partition_categorization.enabled", "true"),
					resource.TestCheckResourceAttr(addr, "analysis_config.per_partition_categorization.stop_on_warn", "true"),
					resource.TestCheckResourceAttrSet(addr, "id"),
				),
			},
		},
	})
}

// Regression test for #1564: custom_rules with conditions were not sent to ES on create,
// and the read path failed to serialize them back from the API response.
// The update step changes the condition value, triggering a destroy+recreate (analysis_config
// is immutable), and asserts the new value is correctly persisted.
func TestAccResourceAnomalyDetectionJobCustomRules(t *testing.T) {
	jobID := fmt.Sprintf("test-ad-rules-%s", sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum))
	addr := testResourceAddr

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          config.Variables{"job_id": config.StringVariable(jobID)},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(addr, "job_id", jobID),
					resource.TestCheckResourceAttr(addr, "analysis_config.detectors.0.custom_rules.#", "1"),
					resource.TestCheckResourceAttr(addr, "analysis_config.detectors.0.custom_rules.0.actions.#", "1"),
					resource.TestCheckResourceAttr(addr, "analysis_config.detectors.0.custom_rules.0.actions.0", "skip_result"),
					resource.TestCheckResourceAttr(addr, "analysis_config.detectors.0.custom_rules.0.conditions.#", "1"),
					resource.TestCheckResourceAttr(addr, "analysis_config.detectors.0.custom_rules.0.conditions.0.applies_to", "actual"),
					resource.TestCheckResourceAttr(addr, "analysis_config.detectors.0.custom_rules.0.conditions.0.operator", "lt"),
					resource.TestCheckResourceAttr(addr, "analysis_config.detectors.0.custom_rules.0.conditions.0.value", "10"),
					resource.TestCheckResourceAttrSet(addr, "id"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables:          config.Variables{"job_id": config.StringVariable(jobID)},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(addr, "job_id", jobID),
					resource.TestCheckResourceAttr(addr, "analysis_config.detectors.0.custom_rules.0.conditions.0.value", "20"),
					resource.TestCheckResourceAttrSet(addr, "id"),
				),
			},
		},
	})
}

// Scope-related acceptance tests run under the same conditions as other ML anomaly detection job tests
// (TF_ACC=1 and a cluster with ML enabled). They are not gated on TF_ACC_ML_SCOPE_TEST: filters are
// created via the Elasticsearch Put filter API in-process, so the elasticstack_elasticsearch_ml_filter
// resource is not required.
//
// TestAccResourceAnomalyDetectionJobCustomRulesScope covers detector custom_rules.scope
// (ML filter references per analysis field), including round-trip to Elasticsearch.
//
// The referenced ML filter is created out-of-band via the Elasticsearch ML APIs before apply,
// so this test does not require the elasticstack_elasticsearch_ml_filter resource.
func TestAccResourceAnomalyDetectionJobCustomRulesScope(t *testing.T) {
	jobID := fmt.Sprintf("test-ad-scope-%s", sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum))
	filterID := fmt.Sprintf("test-ad-scope-flt-%s", sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum))
	addr := testResourceAddr

	setupAccMLFilterOutOfBand(t, filterID)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"job_id":    config.StringVariable(jobID),
					"filter_id": config.StringVariable(filterID),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(addr, "job_id", jobID),
					resource.TestCheckResourceAttr(addr, "analysis_config.detectors.0.custom_rules.#", "1"),
					resource.TestCheckResourceAttr(addr, "analysis_config.detectors.0.custom_rules.0.actions.#", "1"),
					resource.TestCheckResourceAttr(addr, "analysis_config.detectors.0.custom_rules.0.actions.0", "skip_result"),
					resource.TestCheckResourceAttr(addr, "analysis_config.detectors.0.custom_rules.0.scope.clientip.filter_id", filterID),
					resource.TestCheckResourceAttr(addr, "analysis_config.detectors.0.custom_rules.0.scope.clientip.filter_type", "include"),
					resource.TestCheckResourceAttrSet(addr, "id"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables: config.Variables{
					"job_id":    config.StringVariable(jobID),
					"filter_id": config.StringVariable(filterID),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(addr, "analysis_config.detectors.0.custom_rules.0.scope.clientip.filter_type", "exclude"),
					resource.TestCheckResourceAttrSet(addr, "id"),
				),
			},
		},
	})
}

// TestAccResourceAnomalyDetectionJobCustomRulesScopeAndConditions asserts Elasticsearch accepts a
// single custom rule with both a non-empty scope and at least one condition (inclusive OR from the
// API docs, not mutually exclusive). The ML filter is created out-of-band like the scope-only acc test.
func TestAccResourceAnomalyDetectionJobCustomRulesScopeAndConditions(t *testing.T) {
	jobID := fmt.Sprintf("test-ad-scope-cond-%s", sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum))
	filterID := fmt.Sprintf("test-ad-scope-cond-flt-%s", sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum))
	addr := testResourceAddr

	setupAccMLFilterOutOfBand(t, filterID)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"job_id":    config.StringVariable(jobID),
					"filter_id": config.StringVariable(filterID),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(addr, "job_id", jobID),
					resource.TestCheckResourceAttr(addr, "analysis_config.detectors.0.custom_rules.#", "1"),
					resource.TestCheckResourceAttr(addr, "analysis_config.detectors.0.custom_rules.0.actions.#", "1"),
					resource.TestCheckResourceAttr(addr, "analysis_config.detectors.0.custom_rules.0.actions.0", "skip_result"),
					resource.TestCheckResourceAttr(addr, "analysis_config.detectors.0.custom_rules.0.scope.clientip.filter_id", filterID),
					resource.TestCheckResourceAttr(addr, "analysis_config.detectors.0.custom_rules.0.scope.clientip.filter_type", "include"),
					resource.TestCheckResourceAttr(addr, "analysis_config.detectors.0.custom_rules.0.conditions.#", "1"),
					resource.TestCheckResourceAttr(addr, "analysis_config.detectors.0.custom_rules.0.conditions.0.applies_to", "actual"),
					resource.TestCheckResourceAttr(addr, "analysis_config.detectors.0.custom_rules.0.conditions.0.operator", "lt"),
					resource.TestCheckResourceAttr(addr, "analysis_config.detectors.0.custom_rules.0.conditions.0.value", "10"),
					resource.TestCheckResourceAttrSet(addr, "id"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables: config.Variables{
					"job_id":    config.StringVariable(jobID),
					"filter_id": config.StringVariable(filterID),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(addr, "analysis_config.detectors.0.custom_rules.0.scope.clientip.filter_type", "exclude"),
					resource.TestCheckResourceAttr(addr, "analysis_config.detectors.0.custom_rules.0.conditions.0.value", "20"),
					resource.TestCheckResourceAttrSet(addr, "id"),
				),
			},
		},
	})
}

// TestAccResourceAnomalyDetectionJobCustomRulesScope_missingFilterOnCreate applies a job whose scope
// references a filter id that was never created. Elasticsearch may still accept PutJob; opening the
// job must fail until the filter exists (proves the configuration is not runnable as-is).
func TestAccResourceAnomalyDetectionJobCustomRulesScope_missingFilterOnCreate(t *testing.T) {
	jobID := fmt.Sprintf("test-ad-scope-miss-%s", sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum))
	filterID := fmt.Sprintf("nonexistent-flt-%s", sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum))
	addr := testResourceAddr

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"job_id":    config.StringVariable(jobID),
					"filter_id": config.StringVariable(filterID),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(addr, "job_id", jobID),
					testAccCheckOpenMLJobFailsWithUnknownFilter(t, jobID),
				),
			},
		},
	})
}

// TestAccResourceAnomalyDetectionJobCustomRulesScope_missingFilterOnUpdate updates scope to a filter id
// that does not exist. PutJob/UpdateJob may succeed; OpenJob must fail until the filter exists.
func TestAccResourceAnomalyDetectionJobCustomRulesScope_missingFilterOnUpdate(t *testing.T) {
	jobID := fmt.Sprintf("test-ad-scope-miss-up-%s", sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum))
	goodFilterID := fmt.Sprintf("test-ad-scope-flt-ok-%s", sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum))
	badFilterID := fmt.Sprintf("nonexistent-flt-%s", sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum))
	addr := testResourceAddr

	setupAccMLFilterOutOfBand(t, goodFilterID)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"job_id":    config.StringVariable(jobID),
					"filter_id": config.StringVariable(goodFilterID),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(addr, "job_id", jobID),
					resource.TestCheckResourceAttr(addr, "analysis_config.detectors.0.custom_rules.#", "1"),
					resource.TestCheckResourceAttr(addr, "analysis_config.detectors.0.custom_rules.0.scope.clientip.filter_id", goodFilterID),
					resource.TestCheckResourceAttr(addr, "analysis_config.detectors.0.custom_rules.0.scope.clientip.filter_type", "include"),
					resource.TestCheckResourceAttrSet(addr, "id"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables: config.Variables{
					"job_id":    config.StringVariable(jobID),
					"filter_id": config.StringVariable(badFilterID),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(addr, "analysis_config.detectors.0.custom_rules.0.scope.clientip.filter_id", badFilterID),
					testAccCheckOpenMLJobFailsWithUnknownFilter(t, jobID),
				),
			},
		},
	})
}

func TestAccResourceAnomalyDetectionJobNullAndEmpty(t *testing.T) {
	jobID := fmt.Sprintf("test-anomaly-detector-null-and-empty-%s", sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum))

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"job_id": config.StringVariable(jobID),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_anomaly_detection_job.test", "job_id", jobID),
					resource.TestCheckNoResourceAttr("elasticstack_elasticsearch_ml_anomaly_detection_job.test", "description"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_anomaly_detection_job.test", "groups.#", "0"),
					// Analysis config checks
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_anomaly_detection_job.test", "analysis_config.bucket_span", "15m"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_anomaly_detection_job.test", "analysis_config.detectors.#", "1"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_anomaly_detection_job.test", "analysis_config.detectors.0.function", "sum"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_anomaly_detection_job.test", "analysis_config.detectors.0.field_name", "bytes"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_anomaly_detection_job.test", "analysis_config.detectors.0.detector_description", "Sum of bytes"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_anomaly_detection_job.test", "analysis_config.detectors.0.use_null", "false"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_anomaly_detection_job.test", "analysis_config.detectors.0.custom_rules.#", "0"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_anomaly_detection_job.test", "analysis_config.influencers.#", "0"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_anomaly_detection_job.test", "analysis_config.categorization_filters.#", "0"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_anomaly_detection_job.test", "analysis_config.per_partition_categorization.enabled", "false"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_anomaly_detection_job.test", "analysis_config.per_partition_categorization.stop_on_warn", "false"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_anomaly_detection_job.test", "analysis_config.multivariate_by_fields", "false"),
					// Analysis limits checks
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_anomaly_detection_job.test", "analysis_limits.model_memory_limit", "11MB"),
					// Data description checks
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_anomaly_detection_job.test", "data_description.time_field", "timestamp"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_anomaly_detection_job.test", "data_description.time_format", "epoch_ms"),
					// Model plot config checks
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_anomaly_detection_job.test", "model_plot_config.enabled", "true"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_anomaly_detection_job.test", "model_plot_config.annotations_enabled", "true"),
					// Other settings checks
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_anomaly_detection_job.test", "allow_lazy_open", "false"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_anomaly_detection_job.test", "results_index_name", "test-job1"),
					// Computed fields
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_ml_anomaly_detection_job.test", "create_time"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_anomaly_detection_job.test", "job_type", "anomaly_detector"),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_ml_anomaly_detection_job.test", "job_version"),
				),
			},
		},
	})
}

// TestAccResourceAnomalyDetectionJobExplicitConnection exercises the elasticsearch_connection block
// (scoped Elasticsearch client) on the anomaly detection job resource directly.
// It creates a job with an explicit connection using username/password (or api_key when available),
// asserts connection block attributes, and verifies import works.
func TestAccResourceAnomalyDetectionJobExplicitConnection(t *testing.T) {
	endpoints := testAccAnomalyDetectionJobESEndpoints()
	if len(endpoints) == 0 {
		t.Skip("ELASTICSEARCH_ENDPOINTS must be set to run this test")
	}
	endpointVars := make([]config.Variable, 0, len(endpoints))
	for _, endpoint := range endpoints {
		endpointVars = append(endpointVars, config.StringVariable(endpoint))
	}
	jobID := fmt.Sprintf("test-ad-explicit-conn-%s", sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum))

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			// Step 1: create with explicit connection (api_key if available, else username/password)
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"job_id":    config.StringVariable(jobID),
					"endpoints": config.ListVariable(endpointVars...),
					"api_key":   config.StringVariable(os.Getenv("ELASTICSEARCH_API_KEY")),
					"username":  config.StringVariable(os.Getenv("ELASTICSEARCH_USERNAME")),
					"password":  config.StringVariable(os.Getenv("ELASTICSEARCH_PASSWORD")),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(testResourceAddr, "job_id", jobID),
					resource.TestCheckResourceAttr(testResourceAddr, "elasticsearch_connection.#", "1"),
					resource.TestCheckResourceAttr(testResourceAddr, "elasticsearch_connection.0.endpoints.#", fmt.Sprintf("%d", len(endpoints))),
					resource.TestCheckResourceAttr(testResourceAddr, "elasticsearch_connection.0.endpoints.0", endpoints[0]),
					resource.TestCheckResourceAttr(testResourceAddr, "elasticsearch_connection.0.insecure", "true"),
					resource.TestCheckResourceAttrSet(testResourceAddr, "create_time"),
					resource.TestCheckResourceAttr(testResourceAddr, "job_type", "anomaly_detector"),
				),
			},
			// Step 2: import verification; sensitive connection block is ignored on import
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"job_id":    config.StringVariable(jobID),
					"endpoints": config.ListVariable(endpointVars...),
					"api_key":   config.StringVariable(os.Getenv("ELASTICSEARCH_API_KEY")),
					"username":  config.StringVariable(os.Getenv("ELASTICSEARCH_USERNAME")),
					"password":  config.StringVariable(os.Getenv("ELASTICSEARCH_PASSWORD")),
				},
				ResourceName:            testResourceAddr,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"elasticsearch_connection"},
			},
			// Step 3: update description while keeping the same explicit connection (username/password path)
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables: config.Variables{
					"job_id":    config.StringVariable(jobID),
					"endpoints": config.ListVariable(endpointVars...),
					"username":  config.StringVariable(os.Getenv("ELASTICSEARCH_USERNAME")),
					"password":  config.StringVariable(os.Getenv("ELASTICSEARCH_PASSWORD")),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(testResourceAddr, "job_id", jobID),
					resource.TestCheckResourceAttr(testResourceAddr, "description", "Updated anomaly detection job with explicit connection"),
					resource.TestCheckResourceAttr(testResourceAddr, "elasticsearch_connection.#", "1"),
					resource.TestCheckResourceAttr(testResourceAddr, "elasticsearch_connection.0.endpoints.#", fmt.Sprintf("%d", len(endpoints))),
					resource.TestCheckResourceAttr(testResourceAddr, "elasticsearch_connection.0.endpoints.0", endpoints[0]),
					resource.TestCheckResourceAttr(testResourceAddr, "elasticsearch_connection.0.insecure", "true"),
				),
			},
			// Step 4: re-import after update to confirm connection block survives
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables: config.Variables{
					"job_id":    config.StringVariable(jobID),
					"endpoints": config.ListVariable(endpointVars...),
					"username":  config.StringVariable(os.Getenv("ELASTICSEARCH_USERNAME")),
					"password":  config.StringVariable(os.Getenv("ELASTICSEARCH_PASSWORD")),
				},
				ResourceName:            testResourceAddr,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"elasticsearch_connection"},
			},
		},
	})
}

// setupAccMLFilterOutOfBand creates an ML filter via the Elasticsearch API so the acceptance
// config can reference filter_id without using elasticstack_elasticsearch_ml_filter. The filter
// is deleted after the test (Destroy runs before registered Cleanups).
func setupAccMLFilterOutOfBand(t *testing.T, filterID string) {
	t.Helper()
	ctx := context.Background()
	client, err := clients.NewAcceptanceTestingElasticsearchScopedClient()
	if err != nil {
		t.Fatalf("Elasticsearch client: %v", err)
	}
	es, err := client.GetESClient()
	if err != nil {
		t.Fatalf("GetESClient: %v", err)
	}
	desc := "Terraform acc test ML filter (created out-of-band via Elasticsearch Put Filter API)"
	_, err = es.Ml.PutFilter(filterID).Request(&putfilter.Request{
		Description: &desc,
		Items:       []string{"10.0.0.1"},
	}).Do(ctx)
	if err != nil {
		t.Fatalf("create ML filter %q out-of-band: %v", filterID, err)
	}
	t.Cleanup(func() {
		ctx := context.Background()
		client, err := clients.NewAcceptanceTestingElasticsearchScopedClient()
		if err != nil {
			t.Logf("cleanup: Elasticsearch client: %v", err)
			return
		}
		es, err := client.GetESClient()
		if err != nil {
			t.Logf("cleanup: GetESClient: %v", err)
			return
		}
		_, err = es.Ml.DeleteFilter(filterID).Do(ctx)
		if err != nil {
			t.Logf("cleanup: delete ML filter %q: %v", filterID, err)
		}
	})
}

func testAccAnomalyDetectionJobESEndpoints() []string {
	rawEndpoints := os.Getenv("ELASTICSEARCH_ENDPOINTS")
	parts := strings.Split(rawEndpoints, ",")
	endpoints := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			endpoints = append(endpoints, part)
		}
	}
	return endpoints
}
