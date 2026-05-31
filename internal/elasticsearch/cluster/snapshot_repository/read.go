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

package snapshot_repository

import (
	"context"
	"fmt"

	esclients "github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func readSnapshotRepository(ctx context.Context, client *esclients.ElasticsearchScopedClient, resourceID string, state Data) (Data, bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	repo, repoDiags := elasticsearch.GetSnapshotRepository(ctx, client, resourceID)
	diags.Append(repoDiags...)
	if diags.HasError() {
		return state, false, diags
	}

	if repo == nil {
		tflog.Warn(ctx, fmt.Sprintf(`Snapshot repository "%s" not found, removing from state`, resourceID))
		return state, false, diags
	}

	data := state
	data.Name = types.StringValue(resourceID)

	// Clear all type blocks then populate the correct one
	data.Fs = types.ObjectNull(fsAttrTypes())
	data.URL = types.ObjectNull(urlAttrTypes())
	data.Gcs = types.ObjectNull(gcsAttrTypes())
	data.Azure = types.ObjectNull(azureAttrTypes())
	data.S3 = types.ObjectNull(s3AttrTypes())
	data.Hdfs = types.ObjectNull(hdfsAttrTypes())

	switch repo.Type {
	case repoTypeFS:
		fs, fsDiags := settingsToFs(ctx, repo, state)
		diags.Append(fsDiags...)
		if diags.HasError() {
			return state, false, diags
		}
		data.Fs = fs
	case repoTypeURL:
		u, uDiags := settingsToURL(ctx, repo, state)
		diags.Append(uDiags...)
		if diags.HasError() {
			return state, false, diags
		}
		data.URL = u
	case repoTypeGCS:
		gcs, gcsDiags := settingsToGcs(ctx, repo)
		diags.Append(gcsDiags...)
		if diags.HasError() {
			return state, false, diags
		}
		data.Gcs = gcs
	case repoTypeAzure:
		azure, azureDiags := settingsToAzure(ctx, repo)
		diags.Append(azureDiags...)
		if diags.HasError() {
			return state, false, diags
		}
		data.Azure = azure
	case repoTypeS3:
		s3, s3Diags := settingsToS3(ctx, repo, state)
		diags.Append(s3Diags...)
		if diags.HasError() {
			return state, false, diags
		}
		data.S3 = s3
	case repoTypeHDFS:
		hdfs, hdfsDiags := settingsToHdfs(ctx, repo)
		diags.Append(hdfsDiags...)
		if diags.HasError() {
			return state, false, diags
		}
		data.Hdfs = hdfs
	default:
		diags.AddError(
			"Unsupported snapshot repository type",
			fmt.Sprintf("The type %q returned by the API is not supported.", repo.Type),
		)
		return state, false, diags
	}

	return data, true, diags
}

func settingsToFs(ctx context.Context, repo *elasticsearch.SnapshotRepositoryInfo, state Data) (types.Object, diag.Diagnostics) {
	s := repo.Settings

	// Try to inherit compress from state if API does not return it
	compressFallback := true
	if !state.Fs.IsNull() && !state.Fs.IsUnknown() {
		var stateFs FsSettings
		if diags := state.Fs.As(ctx, &stateFs, basetypes.ObjectAsOptions{}); !diags.HasError() {
			compressFallback = stateFs.Compress.ValueBool()
		}
	}

	fs := FsSettings{
		CommonSettings: CommonSettings{
			ChunkSize:              StrSettingNull(s, settingChunkSize),
			Compress:               types.BoolValue(BoolSetting(s, settingCompress, compressFallback)),
			MaxSnapshotBytesPerSec: StrSettingNull(s, settingMaxSnapshotBytesPerSec),
			MaxRestoreBytesPerSec:  StrSettingNull(s, settingMaxRestoreBytesPerSec),
			Readonly:               types.BoolValue(BoolSetting(s, settingReadonly, false)),
		},
		CommonStdSettings: CommonStdSettings{
			MaxNumberOfSnapshots: types.Int64Value(Int64Setting(s, settingMaxNumberOfSnapshots, 500)),
		},
		Location: types.StringValue(StrSetting(s, settingLocation)),
	}
	return types.ObjectValueFrom(ctx, fsAttrTypes(), fs)
}

func settingsToURL(ctx context.Context, repo *elasticsearch.SnapshotRepositoryInfo, state Data) (types.Object, diag.Diagnostics) {
	s := repo.Settings

	compressFallback := true
	if !state.URL.IsNull() && !state.URL.IsUnknown() {
		var stateURL URLSettings
		if diags := state.URL.As(ctx, &stateURL, basetypes.ObjectAsOptions{}); !diags.HasError() {
			compressFallback = stateURL.Compress.ValueBool()
		}
	}

	u := URLSettings{
		CommonSettings: CommonSettings{
			ChunkSize:              StrSettingNull(s, settingChunkSize),
			Compress:               types.BoolValue(BoolSetting(s, settingCompress, compressFallback)),
			MaxSnapshotBytesPerSec: StrSettingNull(s, settingMaxSnapshotBytesPerSec),
			MaxRestoreBytesPerSec:  StrSettingNull(s, settingMaxRestoreBytesPerSec),
			Readonly:               types.BoolValue(BoolSetting(s, settingReadonly, false)),
		},
		CommonStdSettings: CommonStdSettings{
			MaxNumberOfSnapshots: types.Int64Value(Int64Setting(s, settingMaxNumberOfSnapshots, 500)),
		},
		URL:               types.StringValue(StrSetting(s, settingURL)),
		HTTPMaxRetries:    types.Int64Value(Int64Setting(s, settingHTTPMaxRetries, 5)),
		HTTPSocketTimeout: StrSettingNull(s, settingHTTPSocketTimeout),
	}
	return types.ObjectValueFrom(ctx, urlAttrTypes(), u)
}

func settingsToGcs(ctx context.Context, repo *elasticsearch.SnapshotRepositoryInfo) (types.Object, diag.Diagnostics) {
	s := repo.Settings
	gcs := GcsSettings{
		CommonSettings: CommonSettings{
			ChunkSize:              StrSettingNull(s, settingChunkSize),
			Compress:               types.BoolValue(BoolSetting(s, settingCompress, true)),
			MaxSnapshotBytesPerSec: StrSettingNull(s, settingMaxSnapshotBytesPerSec),
			MaxRestoreBytesPerSec:  StrSettingNull(s, settingMaxRestoreBytesPerSec),
			Readonly:               types.BoolValue(BoolSetting(s, settingReadonly, false)),
		},
		Bucket:   types.StringValue(StrSetting(s, settingBucket)),
		Client:   StrSettingNull(s, settingClient),
		BasePath: StrSettingNull(s, settingBasePath),
	}
	return types.ObjectValueFrom(ctx, gcsAttrTypes(), gcs)
}

func settingsToAzure(ctx context.Context, repo *elasticsearch.SnapshotRepositoryInfo) (types.Object, diag.Diagnostics) {
	s := repo.Settings
	azure := AzureSettings{
		CommonSettings: CommonSettings{
			ChunkSize:              StrSettingNull(s, settingChunkSize),
			Compress:               types.BoolValue(BoolSetting(s, settingCompress, true)),
			MaxSnapshotBytesPerSec: StrSettingNull(s, settingMaxSnapshotBytesPerSec),
			MaxRestoreBytesPerSec:  StrSettingNull(s, settingMaxRestoreBytesPerSec),
			Readonly:               types.BoolValue(BoolSetting(s, settingReadonly, false)),
		},
		Container:    types.StringValue(StrSetting(s, settingContainer)),
		Client:       StrSettingNull(s, settingClient),
		BasePath:     StrSettingNull(s, settingBasePath),
		LocationMode: StrSettingNull(s, settingLocationMode),
	}
	return types.ObjectValueFrom(ctx, azureAttrTypes(), azure)
}

func settingsToS3(ctx context.Context, repo *elasticsearch.SnapshotRepositoryInfo, state Data) (types.Object, diag.Diagnostics) {
	s := repo.Settings

	endpointFallback := types.StringNull()
	pathStyleAccessFallback := false
	if !state.S3.IsNull() && !state.S3.IsUnknown() {
		var stateS3 S3Settings
		if diags := state.S3.As(ctx, &stateS3, basetypes.ObjectAsOptions{}); !diags.HasError() {
			endpointFallback = stateS3.Endpoint
			if !stateS3.PathStyleAccess.IsNull() && !stateS3.PathStyleAccess.IsUnknown() {
				pathStyleAccessFallback = stateS3.PathStyleAccess.ValueBool()
			}
		}
	}

	// The Elasticsearch GET response may not echo endpoint and path_style_access
	// (the typed S3RepositorySettings struct also omits both fields). Whether
	// Elasticsearch returns them via the raw settings overlay is version-dependent
	// and difficult to determine empirically once read-side inheritance is in place.
	// We therefore inherit both values from the prior state when the GET response
	// omits them, mirroring the compressFallback pattern in settingsToFs and
	// settingsToURL. The API value wins when present.
	var endpoint types.String
	if endpointStr := StrSetting(s, settingEndpoint); endpointStr != "" {
		endpoint = StrSettingNull(s, settingEndpoint)
	} else if !endpointFallback.IsNull() {
		endpoint = endpointFallback
	} else {
		endpoint = types.StringNull()
	}

	s3 := S3Settings{
		CommonSettings: CommonSettings{
			ChunkSize:              StrSettingNull(s, settingChunkSize),
			Compress:               types.BoolValue(BoolSetting(s, settingCompress, true)),
			MaxSnapshotBytesPerSec: StrSettingNull(s, settingMaxSnapshotBytesPerSec),
			MaxRestoreBytesPerSec:  StrSettingNull(s, settingMaxRestoreBytesPerSec),
			Readonly:               types.BoolValue(BoolSetting(s, settingReadonly, false)),
		},
		Bucket:               types.StringValue(StrSetting(s, settingBucket)),
		Endpoint:             endpoint,
		Client:               StrSettingNull(s, settingClient),
		BasePath:             StrSettingNull(s, settingBasePath),
		ServerSideEncryption: types.BoolValue(BoolSetting(s, settingServerSideEncryption, false)),
		BufferSize:           StrSettingNull(s, settingBufferSize),
		CannedACL:            StrSettingNull(s, settingCannedACL),
		StorageClass:         StrSettingNull(s, settingStorageClass),
		PathStyleAccess:      types.BoolValue(BoolSetting(s, settingPathStyleAccess, pathStyleAccessFallback)),
	}
	return types.ObjectValueFrom(ctx, s3AttrTypes(), s3)
}

func settingsToHdfs(ctx context.Context, repo *elasticsearch.SnapshotRepositoryInfo) (types.Object, diag.Diagnostics) {
	s := repo.Settings
	hdfs := HdfsSettings{
		CommonSettings: CommonSettings{
			ChunkSize:              StrSettingNull(s, settingChunkSize),
			Compress:               types.BoolValue(BoolSetting(s, settingCompress, true)),
			MaxSnapshotBytesPerSec: StrSettingNull(s, settingMaxSnapshotBytesPerSec),
			MaxRestoreBytesPerSec:  StrSettingNull(s, settingMaxRestoreBytesPerSec),
			Readonly:               types.BoolValue(BoolSetting(s, settingReadonly, false)),
		},
		URI:          types.StringValue(StrSetting(s, settingURI)),
		Path:         types.StringValue(StrSetting(s, settingPath)),
		LoadDefaults: types.BoolValue(BoolSetting(s, settingLoadDefaults, true)),
	}
	return types.ObjectValueFrom(ctx, hdfsAttrTypes(), hdfs)
}
