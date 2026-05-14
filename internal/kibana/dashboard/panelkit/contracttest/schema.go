package contracttest

import (
	"context"
	"encoding/json"
	"fmt"
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
	if !(sn.Optional && !sn.Required && !sn.Computed) {
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
		camel := terraformPathToAPICamel(rp)
		if _, ok := jsonNavigateMap(cfg, camel); !ok {
			*issues = append(*issues, fmt.Sprintf("[Schema] required terraform path %#v absent from fixture.config under API segments %#v", rp, camel))
		}
	}
}

func appendValidateRequiredZeroIssues(handler iface.Handler, fixtureJSON string, issues *[]string) {
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

	attrs, err := attrsForShallowFixture(cfg, shallowKeys)
	if err != nil {
		*issues = append(*issues, fmt.Sprintf("[Schema] build ValidatePanel attrs: %v", err))
		return
	}

	target := shallowKeys[0]
	nv, err := nullAttrFromExisting(attrs[target])
	if err != nil {
		*issues = append(*issues, fmt.Sprintf("[Schema] build null attr: %v", err))
		return
	}
	attrs[target] = nv

	ctx := context.Background()
	diags := handler.ValidatePanelConfig(ctx, handler.PanelType(), attrs, path.Root("panels").AtMapKey("stub"))
	if !diags.HasError() {
		*issues = append(*issues, fmt.Sprintf("[Schema] ValidatePanelConfig expected error when %#v is null", target))
	}
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
		apiKey := tfAttrToAPICamel(k)
		raw, ok := cfg[apiKey]
		if !ok {
			return nil, fmt.Errorf("missing TF attr %q fixture key %q in config payload", k, apiKey)
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
