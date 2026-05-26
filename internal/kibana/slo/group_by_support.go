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

package slo

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// resolveGroupBySupport validates group_by version requirements against apiClient,
// mutates apiModel.GroupBy to nil when the server does not support it, and returns
// the supportsMultipleGroupBy flag. Appends diagnostics on error.
func resolveGroupBySupport(
	ctx context.Context,
	apiClient *clients.KibanaScopedClient,
	apiModel *models.Slo,
	diags *diag.Diagnostics,
) bool {
	supportsGroupBy, groupByDiags := apiClient.EnforceMinVersion(ctx, SLOSupportsGroupByMinVersion)
	diags.Append(groupByDiags...)
	if diags.HasError() {
		return false
	}
	if !supportsGroupBy {
		if len(apiModel.GroupBy) > 0 {
			diags.AddError(
				"Unsupported Elastic Stack version",
				"group_by is not supported in this version of the Elastic Stack. group_by requires "+SLOSupportsGroupByMinVersion.String()+" or higher.",
			)
			return false
		}
		apiModel.GroupBy = nil
		return false
	}

	supportsMultipleGroupBy, groupByDiags := apiClient.EnforceMinVersion(ctx, SLOSupportsMultipleGroupByMinVersion)
	diags.Append(groupByDiags...)
	if diags.HasError() {
		return false
	}
	if len(apiModel.GroupBy) > 1 && !supportsMultipleGroupBy {
		diags.AddError(
			"Unsupported Elastic Stack version",
			"multiple group_by fields are not supported in this version of the Elastic Stack. Multiple group_by fields requires "+SLOSupportsMultipleGroupByMinVersion.String(),
		)
		return false
	}
	return supportsMultipleGroupBy
}
