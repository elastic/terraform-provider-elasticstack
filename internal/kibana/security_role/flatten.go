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

func objectFromFieldSecurity(ctx context.Context, fs *map[string][]string) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics
	if fs == nil {
		return types.ObjectNull(fieldSecurityAttrTypes()), diags
	}
	var grants, excepts []string
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

func flattenIndices(ctx context.Context, indices *[]kibanaoapi.SecurityRoleESIndex) (types.Set, diag.Diagnostics) {
	var diags diag.Diagnostics
	if indices == nil || len(*indices) == 0 {
		return types.SetValueMust(types.ObjectType{AttrTypes: esIndexObjectAttrTypes()}, []attr.Value{}), diags
	}
	elems := make([]attr.Value, len(*indices))
	for i, index := range *indices {
		namesSet, d := types.SetValueFrom(ctx, types.StringType, index.Names)
		diags.Append(d...)
		if diags.HasError() {
			return types.SetNull(types.ObjectType{AttrTypes: esIndexObjectAttrTypes()}), diags
		}
		privSet, d := types.SetValueFrom(ctx, types.StringType, index.Privileges)
		diags.Append(d...)
		if diags.HasError() {
			return types.SetNull(types.ObjectType{AttrTypes: esIndexObjectAttrTypes()}), diags
		}
		queryVal := normalizedQueryFromAPI(index.Query)
		var fieldList types.List
		if index.FieldSecurity != nil {
			fieldObj, d := objectFromFieldSecurity(ctx, index.FieldSecurity)
			diags.Append(d...)
			if diags.HasError() {
				return types.SetNull(types.ObjectType{AttrTypes: esIndexObjectAttrTypes()}), diags
			}
			var d2 diag.Diagnostics
			fieldList, d2 = types.ListValue(fieldSecurityListType().ElementType(), []attr.Value{fieldObj})
			diags.Append(d2...)
			if diags.HasError() {
				return types.SetNull(types.ObjectType{AttrTypes: esIndexObjectAttrTypes()}), diags
			}
		} else {
			fieldList = types.ListValueMust(fieldSecurityListType().ElementType(), []attr.Value{})
		}
		obj, d := types.ObjectValue(esIndexObjectAttrTypes(), map[string]attr.Value{
			"names":          namesSet,
			"privileges":     privSet,
			"query":          queryVal,
			"field_security": fieldList,
		})
		diags.Append(d...)
		if diags.HasError() {
			return types.SetNull(types.ObjectType{AttrTypes: esIndexObjectAttrTypes()}), diags
		}
		elems[i] = obj
	}
	set, d := types.SetValue(types.ObjectType{AttrTypes: esIndexObjectAttrTypes()}, elems)
	diags.Append(d...)
	return set, diags
}

func flattenRemoteIndices(ctx context.Context, indices *[]kibanaoapi.SecurityRoleESRemoteIndex) (types.Set, diag.Diagnostics) {
	var diags diag.Diagnostics
	if indices == nil || len(*indices) == 0 {
		return types.SetValueMust(types.ObjectType{AttrTypes: esRemoteIndexObjectAttrTypes()}, []attr.Value{}), diags
	}
	elems := make([]attr.Value, len(*indices))
	for i, index := range *indices {
		clustersSet, d := types.SetValueFrom(ctx, types.StringType, index.Clusters)
		diags.Append(d...)
		if diags.HasError() {
			return types.SetNull(types.ObjectType{AttrTypes: esRemoteIndexObjectAttrTypes()}), diags
		}
		namesSet, d := types.SetValueFrom(ctx, types.StringType, index.Names)
		diags.Append(d...)
		if diags.HasError() {
			return types.SetNull(types.ObjectType{AttrTypes: esRemoteIndexObjectAttrTypes()}), diags
		}
		privSet, d := types.SetValueFrom(ctx, types.StringType, index.Privileges)
		diags.Append(d...)
		if diags.HasError() {
			return types.SetNull(types.ObjectType{AttrTypes: esRemoteIndexObjectAttrTypes()}), diags
		}
		queryVal := normalizedQueryFromAPI(index.Query)
		var fieldList types.List
		if index.FieldSecurity != nil {
			fieldObj, d := objectFromFieldSecurity(ctx, index.FieldSecurity)
			diags.Append(d...)
			if diags.HasError() {
				return types.SetNull(types.ObjectType{AttrTypes: esRemoteIndexObjectAttrTypes()}), diags
			}
			var d2 diag.Diagnostics
			fieldList, d2 = types.ListValue(fieldSecurityListType().ElementType(), []attr.Value{fieldObj})
			diags.Append(d2...)
			if diags.HasError() {
				return types.SetNull(types.ObjectType{AttrTypes: esRemoteIndexObjectAttrTypes()}), diags
			}
		} else {
			fieldList = types.ListValueMust(fieldSecurityListType().ElementType(), []attr.Value{})
		}
		obj, d := types.ObjectValue(esRemoteIndexObjectAttrTypes(), map[string]attr.Value{
			"clusters":       clustersSet,
			"names":          namesSet,
			"privileges":     privSet,
			"query":          queryVal,
			"field_security": fieldList,
		})
		diags.Append(d...)
		if diags.HasError() {
			return types.SetNull(types.ObjectType{AttrTypes: esRemoteIndexObjectAttrTypes()}), diags
		}
		elems[i] = obj
	}
	set, d := types.SetValue(types.ObjectType{AttrTypes: esRemoteIndexObjectAttrTypes()}, elems)
	diags.Append(d...)
	return set, diags
}

func flattenElasticsearch(ctx context.Context, es *kibanaoapi.SecurityRoleES) (types.Set, diag.Diagnostics) {
	var diags diag.Diagnostics
	attrs := map[string]attr.Value{
		"indices":        types.SetNull(types.ObjectType{AttrTypes: esIndexObjectAttrTypes()}),
		"remote_indices": types.SetNull(types.ObjectType{AttrTypes: esRemoteIndexObjectAttrTypes()}),
		"cluster":        types.SetNull(types.StringType),
		"run_as":         types.SetNull(types.StringType),
	}

	if es.Cluster != nil && len(*es.Cluster) > 0 {
		s, d := types.SetValueFrom(ctx, types.StringType, *es.Cluster)
		diags.Append(d...)
		if diags.HasError() {
			return types.SetNull(elasticsearchBlockObjectType()), diags
		}
		attrs["cluster"] = s
	}

	if es.RunAs != nil && len(*es.RunAs) > 0 {
		s, d := types.SetValueFrom(ctx, types.StringType, *es.RunAs)
		diags.Append(d...)
		if diags.HasError() {
			return types.SetNull(elasticsearchBlockObjectType()), diags
		}
		attrs["run_as"] = s
	}

	indicesSet, d := flattenIndices(ctx, es.Indices)
	diags.Append(d...)
	if diags.HasError() {
		return types.SetNull(elasticsearchBlockObjectType()), diags
	}
	attrs["indices"] = indicesSet

	remoteSet, d := flattenRemoteIndices(ctx, es.RemoteIndices)
	diags.Append(d...)
	if diags.HasError() {
		return types.SetNull(elasticsearchBlockObjectType()), diags
	}
	attrs["remote_indices"] = remoteSet

	obj, d := types.ObjectValue(elasticsearchBlockAttrTypes(), attrs)
	diags.Append(d...)
	if diags.HasError() {
		return types.SetNull(elasticsearchBlockObjectType()), diags
	}
	blockSet, d := types.SetValue(elasticsearchBlockObjectType(), []attr.Value{obj})
	diags.Append(d...)
	return blockSet, diags
}

func flattenKibanaFeatures(ctx context.Context, features *map[string][]string) (types.Set, diag.Diagnostics) {
	var diags diag.Diagnostics
	if features == nil || len(*features) == 0 {
		return types.SetValueMust(types.ObjectType{AttrTypes: kibanaFeatureAttrTypes()}, []attr.Value{}), diags
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
	if configs == nil || len(configs) == 0 {
		return types.SetValueMust(kibanaBlockObjectType(), []attr.Value{}), diags
	}
	elems := make([]attr.Value, len(configs))
	for i, cfg := range configs {
		var baseSet types.Set
		if len(cfg.Base) > 0 {
			var base []string
			if err := json.Unmarshal(cfg.Base, &base); err == nil && len(base) > 0 {
				s, d := types.SetValueFrom(ctx, types.StringType, base)
				diags.Append(d...)
				if diags.HasError() {
					return types.SetNull(kibanaBlockObjectType()), diags
				}
				baseSet = s
			}
		}
		if baseSet.IsNull() {
			baseSet = types.SetValueMust(types.StringType, []attr.Value{})
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
