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

// Package globaldatatags provides the shared Terraform model and expand/flatten
// helpers for the `global_data_tags` map attribute used by both
// internal/fleet/agentpolicy and internal/fleet/managedintegration. Both
// resources expose a map keyed by tag name whose value is exactly one of
// string_value/number_value, but each is backed by its own generated kbapi
// union type (Kibana emits a distinct anonymous union per request/response
// shape even though the wire shape is identical) -- so the conversion glue
// between Item and a given kbapi union type still lives in the caller;
// this package only removes the duplicated model shape and union-decoding
// control flow.
package globaldatatags

import (
	"errors"

	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"

	"github.com/hashicorp/terraform-plugin-framework-validators/float32validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const (
	StringValueAttr = "string_value"
	NumberValueAttr = "number_value"

	errExactlyOneValue = "Each entry in global_data_tags must have exactly one of string_value or number_value set."
)

// Item is the Terraform model for a single `global_data_tags` map entry.
type Item struct {
	StringValue types.String  `tfsdk:"string_value"`
	NumberValue types.Float32 `tfsdk:"number_value"`
}

func AttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		StringValueAttr: types.StringType,
		NumberValueAttr: types.Float32Type,
	}
}

func ElementType() attr.Type {
	return types.ObjectType{AttrTypes: AttrTypes()}
}

// Expand converts a decoded map entry into the caller's API union type T,
// using fromString/fromNumber to build T from whichever of string_value or
// number_value is set. If neither (or the conversion itself fails), it
// records a diagnostic on meta.Diags at meta.Path and returns the zero value
// of T; callers that check diags.HasError() before using their results (as
// every current caller does) can treat the returned value as unusable.
func Expand[T any](item Item, meta typeutils.MapMeta, fromString func(string) (T, error), fromNumber func(float32) (T, error)) T {
	var zero, value T
	var err error

	switch {
	case typeutils.IsKnown(item.StringValue):
		value, err = fromString(item.StringValue.ValueString())
	case typeutils.IsKnown(item.NumberValue):
		value, err = fromNumber(item.NumberValue.ValueFloat32())
	default:
		meta.Diags.AddAttributeError(meta.Path, "Invalid global_data_tags entry", errExactlyOneValue)
		return zero
	}

	if err != nil {
		meta.Diags.AddAttributeError(meta.Path, "global_data_tags validation_error_converting_values", err.Error())
		return zero
	}
	return value
}

// NestedObject builds the schema.NestedAttributeObject shared by every
// `global_data_tags` map attribute: the string_value/number_value attributes
// wired with the ConflictsWith + AtLeastOneOf validator pairs that enforce
// "exactly one of the two must be set". stringDescription/numberDescription
// populate each attribute's Description field, or MarkdownDescription when
// markdown is true, so callers can keep their own schema's description style
// (plain vs. Markdown) without duplicating the validator wiring.
func NestedObject(stringDescription, numberDescription string, markdown bool) schema.NestedAttributeObject {
	stringAttr := schema.StringAttribute{
		Optional: true,
		Validators: []validator.String{
			stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName(NumberValueAttr)),
			stringvalidator.AtLeastOneOf(
				path.MatchRelative().AtParent().AtName(StringValueAttr),
				path.MatchRelative().AtParent().AtName(NumberValueAttr),
			),
		},
	}
	numberAttr := schema.Float32Attribute{
		Optional: true,
		Validators: []validator.Float32{
			float32validator.ConflictsWith(path.MatchRelative().AtParent().AtName(StringValueAttr)),
			float32validator.AtLeastOneOf(
				path.MatchRelative().AtParent().AtName(StringValueAttr),
				path.MatchRelative().AtParent().AtName(NumberValueAttr),
			),
		},
	}

	if markdown {
		stringAttr.MarkdownDescription = stringDescription
		numberAttr.MarkdownDescription = numberDescription
	} else {
		stringAttr.Description = stringDescription
		numberAttr.Description = numberDescription
	}

	return schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			StringValueAttr: stringAttr,
			NumberValueAttr: numberAttr,
		},
	}
}

// Flatten decodes a caller's API union value V into Item, preferring the
// number-valued variant to match the union's declaration order (matching the
// AsX...Value1-then-Value0 fallback both packages already use). It returns an
// error if value is neither a string nor a number.
func Flatten[V any](value V, asNumber func(V) (float32, error), asString func(V) (string, error)) (Item, error) {
	if num, err := asNumber(value); err == nil {
		return Item{NumberValue: types.Float32Value(num)}, nil
	}
	if str, err := asString(value); err == nil {
		return Item{StringValue: types.StringValue(str)}, nil
	}
	return Item{}, errors.New("value is neither string_value nor number_value")
}
