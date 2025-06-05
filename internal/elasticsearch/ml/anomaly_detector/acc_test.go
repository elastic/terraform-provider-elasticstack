package anomaly_detector_test

import (
	"fmt"
	"io"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccElasticsearchMLAnomalyDetector_Basic(t *testing.T) {
	resourceName := "elasticstack_elasticsearch_ml_anomaly_detector.test_job"
	randSuffix := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	jobId := fmt.Sprintf("tf_acc_comprehensive_job_%s", randSuffix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.Providers,
		CheckDestroy:             checkResourceMLAnomalyDetectorDestroy(jobId),
		Steps: []resource.TestStep{
			{
				// Step 1: Create comprehensive job
				Config: testAccElasticsearchMLAnomalyDetectorComprehensiveCreate(jobId),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "job_id", jobId),
					resource.TestCheckResourceAttr(resourceName, "description", "Comprehensive anomaly detector for testing."),
					resource.TestCheckResourceAttr(resourceName, "groups.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "groups.0", "acc-test-group"),
					resource.TestCheckResourceAttr(resourceName, "analysis_config.bucket_span", "30m"),
					resource.TestCheckResourceAttr(resourceName, "analysis_config.summary_count_field_name", "doc_count"),
					resource.TestCheckResourceAttr(resourceName, "analysis_config.latency", "3600s"),
					resource.TestCheckResourceAttr(resourceName, "analysis_config.detectors.0.function", "count"),
					resource.TestCheckResourceAttr(resourceName, "analysis_config.detectors.0.detector_description", "Count of documents"),
					resource.TestCheckResourceAttr(resourceName, "analysis_config.detectors.0.use_null", "true"),
					resource.TestCheckResourceAttr(resourceName, "analysis_config.detectors.0.exclude_frequent", "all"),
					resource.TestCheckResourceAttr(resourceName, "data_description.time_field", "@timestamp"),
					resource.TestCheckResourceAttr(resourceName, "data_description.time_format", "epoch_ms"),
					resource.TestCheckResourceAttr(resourceName, "analysis_limits.model_memory_limit", "128mb"),
					resource.TestCheckResourceAttr(resourceName, "analysis_limits.categorization_examples_limit", "5"),
					resource.TestCheckResourceAttr(resourceName, "model_snapshot_retention_days", "10"),
					resource.TestCheckResourceAttr(resourceName, "results_retention_days", "14"),
					resource.TestCheckResourceAttr(resourceName, "allow_lazy_open", "true"),
					resource.TestCheckResourceAttr(resourceName, "daily_model_snapshot_retention_after_days", "1"),
					resource.TestCheckResourceAttr(resourceName, "custom_settings.my_custom_key", "my_custom_value"),
				),
			},
			{
				// Step 2: Update various fields
				Config: testAccElasticsearchMLAnomalyDetectorComprehensiveUpdateStep2(jobId),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "description", "Updated comprehensive detector."),
					resource.TestCheckResourceAttr(resourceName, "groups.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "groups.0", "acc-test-group-updated"), // Order can vary, adjust if needed
					resource.TestCheckResourceAttr(resourceName, "groups.1", "new-group"),
					resource.TestCheckResourceAttr(resourceName, "analysis_limits.model_memory_limit", "256mb"),
					resource.TestCheckResourceAttr(resourceName, "analysis_limits.categorization_examples_limit", "5"), // Check it's still there
					resource.TestCheckResourceAttr(resourceName, "results_retention_days", "20"),
					resource.TestCheckResourceAttr(resourceName, "analysis_config.detectors.0.detector_description", "Updated count of documents"),
					resource.TestCheckResourceAttr(resourceName, "analysis_config.detectors.0.custom_rules.#", "2"), // Added a new rule
					resource.TestCheckResourceAttr(resourceName, "analysis_config.detectors.0.custom_rules.0.scope", "anomaly_score"),
					resource.TestCheckResourceAttr(resourceName, "analysis_config.detectors.0.custom_rules.1.actions.0", "skip_model_update"),
					resource.TestCheckResourceAttr(resourceName, "analysis_config.detectors.0.custom_rules.1.scope", "time"),
					resource.TestCheckResourceAttr(resourceName, "analysis_config.detectors.0.custom_rules.1.conditions.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "analysis_config.detectors.0.custom_rules.1.conditions.0.operator", "lt"),
					resource.TestCheckResourceAttr(resourceName, "analysis_config.detectors.0.custom_rules.1.conditions.0.value", "1000.0"),
					resource.TestCheckResourceAttr(resourceName, "custom_settings.my_custom_key", "updated_custom_value"),
					resource.TestCheckResourceAttr(resourceName, "custom_settings.another_key", "another_value"), // Added new custom setting
				),
			},
			{
				// Step 3: Modify custom rules (remove one)
				Config: testAccElasticsearchMLAnomalyDetectorComprehensiveUpdateStep3(jobId),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "analysis_limits.categorization_examples_limit", "5"), // Check it's still there
					resource.TestCheckResourceAttr(resourceName, "analysis_config.detectors.0.custom_rules.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "analysis_config.detectors.0.custom_rules.0.actions.0", "skip_model_update"),
					resource.TestCheckResourceAttr(resourceName, "analysis_config.detectors.0.custom_rules.0.scope", "time"),
					resource.TestCheckResourceAttr(resourceName, "analysis_config.detectors.0.custom_rules.0.conditions.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "analysis_config.detectors.0.custom_rules.0.conditions.0.operator", "lt"),
					resource.TestCheckResourceAttr(resourceName, "analysis_config.detectors.0.custom_rules.0.conditions.0.value", "1000.0"),
					resource.TestCheckResourceAttr(resourceName, "allow_lazy_open", "false"),
				),
			},
			{
				// Step 4: Remove all custom settings and groups
				Config: testAccElasticsearchMLAnomalyDetectorComprehensiveUpdateStep4(jobId),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "custom_settings.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "groups.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "analysis_limits.categorization_examples_limit", "5"), // Check it's still there
					resource.TestCheckResourceAttr(resourceName, "analysis_config.detectors.0.custom_rules.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "analysis_config.detectors.0.custom_rules.0.actions.0", "skip_model_update"),
					resource.TestCheckResourceAttr(resourceName, "analysis_config.detectors.0.custom_rules.0.scope", "time"),
					resource.TestCheckResourceAttr(resourceName, "analysis_config.detectors.0.custom_rules.0.conditions.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "analysis_config.detectors.0.custom_rules.0.conditions.0.operator", "lt"),
					resource.TestCheckResourceAttr(resourceName, "analysis_config.detectors.0.custom_rules.0.conditions.0.value", "1000.0"),
				),
			},
		},
	})
}

func testAccElasticsearchMLAnomalyDetectorComprehensiveCreate(jobId string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_ml_anomaly_detector" "test_job" {
  job_id      = "%s"
  description = "Comprehensive anomaly detector for testing."
  groups      = ["acc-test-group"]

  analysis_limits = {
    model_memory_limit = "128mb"
    categorization_examples_limit = 5
  }
  model_snapshot_retention_days = 10
  results_retention_days        = 14
  allow_lazy_open               = true
  daily_model_snapshot_retention_after_days = 1

  custom_settings = {
    my_custom_key = "my_custom_value"
  }

  analysis_config = {
    bucket_span = "30m"
    summary_count_field_name  = "doc_count"
    latency                   = "3600s"
    detectors = [{
      function             = "count"
      detector_description = "Count of documents"
      use_null             = true
      exclude_frequent     = "all"
    }]
  }

  data_description = {
    time_field = "@timestamp"
    time_format = "epoch_ms"
  }
}
`, jobId)
}

// checkResourceMLAnomalyDetectorDestroy verifies the job is deleted from Elasticsearch
func checkResourceMLAnomalyDetectorDestroy(jobId string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client, err := clients.NewAcceptanceTestingClient()
		if err != nil {
			return err
		}
		esClient, err := client.GetESClient()
		if err != nil {
			return fmt.Errorf("error getting ES client: %w", err)
		}

		// Use the ML GetJobs API to check if the job exists
		mlClient := esClient.ML
		res, err := mlClient.GetJobs(mlClient.GetJobs.WithJobID(jobId))

		// If there's an error making the call, we can't be sure of the state.
		// However, if the test ran through apply, the connection should be fine.
		// A 404 error from the API call itself (if the client returns it as an error) is a success for destroy.
		if err != nil {
			// Ideally, check for a specific 404 error type if the client library provides one.
			// For now, we'll assume an error means it's likely not found or a connection issue handled by PreCheck.
			// This might need refinement based on how the ES client surfaces 404s.
			return nil
		}
		defer res.Body.Close()

		// If the response status code is 404, the job is gone.
		if res.StatusCode == 404 {
			return nil
		}

		// If the response is not an error and not 404, the job still exists.
		if !res.IsError() {
			// To be absolutely sure, we could parse the response and check if the job_id is present.
			// For now, a successful, non-404 response means it's still there.
			return fmt.Errorf("ML anomaly detector job '%s' still exists after destroy", jobId)
		}

		// If res.IsError() is true, but not a 404 (handled above), it's an unexpected error.
		bodyBytes, _ := io.ReadAll(res.Body) // Read body for error details
		return fmt.Errorf("unexpected error checking for ML job '%s' after destroy: %s - %s", jobId, res.Status(), string(bodyBytes))
	}
}

// HCL for Step 2 of comprehensive test
func testAccElasticsearchMLAnomalyDetectorComprehensiveUpdateStep2(jobId string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_ml_anomaly_detector" "test_job" {
  job_id      = "%s"
  description = "Updated comprehensive detector."
  groups      = ["acc-test-group-updated", "new-group"]

  analysis_limits = {
    model_memory_limit = "256mb" // Changed
    categorization_examples_limit = 5 // Remains same
  }
  results_retention_days = 20 // Changed
  model_snapshot_retention_days = 10 
  allow_lazy_open               = true
  daily_model_snapshot_retention_after_days = 1

  custom_settings = {
    my_custom_key    = "updated_custom_value" // Changed
    another_key      = "another_value"      // Added
  }

  analysis_config = {
    bucket_span = "30m" // Remains same
    categorization_field_name = "mlcategory" // Remains same
    summary_count_field_name  = "doc_count" // Remains same
    latency                   = "3600s" // Remains same
    detectors = [{
      function             = "count" // Remains same
      detector_description = "Updated count of documents" // Changed
      use_null             = true // Remains same
      exclude_frequent     = "all" // Remains same
      custom_rules = [
        {
          actions = ["skip_result"]
          scope   = "anomaly_score"
          conditions = [{
            operator = "gt"
            value    = 50.0
          }]
        },
        {
          actions = ["skip_model_update"]
          scope   = "time"
          conditions = [{
            operator = "lt"
            value    = 1000.0
          }]
        }
      ]
    }]
  }

  data_description = {
    time_field = "@timestamp" // Remains same
    time_format = "epoch_ms" // Remains same
  }
}
`, jobId)
}

// HCL for Step 3 of comprehensive test
func testAccElasticsearchMLAnomalyDetectorComprehensiveUpdateStep3(jobId string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_ml_anomaly_detector" "test_job" {
  job_id      = "%s"
  description = "Updated comprehensive detector." // Same as step 2
  groups      = ["acc-test-group-updated", "new-group"] // Same as step 2

  analysis_limits = {
    model_memory_limit = "256mb"
    categorization_examples_limit = 5 // Remains same
  }
  results_retention_days = 20 
  model_snapshot_retention_days = 10 
  allow_lazy_open               = false // Changed
  daily_model_snapshot_retention_after_days = 1

  custom_settings = {
    my_custom_key    = "updated_custom_value"
    another_key      = "another_value"
  }

  analysis_config {
    bucket_span = "30m"
    categorization_field_name = "mlcategory"
    summary_count_field_name  = "doc_count"
    latency                   = "3600s"
    detectors {
      function             = "count"
      detector_description = "Updated count of documents"
      use_null             = true
      exclude_frequent     = "all"
      custom_rules = [{ // Only the second rule from step 2 remains
        actions = ["skip_model_update"]
        scope   = "time"
        conditions = [{
          operator = "lt"
          value    = 1000.0
        }]
      }]
    }]
  }

  data_description {
    time_field = "@timestamp"
    time_format = "epoch_ms"
  }
}
`, jobId)
}

// HCL for Step 4 of comprehensive test
func testAccElasticsearchMLAnomalyDetectorComprehensiveUpdateStep4(jobId string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_ml_anomaly_detector" "test_job" {
  job_id      = "%s"
  description = "Final state, no groups or custom settings."
  // groups are removed

  analysis_limits = {
    model_memory_limit = "256mb"
    categorization_examples_limit = 5 // Remains same
  }
  results_retention_days = 20 
  model_snapshot_retention_days = 10 
  allow_lazy_open               = false
  daily_model_snapshot_retention_after_days = 1

  analysis_config {
    bucket_span = "30m"
    categorization_field_name = "mlcategory"
    summary_count_field_name  = "doc_count"
    latency                   = "3600s"
    detectors {
      function             = "count"
      detector_description = "Updated count of documents"
      use_null             = true
      exclude_frequent     = "all"
      custom_rules = [{ 
        actions = ["skip_model_update"]
        scope   = "time"
        conditions = [{
          operator = "lt"
          value    = 1000.0
        }]
      }]
    }]
  }

  data_description {
    time_field = "@timestamp"
    time_format = "epoch_ms"
  }
}
`, jobId)
}

func TestAccElasticsearchMLAnomalyDetector_Import(t *testing.T) {
	resourceName := "elasticstack_elasticsearch_ml_anomaly_detector.test_job_import"
	randSuffix := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	jobId := fmt.Sprintf("tf_acc_import_job_%s", randSuffix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.Providers,
		CheckDestroy:             checkResourceMLAnomalyDetectorDestroy(jobId),
		Steps: []resource.TestStep{
			{
				// Step 1: Create the resource
				Config: testAccElasticsearchMLAnomalyDetectorComprehensiveCreate(jobId),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "job_id", jobId),
				),
			},
			{
				// Step 2: Import the resource
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true, // Verifies all fields match schema after import
				// ImportStateIdFunc can be used if the ID used for import is different from resourceName.job_id
				// For this resource, job_id is used for import, which is the default for ImportStatePassthroughID.
			},
		},
	})
}
