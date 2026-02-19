package security_detection_rule

import _ "embed"

//go:embed descriptions/resource.md
var securityDetectionRuleMarkdownDescription string

//go:embed descriptions/filters.md
var filtersMarkdownDescription string

//go:embed descriptions/missing_fields_strategy.md
var missingFieldsStrategyMarkdownDescription string

//go:embed descriptions/building_block_type.md
var buildingBlockTypeMarkdownDescription string
