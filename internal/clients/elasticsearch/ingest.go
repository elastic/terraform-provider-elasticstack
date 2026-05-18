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

package elasticsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	fwdiag "github.com/hashicorp/terraform-plugin-framework/diag"
)

func PutIngestPipeline(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, name string, pipeline map[string]any) fwdiag.Diagnostics {
	pipelineBytes, err := json.Marshal(pipeline)
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}
	typedClient, err := apiClient.GetESClient()
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}
	_, err = typedClient.Ingest.PutPipeline(name).Raw(bytes.NewReader(pipelineBytes)).Do(ctx)
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}
	return nil
}

func GetIngestPipeline(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, name string) (*types.IngestPipeline, fwdiag.Diagnostics) {
	typedClient, err := apiClient.GetESClient()
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}
	res, err := typedClient.Ingest.GetPipeline().Id(name).Do(ctx)
	if err != nil {
		if IsNotFoundElasticsearchError(err) {
			return nil, nil
		}
		return nil, diagutil.FrameworkDiagFromError(err)
	}
	if pipeline, ok := res[name]; ok {
		return &pipeline, nil
	}
	return nil, fwdiag.Diagnostics{
		fwdiag.NewErrorDiagnostic(
			"Unable to find ingest pipeline",
			fmt.Sprintf(`Unable to find "%s" ingest pipeline in the cluster`, name),
		),
	}
}

func DeleteIngestPipeline(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, name string) fwdiag.Diagnostics {
	typedClient, err := apiClient.GetESClient()
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}
	_, err = typedClient.Ingest.DeletePipeline(name).Do(ctx)
	if err != nil {
		if IsNotFoundElasticsearchError(err) {
			return nil
		}
		return diagutil.FrameworkDiagFromError(err)
	}
	return nil
}
