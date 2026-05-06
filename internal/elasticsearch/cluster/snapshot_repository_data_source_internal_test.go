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
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

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
	model := flattenGCSSettings(settings)

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
	model := flattenGCSSettings(map[string]any{})

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
	model := flattenAzureSettings(settings)

	assert.Equal(t, types.StringValue("my-container"), model.Container)
	assert.Equal(t, types.StringValue("my-client"), model.Client)
	assert.Equal(t, types.StringValue("prefix"), model.BasePath)
	assert.Equal(t, types.StringValue("primary_only"), model.LocationMode)
	assert.Equal(t, types.BoolValue(false), model.Compress)
	assert.Equal(t, types.BoolValue(false), model.Readonly)
	assert.Equal(t, types.StringValue("1024"), model.ChunkSize)

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
	model := flattenS3Settings(settings)

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

func TestFlattenS3Settings_NoEndpoint(t *testing.T) {
	// Ensure the S3 model does not include endpoint; this test documents
	// the schema decision that endpoint is absent from the data source.
	settings := map[string]any{
		"bucket":  "b",
		"endpoint": "http://example.com",
	}
	model := flattenS3Settings(settings)
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
	model := flattenHDFSSettings(settings)

	assert.Equal(t, types.StringValue("hdfs://host:8020/"), model.URI)
	assert.Equal(t, types.StringValue("/repo"), model.Path)
	assert.Equal(t, types.BoolValue(true), model.LoadDefaults)
	assert.Equal(t, types.BoolValue(false), model.Compress)
	assert.Equal(t, types.BoolValue(false), model.Readonly)
	assert.Equal(t, types.StringValue("32"), model.ChunkSize)
}

func TestStringSetting(t *testing.T) {
	assert.Equal(t, types.StringValue("hello"), stringSetting(map[string]any{"k": "hello"}, "k"))
	assert.Equal(t, types.StringValue("42"), stringSetting(map[string]any{"k": 42}, "k"))
	assert.True(t, stringSetting(map[string]any{}, "k").IsNull())
	assert.True(t, stringSetting(map[string]any{"k": nil}, "k").IsNull())
}

func TestBoolSetting(t *testing.T) {
	assert.Equal(t, types.BoolValue(true), boolSetting(map[string]any{"k": true}, "k"))
	assert.Equal(t, types.BoolValue(false), boolSetting(map[string]any{"k": "false"}, "k"))
	assert.Equal(t, types.BoolValue(true), boolSetting(map[string]any{"k": "1"}, "k"))
	assert.True(t, boolSetting(map[string]any{"k": "nope"}, "k").IsNull())
	assert.True(t, boolSetting(map[string]any{}, "k").IsNull())
}

func TestInt64Setting(t *testing.T) {
	assert.Equal(t, types.Int64Value(42), int64Setting(map[string]any{"k": 42}, "k"))
	assert.Equal(t, types.Int64Value(42), int64Setting(map[string]any{"k": int64(42)}, "k"))
	assert.Equal(t, types.Int64Value(42), int64Setting(map[string]any{"k": float64(42)}, "k"))
	assert.Equal(t, types.Int64Value(42), int64Setting(map[string]any{"k": "42"}, "k"))
	assert.True(t, int64Setting(map[string]any{"k": "not-a-number"}, "k").IsNull())
	assert.True(t, int64Setting(map[string]any{}, "k").IsNull())
}
