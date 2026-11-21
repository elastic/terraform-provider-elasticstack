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

package trainedmodel_test

import (
	"regexp"
	"strconv"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const dataSourceAddress = "data.elasticstack_elasticsearch_ml_trained_model.test"

func TestAccDataSourceMLTrainedModel_basic(t *testing.T) {
	modelID := acctest.AccTestTrainedModelID

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.EnsureTrainedModel(t)
		},
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				ConfigVariables: config.Variables{
					"model_id": config.StringVariable(modelID),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceAddress, "model_id", modelID),
					resource.TestCheckResourceAttrSet(dataSourceAddress, "id"),
					resource.TestMatchResourceAttr(dataSourceAddress, "id",
						regexp.MustCompile(`^[A-Za-z0-9_-]{22}/`+regexp.QuoteMeta(modelID)+`$`)),
					resource.TestCheckResourceAttr(dataSourceAddress, "model_type", "tree_ensemble"),
					resource.TestCheckResourceAttr(dataSourceAddress, "description", "Terraform acceptance test trained model"),
					// The get-trained-models call does not request "definition_status", so the API omits
					// fully_defined from the response.
					resource.TestCheckNoResourceAttr(dataSourceAddress, "fully_defined"),
					// Model size is deterministic for the same tiny tree model; assert it is a positive integer.
					resource.TestMatchResourceAttr(dataSourceAddress, "model_size_bytes", regexp.MustCompile(`^[1-9][0-9]*$`)),
					resource.TestMatchResourceAttr(dataSourceAddress, "create_time", regexp.MustCompile(`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d{3}Z$`)),
					// Created via the REST API with an explicit user, so created_by reflects the API user,
					// not the "_xpack" value used for models created by data frame analytics.
					resource.TestCheckResourceAttr(dataSourceAddress, "created_by", "api_user"),
					resource.TestMatchResourceAttr(dataSourceAddress, "version", regexp.MustCompile(`^\d+\.\d+\.\d+`)),
					resource.TestMatchResourceAttr(dataSourceAddress, "license_level", regexp.MustCompile(`^(basic|trial|gold|standard|platinum|enterprise)$`)),
					// The fixture is a tree_ensemble (not a PyTorch model), so no platform architecture applies.
					resource.TestCheckNoResourceAttr(dataSourceAddress, "platform_architecture"),
					// The fixture sets no default field map. metadata_json is asserted on the dedicated
					// tagged-model fixture instead: this shared model may accumulate a "model_aliases"
					// metadata entry from the trained-model-alias acceptance tests running against it.
					resource.TestCheckNoResourceAttr(dataSourceAddress, "default_field_map"),
					// The test model has no tags
					resource.TestCheckResourceAttr(dataSourceAddress, "tags.#", "0"),
					// The test model's input field list is exactly ["foo"]
					resource.TestCheckResourceAttr(dataSourceAddress, "input_json", `{"field_names":["foo"]}`),
					// Inference config is present (regression model)
					resource.TestMatchResourceAttr(dataSourceAddress, "inference_config_json", regexp.MustCompile(`"regression"`)),
				),
			},
		},
	})
}

func TestAccDataSourceMLTrainedModel_tags(t *testing.T) {
	modelID := acctest.AccTestTrainedModelIDTagged

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.EnsureTaggedTrainedModel(t)
		},
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				ConfigVariables: config.Variables{
					"model_id": config.StringVariable(modelID),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceAddress, "model_id", modelID),
					resource.TestCheckResourceAttr(dataSourceAddress, "tags.#", strconv.Itoa(len(acctest.AccTestTrainedModelTags))),
					resource.TestCheckTypeSetElemAttr(dataSourceAddress, "tags.*", acctest.AccTestTrainedModelTags[0]),
					resource.TestCheckTypeSetElemAttr(dataSourceAddress, "tags.*", acctest.AccTestTrainedModelTags[1]),
					// This fixture is never aliased by other tests, so metadata_json stays absent.
					resource.TestCheckNoResourceAttr(dataSourceAddress, "metadata_json"),
				),
			},
		},
	})
}

func TestAccDataSourceMLTrainedModel_notFound(t *testing.T) {
	modelID := "non-existent-model-" + sdkacctest.RandStringFromCharSet(8, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				ConfigVariables: config.Variables{
					"model_id": config.StringVariable(modelID),
				},
				ExpectError: regexp.MustCompile(`elasticsearch_ml_trained_model not found`),
			},
		},
	})
}
