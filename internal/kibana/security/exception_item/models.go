package exception_item

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ExceptionItemModel struct {
	ID            types.String `tfsdk:"id"`
	ItemID        types.String `tfsdk:"item_id"`
	ListID        types.String `tfsdk:"list_id"`
	Name          types.String `tfsdk:"name"`
	Description   types.String `tfsdk:"description"`
	Type          types.String `tfsdk:"type"`
	NamespaceType types.String `tfsdk:"namespace_type"`
	OsTypes       types.List   `tfsdk:"os_types"`
	Tags          types.List   `tfsdk:"tags"`
	Meta          types.String `tfsdk:"meta"`
	Entries       types.String `tfsdk:"entries"`
	Comments      types.List   `tfsdk:"comments"`
	ExpireTime    types.String `tfsdk:"expire_time"`
	CreatedAt     types.String `tfsdk:"created_at"`
	CreatedBy     types.String `tfsdk:"created_by"`
	UpdatedAt     types.String `tfsdk:"updated_at"`
	UpdatedBy     types.String `tfsdk:"updated_by"`
	TieBreakerID  types.String `tfsdk:"tie_breaker_id"`
}

type CommentModel struct {
	ID      types.String `tfsdk:"id"`
	Comment types.String `tfsdk:"comment"`
}
