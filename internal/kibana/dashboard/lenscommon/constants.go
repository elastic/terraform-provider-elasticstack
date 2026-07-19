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

package lenscommon

// Terraform schema attribute keys and Lens panel model keys that are reused
// across drilldowns, presentation, and schema definitions. They are extracted
// to constants so the goconst linter is satisfied with a single source of
// truth for the shared identifiers.
const (
	attrType          = "type"
	attrURL           = "url"
	attrLabel         = "label"
	attrValue         = "value"
	attrVisible       = "visible"
	attrMode          = "mode"
	attrTrigger       = "trigger"
	attrOpenInNewTab  = "open_in_new_tab"
	attrColumn        = "column"
	attrFormatJSON    = "format_json"
	attrAlignment     = "alignment"
	attrMetric        = "metric"
	attrDirection     = "direction"
	attrAlignRight    = "right"
	attrRefID         = "ref_id"
	attrTimeRange     = "time_range"
	colorTypeAuto     = "auto"
	sortDirectionDesc = "desc"

	// JSONNullString is the JSON representation of a null value, used to detect
	// empty union types returned by the Kibana API.
	JSONNullString = "null"

	// Drilldown discriminator values stored on Lens panels.
	drilldownTypeDashboard = "dashboard_drilldown"
	drilldownTypeDiscover  = "discover_drilldown"
	drilldownTypeURL       = "url_drilldown"

	// Shared markdown description re-used in multiple schema attribute definitions.
	drilldownLabelDescription = "Human-readable drilldown label."
)
