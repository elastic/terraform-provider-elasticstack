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

package role

import (
	"context"

	estypes "github.com/elastic/go-elasticsearch/v8/typedapi/types"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types/enums/clusterprivilege"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types/enums/indexprivilege"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// marshalIndexQuery converts an Elasticsearch index query union value to a jsontypes.Normalized.
// String values are treated as pre-serialized JSON and passed through unchanged.
func marshalIndexQuery(query any) (jsontypes.Normalized, diag.Diagnostics) {
	var diags diag.Diagnostics
	if q, ok := query.(string); ok {
		return jsontypes.NewNormalizedValue(q), diags
	}
	return typeutils.MarshalToNormalized(query, path.Root("query"), &diags), diags
}

// clusterPrivilegesToStrings converts a list of cluster privilege enums to their string names.
func clusterPrivilegesToStrings(cluster []clusterprivilege.ClusterPrivilege) []string {
	clusterStrings := make([]string, len(cluster))
	for i, cp := range cluster {
		clusterStrings[i] = cp.String()
	}
	return clusterStrings
}

// indexPrivilegesToStrings converts a list of index privilege enums to their string names.
func indexPrivilegesToStrings(privileges []indexprivilege.IndexPrivilege) []string {
	privilegeStrings := make([]string, len(privileges))
	for i, p := range privileges {
		privilegeStrings[i] = p.String()
	}
	return privilegeStrings
}

// fieldSecurityGrantExceptSets converts an API FieldSecurity's Grant/Except slices to
// Terraform sets, shared by the resource (which wraps them in a single object) and the
// data source (which wraps them in a 0-or-1-element list) since the two schemas represent
// field_security with different container types.
func fieldSecurityGrantExceptSets(ctx context.Context, fs *estypes.FieldSecurity) (grant types.Set, except types.Set, diags diag.Diagnostics) {
	if fs == nil {
		return types.SetNull(types.StringType), types.SetNull(types.StringType), diags
	}

	grant, d := types.SetValueFrom(ctx, types.StringType, typeutils.NonNilSlice(fs.Grant))
	diags.Append(d...)
	if diags.HasError() {
		return grant, except, diags
	}

	except, d = types.SetValueFrom(ctx, types.StringType, typeutils.NonNilSlice(fs.Except))
	diags.Append(d...)
	return grant, except, diags
}

// applicationsToSet converts a role's application privilege entries to a Terraform set,
// shared by the resource and data source since both represent applications identically.
func applicationsToSet(ctx context.Context, applications []estypes.ApplicationPrivileges) (types.Set, diag.Diagnostics) {
	var diags diag.Diagnostics
	attrTypes := getApplicationAttrTypes()

	if len(applications) == 0 {
		return types.SetNull(types.ObjectType{AttrTypes: attrTypes}), diags
	}

	appElements := make([]attr.Value, len(applications))
	for i, app := range applications {
		privSet, d := types.SetValueFrom(ctx, types.StringType, typeutils.NonNilSlice(app.Privileges))
		diags.Append(d...)
		if diags.HasError() {
			return types.SetNull(types.ObjectType{AttrTypes: attrTypes}), diags
		}

		resSet, d := types.SetValueFrom(ctx, types.StringType, typeutils.NonNilSlice(app.Resources))
		diags.Append(d...)
		if diags.HasError() {
			return types.SetNull(types.ObjectType{AttrTypes: attrTypes}), diags
		}

		appObj, d := types.ObjectValue(attrTypes, map[string]attr.Value{
			attrApplication: types.StringValue(app.Application),
			attrPrivileges:  privSet,
			attrResources:   resSet,
		})
		diags.Append(d...)
		if diags.HasError() {
			return types.SetNull(types.ObjectType{AttrTypes: attrTypes}), diags
		}

		appElements[i] = appObj
	}

	appSet, d := types.SetValue(types.ObjectType{AttrTypes: attrTypes}, appElements)
	diags.Append(d...)
	return appSet, diags
}
