package synthetics

import (
	"fmt"

	"github.com/disaster37/go-kibana-rest/v8/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
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

// Geographic configuration schema and types
func GeoConfigSchema() schema.Attribute {
	return schema.SingleNestedAttribute{
		Optional:    true,
		Description: "Geographic coordinates (WGS84) for the location",
		Attributes: map[string]schema.Attribute{
			"lat": schema.Float64Attribute{
				Optional:            false,
				Required:            true,
				MarkdownDescription: "The latitude of the location.",
			},
			"lon": schema.Float64Attribute{
				Optional:            false,
				Required:            true,
				MarkdownDescription: "The longitude of the location.",
			},
		},
	}
}

type TFGeoConfigV0 struct {
	Lat types.Float64 `tfsdk:"lat"`
	Lon types.Float64 `tfsdk:"lon"`
}

func (m *TFGeoConfigV0) ToSyntheticGeoConfig() *kbapi.SyntheticGeoConfig {
	return &kbapi.SyntheticGeoConfig{
		Lat: m.Lat.ValueFloat64(),
		Lon: m.Lon.ValueFloat64(),
	}
}

func FromSyntheticGeoConfig(v *kbapi.SyntheticGeoConfig) *TFGeoConfigV0 {
	if v == nil {
		return nil
	}
	return &TFGeoConfigV0{
		Lat: types.Float64Value(v.Lat),
		Lon: types.Float64Value(v.Lon),
	}
}
