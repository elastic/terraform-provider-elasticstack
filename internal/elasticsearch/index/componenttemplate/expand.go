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

package componenttemplate

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/index/templateutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// expandFromData builds a models.ComponentTemplate from a Data plan/state value.
func expandFromData(ctx context.Context, d Data) (models.ComponentTemplate, diag.Diagnostics) {
	var diags diag.Diagnostics
	out := models.ComponentTemplate{
		Name: d.Name.ValueString(),
	}

	out.Meta = templateutil.ExpandMetadataJSON(d.Metadata, &diags)
	if diags.HasError() {
		return out, diags
	}

	if !d.Template.IsNull() && !d.Template.IsUnknown() {
		tpl, d2 := expandTemplateBlock(ctx, d.Template)
		diags.Append(d2...)
		if diags.HasError() {
			return out, diags
		}
		out.Template = tpl
	}

	if !d.Version.IsNull() && !d.Version.IsUnknown() {
		v := d.Version.ValueInt64()
		out.Version = &v
	}

	return out, diags
}

// expandTemplateBlock expands the template block object to *models.Template.
func expandTemplateBlock(ctx context.Context, obj types.Object) (*models.Template, diag.Diagnostics) {
	var tm TemplateModel
	if d := templateutil.DecodeTemplateObject(ctx, obj, &tm); d.HasError() {
		return nil, d
	}

	return templateutil.ExpandTemplateCore(ctx, tm.Alias, tm.Mappings, tm.Settings, tm.DataStreamOptions)
}
