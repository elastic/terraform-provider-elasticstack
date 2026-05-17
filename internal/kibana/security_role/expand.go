// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
//
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

package security_role

import (
	"context"
	"encoding/json"

	kibanaoapi "github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

type esIndexPlan struct {
	Names         types.Set            `tfsdk:"names"`
	Privileges    types.Set            `tfsdk:"privileges"`
	Query         jsontypes.Normalized `tfsdk:"query"`
	FieldSecurity types.List           `tfsdk:"field_security"`
}

type esRemotePlan struct {
	Clusters       types.Set            `tfsdk:"clusters"`
	Names          types.Set            `tfsdk:"names"`
	Privileges     types.Set            `tfsdk:"privileges"`
	Query          jsontypes.Normalized `tfsdk:"query"`
	FieldSecurity  types.List           `tfsdk:"field_security"`
}

type esBlockPlan struct {
	Cluster       types.Set `tfsdk:"cluster"`
	Indices       types.Set `tfsdk:"indices"`
	RemoteIndices types.Set `tfsdk:"remote_indices"`
	RunAs         types.Set `tfsdk:"run_as"`
}

type kibanaFeaturePlan struct {
	Name       types.String `tfsdk:"name"`
	Privileges types.Set    `tfsdk:"privileges"`
}

type kibanaBlockPlan struct {
	Spaces  types.Set `tfsdk:"spaces"`
	Base    types.Set `tfsdk:"base"`
	Feature types.Set `tfsdk:"feature"`
}

func expandFieldSecurity(ctx context.Context, obj types.Object) (map[string][]string, diag.Diagnostics) {
	var diags diag.Diagnostics
	if obj.IsNull() || obj.IsUnknown() {
		return map[string][]string{}, diags
	}
	var fs struct {
		Grant  types.Set `tfsdk:"grant"`
		Except types.Set `tfsdk:"except"`
	}
	diags.Append(obj.As(ctx, &fs, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil, diags
	}
	out := map[string][]string{}
	if !fs.Grant.IsNull() && !fs.Grant.IsUnknown() && len(fs.Grant.Elements()) > 0 {
		var grants []string
		diags.Append(fs.Grant.ElementsAs(ctx, &grants, false)...)
		out["grant"] = grants
	}
	if !fs.Except.IsNull() && !fs.Except.IsUnknown() && len(fs.Except.Elements()) > 0 {
		var excepts []string
		diags.Append(fs.Except.ElementsAs(ctx, &excepts, false)...)
		out["except"] = excepts
	}
	return out, diags
}

func expandFieldSecurityFromList(ctx context.Context, list types.List) (map[string][]string, diag.Diagnostics) {
	if list.IsNull() || list.IsUnknown() || len(list.Elements()) == 0 {
		return map[string][]string{}, nil
	}
	first, ok := list.Elements()[0].(types.Object)
	if !ok {
		var diags diag.Diagnostics
		diags.AddError("Invalid field_security", "expected object element")
		return nil, diags
	}
	return expandFieldSecurity(ctx, first)
}

func expandIndexEntry(ctx context.Context, obj types.Object) (kibanaoapi.SecurityRoleESIndex, diag.Diagnostics) {
	var diags diag.Diagnostics
	var row esIndexPlan
	diags.Append(obj.As(ctx, &row, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return kibanaoapi.SecurityRoleESIndex{}, diags
	}
	var names, privs []string
	diags.Append(row.Names.ElementsAs(ctx, &names, false)...)
	diags.Append(row.Privileges.ElementsAs(ctx, &privs, false)...)
	if diags.HasError() {
		return kibanaoapi.SecurityRoleESIndex{}, diags
	}
	entry := kibanaoapi.SecurityRoleESIndex{
		Names:      names,
		Privileges: privs,
	}
	if typeutils.IsKnown(row.Query) && row.Query.ValueString() != "" {
		q := row.Query.ValueString()
		entry.Query = &q
	}
	if !row.FieldSecurity.IsNull() && !row.FieldSecurity.IsUnknown() {
		fsMap, d := expandFieldSecurityFromList(ctx, row.FieldSecurity)
		diags.Append(d...)
		if diags.HasError() {
			return kibanaoapi.SecurityRoleESIndex{}, diags
		}
		if len(fsMap) > 0 {
			entry.FieldSecurity = &fsMap
		}
	}
	return entry, diags
}

func expandRemoteEntry(ctx context.Context, obj types.Object) (kibanaoapi.SecurityRoleESRemoteIndex, diag.Diagnostics) {
	var diags diag.Diagnostics
	var row esRemotePlan
	diags.Append(obj.As(ctx, &row, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return kibanaoapi.SecurityRoleESRemoteIndex{}, diags
	}
	var names, clusters, privs []string
	diags.Append(row.Names.ElementsAs(ctx, &names, false)...)
	diags.Append(row.Clusters.ElementsAs(ctx, &clusters, false)...)
	diags.Append(row.Privileges.ElementsAs(ctx, &privs, false)...)
	if diags.HasError() {
		return kibanaoapi.SecurityRoleESRemoteIndex{}, diags
	}
	entry := kibanaoapi.SecurityRoleESRemoteIndex{
		Names:      names,
		Clusters:   clusters,
		Privileges: privs,
	}
	if typeutils.IsKnown(row.Query) && row.Query.ValueString() != "" {
		q := row.Query.ValueString()
		entry.Query = &q
	}
	if !row.FieldSecurity.IsNull() && !row.FieldSecurity.IsUnknown() {
		fsMap, d := expandFieldSecurityFromList(ctx, row.FieldSecurity)
		diags.Append(d...)
		if diags.HasError() {
			return kibanaoapi.SecurityRoleESRemoteIndex{}, diags
		}
		if len(fsMap) > 0 {
			entry.FieldSecurity = &fsMap
		}
	}
	return entry, diags
}

func expandElasticsearch(ctx context.Context, set types.Set) (kibanaoapi.SecurityRoleES, diag.Diagnostics) {
	var diags diag.Diagnostics
	var out kibanaoapi.SecurityRoleES
	if set.IsNull() || set.IsUnknown() || len(set.Elements()) == 0 {
		return out, diags
	}
	elems := set.Elements()
	if len(elems) != 1 {
		diags.AddError("Invalid elasticsearch block", "expected exactly one elasticsearch block")
		return out, diags
	}
	obj, ok := elems[0].(types.Object)
	if !ok {
		diags.AddError("Invalid elasticsearch block", "unexpected element type")
		return out, diags
	}
	var block esBlockPlan
	diags.Append(obj.As(ctx, &block, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return out, diags
	}

	if !block.Cluster.IsNull() && !block.Cluster.IsUnknown() && len(block.Cluster.Elements()) > 0 {
		var cluster []string
		diags.Append(block.Cluster.ElementsAs(ctx, &cluster, false)...)
		if diags.HasError() {
			return out, diags
		}
		out.Cluster = &cluster
	}
	if !block.RunAs.IsNull() && !block.RunAs.IsUnknown() && len(block.RunAs.Elements()) > 0 {
		var runs []string
		diags.Append(block.RunAs.ElementsAs(ctx, &runs, false)...)
		if diags.HasError() {
			return out, diags
		}
		out.RunAs = &runs
	}
	if !block.Indices.IsNull() && !block.Indices.IsUnknown() && len(block.Indices.Elements()) > 0 {
		idxElems := block.Indices.Elements()
		indices := make([]kibanaoapi.SecurityRoleESIndex, len(idxElems))
		for i, el := range idxElems {
			idxObj, ok := el.(types.Object)
			if !ok {
				diags.AddError("Invalid indices entry", "unexpected element type")
				return out, diags
			}
			idx, d := expandIndexEntry(ctx, idxObj)
			diags.Append(d...)
			if diags.HasError() {
				return out, diags
			}
			indices[i] = idx
		}
		out.Indices = &indices
	}
	if !block.RemoteIndices.IsNull() && !block.RemoteIndices.IsUnknown() && len(block.RemoteIndices.Elements()) > 0 {
		riElems := block.RemoteIndices.Elements()
		remote := make([]kibanaoapi.SecurityRoleESRemoteIndex, len(riElems))
		for i, el := range riElems {
			riObj, ok := el.(types.Object)
			if !ok {
				diags.AddError("Invalid remote_indices entry", "unexpected element type")
				return out, diags
			}
			ri, d := expandRemoteEntry(ctx, riObj)
			diags.Append(d...)
			if diags.HasError() {
				return out, diags
			}
			remote[i] = ri
		}
		out.RemoteIndices = &remote
	}
	return out, diags
}

func expandKibana(ctx context.Context, set types.Set) ([]kibanaoapi.SecurityRoleKibana, diag.Diagnostics) {
	var diags diag.Diagnostics
	if set.IsNull() || set.IsUnknown() || len(set.Elements()) == 0 {
		return nil, diags
	}
	var entries []kibanaoapi.SecurityRoleKibana
	for _, el := range set.Elements() {
		obj, ok := el.(types.Object)
		if !ok {
			diags.AddError("Invalid kibana block", "unexpected element type")
			return nil, diags
		}
		var block kibanaBlockPlan
		diags.Append(obj.As(ctx, &block, basetypes.ObjectAsOptions{})...)
		if diags.HasError() {
			return nil, diags
		}
		entry := kibanaoapi.SecurityRoleKibana{}
		baseLen := 0
		if !block.Base.IsNull() && !block.Base.IsUnknown() {
			baseLen = len(block.Base.Elements())
		}
		featureLen := 0
		if !block.Feature.IsNull() && !block.Feature.IsUnknown() {
			featureLen = len(block.Feature.Elements())
		}
		if baseLen > 0 && featureLen > 0 {
			diags.AddError(
				"Invalid kibana privileges",
				"Only one of the `feature` or `base` privileges allowed!",
			)
			return nil, diags
		}
		if baseLen > 0 {
			var base []string
			diags.Append(block.Base.ElementsAs(ctx, &base, false)...)
			if diags.HasError() {
				return nil, diags
			}
			raw, err := json.Marshal(base)
			if err != nil {
				diags.AddError("Failed to serialize kibana base privileges", err.Error())
				return nil, diags
			}
			entry.Base = raw
		} else if featureLen > 0 {
			featureMap := map[string][]string{}
			for _, fe := range block.Feature.Elements() {
				fObj, ok := fe.(types.Object)
				if !ok {
					diags.AddError("Invalid kibana feature block", "unexpected element type")
					return nil, diags
				}
				var f kibanaFeaturePlan
				diags.Append(fObj.As(ctx, &f, basetypes.ObjectAsOptions{})...)
				if diags.HasError() {
					return nil, diags
				}
				var privs []string
				diags.Append(f.Privileges.ElementsAs(ctx, &privs, false)...)
				if diags.HasError() {
					return nil, diags
				}
				featureMap[f.Name.ValueString()] = privs
			}
			entry.Feature = &featureMap
		} else {
			diags.AddError(
				"Invalid kibana privileges",
				"Either on of the `feature` or `base` privileges must be set for kibana role!",
			)
			return nil, diags
		}
		if !block.Spaces.IsNull() && !block.Spaces.IsUnknown() && len(block.Spaces.Elements()) > 0 {
			var spaces []string
			diags.Append(block.Spaces.ElementsAs(ctx, &spaces, false)...)
			if diags.HasError() {
				return nil, diags
			}
			entry.Spaces = &spaces
		}
		entries = append(entries, entry)
	}
	return entries, diags
}

func expandMetadata(meta jsontypes.Normalized) (*map[string]any, diag.Diagnostics) {
	var diags diag.Diagnostics
	if !typeutils.IsKnown(meta) || meta.ValueString() == "" {
		return nil, diags
	}
	var m map[string]any
	if err := json.Unmarshal([]byte(meta.ValueString()), &m); err != nil {
		diags.AddError("Invalid metadata JSON", err.Error())
		return nil, diags
	}
	return &m, diags
}

func expandResourceModel(ctx context.Context, m resourceModel) (string, kibanaoapi.SecurityRolePutBody, diag.Diagnostics) {
	var diags diag.Diagnostics
	var body kibanaoapi.SecurityRolePutBody
	roleName := m.Name.ValueString()

	es, d := expandElasticsearch(ctx, m.Elasticsearch)
	diags.Append(d...)
	if diags.HasError() {
		return roleName, body, diags
	}
	body.Elasticsearch = es

	if !m.Kibana.IsNull() && !m.Kibana.IsUnknown() && len(m.Kibana.Elements()) > 0 {
		kib, d := expandKibana(ctx, m.Kibana)
		diags.Append(d...)
		if diags.HasError() {
			return roleName, body, diags
		}
		body.Kibana = kib
	}

	if typeutils.IsKnown(m.Metadata) && m.Metadata.ValueString() != "" {
		meta, d := expandMetadata(m.Metadata)
		diags.Append(d...)
		if diags.HasError() {
			return roleName, body, diags
		}
		body.Metadata = meta
	}

	if typeutils.IsKnown(m.Description) && m.Description.ValueString() != "" {
		desc := m.Description.ValueString()
		body.Description = &desc
	}

	return roleName, body, diags
}
