package security

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

var minSupportedRemoteIndicesVersion = version.Must(version.NewVersion("8.10.0"))
var minSupportedDescriptionVersion = version.Must(version.NewVersion("8.15.0"))

func ResourceRole() *schema.Resource {
	roleSchema := map[string]*schema.Schema{
		"id": {
			Description: "Internal identifier of the resource",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"name": {
			Description: "The name of the role.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"description": {
			Description: "The description of the role.",
			Type:        schema.TypeString,
			Optional:    true,
		},
		"applications": {
			Description: "A list of application privilege entries.",
			Type:        schema.TypeSet,
			Optional:    true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"application": {
						Description: "The name of the application to which this entry applies.",
						Type:        schema.TypeString,
						Required:    true,
					},
					"privileges": {
						Description: "A list of strings, where each element is the name of an application privilege or action.",
						Type:        schema.TypeSet,
						Elem: &schema.Schema{
							Type: schema.TypeString,
						},
						Required: true,
					},
					"resources": {
						Description: "A list resources to which the privileges are applied.",
						Type:        schema.TypeSet,
						Elem: &schema.Schema{
							Type: schema.TypeString,
						},
						Required: true,
					},
				},
			},
		},
		"global": {
			Description:      "An object defining global privileges.",
			Type:             schema.TypeString,
			Optional:         true,
			ValidateFunc:     validation.StringIsJSON,
			DiffSuppressFunc: utils.DiffJsonSuppress,
		},
		"cluster": {
			Description: "A list of cluster privileges. These privileges define the cluster level actions that users with this role are able to execute.",
			Type:        schema.TypeSet,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
			Optional: true,
		},
		"indices": {
			Description: "A list of indices permissions entries.",
			Type:        schema.TypeSet,
			Optional:    true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"field_security": {
						Description: "The document fields that the owners of the role have read access to.",
						Type:        schema.TypeList,
						Optional:    true,
						MaxItems:    1,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"grant": {
									Description: "List of the fields to grant the access to.",
									Type:        schema.TypeSet,
									Optional:    true,
									Elem: &schema.Schema{
										Type: schema.TypeString,
									},
								},
								"except": {
									Description: "List of the fields to which the grants will not be applied.",
									Type:        schema.TypeSet,
									Optional:    true,
									Elem: &schema.Schema{
										Type: schema.TypeString,
									},
								},
							},
						},
					},
					"names": {
						Description: "A list of indices (or index name patterns) to which the permissions in this entry apply.",
						Type:        schema.TypeSet,
						Required:    true,
						Elem: &schema.Schema{
							Type: schema.TypeString,
						},
					},
					"privileges": {
						Description: "The index level privileges that the owners of the role have on the specified indices.",
						Type:        schema.TypeSet,
						Required:    true,
						Elem: &schema.Schema{
							Type: schema.TypeString,
						},
					},
					"query": {
						Description:      "A search query that defines the documents the owners of the role have read access to.",
						Type:             schema.TypeString,
						ValidateFunc:     validation.StringIsJSON,
						DiffSuppressFunc: utils.DiffJsonSuppress,
						Optional:         true,
					},
					"allow_restricted_indices": {
						Description: "Include matching restricted indices in names parameter. Usage is strongly discouraged as it can grant unrestricted operations on critical data, make the entire system unstable or leak sensitive information.",
						Type:        schema.TypeBool,
						Optional:    true,
					},
				},
			},
		},
		"remote_indices": {
			Description: "A list of remote indices permissions entries. Remote indices are effective for remote clusters configured with the API key based model. They have no effect for remote clusters configured with the certificate based model.",
			Type:        schema.TypeSet,
			Optional:    true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"clusters": {
						Description: "A list of cluster aliases to which the permissions in this entry apply.",
						Type:        schema.TypeSet,
						Required:    true,
						Elem: &schema.Schema{
							Type: schema.TypeString,
						},
					},
					"field_security": {
						Description: "The document fields that the owners of the role have read access to.",
						Type:        schema.TypeList,
						Optional:    true,
						MaxItems:    1,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"grant": {
									Description: "List of the fields to grant the access to.",
									Type:        schema.TypeSet,
									Optional:    true,
									Elem: &schema.Schema{
										Type: schema.TypeString,
									},
								},
								"except": {
									Description: "List of the fields to which the grants will not be applied.",
									Type:        schema.TypeSet,
									Optional:    true,
									Elem: &schema.Schema{
										Type: schema.TypeString,
									},
								},
							},
						},
					},
					"query": {
						Description:      "A search query that defines the documents the owners of the role have read access to.",
						Type:             schema.TypeString,
						ValidateFunc:     validation.StringIsJSON,
						DiffSuppressFunc: utils.DiffJsonSuppress,
						Optional:         true,
					},
					"names": {
						Description: "A list of indices (or index name patterns) to which the permissions in this entry apply.",
						Type:        schema.TypeSet,
						Required:    true,
						Elem: &schema.Schema{
							Type: schema.TypeString,
						},
					},
					"privileges": {
						Description: "The index level privileges that the owners of the role have on the specified indices.",
						Type:        schema.TypeSet,
						Required:    true,
						Elem: &schema.Schema{
							Type: schema.TypeString,
						},
					},
				},
			},
		},
		"metadata": {
			Description:      "Optional meta-data.",
			Type:             schema.TypeString,
			Optional:         true,
			Computed:         true,
			ValidateFunc:     validation.StringIsJSON,
			DiffSuppressFunc: utils.DiffJsonSuppress,
		},
		"run_as": {
			Description: "A list of users that the owners of this role can impersonate.",
			Type:        schema.TypeSet,
			Optional:    true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
	}

	utils.AddConnectionSchema(roleSchema)

	return &schema.Resource{
		Description: "Adds and updates roles in the native realm. See, https://www.elastic.co/guide/en/elasticsearch/reference/current/security-api-put-role.html",

		CreateContext: resourceSecurityRolePut,
		UpdateContext: resourceSecurityRolePut,
		ReadContext:   resourceSecurityRoleRead,
		DeleteContext: resourceSecurityRoleDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: roleSchema,
	}
}

func resourceSecurityRolePut(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, diags := clients.NewApiClientFromSDKResource(d, meta)
	if diags.HasError() {
		return diags
	}
	serverVersion, diags := client.ServerVersion(ctx)
	if diags.HasError() {
		return diags
	}
	roleId := d.Get("name").(string)
	id, diags := client.ID(ctx, roleId)
	if diags.HasError() {
		return diags
	}
	var role models.Role
	role.Name = roleId

	// Add description to the role
	if v, ok := d.GetOk("description"); ok {
		// Return an error if the server version is less than the minimum supported version
		if serverVersion.LessThan(minSupportedDescriptionVersion) {
			return diag.FromErr(fmt.Errorf("'description' is supported only for Elasticsearch v%s and above", minSupportedDescriptionVersion.String()))
		}

		description := v.(string)
		role.Description = &description
	}

	if v, ok := d.GetOk("applications"); ok {
		definedApps := v.(*schema.Set)
		applications := make([]models.Application, definedApps.Len())
		for i, app := range definedApps.List() {
			a := app.(map[string]interface{})

			definedPrivs := a["privileges"].(*schema.Set)
			privs := make([]string, definedPrivs.Len())
			for i, pr := range definedPrivs.List() {
				privs[i] = pr.(string)
			}
			definedRess := a["resources"].(*schema.Set)
			ress := make([]string, definedRess.Len())
			for i, res := range definedRess.List() {
				ress[i] = res.(string)
			}

			newApp := models.Application{
				Name:       a["application"].(string),
				Privileges: privs,
				Resources:  ress,
			}
			applications[i] = newApp
		}
		role.Applications = applications
	}

	if v, ok := d.GetOk("global"); ok {
		global := make(map[string]interface{})
		if err := json.NewDecoder(strings.NewReader(v.(string))).Decode(&global); err != nil {
			return diag.FromErr(err)
		}
		role.Global = global
	}

	if v, ok := d.GetOk("cluster"); ok {
		definedCluster := v.(*schema.Set)
		cls := make([]string, definedCluster.Len())
		for i, cl := range definedCluster.List() {
			cls[i] = cl.(string)
		}
		role.Cluster = cls
	}

	if v, ok := d.GetOk("indices"); ok {
		definedIndices := v.(*schema.Set)
		indices := make([]models.IndexPerms, definedIndices.Len())
		for i, idx := range definedIndices.List() {
			index := idx.(map[string]interface{})

			definedNames := index["names"].(*schema.Set)
			names := make([]string, definedNames.Len())
			for i, name := range definedNames.List() {
				names[i] = name.(string)
			}
			definedPrivs := index["privileges"].(*schema.Set)
			privs := make([]string, definedPrivs.Len())
			for i, pr := range definedPrivs.List() {
				privs[i] = pr.(string)
			}

			newIndex := models.IndexPerms{
				Names:      names,
				Privileges: privs,
			}

			if query := index["query"].(string); query != "" {
				newIndex.Query = &query
			}
			if fieldSec := index["field_security"].([]interface{}); len(fieldSec) > 0 {
				fieldSecurity := models.FieldSecurity{}
				// there must be only 1 entry
				definedFieldSec := fieldSec[0].(map[string]interface{})

				// grants
				if gr := definedFieldSec["grant"].(*schema.Set); gr != nil {
					grants := make([]string, gr.Len())
					for i, grant := range gr.List() {
						grants[i] = grant.(string)
					}
					fieldSecurity.Grant = grants
				}
				// except
				if exp := definedFieldSec["except"].(*schema.Set); exp != nil {
					excepts := make([]string, exp.Len())
					for i, except := range exp.List() {
						excepts[i] = except.(string)
					}
					fieldSecurity.Except = excepts
				}
				newIndex.FieldSecurity = &fieldSecurity
			}

			allowRestrictedIndices := index["allow_restricted_indices"].(bool)
			newIndex.AllowRestrictedIndices = &allowRestrictedIndices

			indices[i] = newIndex
		}
		role.Indices = indices
	}

	if v, ok := d.GetOk("remote_indices"); ok {
		definedRemoteIndices := v.(*schema.Set)
		if definedRemoteIndices.Len() > 0 && serverVersion.LessThan(minSupportedRemoteIndicesVersion) {
			return diag.FromErr(fmt.Errorf("'remote_indices' is supported only for Elasticsearch v%s and above", minSupportedRemoteIndicesVersion.String()))
		}
		remoteIndices := make([]models.RemoteIndexPerms, definedRemoteIndices.Len())
		for i, idx := range definedRemoteIndices.List() {
			remoteIndex := idx.(map[string]interface{})

			definedRemoteNames := remoteIndex["names"].(*schema.Set)
			remoteNames := make([]string, definedRemoteNames.Len())
			for i, name := range definedRemoteNames.List() {
				remoteNames[i] = name.(string)
			}
			definedRemoteClusters := remoteIndex["clusters"].(*schema.Set)
			remoteClusters := make([]string, definedRemoteClusters.Len())
			for i, cluster := range definedRemoteClusters.List() {
				remoteClusters[i] = cluster.(string)
			}
			definedRemotePrivs := remoteIndex["privileges"].(*schema.Set)
			remotePrivs := make([]string, definedRemotePrivs.Len())
			for i, pr := range definedRemotePrivs.List() {
				remotePrivs[i] = pr.(string)
			}

			newRemoteIndex := models.RemoteIndexPerms{
				Names:      remoteNames,
				Clusters:   remoteClusters,
				Privileges: remotePrivs,
			}

			if query := remoteIndex["query"].(string); query != "" {
				newRemoteIndex.Query = &query
			}
			if fieldSec := remoteIndex["field_security"].([]interface{}); len(fieldSec) > 0 {
				remoteFieldSecurity := models.FieldSecurity{}
				// there must be only 1 entry
				definedRemoteFieldSec := fieldSec[0].(map[string]interface{})

				// grants
				if gr := definedRemoteFieldSec["grant"].(*schema.Set); gr != nil {
					grants := make([]string, gr.Len())
					for i, grant := range gr.List() {
						grants[i] = grant.(string)
					}
					remoteFieldSecurity.Grant = grants
				}
				// except
				if exp := definedRemoteFieldSec["except"].(*schema.Set); exp != nil {
					excepts := make([]string, exp.Len())
					for i, except := range exp.List() {
						excepts[i] = except.(string)
					}
					remoteFieldSecurity.Except = excepts
				}
				newRemoteIndex.FieldSecurity = &remoteFieldSecurity
			}

			remoteIndices[i] = newRemoteIndex
		}
		role.RemoteIndices = remoteIndices
	}

	if v, ok := d.GetOk("metadata"); ok {
		metadata := make(map[string]interface{})
		if err := json.NewDecoder(strings.NewReader(v.(string))).Decode(&metadata); err != nil {
			return diag.FromErr(err)
		}
		role.Metadata = metadata
	}

	if v, ok := d.GetOk("run_as"); ok {
		definedRuns := v.(*schema.Set)
		runs := make([]string, definedRuns.Len())
		for i, run := range definedRuns.List() {
			runs[i] = run.(string)
		}
		role.RusAs = runs
	}

	if diags := elasticsearch.PutRole(ctx, client, &role); diags.HasError() {
		return diags
	}

	d.SetId(id.String())
	return resourceSecurityRoleRead(ctx, d, meta)
}

func resourceSecurityRoleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, diags := clients.NewApiClientFromSDKResource(d, meta)
	if diags.HasError() {
		return diags
	}
	compId, diags := clients.CompositeIdFromStr(d.Id())
	if diags.HasError() {
		return diags
	}
	roleId := compId.ResourceId

	role, diags := elasticsearch.GetRole(ctx, client, roleId)
	if role == nil && diags == nil {
		tflog.Warn(ctx, fmt.Sprintf(`Role "%s" not found, removing from state`, roleId))
		d.SetId("")
		return diags
	}
	if diags.HasError() {
		return diags
	}

	// set the fields
	if err := d.Set("name", roleId); err != nil {
		return diag.FromErr(err)
	}

	// Set the description if it exists
	if role.Description != nil {
		if err := d.Set("description", *role.Description); err != nil {
			return diag.FromErr(err)
		}
	}

	apps := role.Applications
	applications := flattenApplicationsData(&apps)
	if err := d.Set("applications", applications); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("cluster", role.Cluster); err != nil {
		return diag.FromErr(err)
	}

	if role.Global != nil {
		global, err := json.Marshal(role.Global)
		if err != nil {
			return diag.FromErr(err)
		}
		if err := d.Set("global", string(global)); err != nil {
			return diag.FromErr(err)
		}
	}

	indices := flattenIndicesData(role.Indices)
	if err := d.Set("indices", indices); err != nil {
		return diag.FromErr(err)
	}
	remoteIndices := flattenRemoteIndicesData(role.RemoteIndices)
	if err := d.Set("remote_indices", remoteIndices); err != nil {
		return diag.FromErr(err)
	}

	if role.Metadata != nil {
		metadata, err := json.Marshal(role.Metadata)
		if err != nil {
			return diag.FromErr(err)
		}
		if err := d.Set("metadata", string(metadata)); err != nil {
			return diag.FromErr(err)
		}
	}

	if err := d.Set("run_as", role.RusAs); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func flattenApplicationsData(apps *[]models.Application) []interface{} {
	if apps != nil {
		oapps := make([]interface{}, len(*apps))
		for i, app := range *apps {
			oa := make(map[string]interface{})
			oa["application"] = app.Name
			oa["privileges"] = app.Privileges
			oa["resources"] = app.Resources
			oapps[i] = oa
		}
		return oapps
	}
	return make([]interface{}, 0)
}

func flattenIndicesData(indices []models.IndexPerms) []interface{} {
	oindx := make([]interface{}, len(indices))

	for i, index := range indices {
		oi := make(map[string]interface{})
		oi["names"] = index.Names
		oi["privileges"] = index.Privileges
		oi["query"] = index.Query
		oi["allow_restricted_indices"] = index.AllowRestrictedIndices

		if index.FieldSecurity != nil {
			fsec := make(map[string]interface{})
			fsec["grant"] = index.FieldSecurity.Grant
			fsec["except"] = index.FieldSecurity.Except
			oi["field_security"] = []interface{}{fsec}
		}
		oindx[i] = oi
	}
	return oindx
}

func flattenRemoteIndicesData(remoteIndices []models.RemoteIndexPerms) []interface{} {
	oRemoteIndx := make([]interface{}, len(remoteIndices))

	for i, remoteIndex := range remoteIndices {
		oi := make(map[string]interface{})
		oi["names"] = remoteIndex.Names
		oi["clusters"] = remoteIndex.Clusters
		oi["privileges"] = remoteIndex.Privileges
		oi["query"] = remoteIndex.Query

		if remoteIndex.FieldSecurity != nil {
			fsec := make(map[string]interface{})
			fsec["grant"] = remoteIndex.FieldSecurity.Grant
			fsec["except"] = remoteIndex.FieldSecurity.Except
			oi["field_security"] = []interface{}{fsec}
		}
		oRemoteIndx[i] = oi
	}
	return oRemoteIndx
}

func resourceSecurityRoleDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, diags := clients.NewApiClientFromSDKResource(d, meta)
	if diags.HasError() {
		return diags
	}
	compId, diags := clients.CompositeIdFromStr(d.Id())
	if diags.HasError() {
		return diags
	}

	if diags := elasticsearch.DeleteRole(ctx, client, compId.ResourceId); diags.HasError() {
		return diags
	}

	return diags
}
