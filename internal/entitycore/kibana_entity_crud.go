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

import "github.com/hashicorp/terraform-plugin-framework/diag"

// WriteEntity implements the build→call→populate→wrap sequence shared by Kibana
// Create/Update callbacks: build the API request body, invoke the API call, populate
// the model from the response, and wrap the result in a [KibanaWriteResult]. Diagnostics
// are appended and checked after each step, bailing out on the first error.
//
// buildBody, call, and populate close over whatever the caller needs (context, client,
// space/write identifiers) that isn't uniform across entities. populate must return the
// model reflecting the API response — callers whose model's populateFromAPI method uses a
// pointer receiver should return the same (now-mutated) model passed into the closure.
func WriteEntity[T KibanaResourceModel, Body any, Resp any](
	buildBody func() (Body, diag.Diagnostics),
	call func(Body) (Resp, diag.Diagnostics),
	populate func(Resp) (T, diag.Diagnostics),
) (KibanaWriteResult[T], diag.Diagnostics) {
	var diags diag.Diagnostics

	body, bodyDiags := buildBody()
	diags.Append(bodyDiags...)
	if diags.HasError() {
		return KibanaWriteResult[T]{}, diags
	}

	resp, callDiags := call(body)
	diags.Append(callDiags...)
	if diags.HasError() {
		return KibanaWriteResult[T]{}, diags
	}

	model, populateDiags := populate(resp)
	diags.Append(populateDiags...)
	if diags.HasError() {
		return KibanaWriteResult[T]{}, diags
	}

	return KibanaWriteResult[T]{Model: model}, diags
}

// ReadEntity implements the call→check→populate sequence shared by Kibana Read callbacks:
// invoke the API get call, treat a not-found response as a miss, and otherwise populate the
// model from the response. Diagnostics are appended and checked after each step.
//
// call, isNotFound, and populate close over whatever the caller needs (context, client,
// resource/space identifiers). populate must return the model reflecting the API response,
// matching the same pointer-receiver caveat as [WriteEntity].
func ReadEntity[T any, Resp any](
	model T,
	call func() (Resp, diag.Diagnostics),
	isNotFound func(Resp) bool,
	populate func(Resp) (T, diag.Diagnostics),
) (T, bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	resp, callDiags := call()
	diags.Append(callDiags...)
	if diags.HasError() {
		return model, false, diags
	}

	if isNotFound(resp) {
		return model, false, diags
	}

	updated, populateDiags := populate(resp)
	diags.Append(populateDiags...)
	if diags.HasError() {
		return model, false, diags
	}

	return updated, true, diags
}
