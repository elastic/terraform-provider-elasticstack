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
