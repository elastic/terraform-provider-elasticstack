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

package synthetics

import (
	"fmt"
	"strings"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const (
	MetadataPrefix = "_kibana_synthetics_"
)

// GetCompositeID parses a composite ID and returns the parsed components
func GetCompositeID(id string) (*clients.CompositeID, diag.Diagnostics) {
	compositeID, sdkDiag := clients.CompositeIDFromStr(id)
	dg := diag.Diagnostics{}
	if sdkDiag.HasError() {
		dg.AddError(fmt.Sprintf("Failed to parse monitor ID %s", id), fmt.Sprintf("Resource ID must have following format: <cluster_uuid>/<resource identifier>. Current value: %s", id))
		return nil, dg
	}
	return compositeID, dg
}

// TryReadCompositeID parses a composite ID when the id contains "/".
// Plain IDs are treated as legacy non-composite identifiers and return nil
// without diagnostics. Kibana synthetics IDs allow at most one slash (for example
// "<space_id>/<monitor_id>" or legacy "/<monitor_id>"); additional slashes are rejected.
func TryReadCompositeID(id string) (*clients.CompositeID, diag.Diagnostics) {
	if !strings.Contains(id, "/") {
		return nil, diag.Diagnostics{}
	}
	if strings.Count(id, "/") > 1 {
		dg := diag.Diagnostics{}
		dg.AddError(
			fmt.Sprintf("Failed to parse monitor ID %s", id),
			fmt.Sprintf("Resource ID must have following format: <cluster_uuid>/<resource identifier>. Current value: %s", id),
		)
		return nil, dg
	}
	return GetCompositeID(id)
}

// ValueStringSlice converts []types.String to []string.
//
// Deprecated: use typeutils.ValueStringSlice directly.
func ValueStringSlice(v []types.String) []string {
	return typeutils.ValueStringSlice(v)
}

// StringSliceValue converts []string to []types.String.
//
// Deprecated: use typeutils.StringSliceValue directly.
func StringSliceValue(v []string) []types.String {
	return typeutils.StringSliceValue(v)
}
