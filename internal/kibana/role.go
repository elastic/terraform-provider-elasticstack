// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package kibana

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	_ "embed"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	kibanaoapi "github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	providerSchema "github.com/elastic/terraform-provider-elasticstack/internal/schema"
	"github.com/elastic/terraform-provider-elasticstack/internal/tfsdkutils"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

var (
	minSupportedRemoteIndicesVersion = version.Must(version.NewVersion("8.10.0"))
	minSupportedDescriptionVersion   = version.Must(version.NewVersion("8.15.0"))
)

//go:embed role.md
var roleDescription string

func ResourceRole() *schema.Resource {
	roleSchema := map[string]*schema.Schema{
		"name": {
			Description: "The name for the role.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"elasticsearch": {
			Description: "Elasticsearch cluster and index privileges.",
			Type:        schema.TypeSet,
			Required:    true,
			MaxItems:    1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"cluster": {
						Description: "List of the cluster privileges.",
						Type:        schema.TypeSet,
						Optional:    true,
						Elem: &schema.Schema{
							Type: schema.TypeString,
						},
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
								"query": {
									Description:      "A search query that defines the documents the owners of the role have read access to.",
									Type:             schema.TypeString,
									ValidateFunc:     validation.StringIsJSON,
									DiffSuppressFunc: tfsdkutils.DiffJSONSuppress,
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
					"remote_indices": {
						Description: remoteIndicesPermissionsDescription,
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
									DiffSuppressFunc: tfsdkutils.DiffJSONSuppress,
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
					"run_as": {
						Description: "A list of usernames the owners of this role can impersonate.",
						Type:        schema.TypeSet,
						Optional:    true,
						Elem: &schema.Schema{
							Type: schema.TypeString,
						},
					},
				},
			},
		},
		"kibana": {
			Description: "The list of objects that specify the Kibana privileges for the role.",
			Type:        schema.TypeSet,
			Optional:    true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"base": {
						Description: "A base privilege. When specified, the base must be [\"all\"] or [\"read\"]. When the base privileges are specified, you are unable to use the \"feature\" section.",
						Type:        schema.TypeSet,
						Optional:    true,
						MaxItems:    1,
						Elem: &schema.Schema{
							Type:         schema.TypeString,
							ValidateFunc: validation.StringInSlice([]string{"all", "read"}, true),
						},
					},
					"feature": {
						Description: "List of privileges for specific features. When the feature privileges are specified, you are unable to use the \"base\" section.",
						Type:        schema.TypeSet,
						Optional:    true,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"name": {
									Description: "Feature name.",
									Type:        schema.TypeString,
									Required:    true,
								},
								"privileges": {
									Description: "Feature privileges.",
									Type:        schema.TypeSet,
									Required:    true,
									Elem: &schema.Schema{
										Type: schema.TypeString,
									},
								},
							},
						},
					},
					"spaces": {
						Description: "The spaces to apply the privileges to. To grant access to all spaces, set to [\"*\"], or omit the value.",
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
			DiffSuppressFunc: tfsdkutils.DiffJSONSuppress,
		},
		"description": {
			Description: "Optional description for the role",
			Type:        schema.TypeString,
			Optional:    true,
		},
		"kibana_connection": providerSchema.GetKibanaEntityConnectionSchema(),
	}

	return &schema.Resource{
		Description: roleDescription,

		CreateContext: resourceRoleUpsert,
		UpdateContext: resourceRoleUpsert,
		ReadContext:   resourceRoleRead,
		DeleteContext: resourceRoleDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: roleSchema,
	}
}

func resourceRoleUpsert(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	factory, diags := clients.ConvertMetaToFactory(meta)
	if diags.HasError() {
		return diags
	}
	client, diags := factory.GetKibanaClientFromSDK(d)
	if diags.HasError() {
		return diags
	}

	serverVersion, diags := client.ServerVersion(ctx)
	if diags.HasError() {
		return diags
	}

	oapiClient, err := client.GetKibanaOapiClient()
	if err != nil {
		return diag.FromErr(err)
	}

	roleName := d.Get("name").(string)

	body := kibanaoapi.SecurityRolePutBody{}

	if v, ok := d.GetOk("kibana"); ok {
		kibanaPrivs, ds := expandKibanaRoleKibana(v)
		if ds != nil {
			return ds
		}
		body.Kibana = kibanaPrivs
	}

	if v, ok := d.GetOk("elasticsearch"); ok {
		ds := expandKibanaRoleElasticsearchInto(v, serverVersion, &body.Elasticsearch)
		if ds != nil {
			return ds
		}
	}

	if v, ok := d.GetOk("metadata"); ok {
		metadata, ds := expandKibanaRoleMetadata(v)
		if ds != nil {
			return ds
		}
		body.Metadata = &metadata
	}

	if v, ok := d.GetOk("description"); ok {
		if serverVersion.LessThan(minSupportedDescriptionVersion) {
			return diag.FromErr(fmt.Errorf("'description' is supported only for Kibana v%s and above", minSupportedDescriptionVersion.String()))
		}
		desc := v.(string)
		body.Description = &desc
	}

	createOnly := d.IsNewResource()
	params := kbapi.PutSecurityRoleNameParams{
		CreateOnly: &createOnly,
	}

	if ds := kibanaoapi.PutSecurityRole(ctx, oapiClient, roleName, params, body); ds != nil {
		return ds
	}

	d.SetId(roleName)
	return resourceRoleRead(ctx, d, meta)
}

func resourceRoleRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	factory, diags := clients.ConvertMetaToFactory(meta)
	if diags.HasError() {
		return diags
	}
	client, diags := factory.GetKibanaClientFromSDK(d)
	if diags.HasError() {
		return diags
	}

	oapiClient, err := client.GetKibanaOapiClient()
	if err != nil {
		return diag.FromErr(err)
	}

	name := d.Id()

	role, ds := kibanaoapi.GetSecurityRole(ctx, oapiClient, name)
	if ds != nil {
		return ds
	}
	if role == nil {
		d.SetId("")
		return diags
	}

	if err := d.Set("name", name); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("elasticsearch", flattenKibanaRoleElasticsearchData(role)); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("kibana", flattenKibanaRoleKibanaData(role.Kibana)); err != nil {
		return diag.FromErr(err)
	}
	description := ""
	if role.Description != nil {
		description = *role.Description
	}
	if err := d.Set("description", description); err != nil {
		return diag.FromErr(err)
	}
	if role.Metadata != nil {
		metadata, err := json.Marshal(*role.Metadata)
		if err != nil {
			return diag.FromErr(err)
		}
		if err := d.Set("metadata", string(metadata)); err != nil {
			return diag.FromErr(err)
		}
	}
	return diags
}

func resourceRoleDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	factory, diags := clients.ConvertMetaToFactory(meta)
	if diags.HasError() {
		return diags
	}
	client, diags := factory.GetKibanaClientFromSDK(d)
	if diags.HasError() {
		return diags
	}

	oapiClient, err := client.GetKibanaOapiClient()
	if err != nil {
		return diag.FromErr(err)
	}

	resourceID := d.Id()

	if ds := kibanaoapi.DeleteSecurityRole(ctx, oapiClient, resourceID); ds != nil {
		return ds
	}

	d.SetId("")
	return diags
}

// Helper functions

func expandKibanaRoleMetadata(v any) (map[string]any, diag.Diagnostics) {
	metadata := make(map[string]any)
	if err := json.NewDecoder(strings.NewReader(v.(string))).Decode(&metadata); err != nil {
		return nil, diag.FromErr(err)
	}
	return metadata, nil
}

// expandKibanaRoleElasticsearchInto expands the elasticsearch block from Terraform
// state into a SecurityRoleES struct, modifying es in place.
func expandKibanaRoleElasticsearchInto(v any, serverVersion *version.Version, es *kibanaoapi.SecurityRoleES) diag.Diagnostics {
	if definedElasticConfigs := v.(*schema.Set); definedElasticConfigs.Len() > 0 {
		userElasticConfig := definedElasticConfigs.List()[0].(map[string]any)

		if v, ok := userElasticConfig["cluster"]; ok {
			definedCluster := v.(*schema.Set)
			if definedCluster.Len() > 0 {
				cls := make([]string, definedCluster.Len())
				for i, cl := range definedCluster.List() {
					cls[i] = cl.(string)
				}
				es.Cluster = &cls
			}
		}

		if v, ok := userElasticConfig["indices"]; ok {
			definedIndices := v.(*schema.Set)
			if definedIndices.Len() > 0 {
				indices := expandIndices(definedIndices)
				es.Indices = &indices
			}
		}

		if v, ok := userElasticConfig["remote_indices"]; ok {
			definedRemoteIndices := v.(*schema.Set)
			if definedRemoteIndices.Len() > 0 {
				if serverVersion.LessThan(minSupportedRemoteIndicesVersion) {
					return diag.FromErr(fmt.Errorf("'remote_indices' is supported only for Kibana v%s and above", minSupportedRemoteIndicesVersion.String()))
				}

				remoteIndices := expandRemoteIndices(definedRemoteIndices)
				es.RemoteIndices = &remoteIndices
			}
		}

		if v, ok := userElasticConfig["run_as"]; ok {
			definedRuns := v.(*schema.Set)
			if definedRuns.Len() > 0 {
				runs := make([]string, definedRuns.Len())
				for i, run := range definedRuns.List() {
					runs[i] = run.(string)
				}
				es.RunAs = &runs
			}
		}
	}
	return nil
}

func expandIndices(definedIndices *schema.Set) []kibanaoapi.SecurityRoleESIndex {
	indices := make([]kibanaoapi.SecurityRoleESIndex, definedIndices.Len())
	for i, idx := range definedIndices.List() {
		index := idx.(map[string]any)

		definedNames := index["names"].(*schema.Set)
		names := make([]string, definedNames.Len())
		for j, name := range definedNames.List() {
			names[j] = name.(string)
		}
		definedPrivileges := index["privileges"].(*schema.Set)
		privileges := make([]string, definedPrivileges.Len())
		for j, pr := range definedPrivileges.List() {
			privileges[j] = pr.(string)
		}

		entry := kibanaoapi.SecurityRoleESIndex{
			Names:      names,
			Privileges: privileges,
		}

		if query := index["query"].(string); query != "" {
			entry.Query = &query
		}
		if fieldSec := index["field_security"].([]any); len(fieldSec) > 0 {
			fieldSecurity := expandFieldSecurity(fieldSec[0].(map[string]any))
			if len(fieldSecurity) > 0 {
				entry.FieldSecurity = &fieldSecurity
			}
		}

		indices[i] = entry
	}
	return indices
}

func expandRemoteIndices(definedRemoteIndices *schema.Set) []kibanaoapi.SecurityRoleESRemoteIndex {
	remoteIndices := make([]kibanaoapi.SecurityRoleESRemoteIndex, definedRemoteIndices.Len())
	for i, idx := range definedRemoteIndices.List() {
		index := idx.(map[string]any)

		definedNames := index["names"].(*schema.Set)
		names := make([]string, definedNames.Len())
		for j, name := range definedNames.List() {
			names[j] = name.(string)
		}
		definedClusters := index["clusters"].(*schema.Set)
		clusters := make([]string, definedClusters.Len())
		for j, cluster := range definedClusters.List() {
			clusters[j] = cluster.(string)
		}
		definedPrivileges := index["privileges"].(*schema.Set)
		privileges := make([]string, definedPrivileges.Len())
		for j, pr := range definedPrivileges.List() {
			privileges[j] = pr.(string)
		}

		entry := kibanaoapi.SecurityRoleESRemoteIndex{
			Names:      names,
			Clusters:   clusters,
			Privileges: privileges,
		}

		if query := index["query"].(string); query != "" {
			entry.Query = &query
		}
		if fieldSec := index["field_security"].([]any); len(fieldSec) > 0 {
			fieldSecurity := expandFieldSecurity(fieldSec[0].(map[string]any))
			if len(fieldSecurity) > 0 {
				entry.FieldSecurity = &fieldSecurity
			}
		}

		remoteIndices[i] = entry
	}
	return remoteIndices
}

func expandFieldSecurity(definedFieldSec map[string]any) map[string][]string {
	fieldSecurity := map[string][]string{}
	if gr := definedFieldSec["grant"].(*schema.Set); gr != nil && gr.Len() > 0 {
		grants := make([]string, gr.Len())
		for i, grant := range gr.List() {
			grants[i] = grant.(string)
		}
		fieldSecurity["grant"] = grants
	}
	if exp := definedFieldSec["except"].(*schema.Set); exp != nil && exp.Len() > 0 {
		excepts := make([]string, exp.Len())
		for i, except := range exp.List() {
			excepts[i] = except.(string)
		}
		fieldSecurity["except"] = excepts
	}
	return fieldSecurity
}

func expandKibanaRoleKibana(v any) ([]kibanaoapi.SecurityRoleKibana, diag.Diagnostics) {
	definedKibanaConfigs := v.(*schema.Set)
	entries := make([]kibanaoapi.SecurityRoleKibana, 0, definedKibanaConfigs.Len())

	for _, item := range definedKibanaConfigs.List() {
		each := item.(map[string]any)
		entry := kibanaoapi.SecurityRoleKibana{}

		if basePrivileges, ok := each["base"].(*schema.Set); ok && basePrivileges.Len() > 0 {
			if features, ok := each["feature"].(*schema.Set); ok && features.Len() > 0 {
				return nil, diag.Errorf("Only one of the `feature` or `base` privileges allowed!")
			}
			base := make([]string, basePrivileges.Len())
			for i, name := range basePrivileges.List() {
				base[i] = name.(string)
			}
			baseJSON, err := json.Marshal(base)
			if err != nil {
				return nil, diag.FromErr(err)
			}
			entry.Base = json.RawMessage(baseJSON)
		} else if kibanaFeatures, ok := each["feature"].(*schema.Set); ok && kibanaFeatures.Len() > 0 {
			featureMap := map[string][]string{}
			for _, item := range kibanaFeatures.List() {
				featureData := item.(map[string]any)
				featurePrivileges := featureData["privileges"].(*schema.Set)
				privs := make([]string, featurePrivileges.Len())
				for i, f := range featurePrivileges.List() {
					privs[i] = f.(string)
				}
				featureMap[featureData["name"].(string)] = privs
			}
			entry.Feature = &featureMap
		} else {
			return nil, diag.Errorf("Either on of the `feature` or `base` privileges must be set for kibana role!")
		}

		if roleSpaces, ok := each["spaces"].(*schema.Set); ok && roleSpaces.Len() > 0 {
			spaces := make([]string, roleSpaces.Len())
			for i, name := range roleSpaces.List() {
				spaces[i] = name.(string)
			}
			entry.Spaces = &spaces
		}
		entries = append(entries, entry)
	}

	return entries, nil
}

func flattenKibanaRoleIndicesData(indices *[]kibanaoapi.SecurityRoleESIndex) []any {
	if indices == nil {
		return []any{}
	}
	oindx := make([]any, len(*indices))

	for i, index := range *indices {
		oi := make(map[string]any)
		oi["names"] = index.Names
		oi["privileges"] = index.Privileges
		if index.Query != nil {
			oi["query"] = *index.Query
		} else {
			oi["query"] = ""
		}

		if index.FieldSecurity != nil {
			fsec := make(map[string]any)
			if grantV, ok := (*index.FieldSecurity)["grant"]; ok {
				fsec["grant"] = grantV
			}
			if exceptV, ok := (*index.FieldSecurity)["except"]; ok {
				fsec["except"] = exceptV
			}
			oi["field_security"] = []any{fsec}
		}
		oindx[i] = oi
	}
	return oindx
}

func flattenKibanaRoleRemoteIndicesData(indices *[]kibanaoapi.SecurityRoleESRemoteIndex) []any {
	if indices == nil {
		return []any{}
	}
	oindx := make([]any, len(*indices))

	for i, index := range *indices {
		oi := make(map[string]any)
		oi["clusters"] = index.Clusters
		oi["names"] = index.Names
		oi["privileges"] = index.Privileges
		if index.Query != nil {
			oi["query"] = *index.Query
		} else {
			oi["query"] = ""
		}

		if index.FieldSecurity != nil {
			fsec := make(map[string]any)
			if grantV, ok := (*index.FieldSecurity)["grant"]; ok {
				fsec["grant"] = grantV
			}
			if exceptV, ok := (*index.FieldSecurity)["except"]; ok {
				fsec["except"] = exceptV
			}
			oi["field_security"] = []any{fsec}
		}
		oindx[i] = oi
	}
	return oindx
}

func flattenKibanaRoleElasticsearchData(role *kibanaoapi.SecurityRole) []any {
	result := make(map[string]any)
	es := role.Elasticsearch

	if es.Cluster != nil && len(*es.Cluster) > 0 {
		result["cluster"] = *es.Cluster
	}
	result["indices"] = flattenKibanaRoleIndicesData(es.Indices)
	result["remote_indices"] = flattenKibanaRoleRemoteIndicesData(es.RemoteIndices)
	if es.RunAs != nil && len(*es.RunAs) > 0 {
		result["run_as"] = *es.RunAs
	}
	return []any{result}
}

func flattenKibanaRoleKibanaFeatureData(features *map[string][]string) []any {
	if features == nil {
		return []any{}
	}
	result := make([]any, len(*features))
	i := 0
	for k, v := range *features {
		m := make(map[string]any)
		m["name"] = k
		m["privileges"] = v
		result[i] = m
		i++
	}
	return result
}

func flattenKibanaRoleKibanaData(kibanaConfigs []kibanaoapi.SecurityRoleKibana) []any {
	if kibanaConfigs == nil {
		return []any{}
	}

	result := make([]any, len(kibanaConfigs))
	for i, config := range kibanaConfigs {
		nk := make(map[string]any)

		// base is stored as json.RawMessage - decode to []string
		if len(config.Base) > 0 {
			var base []string
			if err := json.Unmarshal(config.Base, &base); err == nil && len(base) > 0 {
				nk["base"] = base
			}
		}
		if _, ok := nk["base"]; !ok {
			nk["base"] = []string{}
		}

		nk["feature"] = flattenKibanaRoleKibanaFeatureData(config.Feature)

		if config.Spaces != nil {
			nk["spaces"] = *config.Spaces
		} else {
			nk["spaces"] = []string{}
		}

		result[i] = nk
	}
	return result
}
