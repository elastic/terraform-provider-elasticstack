package synthetics

import (
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
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
