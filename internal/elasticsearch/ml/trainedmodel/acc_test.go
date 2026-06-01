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
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	esclient "github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const dataSourceAddress = "data.elasticstack_elasticsearch_ml_trained_model.test"

// skipIfNoTrainedModel checks whether a built-in trained model exists in the
// acceptance cluster. Returns (true, nil) when the model is absent so the
// test should be skipped.
func skipIfNoTrainedModel(modelID string) func() (bool, error) {
	return func() (bool, error) {
		client, err := clients.NewAcceptanceTestingElasticsearchScopedClient()
		if err != nil {
			return false, err
		}

		_, found, diags := esclient.GetTrainedModel(context.Background(), client, modelID)
		if diags.HasError() {
			return false, fmt.Errorf("error checking for trained model: %v", diags.Errors())
		}

		if !found {
			return true, nil
		}
		return false, nil
	}
}

func TestAccDataSourceMLTrainedModel_basic(t *testing.T) {
	modelID := "lang_ident_model_current"

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				ConfigVariables: config.Variables{
					"model_id": config.StringVariable(modelID),
				},
				SkipFunc: skipIfNoTrainedModel(modelID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceAddress, "model_id", modelID),
					resource.TestCheckResourceAttrSet(dataSourceAddress, "id"),
					resource.TestMatchResourceAttr(dataSourceAddress, "id",
						regexp.MustCompile(`^[A-Za-z0-9_-]{22}/`+regexp.QuoteMeta(modelID)+`$`)),
					resource.TestCheckResourceAttrSet(dataSourceAddress, "model_type"),
					resource.TestCheckResourceAttrSet(dataSourceAddress, "version"),
					resource.TestCheckResourceAttrSet(dataSourceAddress, "create_time"),
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
