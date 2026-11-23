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

package ilm

import (
	"context"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// phaseMapToObjectValue decodes a phase's dynamic JSON action data into the
// Terraform Object the phase's schema expects, delegating the recursive
// JSON-to-attr.Value decoding to typeutils.DecodeAnyToAttrValue.
func phaseMapToObjectValue(ctx context.Context, phaseName string, data map[string]any) (types.Object, diag.Diagnostics) {
	ot := phaseObjectType(phaseName)
	v, diags := typeutils.DecodeAnyToAttrValue(ctx, ot, data)
	if diags.HasError() {
		return types.ObjectUnknown(ot.AttrTypes), diags
	}
	obj, ok := v.(types.Object)
	if !ok {
		diags.AddError("Internal error", fmt.Sprintf("expected types.Object, got %T", v))
		return types.ObjectUnknown(ot.AttrTypes), diags
	}
	return obj, diags
}
