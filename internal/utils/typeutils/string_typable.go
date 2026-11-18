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

package typeutils

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// StringTypableValueFromTerraform implements the ValueFromTerraform adapter shared by every
// custom basetypes.StringTypable type in this codebase: it converts the raw tftypes.Value into
// a basetypes.StringValue via stringType, then delegates to valueFromString (normally the
// receiver's own ValueFromString method) to produce the final attr.Value.
func StringTypableValueFromTerraform(
	ctx context.Context,
	stringType basetypes.StringType,
	valueFromString func(context.Context, basetypes.StringValue) (basetypes.StringValuable, diag.Diagnostics),
	in tftypes.Value,
) (attr.Value, error) {
	attrValue, err := stringType.ValueFromTerraform(ctx, in)
	if err != nil {
		return nil, err
	}

	stringValue, ok := attrValue.(basetypes.StringValue)
	if !ok {
		return nil, fmt.Errorf("unexpected value type of %T", attrValue)
	}

	stringValuable, diags := valueFromString(ctx, stringValue)
	if diags.HasError() {
		return nil, fmt.Errorf("unexpected error converting StringValue to StringValuable: %v", diags)
	}

	return stringValuable, nil
}

// StringTypableEqual implements the Equal method shared by every custom basetypes.StringTypable
// type in this package that embeds basetypes.StringType: since basetypes.StringType carries no
// distinguishing state, equality between two instances of T reduces to asserting o to that same
// concrete type.
func StringTypableEqual[T attr.Type](o attr.Type) bool {
	_, ok := o.(T)
	return ok
}

// StringValuableEqual implements the Equal method shared by every custom basetypes.StringValuable
// type in this package that embeds basetypes.StringValue: it asserts o to the same concrete type T,
// then delegates to the embedded StringValue's own Equal. embeddedOf extracts the embedded
// basetypes.StringValue from a T (normally a field access, e.g. func(d Duration) basetypes.StringValue
// { return d.StringValue }).
func StringValuableEqual[T attr.Value](self basetypes.StringValue, o attr.Value, embeddedOf func(T) basetypes.StringValue) bool {
	other, ok := o.(T)
	if !ok {
		return false
	}

	return self.Equal(embeddedOf(other))
}
