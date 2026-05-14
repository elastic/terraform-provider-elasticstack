package contracttest

import (
	"slices"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

type leafPaths struct {
	optional [][]string
	required [][]string
}

func collectLeafPaths(root schema.SingleNestedAttribute) leafPaths {
	var out leafPaths
	for name, a := range root.Attributes {
		walkAttributes(a, []string{name}, &out)
	}
	return out
}

func walkAttributes(a schema.Attribute, parent []string, out *leafPaths) {
	switch at := a.(type) {
	case schema.StringAttribute:
		recordSchemaLeaf(parent, at.Required, at.Optional, at.Computed, out)
	case schema.BoolAttribute:
		recordSchemaLeaf(parent, at.Required, at.Optional, at.Computed, out)
	case schema.Float64Attribute:
		recordSchemaLeaf(parent, at.Required, at.Optional, at.Computed, out)
	case schema.Int64Attribute:
		recordSchemaLeaf(parent, at.Required, at.Optional, at.Computed, out)
	case schema.ListAttribute, schema.ListNestedAttribute, schema.MapAttribute:
		return
	case schema.SingleNestedAttribute:
		if at.Computed {
			return
		}
		for name, nested := range at.Attributes {
			walkAttributes(nested, append(slices.Clone(parent), name), out)
		}
	default:
		// Omit unknown composite kinds for contract generation.
	}
}

func recordSchemaLeaf(path []string, required, optional, computed bool, out *leafPaths) {
	if computed || len(path) == 0 {
		return
	}
	path = slices.Clone(path)
	switch {
	case required:
		out.required = append(out.required, path)
	case optional && !required:
		out.optional = append(out.optional, path)
	}
}
