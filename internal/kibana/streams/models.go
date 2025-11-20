package streams

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// ingestAttrTypes describes the shape of the `ingest` nested object in state.
var ingestAttrTypes = map[string]attr.Type{
	"type": types.StringType,
}

// groupModel represents the group configuration for a group stream.
type groupModel struct {
	Members  []types.String `tfsdk:"members"`
	Metadata types.Map      `tfsdk:"metadata"`
	Tags     []types.String `tfsdk:"tags"`
}

// streamModel is the top-level Terraform representation of a Kibana stream.
type streamModel struct {
	ID              types.String `tfsdk:"id"`
	Name            types.String `tfsdk:"name"`
	SpaceID         types.String `tfsdk:"space_id"`
	Description     types.String `tfsdk:"description"`
	Type            types.String `tfsdk:"type"`
	CreateIfMissing types.Bool   `tfsdk:"create_if_missing"`

	Ingest types.Object `tfsdk:"ingest"`
	Group  *groupModel  `tfsdk:"group"`
}
