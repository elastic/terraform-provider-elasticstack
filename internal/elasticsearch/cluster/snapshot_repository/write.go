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

	id, idDiags := client.ID(ctx, resourceID)
	diags.Append(idDiags...)
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

	diags.Append(elasticsearch.PutSnapshotRepository(ctx, client, resourceID, repoType, settings, verify)...)
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
		return repoTypeFS, fsToSettings(fs), diags
	}
	if !data.URL.IsNull() && !data.URL.IsUnknown() {
		var u URLSettings
		diags.Append(data.URL.As(ctx, &u, basetypes.ObjectAsOptions{})...)
		if diags.HasError() {
			return "", nil, diags
		}
		return repoTypeURL, urlToSettings(u), diags
	}
	if !data.Gcs.IsNull() && !data.Gcs.IsUnknown() {
		var gcs GcsSettings
		diags.Append(data.Gcs.As(ctx, &gcs, basetypes.ObjectAsOptions{})...)
		if diags.HasError() {
			return "", nil, diags
		}
		return repoTypeGCS, gcsToSettings(gcs), diags
	}
	if !data.Azure.IsNull() && !data.Azure.IsUnknown() {
		var azure AzureSettings
		diags.Append(data.Azure.As(ctx, &azure, basetypes.ObjectAsOptions{})...)
		if diags.HasError() {
			return "", nil, diags
		}
		return repoTypeAzure, azureToSettings(azure), diags
	}
	if !data.S3.IsNull() && !data.S3.IsUnknown() {
		var s3 S3Settings
		diags.Append(data.S3.As(ctx, &s3, basetypes.ObjectAsOptions{})...)
		if diags.HasError() {
			return "", nil, diags
		}
		return repoTypeS3, s3ToSettings(s3), diags
	}
	if !data.Hdfs.IsNull() && !data.Hdfs.IsUnknown() {
		var hdfs HdfsSettings
		diags.Append(data.Hdfs.As(ctx, &hdfs, basetypes.ObjectAsOptions{})...)
		if diags.HasError() {
			return "", nil, diags
		}
		return repoTypeHDFS, hdfsToSettings(hdfs), diags
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
		settingLocation: fs.Location.ValueString(),
		settingCompress: fs.Compress.ValueBool(),
		settingReadonly: fs.Readonly.ValueBool(),
	}
	setIfNotEmpty(m, settingChunkSize, fs.ChunkSize.ValueString())
	setIfNotEmpty(m, settingMaxSnapshotBytesPerSec, fs.MaxSnapshotBytesPerSec.ValueString())
	setIfNotEmpty(m, settingMaxRestoreBytesPerSec, fs.MaxRestoreBytesPerSec.ValueString())
	if !fs.MaxNumberOfSnapshots.IsNull() && !fs.MaxNumberOfSnapshots.IsUnknown() {
		m[settingMaxNumberOfSnapshots] = fs.MaxNumberOfSnapshots.ValueInt64()
	}
	return m
}

func urlToSettings(u URLSettings) map[string]any {
	m := map[string]any{
		settingURL:      u.URL.ValueString(),
		settingCompress: u.Compress.ValueBool(),
		settingReadonly: u.Readonly.ValueBool(),
	}
	setIfNotEmpty(m, settingChunkSize, u.ChunkSize.ValueString())
	setIfNotEmpty(m, settingMaxSnapshotBytesPerSec, u.MaxSnapshotBytesPerSec.ValueString())
	setIfNotEmpty(m, settingMaxRestoreBytesPerSec, u.MaxRestoreBytesPerSec.ValueString())
	if !u.HTTPMaxRetries.IsNull() && !u.HTTPMaxRetries.IsUnknown() {
		m[settingHTTPMaxRetries] = u.HTTPMaxRetries.ValueInt64()
	}
	setIfNotEmpty(m, settingHTTPSocketTimeout, u.HTTPSocketTimeout.ValueString())
	if !u.MaxNumberOfSnapshots.IsNull() && !u.MaxNumberOfSnapshots.IsUnknown() {
		m[settingMaxNumberOfSnapshots] = u.MaxNumberOfSnapshots.ValueInt64()
	}
	return m
}

func gcsToSettings(gcs GcsSettings) map[string]any {
	m := map[string]any{
		settingBucket:   gcs.Bucket.ValueString(),
		settingCompress: gcs.Compress.ValueBool(),
		settingReadonly: gcs.Readonly.ValueBool(),
	}
	setIfNotEmpty(m, settingClient, gcs.Client.ValueString())
	setIfNotEmpty(m, settingBasePath, gcs.BasePath.ValueString())
	setIfNotEmpty(m, settingChunkSize, gcs.ChunkSize.ValueString())
	setIfNotEmpty(m, settingMaxSnapshotBytesPerSec, gcs.MaxSnapshotBytesPerSec.ValueString())
	setIfNotEmpty(m, settingMaxRestoreBytesPerSec, gcs.MaxRestoreBytesPerSec.ValueString())
	return m
}

func azureToSettings(azure AzureSettings) map[string]any {
	m := map[string]any{
		settingContainer: azure.Container.ValueString(),
		settingCompress:  azure.Compress.ValueBool(),
		settingReadonly:  azure.Readonly.ValueBool(),
	}
	setIfNotEmpty(m, settingClient, azure.Client.ValueString())
	setIfNotEmpty(m, settingBasePath, azure.BasePath.ValueString())
	setIfNotEmpty(m, settingLocationMode, azure.LocationMode.ValueString())
	setIfNotEmpty(m, settingChunkSize, azure.ChunkSize.ValueString())
	setIfNotEmpty(m, settingMaxSnapshotBytesPerSec, azure.MaxSnapshotBytesPerSec.ValueString())
	setIfNotEmpty(m, settingMaxRestoreBytesPerSec, azure.MaxRestoreBytesPerSec.ValueString())
	return m
}

func s3ToSettings(s3 S3Settings) map[string]any {
	m := map[string]any{
		settingBucket:               s3.Bucket.ValueString(),
		settingCompress:             s3.Compress.ValueBool(),
		settingReadonly:             s3.Readonly.ValueBool(),
		settingServerSideEncryption: s3.ServerSideEncryption.ValueBool(),
		settingPathStyleAccess:      s3.PathStyleAccess.ValueBool(),
	}
	setIfNotEmpty(m, settingEndpoint, s3.Endpoint.ValueString())
	setIfNotEmpty(m, settingClient, s3.Client.ValueString())
	setIfNotEmpty(m, settingBasePath, s3.BasePath.ValueString())
	setIfNotEmpty(m, settingBufferSize, s3.BufferSize.ValueString())
	setIfNotEmpty(m, settingCannedACL, s3.CannedACL.ValueString())
	setIfNotEmpty(m, settingStorageClass, s3.StorageClass.ValueString())
	setIfNotEmpty(m, settingChunkSize, s3.ChunkSize.ValueString())
	setIfNotEmpty(m, settingMaxSnapshotBytesPerSec, s3.MaxSnapshotBytesPerSec.ValueString())
	setIfNotEmpty(m, settingMaxRestoreBytesPerSec, s3.MaxRestoreBytesPerSec.ValueString())
	return m
}

func hdfsToSettings(hdfs HdfsSettings) map[string]any {
	m := map[string]any{
		settingURI:          hdfs.URI.ValueString(),
		settingPath:         hdfs.Path.ValueString(),
		settingLoadDefaults: hdfs.LoadDefaults.ValueBool(),
		settingCompress:     hdfs.Compress.ValueBool(),
		settingReadonly:     hdfs.Readonly.ValueBool(),
	}
	setIfNotEmpty(m, settingChunkSize, hdfs.ChunkSize.ValueString())
	setIfNotEmpty(m, settingMaxSnapshotBytesPerSec, hdfs.MaxSnapshotBytesPerSec.ValueString())
	setIfNotEmpty(m, settingMaxRestoreBytesPerSec, hdfs.MaxRestoreBytesPerSec.ValueString())
	return m
}
