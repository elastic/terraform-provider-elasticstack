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
	"strconv"

	esclients "github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func readSnapshotRepository(ctx context.Context, client *esclients.ElasticsearchScopedClient, resourceID string, state Data) (Data, bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	repo, sdkDiags := elasticsearch.GetSnapshotRepository(ctx, client, resourceID)
	diags.Append(diagutil.FrameworkDiagsFromSDK(sdkDiags)...)
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
	case "fs":
		fs, fsDiags := settingsToFs(ctx, repo, state)
		diags.Append(fsDiags...)
		if diags.HasError() {
			return state, false, diags
		}
		data.Fs = fs
	case "url":
		u, uDiags := settingsToURL(ctx, repo, state)
		diags.Append(uDiags...)
		if diags.HasError() {
			return state, false, diags
		}
		data.URL = u
	case "gcs":
		gcs, gcsDiags := settingsToGcs(ctx, repo)
		diags.Append(gcsDiags...)
		if diags.HasError() {
			return state, false, diags
		}
		data.Gcs = gcs
	case "azure":
		azure, azureDiags := settingsToAzure(ctx, repo)
		diags.Append(azureDiags...)
		if diags.HasError() {
			return state, false, diags
		}
		data.Azure = azure
	case "s3":
		s3, s3Diags := settingsToS3(ctx, repo)
		diags.Append(s3Diags...)
		if diags.HasError() {
			return state, false, diags
		}
		data.S3 = s3
	case "hdfs":
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

// strSetting extracts a string setting with a fallback.
func strSetting(settings map[string]any, key string) string {
	v, ok := settings[key]
	if !ok {
		return ""
	}
	switch val := v.(type) {
	case string:
		return val
	default:
		return fmt.Sprintf("%v", v)
	}
}

// strSettingNull extracts a string setting, returning null when the key is absent.
func strSettingNull(settings map[string]any, key string) types.String {
	v, ok := settings[key]
	if !ok {
		return types.StringNull()
	}
	switch val := v.(type) {
	case string:
		return types.StringValue(val)
	default:
		return types.StringValue(fmt.Sprintf("%v", v))
	}
}

// boolSetting extracts a bool setting with a fallback.
func boolSetting(settings map[string]any, key string, fallback bool) bool {
	v, ok := settings[key]
	if !ok {
		return fallback
	}
	switch val := v.(type) {
	case bool:
		return val
	case string:
		b, err := strconv.ParseBool(val)
		if err != nil {
			return fallback
		}
		return b
	default:
		return fallback
	}
}

// int64Setting extracts an int64 setting with a fallback.
func int64Setting(settings map[string]any, key string, fallback int64) int64 {
	v, ok := settings[key]
	if !ok {
		return fallback
	}
	switch val := v.(type) {
	case int:
		return int64(val)
	case int64:
		return val
	case float64:
		return int64(val)
	case string:
		i, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			f, err2 := strconv.ParseFloat(val, 64)
			if err2 != nil {
				return fallback
			}
			return int64(f)
		}
		return i
	default:
		return fallback
	}
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
			ChunkSize:              strSettingNull(s, "chunk_size"),
			Compress:               types.BoolValue(boolSetting(s, "compress", compressFallback)),
			MaxSnapshotBytesPerSec: strSettingNull(s, "max_snapshot_bytes_per_sec"),
			MaxRestoreBytesPerSec:  strSettingNull(s, "max_restore_bytes_per_sec"),
			Readonly:               types.BoolValue(boolSetting(s, "readonly", false)),
		},
		CommonStdSettings: CommonStdSettings{
			MaxNumberOfSnapshots: types.Int64Value(int64Setting(s, "max_number_of_snapshots", 500)),
		},
		Location: types.StringValue(strSetting(s, "location")),
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
			ChunkSize:              strSettingNull(s, "chunk_size"),
			Compress:               types.BoolValue(boolSetting(s, "compress", compressFallback)),
			MaxSnapshotBytesPerSec: strSettingNull(s, "max_snapshot_bytes_per_sec"),
			MaxRestoreBytesPerSec:  strSettingNull(s, "max_restore_bytes_per_sec"),
			Readonly:               types.BoolValue(boolSetting(s, "readonly", false)),
		},
		CommonStdSettings: CommonStdSettings{
			MaxNumberOfSnapshots: types.Int64Value(int64Setting(s, "max_number_of_snapshots", 500)),
		},
		URL:               types.StringValue(strSetting(s, "url")),
		HTTPMaxRetries:    types.Int64Value(int64Setting(s, "http_max_retries", 5)),
		HTTPSocketTimeout: strSettingNull(s, "http_socket_timeout"),
	}
	return types.ObjectValueFrom(ctx, urlAttrTypes(), u)
}

func settingsToGcs(ctx context.Context, repo *elasticsearch.SnapshotRepositoryInfo) (types.Object, diag.Diagnostics) {
	s := repo.Settings
	gcs := GcsSettings{
		CommonSettings: CommonSettings{
			ChunkSize:              strSettingNull(s, "chunk_size"),
			Compress:               types.BoolValue(boolSetting(s, "compress", true)),
			MaxSnapshotBytesPerSec: strSettingNull(s, "max_snapshot_bytes_per_sec"),
			MaxRestoreBytesPerSec:  strSettingNull(s, "max_restore_bytes_per_sec"),
			Readonly:               types.BoolValue(boolSetting(s, "readonly", false)),
		},
		Bucket:   types.StringValue(strSetting(s, "bucket")),
		Client:   strSettingNull(s, "client"),
		BasePath: strSettingNull(s, "base_path"),
	}
	return types.ObjectValueFrom(ctx, gcsAttrTypes(), gcs)
}

func settingsToAzure(ctx context.Context, repo *elasticsearch.SnapshotRepositoryInfo) (types.Object, diag.Diagnostics) {
	s := repo.Settings
	azure := AzureSettings{
		CommonSettings: CommonSettings{
			ChunkSize:              strSettingNull(s, "chunk_size"),
			Compress:               types.BoolValue(boolSetting(s, "compress", true)),
			MaxSnapshotBytesPerSec: strSettingNull(s, "max_snapshot_bytes_per_sec"),
			MaxRestoreBytesPerSec:  strSettingNull(s, "max_restore_bytes_per_sec"),
			Readonly:               types.BoolValue(boolSetting(s, "readonly", false)),
		},
		Container:    types.StringValue(strSetting(s, "container")),
		Client:       strSettingNull(s, "client"),
		BasePath:     strSettingNull(s, "base_path"),
		LocationMode: strSettingNull(s, "location_mode"),
	}
	return types.ObjectValueFrom(ctx, azureAttrTypes(), azure)
}

func settingsToS3(ctx context.Context, repo *elasticsearch.SnapshotRepositoryInfo) (types.Object, diag.Diagnostics) {
	s := repo.Settings
	s3 := S3Settings{
		CommonSettings: CommonSettings{
			ChunkSize:              strSettingNull(s, "chunk_size"),
			Compress:               types.BoolValue(boolSetting(s, "compress", true)),
			MaxSnapshotBytesPerSec: strSettingNull(s, "max_snapshot_bytes_per_sec"),
			MaxRestoreBytesPerSec:  strSettingNull(s, "max_restore_bytes_per_sec"),
			Readonly:               types.BoolValue(boolSetting(s, "readonly", false)),
		},
		Bucket:               types.StringValue(strSetting(s, "bucket")),
		Endpoint:             strSettingNull(s, "endpoint"),
		Client:               strSettingNull(s, "client"),
		BasePath:             strSettingNull(s, "base_path"),
		ServerSideEncryption: types.BoolValue(boolSetting(s, "server_side_encryption", false)),
		BufferSize:           strSettingNull(s, "buffer_size"),
		CannedACL:            strSettingNull(s, "canned_acl"),
		StorageClass:         strSettingNull(s, "storage_class"),
		PathStyleAccess:      types.BoolValue(boolSetting(s, "path_style_access", false)),
	}
	return types.ObjectValueFrom(ctx, s3AttrTypes(), s3)
}

func settingsToHdfs(ctx context.Context, repo *elasticsearch.SnapshotRepositoryInfo) (types.Object, diag.Diagnostics) {
	s := repo.Settings
	hdfs := HdfsSettings{
		CommonSettings: CommonSettings{
			ChunkSize:              strSettingNull(s, "chunk_size"),
			Compress:               types.BoolValue(boolSetting(s, "compress", true)),
			MaxSnapshotBytesPerSec: strSettingNull(s, "max_snapshot_bytes_per_sec"),
			MaxRestoreBytesPerSec:  strSettingNull(s, "max_restore_bytes_per_sec"),
			Readonly:               types.BoolValue(boolSetting(s, "readonly", false)),
		},
		URI:          types.StringValue(strSetting(s, "uri")),
		Path:         types.StringValue(strSetting(s, "path")),
		LoadDefaults: types.BoolValue(boolSetting(s, "load_defaults", true)),
	}
	return types.ObjectValueFrom(ctx, hdfsAttrTypes(), hdfs)
}
