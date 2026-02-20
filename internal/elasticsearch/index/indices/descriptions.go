package indices

import _ "embed"

//go:embed descriptions/target.md
var targetDescription string

//go:embed descriptions/codec.md
var codecDescription string

//go:embed descriptions/shard_check_on_startup.md
var shardCheckOnStartupDescription string

//go:embed descriptions/final_pipeline.md
var finalPipelineDescription string

//go:embed descriptions/indexing_slowlog_source.md
var indexingSlowlogSourceDescription string

//go:embed descriptions/deletion_protection.md
var deletionProtectionDescription string

//go:embed descriptions/wait_for_active_shards.md
var waitForActiveShardsDescription string

//go:embed descriptions/master_timeout.md
var masterTimeoutDescription string

//go:embed descriptions/mappings.md
var mappingsDescription string
