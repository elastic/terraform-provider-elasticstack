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

package templateutil

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/index"
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/index/aliasutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/index/datastreamoptions"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// ExpandTemplateCore builds a *models.Template from the shared template block
// fields (aliases, mappings, settings, data_stream_options). Callers may extend
// the result with package-specific fields such as lifecycle before returning.
func ExpandTemplateCore(
	ctx context.Context,
	alias types.Set,
	mappings index.MappingsValue,
	settings customtypes.IndexSettingsValue,
	dataStreamOptions types.Object,
) (*models.Template, diag.Diagnostics) {
	var diags diag.Diagnostics
	t := &models.Template{}

	if !alias.IsNull() && !alias.IsUnknown() {
		aliases, d := aliasutil.ExpandAliasSet(ctx, alias)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		t.Aliases = aliases
	}

	if !mappings.IsNull() && !mappings.IsUnknown() {
		s := strings.TrimSpace(mappings.ValueString())
		if s != "" {
			maps := make(map[string]any)
			if err := json.Unmarshal([]byte(s), &maps); err != nil {
				diags.AddError("Invalid template.mappings JSON", err.Error())
				return nil, diags
			}
			t.Mappings = maps
		}
	}

	if !settings.IsNull() && !settings.IsUnknown() {
		s := strings.TrimSpace(settings.ValueString())
		if s != "" {
			sets := make(map[string]any)
			if err := json.Unmarshal([]byte(s), &sets); err != nil {
				diags.AddError("Invalid template.settings JSON", err.Error())
				return nil, diags
			}
			t.Settings = sets
		}
	}

	if !dataStreamOptions.IsNull() && !dataStreamOptions.IsUnknown() {
		dso, d := datastreamoptions.Expand(dataStreamOptions)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		t.DataStreamOptions = dso
	}

	return t, diags
}

// DecodeTemplateObject unmarshals a Terraform types.Object into the provided
// model pointer when the object is known. It returns empty diagnostics for
// null or unknown objects. This helper removes the IsNull/IsUnknown and
// ObjectAs boilerplate duplicated in template block expand functions.
func DecodeTemplateObject(ctx context.Context, obj types.Object, model any) diag.Diagnostics {
	var diags diag.Diagnostics
	if obj.IsNull() || obj.IsUnknown() {
		return diags
	}
	diags.Append(obj.As(ctx, model, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true})...)
	return diags
}
