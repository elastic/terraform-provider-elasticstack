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
	"net/http"

	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types/enums/expandwildcard"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	fwdiags "github.com/hashicorp/terraform-plugin-framework/diag"
)

func PutDataStream(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, dataStreamName string) fwdiags.Diagnostics {
	typedClient := apiClient.GetESClient()
	_, err := typedClient.Indices.CreateDataStream(dataStreamName).Do(ctx)
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}
	return nil
}

func GetDataStream(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, dataStreamName string) (*types.DataStream, fwdiags.Diagnostics) {
	typedClient := apiClient.GetESClient()
	res, err := typedClient.Indices.GetDataStream().Name(dataStreamName).Do(ctx)
	if err != nil {
		if IsNotFoundElasticsearchError(err) {
			return nil, nil
		}
		return nil, diagutil.FrameworkDiagFromError(err)
	}
	if len(res.DataStreams) == 0 {
		return nil, nil
	}
	ds := res.DataStreams[0]
	return &ds, nil
}

func DeleteDataStream(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, dataStreamName string) fwdiags.Diagnostics {
	typedClient := apiClient.GetESClient()
	_, err := typedClient.Indices.DeleteDataStream(dataStreamName).Do(ctx)
	if err != nil {
		if IsNotFoundElasticsearchError(err) {
			return nil
		}
		return diagutil.FrameworkDiagFromError(err)
	}
	return nil
}

func PutDataStreamLifecycle(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, dataStreamName string, expandWildcards string, lifecycle models.LifecycleSettings) fwdiags.Diagnostics {
	typedClient := apiClient.GetESClient()

	reqBody := map[string]any{}
	if lifecycle.DataRetention != "" {
		reqBody["data_retention"] = lifecycle.DataRetention
	}
	reqBody["enabled"] = lifecycle.Enabled
	if len(lifecycle.Downsampling) > 0 {
		ds := make([]map[string]any, len(lifecycle.Downsampling))
		for i, d := range lifecycle.Downsampling {
			ds[i] = map[string]any{
				"after":          d.After,
				"fixed_interval": d.FixedInterval,
			}
		}
		reqBody["downsampling"] = ds
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}

	builder := typedClient.Indices.PutDataLifecycle(dataStreamName).Raw(bytes.NewReader(bodyBytes))
	if expandWildcards != "" {
		builder = builder.ExpandWildcards(expandwildcard.ExpandWildcard{Name: expandWildcards})
	}
	_, err = builder.Do(ctx)
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}
	return nil
}

func GetDataStreamLifecycle(
	ctx context.Context,
	apiClient *clients.ElasticsearchScopedClient,
	dataStreamName string,
	expandWildcards string,
) (*models.DataStreamLifecycleResponse, fwdiags.Diagnostics) {
	typedClient := apiClient.GetESClient()

	call := typedClient.Indices.GetDataLifecycle(dataStreamName)
	if expandWildcards != "" {
		call = call.ExpandWildcards(expandwildcard.ExpandWildcard{Name: expandWildcards})
	}
	res, err := call.Perform(ctx)
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if d := diagutil.CheckHTTPErrorFromFW(res, "Unable to get data stream lifecycle"); d.HasError() {
		return nil, d
	}

	var response models.DataStreamLifecycleResponse
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}

	return &response, nil
}

func DeleteDataStreamLifecycle(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, dataStreamName string, expandWildcards string) fwdiags.Diagnostics {
	// On Elastic Cloud Serverless, data stream lifecycle and retention are
	// managed by Elastic. The delete-lifecycle API answers with HTTP 410
	// (api_not_available_exception), which is uncatchable and leaves
	// `terraform destroy` unable to complete. There is nothing for the provider
	// to delete, so skip the request and surface a warning instead of failing.
	isServerless, diags := apiClient.IsServerless(ctx)
	if diags.HasError() {
		return diags
	}
	if isServerless {
		return fwdiags.Diagnostics{
			fwdiags.NewWarningDiagnostic(
				"Data stream lifecycle removal skipped on serverless",
				"Elastic Cloud Serverless manages data stream lifecycle and retention automatically, "+
					"so the lifecycle cannot be removed through the API. The resource has been removed "+
					"from Terraform state without changing the data stream on the server.",
			),
		}
	}

	typedClient := apiClient.GetESClient()
	builder := typedClient.Indices.DeleteDataLifecycle(dataStreamName)
	if expandWildcards != "" {
		builder = builder.ExpandWildcards(expandwildcard.ExpandWildcard{Name: expandWildcards})
	}
	_, err := builder.Do(ctx)
	if err != nil {
		if IsNotFoundElasticsearchError(err) {
			return nil
		}
		return diagutil.FrameworkDiagFromError(err)
	}
	return nil
}
