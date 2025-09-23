package role

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type RoleData struct {
	Id                      types.String `tfsdk:"id"`
	ElasticsearchConnection types.List   `tfsdk:"elasticsearch_connection"`
	Name                    types.String `tfsdk:"name"`
	Description             types.String `tfsdk:"description"`
	Applications            types.Set    `tfsdk:"applications"`
	Global                  types.String `tfsdk:"global"`
	Cluster                 types.Set    `tfsdk:"cluster"`
	Indices                 types.Set    `tfsdk:"indices"`
	RemoteIndices           types.Set    `tfsdk:"remote_indices"`
	Metadata                types.String `tfsdk:"metadata"`
	RunAs                   types.Set    `tfsdk:"run_as"`
}

type ApplicationData struct {
	Application types.String `tfsdk:"application"`
	Privileges  types.Set    `tfsdk:"privileges"`
	Resources   types.Set    `tfsdk:"resources"`
}

type IndexPermsData struct {
	FieldSecurity           types.List `tfsdk:"field_security"`
	Names                   types.Set  `tfsdk:"names"`
	Privileges              types.Set  `tfsdk:"privileges"`
	Query                   types.String `tfsdk:"query"`
	AllowRestrictedIndices  types.Bool `tfsdk:"allow_restricted_indices"`
}

type RemoteIndexPermsData struct {
	Clusters      types.Set    `tfsdk:"clusters"`
	FieldSecurity types.List   `tfsdk:"field_security"`
	Query         types.String `tfsdk:"query"`
	Names         types.Set    `tfsdk:"names"`
	Privileges    types.Set    `tfsdk:"privileges"`
}

type FieldSecurityData struct {
	Grant  types.Set `tfsdk:"grant"`
	Except types.Set `tfsdk:"except"`
}