package ingest

import _ "embed"

//go:embed descriptions/pipeline_on_failure.md
var ingestPipelineOnFailureDescription string

//go:embed descriptions/pipeline_processors.md
var ingestPipelineProcessorsDescription string

//go:embed descriptions/append_media_type.md
var processorAppendMediaTypeDescription string

//go:embed descriptions/append_data_source.md
var processorAppendDataSourceDescription string

//go:embed descriptions/circle_data_source.md
var processorCircleDataSourceDescription string

//go:embed descriptions/community_id_seed.md
var communityIDSeedDescription string

//go:embed descriptions/dot_expander_data_source.md
var processorDotExpanderDataSourceDescription string

//go:embed descriptions/drop_data_source.md
var processorDropDataSourceDescription string

//go:embed descriptions/enrich_data_source.md
var processorEnrichDataSourceDescription string

//go:embed descriptions/fail_data_source.md
var processorFailDataSourceDescription string

//go:embed descriptions/fingerprint_data_source.md
var processorFingerprintDataSourceDescription string

//go:embed descriptions/geoip_database_file.md
var processorGeoIPDatabaseFileDescription string

//go:embed descriptions/gsub_data_source.md
var processorGsubDataSourceDescription string

//go:embed descriptions/html_strip_data_source.md
var processorHTMLStripDataSourceDescription string

//go:embed descriptions/join_data_source.md
var processorJoinDataSourceDescription string

//go:embed descriptions/json_add_to_root_conflict.md
var processorJSONAddToRootConflictDescription string

//go:embed descriptions/json_data_source.md
var processorJSONDataSourceDescription string

//go:embed descriptions/kv_data_source.md
var processorKVDataSourceDescription string

//go:embed descriptions/lowercase_data_source.md
var processorLowercaseDataSourceDescription string

//go:embed descriptions/pipeline_data_source.md
var processorPipelineDataSourceDescription string

//go:embed descriptions/registered_domain_data_source.md
var processorRegisteredDomainDataSourceDescription string

//go:embed descriptions/remove_data_source.md
var processorRemoveDataSourceDescription string

//go:embed descriptions/rename_data_source.md
var processorRenameDataSourceDescription string

//go:embed descriptions/reroute_data_source.md
var processorRerouteDataSourceDescription string

//go:embed descriptions/script_data_source.md
var processorScriptDataSourceDescription string

//go:embed descriptions/set_data_source.md
var processorSetDataSourceDescription string

//go:embed descriptions/set_security_user_data_source.md
var processorSetSecurityUserDataSourceDescription string

//go:embed descriptions/sort_data_source.md
var processorSortDataSourceDescription string

//go:embed descriptions/split_data_source.md
var processorSplitDataSourceDescription string

//go:embed descriptions/trim_data_source.md
var processorTrimDataSourceDescription string

//go:embed descriptions/uppercase_data_source.md
var processorUppercaseDataSourceDescription string

//go:embed descriptions/uri_parts_data_source.md
var processorURIPartsDataSourceDescription string

//go:embed descriptions/urldecode_data_source.md
var processorURLDecodeDataSourceDescription string

//go:embed descriptions/user_agent_data_source.md
var processorUserAgentDataSourceDescription string
