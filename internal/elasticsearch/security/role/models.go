package role

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type RoleData struct {
	Id                      types.String         `tfsdk:"id"`
	ElasticsearchConnection types.List           `tfsdk:"elasticsearch_connection"`
	Name                    types.String         `tfsdk:"name"`
	Description             types.String         `tfsdk:"description"`
	Applications            types.Set            `tfsdk:"applications"`
	Global                  jsontypes.Normalized `tfsdk:"global"`
	Cluster                 types.Set            `tfsdk:"cluster"`
	Indices                 types.Set            `tfsdk:"indices"`
	RemoteIndices           types.Set            `tfsdk:"remote_indices"`
	Metadata                jsontypes.Normalized `tfsdk:"metadata"`
	RunAs                   types.Set            `tfsdk:"run_as"`
}

type ApplicationData struct {
	Application types.String `tfsdk:"application"`
	Privileges  types.Set    `tfsdk:"privileges"`
	Resources   types.Set    `tfsdk:"resources"`
}

type IndexPermsData struct {
	FieldSecurity          types.List           `tfsdk:"field_security"`
	Names                  types.Set            `tfsdk:"names"`
	Privileges             types.Set            `tfsdk:"privileges"`
	Query                  jsontypes.Normalized `tfsdk:"query"`
	AllowRestrictedIndices types.Bool           `tfsdk:"allow_restricted_indices"`
}

type RemoteIndexPermsData struct {
	Clusters      types.Set            `tfsdk:"clusters"`
	FieldSecurity types.List           `tfsdk:"field_security"`
	Query         jsontypes.Normalized `tfsdk:"query"`
	Names         types.Set            `tfsdk:"names"`
	Privileges    types.Set            `tfsdk:"privileges"`
}

type FieldSecurityData struct {
	Grant  types.Set `tfsdk:"grant"`
	Except types.Set `tfsdk:"except"`
}

// toAPIModel converts the Terraform model to the API model
func (data *RoleData) toAPIModel(ctx context.Context) (*models.Role, diag.Diagnostics) {
	var diags diag.Diagnostics
	var role models.Role

	role.Name = data.Name.ValueString()

	// Description
	if !data.Description.IsNull() {
		description := data.Description.ValueString()
		role.Description = &description
	}

	// Applications
	if !data.Applications.IsNull() {
		var applicationsList []ApplicationData
		diags.Append(data.Applications.ElementsAs(ctx, &applicationsList, false)...)
		if diags.HasError() {
			return nil, diags
		}

		applications := make([]models.Application, len(applicationsList))
		for i, app := range applicationsList {
			var privileges, resources []string
			diags.Append(app.Privileges.ElementsAs(ctx, &privileges, false)...)
			diags.Append(app.Resources.ElementsAs(ctx, &resources, false)...)
			if diags.HasError() {
				return nil, diags
			}

			applications[i] = models.Application{
				Name:       app.Application.ValueString(),
				Privileges: privileges,
				Resources:  resources,
			}
		}
		role.Applications = applications
	}

	// Global
	if !data.Global.IsNull() {
		var global map[string]interface{}
		if err := json.Unmarshal([]byte(data.Global.ValueString()), &global); err != nil {
			diags.AddError("Invalid JSON", fmt.Sprintf("Error parsing global JSON: %s", err))
			return nil, diags
		}
		role.Global = global
	}

	// Cluster
	if !data.Cluster.IsNull() {
		var cluster []string
		diags.Append(data.Cluster.ElementsAs(ctx, &cluster, false)...)
		if diags.HasError() {
			return nil, diags
		}
		role.Cluster = cluster
	}

	// Indices
	if !data.Indices.IsNull() {
		var indicesList []IndexPermsData
		diags.Append(data.Indices.ElementsAs(ctx, &indicesList, false)...)
		if diags.HasError() {
			return nil, diags
		}

		indices := make([]models.IndexPerms, len(indicesList))
		for i, idx := range indicesList {
			var names, privileges []string
			diags.Append(idx.Names.ElementsAs(ctx, &names, false)...)
			diags.Append(idx.Privileges.ElementsAs(ctx, &privileges, false)...)
			if diags.HasError() {
				return nil, diags
			}

			newIndex := models.IndexPerms{
				Names:      names,
				Privileges: privileges,
			}

			if !idx.Query.IsNull() {
				query := idx.Query.ValueString()
				newIndex.Query = &query
			}

			// Field Security
			if !idx.FieldSecurity.IsNull() {
				var fieldSecList []FieldSecurityData
				diags.Append(idx.FieldSecurity.ElementsAs(ctx, &fieldSecList, false)...)
				if diags.HasError() {
					return nil, diags
				}

				if len(fieldSecList) > 0 {
					fieldSec := fieldSecList[0]
					fieldSecurity := models.FieldSecurity{}

					if !fieldSec.Grant.IsNull() {
						var grants []string
						diags.Append(fieldSec.Grant.ElementsAs(ctx, &grants, false)...)
						if diags.HasError() {
							return nil, diags
						}
						fieldSecurity.Grant = grants
					}

					if !fieldSec.Except.IsNull() {
						var excepts []string
						diags.Append(fieldSec.Except.ElementsAs(ctx, &excepts, false)...)
						if diags.HasError() {
							return nil, diags
						}
						fieldSecurity.Except = excepts
					}

					newIndex.FieldSecurity = &fieldSecurity
				}
			}

			if !idx.AllowRestrictedIndices.IsNull() {
				allowRestrictedIndices := idx.AllowRestrictedIndices.ValueBool()
				newIndex.AllowRestrictedIndices = &allowRestrictedIndices
			}

			indices[i] = newIndex
		}
		role.Indices = indices
	}

	// Remote Indices
	if !data.RemoteIndices.IsNull() {
		var remoteIndicesList []RemoteIndexPermsData
		diags.Append(data.RemoteIndices.ElementsAs(ctx, &remoteIndicesList, false)...)
		if diags.HasError() {
			return nil, diags
		}

		remoteIndices := make([]models.RemoteIndexPerms, len(remoteIndicesList))
		for i, remoteIdx := range remoteIndicesList {
			var names, clusters, privileges []string
			diags.Append(remoteIdx.Names.ElementsAs(ctx, &names, false)...)
			diags.Append(remoteIdx.Clusters.ElementsAs(ctx, &clusters, false)...)
			diags.Append(remoteIdx.Privileges.ElementsAs(ctx, &privileges, false)...)
			if diags.HasError() {
				return nil, diags
			}

			newRemoteIndex := models.RemoteIndexPerms{
				Names:      names,
				Clusters:   clusters,
				Privileges: privileges,
			}

			if !remoteIdx.Query.IsNull() {
				query := remoteIdx.Query.ValueString()
				newRemoteIndex.Query = &query
			}

			// Field Security
			if !remoteIdx.FieldSecurity.IsNull() {
				var fieldSecList []FieldSecurityData
				diags.Append(remoteIdx.FieldSecurity.ElementsAs(ctx, &fieldSecList, false)...)
				if diags.HasError() {
					return nil, diags
				}

				if len(fieldSecList) > 0 {
					fieldSec := fieldSecList[0]
					remoteFieldSecurity := models.FieldSecurity{}

					if !fieldSec.Grant.IsNull() {
						var grants []string
						diags.Append(fieldSec.Grant.ElementsAs(ctx, &grants, false)...)
						if diags.HasError() {
							return nil, diags
						}
						remoteFieldSecurity.Grant = grants
					}

					if !fieldSec.Except.IsNull() {
						var excepts []string
						diags.Append(fieldSec.Except.ElementsAs(ctx, &excepts, false)...)
						if diags.HasError() {
							return nil, diags
						}
						remoteFieldSecurity.Except = excepts
					}

					newRemoteIndex.FieldSecurity = &remoteFieldSecurity
				}
			}

			remoteIndices[i] = newRemoteIndex
		}
		role.RemoteIndices = remoteIndices
	}

	// Metadata
	if !data.Metadata.IsNull() {
		var metadata map[string]interface{}
		if err := json.Unmarshal([]byte(data.Metadata.ValueString()), &metadata); err != nil {
			diags.AddError("Invalid JSON", fmt.Sprintf("Error parsing metadata JSON: %s", err))
			return nil, diags
		}
		role.Metadata = metadata
	}

	// Run As
	if !data.RunAs.IsNull() {
		var runAs []string
		diags.Append(data.RunAs.ElementsAs(ctx, &runAs, false)...)
		if diags.HasError() {
			return nil, diags
		}
		role.RusAs = runAs
	}

	return &role, diags
}

// fromAPIModel converts the API model to the Terraform model
func (data *RoleData) fromAPIModel(ctx context.Context, role *models.Role) diag.Diagnostics {
	var diags diag.Diagnostics

	data.Name = types.StringValue(role.Name)

	// Description
	data.Description = types.StringPointerValue(role.Description)

	// Applications
	if len(role.Applications) > 0 {
		appElements := make([]attr.Value, len(role.Applications))
		for i, app := range role.Applications {
			privSet, d := types.SetValueFrom(ctx, types.StringType, app.Privileges)
			diags.Append(d...)
			if diags.HasError() {
				return diags
			}

			resSet, d := types.SetValueFrom(ctx, types.StringType, app.Resources)
			diags.Append(d...)
			if diags.HasError() {
				return diags
			}

			appObj, d := types.ObjectValue(getApplicationAttrTypes(), map[string]attr.Value{
				"application": types.StringValue(app.Name),
				"privileges":  privSet,
				"resources":   resSet,
			})
			diags.Append(d...)
			if diags.HasError() {
				return diags
			}

			appElements[i] = appObj
		}

		appSet, d := types.SetValue(types.ObjectType{AttrTypes: getApplicationAttrTypes()}, appElements)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}
		data.Applications = appSet
	} else {
		data.Applications = types.SetNull(types.ObjectType{AttrTypes: getApplicationAttrTypes()})
	}

	// Cluster
	if len(role.Cluster) > 0 {
		clusterSet, d := types.SetValueFrom(ctx, types.StringType, role.Cluster)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}
		data.Cluster = clusterSet
	} else {
		data.Cluster = types.SetNull(types.StringType)
	}

	// Global
	if role.Global != nil {
		global, err := json.Marshal(role.Global)
		if err != nil {
			diags.AddError("JSON Marshal Error", fmt.Sprintf("Error marshaling global JSON: %s", err))
			return diags
		}
		data.Global = jsontypes.NewNormalizedValue(string(global))
	} else {
		data.Global = jsontypes.NewNormalizedNull()
	}

	// Indices
	if len(role.Indices) > 0 {
		indicesElements := make([]attr.Value, len(role.Indices))
		for i, index := range role.Indices {
			namesSet, d := types.SetValueFrom(ctx, types.StringType, index.Names)
			diags.Append(d...)
			if diags.HasError() {
				return diags
			}

			privSet, d := types.SetValueFrom(ctx, types.StringType, index.Privileges)
			diags.Append(d...)
			if diags.HasError() {
				return diags
			}

			var queryVal jsontypes.Normalized
			if index.Query != nil {
				queryVal = jsontypes.NewNormalizedValue(*index.Query)
			} else {
				queryVal = jsontypes.NewNormalizedNull()
			}

			allowRestrictedVal := types.BoolPointerValue(index.AllowRestrictedIndices)

			var fieldSecList types.List
			if index.FieldSecurity != nil {
				grantSet, d := types.SetValueFrom(ctx, types.StringType, index.FieldSecurity.Grant)
				diags.Append(d...)
				if diags.HasError() {
					return diags
				}

				exceptSet, d := types.SetValueFrom(ctx, types.StringType, index.FieldSecurity.Except)
				diags.Append(d...)
				if diags.HasError() {
					return diags
				}

				fieldSecObj, d := types.ObjectValue(getFieldSecurityAttrTypes(), map[string]attr.Value{
					"grant":  grantSet,
					"except": exceptSet,
				})
				diags.Append(d...)
				if diags.HasError() {
					return diags
				}

				fieldSecList, d = types.ListValue(types.ObjectType{AttrTypes: getFieldSecurityAttrTypes()}, []attr.Value{fieldSecObj})
				diags.Append(d...)
				if diags.HasError() {
					return diags
				}
			} else {
				fieldSecList = types.ListNull(types.ObjectType{AttrTypes: getFieldSecurityAttrTypes()})
			}

			indexObj, d := types.ObjectValue(getIndexPermsAttrTypes(), map[string]attr.Value{
				"field_security":           fieldSecList,
				"names":                    namesSet,
				"privileges":               privSet,
				"query":                    queryVal,
				"allow_restricted_indices": allowRestrictedVal,
			})
			diags.Append(d...)
			if diags.HasError() {
				return diags
			}

			indicesElements[i] = indexObj
		}

		indicesSet, d := types.SetValue(types.ObjectType{AttrTypes: getIndexPermsAttrTypes()}, indicesElements)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}
		data.Indices = indicesSet
	} else {
		data.Indices = types.SetNull(types.ObjectType{AttrTypes: getIndexPermsAttrTypes()})
	}

	// Remote Indices
	if len(role.RemoteIndices) > 0 {
		remoteIndicesElements := make([]attr.Value, len(role.RemoteIndices))
		for i, remoteIndex := range role.RemoteIndices {
			clustersSet, d := types.SetValueFrom(ctx, types.StringType, remoteIndex.Clusters)
			diags.Append(d...)
			if diags.HasError() {
				return diags
			}

			namesSet, d := types.SetValueFrom(ctx, types.StringType, remoteIndex.Names)
			diags.Append(d...)
			if diags.HasError() {
				return diags
			}

			privSet, d := types.SetValueFrom(ctx, types.StringType, remoteIndex.Privileges)
			diags.Append(d...)
			if diags.HasError() {
				return diags
			}

			var queryVal jsontypes.Normalized
			if remoteIndex.Query != nil {
				queryVal = jsontypes.NewNormalizedValue(*remoteIndex.Query)
			} else {
				queryVal = jsontypes.NewNormalizedNull()
			}

			var fieldSecList types.List
			if remoteIndex.FieldSecurity != nil {
				grantSet, d := types.SetValueFrom(ctx, types.StringType, remoteIndex.FieldSecurity.Grant)
				diags.Append(d...)
				if diags.HasError() {
					return diags
				}

				exceptSet, d := types.SetValueFrom(ctx, types.StringType, remoteIndex.FieldSecurity.Except)
				diags.Append(d...)
				if diags.HasError() {
					return diags
				}

				fieldSecObj, d := types.ObjectValue(getFieldSecurityAttrTypes(), map[string]attr.Value{
					"grant":  grantSet,
					"except": exceptSet,
				})
				diags.Append(d...)
				if diags.HasError() {
					return diags
				}

				fieldSecList, d = types.ListValue(types.ObjectType{AttrTypes: getFieldSecurityAttrTypes()}, []attr.Value{fieldSecObj})
				diags.Append(d...)
				if diags.HasError() {
					return diags
				}
			} else {
				fieldSecList = types.ListNull(types.ObjectType{AttrTypes: getFieldSecurityAttrTypes()})
			}

			remoteIndexObj, d := types.ObjectValue(getRemoteIndexPermsAttrTypes(), map[string]attr.Value{
				"clusters":       clustersSet,
				"field_security": fieldSecList,
				"query":          queryVal,
				"names":          namesSet,
				"privileges":     privSet,
			})
			diags.Append(d...)
			if diags.HasError() {
				return diags
			}

			remoteIndicesElements[i] = remoteIndexObj
		}

		remoteIndicesSet, d := types.SetValue(types.ObjectType{AttrTypes: getRemoteIndexPermsAttrTypes()}, remoteIndicesElements)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}
		data.RemoteIndices = remoteIndicesSet
	} else {
		data.RemoteIndices = types.SetNull(types.ObjectType{AttrTypes: getRemoteIndexPermsAttrTypes()})
	}

	// Metadata
	if role.Metadata != nil {
		metadata, err := json.Marshal(role.Metadata)
		if err != nil {
			diags.AddError("JSON Marshal Error", fmt.Sprintf("Error marshaling metadata JSON: %s", err))
			return diags
		}
		data.Metadata = jsontypes.NewNormalizedValue(string(metadata))
	} else {
		data.Metadata = jsontypes.NewNormalizedNull()
	}

	// Run As
	if len(role.RusAs) > 0 {
		runAsSet, d := types.SetValueFrom(ctx, types.StringType, role.RusAs)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}
		data.RunAs = runAsSet
	} else {
		data.RunAs = types.SetNull(types.StringType)
	}

	return diags
}
