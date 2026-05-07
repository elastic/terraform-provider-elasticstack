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

package cluster

import (
	"context"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFlattenFsSettings(t *testing.T) {
	settings := map[string]any{
		"location":                   "/tmp",
		"chunk_size":                 "1gb",
		"compress":                   true,
		"max_snapshot_bytes_per_sec": "20mb",
		"max_restore_bytes_per_sec":  "10mb",
		"readonly":                   "false",
		"max_number_of_snapshots":    50,
	}
	model, err := flattenFsSettings(settings)
	require.NoError(t, err)

	assert.Equal(t, types.StringValue("/tmp"), model.Location)
	assert.Equal(t, types.StringValue("1gb"), model.ChunkSize)
	assert.Equal(t, types.BoolValue(true), model.Compress)
	assert.Equal(t, types.StringValue("20mb"), model.MaxSnapshotBytesPerSec)
	assert.Equal(t, types.StringValue("10mb"), model.MaxRestoreBytesPerSec)
	assert.Equal(t, types.BoolValue(false), model.Readonly)
	assert.Equal(t, types.Int64Value(50), model.MaxNumberOfSnapshots)
}

func TestFlattenFsSettings_Nulls(t *testing.T) {
	model, err := flattenFsSettings(map[string]any{})
	require.NoError(t, err)

	assert.True(t, model.Location.IsNull())
	assert.True(t, model.ChunkSize.IsNull())
	assert.True(t, model.Compress.IsNull())
	assert.True(t, model.MaxSnapshotBytesPerSec.IsNull())
	assert.True(t, model.MaxRestoreBytesPerSec.IsNull())
	assert.True(t, model.Readonly.IsNull())
	assert.True(t, model.MaxNumberOfSnapshots.IsNull())
}

func TestFlattenURLSettings(t *testing.T) {
	settings := map[string]any{
		"url":                        "file:/tmp",
		"chunk_size":                 "1gb",
		"compress":                   "true",
		"max_snapshot_bytes_per_sec": "40mb",
		"max_restore_bytes_per_sec":  "10mb",
		"readonly":                   false,
		"max_number_of_snapshots":    500,
		"http_max_retries":           3,
		"http_socket_timeout":        "30s",
	}
	model, err := flattenURLSettings(settings)
	require.NoError(t, err)

	assert.Equal(t, types.StringValue("file:/tmp"), model.URL)
	assert.Equal(t, types.StringValue("1gb"), model.ChunkSize)
	assert.Equal(t, types.BoolValue(true), model.Compress)
	assert.Equal(t, types.StringValue("40mb"), model.MaxSnapshotBytesPerSec)
	assert.Equal(t, types.StringValue("10mb"), model.MaxRestoreBytesPerSec)
	assert.Equal(t, types.BoolValue(false), model.Readonly)
	assert.Equal(t, types.Int64Value(500), model.MaxNumberOfSnapshots)
	assert.Equal(t, types.Int64Value(3), model.HTTPMaxRetries)
	assert.Equal(t, types.StringValue("30s"), model.HTTPSocketTimeout)
}

func TestFlattenURLSettings_Nulls(t *testing.T) {
	model, err := flattenURLSettings(map[string]any{})
	require.NoError(t, err)

	assert.True(t, model.URL.IsNull())
	assert.True(t, model.ChunkSize.IsNull())
	assert.True(t, model.Compress.IsNull())
	assert.True(t, model.MaxSnapshotBytesPerSec.IsNull())
	assert.True(t, model.MaxRestoreBytesPerSec.IsNull())
	assert.True(t, model.Readonly.IsNull())
	assert.True(t, model.MaxNumberOfSnapshots.IsNull())
	assert.True(t, model.HTTPMaxRetries.IsNull())
	assert.True(t, model.HTTPSocketTimeout.IsNull())
}

func TestFlattenGCSSettings(t *testing.T) {
	settings := map[string]any{
		"bucket":                     "my-bucket",
		"client":                     "my-client",
		"base_path":                  "prefix",
		"chunk_size":                 "64mb",
		"compress":                   true,
		"max_snapshot_bytes_per_sec": "40mb",
		"max_restore_bytes_per_sec":  "20mb",
		"readonly":                   "true",
	}
	model, err := flattenGCSSettings(settings)
	require.NoError(t, err)

	assert.Equal(t, types.StringValue("my-bucket"), model.Bucket)
	assert.Equal(t, types.StringValue("my-client"), model.Client)
	assert.Equal(t, types.StringValue("prefix"), model.BasePath)
	assert.Equal(t, types.StringValue("64mb"), model.ChunkSize)
	assert.Equal(t, types.BoolValue(true), model.Compress)
	assert.Equal(t, types.StringValue("40mb"), model.MaxSnapshotBytesPerSec)
	assert.Equal(t, types.StringValue("20mb"), model.MaxRestoreBytesPerSec)
	assert.Equal(t, types.BoolValue(true), model.Readonly)
}

func TestFlattenGCSSettings_Nulls(t *testing.T) {
	model, err := flattenGCSSettings(map[string]any{})
	require.NoError(t, err)

	assert.True(t, model.Bucket.IsNull())
	assert.True(t, model.Client.IsNull())
	assert.True(t, model.BasePath.IsNull())
	assert.True(t, model.ChunkSize.IsNull())
	assert.True(t, model.Compress.IsNull())
	assert.True(t, model.MaxSnapshotBytesPerSec.IsNull())
	assert.True(t, model.MaxRestoreBytesPerSec.IsNull())
	assert.True(t, model.Readonly.IsNull())
}

func TestFlattenAzureSettings(t *testing.T) {
	settings := map[string]any{
		"container":     "my-container",
		"client":        "my-client",
		"base_path":     "prefix",
		"location_mode": "primary_only",
		"compress":      "false",
		"readonly":      false,
		"chunk_size":    1024,
	}
	model, err := flattenAzureSettings(settings)
	require.NoError(t, err)

	assert.Equal(t, types.StringValue("my-container"), model.Container)
	assert.Equal(t, types.StringValue("my-client"), model.Client)
	assert.Equal(t, types.StringValue("prefix"), model.BasePath)
	assert.Equal(t, types.StringValue("primary_only"), model.LocationMode)
	assert.Equal(t, types.BoolValue(false), model.Compress)
	assert.Equal(t, types.BoolValue(false), model.Readonly)
	assert.Equal(t, types.StringValue("1024"), model.ChunkSize)
}

func TestFlattenAzureSettings_Nulls(t *testing.T) {
	model, err := flattenAzureSettings(map[string]any{})
	require.NoError(t, err)

	assert.True(t, model.Container.IsNull())
	assert.True(t, model.Client.IsNull())
	assert.True(t, model.BasePath.IsNull())
	assert.True(t, model.LocationMode.IsNull())
	assert.True(t, model.ChunkSize.IsNull())
	assert.True(t, model.Compress.IsNull())
	assert.True(t, model.MaxSnapshotBytesPerSec.IsNull())
	assert.True(t, model.MaxRestoreBytesPerSec.IsNull())
	assert.True(t, model.Readonly.IsNull())
}

func TestFlattenS3Settings(t *testing.T) {
	settings := map[string]any{
		"bucket":                  "my-bucket",
		"client":                  "default",
		"base_path":               "backups",
		"server_side_encryption":  true,
		"buffer_size":             "5mb",
		"canned_acl":              "private",
		"storage_class":           "standard",
		"path_style_access":       "false",
		"max_number_of_snapshots": 100,
		"compress":                true,
	}
	model, err := flattenS3Settings(settings)
	require.NoError(t, err)

	assert.Equal(t, types.StringValue("my-bucket"), model.Bucket)
	assert.Equal(t, types.StringValue("default"), model.Client)
	assert.Equal(t, types.StringValue("backups"), model.BasePath)
	assert.Equal(t, types.BoolValue(true), model.ServerSideEncryption)
	assert.Equal(t, types.StringValue("5mb"), model.BufferSize)
	assert.Equal(t, types.StringValue("private"), model.CannedACL)
	assert.Equal(t, types.StringValue("standard"), model.StorageClass)
	assert.Equal(t, types.BoolValue(false), model.PathStyleAccess)
	assert.Equal(t, types.BoolValue(true), model.Compress)
}

func TestFlattenS3Settings_Nulls(t *testing.T) {
	model, err := flattenS3Settings(map[string]any{})
	require.NoError(t, err)

	assert.True(t, model.Bucket.IsNull())
	assert.True(t, model.Client.IsNull())
	assert.True(t, model.BasePath.IsNull())
	assert.True(t, model.ServerSideEncryption.IsNull())
	assert.True(t, model.BufferSize.IsNull())
	assert.True(t, model.CannedACL.IsNull())
	assert.True(t, model.StorageClass.IsNull())
	assert.True(t, model.PathStyleAccess.IsNull())
	assert.True(t, model.ChunkSize.IsNull())
	assert.True(t, model.Compress.IsNull())
	assert.True(t, model.MaxSnapshotBytesPerSec.IsNull())
	assert.True(t, model.MaxRestoreBytesPerSec.IsNull())
	assert.True(t, model.Readonly.IsNull())
}

func TestFlattenS3Settings_NoEndpoint(t *testing.T) {
	// Ensure the S3 model does not include endpoint; this test documents
	// the schema decision that endpoint is absent from the data source.
	settings := map[string]any{
		"bucket":   "b",
		"endpoint": "http://example.com",
	}
	model, err := flattenS3Settings(settings)
	require.NoError(t, err)
	// The endpoint key should be ignored by the flatten helper.
	assert.Equal(t, types.StringValue("b"), model.Bucket)
}

func TestFlattenHdfsSettings(t *testing.T) {
	settings := map[string]any{
		"uri":           "hdfs://host:8020/",
		"path":          "/repo",
		"load_defaults": "true",
		"compress":      false,
		"readonly":      false,
		"chunk_size":    "32",
	}
	model, err := flattenHDFSSettings(settings)
	require.NoError(t, err)

	assert.Equal(t, types.StringValue("hdfs://host:8020/"), model.URI)
	assert.Equal(t, types.StringValue("/repo"), model.Path)
	assert.Equal(t, types.BoolValue(true), model.LoadDefaults)
	assert.Equal(t, types.BoolValue(false), model.Compress)
	assert.Equal(t, types.BoolValue(false), model.Readonly)
	assert.Equal(t, types.StringValue("32"), model.ChunkSize)
}

func TestFlattenHdfsSettings_Nulls(t *testing.T) {
	model, err := flattenHDFSSettings(map[string]any{})
	require.NoError(t, err)

	assert.True(t, model.URI.IsNull())
	assert.True(t, model.Path.IsNull())
	assert.True(t, model.LoadDefaults.IsNull())
	assert.True(t, model.Compress.IsNull())
	assert.True(t, model.Readonly.IsNull())
	assert.True(t, model.ChunkSize.IsNull())
	assert.True(t, model.MaxSnapshotBytesPerSec.IsNull())
	assert.True(t, model.MaxRestoreBytesPerSec.IsNull())
}

func TestStringSetting(t *testing.T) {
	assert.Equal(t, types.StringValue("hello"), stringSetting(map[string]any{"k": "hello"}, "k"))
	assert.Equal(t, types.StringValue("42"), stringSetting(map[string]any{"k": 42}, "k"))
	assert.True(t, stringSetting(map[string]any{}, "k").IsNull())
	assert.True(t, stringSetting(map[string]any{"k": nil}, "k").IsNull())
}

func TestBoolSetting(t *testing.T) {
	b, err := boolSetting(map[string]any{"k": true}, "k")
	require.NoError(t, err)
	assert.Equal(t, types.BoolValue(true), b)

	b, err = boolSetting(map[string]any{"k": "false"}, "k")
	require.NoError(t, err)
	assert.Equal(t, types.BoolValue(false), b)

	b, err = boolSetting(map[string]any{"k": "1"}, "k")
	require.NoError(t, err)
	assert.Equal(t, types.BoolValue(true), b)

	b, err = boolSetting(map[string]any{"k": "nope"}, "k")
	require.Error(t, err)
	assert.True(t, b.IsNull())

	b, err = boolSetting(map[string]any{}, "k")
	require.NoError(t, err)
	assert.True(t, b.IsNull())
}

func TestInt64Setting(t *testing.T) {
	i, err := int64Setting(map[string]any{"k": 42}, "k")
	require.NoError(t, err)
	assert.Equal(t, types.Int64Value(42), i)

	i, err = int64Setting(map[string]any{"k": int64(42)}, "k")
	require.NoError(t, err)
	assert.Equal(t, types.Int64Value(42), i)

	i, err = int64Setting(map[string]any{"k": float64(42)}, "k")
	require.NoError(t, err)
	assert.Equal(t, types.Int64Value(42), i)

	i, err = int64Setting(map[string]any{"k": "42"}, "k")
	require.NoError(t, err)
	assert.Equal(t, types.Int64Value(42), i)

	i, err = int64Setting(map[string]any{"k": "not-a-number"}, "k")
	require.Error(t, err)
	assert.True(t, i.IsNull())

	i, err = int64Setting(map[string]any{}, "k")
	require.NoError(t, err)
	assert.True(t, i.IsNull())
}

func TestPopulateRepositoryTypeBlocks_Success(t *testing.T) {
	tests := []struct {
		name          string
		repo          *elasticsearch.SnapshotRepositoryInfo
		wantPopulated string
	}{
		{
			name: "fs",
			repo: &elasticsearch.SnapshotRepositoryInfo{
				Type: repoTypeFS,
				Settings: map[string]any{
					"location": "/tmp",
					"compress": true,
				},
			},
			wantPopulated: repoTypeFS,
		},
		{
			name: "url",
			repo: &elasticsearch.SnapshotRepositoryInfo{
				Type: repoTypeURL,
				Settings: map[string]any{
					"url":      "file:/tmp",
					"readonly": "false",
				},
			},
			wantPopulated: repoTypeURL,
		},
		{
			name: "gcs",
			repo: &elasticsearch.SnapshotRepositoryInfo{
				Type: repoTypeGCS,
				Settings: map[string]any{
					"bucket": "my-bucket",
					"client": "default",
				},
			},
			wantPopulated: repoTypeGCS,
		},
		{
			name: "azure",
			repo: &elasticsearch.SnapshotRepositoryInfo{
				Type: repoTypeAzure,
				Settings: map[string]any{
					"container": "my-container",
					"client":    "default",
				},
			},
			wantPopulated: repoTypeAzure,
		},
		{
			name: "s3",
			repo: &elasticsearch.SnapshotRepositoryInfo{
				Type: repoTypeS3,
				Settings: map[string]any{
					"bucket":   "my-bucket",
					"client":   "default",
					"compress": true,
				},
			},
			wantPopulated: repoTypeS3,
		},
		{
			name: "hdfs",
			repo: &elasticsearch.SnapshotRepositoryInfo{
				Type: repoTypeHDFS,
				Settings: map[string]any{
					"uri":  "hdfs://host:8020/",
					"path": "/repo",
				},
			},
			wantPopulated: repoTypeHDFS,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			config := snapshotRepositoryDataSourceModel{
				Name: types.StringValue("test-repo"),
			}
			config, initDiags := initEmptyTypeBlocks(config)
			require.False(t, initDiags.HasError(), "unexpected init diagnostics: %v", initDiags)

			got, diags := populateRepositoryTypeBlocks(ctx, config, tt.repo)
			require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)

			assertListLength := func(list types.List, want int) {
				assert.False(t, list.IsNull(), "list should not be null")
				assert.Len(t, list.Elements(), want, "unexpected list length")
			}

			populated := map[string]int{
				repoTypeFS:    0,
				repoTypeURL:   0,
				repoTypeGCS:   0,
				repoTypeAzure: 0,
				repoTypeS3:    0,
				repoTypeHDFS:  0,
			}
			populated[tt.wantPopulated] = 1

			assertListLength(got.Fs, populated[repoTypeFS])
			assertListLength(got.URL, populated[repoTypeURL])
			assertListLength(got.GCS, populated[repoTypeGCS])
			assertListLength(got.Azure, populated[repoTypeAzure])
			assertListLength(got.S3, populated[repoTypeS3])
			assertListLength(got.HDFS, populated[repoTypeHDFS])
		})
	}
}

func TestInitEmptyTypeBlocks(t *testing.T) {
	config := snapshotRepositoryDataSourceModel{
		Name: types.StringValue("missing-repo"),
	}
	got, diags := initEmptyTypeBlocks(config)
	require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)

	assert.False(t, got.Fs.IsNull())
	assert.Empty(t, got.Fs.Elements())
	assert.False(t, got.URL.IsNull())
	assert.Empty(t, got.URL.Elements())
	assert.False(t, got.GCS.IsNull())
	assert.Empty(t, got.GCS.Elements())
	assert.False(t, got.Azure.IsNull())
	assert.Empty(t, got.Azure.Elements())
	assert.False(t, got.S3.IsNull())
	assert.Empty(t, got.S3.Elements())
	assert.False(t, got.HDFS.IsNull())
	assert.Empty(t, got.HDFS.Elements())
}

func TestPopulateRepositoryTypeBlocks_UnsupportedType(t *testing.T) {
	config := snapshotRepositoryDataSourceModel{
		Name: types.StringValue("test-repo"),
	}
	repo := &elasticsearch.SnapshotRepositoryInfo{
		Type:     "unknown-plugin-type",
		Settings: map[string]any{},
	}

	_, diags := populateRepositoryTypeBlocks(context.Background(), config, repo)
	require.True(t, diags.HasError())
	summary := diags.Errors()[0].Summary()
	assert.Contains(t, summary, "unsupported type")
}
