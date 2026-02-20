package datafeed

import _ "embed"

//go:embed descriptions/schema.md
var schemaMarkdownDescription string

//go:embed descriptions/datafeed_id.md
var datafeedIDMarkdownDescription string

//go:embed descriptions/query.md
var queryMarkdownDescription string

//go:embed descriptions/aggregations.md
var aggregationsMarkdownDescription string

//go:embed descriptions/script_fields.md
var scriptFieldsMarkdownDescription string

//go:embed descriptions/scroll_size.md
var scrollSizeMarkdownDescription string

//go:embed descriptions/frequency.md
var frequencyMarkdownDescription string

//go:embed descriptions/query_delay.md
var queryDelayMarkdownDescription string

//go:embed descriptions/max_empty_searches.md
var maxEmptySearchesMarkdownDescription string

//go:embed descriptions/chunking_config.md
var chunkingConfigMarkdownDescription string

//go:embed descriptions/chunking_mode.md
var chunkingModeMarkdownDescription string

//go:embed descriptions/delayed_data_check_config.md
var delayedDataCheckConfigMarkdownDescription string

//go:embed descriptions/check_window.md
var checkWindowMarkdownDescription string

//go:embed descriptions/expand_wildcards.md
var expandWildcardsMarkdownDescription string
