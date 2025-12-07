package user

import (
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type UserData struct {
	Id                      types.String         `tfsdk:"id"`
	ElasticsearchConnection types.List           `tfsdk:"elasticsearch_connection"`
	Username                types.String         `tfsdk:"username"`
	Password                types.String         `tfsdk:"password"`
	PasswordHash            types.String         `tfsdk:"password_hash"`
	PasswordWo              types.String         `tfsdk:"password_wo"`
	PasswordWoVersion       types.String         `tfsdk:"password_wo_version"`
	FullName                types.String         `tfsdk:"full_name"`
	Email                   types.String         `tfsdk:"email"`
	Roles                   types.Set            `tfsdk:"roles"`
	Metadata                jsontypes.Normalized `tfsdk:"metadata"`
	Enabled                 types.Bool           `tfsdk:"enabled"`
}
