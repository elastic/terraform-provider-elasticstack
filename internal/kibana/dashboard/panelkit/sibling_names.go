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

package panelkit

// typedSiblingPanelConfigBlockNames holds panel-level attribute names configured as mutually
// exclusive siblings: optional `config_json`, each registered handler's typed `*_config` block name,
// and unmigrated panel blocks (`vis_config`, `lens_dashboard_app_config`, `discover_session_config`).
var typedSiblingPanelConfigBlockNames []string

// SetTypedSiblingPanelConfigBlockNames configures the mutually exclusive sibling name list used for
// panel-level ConflictsWith / MarkdownDescription hints. Dashboard init calls this after
// registering handlers and appending non-*_config panel config siblings used by unmigrated panels.
func SetTypedSiblingPanelConfigBlockNames(names []string) {
	typedSiblingPanelConfigBlockNames = append([]string(nil), names...)
}

// TypedSiblingPanelConfigBlockNames returns a copy of the mutually exclusive sibling name list.
func TypedSiblingPanelConfigBlockNames() []string {
	out := make([]string, len(typedSiblingPanelConfigBlockNames))
	copy(out, typedSiblingPanelConfigBlockNames)
	return out
}
