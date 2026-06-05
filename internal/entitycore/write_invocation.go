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

package entitycore

import (
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

// resourceWriteInvocation carries the framework state objects needed by both
// the Elasticsearch and Kibana write paths. All fields are concrete framework
// types, so no type parameter is required. Both runWrite (resource_envelope.go)
// and runKibanaWrite (kibana_resource_envelope.go) consume this struct.
type resourceWriteInvocation struct {
	plan         tfsdk.Plan
	priorState   *tfsdk.State
	config       tfsdk.Config
	outState     *tfsdk.State
	privateState PrivateStateStorage
	isUpdate     bool
}

// requireReadFuncDiag returns an error diagnostic when the read callback for an
// envelope is nil. component ("elasticsearch", "kibana", …) is capitalized to
// form the human-readable envelope name used in both the summary and detail.
func requireReadFuncDiag(component Component) diag.Diagnostics {
	return requireCallbackDiag(component, "read")
}

// requireDeleteFuncDiag returns an error diagnostic when the delete callback
// for an envelope is nil.
func requireDeleteFuncDiag(component Component) diag.Diagnostics {
	return requireCallbackDiag(component, "delete")
}

func requireCallbackDiag(component Component, callback string) diag.Diagnostics {
	var diags diag.Diagnostics
	name := string(component)
	if len(name) > 0 {
		name = strings.ToUpper(name[:1]) + name[1:]
	}
	diags.AddError(
		name+" envelope configuration error",
		"The "+callback+" callback passed via "+name+"ResourceOptions must not be nil.",
	)
	return diags
}
