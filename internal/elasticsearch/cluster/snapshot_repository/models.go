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
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Data holds the Terraform state for the snapshot repository resource.
type Data struct {
	ID                      types.String `tfsdk:"id"`
	ElasticsearchConnection types.List   `tfsdk:"elasticsearch_connection"`
	Name                    types.String `tfsdk:"name"`
	Verify                  types.Bool   `tfsdk:"verify"`
	Fs                      types.Object `tfsdk:"fs"`
	URL                     types.Object `tfsdk:"url"`
	Azure                   types.Object `tfsdk:"azure"`
	Gcs                     types.Object `tfsdk:"gcs"`
	S3                      types.Object `tfsdk:"s3"`
	Hdfs                    types.Object `tfsdk:"hdfs"`
}

func (d Data) GetID() types.String                    { return d.ID }
func (d Data) GetResourceID() types.String            { return d.Name }
func (d Data) GetElasticsearchConnection() types.List { return d.ElasticsearchConnection }

// CommonSettings holds fields shared across most repository types.
type CommonSettings struct {
	ChunkSize              types.String `tfsdk:"chunk_size"`
	Compress               types.Bool   `tfsdk:"compress"`
	MaxSnapshotBytesPerSec types.String `tfsdk:"max_snapshot_bytes_per_sec"`
	MaxRestoreBytesPerSec  types.String `tfsdk:"max_restore_bytes_per_sec"`
	Readonly               types.Bool   `tfsdk:"readonly"`
}

// CommonStdSettings holds fields shared across standard (non-URL) repositories.
type CommonStdSettings struct {
	MaxNumberOfSnapshots types.Int64 `tfsdk:"max_number_of_snapshots"`
}

// FsSettings is used for the `fs` block.
type FsSettings struct {
	CommonSettings
	CommonStdSettings
	Location types.String `tfsdk:"location"`
}

// URLSettings is used for the `url` block.
type URLSettings struct {
	CommonSettings
	CommonStdSettings
	URL               types.String `tfsdk:"url"`
	HTTPMaxRetries    types.Int64  `tfsdk:"http_max_retries"`
	HTTPSocketTimeout types.String `tfsdk:"http_socket_timeout"`
}

// GcsSettings is used for the `gcs` block.
type GcsSettings struct {
	CommonSettings
	Bucket   types.String `tfsdk:"bucket"`
	Client   types.String `tfsdk:"client"`
	BasePath types.String `tfsdk:"base_path"`
}

// AzureSettings is used for the `azure` block.
type AzureSettings struct {
	CommonSettings
	Container    types.String `tfsdk:"container"`
	Client       types.String `tfsdk:"client"`
	BasePath     types.String `tfsdk:"base_path"`
	LocationMode types.String `tfsdk:"location_mode"`
}

// S3Settings is used for the `s3` block.
type S3Settings struct {
	CommonSettings
	Bucket               types.String `tfsdk:"bucket"`
	Endpoint             types.String `tfsdk:"endpoint"`
	Client               types.String `tfsdk:"client"`
	BasePath             types.String `tfsdk:"base_path"`
	ServerSideEncryption types.Bool   `tfsdk:"server_side_encryption"`
	BufferSize           types.String `tfsdk:"buffer_size"`
	CannedACL            types.String `tfsdk:"canned_acl"`
	StorageClass         types.String `tfsdk:"storage_class"`
	PathStyleAccess      types.Bool   `tfsdk:"path_style_access"`
}

// HdfsSettings is used for the `hdfs` block.
type HdfsSettings struct {
	CommonSettings
	URI          types.String `tfsdk:"uri"`
	Path         types.String `tfsdk:"path"`
	LoadDefaults types.Bool   `tfsdk:"load_defaults"`
}
