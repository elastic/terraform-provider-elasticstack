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
	"fmt"
	"sort"
	"strings"

	kibanaoapi "github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func normalizedQueryFromAPI(q *string) jsontypes.Normalized {
	if q == nil {
		return jsontypes.NewNormalizedNull()
	}
	if *q == "" {
		return jsontypes.NewNormalizedNull()
	}
	return jsontypes.NewNormalizedValue(*q)
}

// alignSetRepresentation reconciles "empty vs null" set drift between what
// the API can express (data or absent) and what the user's plan / prior
// state expects. Plugin Framework treats null and known-empty sets as
// distinct values for `SetNestedBlock` element identity, so a config that
// writes `base = []` must round-trip to `[]` while one that omits `base`
// must round-trip to null. When `fromAPI` carries data it wins; otherwise
// the representation is taken from `hint`.
func alignSetRepresentation(hint, fromAPI types.Set, elemType attr.Type) types.Set {
	if !fromAPI.IsNull() && !fromAPI.IsUnknown() && len(fromAPI.Elements()) > 0 {
		return fromAPI
	}
	if hint.IsNull() || hint.IsUnknown() {
		return types.SetNull(elemType)
	}
	if len(hint.Elements()) == 0 {
		return hint
	}
	return fromAPI
}

// setStringKey returns a canonical, order-independent key for a set of
// strings, used to match plan-side entries against API-side entries within
// the kibana / indices / remote_indices set blocks.
func setStringKey(s types.Set) string {
	if s.IsNull() || s.IsUnknown() {
		return ""
	}
	elems := s.Elements()
	parts := make([]string, 0, len(elems))
	for _, e := range elems {
		sv, ok := e.(types.String)
		if !ok {
			continue
		}
		parts = append(parts, sv.ValueString())
	}
	sort.Strings(parts)
	return strings.Join(parts, "\x00")
}

// indexHintSet builds a lookup of hint set objects keyed by `keyFn`. Objects
// in `hint` that don't decode to `types.Object` are skipped.
func indexHintSet(hint types.Set, keyFn func(types.Object) string) map[string]types.Object {
	out := map[string]types.Object{}
	if hint.IsNull() || hint.IsUnknown() {
		return out
	}
	for _, el := range hint.Elements() {
		obj, ok := el.(types.Object)
		if !ok {
			continue
		}
		out[keyFn(obj)] = obj
	}
	return out
}

func kibanaHintKey(obj types.Object) string {
	spaces, _ := obj.Attributes()["spaces"].(types.Set)
	return setStringKey(spaces)
}

func indicesHintKey(obj types.Object) string {
	names, _ := obj.Attributes()["names"].(types.Set)
	return setStringKey(names)
}

func remoteIndicesHintKey(obj types.Object) string {
	attrs := obj.Attributes()
	clusters, _ := attrs["clusters"].(types.Set)
	names, _ := attrs["names"].(types.Set)
	return setStringKey(clusters) + "|" + setStringKey(names)
}

// fieldSecurityHint extracts the `field_security` object attribute from an
// indices/remote_indices hint entry, returning a null object if absent.
func fieldSecurityHint(obj types.Object) types.Object {
	if obj.IsNull() || obj.IsUnknown() {
		return types.ObjectNull(fieldSecurityAttrTypes())
	}
	fs, ok := obj.Attributes()["field_security"].(types.Object)
	if !ok {
		return types.ObjectNull(fieldSecurityAttrTypes())
	}
	return fs
}

// objectFromFieldSecurityResource builds the resource-side field_security
// object. The Kibana API omits absent keys, so `hint` (via
// alignSetRepresentation) is consulted to decide null vs known-empty for
// missing `grant` / `except`. When the API returns no field_security at all,
// a known hint object is mirrored verbatim; otherwise a null object is
// returned (field_security is a SingleNestedBlock).
func objectFromFieldSecurityResource(ctx context.Context, fs *map[string][]string, hint types.Object) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics
	if fs == nil {
		if !hint.IsNull() && !hint.IsUnknown() {
			return hint, diags
		}
		return types.ObjectNull(fieldSecurityAttrTypes()), diags
	}
	grants := []string{}
	excepts := []string{}
	hasGrant := false
	hasExcept := false
	if g, ok := (*fs)["grant"]; ok {
		grants = g
		hasGrant = true
	}
	if e, ok := (*fs)["except"]; ok {
		excepts = e
		hasExcept = true
	}
	grantSet, d := types.SetValueFrom(ctx, types.StringType, grants)
	diags.Append(d...)
	if diags.HasError() {
		return types.ObjectNull(fieldSecurityAttrTypes()), diags
	}
	exceptSet, d := types.SetValueFrom(ctx, types.StringType, excepts)
	diags.Append(d...)
	if diags.HasError() {
		return types.ObjectNull(fieldSecurityAttrTypes()), diags
	}
	// Align null vs known-empty representation with the hint for keys the
	// API omitted. Keys the API actually returned are taken at face value.
	if !hasGrant {
		grantSet = alignSetRepresentation(objAttrSet(hint, "grant", types.StringType), grantSet, types.StringType)
	}
	if !hasExcept {
		exceptSet = alignSetRepresentation(objAttrSet(hint, "except", types.StringType), exceptSet, types.StringType)
	}
	obj, d := types.ObjectValue(fieldSecurityAttrTypes(), map[string]attr.Value{
		"grant":  grantSet,
		"except": exceptSet,
	})
	diags.Append(d...)
	return obj, diags
}

// flattenIndicesResource builds the resource-side indices set: optional
// `cluster`/`run_as`-like nested sets are null when absent, and field_security
// is a single object (or null) rather than a list. `hint` is the plan/state
// `indices` set used to preserve null-vs-empty representation on a per-entry
// basis, keyed by the entry's `names`.
func flattenIndicesResource(ctx context.Context, indices *[]kibanaoapi.SecurityRoleESIndex, hint types.Set) (types.Set, diag.Diagnostics) {
	var diags diag.Diagnostics
	objType := types.ObjectType{AttrTypes: esIndexResourceAttrTypes()}
	if indices == nil || len(*indices) == 0 {
		return types.SetNull(objType), diags
	}
	hintIdx := indexHintSet(hint, indicesHintKey)
	elems := make([]attr.Value, len(*indices))
	for i, index := range *indices {
		namesSet, d := types.SetValueFrom(ctx, types.StringType, index.Names)
		diags.Append(d...)
		if diags.HasError() {
			return types.SetNull(objType), diags
		}
		privSet, d := types.SetValueFrom(ctx, types.StringType, index.Privileges)
		diags.Append(d...)
		if diags.HasError() {
			return types.SetNull(objType), diags
		}
		entryHint := hintIdx[setStringKey(namesSet)]
		fieldObj, d := objectFromFieldSecurityResource(ctx, index.FieldSecurity, fieldSecurityHint(entryHint))
		diags.Append(d...)
		if diags.HasError() {
			return types.SetNull(objType), diags
		}
		obj, d := types.ObjectValue(esIndexResourceAttrTypes(), map[string]attr.Value{
			"names":          namesSet,
			"privileges":     privSet,
			"query":          normalizedQueryFromAPI(index.Query),
			"field_security": fieldObj,
		})
		diags.Append(d...)
		if diags.HasError() {
			return types.SetNull(objType), diags
		}
		elems[i] = obj
	}
	set, d := types.SetValue(objType, elems)
	diags.Append(d...)
	return set, diags
}

func flattenRemoteIndicesResource(ctx context.Context, indices *[]kibanaoapi.SecurityRoleESRemoteIndex, hint types.Set) (types.Set, diag.Diagnostics) {
	var diags diag.Diagnostics
	objType := types.ObjectType{AttrTypes: esRemoteIndexResourceAttrTypes()}
	if indices == nil || len(*indices) == 0 {
		return types.SetNull(objType), diags
	}
	hintIdx := indexHintSet(hint, remoteIndicesHintKey)
	elems := make([]attr.Value, len(*indices))
	for i, index := range *indices {
		clustersSet, d := types.SetValueFrom(ctx, types.StringType, index.Clusters)
		diags.Append(d...)
		if diags.HasError() {
			return types.SetNull(objType), diags
		}
		namesSet, d := types.SetValueFrom(ctx, types.StringType, index.Names)
		diags.Append(d...)
		if diags.HasError() {
			return types.SetNull(objType), diags
		}
		privSet, d := types.SetValueFrom(ctx, types.StringType, index.Privileges)
		diags.Append(d...)
		if diags.HasError() {
			return types.SetNull(objType), diags
		}
		entryHint := hintIdx[setStringKey(clustersSet)+"|"+setStringKey(namesSet)]
		fieldObj, d := objectFromFieldSecurityResource(ctx, index.FieldSecurity, fieldSecurityHint(entryHint))
		diags.Append(d...)
		if diags.HasError() {
			return types.SetNull(objType), diags
		}
		obj, d := types.ObjectValue(esRemoteIndexResourceAttrTypes(), map[string]attr.Value{
			"clusters":       clustersSet,
			"names":          namesSet,
			"privileges":     privSet,
			"query":          normalizedQueryFromAPI(index.Query),
			"field_security": fieldObj,
		})
		diags.Append(d...)
		if diags.HasError() {
			return types.SetNull(objType), diags
		}
		elems[i] = obj
	}
	set, d := types.SetValue(objType, elems)
	diags.Append(d...)
	return set, diags
}

// flattenElasticsearchObject is the resource-side flattener: returns a single
// elasticsearch object (matching the SingleNestedBlock schema) with all
// SDK-legacy optional sets normalised to null when the API omits them. `hint`
// is the plan/state `elasticsearch` object used to preserve null-vs-empty
// representation for `cluster`, `run_as`, and per-entry field_security.
func flattenElasticsearchObject(ctx context.Context, es *kibanaoapi.SecurityRoleES, hint types.Object) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics
	attrs := map[string]attr.Value{
		"cluster":        types.SetNull(types.StringType),
		"run_as":         types.SetNull(types.StringType),
		"indices":        types.SetNull(types.ObjectType{AttrTypes: esIndexResourceAttrTypes()}),
		"remote_indices": types.SetNull(types.ObjectType{AttrTypes: esRemoteIndexResourceAttrTypes()}),
	}
	clusterFromAPI := types.SetNull(types.StringType)
	if es.Cluster != nil && len(*es.Cluster) > 0 {
		s, d := types.SetValueFrom(ctx, types.StringType, *es.Cluster)
		diags.Append(d...)
		if diags.HasError() {
			return types.ObjectNull(elasticsearchResourceAttrTypes()), diags
		}
		clusterFromAPI = s
	}
	attrs["cluster"] = alignSetRepresentation(objAttrSet(hint, "cluster", types.StringType), clusterFromAPI, types.StringType)

	runAsFromAPI := types.SetNull(types.StringType)
	if es.RunAs != nil && len(*es.RunAs) > 0 {
		s, d := types.SetValueFrom(ctx, types.StringType, *es.RunAs)
		diags.Append(d...)
		if diags.HasError() {
			return types.ObjectNull(elasticsearchResourceAttrTypes()), diags
		}
		runAsFromAPI = s
	}
	attrs["run_as"] = alignSetRepresentation(objAttrSet(hint, "run_as", types.StringType), runAsFromAPI, types.StringType)

	indicesHint := objAttrSet(hint, "indices", types.ObjectType{AttrTypes: esIndexResourceAttrTypes()})
	remoteHint := objAttrSet(hint, "remote_indices", types.ObjectType{AttrTypes: esRemoteIndexResourceAttrTypes()})
	indicesSet, d := flattenIndicesResource(ctx, es.Indices, indicesHint)
	diags.Append(d...)
	if diags.HasError() {
		return types.ObjectNull(elasticsearchResourceAttrTypes()), diags
	}
	attrs["indices"] = indicesSet
	remoteSet, d := flattenRemoteIndicesResource(ctx, es.RemoteIndices, remoteHint)
	diags.Append(d...)
	if diags.HasError() {
		return types.ObjectNull(elasticsearchResourceAttrTypes()), diags
	}
	attrs["remote_indices"] = remoteSet
	obj, d := types.ObjectValue(elasticsearchResourceAttrTypes(), attrs)
	diags.Append(d...)
	return obj, diags
}

func flattenKibanaFeatures(ctx context.Context, features *map[string][]string) (types.Set, diag.Diagnostics) {
	var diags diag.Diagnostics
	if features == nil || len(*features) == 0 {
		return types.SetNull(types.ObjectType{AttrTypes: kibanaFeatureAttrTypes()}), diags
	}
	elems := make([]attr.Value, 0, len(*features))
	for k, privs := range *features {
		privSet, d := types.SetValueFrom(ctx, types.StringType, privs)
		diags.Append(d...)
		if diags.HasError() {
			return types.SetNull(types.ObjectType{AttrTypes: kibanaFeatureAttrTypes()}), diags
		}
		obj, d := types.ObjectValue(kibanaFeatureAttrTypes(), map[string]attr.Value{
			"name":       types.StringValue(k),
			"privileges": privSet,
		})
		diags.Append(d...)
		if diags.HasError() {
			return types.SetNull(types.ObjectType{AttrTypes: kibanaFeatureAttrTypes()}), diags
		}
		elems = append(elems, obj)
	}
	set, d := types.SetValue(types.ObjectType{AttrTypes: kibanaFeatureAttrTypes()}, elems)
	diags.Append(d...)
	return set, diags
}

// flattenKibana builds the resource-side kibana set. `hint` is the plan/state
// kibana set used to preserve the null-vs-empty representation of the
// optional `base` attribute, matched by `spaces`.
func flattenKibana(ctx context.Context, configs []kibanaoapi.SecurityRoleKibana, hint types.Set) (types.Set, diag.Diagnostics) {
	var diags diag.Diagnostics
	if len(configs) == 0 {
		return types.SetValueMust(kibanaBlockObjectType(), []attr.Value{}), diags
	}
	hintIdx := indexHintSet(hint, kibanaHintKey)
	elems := make([]attr.Value, len(configs))
	for i, cfg := range configs {
		baseFromAPI := types.SetNull(types.StringType)
		if len(cfg.Base) > 0 {
			var base []string
			if err := json.Unmarshal(cfg.Base, &base); err != nil {
				diags.AddError(
					"Invalid kibana base privileges",
					fmt.Sprintf("API returned a base payload that is not a JSON array of strings: %v", err),
				)
				return types.SetNull(kibanaBlockObjectType()), diags
			}
			if len(base) > 0 {
				s, d := types.SetValueFrom(ctx, types.StringType, base)
				diags.Append(d...)
				if diags.HasError() {
					return types.SetNull(kibanaBlockObjectType()), diags
				}
				baseFromAPI = s
			}
		}
		featureSet, d := flattenKibanaFeatures(ctx, cfg.Feature)
		diags.Append(d...)
		if diags.HasError() {
			return types.SetNull(kibanaBlockObjectType()), diags
		}
		var spacesSet types.Set
		if cfg.Spaces != nil && len(*cfg.Spaces) > 0 {
			s, d := types.SetValueFrom(ctx, types.StringType, *cfg.Spaces)
			diags.Append(d...)
			if diags.HasError() {
				return types.SetNull(kibanaBlockObjectType()), diags
			}
			spacesSet = s
		} else {
			spacesSet = types.SetValueMust(types.StringType, []attr.Value{})
		}
		entryHint := hintIdx[setStringKey(spacesSet)]
		baseSet := alignSetRepresentation(objAttrSet(entryHint, "base", types.StringType), baseFromAPI, types.StringType)
		obj, d := types.ObjectValue(kibanaBlockAttrTypes(), map[string]attr.Value{
			"spaces":  spacesSet,
			"base":    baseSet,
			"feature": featureSet,
		})
		diags.Append(d...)
		if diags.HasError() {
			return types.SetNull(kibanaBlockObjectType()), diags
		}
		elems[i] = obj
	}
	set, d := types.SetValue(kibanaBlockObjectType(), elems)
	diags.Append(d...)
	return set, diags
}

func metadataFromAPI(role *kibanaoapi.SecurityRole) (jsontypes.Normalized, diag.Diagnostics) {
	var diags diag.Diagnostics
	if role.Metadata == nil {
		return jsontypes.NewNormalizedNull(), diags
	}
	b, err := json.Marshal(*role.Metadata)
	if err != nil {
		diags.AddError("Failed to marshal role metadata", err.Error())
		return jsontypes.NewNormalizedNull(), diags
	}
	return jsontypes.NewNormalizedValue(string(b)), diags
}
