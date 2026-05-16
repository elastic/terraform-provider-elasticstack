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

package dashboard

import _ "embed"

//go:embed descriptions/dashboard_filters.md
var dashboardFiltersDescription string

//go:embed descriptions/dashboard_filter_json.md
var dashboardFilterJSONDescription string

//go:embed descriptions/pinned_panels.md
var pinnedPanelsDescription string

//go:embed descriptions/pinned_panel_control_note.md
var pinnedPanelControlNote string

//go:embed descriptions/pinned_panel_type.md
var pinnedPanelTypeDescription string
