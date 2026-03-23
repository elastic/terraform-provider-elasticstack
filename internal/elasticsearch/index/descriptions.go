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
