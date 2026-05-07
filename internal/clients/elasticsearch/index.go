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
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/elastic/go-elasticsearch/v8/typedapi/ilm/putlifecycle"
	"github.com/elastic/go-elasticsearch/v8/typedapi/indices/create"
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
func encodeDateMathIndexName(name string) string {
	// url.PathEscape does not encode '/' by default; we need '/' encoded too
	// so the Go HTTP client does not split the path at that point.
	return strings.ReplaceAll(url.PathEscape(name), "/", "%2F")
}

func PutIlm(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, policy *models.Policy) fwdiags.Diagnostics {
	policyBytes, err := json.Marshal(map[string]any{"policy": policy})
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}

	typedClient, err := apiClient.GetESClient()
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
	typedClient, err := apiClient.GetESClient()
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}
	res, err := typedClient.Ilm.GetLifecycle().Policy(policyName).Do(ctx)
	if err != nil {
		if isNotFoundElasticsearchError(err) {
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

// GetIndicesWithILMPolicy returns the names of all indices currently using
// the given ILM policy.
//
// It queries GET /_ilm/policy/<policyName> and reads the
// `<policy>.in_use_by.indices` field, which Elasticsearch maintains per
// policy. This is a single targeted lookup keyed by the policy and avoids
// scanning indices cluster-wide.
//
// The typed client's generated `Lifecycle` struct does not expose
// `in_use_by`, so this function uses Perform to obtain the raw HTTP response
// and decodes the relevant subset of the body itself.
func GetIndicesWithILMPolicy(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, policyName string) ([]string, fwdiags.Diagnostics) {
	typedClient, err := apiClient.GetESClient()
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}

	res, err := typedClient.Ilm.GetLifecycle().Policy(policyName).Perform(ctx)
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if res.StatusCode >= 300 {
		body, _ := io.ReadAll(res.Body)
		return nil, fwdiags.Diagnostics{
			fwdiags.NewErrorDiagnostic(
				"Unable to fetch ILM policy",
				fmt.Sprintf("Elasticsearch returned status %d for GET /_ilm/policy/%s: %s", res.StatusCode, policyName, strings.TrimSpace(string(body))),
			),
		}
	}

	// The response is shaped as:
	//   { "<policy_name>": { "in_use_by": { "indices": [...], ... }, ... } }
	var decoded map[string]struct {
		InUseBy struct {
			Indices []string `json:"indices"`
		} `json:"in_use_by"`
	}
	if err := json.NewDecoder(res.Body).Decode(&decoded); err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}

	entry, ok := decoded[policyName]
	if !ok {
		return nil, nil
	}
	return entry.InUseBy.Indices, nil
}

// ClearILMPolicyFromIndices removes the ILM policy reference from the
// provided indices by setting index.lifecycle.name to null.
// It issues PUT /{indices}/_settings.
func ClearILMPolicyFromIndices(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, indices []string) fwdiags.Diagnostics {
	if len(indices) == 0 {
		return nil
	}

	settingsBytes, err := json.Marshal(map[string]any{"index.lifecycle.name": nil})
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}

	typedClient, err := apiClient.GetESClient()
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}

	_, err = typedClient.Indices.PutSettings().Indices(strings.Join(indices, ",")).Raw(bytes.NewReader(settingsBytes)).Do(ctx)
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}
	return nil
}

func DeleteIlm(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, policyName string) fwdiags.Diagnostics {
	typedClient, err := apiClient.GetESClient()
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}
	_, err = typedClient.Ilm.DeleteLifecycle(policyName).Do(ctx)
	if err != nil {
		if isNotFoundElasticsearchError(err) {
			return nil
		}
		return diagutil.FrameworkDiagFromError(err)
	}
	return nil
}

func PutComponentTemplate(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, template *models.ComponentTemplate) sdkdiag.Diagnostics {
	typedClient, err := apiClient.GetESClient()
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
	typedClient, err := apiClient.GetESClient()
	if err != nil {
		return nil, sdkdiag.FromErr(err)
	}
	res, err := typedClient.Cluster.GetComponentTemplate().Name(templateName).Do(ctx)
	if err != nil {
		if isNotFoundElasticsearchError(err) {
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
	typedClient, err := apiClient.GetESClient()
	if err != nil {
		return sdkdiag.FromErr(err)
	}
	_, err = typedClient.Cluster.DeleteComponentTemplate(templateName).Do(ctx)
	if err != nil {
		if isNotFoundElasticsearchError(err) {
			return nil
		}
		return sdkdiag.FromErr(err)
	}
	return nil
}

func PutIndexTemplate(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, template *models.IndexTemplate) fwdiags.Diagnostics {
	typedClient, err := apiClient.GetESClient()
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
	typedClient, err := apiClient.GetESClient()
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}
	res, err := typedClient.Indices.GetIndexTemplate().Name(templateName).Do(ctx)
	if err != nil {
		if isNotFoundElasticsearchError(err) {
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
	typedClient, err := apiClient.GetESClient()
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}
	_, err = typedClient.Indices.DeleteIndexTemplate(templateName).Do(ctx)
	if err != nil {
		if isNotFoundElasticsearchError(err) {
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
	typedClient, err := apiClient.GetESClient()
	if err != nil {
		return "", diagutil.FrameworkDiagFromError(err)
	}

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
	typedClient, err := apiClient.GetESClient()
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}
	_, err = typedClient.Indices.Delete(name).Do(ctx)
	if err != nil {
		if isNotFoundElasticsearchError(err) {
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
	typedClient, err := apiClient.GetESClient()
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}
	res, err := typedClient.Indices.Get(name).FlatSettings(true).Do(ctx)
	if err != nil {
		if isNotFoundElasticsearchError(err) {
			return nil, nil
		}
		return nil, diagutil.FrameworkDiagFromError(err)
	}
	return res, nil
}

func DeleteIndexAlias(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, index string, aliases []string) fwdiags.Diagnostics {
	typedClient, err := apiClient.GetESClient()
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}
	_, err = typedClient.Indices.DeleteAlias(index, strings.Join(aliases, ",")).Do(ctx)
	if err != nil {
		if isNotFoundElasticsearchError(err) {
			return nil
		}
		return diagutil.FrameworkDiagFromError(err)
	}
	return nil
}

func UpdateIndexAlias(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, index string, alias *models.IndexAlias) fwdiags.Diagnostics {
	aliasBytes, err := json.Marshal(alias)
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}
	typedClient, err := apiClient.GetESClient()
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
		return diagutil.FrameworkDiagFromError(err)
	}
	typedClient, err := apiClient.GetESClient()
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
	typedClient, err := apiClient.GetESClient()
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
	typedClient, err := apiClient.GetESClient()
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
	typedClient, err := apiClient.GetESClient()
	if err != nil {
		return nil, sdkdiag.FromErr(err)
	}
	res, err := typedClient.Indices.GetDataStream().Name(dataStreamName).Do(ctx)
	if err != nil {
		if isNotFoundElasticsearchError(err) {
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
	typedClient, err := apiClient.GetESClient()
	if err != nil {
		return sdkdiag.FromErr(err)
	}
	_, err = typedClient.Indices.DeleteDataStream(dataStreamName).Do(ctx)
	if err != nil {
		if isNotFoundElasticsearchError(err) {
			return nil
		}
		return sdkdiag.FromErr(err)
	}
	return nil
}

func PutDataStreamLifecycle(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, dataStreamName string, expandWildcards string, lifecycle models.LifecycleSettings) fwdiags.Diagnostics {
	typedClient, err := apiClient.GetESClient()
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}

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
	typedClient, err := apiClient.GetESClient()
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}

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
		_, _ = io.Copy(io.Discard, res.Body)
		return nil, nil
	}

	if res.StatusCode >= http.StatusMultipleChoices {
		errorResponse := types.NewElasticsearchError()
		if err := json.NewDecoder(res.Body).Decode(errorResponse); err != nil {
			return nil, diagutil.FrameworkDiagFromError(err)
		}
		if errorResponse.Status == 0 {
			errorResponse.Status = res.StatusCode
		}
		return nil, diagutil.FrameworkDiagFromError(errorResponse)
	}

	var response models.DataStreamLifecycleResponse
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}

	return &response, nil
}

func DeleteDataStreamLifecycle(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, dataStreamName string, expandWildcards string) fwdiags.Diagnostics {
	typedClient, err := apiClient.GetESClient()
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}
	builder := typedClient.Indices.DeleteDataLifecycle(dataStreamName)
	if expandWildcards != "" {
		builder = builder.ExpandWildcards(expandwildcard.ExpandWildcard{Name: expandWildcards})
	}
	_, err = builder.Do(ctx)
	if err != nil {
		if isNotFoundElasticsearchError(err) {
			return nil
		}
		return diagutil.FrameworkDiagFromError(err)
	}
	return nil
}

func GetAlias(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, aliasName string) (map[string]types.IndexAliases, fwdiags.Diagnostics) {
	typedClient, err := apiClient.GetESClient()
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}
	res, err := typedClient.Indices.GetAlias().Name(aliasName).Do(ctx)
	if err != nil {
		if isNotFoundElasticsearchError(err) {
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
	aliasActions := make([]map[string]any, 0, len(actions))

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

	typedClient, err := apiClient.GetESClient()
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
	typedClient, err := apiClient.GetESClient()
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
	typedClient, err := apiClient.GetESClient()
	if err != nil {
		return nil, sdkdiag.FromErr(err)
	}
	res, err := typedClient.Ingest.GetPipeline().Id(name).Do(ctx)
	if err != nil {
		if isNotFoundElasticsearchError(err) {
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
	typedClient, err := apiClient.GetESClient()
	if err != nil {
		return sdkdiag.FromErr(err)
	}
	_, err = typedClient.Ingest.DeletePipeline(name).Do(ctx)
	if err != nil {
		if isNotFoundElasticsearchError(err) {
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
