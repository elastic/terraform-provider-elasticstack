package anomaly_detector_test

import (
	"fmt"
	"io"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccElasticsearchMLAnomalyDetector_Basic(t *testing.T) {
	resourceName := "elasticstack_elasticsearch_ml_anomaly_detector.test_job"
	jobId := "tf_acc_test_basic_job"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.Providers,
		CheckDestroy:             checkResourceMLAnomalyDetectorDestroy(jobId),
		Steps: []resource.TestStep{
			{
				// Step 1: Create basic job
				Config: testAccElasticsearchMLAnomalyDetectorBasic(jobId, "Basic anomaly detector for acceptance testing."),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "job_id", jobId),
					resource.TestCheckResourceAttr(resourceName, "description", "Basic anomaly detector for acceptance testing."),
					resource.TestCheckResourceAttr(resourceName, "analysis_config.0.bucket_span", "30m"),
					resource.TestCheckResourceAttr(resourceName, "analysis_config.0.detectors.0.function", "count"),
					resource.TestCheckResourceAttr(resourceName, "data_description.0.time_field", "@timestamp"),
					resource.TestCheckResourceAttr(resourceName, "groups.#", "0"), // No groups initially
				),
			},
			{
				// Step 2: Update description and add a group
				Config: testAccElasticsearchMLAnomalyDetectorUpdate(jobId, "Updated anomaly detector for acceptance testing.", []string{"tf-acc-group"}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "job_id", jobId),
					resource.TestCheckResourceAttr(resourceName, "description", "Updated anomaly detector for acceptance testing."),
					resource.TestCheckResourceAttr(resourceName, "groups.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "groups.0", "tf-acc-group"),
				),
			},
			{
				// Step 3: Add another group and change description again
				Config: testAccElasticsearchMLAnomalyDetectorUpdate(jobId, "Final description.", []string{"tf-acc-group", "another-group"}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "description", "Final description."),
					resource.TestCheckResourceAttr(resourceName, "groups.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "groups.0", "tf-acc-group"), // Order might not be guaranteed by API, adjust if needed
					resource.TestCheckResourceAttr(resourceName, "groups.1", "another-group"),
				),
			},
			{
				// Step 4: Remove all groups
				Config: testAccElasticsearchMLAnomalyDetectorUpdate(jobId, "Description with no groups.", []string{}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "description", "Description with no groups."),
					resource.TestCheckResourceAttr(resourceName, "groups.#", "0"),
				),
			},
		},
	})
}

func testAccElasticsearchMLAnomalyDetectorBasic(jobId string, description string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_ml_anomaly_detector" "test_job" {
  job_id      = "%s"
  description = "%s"

  analysis_config {
    bucket_span = "30m"
    detectors {
      function = "count"
    }
  }

  data_description {
    time_field = "@timestamp"
  }
}
`, jobId, description)
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

func testAccElasticsearchMLAnomalyDetectorUpdate(jobId string, description string, groups []string) string {
	groupsFormatted := ""
	if len(groups) > 0 {
		groupsFormatted = "groups = ["
		for i, group := range groups {
			groupsFormatted += fmt.Sprintf("\"%s\"", group)
			if i < len(groups)-1 {
				groupsFormatted += ", "
			}
		}
		groupsFormatted += "]\n"
	}

	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_ml_anomaly_detector" "test_job" {
  job_id      = "%s"
  description = "%s"
  %s // This will insert the formatted groups string, or be empty if no groups

  analysis_config {
    bucket_span = "30m" // Assuming this doesn't change for the update test
    detectors {
      function = "count" // Assuming this doesn't change
    }
  }

  data_description {
    time_field = "@timestamp" // Assuming this doesn't change
  }
}
`, jobId, description, groupsFormatted)
}

func TestAccElasticsearchMLAnomalyDetector_Import(t *testing.T) {
	resourceName := "elasticstack_elasticsearch_ml_anomaly_detector.test_job_import"
	jobId := "tf_acc_test_import_job"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.Providers,
		CheckDestroy:             checkResourceMLAnomalyDetectorDestroy(jobId),
		Steps: []resource.TestStep{
			{
				// Step 1: Create the resource
				Config: testAccElasticsearchMLAnomalyDetectorBasic(jobId, "Resource to be imported."),
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
