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

package policyshape

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// DecodedInput is the once-decoded form of a single `inputs` map element, with
// its nested `streams` map (if any) already decoded. DecodeInputs produces
// this so callers don't each independently re-run the reflection-based
// typeutils.MapTypeAs decode over the same inputs+streams structure.
type DecodedInput[InputM any] struct {
	Model   InputM
	Streams map[string]InputStreamModel // nil if streams is null/unknown
}

// streamsGetter is satisfied by any input model that exposes its Streams field
// as a types.Map — both policyshape.InputModel and managedIntegrationInputModel
// meet this constraint through a simple struct-field access.
type streamsGetter interface {
	GetStreams() types.Map
}

// DecodeInputs decodes the `inputs` attribute of an InputsValue, and each
// input's nested `streams` map, exactly once. Returns nil if `inputs` itself
// is null/unknown or fails to decode.
//
// InputM must be a struct decodable from a single map element of inputs (e.g.
// policyshape.InputModel or managedIntegrationInputModel) and must implement
// streamsGetter so DecodeInputs can reach the nested streams map without
// knowing the concrete field name.
func DecodeInputs[InputM streamsGetter](ctx context.Context, inputs InputsValue, attrPath path.Path, diags *diag.Diagnostics) map[string]DecodedInput[InputM] {
	if !typeutils.IsKnown(inputs.MapValue) {
		return nil
	}

	inputsMap := typeutils.MapTypeAs[InputM](ctx, inputs.MapValue, attrPath, diags)
	if inputsMap == nil {
		return nil
	}

	decoded := make(map[string]DecodedInput[InputM], len(inputsMap))
	for inputID, inputModel := range inputsMap {
		d := DecodedInput[InputM]{Model: inputModel}
		if typeutils.IsKnown(inputModel.GetStreams()) {
			streamPath := attrPath.AtMapKey(inputID).AtName("streams")
			d.Streams = typeutils.MapTypeAs[InputStreamModel](ctx, inputModel.GetStreams(), streamPath, diags)
		}
		decoded[inputID] = d
	}
	return decoded
}
