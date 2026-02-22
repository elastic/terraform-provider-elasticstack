package index

import _ "embed"

//go:embed descriptions/component_template_resource.md
var componentTemplateResourceDescription string

//go:embed descriptions/component_template_alias_name.md
var componentTemplateAliasNameDescription string

//go:embed descriptions/index_template_mappings.md
var indexTemplateMappingsDescription string

//go:embed descriptions/component_template_settings.md
var componentTemplateSettingsDescription string

//go:embed descriptions/index_template_resource.md
var indexTemplateResourceDescription string

//go:embed descriptions/ilm_resource.md
var ilmResourceDescription string

//go:embed descriptions/ilm_set_priority_action.md
var ilmSetPriorityActionDescription string
