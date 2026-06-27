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
	"testing"

	esclients "github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/stretchr/testify/require"
)

func TestExtractSettings(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	fsObj, _ := types.ObjectValueFrom(ctx, fsAttrTypes(), FsSettings{
		CommonSettings: CommonSettings{
			Compress: types.BoolValue(true),
		},
		CommonStdSettings: CommonStdSettings{MaxNumberOfSnapshots: types.Int64Value(500)},
		Location:          types.StringValue("/tmp"),
	})

	urlObj, _ := types.ObjectValueFrom(ctx, urlAttrTypes(), URLSettings{
		CommonSettings: CommonSettings{
			Compress: types.BoolValue(true),
		},
		CommonStdSettings: CommonStdSettings{MaxNumberOfSnapshots: types.Int64Value(500)},
		URL:               types.StringValue("file:///tmp"),
	})

	cases := []struct {
		name        string
		data        Data
		wantType    string
		wantSetting string
		wantErr     bool
	}{
		{
			name:        "fs",
			data:        Data{Fs: fsObj},
			wantType:    "fs",
			wantSetting: "location",
		},
		{
			name:        "url",
			data:        Data{URL: urlObj},
			wantType:    "url",
			wantSetting: "url",
		},
		{
			name:    "none",
			data:    Data{},
			wantErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			repoType, settings, diags := extractSettings(ctx, tc.data)
			if tc.wantErr {
				require.True(t, diags.HasError())
				return
			}
			require.False(t, diags.HasError())
			require.Equal(t, tc.wantType, repoType)
			require.Contains(t, settings, tc.wantSetting)
		})
	}
}

func TestStrSettingNull(t *testing.T) {
	t.Parallel()

	settings := map[string]any{"present": "value", "empty": ""}

	require.True(t, strSettingNull(settings, "missing").IsNull())
	require.Equal(t, "value", strSettingNull(settings, "present").ValueString())
	require.Empty(t, strSettingNull(settings, "empty").ValueString())
}

func TestFsToSettingsDefaults(t *testing.T) {
	t.Parallel()

	fs := FsSettings{
		CommonSettings: CommonSettings{
			ChunkSize:              types.StringNull(),
			Compress:               types.BoolValue(true),
			MaxSnapshotBytesPerSec: types.StringNull(),
			MaxRestoreBytesPerSec:  types.StringNull(),
			Readonly:               types.BoolValue(false),
		},
		CommonStdSettings: CommonStdSettings{MaxNumberOfSnapshots: types.Int64Value(500)},
		Location:          types.StringValue("/tmp"),
	}

	m := fsToSettings(fs)
	require.Equal(t, "/tmp", m["location"])
	require.Equal(t, true, m["compress"])
	require.Equal(t, false, m["readonly"])
	require.Equal(t, int64(500), m["max_number_of_snapshots"])
	require.NotContains(t, m, "chunk_size")
	require.NotContains(t, m, "max_snapshot_bytes_per_sec")
}

func TestS3ToSettingsWithDefaults(t *testing.T) {
	t.Parallel()

	s3 := S3Settings{
		CommonSettings: CommonSettings{
			Compress: types.BoolValue(true),
			Readonly: types.BoolValue(false),
		},
		Bucket:               types.StringValue("mybucket"),
		Endpoint:             types.StringNull(),
		Client:               types.StringValue("default"),
		BasePath:             types.StringNull(),
		ServerSideEncryption: types.BoolValue(false),
		BufferSize:           types.StringNull(),
		CannedACL:            types.StringValue("private"),
		StorageClass:         types.StringValue("standard"),
		PathStyleAccess:      types.BoolValue(false),
	}

	m := s3ToSettings(s3)
	require.Equal(t, "mybucket", m["bucket"])
	require.Equal(t, true, m["compress"])
	require.Equal(t, false, m["readonly"])
	require.Equal(t, false, m["server_side_encryption"])
	require.Equal(t, false, m["path_style_access"])
	require.Equal(t, "default", m["client"])
	require.Equal(t, "private", m["canned_acl"])
	require.Equal(t, "standard", m["storage_class"])
	require.NotContains(t, m, "endpoint")
	require.NotContains(t, m, "base_path")
}

func TestS3ToSettingsWithEndpoint(t *testing.T) {
	t.Parallel()

	s3 := S3Settings{
		CommonSettings: CommonSettings{
			Compress: types.BoolValue(true),
			Readonly: types.BoolValue(false),
		},
		Bucket:               types.StringValue("mybucket"),
		Endpoint:             types.StringValue("https://minio.example.com:9000"),
		Client:               types.StringValue("default"),
		BasePath:             types.StringNull(),
		ServerSideEncryption: types.BoolValue(false),
		BufferSize:           types.StringNull(),
		CannedACL:            types.StringValue("private"),
		StorageClass:         types.StringValue("standard"),
		PathStyleAccess:      types.BoolValue(true),
	}

	m := s3ToSettings(s3)
	require.Equal(t, "mybucket", m["bucket"])
	require.Equal(t, true, m["compress"])
	require.Equal(t, false, m["readonly"])
	require.Equal(t, false, m["server_side_encryption"])
	require.Equal(t, true, m["path_style_access"])
	require.Equal(t, "default", m["client"])
	require.Equal(t, "private", m["canned_acl"])
	require.Equal(t, "standard", m["storage_class"])
	require.Contains(t, m, "endpoint")
	require.Equal(t, "https://minio.example.com:9000", m["endpoint"])
	require.NotContains(t, m, "base_path")
}

func TestSettingsToS3StateInheritance(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	cases := []struct {
		name                string
		apiSettings         map[string]any
		stateS3             S3Settings
		stateS3Null         bool
		wantEndpoint        types.String
		wantPathStyleAccess bool
	}{
		{
			name: "ES echoes both",
			apiSettings: map[string]any{
				"bucket":            "api-bucket",
				"endpoint":          "https://api.example",
				"path_style_access": true,
			},
			stateS3: s3SettingsForState(
				types.StringValue("https://prior.example"),
				types.BoolValue(false),
			),
			wantEndpoint:        types.StringValue("https://api.example"),
			wantPathStyleAccess: true,
		},
		{
			name: "ES omits both, prior state populated",
			apiSettings: map[string]any{
				"bucket": "api-bucket",
			},
			stateS3: s3SettingsForState(
				types.StringValue("https://prior.example"),
				types.BoolValue(true),
			),
			wantEndpoint:        types.StringValue("https://prior.example"),
			wantPathStyleAccess: true,
		},
		{
			name: "ES omits both, no prior state (import)",
			apiSettings: map[string]any{
				"bucket": "api-bucket",
			},
			stateS3Null:         true,
			wantEndpoint:        types.StringNull(),
			wantPathStyleAccess: false,
		},
		{
			name: "ES omits both, prior state has null fields",
			apiSettings: map[string]any{
				"bucket": "api-bucket",
			},
			stateS3: s3SettingsForState(
				types.StringNull(),
				types.BoolValue(false),
			),
			wantEndpoint:        types.StringNull(),
			wantPathStyleAccess: false,
		},
		{
			name: "ES echoes endpoint but not path_style_access",
			apiSettings: map[string]any{
				"bucket":   "api-bucket",
				"endpoint": "https://api.example",
			},
			stateS3: s3SettingsForState(
				types.StringValue("https://prior.example"),
				types.BoolValue(true),
			),
			wantEndpoint:        types.StringValue("https://api.example"),
			wantPathStyleAccess: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			repo := &esclients.SnapshotRepositoryInfo{
				Type:     "s3",
				Settings: tc.apiSettings,
			}

			state := Data{}
			if !tc.stateS3Null {
				state.S3 = mustS3Object(ctx, t, tc.stateS3)
			}

			result, diags := settingsToS3(ctx, repo, state)
			require.False(t, diags.HasError(), diags.Errors())

			var got S3Settings
			require.False(t, result.As(ctx, &got, basetypes.ObjectAsOptions{}).HasError())

			if tc.wantEndpoint.IsNull() {
				require.True(t, got.Endpoint.IsNull())
			} else {
				require.Equal(t, tc.wantEndpoint.ValueString(), got.Endpoint.ValueString())
			}
			require.Equal(t, tc.wantPathStyleAccess, got.PathStyleAccess.ValueBool())
		})
	}
}

func s3SettingsForState(endpoint types.String, pathStyleAccess types.Bool) S3Settings {
	return S3Settings{
		CommonSettings: CommonSettings{
			Compress: types.BoolValue(true),
			Readonly: types.BoolValue(false),
		},
		Bucket:               types.StringValue("state-bucket"),
		Endpoint:             endpoint,
		Client:               types.StringValue("default"),
		BasePath:             types.StringNull(),
		ServerSideEncryption: types.BoolValue(false),
		BufferSize:           types.StringNull(),
		CannedACL:            types.StringNull(),
		StorageClass:         types.StringNull(),
		PathStyleAccess:      pathStyleAccess,
	}
}

func mustS3Object(ctx context.Context, t *testing.T, s3 S3Settings) types.Object {
	t.Helper()
	obj, diags := types.ObjectValueFrom(ctx, s3AttrTypes(), s3)
	require.False(t, diags.HasError(), diags.Errors())
	return obj
}

// TestSettingsToGcsAbsentFieldsAreNull verifies that optional string fields
// absent in the API response are mapped to null (not empty string). This
// prevents the "inconsistent result after apply" error described in issue #3709,
// where the plan holds "" but the post-apply read returns null.
func TestSettingsToGcsAbsentFieldsAreNull(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	repo := &esclients.SnapshotRepositoryInfo{
		Type: "gcs",
		Settings: map[string]any{
			"bucket": "my-bucket",
		},
	}

	obj, diags := settingsToGcs(ctx, repo)
	require.False(t, diags.HasError(), diags.Errors())

	var gcs GcsSettings
	require.False(t, obj.As(ctx, &gcs, basetypes.ObjectAsOptions{}).HasError())

	require.True(t, gcs.ChunkSize.IsNull(), "chunk_size should be null when absent from API")
	require.True(t, gcs.MaxSnapshotBytesPerSec.IsNull(), "max_snapshot_bytes_per_sec should be null when absent from API")
	require.True(t, gcs.MaxRestoreBytesPerSec.IsNull(), "max_restore_bytes_per_sec should be null when absent from API")
}
