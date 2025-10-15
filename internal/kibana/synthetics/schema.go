package synthetics

import (
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const (
	MetadataPrefix = "_kibana_synthetics_"
)

// GetCompositeId parses a composite ID and returns the parsed components
func GetCompositeId(id string) (*clients.CompositeId, diag.Diagnostics) {
	compositeID, sdkDiag := clients.CompositeIdFromStr(id)
	dg := diag.Diagnostics{}
	if sdkDiag.HasError() {
		dg.AddError(fmt.Sprintf("Failed to parse monitor ID %s", id), fmt.Sprintf("Resource ID must have following format: <cluster_uuid>/<resource identifier>. Current value: %s", id))
		return nil, dg
	}
	return compositeID, dg
}

// Shared utility functions for type conversions
func ValueStringSlice(v []types.String) []string {
	var res []string
	for _, s := range v {
		res = append(res, s.ValueString())
	}
	return res
}

func StringSliceValue(v []string) []types.String {
	var res []types.String
	for _, s := range v {
		res = append(res, types.StringValue(s))
	}
	return res
}

func MapStringValue(v map[string]string) types.Map {
	if len(v) == 0 {
		return types.MapNull(types.StringType)
	}
	elements := make(map[string]attr.Value)
	for k, val := range v {
		elements[k] = types.StringValue(val)
	}
	mapValue, _ := types.MapValue(types.StringType, elements)
	return mapValue
}

func ValueStringMap(v types.Map) map[string]string {
	if v.IsNull() || v.IsUnknown() {
		return make(map[string]string)
	}
	result := make(map[string]string)
	for k, val := range v.Elements() {
		if strVal, ok := val.(types.String); ok {
			result[k] = strVal.ValueString()
		}
	}
	return result
}
