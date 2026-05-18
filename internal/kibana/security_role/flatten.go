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

// objectFromFieldSecurityResource builds the resource-side field_security
// object. The Kibana API omits keys it has no value for (and never returns
// them as `null`), but configs commonly set `except = []` or `grant = []`
// explicitly. To keep Create→Read consistent for those configs we normalise
// missing keys to known-empty sets rather than null. A nil API value yields
// ObjectNull (field_security itself is a SingleNestedBlock, so absence is
// represented as a null object).
func objectFromFieldSecurityResource(ctx context.Context, fs *map[string][]string) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics
	if fs == nil {
		return types.ObjectNull(fieldSecurityAttrTypes()), diags
	}
	grants := []string{}
	excepts := []string{}
	if g, ok := (*fs)["grant"]; ok {
		grants = g
	}
	if e, ok := (*fs)["except"]; ok {
		excepts = e
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
	obj, d := types.ObjectValue(fieldSecurityAttrTypes(), map[string]attr.Value{
		"grant":  grantSet,
		"except": exceptSet,
	})
	diags.Append(d...)
	return obj, diags
}

// flattenIndicesResource builds the resource-side indices set: optional
// `cluster`/`run_as`-like nested sets are null when absent, and field_security
// is a single object (or null) rather than a list.
func flattenIndicesResource(ctx context.Context, indices *[]kibanaoapi.SecurityRoleESIndex) (types.Set, diag.Diagnostics) {
	var diags diag.Diagnostics
	objType := types.ObjectType{AttrTypes: esIndexResourceAttrTypes()}
	if indices == nil || len(*indices) == 0 {
		return types.SetNull(objType), diags
	}
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
		fieldObj, d := objectFromFieldSecurityResource(ctx, index.FieldSecurity)
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

func flattenRemoteIndicesResource(ctx context.Context, indices *[]kibanaoapi.SecurityRoleESRemoteIndex) (types.Set, diag.Diagnostics) {
	var diags diag.Diagnostics
	objType := types.ObjectType{AttrTypes: esRemoteIndexResourceAttrTypes()}
	if indices == nil || len(*indices) == 0 {
		return types.SetNull(objType), diags
	}
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
		fieldObj, d := objectFromFieldSecurityResource(ctx, index.FieldSecurity)
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
// SDK-legacy optional sets normalised to null when the API omits them.
func flattenElasticsearchObject(ctx context.Context, es *kibanaoapi.SecurityRoleES) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics
	attrs := map[string]attr.Value{
		"cluster":        types.SetNull(types.StringType),
		"run_as":         types.SetNull(types.StringType),
		"indices":        types.SetNull(types.ObjectType{AttrTypes: esIndexResourceAttrTypes()}),
		"remote_indices": types.SetNull(types.ObjectType{AttrTypes: esRemoteIndexResourceAttrTypes()}),
	}
	if es.Cluster != nil && len(*es.Cluster) > 0 {
		s, d := types.SetValueFrom(ctx, types.StringType, *es.Cluster)
		diags.Append(d...)
		if diags.HasError() {
			return types.ObjectNull(elasticsearchResourceAttrTypes()), diags
		}
		attrs["cluster"] = s
	}
	if es.RunAs != nil && len(*es.RunAs) > 0 {
		s, d := types.SetValueFrom(ctx, types.StringType, *es.RunAs)
		diags.Append(d...)
		if diags.HasError() {
			return types.ObjectNull(elasticsearchResourceAttrTypes()), diags
		}
		attrs["run_as"] = s
	}
	indicesSet, d := flattenIndicesResource(ctx, es.Indices)
	diags.Append(d...)
	if diags.HasError() {
		return types.ObjectNull(elasticsearchResourceAttrTypes()), diags
	}
	attrs["indices"] = indicesSet
	remoteSet, d := flattenRemoteIndicesResource(ctx, es.RemoteIndices)
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

func flattenKibana(ctx context.Context, configs []kibanaoapi.SecurityRoleKibana) (types.Set, diag.Diagnostics) {
	var diags diag.Diagnostics
	if len(configs) == 0 {
		return types.SetValueMust(kibanaBlockObjectType(), []attr.Value{}), diags
	}
	elems := make([]attr.Value, len(configs))
	for i, cfg := range configs {
		var baseSet types.Set
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
				baseSet = s
			}
		}
		if baseSet.IsNull() {
			baseSet = types.SetNull(types.StringType)
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
