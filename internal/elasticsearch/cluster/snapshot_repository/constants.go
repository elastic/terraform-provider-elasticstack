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

// Snapshot repository types recognised by the Elasticsearch _snapshot API.
const (
	repoTypeFS    = "fs"
	repoTypeURL   = "url"
	repoTypeGCS   = "gcs"
	repoTypeAzure = "azure"
	repoTypeS3    = "s3"
	repoTypeHDFS  = "hdfs"
)

// Snapshot repository setting keys. These are used both as Terraform schema
// attribute names and as keys in the request/response JSON sent to the
// Elasticsearch _snapshot API.
const (
	settingChunkSize              = "chunk_size"
	settingCompress               = "compress"
	settingMaxSnapshotBytesPerSec = "max_snapshot_bytes_per_sec"
	settingMaxRestoreBytesPerSec  = "max_restore_bytes_per_sec"
	settingReadonly               = "readonly"
	settingMaxNumberOfSnapshots   = "max_number_of_snapshots"
	settingLocation               = "location"
	settingURL                    = "url"
	settingHTTPMaxRetries         = "http_max_retries"
	settingHTTPSocketTimeout      = "http_socket_timeout"
	settingBucket                 = "bucket"
	settingClient                 = "client"
	settingBasePath               = "base_path"
	settingContainer              = "container"
	settingLocationMode           = "location_mode"
	settingEndpoint               = "endpoint"
	settingServerSideEncryption   = "server_side_encryption"
	settingBufferSize             = "buffer_size"
	settingCannedACL              = "canned_acl"
	settingStorageClass           = "storage_class"
	settingPathStyleAccess        = "path_style_access"
	settingURI                    = "uri"
	settingPath                   = "path"
	settingLoadDefaults           = "load_defaults"
)
