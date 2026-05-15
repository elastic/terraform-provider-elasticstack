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

	esclients "github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// writeSnapshotRepository handles both Create and Update; the repository PUT
// API is idempotent so the same callback serves both lifecycle methods.
func writeSnapshotRepository(ctx context.Context, client *esclients.ElasticsearchScopedClient, req entitycore.WriteRequest[Data]) (entitycore.WriteResult[Data], diag.Diagnostics) {
	var diags diag.Diagnostics
	data := req.Plan
	resourceID := req.WriteID

	id, sdkDiags := client.ID(ctx, resourceID)
	diags.Append(diagutil.FrameworkDiagsFromSDK(sdkDiags)...)
	if diags.HasError() {
		return entitycore.WriteResult[Data]{}, diags
	}

	repoType, settings, settingsDiags := extractSettings(ctx, data)
	diags.Append(settingsDiags...)
	if diags.HasError() {
		return entitycore.WriteResult[Data]{}, diags
	}

	verify := true
	if !data.Verify.IsNull() && !data.Verify.IsUnknown() {
		verify = data.Verify.ValueBool()
	}

	diags.Append(diagutil.FrameworkDiagsFromSDK(elasticsearch.PutSnapshotRepository(ctx, client, resourceID, repoType, settings, verify))...)
	if diags.HasError() {
		return entitycore.WriteResult[Data]{}, diags
	}

	data.ID = types.StringValue(id.String())
	return entitycore.WriteResult[Data]{Model: data}, diags
}

// extractSettings determines the repository type and builds the settings map.
func extractSettings(ctx context.Context, data Data) (string, map[string]any, diag.Diagnostics) {
	var diags diag.Diagnostics

	if !data.Fs.IsNull() && !data.Fs.IsUnknown() {
		var fs FsSettings
		diags.Append(data.Fs.As(ctx, &fs, basetypes.ObjectAsOptions{})...)
		if diags.HasError() {
			return "", nil, diags
		}
		return "fs", fsToSettings(fs), diags
	}
	if !data.URL.IsNull() && !data.URL.IsUnknown() {
		var u URLSettings
		diags.Append(data.URL.As(ctx, &u, basetypes.ObjectAsOptions{})...)
		if diags.HasError() {
			return "", nil, diags
		}
		return "url", urlToSettings(u), diags
	}
	if !data.Gcs.IsNull() && !data.Gcs.IsUnknown() {
		var gcs GcsSettings
		diags.Append(data.Gcs.As(ctx, &gcs, basetypes.ObjectAsOptions{})...)
		if diags.HasError() {
			return "", nil, diags
		}
		return "gcs", gcsToSettings(gcs), diags
	}
	if !data.Azure.IsNull() && !data.Azure.IsUnknown() {
		var azure AzureSettings
		diags.Append(data.Azure.As(ctx, &azure, basetypes.ObjectAsOptions{})...)
		if diags.HasError() {
			return "", nil, diags
		}
		return "azure", azureToSettings(azure), diags
	}
	if !data.S3.IsNull() && !data.S3.IsUnknown() {
		var s3 S3Settings
		diags.Append(data.S3.As(ctx, &s3, basetypes.ObjectAsOptions{})...)
		if diags.HasError() {
			return "", nil, diags
		}
		return "s3", s3ToSettings(s3), diags
	}
	if !data.Hdfs.IsNull() && !data.Hdfs.IsUnknown() {
		var hdfs HdfsSettings
		diags.Append(data.Hdfs.As(ctx, &hdfs, basetypes.ObjectAsOptions{})...)
		if diags.HasError() {
			return "", nil, diags
		}
		return "hdfs", hdfsToSettings(hdfs), diags
	}

	diags.AddError("No repository type set", "Exactly one repository type block must be set")
	return "", nil, diags
}

func setIfNotEmpty(m map[string]any, key string, val string) {
	if val != "" {
		m[key] = val
	}
}

func fsToSettings(fs FsSettings) map[string]any {
	m := map[string]any{
		"location": fs.Location.ValueString(),
		"compress": fs.Compress.ValueBool(),
		"readonly": fs.Readonly.ValueBool(),
	}
	setIfNotEmpty(m, "chunk_size", fs.ChunkSize.ValueString())
	setIfNotEmpty(m, "max_snapshot_bytes_per_sec", fs.MaxSnapshotBytesPerSec.ValueString())
	setIfNotEmpty(m, "max_restore_bytes_per_sec", fs.MaxRestoreBytesPerSec.ValueString())
	if !fs.MaxNumberOfSnapshots.IsNull() && !fs.MaxNumberOfSnapshots.IsUnknown() {
		m["max_number_of_snapshots"] = fs.MaxNumberOfSnapshots.ValueInt64()
	}
	return m
}

func urlToSettings(u URLSettings) map[string]any {
	m := map[string]any{
		"url":      u.URL.ValueString(),
		"compress": u.Compress.ValueBool(),
		"readonly": u.Readonly.ValueBool(),
	}
	setIfNotEmpty(m, "chunk_size", u.ChunkSize.ValueString())
	setIfNotEmpty(m, "max_snapshot_bytes_per_sec", u.MaxSnapshotBytesPerSec.ValueString())
	setIfNotEmpty(m, "max_restore_bytes_per_sec", u.MaxRestoreBytesPerSec.ValueString())
	if !u.HTTPMaxRetries.IsNull() && !u.HTTPMaxRetries.IsUnknown() {
		m["http_max_retries"] = u.HTTPMaxRetries.ValueInt64()
	}
	setIfNotEmpty(m, "http_socket_timeout", u.HTTPSocketTimeout.ValueString())
	if !u.MaxNumberOfSnapshots.IsNull() && !u.MaxNumberOfSnapshots.IsUnknown() {
		m["max_number_of_snapshots"] = u.MaxNumberOfSnapshots.ValueInt64()
	}
	return m
}

func gcsToSettings(gcs GcsSettings) map[string]any {
	m := map[string]any{
		"bucket":   gcs.Bucket.ValueString(),
		"compress": gcs.Compress.ValueBool(),
		"readonly": gcs.Readonly.ValueBool(),
	}
	setIfNotEmpty(m, "client", gcs.Client.ValueString())
	setIfNotEmpty(m, "base_path", gcs.BasePath.ValueString())
	setIfNotEmpty(m, "chunk_size", gcs.ChunkSize.ValueString())
	setIfNotEmpty(m, "max_snapshot_bytes_per_sec", gcs.MaxSnapshotBytesPerSec.ValueString())
	setIfNotEmpty(m, "max_restore_bytes_per_sec", gcs.MaxRestoreBytesPerSec.ValueString())
	return m
}

func azureToSettings(azure AzureSettings) map[string]any {
	m := map[string]any{
		"container": azure.Container.ValueString(),
		"compress":  azure.Compress.ValueBool(),
		"readonly":  azure.Readonly.ValueBool(),
	}
	setIfNotEmpty(m, "client", azure.Client.ValueString())
	setIfNotEmpty(m, "base_path", azure.BasePath.ValueString())
	setIfNotEmpty(m, "location_mode", azure.LocationMode.ValueString())
	setIfNotEmpty(m, "chunk_size", azure.ChunkSize.ValueString())
	setIfNotEmpty(m, "max_snapshot_bytes_per_sec", azure.MaxSnapshotBytesPerSec.ValueString())
	setIfNotEmpty(m, "max_restore_bytes_per_sec", azure.MaxRestoreBytesPerSec.ValueString())
	return m
}

func s3ToSettings(s3 S3Settings) map[string]any {
	m := map[string]any{
		"bucket":                 s3.Bucket.ValueString(),
		"compress":               s3.Compress.ValueBool(),
		"readonly":               s3.Readonly.ValueBool(),
		"server_side_encryption": s3.ServerSideEncryption.ValueBool(),
		"path_style_access":      s3.PathStyleAccess.ValueBool(),
	}
	setIfNotEmpty(m, "endpoint", s3.Endpoint.ValueString())
	setIfNotEmpty(m, "client", s3.Client.ValueString())
	setIfNotEmpty(m, "base_path", s3.BasePath.ValueString())
	setIfNotEmpty(m, "buffer_size", s3.BufferSize.ValueString())
	setIfNotEmpty(m, "canned_acl", s3.CannedACL.ValueString())
	setIfNotEmpty(m, "storage_class", s3.StorageClass.ValueString())
	setIfNotEmpty(m, "chunk_size", s3.ChunkSize.ValueString())
	setIfNotEmpty(m, "max_snapshot_bytes_per_sec", s3.MaxSnapshotBytesPerSec.ValueString())
	setIfNotEmpty(m, "max_restore_bytes_per_sec", s3.MaxRestoreBytesPerSec.ValueString())
	return m
}

func hdfsToSettings(hdfs HdfsSettings) map[string]any {
	m := map[string]any{
		"uri":           hdfs.URI.ValueString(),
		"path":          hdfs.Path.ValueString(),
		"load_defaults": hdfs.LoadDefaults.ValueBool(),
		"compress":      hdfs.Compress.ValueBool(),
		"readonly":      hdfs.Readonly.ValueBool(),
	}
	setIfNotEmpty(m, "chunk_size", hdfs.ChunkSize.ValueString())
	setIfNotEmpty(m, "max_snapshot_bytes_per_sec", hdfs.MaxSnapshotBytesPerSec.ValueString())
	setIfNotEmpty(m, "max_restore_bytes_per_sec", hdfs.MaxRestoreBytesPerSec.ValueString())
	return m
}
