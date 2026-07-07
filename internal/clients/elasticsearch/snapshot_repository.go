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

	"github.com/elastic/go-elasticsearch/v9/typedapi/snapshot/getrepository"
	"github.com/elastic/go-elasticsearch/v9/typedapi/types"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	fwdiag "github.com/hashicorp/terraform-plugin-framework/diag"
)

type SnapshotRepositoryInfo struct {
	Name     string
	Type     string
	Settings map[string]any
}

func PutSnapshotRepository(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, name string, repoType string, settings map[string]any, verify bool) fwdiag.Diagnostics {
	typedClient := apiClient.GetESClient()

	settingsBytes, err := json.Marshal(settings)
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}

	var repo types.Repository
	switch repoType {
	case "fs":
		var s types.SharedFileSystemRepository
		if err := json.Unmarshal(settingsBytes, &s.Settings); err != nil {
			return diagutil.FrameworkDiagFromError(err)
		}
		s.Type = "fs"
		repo = s
	case "url":
		var s types.ReadOnlyUrlRepository
		if err := json.Unmarshal(settingsBytes, &s.Settings); err != nil {
			return diagutil.FrameworkDiagFromError(err)
		}
		s.Type = "url"
		repo = s
	case "s3":
		// types.S3RepositorySettings from go-elasticsearch omits endpoint and
		// path_style_access, so unmarshaling the settings map silently drops those
		// fields. Send raw JSON instead, mirroring the HDFS bypass below.
		body := map[string]any{
			"type":     repoType,
			"settings": settings,
		}
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			return diagutil.FrameworkDiagFromError(err)
		}
		req := typedClient.Snapshot.CreateRepository(name).Raw(bytes.NewReader(bodyBytes)).Verify(verify)
		_, err = req.Do(ctx)
		if err != nil {
			return diagutil.FrameworkDiagFromError(err)
		}
		return nil
	case "gcs":
		var s types.GcsRepository
		if err := json.Unmarshal(settingsBytes, &s.Settings); err != nil {
			return diagutil.FrameworkDiagFromError(err)
		}
		s.Type = "gcs"
		repo = s
	case "azure":
		var s types.AzureRepository
		if err := json.Unmarshal(settingsBytes, &s.Settings); err != nil {
			return diagutil.FrameworkDiagFromError(err)
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
			return diagutil.FrameworkDiagFromError(err)
		}
		req := typedClient.Snapshot.CreateRepository(name).Raw(bytes.NewReader(bodyBytes)).Verify(verify)
		_, err = req.Do(ctx)
		if err != nil {
			return diagutil.FrameworkDiagFromError(err)
		}
		return nil
	case "source":
		var s types.SourceOnlyRepository
		if err := json.Unmarshal(settingsBytes, &s.Settings); err != nil {
			return diagutil.FrameworkDiagFromError(err)
		}
		s.Type = "source"
		repo = s
	default:
		return fwdiag.Diagnostics{
			fwdiag.NewErrorDiagnostic(fmt.Sprintf("unsupported snapshot repository type: %s", repoType), ""),
		}
	}

	_, err = typedClient.Snapshot.CreateRepository(name).Request(&repo).Verify(verify).Do(ctx)
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}
	return nil
}

func GetSnapshotRepository(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, name string) (*SnapshotRepositoryInfo, fwdiag.Diagnostics) {
	typedClient := apiClient.GetESClient()
	res, err := typedClient.Snapshot.GetRepository().Repository(name).Perform(ctx)
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}
	defer res.Body.Close()
	if res.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if d := diagutil.CheckHTTPErrorFromFW(res, "Unable to get snapshot repository"); d.HasError() {
		return nil, d
	}

	bodyBytes, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(fmt.Errorf("failed to read snapshot repository response: %w", err))
	}

	// Try typed parsing first for known types.
	typedResp := getrepository.NewResponse()
	if err := json.Unmarshal(bodyBytes, &typedResp); err == nil {
		if repo, ok := typedResp[name]; ok {
			info, err := extractSnapshotRepositoryInfo(repo)
			if err != nil {
				return nil, diagutil.FrameworkDiagFromError(err)
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

			return info, nil
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
			}, nil
		}
	}

	return nil, fwdiag.Diagnostics{
		fwdiag.NewErrorDiagnostic(
			"Unable to find requested repository",
			fmt.Sprintf(`Repository "%s" is missing in the ES API response`, name),
		),
	}
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

func DeleteSnapshotRepository(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, name string) fwdiag.Diagnostics {
	typedClient := apiClient.GetESClient()
	_, err := typedClient.Snapshot.DeleteRepository(name).Do(ctx)
	if err != nil {
		if IsNotFoundElasticsearchError(err) {
			return nil
		}
		return diagutil.FrameworkDiagFromError(err)
	}
	return nil
}
