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

package acctest

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"sync"

	"golang.org/x/sync/errgroup"
	"testing"

	"github.com/elastic/go-elasticsearch/v9/typedapi/ml/puttrainedmodel"
	"github.com/elastic/go-elasticsearch/v9/typedapi/ml/puttrainedmodeldefinitionpart"
	"github.com/elastic/go-elasticsearch/v9/typedapi/ml/puttrainedmodelvocabulary"
	estypes "github.com/elastic/go-elasticsearch/v9/typedapi/types"
	"github.com/elastic/go-elasticsearch/v9/typedapi/types/enums/trainedmodeltype"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	esclient "github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// FixturePyTorchModelID is the stable model ID for the fixture PyTorch model.
const FixturePyTorchModelID = "terraform-acc-test-pytorch-fixture"

// fixtureEnsureOnce ensures the fixture is created at most once per test run.
var fixtureEnsureOnce sync.Once
var errFixtureEnsure error

// EnsureFixturePyTorchModel creates a TorchScript model fixture using only raw
// Elasticsearch ML APIs (no Eland, no Python).
// It returns the model ID. On capacity failure it calls t.Skip.
func EnsureFixturePyTorchModel(t *testing.T) string {
	t.Helper()
	fixtureEnsureOnce.Do(func() {
		errFixtureEnsure = ensureFixturePyTorchModel(t)
	})
	if errFixtureEnsure != nil {
		t.Fatal(errFixtureEnsure)
	}
	return FixturePyTorchModelID
}

func ensureFixturePyTorchModel(t *testing.T) error {
	t.Helper()

	client, err := clients.NewAcceptanceTestingElasticsearchScopedClient()
	if err != nil {
		return fmt.Errorf("failed to create elasticsearch client: %w", err)
	}
	es := client.GetESClient()
	ctx := context.Background()

	// Check if the model already exists from a previous run.
	_, found, _ := esclient.GetTrainedModel(ctx, client, FixturePyTorchModelID)
	if found {
		return nil
	}

	fixtureDir := testDataFixtureDir()

	configData, err := os.ReadFile(filepath.Join(fixtureDir, "model_config.json"))
	if err != nil {
		return fmt.Errorf("reading model_config.json: %w", err)
	}
	var cfg struct {
		Description     string          `json:"description"`
		Input           json.RawMessage `json:"input"`
		InferenceConfig json.RawMessage `json:"inference_config"`
	}
	if err := json.Unmarshal(configData, &cfg); err != nil {
		return fmt.Errorf("parsing model_config.json: %w", err)
	}
	desc := cfg.Description
	input := &estypes.Input{}
	if err := json.Unmarshal(cfg.Input, input); err != nil {
		return fmt.Errorf("parsing input: %w", err)
	}
	inferenceConfig := &estypes.InferenceConfigCreateContainer{}
	if err := json.Unmarshal(cfg.InferenceConfig, inferenceConfig); err != nil {
		return fmt.Errorf("parsing inference_config: %w", err)
	}
	modelType := trainedmodeltype.Pytorch

	putReq := &puttrainedmodel.Request{
		Description:     &desc,
		ModelType:       &modelType,
		Input:           input,
		InferenceConfig: inferenceConfig,
	}

	_, err = es.Ml.PutTrainedModel(FixturePyTorchModelID).Request(putReq).Do(ctx)
	if err != nil {
		return fmt.Errorf("creating model config for %q: %w", FixturePyTorchModelID, err)
	}

	ptData, err := os.ReadFile(filepath.Join(fixtureDir, "traced_pytorch_model.pt"))
	if err != nil {
		return fmt.Errorf("reading traced_pytorch_model.pt: %w", err)
	}

	const chunkSize = 1024 * 1024
	totalParts := (len(ptData) + chunkSize - 1) / chunkSize
	group, groupCtx := errgroup.WithContext(ctx)
	group.SetLimit(4)
	for i := range totalParts {
		partIndex := i
		start := partIndex * chunkSize
		end := min(start+chunkSize, len(ptData))
		definition := base64.StdEncoding.EncodeToString(ptData[start:end])
		group.Go(func() error {
			_, err := es.Ml.PutTrainedModelDefinitionPart(FixturePyTorchModelID, strconv.Itoa(partIndex)).
				Request(&puttrainedmodeldefinitionpart.Request{
					Definition:            definition,
					TotalDefinitionLength: int64(len(ptData)),
					TotalParts:            totalParts,
				}).
				Do(groupCtx)
			if err != nil {
				return fmt.Errorf("uploading definition part %d: %w", partIndex, err)
			}
			return nil
		})
	}
	if err := group.Wait(); err != nil {
		return err
	}

	vocabData, err := os.ReadFile(filepath.Join(fixtureDir, "vocabulary.json"))
	if err != nil {
		return fmt.Errorf("reading vocabulary.json: %w", err)
	}
	var vocabReq struct {
		Vocabulary []string `json:"vocabulary"`
	}
	if err := json.Unmarshal(vocabData, &vocabReq); err != nil {
		return fmt.Errorf("parsing vocabulary.json: %w", err)
	}
	_, err = es.Ml.PutTrainedModelVocabulary(FixturePyTorchModelID).
		Request(&puttrainedmodelvocabulary.Request{
			Vocabulary: vocabReq.Vocabulary,
		}).
		Do(ctx)
	if err != nil {
		return fmt.Errorf("uploading vocabulary: %w", err)
	}

	t.Cleanup(func() {
		_, stopErr := es.Ml.StopTrainedModelDeployment(FixturePyTorchModelID).Force(true).Do(ctx)
		if stopErr != nil {
			tflog.Warn(ctx, "failed to stop fixture model deployment during cleanup", map[string]any{
				"model_id": FixturePyTorchModelID,
				"error":    stopErr.Error(),
			})
		}
	})

	return nil
}

// testDataFixtureDir resolves the path to the model fixture directory.
func testDataFixtureDir() string {
	_, srcFile, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(srcFile), "testdata", "model-fixture")
}
