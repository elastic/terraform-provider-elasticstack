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
