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

package ml

import (
	"fmt"

	fwdiags "github.com/hashicorp/terraform-plugin-framework/diag"
)

// RequireNonEmptyID returns diagnostics with a single error when id is
// empty, using fieldName (e.g. "calendar_id") in the message. Callers should
// return their own zero-value result together with these diagnostics when
// diags.HasError() is true.
func RequireNonEmptyID(id, fieldName string) fwdiags.Diagnostics {
	var diags fwdiags.Diagnostics
	if id == "" {
		diags.AddError("Invalid resource ID", fmt.Sprintf("%s cannot be empty", fieldName))
	}
	return diags
}
