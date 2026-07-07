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
	"io"
	"strings"

	"github.com/elastic/go-elasticsearch/v9/typedapi/indices/create"
	"github.com/elastic/go-elasticsearch/v9/typedapi/types"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	fwdiags "github.com/hashicorp/terraform-plugin-framework/diag"
)

// PutIndex creates an Elasticsearch index and returns the concrete index name from the
// API response together with any diagnostics.  When the configured name is a plain date
// math expression it is URI-encoded before being sent in the API request path so the Go
// HTTP client does not rewrite the angle brackets or braces.
func PutIndex(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, index *models.Index, params *models.PutIndexParams) (string, fwdiags.Diagnostics) {
	typedClient := apiClient.GetESClient()

	indexBytes, err := json.Marshal(index)
	if err != nil {
		return "", diagutil.FrameworkDiagFromError(err)
	}

	call := typedClient.Indices.Create(index.Name).Raw(bytes.NewReader(indexBytes))
	if params.WaitForActiveShards != "" {
		call = call.WaitForActiveShards(params.WaitForActiveShards)
	}
	if params.MasterTimeout > 0 {
		call = call.MasterTimeout(durationToMsString(params.MasterTimeout))
	}
	if params.Timeout > 0 {
		call = call.Timeout(durationToMsString(params.Timeout))
	}

	// For date-math index names we must build the request manually and set
	// URL.RawPath so the already-encoded characters (including %2F) are not
	// double-encoded by url.URL.String().
	// Remove when https://github.com/elastic/go-elasticsearch/pull/1425 is available.
	var res *create.Response
	if DateMathIndexNameRe.MatchString(index.Name) {
		req, err := call.HttpRequest(ctx)
		if err != nil {
			return "", diagutil.FrameworkDiagFromError(err)
		}
		req.URL.RawPath = "/" + encodeDateMathIndexName(index.Name)
		req.URL.Path = "/" + index.Name

		httpRes, err := typedClient.Transport.Perform(req)
		if err != nil {
			return "", diagutil.FrameworkDiagFromError(err)
		}
		defer httpRes.Body.Close()

		if httpRes.StatusCode >= 400 {
			body, _ := io.ReadAll(httpRes.Body)
			return "", fwdiags.Diagnostics{fwdiags.NewErrorDiagnostic(
				fmt.Sprintf("Unable to create index: %s", index.Name),
				fmt.Sprintf("status: %d, body: %s", httpRes.StatusCode, string(body)),
			)}
		}
		// Indices.Create response always contains the resolved index name.
		// We cannot parse the typed response here because the typed
		// response would try to decode an error body on non-2xx.
		var createRes create.Response
		if err := json.NewDecoder(httpRes.Body).Decode(&createRes); err != nil {
			return "", diagutil.FrameworkDiagFromError(err)
		}
		res = &createRes
	} else {
		var err error
		res, err = call.Do(ctx)
		if err != nil {
			return "", diagutil.FrameworkDiagFromError(err)
		}
	}

	concreteName := res.Index
	if concreteName == "" {
		concreteName = index.Name
	}
	return concreteName, nil
}

func DeleteIndex(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, name string) fwdiags.Diagnostics {
	typedClient := apiClient.GetESClient()
	_, err := typedClient.Indices.Delete(name).Do(ctx)
	if err != nil {
		if IsNotFoundElasticsearchError(err) {
			return nil
		}
		return diagutil.FrameworkDiagFromError(err)
	}
	return nil
}

// GetIndex retrieves a single index by its concrete name.  The caller is responsible
// for supplying the concrete index name (e.g. from id.ResourceID or the create
// response), not a date math expression.  The response map key for a concrete name
// always equals the requested name, so a direct key lookup is sufficient.
func GetIndex(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, name string) (*types.IndexState, fwdiags.Diagnostics) {
	indices, diags := GetIndices(ctx, apiClient, name)
	if diags.HasError() {
		return nil, diags
	}

	if index, ok := indices[name]; ok {
		return &index, nil
	}

	return nil, nil
}

func GetIndices(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, name string) (map[string]types.IndexState, fwdiags.Diagnostics) {
	typedClient := apiClient.GetESClient()
	res, err := typedClient.Indices.Get(name).FlatSettings(true).Do(ctx)
	if err != nil {
		if IsNotFoundElasticsearchError(err) {
			return nil, nil
		}
		return nil, diagutil.FrameworkDiagFromError(err)
	}
	return res, nil
}

func UpdateIndexSettings(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, index string, settings map[string]any) fwdiags.Diagnostics {
	settingsBytes, err := json.Marshal(settings)
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}
	typedClient := apiClient.GetESClient()
	_, err = typedClient.Indices.PutSettings().Indices(index).Raw(bytes.NewReader(settingsBytes)).Do(ctx)
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}
	return nil
}

func UpdateIndexMappings(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, index, mappings string) fwdiags.Diagnostics {
	typedClient := apiClient.GetESClient()
	_, err := typedClient.Indices.PutMapping(index).Raw(strings.NewReader(mappings)).Do(ctx)
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}
	return nil
}
