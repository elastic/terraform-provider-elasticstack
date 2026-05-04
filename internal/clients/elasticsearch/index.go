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
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/elastic/go-elasticsearch/v8/typedapi/ilm/putlifecycle"
	"github.com/elastic/go-elasticsearch/v8/typedapi/indices/getdatalifecycle"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types/enums/expandwildcard"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	fwdiags "github.com/hashicorp/terraform-plugin-framework/diag"
	sdkdiag "github.com/hashicorp/terraform-plugin-sdk/v2/diag"
)

// DateMathIndexNameRe matches plain Elasticsearch date math index name expressions.
// The pattern enforces:
//   - opening `<`
//   - a static prefix that starts with a valid non-start character (not -, _, +) and
//     uses only the same character set allowed in ordinary static index names
//   - at least one `{…}` section (the date math expression itself)
//   - a closing `>` immediately after the last `}`
//
// This keeps the two validation paths (static vs date-math) consistent and avoids
// accepting expressions that would be rejected as static names.
var DateMathIndexNameRe = regexp.MustCompile(`^<[^-_+][a-z0-9!$%&'()+.;=@[\]^{}~_-]*\{[^<>]+\}>$`)

// encodeDateMathIndexName URI-encodes a plain date math index name for use in an API
// request path.  Characters inside the expression that have special meaning in a URL
// path are percent-encoded so the Go HTTP client does not rewrite them.
// encodeDateMathIndexName URI-encodes a plain date math index name for use in an API
// request path.  Characters inside the expression that have special meaning in a URL
// path are percent-encoded so the Go HTTP client does not rewrite them.
func encodeDateMathIndexName(name string) string {
	// url.PathEscape encodes the string so it is safe for a path segment.
	return url.PathEscape(name)
}

func PutIlm(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, policy *models.Policy) fwdiags.Diagnostics {
	policyBytes, err := json.Marshal(map[string]any{"policy": policy})
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}

	typedClient, err := apiClient.GetESTypedClient()
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}
	var req putlifecycle.Request
	if err := json.Unmarshal(policyBytes, &req); err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}
	_, err = typedClient.Ilm.PutLifecycle(policy.Name).Request(&req).Do(ctx)
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}
	return nil
}

func GetIlm(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, policyName string) (*types.Lifecycle, fwdiags.Diagnostics) {
	typedClient, err := apiClient.GetESTypedClient()
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}
	res, err := typedClient.Ilm.GetLifecycle().Policy(policyName).Do(ctx)
	if err != nil {
		var esErr *types.ElasticsearchError
		if errors.As(err, &esErr) && esErr.Status == 404 {
			return nil, nil
		}
		return nil, diagutil.FrameworkDiagFromError(err)
	}
	if lifecycle, ok := res[policyName]; ok {
		return &lifecycle, nil
	}
	return nil, fwdiags.Diagnostics{
		fwdiags.NewErrorDiagnostic(
			"Unable to find a ILM policy in the cluster",
			fmt.Sprintf(`Unable to find "%s" ILM policy in the cluster`, policyName),
		),
	}
}

func DeleteIlm(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, policyName string) fwdiags.Diagnostics {
	typedClient, err := apiClient.GetESTypedClient()
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}
	_, err = typedClient.Ilm.DeleteLifecycle(policyName).Do(ctx)
	if err != nil {
		var esErr *types.ElasticsearchError
		if errors.As(err, &esErr) && esErr.Status == 404 {
			return nil
		}
		return diagutil.FrameworkDiagFromError(err)
	}
	return nil
}

func PutComponentTemplate(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, template *models.ComponentTemplate) sdkdiag.Diagnostics {
	typedClient, err := apiClient.GetESTypedClient()
	if err != nil {
		return sdkdiag.FromErr(err)
	}

	templateBytes, err := json.Marshal(template)
	if err != nil {
		return sdkdiag.FromErr(err)
	}

	_, err = typedClient.Cluster.PutComponentTemplate(template.Name).Raw(bytes.NewReader(templateBytes)).Do(ctx)
	if err != nil {
		return sdkdiag.FromErr(err)
	}
	return nil
}

// GetComponentTemplate returns a component template by name.
func GetComponentTemplate(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, templateName string) (*types.ClusterComponentTemplate, sdkdiag.Diagnostics) {
	typedClient, err := apiClient.GetESTypedClient()
	if err != nil {
		return nil, sdkdiag.FromErr(err)
	}
	res, err := typedClient.Cluster.GetComponentTemplate().Name(templateName).Do(ctx)
	if err != nil {
		var esErr *types.ElasticsearchError
		if errors.As(err, &esErr) && esErr.Status == 404 {
			return nil, nil
		}
		return nil, sdkdiag.FromErr(err)
	}
	if len(res.ComponentTemplates) != 1 {
		return nil, sdkdiag.Diagnostics{
			sdkdiag.Diagnostic{
				Severity: sdkdiag.Error,
				Summary:  "Wrong number of templates returned",
				Detail:   fmt.Sprintf("Elasticsearch API returned %d when requested '%s' component template.", len(res.ComponentTemplates), templateName),
			},
		}
	}
	tpl := res.ComponentTemplates[0]
	return &tpl, nil
}

func DeleteComponentTemplate(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, templateName string) sdkdiag.Diagnostics {
	typedClient, err := apiClient.GetESTypedClient()
	if err != nil {
		return sdkdiag.FromErr(err)
	}
	_, err = typedClient.Cluster.DeleteComponentTemplate(templateName).Do(ctx)
	if err != nil {
		var esErr *types.ElasticsearchError
		if errors.As(err, &esErr) && esErr.Status == 404 {
			return nil
		}
		return sdkdiag.FromErr(err)
	}
	return nil
}

func PutIndexTemplate(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, template *models.IndexTemplate) fwdiags.Diagnostics {
	typedClient, err := apiClient.GetESTypedClient()
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}

	templateBytes, err := json.Marshal(template)
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}

	_, err = typedClient.Indices.PutIndexTemplate(template.Name).Raw(bytes.NewReader(templateBytes)).Do(ctx)
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}
	return nil
}

func GetIndexTemplate(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, templateName string) (*types.IndexTemplateItem, fwdiags.Diagnostics) {
	typedClient, err := apiClient.GetESTypedClient()
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}
	res, err := typedClient.Indices.GetIndexTemplate().Name(templateName).Do(ctx)
	if err != nil {
		var esErr *types.ElasticsearchError
		if errors.As(err, &esErr) && esErr.Status == 404 {
			return nil, nil
		}
		return nil, diagutil.FrameworkDiagFromError(err)
	}
	if len(res.IndexTemplates) != 1 {
		return nil, fwdiags.Diagnostics{
			fwdiags.NewErrorDiagnostic(
				"Wrong number of templates returned",
				fmt.Sprintf("Elasticsearch API returned %d when requested '%s' template.", len(res.IndexTemplates), templateName),
			),
		}
	}
	tpl := res.IndexTemplates[0]
	return &tpl, nil
}

func DeleteIndexTemplate(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, templateName string) fwdiags.Diagnostics {
	typedClient, err := apiClient.GetESTypedClient()
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}
	_, err = typedClient.Indices.DeleteIndexTemplate(templateName).Do(ctx)
	if err != nil {
		var esErr *types.ElasticsearchError
		if errors.As(err, &esErr) && esErr.Status == 404 {
			return nil
		}
		return diagutil.FrameworkDiagFromError(err)
	}
	return nil
}

// PutIndex creates an Elasticsearch index and returns the concrete index name from the
// API response together with any diagnostics.  When the configured name is a plain date
// math expression it is URI-encoded before being sent in the API request path so the Go
// HTTP client does not rewrite the angle brackets or braces.
func PutIndex(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, index *models.Index, params *models.PutIndexParams) (string, fwdiags.Diagnostics) {
	esClient, err := apiClient.GetESClient()
	if err != nil {
		return "", diagutil.FrameworkDiagFromError(err)
	}

	indexBytes, err := json.Marshal(index)
	if err != nil {
		return "", diagutil.FrameworkDiagFromError(err)
	}

	indexAPIName := index.Name
	if DateMathIndexNameRe.MatchString(index.Name) {
		indexAPIName = encodeDateMathIndexName(index.Name)
	}

	opts := []func(*esapi.IndicesCreateRequest){
		esClient.Indices.Create.WithBody(bytes.NewReader(indexBytes)),
		esClient.Indices.Create.WithContext(ctx),
		esClient.Indices.Create.WithWaitForActiveShards(params.WaitForActiveShards),
	}
	if params.MasterTimeout > 0 {
		opts = append(opts, esClient.Indices.Create.WithMasterTimeout(params.MasterTimeout))
	}
	if params.Timeout > 0 {
		opts = append(opts, esClient.Indices.Create.WithTimeout(params.Timeout))
	}
	res, err := esClient.Indices.Create(indexAPIName, opts...)
	if err != nil {
		return "", diagutil.FrameworkDiagFromError(err)
	}
	defer res.Body.Close()
	if sdkDiags := diagutil.CheckError(res, fmt.Sprintf("Unable to create index: %s", index.Name)); sdkDiags.HasError() {
		return "", diagutil.FrameworkDiagsFromSDK(sdkDiags)
	}
	var response struct {
		Index string `json:"index"`
	}
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return "", diagutil.FrameworkDiagFromError(err)
	}
	concreteName := response.Index
	if concreteName == "" {
		concreteName = index.Name
	}
	return concreteName, nil
}

func DeleteIndex(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, name string) fwdiags.Diagnostics {
	typedClient, err := apiClient.GetESTypedClient()
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}
	_, err = typedClient.Indices.Delete(name).Do(ctx)
	if err != nil {
		var esErr *types.ElasticsearchError
		if errors.As(err, &esErr) && esErr.Status == 404 {
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
	typedClient, err := apiClient.GetESTypedClient()
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}
	res, err := typedClient.Indices.Get(name).FlatSettings(true).Do(ctx)
	if err != nil {
		var esErr *types.ElasticsearchError
		if errors.As(err, &esErr) && esErr.Status == 404 {
			return nil, nil
		}
		return nil, diagutil.FrameworkDiagFromError(err)
	}
	return res, nil
}

func DeleteIndexAlias(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, index string, aliases []string) fwdiags.Diagnostics {
	typedClient, err := apiClient.GetESTypedClient()
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}
	_, err = typedClient.Indices.DeleteAlias(index, strings.Join(aliases, ",")).Do(ctx)
	if err != nil {
		var esErr *types.ElasticsearchError
		if errors.As(err, &esErr) && esErr.Status == 404 {
			return nil
		}
		return diagutil.FrameworkDiagFromError(err)
	}
	return nil
}

func UpdateIndexAlias(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, index string, alias *models.IndexAlias) fwdiags.Diagnostics {
	aliasBytes, err := json.Marshal(alias)
	if err != nil {
		return fwdiags.Diagnostics{fwdiags.NewErrorDiagnostic(err.Error(), err.Error())}
	}
	typedClient, err := apiClient.GetESTypedClient()
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}
	_, err = typedClient.Indices.PutAlias(index, alias.Name).Raw(bytes.NewReader(aliasBytes)).Do(ctx)
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}
	return nil
}

func UpdateIndexSettings(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, index string, settings map[string]any) fwdiags.Diagnostics {
	settingsBytes, err := json.Marshal(settings)
	if err != nil {
		return fwdiags.Diagnostics{fwdiags.NewErrorDiagnostic(err.Error(), err.Error())}
	}
	typedClient, err := apiClient.GetESTypedClient()
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}
	_, err = typedClient.Indices.PutSettings().Indices(index).Raw(bytes.NewReader(settingsBytes)).Do(ctx)
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}
	return nil
}

func UpdateIndexMappings(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, index, mappings string) fwdiags.Diagnostics {
	typedClient, err := apiClient.GetESTypedClient()
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}
	_, err = typedClient.Indices.PutMapping(index).Raw(strings.NewReader(mappings)).Do(ctx)
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}
	return nil
}

func PutDataStream(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, dataStreamName string) sdkdiag.Diagnostics {
	typedClient, err := apiClient.GetESTypedClient()
	if err != nil {
		return sdkdiag.FromErr(err)
	}
	_, err = typedClient.Indices.CreateDataStream(dataStreamName).Do(ctx)
	if err != nil {
		return sdkdiag.FromErr(err)
	}
	return nil
}

func GetDataStream(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, dataStreamName string) (*types.DataStream, sdkdiag.Diagnostics) {
	typedClient, err := apiClient.GetESTypedClient()
	if err != nil {
		return nil, sdkdiag.FromErr(err)
	}
	res, err := typedClient.Indices.GetDataStream().Name(dataStreamName).Do(ctx)
	if err != nil {
		var esErr *types.ElasticsearchError
		if errors.As(err, &esErr) && esErr.Status == 404 {
			return nil, nil
		}
		return nil, sdkdiag.FromErr(err)
	}
	if len(res.DataStreams) == 0 {
		return nil, nil
	}
	ds := res.DataStreams[0]
	return &ds, nil
}

func DeleteDataStream(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, dataStreamName string) sdkdiag.Diagnostics {
	typedClient, err := apiClient.GetESTypedClient()
	if err != nil {
		return sdkdiag.FromErr(err)
	}
	_, err = typedClient.Indices.DeleteDataStream(dataStreamName).Do(ctx)
	if err != nil {
		var esErr *types.ElasticsearchError
		if errors.As(err, &esErr) && esErr.Status == 404 {
			return nil
		}
		return sdkdiag.FromErr(err)
	}
	return nil
}

func PutDataStreamLifecycle(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, dataStreamName string, expandWildcards string, lifecycle models.LifecycleSettings) fwdiags.Diagnostics {
	typedClient, err := apiClient.GetESTypedClient()
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}

	reqBody := map[string]any{}
	if lifecycle.DataRetention != "" {
		reqBody["data_retention"] = lifecycle.DataRetention
	}
	if lifecycle.Enabled {
		reqBody["enabled"] = lifecycle.Enabled
	}
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

func GetDataStreamLifecycle(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, dataStreamName string, expandWildcards string) (*getdatalifecycle.Response, fwdiags.Diagnostics) {
	esClient, err := apiClient.GetESClient()
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}

	opts := []func(*esapi.IndicesGetDataLifecycleRequest){}
	if expandWildcards != "" {
		opts = append(opts, esClient.Indices.GetDataLifecycle.WithExpandWildcards(expandWildcards))
	}

	opts = append(opts, esClient.Indices.GetDataLifecycle.WithContext(ctx))
	res, err := esClient.Indices.GetDataLifecycle([]string{dataStreamName}, opts...)
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if sdkDiags := diagutil.CheckError(res, fmt.Sprintf("Unable to get requested DataStreamLifecycle: %s", dataStreamName)); sdkDiags.HasError() {
		return nil, diagutil.FrameworkDiagsFromSDK(sdkDiags)
	}

	var rawResp struct {
		DataStreams []struct {
			Lifecycle struct {
				DataRetention any `json:"data_retention"`
				Downsampling  []struct {
					After         string `json:"after"`
					FixedInterval string `json:"fixed_interval"`
				} `json:"downsampling"`
				Enabled *bool `json:"enabled"`
			} `json:"lifecycle"`
			Name string `json:"name"`
		} `json:"data_streams"`
	}
	if err := json.NewDecoder(res.Body).Decode(&rawResp); err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}

	typedResp := &getdatalifecycle.Response{}
	for _, ds := range rawResp.DataStreams {
		lifecycle := &types.DataStreamLifecycleWithRollover{}
		if v, ok := ds.Lifecycle.DataRetention.(string); ok && v != "" {
			lifecycle.DataRetention = types.Duration(v)
		}
		if ds.Lifecycle.Enabled != nil {
			lifecycle.Enabled = ds.Lifecycle.Enabled
		}
		for _, d := range ds.Lifecycle.Downsampling {
			lifecycle.Downsampling = append(lifecycle.Downsampling, types.DownsamplingRound{
				After: types.Duration(d.After),
				Config: types.DownsampleConfig{
					FixedInterval: d.FixedInterval,
				},
			})
		}
		typedResp.DataStreams = append(typedResp.DataStreams, types.DataStreamWithLifecycle{
			Lifecycle: lifecycle,
			Name:      ds.Name,
		})
	}

	return typedResp, nil
}

func DeleteDataStreamLifecycle(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, dataStreamName string, expandWildcards string) fwdiags.Diagnostics {
	typedClient, err := apiClient.GetESTypedClient()
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}
	builder := typedClient.Indices.DeleteDataLifecycle(dataStreamName)
	if expandWildcards != "" {
		builder = builder.ExpandWildcards(expandwildcard.ExpandWildcard{Name: expandWildcards})
	}
	_, err = builder.Do(ctx)
	if err != nil {
		var esErr *types.ElasticsearchError
		if errors.As(err, &esErr) && esErr.Status == 404 {
			return nil
		}
		return diagutil.FrameworkDiagFromError(err)
	}
	return nil
}

func GetAlias(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, aliasName string) (map[string]types.IndexAliases, fwdiags.Diagnostics) {
	typedClient, err := apiClient.GetESTypedClient()
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}
	res, err := typedClient.Indices.GetAlias().Name(aliasName).Do(ctx)
	if err != nil {
		var esErr *types.ElasticsearchError
		if errors.As(err, &esErr) && esErr.Status == 404 {
			return nil, nil
		}
		return nil, diagutil.FrameworkDiagFromError(err)
	}
	return res, nil
}

// AliasAction represents a single action in an atomic alias update operation
type AliasAction struct {
	Type          string
	Index         string
	Alias         string
	IsWriteIndex  bool
	Filter        map[string]any
	IndexRouting  string
	IsHidden      bool
	Routing       string
	SearchRouting string
}

// UpdateAliasesAtomic performs atomic alias updates using multiple actions
func UpdateAliasesAtomic(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, actions []AliasAction) fwdiags.Diagnostics {
	var aliasActions []map[string]any

	for _, action := range actions {
		switch action.Type {
		case "remove":
			aliasActions = append(aliasActions, map[string]any{
				"remove": map[string]any{
					"index": action.Index,
					"alias": action.Alias,
				},
			})
		case "add":
			addDetails := map[string]any{
				"index": action.Index,
				"alias": action.Alias,
			}

			if action.IsWriteIndex {
				addDetails["is_write_index"] = true
			}
			if action.Filter != nil {
				addDetails["filter"] = action.Filter
			}
			if action.IndexRouting != "" {
				addDetails["index_routing"] = action.IndexRouting
			}
			if action.SearchRouting != "" {
				addDetails["search_routing"] = action.SearchRouting
			}
			if action.Routing != "" {
				addDetails["routing"] = action.Routing
			}
			if action.IsHidden {
				addDetails["is_hidden"] = action.IsHidden
			}

			aliasActions = append(aliasActions, map[string]any{
				"add": addDetails,
			})
		}
	}

	requestBody := map[string]any{
		"actions": aliasActions,
	}

	aliasBytes, err := json.Marshal(requestBody)
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}

	typedClient, err := apiClient.GetESTypedClient()
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}
	_, err = typedClient.Indices.UpdateAliases().Raw(bytes.NewReader(aliasBytes)).Do(ctx)
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}
	return nil
}

func PutIngestPipeline(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, name string, pipeline map[string]any) sdkdiag.Diagnostics {
	pipelineBytes, err := json.Marshal(pipeline)
	if err != nil {
		return sdkdiag.FromErr(err)
	}
	typedClient, err := apiClient.GetESTypedClient()
	if err != nil {
		return sdkdiag.FromErr(err)
	}
	_, err = typedClient.Ingest.PutPipeline(name).Raw(bytes.NewReader(pipelineBytes)).Do(ctx)
	if err != nil {
		return sdkdiag.FromErr(err)
	}
	return nil
}

func GetIngestPipeline(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, name string) (*types.IngestPipeline, sdkdiag.Diagnostics) {
	typedClient, err := apiClient.GetESTypedClient()
	if err != nil {
		return nil, sdkdiag.FromErr(err)
	}
	res, err := typedClient.Ingest.GetPipeline().Id(name).Do(ctx)
	if err != nil {
		var esErr *types.ElasticsearchError
		if errors.As(err, &esErr) && esErr.Status == 404 {
			return nil, nil
		}
		return nil, sdkdiag.FromErr(err)
	}
	if pipeline, ok := res[name]; ok {
		return &pipeline, nil
	}
	return nil, sdkdiag.Diagnostics{
		sdkdiag.Diagnostic{
			Severity: sdkdiag.Error,
			Summary:  "Unable to find ingest pipeline",
			Detail:   fmt.Sprintf(`Unable to find "%s" ingest pipeline in the cluster`, name),
		},
	}
}

func DeleteIngestPipeline(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, name string) sdkdiag.Diagnostics {
	typedClient, err := apiClient.GetESTypedClient()
	if err != nil {
		return sdkdiag.FromErr(err)
	}
	_, err = typedClient.Ingest.DeletePipeline(name).Do(ctx)
	if err != nil {
		var esErr *types.ElasticsearchError
		if errors.As(err, &esErr) && esErr.Status == 404 {
			return nil
		}
		return sdkdiag.FromErr(err)
	}
	return nil
}

// NormalizeQueryFilter recursively compacts expanded single-key query values
// produced by the typed client back to their shorthand form.
// For example: {"term":{"field":{"value":"x"}}} → {"term":{"field":"x"}}
func NormalizeQueryFilter(v any) any {
	switch val := v.(type) {
	case map[string]any:
		// If this map has exactly one key "value" with a scalar value, compact it.
		if len(val) == 1 {
			if inner, ok := val["value"]; ok {
				switch inner.(type) {
				case string, float64, bool, int, int64:
					return inner
				}
			}
		}
		out := make(map[string]any, len(val))
		for k, vv := range val {
			out[k] = NormalizeQueryFilter(vv)
		}
		return out
	case []any:
		out := make([]any, len(val))
		for i, vv := range val {
			out[i] = NormalizeQueryFilter(vv)
		}
		return out
	default:
		return v
	}
}

// IndexStateToModel converts a typed IndexState into the legacy models.Index shape
// used by the index resource and indices data source.
func IndexStateToModel(state types.IndexState) models.Index {
	aliases := make(map[string]models.IndexAlias, len(state.Aliases))
	for name, alias := range state.Aliases {
		ia := models.IndexAlias{Name: name}
		if alias.Filter != nil {
			filterBytes, _ := json.Marshal(alias.Filter)
			var filterMap map[string]any
			_ = json.Unmarshal(filterBytes, &filterMap)
			ia.Filter = NormalizeQueryFilter(filterMap).(map[string]any)
		}
		if alias.IndexRouting != nil {
			ia.IndexRouting = *alias.IndexRouting
		}
		if alias.IsHidden != nil {
			ia.IsHidden = *alias.IsHidden
		}
		if alias.IsWriteIndex != nil {
			ia.IsWriteIndex = *alias.IsWriteIndex
		}
		if alias.Routing != nil {
			ia.Routing = *alias.Routing
		}
		if alias.SearchRouting != nil {
			ia.SearchRouting = *alias.SearchRouting
		}
		aliases[name] = ia
	}

	var mappings map[string]any
	if state.Mappings != nil {
		mappingBytes, _ := json.Marshal(state.Mappings)
		_ = json.Unmarshal(mappingBytes, &mappings)
	}

	var settings map[string]any
	if state.Settings != nil {
		settingsBytes, _ := json.Marshal(state.Settings)
		_ = json.Unmarshal(settingsBytes, &settings)
	}

	return models.Index{
		Aliases:  aliases,
		Mappings: mappings,
		Settings: settings,
	}
}
