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

package elasticdefendintegrationpolicy

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Helper to extract sub-map from a map.
func getMap(m map[string]any, key string) map[string]any {
	if m == nil {
		return nil
	}
	if v, ok := m[key]; ok {
		if sm, ok := v.(map[string]any); ok {
			return sm
		}
	}
	return nil
}

// mapOptionalObject maps a sub-section from data using getMap; returns
// ObjectValueFrom when present, ObjectNull otherwise.
func mapOptionalObject[M any](ctx context.Context, data map[string]any, key string, attrTypes map[string]attr.Type, build func(map[string]any) M) (types.Object, diag.Diagnostics) {
	sub := getMap(data, key)
	if sub == nil {
		return types.ObjectNull(attrTypes), nil
	}
	return types.ObjectValueFrom(ctx, attrTypes, build(sub))
}
