package role

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	MinSupportedRemoteIndicesVersion = version.Must(version.NewVersion("8.10.0"))
	MinSupportedDescriptionVersion   = version.Must(version.NewVersion("8.15.0"))
)

func (r *roleResource) update(ctx context.Context, plan tfsdk.Plan, state *tfsdk.State) diag.Diagnostics {
	var data RoleData
	var diags diag.Diagnostics
	diags.Append(plan.Get(ctx, &data)...)
	if diags.HasError() {
		return diags
	}

	roleId := data.Name.ValueString()
	id, sdkDiags := r.client.ID(ctx, roleId)
	diags.Append(diagutil.FrameworkDiagsFromSDK(sdkDiags)...)
	if diags.HasError() {
		return diags
	}

	client, diags := clients.MaybeNewApiClientFromFrameworkResource(ctx, data.ElasticsearchConnection, r.client)
	diags.Append(diags...)
	if diags.HasError() {
		return diags
	}

	serverVersion, sdkDiags := client.ServerVersion(ctx)
	diags.Append(diagutil.FrameworkDiagsFromSDK(sdkDiags)...)
	if diags.HasError() {
		return diags
	}

	var role models.Role
	role.Name = roleId

	// Add description to the role
	if utils.IsKnown(data.Description) && !data.Description.IsNull() {
		// Return an error if the server version is less than the minimum supported version
		if serverVersion.LessThan(MinSupportedDescriptionVersion) {
			diags.AddError("Unsupported Feature", fmt.Sprintf("'description' is supported only for Elasticsearch v%s and above", MinSupportedDescriptionVersion.String()))
			return diags
		}

		description := data.Description.ValueString()
		role.Description = &description
	}

	// Applications
	if utils.IsKnown(data.Applications) && !data.Applications.IsNull() {
		var applicationsList []ApplicationData
		diags.Append(data.Applications.ElementsAs(ctx, &applicationsList, false)...)
		if diags.HasError() {
			return diags
		}

		applications := make([]models.Application, len(applicationsList))
		for i, app := range applicationsList {
			var privileges, resources []string
			diags.Append(app.Privileges.ElementsAs(ctx, &privileges, false)...)
			diags.Append(app.Resources.ElementsAs(ctx, &resources, false)...)
			if diags.HasError() {
				return diags
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
	if utils.IsKnown(data.Global) && !data.Global.IsNull() {
		global := make(map[string]interface{})
		if err := json.NewDecoder(strings.NewReader(data.Global.ValueString())).Decode(&global); err != nil {
			diags.AddError("Invalid JSON", fmt.Sprintf("Error parsing global JSON: %s", err))
			return diags
		}
		role.Global = global
	}

	// Cluster
	if utils.IsKnown(data.Cluster) && !data.Cluster.IsNull() {
		var cluster []string
		diags.Append(data.Cluster.ElementsAs(ctx, &cluster, false)...)
		if diags.HasError() {
			return diags
		}
		role.Cluster = cluster
	}

	// Indices
	if utils.IsKnown(data.Indices) && !data.Indices.IsNull() {
		var indicesList []IndexPermsData
		diags.Append(data.Indices.ElementsAs(ctx, &indicesList, false)...)
		if diags.HasError() {
			return diags
		}

		indices := make([]models.IndexPerms, len(indicesList))
		for i, idx := range indicesList {
			var names, privileges []string
			diags.Append(idx.Names.ElementsAs(ctx, &names, false)...)
			diags.Append(idx.Privileges.ElementsAs(ctx, &privileges, false)...)
			if diags.HasError() {
				return diags
			}

			newIndex := models.IndexPerms{
				Names:      names,
				Privileges: privileges,
			}

			if utils.IsKnown(idx.Query) && !idx.Query.IsNull() {
				query := idx.Query.ValueString()
				newIndex.Query = &query
			}

			// Field Security
			if utils.IsKnown(idx.FieldSecurity) && !idx.FieldSecurity.IsNull() {
				var fieldSecList []FieldSecurityData
				diags.Append(idx.FieldSecurity.ElementsAs(ctx, &fieldSecList, false)...)
				if diags.HasError() {
					return diags
				}

				if len(fieldSecList) > 0 {
					fieldSec := fieldSecList[0]
					fieldSecurity := models.FieldSecurity{}

					if utils.IsKnown(fieldSec.Grant) && !fieldSec.Grant.IsNull() {
						var grants []string
						diags.Append(fieldSec.Grant.ElementsAs(ctx, &grants, false)...)
						if diags.HasError() {
							return diags
						}
						fieldSecurity.Grant = grants
					}

					if utils.IsKnown(fieldSec.Except) && !fieldSec.Except.IsNull() {
						var excepts []string
						diags.Append(fieldSec.Except.ElementsAs(ctx, &excepts, false)...)
						if diags.HasError() {
							return diags
						}
						fieldSecurity.Except = excepts
					}

					newIndex.FieldSecurity = &fieldSecurity
				}
			}

			if utils.IsKnown(idx.AllowRestrictedIndices) && !idx.AllowRestrictedIndices.IsNull() {
				allowRestrictedIndices := idx.AllowRestrictedIndices.ValueBool()
				newIndex.AllowRestrictedIndices = &allowRestrictedIndices
			}

			indices[i] = newIndex
		}
		role.Indices = indices
	}

	// Remote Indices
	if utils.IsKnown(data.RemoteIndices) && !data.RemoteIndices.IsNull() {
		var remoteIndicesList []RemoteIndexPermsData
		diags.Append(data.RemoteIndices.ElementsAs(ctx, &remoteIndicesList, false)...)
		if diags.HasError() {
			return diags
		}

		if len(remoteIndicesList) > 0 && serverVersion.LessThan(MinSupportedRemoteIndicesVersion) {
			diags.AddError("Unsupported Feature", fmt.Sprintf("'remote_indices' is supported only for Elasticsearch v%s and above", MinSupportedRemoteIndicesVersion.String()))
			return diags
		}

		remoteIndices := make([]models.RemoteIndexPerms, len(remoteIndicesList))
		for i, remoteIdx := range remoteIndicesList {
			var names, clusters, privileges []string
			diags.Append(remoteIdx.Names.ElementsAs(ctx, &names, false)...)
			diags.Append(remoteIdx.Clusters.ElementsAs(ctx, &clusters, false)...)
			diags.Append(remoteIdx.Privileges.ElementsAs(ctx, &privileges, false)...)
			if diags.HasError() {
				return diags
			}

			newRemoteIndex := models.RemoteIndexPerms{
				Names:      names,
				Clusters:   clusters,
				Privileges: privileges,
			}

			if utils.IsKnown(remoteIdx.Query) && !remoteIdx.Query.IsNull() {
				query := remoteIdx.Query.ValueString()
				newRemoteIndex.Query = &query
			}

			// Field Security
			if utils.IsKnown(remoteIdx.FieldSecurity) && !remoteIdx.FieldSecurity.IsNull() {
				var fieldSecList []FieldSecurityData
				diags.Append(remoteIdx.FieldSecurity.ElementsAs(ctx, &fieldSecList, false)...)
				if diags.HasError() {
					return diags
				}

				if len(fieldSecList) > 0 {
					fieldSec := fieldSecList[0]
					remoteFieldSecurity := models.FieldSecurity{}

					if utils.IsKnown(fieldSec.Grant) && !fieldSec.Grant.IsNull() {
						var grants []string
						diags.Append(fieldSec.Grant.ElementsAs(ctx, &grants, false)...)
						if diags.HasError() {
							return diags
						}
						remoteFieldSecurity.Grant = grants
					}

					if utils.IsKnown(fieldSec.Except) && !fieldSec.Except.IsNull() {
						var excepts []string
						diags.Append(fieldSec.Except.ElementsAs(ctx, &excepts, false)...)
						if diags.HasError() {
							return diags
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
	if utils.IsKnown(data.Metadata) && !data.Metadata.IsNull() {
		metadata := make(map[string]interface{})
		if err := json.NewDecoder(strings.NewReader(data.Metadata.ValueString())).Decode(&metadata); err != nil {
			diags.AddError("Invalid JSON", fmt.Sprintf("Error parsing metadata JSON: %s", err))
			return diags
		}
		role.Metadata = metadata
	}

	// Run As
	if utils.IsKnown(data.RunAs) && !data.RunAs.IsNull() {
		var runAs []string
		diags.Append(data.RunAs.ElementsAs(ctx, &runAs, false)...)
		if diags.HasError() {
			return diags
		}
		role.RusAs = runAs
	}

	// Put the role
	sdkDiags = elasticsearch.PutRole(ctx, client, &role)
	diags.Append(diagutil.FrameworkDiagsFromSDK(sdkDiags)...)
	if diags.HasError() {
		return diags
	}

	data.Id = types.StringValue(id.String())
	diags.Append(state.Set(ctx, &data)...)
	return diags
}

func (r *roleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	diags := r.update(ctx, req.Plan, &resp.State)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
