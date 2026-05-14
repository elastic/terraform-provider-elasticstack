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

package contracttest

import (
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"sort"

	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panel/iface"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func appendOuterSchemaIssues(handler iface.Handler, issues *[]string) {
	a := handler.SchemaAttribute()
	sn, ok := a.(schema.SingleNestedAttribute)
	if !ok {
		*issues = append(*issues, "[Schema] SchemaAttribute must be a SingleNestedAttribute")
		return
	}
	if !sn.Optional || sn.Required || sn.Computed {
		*issues = append(*issues, "[Schema] outer SingleNested must be Optional=true, Required=false, Computed=false")
	}
}

func appendRequiredJSONPresenceIssues(handler iface.Handler, fixtureJSON string, issues *[]string) {
	sn, ok := handler.SchemaAttribute().(schema.SingleNestedAttribute)
	if !ok {
		return
	}
	lp := collectLeafPaths(sn)
	cfg, err := parseFixtureConfig(fixtureJSON)
	if err != nil {
		*issues = append(*issues, fmt.Sprintf("[Schema] parse fixture.config: %v", err))
		return
	}
	for _, rp := range lp.required {
		if !fixtureHasTerraformNestedKey(cfg, rp) {
			*issues = append(*issues, fmt.Sprintf("[Schema] required terraform path %#v absent from fixture.config payload", rp))
		}
	}
}

// appendValidateRequiredZeroIssues validates each shallow required attribute independently (one key per ValidatePanelConfig call).
func appendValidateRequiredZeroIssues(ctx context.Context, handler iface.Handler, fixtureJSON string, issues *[]string) {
	sn, ok := handler.SchemaAttribute().(schema.SingleNestedAttribute)
	if !ok {
		return
	}
	lp := collectLeafPaths(sn)
	cfg, err := parseFixtureConfig(fixtureJSON)
	if err != nil {
		*issues = append(*issues, fmt.Sprintf("[Schema] parse fixture.config for validation: %v", err))
		return
	}
	shallowKeys := flattenShallowRequired(lp)
	if len(shallowKeys) == 0 {
		return
	}
	sort.Strings(shallowKeys)

	attrsTemplate, err := attrsForShallowFixture(cfg, shallowKeys)
	if err != nil {
		*issues = append(*issues, fmt.Sprintf("[Schema] build ValidatePanel attrs: %v", err))
		return
	}

	rootPath := path.Root("panels").AtMapKey("stub")

	for _, target := range shallowKeys {
		nv, err := nullAttrFromExisting(attrsTemplate[target])
		if err != nil {
			*issues = append(*issues, fmt.Sprintf("[Schema] ValidatePanelConfig null %q: %v", target, err))
			continue
		}
		attrs := shallowCloneAttrs(attrsTemplate)
		attrs[target] = nv
		diags := handler.ValidatePanelConfig(ctx, attrs, rootPath)
		if !diags.HasError() {
			*issues = append(*issues, fmt.Sprintf("[Schema] ValidatePanelConfig expected error when %#v is zeroed/null", target))
		}
	}
}

func shallowCloneAttrs(in map[string]attr.Value) map[string]attr.Value {
	out := make(map[string]attr.Value, len(in))
	maps.Copy(out, in)
	return out
}

func flattenShallowRequired(lp leafPaths) []string {
	out := []string(nil)
	for _, rp := range lp.required {
		if len(rp) == 1 {
			out = append(out, rp[0])
		}
	}
	return out
}

func attrsForShallowFixture(cfg map[string]any, shallowKeys []string) (map[string]attr.Value, error) {
	attrs := make(map[string]attr.Value)
	for _, k := range shallowKeys {
		raw, ok := rawFixtureScalarAtConfig(cfg, k)
		if !ok {
			return nil, fmt.Errorf("missing TF attr %q in config payload", k)
		}
		nv, err := attrFromFixtureScalar(raw)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", k, err)
		}
		attrs[k] = nv
	}
	return attrs, nil
}

func attrFromFixtureScalar(v any) (attr.Value, error) {
	switch v := v.(type) {
	case string:
		return types.StringValue(v), nil
	case bool:
		return types.BoolValue(v), nil
	case float64:
		return types.Float64Value(v), nil
	case json.Number:
		n, err := v.Int64()
		if err == nil {
			return types.Int64Value(n), nil
		}
		f, err := v.Float64()
		if err != nil {
			return nil, err
		}
		return types.Float64Value(f), nil
	default:
		return nil, fmt.Errorf("unsupported fixture scalar %T", v)
	}
}

func nullAttrFromExisting(v attr.Value) (attr.Value, error) {
	switch v.(type) {
	case types.String:
		return types.StringNull(), nil
	case types.Bool:
		return types.BoolNull(), nil
	case types.Float64:
		return types.Float64Null(), nil
	case types.Int64:
		return types.Int64Null(), nil
	default:
		return nil, fmt.Errorf("unsupported attr %T", v)
	}
}
