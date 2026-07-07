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

package trainedmodelalias_test

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"testing"

	"github.com/elastic/go-elasticsearch/v9/typedapi/types"
	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

const mlTrainedModelAliasResourceAddress = "elasticstack_elasticsearch_ml_trained_model_alias.test"

func TestAccResourceMLTrainedModelAlias_basic(t *testing.T) {
	versionutils.SkipIfUnsupported(t, version.Must(version.NewVersion("8.8.0")), versionutils.FlavorAny)
	acctest.EnsureTrainedModel(t)

	modelID := acctest.AccTestTrainedModelID
	modelAlias := fmt.Sprintf("test-alias-%s%s", sdkacctest.RandStringFromCharSet(9, sdkacctest.CharSetAlphaNum), sdkacctest.RandStringFromCharSet(1, sdkacctest.CharSetAlpha))

	t.Cleanup(func() {
		deleteMLTrainedModelAliasBestEffort(t.Context(), t, modelAlias)
	})

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceMLTrainedModelAliasDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: map[string]config.Variable{
					"model_alias": config.StringVariable(modelAlias),
					"model_id":    config.StringVariable(modelID),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(mlTrainedModelAliasResourceAddress, "model_alias", modelAlias),
					resource.TestCheckResourceAttr(mlTrainedModelAliasResourceAddress, "model_id", modelID),
					resource.TestCheckResourceAttr(mlTrainedModelAliasResourceAddress, "reassign", "true"),
					resource.TestMatchResourceAttr(mlTrainedModelAliasResourceAddress, "id",
						regexp.MustCompile(`^[A-Za-z0-9_-]{22}/`+regexp.QuoteMeta(modelAlias)+`$`)),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: map[string]config.Variable{
					"model_alias": config.StringVariable(modelAlias),
					"model_id":    config.StringVariable(modelID),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(mlTrainedModelAliasResourceAddress, "model_alias", modelAlias),
					resource.TestCheckResourceAttr(mlTrainedModelAliasResourceAddress, "model_id", modelID),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("import"),
				ResourceName:             mlTrainedModelAliasResourceAddress,
				ImportState:              true,
				ImportStateVerify:        true,
				ImportStateVerifyIgnore:  []string{"reassign"},
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs := s.RootModule().Resources[mlTrainedModelAliasResourceAddress]
					return rs.Primary.ID + "/" + rs.Primary.Attributes["model_id"], nil
				},
				ConfigVariables: map[string]config.Variable{
					"model_alias": config.StringVariable(modelAlias),
					"model_id":    config.StringVariable(modelID),
				},
			},
		},
	})
}

func TestAccResourceMLTrainedModelAlias_reassign(t *testing.T) {
	versionutils.SkipIfUnsupported(t, version.Must(version.NewVersion("8.8.0")), versionutils.FlavorAny)
	acctest.EnsureTrainedModelByID(t, acctest.AccTestTrainedModelID)
	acctest.EnsureTrainedModelByID(t, acctest.AccTestTrainedModelID2)

	modelAlias := fmt.Sprintf("test-alias-reassign-%s%s", sdkacctest.RandStringFromCharSet(9, sdkacctest.CharSetAlphaNum), sdkacctest.RandStringFromCharSet(1, sdkacctest.CharSetAlpha))

	t.Cleanup(func() {
		deleteMLTrainedModelAliasBestEffort(t.Context(), t, modelAlias)
	})

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceMLTrainedModelAliasDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: map[string]config.Variable{
					"model_alias": config.StringVariable(modelAlias),
					"model_a":     config.StringVariable(acctest.AccTestTrainedModelID),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(mlTrainedModelAliasResourceAddress, "model_alias", modelAlias),
					resource.TestCheckResourceAttr(mlTrainedModelAliasResourceAddress, "model_id", acctest.AccTestTrainedModelID),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables: map[string]config.Variable{
					"model_alias": config.StringVariable(modelAlias),
					"model_b":     config.StringVariable(acctest.AccTestTrainedModelID2),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(mlTrainedModelAliasResourceAddress, "model_alias", modelAlias),
					resource.TestCheckResourceAttr(mlTrainedModelAliasResourceAddress, "model_id", acctest.AccTestTrainedModelID2),
				),
			},
		},
	})
}

func TestAccResourceMLTrainedModelAlias_collisionWithReassignDisabled(t *testing.T) {
	versionutils.SkipIfUnsupported(t, version.Must(version.NewVersion("8.8.0")), versionutils.FlavorAny)
	acctest.EnsureTrainedModelByID(t, acctest.AccTestTrainedModelID)
	acctest.EnsureTrainedModelByID(t, acctest.AccTestTrainedModelID2)

	modelAlias := fmt.Sprintf("test-alias-collision-%s%s", sdkacctest.RandStringFromCharSet(9, sdkacctest.CharSetAlphaNum), sdkacctest.RandStringFromCharSet(1, sdkacctest.CharSetAlpha))

	// Pre-create the alias on model A.
	ctx := context.Background()
	client, err := clients.NewAcceptanceTestingElasticsearchScopedClient()
	if err != nil {
		t.Skipf("Cannot connect to Elasticsearch: %v", err)
	}
	_, err = client.GetESClient().Ml.PutTrainedModelAlias(acctest.AccTestTrainedModelID, modelAlias).Reassign(true).Do(ctx)
	if err != nil {
		t.Fatalf("failed to pre-create alias for collision test: %v", err)
	}

	t.Cleanup(func() {
		deleteMLTrainedModelAliasBestEffort(t.Context(), t, modelAlias)
	})

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceMLTrainedModelAliasDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: map[string]config.Variable{
					"model_alias": config.StringVariable(modelAlias),
					"model_id":    config.StringVariable(acctest.AccTestTrainedModelID2),
				},
				ExpectError: regexp.MustCompile(`.*`),
			},
		},
	})
}

func TestAccResourceMLTrainedModelAlias_updateReassignFlag(t *testing.T) {
	versionutils.SkipIfUnsupported(t, version.Must(version.NewVersion("8.8.0")), versionutils.FlavorAny)
	acctest.EnsureTrainedModel(t)

	modelID := acctest.AccTestTrainedModelID
	modelAlias := fmt.Sprintf("test-alias-flag-%s%s", sdkacctest.RandStringFromCharSet(9, sdkacctest.CharSetAlphaNum), sdkacctest.RandStringFromCharSet(1, sdkacctest.CharSetAlpha))

	t.Cleanup(func() {
		deleteMLTrainedModelAliasBestEffort(t.Context(), t, modelAlias)
	})

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceMLTrainedModelAliasDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: map[string]config.Variable{
					"model_alias": config.StringVariable(modelAlias),
					"model_id":    config.StringVariable(modelID),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(mlTrainedModelAliasResourceAddress, "model_alias", modelAlias),
					resource.TestCheckResourceAttr(mlTrainedModelAliasResourceAddress, "model_id", modelID),
					resource.TestCheckResourceAttr(mlTrainedModelAliasResourceAddress, "reassign", "true"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables: map[string]config.Variable{
					"model_alias": config.StringVariable(modelAlias),
					"model_id":    config.StringVariable(modelID),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(mlTrainedModelAliasResourceAddress, "model_alias", modelAlias),
					resource.TestCheckResourceAttr(mlTrainedModelAliasResourceAddress, "model_id", modelID),
					resource.TestCheckResourceAttr(mlTrainedModelAliasResourceAddress, "reassign", "false"),
				),
			},
		},
	})
}

func checkResourceMLTrainedModelAliasDestroy(s *terraform.State) error {
	client, err := clients.NewAcceptanceTestingElasticsearchScopedClient()
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "elasticstack_elasticsearch_ml_trained_model_alias" {
			continue
		}

		alias := rs.Primary.Attributes["model_alias"]
		if alias == "" {
			continue
		}

		res, err := client.GetESClient().Ml.GetTrainedModels().ModelId(alias).AllowNoMatch(true).Do(context.Background())
		if err != nil {
			var esErr *types.ElasticsearchError
			if errors.As(err, &esErr) && esErr.Status == 404 {
				continue
			}
			return err
		}
		if res == nil || len(res.TrainedModelConfigs) == 0 {
			continue
		}

		return fmt.Errorf("ML trained model alias %q still exists", alias)
	}

	return nil
}

func deleteMLTrainedModelAliasBestEffort(ctx context.Context, t *testing.T, alias string) {
	t.Helper()

	client, err := clients.NewAcceptanceTestingElasticsearchScopedClient()
	if err != nil {
		t.Logf("ML trained model alias cleanup: no client: %v", err)
		return
	}
	typed := client.GetESClient()

	// Resolve current model ID via GET, then delete
	res, err := typed.Ml.GetTrainedModels().ModelId(alias).AllowNoMatch(true).Do(ctx)
	if err != nil {
		var esErr *types.ElasticsearchError
		if errors.As(err, &esErr) && esErr.Status == 404 {
			return
		}
		t.Logf("ML trained model alias GET cleanup %q: %v", alias, err)
		return
	}
	if res == nil || len(res.TrainedModelConfigs) == 0 {
		return
	}

	modelID := res.TrainedModelConfigs[0].ModelId
	_, err = typed.Ml.DeleteTrainedModelAlias(modelID, alias).Do(ctx)
	if err != nil {
		var esErr *types.ElasticsearchError
		if errors.As(err, &esErr) && esErr.Status == 404 {
			return
		}
		t.Logf("ML trained model alias DELETE cleanup %q: %v", alias, err)
	}
}
