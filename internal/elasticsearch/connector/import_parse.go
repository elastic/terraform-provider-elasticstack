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

package connector

import (
	"strings"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

const invalidImportIDSummary = "Invalid import ID"

const invalidImportIDDetail = "Import ID must be a non-empty connector_id or composite <cluster_uuid>/<connector_id>."

func parseConnectorImportID(raw string) (string, diag.Diagnostics) {
	var diags diag.Diagnostics

	importID := strings.TrimSpace(raw)
	importID = strings.TrimRight(importID, "/")
	importID = strings.TrimLeft(importID, "/")
	if importID == "" {
		diags.AddError(invalidImportIDSummary, invalidImportIDDetail)
		return "", diags
	}

	if strings.Contains(importID, "/") {
		compID, compDiags := clients.CompositeIDFromStr(importID)
		diags.Append(compDiags...)
		if diags.HasError() {
			return "", diags
		}
		return compID.ResourceID, diags
	}

	return importID, diags
}
