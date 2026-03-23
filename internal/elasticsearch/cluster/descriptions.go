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

import _ "embed"

//go:embed descriptions/settings_resource.md
var settingsResourceDescription string

//go:embed descriptions/slm_resource.md
var slmResourceDescription string

//go:embed descriptions/snapshot_repository_resource.md
var snapshotRepositoryResourceDescription string

//go:embed descriptions/snapshot_repository_location_mode.md
var snapshotRepositoryLocationModeDescription string

//go:embed descriptions/snapshot_repository_url_attr.md
var snapshotRepositoryURLAttrDescription string

//go:embed descriptions/snapshot_repository_gcs_attr.md
var snapshotRepositoryGCSAttrDescription string

//go:embed descriptions/snapshot_repository_azure_attr.md
var snapshotRepositoryAzureAttrDescription string

//go:embed descriptions/snapshot_repository_s3_attr.md
var snapshotRepositoryS3AttrDescription string

//go:embed descriptions/snapshot_repository_hdfs_attr.md
var snapshotRepositoryHDFSAttrDescription string
