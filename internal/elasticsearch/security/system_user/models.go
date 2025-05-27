package system_user

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type SystemUserData struct {
	Id                      types.String `tfsdk:"id"`
	ElasticsearchConnection types.List   `tfsdk:"elasticsearch_connection"`
	Username                types.String `tfsdk:"username"`
	Password                types.String `tfsdk:"password"`
	PasswordHash            types.String `tfsdk:"password_hash"`
	Enabled                 types.Bool   `tfsdk:"enabled"`
}
