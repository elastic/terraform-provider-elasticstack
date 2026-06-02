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
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceAddress, "model_id", modelID),
					resource.TestCheckResourceAttr(dataSourceAddress, "id", ""),
				),
			},
		},
	})
}
