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

package securitydetectionrule

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// convertListFieldToAPI converts a Terraform list attribute to an API slice using mapItem to
// convert each element. Returns nil when the list is unknown, null, or has no elements.
func convertListFieldToAPI[TModel, TAPI any](
	ctx context.Context,
	list types.List,
	p path.Path,
	diags *diag.Diagnostics,
	mapItem func(TModel, typeutils.ListMeta) TAPI,
) []TAPI {
	if !typeutils.IsKnown(list) || len(list.Elements()) == 0 {
		return nil
	}
	return typeutils.ListTypeToSlice(ctx, list, p, diags, mapItem)
}

// convertAPISliceToListField converts an API slice to a Terraform list attribute using mapItem to
// convert each element. Returns a null list of elementType when apiSlice is empty.
func convertAPISliceToListField[TAPI, TModel any](
	ctx context.Context,
	apiSlice []TAPI,
	elementType attr.Type,
	mapItem func(TAPI) TModel,
) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics

	if len(apiSlice) == 0 {
		return types.ListNull(elementType), diags
	}

	items := make([]TModel, 0, len(apiSlice))
	for _, apiItem := range apiSlice {
		items = append(items, mapItem(apiItem))
	}

	listValue, listDiags := types.ListValueFrom(ctx, elementType, items)
	diags.Append(listDiags...)
	return listValue, diags
}
