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

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// lensQueryESQLMode returns whether a Lens chart's optional `query` object selects
// ES|QL mode (i.e. `query` is omitted, or both `expression` and `language` are
// null). ok is false when the configuration is still unknown and validation
// should defer.
func lensQueryESQLMode(ctx context.Context, config tfsdk.Config, attrPath path.Path, diags *diag.Diagnostics) (esqlMode bool, ok bool) {
	var queryObj types.Object
	diags.Append(config.GetAttribute(ctx, attrPath.AtName("query"), &queryObj)...)
	if diags.HasError() {
		return false, false
	}
	if queryObj.IsUnknown() {
		return false, false
	}
	if queryObj.IsNull() {
		return true, true
	}

	var lang, expr types.String
	diags.Append(config.GetAttribute(ctx, attrPath.AtName("query").AtName("language"), &lang)...)
	diags.Append(config.GetAttribute(ctx, attrPath.AtName("query").AtName("expression"), &expr)...)
	if diags.HasError() {
		return false, false
	}
	return lang.IsNull() && expr.IsNull(), true
}
