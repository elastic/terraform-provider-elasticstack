package role

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func (r *roleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data RoleData
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	compId, diags := clients.CompositeIdFromStrFw(data.Id.ValueString())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	roleId := compId.ResourceId

	client, diags := clients.MaybeNewApiClientFromFrameworkResource(ctx, data.ElasticsearchConnection, r.client)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	role, sdkDiags := elasticsearch.GetRole(ctx, client, roleId)
	resp.Diagnostics.Append(diagutil.FrameworkDiagsFromSDK(sdkDiags)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if role == nil {
		tflog.Warn(ctx, fmt.Sprintf(`Role "%s" not found, removing from state`, roleId))
		resp.State.RemoveResource(ctx)
		return
	}

	// Set the fields
	data.Name = types.StringValue(roleId)

	// Set the description if it exists
	if role.Description != nil {
		data.Description = types.StringValue(*role.Description)
	} else {
		data.Description = types.StringNull()
	}

	// Applications
	if len(role.Applications) > 0 {
		appElements := make([]attr.Value, len(role.Applications))
		for i, app := range role.Applications {
			privSet, diags := types.SetValueFrom(ctx, types.StringType, app.Privileges)
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}

			resSet, diags := types.SetValueFrom(ctx, types.StringType, app.Resources)
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}

			appObj, diags := types.ObjectValue(map[string]attr.Type{
				"application": types.StringType,
				"privileges":  types.SetType{ElemType: types.StringType},
				"resources":   types.SetType{ElemType: types.StringType},
			}, map[string]attr.Value{
				"application": types.StringValue(app.Name),
				"privileges":  privSet,
				"resources":   resSet,
			})
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}

			appElements[i] = appObj
		}

		appSet, diags := types.SetValue(types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"application": types.StringType,
				"privileges":  types.SetType{ElemType: types.StringType},
				"resources":   types.SetType{ElemType: types.StringType},
			},
		}, appElements)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		data.Applications = appSet
	} else {
		data.Applications = types.SetNull(types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"application": types.StringType,
				"privileges":  types.SetType{ElemType: types.StringType},
				"resources":   types.SetType{ElemType: types.StringType},
			},
		})
	}

	// Cluster
	if len(role.Cluster) > 0 {
		clusterSet, diags := types.SetValueFrom(ctx, types.StringType, role.Cluster)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		data.Cluster = clusterSet
	} else {
		data.Cluster = types.SetNull(types.StringType)
	}

	// Global
	if role.Global != nil {
		global, err := json.Marshal(role.Global)
		if err != nil {
			resp.Diagnostics.AddError("JSON Marshal Error", fmt.Sprintf("Error marshaling global JSON: %s", err))
			return
		}
		data.Global = types.StringValue(string(global))
	} else {
		data.Global = types.StringNull()
	}

	// Indices
	if len(role.Indices) > 0 {
		indicesElements := make([]attr.Value, len(role.Indices))
		for i, index := range role.Indices {
			namesSet, diags := types.SetValueFrom(ctx, types.StringType, index.Names)
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}

			privSet, diags := types.SetValueFrom(ctx, types.StringType, index.Privileges)
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}

			var queryVal types.String
			if index.Query != nil {
				queryVal = types.StringValue(*index.Query)
			} else {
				queryVal = types.StringNull()
			}

			var allowRestrictedVal types.Bool
			if index.AllowRestrictedIndices != nil {
				allowRestrictedVal = types.BoolValue(*index.AllowRestrictedIndices)
			} else {
				allowRestrictedVal = types.BoolNull()
			}

			var fieldSecList types.List
			if index.FieldSecurity != nil {
				grantSet, diags := types.SetValueFrom(ctx, types.StringType, index.FieldSecurity.Grant)
				resp.Diagnostics.Append(diags...)
				if resp.Diagnostics.HasError() {
					return
				}

				exceptSet, diags := types.SetValueFrom(ctx, types.StringType, index.FieldSecurity.Except)
				resp.Diagnostics.Append(diags...)
				if resp.Diagnostics.HasError() {
					return
				}

				fieldSecObj, diags := types.ObjectValue(map[string]attr.Type{
					"grant":  types.SetType{ElemType: types.StringType},
					"except": types.SetType{ElemType: types.StringType},
				}, map[string]attr.Value{
					"grant":  grantSet,
					"except": exceptSet,
				})
				resp.Diagnostics.Append(diags...)
				if resp.Diagnostics.HasError() {
					return
				}

				fieldSecList, diags = types.ListValue(types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"grant":  types.SetType{ElemType: types.StringType},
						"except": types.SetType{ElemType: types.StringType},
					},
				}, []attr.Value{fieldSecObj})
				resp.Diagnostics.Append(diags...)
				if resp.Diagnostics.HasError() {
					return
				}
			} else {
				fieldSecList = types.ListNull(types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"grant":  types.SetType{ElemType: types.StringType},
						"except": types.SetType{ElemType: types.StringType},
					},
				})
			}

			indexObj, diags := types.ObjectValue(map[string]attr.Type{
				"field_security":           types.ListType{ElemType: types.ObjectType{AttrTypes: map[string]attr.Type{"grant": types.SetType{ElemType: types.StringType}, "except": types.SetType{ElemType: types.StringType}}}},
				"names":                    types.SetType{ElemType: types.StringType},
				"privileges":               types.SetType{ElemType: types.StringType},
				"query":                    types.StringType,
				"allow_restricted_indices": types.BoolType,
			}, map[string]attr.Value{
				"field_security":           fieldSecList,
				"names":                    namesSet,
				"privileges":               privSet,
				"query":                    queryVal,
				"allow_restricted_indices": allowRestrictedVal,
			})
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}

			indicesElements[i] = indexObj
		}

		indicesSet, diags := types.SetValue(types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"field_security":           types.ListType{ElemType: types.ObjectType{AttrTypes: map[string]attr.Type{"grant": types.SetType{ElemType: types.StringType}, "except": types.SetType{ElemType: types.StringType}}}},
				"names":                    types.SetType{ElemType: types.StringType},
				"privileges":               types.SetType{ElemType: types.StringType},
				"query":                    types.StringType,
				"allow_restricted_indices": types.BoolType,
			},
		}, indicesElements)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		data.Indices = indicesSet
	} else {
		data.Indices = types.SetNull(types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"field_security":           types.ListType{ElemType: types.ObjectType{AttrTypes: map[string]attr.Type{"grant": types.SetType{ElemType: types.StringType}, "except": types.SetType{ElemType: types.StringType}}}},
				"names":                    types.SetType{ElemType: types.StringType},
				"privileges":               types.SetType{ElemType: types.StringType},
				"query":                    types.StringType,
				"allow_restricted_indices": types.BoolType,
			},
		})
	}

	// Remote Indices
	if len(role.RemoteIndices) > 0 {
		remoteIndicesElements := make([]attr.Value, len(role.RemoteIndices))
		for i, remoteIndex := range role.RemoteIndices {
			clustersSet, diags := types.SetValueFrom(ctx, types.StringType, remoteIndex.Clusters)
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}

			namesSet, diags := types.SetValueFrom(ctx, types.StringType, remoteIndex.Names)
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}

			privSet, diags := types.SetValueFrom(ctx, types.StringType, remoteIndex.Privileges)
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}

			var queryVal types.String
			if remoteIndex.Query != nil {
				queryVal = types.StringValue(*remoteIndex.Query)
			} else {
				queryVal = types.StringNull()
			}

			var fieldSecList types.List
			if remoteIndex.FieldSecurity != nil {
				grantSet, diags := types.SetValueFrom(ctx, types.StringType, remoteIndex.FieldSecurity.Grant)
				resp.Diagnostics.Append(diags...)
				if resp.Diagnostics.HasError() {
					return
				}

				exceptSet, diags := types.SetValueFrom(ctx, types.StringType, remoteIndex.FieldSecurity.Except)
				resp.Diagnostics.Append(diags...)
				if resp.Diagnostics.HasError() {
					return
				}

				fieldSecObj, diags := types.ObjectValue(map[string]attr.Type{
					"grant":  types.SetType{ElemType: types.StringType},
					"except": types.SetType{ElemType: types.StringType},
				}, map[string]attr.Value{
					"grant":  grantSet,
					"except": exceptSet,
				})
				resp.Diagnostics.Append(diags...)
				if resp.Diagnostics.HasError() {
					return
				}

				fieldSecList, diags = types.ListValue(types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"grant":  types.SetType{ElemType: types.StringType},
						"except": types.SetType{ElemType: types.StringType},
					},
				}, []attr.Value{fieldSecObj})
				resp.Diagnostics.Append(diags...)
				if resp.Diagnostics.HasError() {
					return
				}
			} else {
				fieldSecList = types.ListNull(types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"grant":  types.SetType{ElemType: types.StringType},
						"except": types.SetType{ElemType: types.StringType},
					},
				})
			}

			remoteIndexObj, diags := types.ObjectValue(map[string]attr.Type{
				"clusters":       types.SetType{ElemType: types.StringType},
				"field_security": types.ListType{ElemType: types.ObjectType{AttrTypes: map[string]attr.Type{"grant": types.SetType{ElemType: types.StringType}, "except": types.SetType{ElemType: types.StringType}}}},
				"query":          types.StringType,
				"names":          types.SetType{ElemType: types.StringType},
				"privileges":     types.SetType{ElemType: types.StringType},
			}, map[string]attr.Value{
				"clusters":       clustersSet,
				"field_security": fieldSecList,
				"query":          queryVal,
				"names":          namesSet,
				"privileges":     privSet,
			})
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}

			remoteIndicesElements[i] = remoteIndexObj
		}

		remoteIndicesSet, diags := types.SetValue(types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"clusters":       types.SetType{ElemType: types.StringType},
				"field_security": types.ListType{ElemType: types.ObjectType{AttrTypes: map[string]attr.Type{"grant": types.SetType{ElemType: types.StringType}, "except": types.SetType{ElemType: types.StringType}}}},
				"query":          types.StringType,
				"names":          types.SetType{ElemType: types.StringType},
				"privileges":     types.SetType{ElemType: types.StringType},
			},
		}, remoteIndicesElements)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		data.RemoteIndices = remoteIndicesSet
	} else {
		data.RemoteIndices = types.SetNull(types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"clusters":       types.SetType{ElemType: types.StringType},
				"field_security": types.ListType{ElemType: types.ObjectType{AttrTypes: map[string]attr.Type{"grant": types.SetType{ElemType: types.StringType}, "except": types.SetType{ElemType: types.StringType}}}},
				"query":          types.StringType,
				"names":          types.SetType{ElemType: types.StringType},
				"privileges":     types.SetType{ElemType: types.StringType},
			},
		})
	}

	// Metadata
	if role.Metadata != nil {
		metadata, err := json.Marshal(role.Metadata)
		if err != nil {
			resp.Diagnostics.AddError("JSON Marshal Error", fmt.Sprintf("Error marshaling metadata JSON: %s", err))
			return
		}
		data.Metadata = types.StringValue(string(metadata))
	} else {
		data.Metadata = types.StringNull()
	}

	// Run As
	if len(role.RusAs) > 0 {
		runAsSet, diags := types.SetValueFrom(ctx, types.StringType, role.RusAs)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		data.RunAs = runAsSet
	} else {
		data.RunAs = types.SetNull(types.StringType)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
