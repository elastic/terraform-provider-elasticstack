package security_exception_list

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ExceptionListModel struct {
	ID            types.String `tfsdk:"id"`
	SpaceID       types.String `tfsdk:"space_id"`
	ListID        types.String `tfsdk:"list_id"`
	Name          types.String `tfsdk:"name"`
	Description   types.String `tfsdk:"description"`
	Type          types.String `tfsdk:"type"`
	NamespaceType types.String `tfsdk:"namespace_type"`
	OsTypes       types.List   `tfsdk:"os_types"`
	Tags          types.List   `tfsdk:"tags"`
	Meta          types.String `tfsdk:"meta"`
	CreatedAt     types.String `tfsdk:"created_at"`
	CreatedBy     types.String `tfsdk:"created_by"`
	UpdatedAt     types.String `tfsdk:"updated_at"`
	UpdatedBy     types.String `tfsdk:"updated_by"`
	Immutable     types.Bool   `tfsdk:"immutable"`
	TieBreakerID  types.String `tfsdk:"tie_breaker_id"`
}
