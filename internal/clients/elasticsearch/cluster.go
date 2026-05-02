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
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/elastic/go-elasticsearch/v8/typedapi/core/info"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	fwdiag "github.com/hashicorp/terraform-plugin-framework/diag"
	sdkdiag "github.com/hashicorp/terraform-plugin-sdk/v2/diag"
)

func GetClusterInfo(ctx context.Context, apiClient *clients.ElasticsearchScopedClient) (*info.Response, sdkdiag.Diagnostics) {
	var diags sdkdiag.Diagnostics
	typedClient, err := apiClient.GetESTypedClient()
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

func PutSnapshotRepository(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, name string, repoType string, settings map[string]any, verify bool) sdkdiag.Diagnostics {
	var diags sdkdiag.Diagnostics
	typedClient, err := apiClient.GetESTypedClient()
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
	typedClient, err := apiClient.GetESTypedClient()
	if err != nil {
		return nil, sdkdiag.FromErr(err)
	}
	resp, err := typedClient.Snapshot.GetRepository().Repository(name).Do(ctx)
	if err != nil {
		var esErr *types.ElasticsearchError
		if errors.As(err, &esErr) && esErr.Status == 404 {
			return nil, nil
		}
		return nil, sdkdiag.FromErr(err)
	}

	repo, ok := resp[name]
	if !ok {
		diags = append(diags, sdkdiag.Diagnostic{
			Severity: sdkdiag.Error,
			Summary:  "Unable to find requested repository",
			Detail:   fmt.Sprintf(`Repository "%s" is missing in the ES API response`, name),
		})
		return nil, diags
	}

	info, err := extractSnapshotRepositoryInfo(repo)
	if err != nil {
		return nil, sdkdiag.FromErr(err)
	}
	info.Name = name
	return info, diags
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
	if settings != nil {
		settingsBytes, err := json.Marshal(settings)
		if err != nil {
			return nil, err
		}
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
	typedClient, err := apiClient.GetESTypedClient()
	if err != nil {
		return sdkdiag.FromErr(err)
	}
	_, err = typedClient.Snapshot.DeleteRepository(name).Do(ctx)
	if err != nil {
		var esErr *types.ElasticsearchError
		if errors.As(err, &esErr) && esErr.Status == 404 {
			return diags
		}
		return sdkdiag.FromErr(err)
	}
	return diags
}

func PutSlm(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, slm *types.SLMPolicy) sdkdiag.Diagnostics {
	var diags sdkdiag.Diagnostics
	typedClient, err := apiClient.GetESTypedClient()
	if err != nil {
		return sdkdiag.FromErr(err)
	}

	req := typedClient.Slm.PutLifecycle(slm.Name)
	if slm.Config != nil {
		req.Config(slm.Config)
	}
	req.Repository(slm.Repository)
	if slm.Retention != nil {
		req.Retention(slm.Retention)
	}
	req.Schedule(slm.Schedule)

	_, err = req.Do(ctx)
	if err != nil {
		return sdkdiag.FromErr(err)
	}
	return diags
}

func GetSlm(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, slmName string) (*types.SLMPolicy, sdkdiag.Diagnostics) {
	var diags sdkdiag.Diagnostics
	typedClient, err := apiClient.GetESTypedClient()
	if err != nil {
		return nil, sdkdiag.FromErr(err)
	}
	resp, err := typedClient.Slm.GetLifecycle().PolicyId(slmName).Do(ctx)
	if err != nil {
		var esErr *types.ElasticsearchError
		if errors.As(err, &esErr) && esErr.Status == 404 {
			return nil, nil
		}
		return nil, sdkdiag.FromErr(err)
	}

	if slm, ok := resp[slmName]; ok {
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
	typedClient, err := apiClient.GetESTypedClient()
	if err != nil {
		return sdkdiag.FromErr(err)
	}
	_, err = typedClient.Slm.DeleteLifecycle(slmName).Do(ctx)
	if err != nil {
		var esErr *types.ElasticsearchError
		if errors.As(err, &esErr) && esErr.Status == 404 {
			return diags
		}
		return sdkdiag.FromErr(err)
	}
	return diags
}

func PutSettings(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, settings map[string]any) sdkdiag.Diagnostics {
	var diags sdkdiag.Diagnostics
	typedClient, err := apiClient.GetESTypedClient()
	if err != nil {
		return sdkdiag.FromErr(err)
	}

	req := typedClient.Cluster.PutSettings()

	if persistent, ok := settings["persistent"].(map[string]any); ok {
		req.Persistent(toRawMessageMap(persistent))
	}
	if transient, ok := settings["transient"].(map[string]any); ok {
		req.Transient(toRawMessageMap(transient))
	}

	_, err = req.Do(ctx)
	if err != nil {
		return sdkdiag.FromErr(err)
	}
	return diags
}

func toRawMessageMap(m map[string]any) map[string]json.RawMessage {
	result := make(map[string]json.RawMessage, len(m))
	for k, v := range m {
		data, err := json.Marshal(v)
		if err == nil {
			result[k] = data
		}
	}
	return result
}

func GetSettings(ctx context.Context, apiClient *clients.ElasticsearchScopedClient) (map[string]any, sdkdiag.Diagnostics) {
	var diags sdkdiag.Diagnostics
	typedClient, err := apiClient.GetESTypedClient()
	if err != nil {
		return nil, sdkdiag.FromErr(err)
	}
	resp, err := typedClient.Cluster.GetSettings().FlatSettings(true).Do(ctx)
	if err != nil {
		return nil, sdkdiag.FromErr(err)
	}

	result := make(map[string]any)
	result["persistent"] = flattenRawMessageMap(resp.Persistent)
	result["transient"] = flattenRawMessageMap(resp.Transient)
	result["defaults"] = flattenRawMessageMap(resp.Defaults)
	return result, diags
}

func flattenRawMessageMap(m map[string]json.RawMessage) map[string]any {
	result := make(map[string]any, len(m))
	for k, v := range m {
		var val any
		if err := json.Unmarshal(v, &val); err == nil {
			result[k] = val
		}
	}
	return result
}

func GetScript(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, id string) (*types.StoredScript, fwdiag.Diagnostics) {
	typedClient, err := apiClient.GetESTypedClient()
	if err != nil {
		var diags fwdiag.Diagnostics
		diags.AddError("Failed to get ES client", err.Error())
		return nil, diags
	}
	resp, err := typedClient.Core.GetScript(id).Do(ctx)
	if err != nil {
		var esErr *types.ElasticsearchError
		if errors.As(err, &esErr) && esErr.Status == 404 {
			return nil, nil
		}
		var diags fwdiag.Diagnostics
		diags.AddError("Failed to get script", err.Error())
		return nil, diags
	}
	return resp.Script, nil
}

func PutScript(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, id string, context string, script *types.StoredScript) fwdiag.Diagnostics {
	var diags fwdiag.Diagnostics
	typedClient, err := apiClient.GetESTypedClient()
	if err != nil {
		diags.AddError("Failed to get ES client", err.Error())
		return diags
	}

	req := typedClient.Core.PutScript(id).Script(script)
	if context != "" {
		req.Context(context)
	}

	_, err = req.Do(ctx)
	if err != nil {
		diags.AddError("Failed to put script", err.Error())
		return diags
	}
	return diags
}

func DeleteScript(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, id string) fwdiag.Diagnostics {
	var diags fwdiag.Diagnostics
	typedClient, err := apiClient.GetESTypedClient()
	if err != nil {
		diags.AddError("Failed to get ES client", err.Error())
		return diags
	}
	_, err = typedClient.Core.DeleteScript(id).Do(ctx)
	if err != nil {
		var esErr *types.ElasticsearchError
		if errors.As(err, &esErr) && esErr.Status == 404 {
			return diags
		}
		diags.AddError("Failed to delete script", err.Error())
		return diags
	}
	return diags
}
