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
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// AssertSameType type-asserts other to the concrete type of expected (T), for
// use at the top of a XSemanticEquals method (StringSemanticEquals,
// ListSemanticEquals, MapSemanticEquals, ObjectSemanticEquals,
// SetSemanticEquals, Float64SemanticEquals, ...). The terraform-plugin-framework
// contract for these methods guarantees other is normally the same concrete
// type as the receiver; this centralizes the "unexpected value type" diagnostic
// raised on the rare occasion that guarantee doesn't hold.
func AssertSameType[T attr.Value](expected T, other attr.Value) (T, bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	got, ok := other.(T)
	if !ok {
		diags.AddError(
			"Semantic Equality Check Error",
			"An unexpected value type was received while performing semantic equality checks. "+
				"Please report this to the provider developers.\n\n"+
				"Expected Value Type: "+fmt.Sprintf("%T", expected)+"\n"+
				"Got Value Type: "+fmt.Sprintf("%T", other),
		)
		return got, false, diags
	}

	return got, true, diags
}

// UnmarshalJSONForSemanticEquals unmarshals raw JSON into a value of type T,
// centralizing the "Semantic Equality Check Error" diagnostic raised on
// failure across XSemanticEquals implementations that decode JSON-typed
// values before comparing them structurally.
func UnmarshalJSONForSemanticEquals[T any](raw string) (T, diag.Diagnostics) {
	var diags diag.Diagnostics

	var out T
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		diags.AddError("Semantic Equality Check Error", err.Error())
	}

	return out, diags
}
