package panelkit

import (
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// PreserveString keeps an existing null/unknown string when it is not known; otherwise updates from the API pointer.
func PreserveString(existing types.String, api *string) types.String {
	if !typeutils.IsKnown(existing) {
		return existing
	}
	return types.StringPointerValue(api)
}

// PreserveBool keeps an existing null/unknown bool when it is not known; otherwise updates from the API pointer.
func PreserveBool(existing types.Bool, api *bool) types.Bool {
	if !typeutils.IsKnown(existing) {
		return existing
	}
	return types.BoolPointerValue(api)
}

// PreserveFloat64 keeps an existing null/unknown float when it is not known; otherwise updates from the API pointer.
func PreserveFloat64(existing types.Float64, api *float64) types.Float64 {
	if !typeutils.IsKnown(existing) {
		return existing
	}
	return types.Float64PointerValue(api)
}

// PreserveList keeps an existing null/unknown list when it is not known; otherwise replaces with next.
func PreserveList(existing, next attr.Value) attr.Value {
	if !typeutils.IsKnown(existing) {
		return existing
	}
	return next
}
