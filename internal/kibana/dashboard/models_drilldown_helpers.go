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

import "github.com/hashicorp/terraform-plugin-framework/types"

// Default values for optional booleans on drilldowns are represented as package-level constants in panel
// packages where needed. K presets these when attributes are omitted; the provider nulls matching values
// on import where appropriate.

const (
	// Kibana defaults for URL drilldown booleans (panels that still live in legacy mapping).
	drilldownURLEncodeURLDefault    = true
	drilldownURLOpenInNewTabDefault = false
)

// panelDrilldownBoolImportPreserving maps optional API booleans on import: nil or value equal to
// the server-side default becomes null in Terraform state so practitioners can omit those
// attributes without drift.
func panelDrilldownBoolImportPreserving(api *bool, serverDefault bool) types.Bool {
	if api == nil {
		return types.BoolNull()
	}
	if *api == serverDefault {
		return types.BoolNull()
	}
	return types.BoolValue(*api)
}
