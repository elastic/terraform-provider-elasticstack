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

package agentbuilder

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// PopulateSet sets dst to a types.Set containing the strings in src.
// If src is empty, dst is set to a null set.
func PopulateSet(ctx context.Context, src []string, dst *types.Set) diag.Diagnostics {
	if len(src) > 0 {
		v, d := types.SetValueFrom(ctx, types.StringType, src)
		*dst = v
		return d
	}
	*dst = types.SetNull(types.StringType)
	return nil
}

// SetToStrings converts a types.Set to a []string.
// Returns nil when the set is null or unknown.
func SetToStrings(ctx context.Context, set types.Set) ([]string, diag.Diagnostics) {
	if set.IsNull() || set.IsUnknown() {
		return nil, nil
	}
	var out []string
	d := set.ElementsAs(ctx, &out, false)
	return out, d
}
