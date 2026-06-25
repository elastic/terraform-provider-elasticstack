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

package repository

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

// strSettingNullWithFallback returns the API value for key if present, otherwise
// the prior state value when it is known. This keeps explicitly configured empty
// strings ("") from drifting to null after apply because the write path omits
// them from the API request.
func strSettingNullWithFallback(settings map[string]any, key string, fallback types.String) types.String {
	v := strSettingNull(settings, key)
	if !v.IsNull() {
		return v
	}
	if !fallback.IsNull() && !fallback.IsUnknown() {
		return fallback
	}
	return types.StringNull()
}

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
		gcs, gcsDiags := settingsToGcs(ctx, repo, state)
		diags.Append(gcsDiags...)
		if diags.HasError() {
			return state, false, diags
		}
		data.Gcs = gcs
	case repoTypeAzure:
		azure, azureDiags := settingsToAzure(ctx, repo, state)
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
		hdfs, hdfsDiags := settingsToHdfs(ctx, repo, state)
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

	var diags diag.Diagnostics
	var priorFs FsSettings
	if !state.Fs.IsNull() && !state.Fs.IsUnknown() {
		diags.Append(state.Fs.As(ctx, &priorFs, basetypes.ObjectAsOptions{})...)
	}

	// Try to inherit compress from state if API does not return it
	compressFallback := true
	if !state.Fs.IsNull() && !state.Fs.IsUnknown() {
		var stateFs FsSettings
		if asDiags := state.Fs.As(ctx, &stateFs, basetypes.ObjectAsOptions{}); !asDiags.HasError() {
			compressFallback = stateFs.Compress.ValueBool()
		}
	}

	fs := FsSettings{
		CommonSettings: CommonSettings{
			ChunkSize:              strSettingNullWithFallback(s, settingChunkSize, priorFs.ChunkSize),
			Compress:               types.BoolValue(boolSetting(s, settingCompress, compressFallback)),
			MaxSnapshotBytesPerSec: strSettingNullWithFallback(s, settingMaxSnapshotBytesPerSec, priorFs.MaxSnapshotBytesPerSec),
			MaxRestoreBytesPerSec:  strSettingNullWithFallback(s, settingMaxRestoreBytesPerSec, priorFs.MaxRestoreBytesPerSec),
			Readonly:               types.BoolValue(boolSetting(s, settingReadonly, false)),
		},
		CommonStdSettings: CommonStdSettings{
			MaxNumberOfSnapshots: types.Int64Value(int64Setting(s, settingMaxNumberOfSnapshots, 500)),
		},
		Location: types.StringValue(strSetting(s, settingLocation)),
	}
	obj, objDiags := types.ObjectValueFrom(ctx, fsAttrTypes(), fs)
	diags.Append(objDiags...)
	return obj, diags
}

func settingsToURL(ctx context.Context, repo *elasticsearch.SnapshotRepositoryInfo, state Data) (types.Object, diag.Diagnostics) {
	s := repo.Settings

	var diags diag.Diagnostics
	var priorURL URLSettings
	if !state.URL.IsNull() && !state.URL.IsUnknown() {
		diags.Append(state.URL.As(ctx, &priorURL, basetypes.ObjectAsOptions{})...)
	}

	compressFallback := true
	if !state.URL.IsNull() && !state.URL.IsUnknown() {
		var stateURL URLSettings
		if asDiags := state.URL.As(ctx, &stateURL, basetypes.ObjectAsOptions{}); !asDiags.HasError() {
			compressFallback = stateURL.Compress.ValueBool()
		}
	}

	u := URLSettings{
		CommonSettings: CommonSettings{
			ChunkSize:              strSettingNullWithFallback(s, settingChunkSize, priorURL.ChunkSize),
			Compress:               types.BoolValue(boolSetting(s, settingCompress, compressFallback)),
			MaxSnapshotBytesPerSec: strSettingNullWithFallback(s, settingMaxSnapshotBytesPerSec, priorURL.MaxSnapshotBytesPerSec),
			MaxRestoreBytesPerSec:  strSettingNullWithFallback(s, settingMaxRestoreBytesPerSec, priorURL.MaxRestoreBytesPerSec),
			Readonly:               types.BoolValue(boolSetting(s, settingReadonly, false)),
		},
		CommonStdSettings: CommonStdSettings{
			MaxNumberOfSnapshots: types.Int64Value(int64Setting(s, settingMaxNumberOfSnapshots, 500)),
		},
		URL:               types.StringValue(strSetting(s, settingURL)),
		HTTPMaxRetries:    types.Int64Value(int64Setting(s, settingHTTPMaxRetries, 5)),
		HTTPSocketTimeout: strSettingNullWithFallback(s, settingHTTPSocketTimeout, priorURL.HTTPSocketTimeout),
	}
	obj, objDiags := types.ObjectValueFrom(ctx, urlAttrTypes(), u)
	diags.Append(objDiags...)
	return obj, diags
}

func settingsToGcs(ctx context.Context, repo *elasticsearch.SnapshotRepositoryInfo, state Data) (types.Object, diag.Diagnostics) {
	s := repo.Settings

	var diags diag.Diagnostics
	var priorGcs GcsSettings
	if !state.Gcs.IsNull() && !state.Gcs.IsUnknown() {
		diags.Append(state.Gcs.As(ctx, &priorGcs, basetypes.ObjectAsOptions{})...)
	}

	gcs := GcsSettings{
		CommonSettings: CommonSettings{
			ChunkSize:              strSettingNullWithFallback(s, settingChunkSize, priorGcs.ChunkSize),
			Compress:               types.BoolValue(boolSetting(s, settingCompress, true)),
			MaxSnapshotBytesPerSec: strSettingNullWithFallback(s, settingMaxSnapshotBytesPerSec, priorGcs.MaxSnapshotBytesPerSec),
			MaxRestoreBytesPerSec:  strSettingNullWithFallback(s, settingMaxRestoreBytesPerSec, priorGcs.MaxRestoreBytesPerSec),
			Readonly:               types.BoolValue(boolSetting(s, settingReadonly, false)),
		},
		Bucket:   types.StringValue(strSetting(s, settingBucket)),
		Client:   strSettingNull(s, settingClient),
		BasePath: strSettingNull(s, settingBasePath),
	}
	obj, objDiags := types.ObjectValueFrom(ctx, gcsAttrTypes(), gcs)
	diags.Append(objDiags...)
	return obj, diags
}

func settingsToAzure(ctx context.Context, repo *elasticsearch.SnapshotRepositoryInfo, state Data) (types.Object, diag.Diagnostics) {
	s := repo.Settings

	var diags diag.Diagnostics
	var priorAzure AzureSettings
	if !state.Azure.IsNull() && !state.Azure.IsUnknown() {
		diags.Append(state.Azure.As(ctx, &priorAzure, basetypes.ObjectAsOptions{})...)
	}

	azure := AzureSettings{
		CommonSettings: CommonSettings{
			ChunkSize:              strSettingNullWithFallback(s, settingChunkSize, priorAzure.ChunkSize),
			Compress:               types.BoolValue(boolSetting(s, settingCompress, true)),
			MaxSnapshotBytesPerSec: strSettingNullWithFallback(s, settingMaxSnapshotBytesPerSec, priorAzure.MaxSnapshotBytesPerSec),
			MaxRestoreBytesPerSec:  strSettingNullWithFallback(s, settingMaxRestoreBytesPerSec, priorAzure.MaxRestoreBytesPerSec),
			Readonly:               types.BoolValue(boolSetting(s, settingReadonly, false)),
		},
		Container:    types.StringValue(strSetting(s, settingContainer)),
		Client:       strSettingNull(s, settingClient),
		BasePath:     strSettingNull(s, settingBasePath),
		LocationMode: strSettingNull(s, settingLocationMode),
	}
	obj, objDiags := types.ObjectValueFrom(ctx, azureAttrTypes(), azure)
	diags.Append(objDiags...)
	return obj, diags
}

func settingsToS3(ctx context.Context, repo *elasticsearch.SnapshotRepositoryInfo, state Data) (types.Object, diag.Diagnostics) {
	s := repo.Settings

	var diags diag.Diagnostics
	var priorS3 S3Settings
	if !state.S3.IsNull() && !state.S3.IsUnknown() {
		diags.Append(state.S3.As(ctx, &priorS3, basetypes.ObjectAsOptions{})...)
	}

	endpointFallback := types.StringNull()
	pathStyleAccessFallback := false
	if !state.S3.IsNull() && !state.S3.IsUnknown() {
		var stateS3 S3Settings
		if asDiags := state.S3.As(ctx, &stateS3, basetypes.ObjectAsOptions{}); !asDiags.HasError() {
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
	if endpointStr := strSetting(s, settingEndpoint); endpointStr != "" {
		endpoint = strSettingNull(s, settingEndpoint)
	} else if !endpointFallback.IsNull() {
		endpoint = endpointFallback
	} else {
		endpoint = types.StringNull()
	}

	s3 := S3Settings{
		CommonSettings: CommonSettings{
			ChunkSize:              strSettingNullWithFallback(s, settingChunkSize, priorS3.ChunkSize),
			Compress:               types.BoolValue(boolSetting(s, settingCompress, true)),
			MaxSnapshotBytesPerSec: strSettingNullWithFallback(s, settingMaxSnapshotBytesPerSec, priorS3.MaxSnapshotBytesPerSec),
			MaxRestoreBytesPerSec:  strSettingNullWithFallback(s, settingMaxRestoreBytesPerSec, priorS3.MaxRestoreBytesPerSec),
			Readonly:               types.BoolValue(boolSetting(s, settingReadonly, false)),
		},
		Bucket:               types.StringValue(strSetting(s, settingBucket)),
		Endpoint:             endpoint,
		Client:               strSettingNull(s, settingClient),
		BasePath:             strSettingNull(s, settingBasePath),
		ServerSideEncryption: types.BoolValue(boolSetting(s, settingServerSideEncryption, false)),
		BufferSize:           strSettingNull(s, settingBufferSize),
		CannedACL:            strSettingNull(s, settingCannedACL),
		StorageClass:         strSettingNull(s, settingStorageClass),
		PathStyleAccess:      types.BoolValue(boolSetting(s, settingPathStyleAccess, pathStyleAccessFallback)),
	}
	obj, objDiags := types.ObjectValueFrom(ctx, s3AttrTypes(), s3)
	diags.Append(objDiags...)
	return obj, diags
}

func settingsToHdfs(ctx context.Context, repo *elasticsearch.SnapshotRepositoryInfo, state Data) (types.Object, diag.Diagnostics) {
	s := repo.Settings

	var diags diag.Diagnostics
	var priorHdfs HdfsSettings
	if !state.Hdfs.IsNull() && !state.Hdfs.IsUnknown() {
		diags.Append(state.Hdfs.As(ctx, &priorHdfs, basetypes.ObjectAsOptions{})...)
	}

	hdfs := HdfsSettings{
		CommonSettings: CommonSettings{
			ChunkSize:              strSettingNullWithFallback(s, settingChunkSize, priorHdfs.ChunkSize),
			Compress:               types.BoolValue(boolSetting(s, settingCompress, true)),
			MaxSnapshotBytesPerSec: strSettingNullWithFallback(s, settingMaxSnapshotBytesPerSec, priorHdfs.MaxSnapshotBytesPerSec),
			MaxRestoreBytesPerSec:  strSettingNullWithFallback(s, settingMaxRestoreBytesPerSec, priorHdfs.MaxRestoreBytesPerSec),
			Readonly:               types.BoolValue(boolSetting(s, settingReadonly, false)),
		},
		URI:          types.StringValue(strSetting(s, settingURI)),
		Path:         types.StringValue(strSetting(s, settingPath)),
		LoadDefaults: types.BoolValue(boolSetting(s, settingLoadDefaults, true)),
	}
	obj, objDiags := types.ObjectValueFrom(ctx, hdfsAttrTypes(), hdfs)
	diags.Append(objDiags...)
	return obj, diags
}
