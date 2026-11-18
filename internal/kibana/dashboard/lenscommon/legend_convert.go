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

import (
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// LegendSizeTruncateVisibilityFromAPI maps the legend fields shared by every lens chart type
// (size, truncateAfterLines, visibility) from an API response into Terraform values. Callers
// convert their concrete enum-typed pointers to *string before calling this.
func LegendSizeTruncateVisibilityFromAPI(size *string, truncateAfterLines *float32, visibility *string) (sizeOut types.String, truncateAfterLinesOut types.Int64, visibilityOut types.String) {
	if size != nil {
		sizeOut = types.StringValue(*size)
	} else {
		sizeOut = types.StringNull()
	}
	if truncateAfterLines != nil {
		truncateAfterLinesOut = types.Int64Value(int64(*truncateAfterLines))
	} else {
		truncateAfterLinesOut = types.Int64Null()
	}
	if visibility != nil {
		visibilityOut = types.StringValue(*visibility)
	} else {
		visibilityOut = types.StringNull()
	}
	return
}

// LegendSizeTruncateVisibilityToAPI maps the legend fields shared by every lens chart type (size,
// truncateAfterLines, visibility) from the Terraform model into API pointer fields. Size is
// required; if it is unknown, an error is appended to diags using missingSizeSummary/Detail and
// sizeOut is left nil. Callers convert the returned *string values to their concrete enum types.
func LegendSizeTruncateVisibilityToAPI(
	size, visibility types.String,
	truncateAfterLines types.Int64,
	missingSizeSummary, missingSizeDetail string,
	diags *diag.Diagnostics,
) (sizeOut *string, truncateAfterLinesOut *float32, visibilityOut *string) {
	if typeutils.IsKnown(size) {
		sizeOut = new(size.ValueString())
	} else {
		diags.AddError(missingSizeSummary, missingSizeDetail)
	}
	if typeutils.IsKnown(truncateAfterLines) {
		truncateAfterLinesOut = new(float32(truncateAfterLines.ValueInt64()))
	}
	if typeutils.IsKnown(visibility) {
		visibilityOut = new(visibility.ValueString())
	}
	return
}
