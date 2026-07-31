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

package fleet

import (
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

// VersionGateError returns a single attribute-scoped error diagnostic
// reporting that the value at attr is not supported by the connected Elastic
// Stack version. detail should already describe the gated feature and the
// minimum required version, e.g. "Space IDs are only supported in Elastic
// Stack 9.1.0 and above".
func VersionGateError(attr path.Path, detail string) diag.Diagnostics {
	return diag.Diagnostics{
		diag.NewAttributeErrorDiagnostic(attr, "Unsupported Elasticsearch version", detail),
	}
}
