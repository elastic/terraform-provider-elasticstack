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

package security_role

import (
	"context"
	"encoding/json"

	kibanaoapi "github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// objAttrSet fetches a `types.Set` attribute from a decoded object. Returns
// a null set of the supplied element type if the attribute is missing or has
// an unexpected shape (the schema guarantees this won't happen at runtime).
func objAttrSet(obj types.Object, name string, elemType attr.Type) types.Set {
	if obj.IsNull() || obj.IsUnknown() {
		return types.SetNull(elemType)
	}
	s, ok := obj.Attributes()[name].(types.Set)
	if !ok {
		return types.SetNull(elemType)
	}
	return s
}

// kibanaPrivilegeCounts returns the `base` / `feature` sets from a decoded
// kibana block object and their element counts, treating null/unknown as 0.
func kibanaPrivilegeCounts(obj types.Object) (base, feature types.Set, baseLen, featureLen int) {
	base = objAttrSet(obj, "base", types.StringType)
	feature = objAttrSet(obj, "feature", types.ObjectType{AttrTypes: kibanaFeatureAttrTypes()})
	if !base.IsNull() && !base.IsUnknown() {
		baseLen = len(base.Elements())
	}
	if !feature.IsNull() && !feature.IsUnknown() {
		featureLen = len(feature.Elements())
	}
	return
}

func expandFieldSecurity(ctx context.Context, obj types.Object) (map[string][]string, diag.Diagnostics) {
	var diags diag.Diagnostics
	if obj.IsNull() || obj.IsUnknown() {
		return map[string][]string{}, diags
	}
	grant := objAttrSet(obj, "grant", types.StringType)
	except := objAttrSet(obj, "except", types.StringType)
	out := map[string][]string{}
	if !grant.IsNull() && !grant.IsUnknown() && len(grant.Elements()) > 0 {
		var grants []string
		diags.Append(grant.ElementsAs(ctx, &grants, false)...)
		out["grant"] = grants
	}
	if !except.IsNull() && !except.IsUnknown() && len(except.Elements()) > 0 {
		var excepts []string
		diags.Append(except.ElementsAs(ctx, &excepts, false)...)
		out["except"] = excepts
	}
	return out, diags
}

// expandedEntry captures the fields shared between `indices` and
// `remote_indices` entries; `Clusters` is only populated for the remote
// variant.
type expandedEntry struct {
	Names                  []string
	Clusters               []string
	Privileges             []string
	Query                  *string
	AllowRestrictedIndices *bool
	FS                     *map[string][]string
}

// expandEntryCommon reads names/privileges/query/field_security (and
// optionally clusters) from a decoded entry object. Uses direct attribute
// access rather than `obj.As` so the same code serves the `indices` and
// `remote_indices` schemas, which differ only in the presence of `clusters`.
func expandEntryCommon(ctx context.Context, obj types.Object, wantClusters bool) (expandedEntry, diag.Diagnostics) {
	var (
		diags diag.Diagnostics
		out   expandedEntry
	)
	diags.Append(objAttrSet(obj, "names", types.StringType).ElementsAs(ctx, &out.Names, false)...)
	diags.Append(objAttrSet(obj, "privileges", types.StringType).ElementsAs(ctx, &out.Privileges, false)...)
	if wantClusters {
		diags.Append(objAttrSet(obj, "clusters", types.StringType).ElementsAs(ctx, &out.Clusters, false)...)
	}
	if diags.HasError() {
		return expandedEntry{}, diags
	}
	if q, ok := obj.Attributes()["query"].(jsontypes.Normalized); ok && typeutils.IsKnown(q) && q.ValueString() != "" {
		v := q.ValueString()
		out.Query = &v
	}
	if fs, ok := obj.Attributes()["field_security"].(types.Object); ok && !fs.IsNull() && !fs.IsUnknown() {
		fsMap, d := expandFieldSecurity(ctx, fs)
		diags.Append(d...)
		if diags.HasError() {
			return expandedEntry{}, diags
		}
		if len(fsMap) > 0 {
			out.FS = &fsMap
		}
	}
	if ari, ok := obj.Attributes()[attrAllowRestrictedIndices].(types.Bool); ok && typeutils.IsKnown(ari) {
		v := ari.ValueBool()
		out.AllowRestrictedIndices = &v
	}
	return out, diags
}

func expandIndexEntry(ctx context.Context, obj types.Object) (kibanaoapi.SecurityRoleESIndex, diag.Diagnostics) {
	e, diags := expandEntryCommon(ctx, obj, false)
	if diags.HasError() {
		return kibanaoapi.SecurityRoleESIndex{}, diags
	}
	return kibanaoapi.SecurityRoleESIndex{
		Names:         e.Names,
		Privileges:    e.Privileges,
		Query:         e.Query,
		FieldSecurity: e.FS,
	}, diags
}

func expandRemoteEntry(ctx context.Context, obj types.Object) (kibanaoapi.SecurityRoleESRemoteIndex, diag.Diagnostics) {
	e, diags := expandEntryCommon(ctx, obj, true)
	if diags.HasError() {
		return kibanaoapi.SecurityRoleESRemoteIndex{}, diags
	}
	return kibanaoapi.SecurityRoleESRemoteIndex{
		Names:                  e.Names,
		Clusters:               e.Clusters,
		Privileges:             e.Privileges,
		Query:                  e.Query,
		AllowRestrictedIndices: e.AllowRestrictedIndices,
		FieldSecurity:          e.FS,
	}, diags
}

// expandObjectSet iterates over a set of typed objects, calling expandFn on
// each element. Returns nil when the set is null, unknown, or empty so callers
// can omit the pointer field from the API body unchanged.
func expandObjectSet[T any](
	ctx context.Context,
	s types.Set,
	expandFn func(context.Context, types.Object) (T, diag.Diagnostics),
	errLabel string,
) ([]T, diag.Diagnostics) {
	var diags diag.Diagnostics
	if s.IsNull() || s.IsUnknown() || len(s.Elements()) == 0 {
		return nil, diags
	}
	elems := s.Elements()
	out := make([]T, len(elems))
	for i, el := range elems {
		obj, ok := el.(types.Object)
		if !ok {
			diags.AddError(errLabel, "unexpected element type")
			return nil, diags
		}
		v, d := expandFn(ctx, obj)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		out[i] = v
	}
	return out, diags
}

// expandStringSlicePtr extracts a `*[]string` from an optional set
// attribute, returning nil when the set is null/unknown/empty so the API
// body omits the key.
func expandStringSlicePtr(ctx context.Context, s types.Set) (*[]string, diag.Diagnostics) {
	var diags diag.Diagnostics
	if s.IsNull() || s.IsUnknown() || len(s.Elements()) == 0 {
		return nil, diags
	}
	var out []string
	diags.Append(s.ElementsAs(ctx, &out, false)...)
	if diags.HasError() {
		return nil, diags
	}
	return &out, diags
}

func expandElasticsearch(ctx context.Context, obj types.Object) (kibanaoapi.SecurityRoleES, diag.Diagnostics) {
	var diags diag.Diagnostics
	var out kibanaoapi.SecurityRoleES
	if obj.IsNull() || obj.IsUnknown() {
		return out, diags
	}

	cluster, d := expandStringSlicePtr(ctx, objAttrSet(obj, "cluster", types.StringType))
	diags.Append(d...)
	if diags.HasError() {
		return out, diags
	}
	out.Cluster = cluster

	runAs, d := expandStringSlicePtr(ctx, objAttrSet(obj, "run_as", types.StringType))
	diags.Append(d...)
	if diags.HasError() {
		return out, diags
	}
	out.RunAs = runAs

	indicesSet := objAttrSet(obj, "indices", types.ObjectType{AttrTypes: esIndexResourceAttrTypes()})
	indices, d := expandObjectSet(ctx, indicesSet, expandIndexEntry, "Invalid indices entry")
	diags.Append(d...)
	if diags.HasError() {
		return out, diags
	}
	if indices != nil {
		out.Indices = &indices
	}

	remoteSet := objAttrSet(obj, "remote_indices", types.ObjectType{AttrTypes: esRemoteIndexResourceAttrTypes()})
	remote, d := expandObjectSet(ctx, remoteSet, expandRemoteEntry, "Invalid remote_indices entry")
	diags.Append(d...)
	if diags.HasError() {
		return out, diags
	}
	if remote != nil {
		out.RemoteIndices = &remote
	}
	return out, diags
}

func expandKibana(ctx context.Context, set types.Set) ([]kibanaoapi.SecurityRoleKibana, diag.Diagnostics) {
	var diags diag.Diagnostics
	elems := set.Elements()
	if set.IsNull() || set.IsUnknown() || len(elems) == 0 {
		return nil, diags
	}
	entries := make([]kibanaoapi.SecurityRoleKibana, 0, len(elems))
	for _, el := range elems {
		obj, ok := el.(types.Object)
		if !ok {
			diags.AddError("Invalid kibana block", "unexpected element type")
			return nil, diags
		}
		base, feature, baseLen, featureLen := kibanaPrivilegeCounts(obj)
		diags.Append(validateKibanaPrivileges(baseLen, featureLen)...)
		if diags.HasError() {
			return nil, diags
		}

		entry := kibanaoapi.SecurityRoleKibana{}
		switch {
		case baseLen > 0:
			var basePrivs []string
			diags.Append(base.ElementsAs(ctx, &basePrivs, false)...)
			if diags.HasError() {
				return nil, diags
			}
			raw, err := json.Marshal(basePrivs)
			if err != nil {
				diags.AddError("Failed to serialize kibana base privileges", err.Error())
				return nil, diags
			}
			entry.Base = raw
		case featureLen > 0:
			featureMap := map[string][]string{}
			for _, fe := range feature.Elements() {
				fObj, ok := fe.(types.Object)
				if !ok {
					diags.AddError("Invalid kibana feature block", "unexpected element type")
					return nil, diags
				}
				name, _ := fObj.Attributes()["name"].(types.String)
				privsSet := objAttrSet(fObj, "privileges", types.StringType)
				var privs []string
				diags.Append(privsSet.ElementsAs(ctx, &privs, false)...)
				if diags.HasError() {
					return nil, diags
				}
				featureMap[name.ValueString()] = privs
			}
			entry.Feature = &featureMap
		}

		spaces, d := expandStringSlicePtr(ctx, objAttrSet(obj, "spaces", types.StringType))
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		entry.Spaces = spaces
		entries = append(entries, entry)
	}
	return entries, diags
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

	if meta := typeutils.NormalizedTypeToMap[any](m.Metadata, path.Root("metadata"), &diags); diags.HasError() {
		return roleName, body, diags
	} else if meta != nil {
		body.Metadata = &meta
	}

	if typeutils.IsKnown(m.Description) && m.Description.ValueString() != "" {
		desc := m.Description.ValueString()
		body.Description = &desc
	}

	return roleName, body, diags
}
