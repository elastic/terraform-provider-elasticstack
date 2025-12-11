package anomaly_detection_job_test

import (
	_ "embed"
	"fmt"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

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
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_anomaly_detection_job.test", "analysis_config.detectors.#", "2"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_anomaly_detection_job.test", "analysis_config.detectors.0.function", "count"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_anomaly_detection_job.test", "analysis_config.detectors.1.function", "mean"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_anomaly_detection_job.test", "analysis_config.detectors.1.field_name", "response_time"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_anomaly_detection_job.test", "analysis_config.influencers.#", "1"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_anomaly_detection_job.test", "analysis_config.influencers.0", "status_code"),
					// Analysis limits checks
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_anomaly_detection_job.test", "analysis_limits.model_memory_limit", "100mb"),
					// Data description checks
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_anomaly_detection_job.test", "data_description.time_field", "@timestamp"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_anomaly_detection_job.test", "data_description.time_format", "epoch_ms"),
					// Model plot config checks
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_anomaly_detection_job.test", "model_plot_config.enabled", "true"),
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
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_anomaly_detection_job.test", "analysis_config.detectors.#", "2"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_anomaly_detection_job.test", "analysis_config.detectors.0.function", "count"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_anomaly_detection_job.test", "analysis_config.detectors.1.function", "mean"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_anomaly_detection_job.test", "analysis_config.detectors.1.field_name", "response_time"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_anomaly_detection_job.test", "analysis_config.influencers.#", "1"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_anomaly_detection_job.test", "analysis_config.influencers.0", "status_code"),
					// Updated analysis limits checks
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_ml_anomaly_detection_job.test", "analysis_limits.model_memory_limit", "256mb"),
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
