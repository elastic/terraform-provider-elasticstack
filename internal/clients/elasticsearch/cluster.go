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

	"github.com/elastic/go-elasticsearch/v8/typedapi/core/info"
	"github.com/elastic/go-elasticsearch/v8/typedapi/snapshot/getrepository"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	fwdiag "github.com/hashicorp/terraform-plugin-framework/diag"
	sdkdiag "github.com/hashicorp/terraform-plugin-sdk/v2/diag"
)

func GetClusterInfo(ctx context.Context, apiClient *clients.ElasticsearchScopedClient) (*info.Response, sdkdiag.Diagnostics) {
	var diags sdkdiag.Diagnostics
	typedClient, err := apiClient.GetESClient()
	if err != nil {
		return nil, sdkdiag.FromErr(err)
	}
	res, err := typedClient.Core.Info().Do(ctx)
	if err != nil {
		return nil, sdkdiag.FromErr(err)
	}
	return res, diags
}

type SnapshotRepositoryInfo struct {
	Name     string
	Type     string
	Settings map[string]any
}

type SlmPolicy struct {
	Name       string        `json:"name"`
	Schedule   string        `json:"schedule"`
	Repository string        `json:"repository"`
	Config     *SlmConfig    `json:"config"`
	Retention  *SlmRetention `json:"retention"`
}

type SlmConfig struct {
	FeatureStates      []string       `json:"feature_states"`
	ExpandWildcards    string         `json:"expand_wildcards"`
	IgnoreUnavailable  *bool          `json:"ignore_unavailable"`
	IncludeGlobalState *bool          `json:"include_global_state"`
	Indices            []string       `json:"indices"`
	Metadata           types.Metadata `json:"metadata"`
	Partial            *bool          `json:"partial"`
}

type SlmRetention struct {
	ExpireAfter *string `json:"expire_after"`
	MaxCount    *int    `json:"max_count"`
	MinCount    *int    `json:"min_count"`
}

func PutSnapshotRepository(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, name string, repoType string, settings map[string]any, verify bool) sdkdiag.Diagnostics {
	var diags sdkdiag.Diagnostics
	typedClient, err := apiClient.GetESClient()
	if err != nil {
		return sdkdiag.FromErr(err)
	}

	settingsBytes, err := json.Marshal(settings)
	if err != nil {
		return sdkdiag.FromErr(err)
	}

	var repo types.Repository
	switch repoType {
	case "fs":
		var s types.SharedFileSystemRepository
		if err := json.Unmarshal(settingsBytes, &s.Settings); err != nil {
			return sdkdiag.FromErr(err)
		}
		s.Type = "fs"
		repo = s
	case "url":
		var s types.ReadOnlyUrlRepository
		if err := json.Unmarshal(settingsBytes, &s.Settings); err != nil {
			return sdkdiag.FromErr(err)
		}
		s.Type = "url"
		repo = s
	case "s3":
		var s types.S3Repository
		if err := json.Unmarshal(settingsBytes, &s.Settings); err != nil {
			return sdkdiag.FromErr(err)
		}
		s.Type = "s3"
		repo = s
	case "gcs":
		var s types.GcsRepository
		if err := json.Unmarshal(settingsBytes, &s.Settings); err != nil {
			return sdkdiag.FromErr(err)
		}
		s.Type = "gcs"
		repo = s
	case "azure":
		var s types.AzureRepository
		if err := json.Unmarshal(settingsBytes, &s.Settings); err != nil {
			return sdkdiag.FromErr(err)
		}
		s.Type = "azure"
		repo = s
	case "hdfs":
		// The go-elasticsearch Typed API's types.Repository union does not include
		// HdfsRepository, and the getrepository Response.UnmarshalJSON only handles
		// azure, gcs, s3, fs, url, source. HDFS snapshots require the repository-hdfs
		// plugin, which is outside the core typed API spec. Therefore HDFS must be
		// sent/received as raw JSON to preserve backward compatibility.
		body := map[string]any{
			"type":     repoType,
			"settings": settings,
		}
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			return sdkdiag.FromErr(err)
		}
		req := typedClient.Snapshot.CreateRepository(name).Raw(bytes.NewReader(bodyBytes)).Verify(verify)
		_, err = req.Do(ctx)
		if err != nil {
			return sdkdiag.FromErr(err)
		}
		return diags
	case "source":
		var s types.SourceOnlyRepository
		if err := json.Unmarshal(settingsBytes, &s.Settings); err != nil {
			return sdkdiag.FromErr(err)
		}
		s.Type = "source"
		repo = s
	default:
		return sdkdiag.Errorf("unsupported snapshot repository type: %s", repoType)
	}

	_, err = typedClient.Snapshot.CreateRepository(name).Request(&repo).Verify(verify).Do(ctx)
	if err != nil {
		return sdkdiag.FromErr(err)
	}
	return diags
}

func GetSnapshotRepository(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, name string) (*SnapshotRepositoryInfo, sdkdiag.Diagnostics) {
	var diags sdkdiag.Diagnostics
	typedClient, err := apiClient.GetESClient()
	if err != nil {
		return nil, sdkdiag.FromErr(err)
	}
	res, err := typedClient.Snapshot.GetRepository().Repository(name).Perform(ctx)
	if err != nil {
		return nil, sdkdiag.FromErr(err)
	}
	defer res.Body.Close()
	if res.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if res.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(res.Body)
		return nil, sdkdiag.FromErr(fmt.Errorf("unexpected status code %d from snapshot repository API: %s", res.StatusCode, string(bodyBytes)))
	}

	bodyBytes, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, sdkdiag.FromErr(fmt.Errorf("failed to read snapshot repository response: %w", err))
	}

	// Try typed parsing first for known types.
	typedResp := getrepository.NewResponse()
	if err := json.Unmarshal(bodyBytes, &typedResp); err == nil {
		if repo, ok := typedResp[name]; ok {
			info, err := extractSnapshotRepositoryInfo(repo)
			if err != nil {
				return nil, sdkdiag.FromErr(err)
			}
			info.Name = name

			// Overlay raw settings to preserve fields omitted by typed structs
			// (e.g. readonly on ReadOnlyUrlRepository).
			var rawResp map[string]struct {
				Type     string         `json:"type"`
				Settings map[string]any `json:"settings"`
			}
			if err := json.Unmarshal(bodyBytes, &rawResp); err == nil {
				if rawRepo, ok := rawResp[name]; ok {
					info.Settings = rawRepo.Settings
				}
			}

			return info, diags
		}
	}

	// Fall back to raw JSON parsing for plugin-backed repository types (e.g. hdfs)
	// that are not part of the go-elasticsearch Typed API types.Repository union.
	var rawResp map[string]struct {
		Type     string         `json:"type"`
		Settings map[string]any `json:"settings"`
	}
	if err := json.Unmarshal(bodyBytes, &rawResp); err == nil {
		if repo, ok := rawResp[name]; ok {
			return &SnapshotRepositoryInfo{
				Name:     name,
				Type:     repo.Type,
				Settings: repo.Settings,
			}, diags
		}
	}

	diags = append(diags, sdkdiag.Diagnostic{
		Severity: sdkdiag.Error,
		Summary:  "Unable to find requested repository",
		Detail:   fmt.Sprintf(`Repository "%s" is missing in the ES API response`, name),
	})
	return nil, diags
}

func extractSnapshotRepositoryInfo(repo types.Repository) (*SnapshotRepositoryInfo, error) {
	var repoType string
	var settings any

	switch r := repo.(type) {
	case types.SharedFileSystemRepository:
		repoType = r.Type
		settings = r.Settings
	case *types.SharedFileSystemRepository:
		repoType = r.Type
		settings = r.Settings
	case types.ReadOnlyUrlRepository:
		repoType = r.Type
		settings = r.Settings
	case *types.ReadOnlyUrlRepository:
		repoType = r.Type
		settings = r.Settings
	case types.S3Repository:
		repoType = r.Type
		settings = r.Settings
	case *types.S3Repository:
		repoType = r.Type
		settings = r.Settings
	case types.GcsRepository:
		repoType = r.Type
		settings = r.Settings
	case *types.GcsRepository:
		repoType = r.Type
		settings = r.Settings
	case types.AzureRepository:
		repoType = r.Type
		settings = r.Settings
	case *types.AzureRepository:
		repoType = r.Type
		settings = r.Settings
	case types.SourceOnlyRepository:
		repoType = r.Type
		settings = r.Settings
	case *types.SourceOnlyRepository:
		repoType = r.Type
		settings = r.Settings
	default:
		return nil, fmt.Errorf("unsupported snapshot repository type: %T", repo)
	}

	settingsMap := make(map[string]any)
	settingsBytes, err := json.Marshal(settings)
	if err != nil {
		return nil, err
	}
	if !bytes.Equal(settingsBytes, []byte("null")) {
		if err := json.Unmarshal(settingsBytes, &settingsMap); err != nil {
			return nil, err
		}
	}

	return &SnapshotRepositoryInfo{
		Type:     repoType,
		Settings: settingsMap,
	}, nil
}

func DeleteSnapshotRepository(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, name string) sdkdiag.Diagnostics {
	var diags sdkdiag.Diagnostics
	typedClient, err := apiClient.GetESClient()
	if err != nil {
		return sdkdiag.FromErr(err)
	}
	_, err = typedClient.Snapshot.DeleteRepository(name).Do(ctx)
	if err != nil {
		if isNotFoundElasticsearchError(err) {
			return diags
		}
		return sdkdiag.FromErr(err)
	}
	return diags
}

func PutSlm(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, policyID string, slm *SlmPolicy) sdkdiag.Diagnostics {
	var diags sdkdiag.Diagnostics
	typedClient, err := apiClient.GetESClient()
	if err != nil {
		return sdkdiag.FromErr(err)
	}

	req := typedClient.Slm.PutLifecycle(policyID)

	// Build request body manually to support expand_wildcards and accurate
	// Retention field omission (types.Retention lacks omitempty on MaxCount/MinCount).
	body := map[string]any{
		"name":       slm.Name,
		"repository": slm.Repository,
		"schedule":   slm.Schedule,
	}
	if slm.Config != nil {
		config := map[string]any{}
		if len(slm.Config.FeatureStates) > 0 {
			config["feature_states"] = slm.Config.FeatureStates
		}
		if slm.Config.IgnoreUnavailable != nil {
			config["ignore_unavailable"] = *slm.Config.IgnoreUnavailable
		}
		if slm.Config.IncludeGlobalState != nil {
			config["include_global_state"] = *slm.Config.IncludeGlobalState
		}
		if len(slm.Config.Indices) > 0 {
			config["indices"] = slm.Config.Indices
		}
		if slm.Config.Metadata != nil {
			meta := make(map[string]any)
			for k, v := range slm.Config.Metadata {
				var val any
				if err := json.Unmarshal(v, &val); err != nil {
					return sdkdiag.FromErr(fmt.Errorf("failed to unmarshal metadata key %q: %w", k, err))
				}
				meta[k] = val
			}
			config["metadata"] = meta
		}
		if slm.Config.Partial != nil {
			config["partial"] = *slm.Config.Partial
		}
		if slm.Config.ExpandWildcards != "" {
			config["expand_wildcards"] = slm.Config.ExpandWildcards
		}
		if len(config) > 0 {
			body["config"] = config
		}
	}
	if slm.Retention != nil {
		retention := map[string]any{}
		if slm.Retention.ExpireAfter != nil {
			retention["expire_after"] = *slm.Retention.ExpireAfter
		}
		if slm.Retention.MaxCount != nil {
			retention["max_count"] = *slm.Retention.MaxCount
		}
		if slm.Retention.MinCount != nil {
			retention["min_count"] = *slm.Retention.MinCount
		}
		if len(retention) > 0 {
			body["retention"] = retention
		}
	}

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return sdkdiag.FromErr(err)
	}
	req.Raw(bytes.NewReader(bodyBytes))

	_, err = req.Do(ctx)
	if err != nil {
		return sdkdiag.FromErr(err)
	}
	return diags
}

func GetSlm(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, slmName string) (*SlmPolicy, sdkdiag.Diagnostics) {
	var diags sdkdiag.Diagnostics
	typedClient, err := apiClient.GetESClient()
	if err != nil {
		return nil, sdkdiag.FromErr(err)
	}
	res, err := typedClient.Slm.GetLifecycle().PolicyId(slmName).Perform(ctx)
	if err != nil {
		return nil, sdkdiag.FromErr(err)
	}
	defer res.Body.Close()
	if res.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	bodyBytes, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, sdkdiag.FromErr(err)
	}

	var rawResp map[string]struct {
		Policy SlmPolicy `json:"policy"`
	}
	if err := json.Unmarshal(bodyBytes, &rawResp); err != nil {
		return nil, sdkdiag.FromErr(err)
	}
	if slm, ok := rawResp[slmName]; ok {
		return &slm.Policy, diags
	}
	diags = append(diags, sdkdiag.Diagnostic{
		Severity: sdkdiag.Error,
		Summary:  "Unable to find the SLM policy in the response",
		Detail:   fmt.Sprintf(`Unable to find "%s" policy in the ES API response.`, slmName),
	})
	return nil, diags
}

func DeleteSlm(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, slmName string) sdkdiag.Diagnostics {
	var diags sdkdiag.Diagnostics
	typedClient, err := apiClient.GetESClient()
	if err != nil {
		return sdkdiag.FromErr(err)
	}
	_, err = typedClient.Slm.DeleteLifecycle(slmName).Do(ctx)
	if err != nil {
		if isNotFoundElasticsearchError(err) {
			return diags
		}
		return sdkdiag.FromErr(err)
	}
	return diags
}

func PutSettings(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, settings map[string]any) sdkdiag.Diagnostics {
	var diags sdkdiag.Diagnostics
	typedClient, err := apiClient.GetESClient()
	if err != nil {
		return sdkdiag.FromErr(err)
	}

	req := typedClient.Cluster.PutSettings()

	if persistent, ok := settings["persistent"].(map[string]any); ok {
		raw, err := toRawMessageMap(persistent)
		if err != nil {
			return sdkdiag.FromErr(err)
		}
		req.Persistent(raw)
	}
	if transient, ok := settings["transient"].(map[string]any); ok {
		raw, err := toRawMessageMap(transient)
		if err != nil {
			return sdkdiag.FromErr(err)
		}
		req.Transient(raw)
	}

	_, err = req.Do(ctx)
	if err != nil {
		return sdkdiag.FromErr(err)
	}
	return diags
}

func toRawMessageMap(m map[string]any) (map[string]json.RawMessage, error) {
	result := make(map[string]json.RawMessage, len(m))
	for k, v := range m {
		data, err := json.Marshal(v)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal setting %q: %w", k, err)
		}
		result[k] = data
	}
	return result, nil
}

func GetSettings(ctx context.Context, apiClient *clients.ElasticsearchScopedClient) (map[string]any, sdkdiag.Diagnostics) {
	var diags sdkdiag.Diagnostics
	typedClient, err := apiClient.GetESClient()
	if err != nil {
		return nil, sdkdiag.FromErr(err)
	}
	resp, err := typedClient.Cluster.GetSettings().FlatSettings(true).Do(ctx)
	if err != nil {
		return nil, sdkdiag.FromErr(err)
	}

	result := make(map[string]any)
	result["persistent"], err = flattenRawMessageMap(resp.Persistent)
	if err != nil {
		return nil, sdkdiag.FromErr(err)
	}
	result["transient"], err = flattenRawMessageMap(resp.Transient)
	if err != nil {
		return nil, sdkdiag.FromErr(err)
	}
	result["defaults"], err = flattenRawMessageMap(resp.Defaults)
	if err != nil {
		return nil, sdkdiag.FromErr(err)
	}
	return result, diags
}

func flattenRawMessageMap(m map[string]json.RawMessage) (map[string]any, error) {
	result := make(map[string]any, len(m))
	for k, v := range m {
		var val any
		if err := json.Unmarshal(v, &val); err != nil {
			return nil, fmt.Errorf("failed to unmarshal setting %q: %w", k, err)
		}
		result[k] = val
	}
	return result, nil
}

func GetScript(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, id string) (*types.StoredScript, fwdiag.Diagnostics) {
	typedClient, err := apiClient.GetESClient()
	if err != nil {
		var diags fwdiag.Diagnostics
		diags.AddError("Failed to get ES client", err.Error())
		return nil, diags
	}
	resp, err := typedClient.Core.GetScript(id).Do(ctx)
	if err != nil {
		if isNotFoundElasticsearchError(err) {
			return nil, nil
		}
		var diags fwdiag.Diagnostics
		diags.AddError("Failed to get script", err.Error())
		return nil, diags
	}
	return resp.Script, nil
}

func PutScript(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, id string, context string, script *types.StoredScript, params map[string]any) fwdiag.Diagnostics {
	var diags fwdiag.Diagnostics
	typedClient, err := apiClient.GetESClient()
	if err != nil {
		diags.AddError("Failed to get ES client", err.Error())
		return diags
	}

	req := typedClient.Core.PutScript(id)
	if context != "" {
		req.Context(context)
	}

	// Build request body manually to support params (types.StoredScript lacks Params).
	type storedScriptWithParams struct {
		types.StoredScript
		Params map[string]any `json:"params,omitempty"`
	}
	body := struct {
		Script storedScriptWithParams `json:"script"`
	}{
		Script: storedScriptWithParams{
			StoredScript: *script,
			Params:       params,
		},
	}
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		diags.AddError("Failed to marshal script request", err.Error())
		return diags
	}
	req.Raw(bytes.NewReader(bodyBytes))

	_, err = req.Do(ctx)
	if err != nil {
		diags.AddError("Failed to put script", err.Error())
		return diags
	}
	return diags
}

func DeleteScript(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, id string) fwdiag.Diagnostics {
	var diags fwdiag.Diagnostics
	typedClient, err := apiClient.GetESClient()
	if err != nil {
		diags.AddError("Failed to get ES client", err.Error())
		return diags
	}
	_, err = typedClient.Core.DeleteScript(id).Do(ctx)
	if err != nil {
		if isNotFoundElasticsearchError(err) {
			return diags
		}
		diags.AddError("Failed to delete script", err.Error())
		return diags
	}
	return diags
}
