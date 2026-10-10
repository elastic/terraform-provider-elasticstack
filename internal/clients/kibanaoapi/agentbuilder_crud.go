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

package kibanaoapi

import (
	"net/http"

	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// crudCall invokes a single generated-SDK method and returns its HTTP status
// code and raw response body. It returns a non-nil err only when the call
// itself failed (e.g. request construction or transport error), not for
// non-2xx HTTP responses.
type crudCall func() (statusCode int, body []byte, err error)

// crudGet runs call and unmarshals a 200 response into T, matching the
// error-handling and dispatch shape shared by the agentbuilder_* Get wrappers.
// Returns (nil, nil) on 404.
func crudGet[T any](call crudCall) (*T, diag.Diagnostics) {
	statusCode, body, err := call()
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}
	return HandleGetRawResponse[T](statusCode, body)
}

// crudMutate runs call and unmarshals a 200 response into T, matching the
// error-handling and dispatch shape shared by the agentbuilder_* Create/Update
// wrappers.
func crudMutate[T any](call crudCall) (*T, diag.Diagnostics) {
	statusCode, body, err := call()
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}
	return HandleMutateRawResponse[T](statusCode, body)
}

// crudDelete runs call and reports a diagnostic unless the response is 200 or
// 404, matching the error-handling and dispatch shape shared by the
// agentbuilder_* Delete wrappers.
func crudDelete(call crudCall) diag.Diagnostics {
	statusCode, body, err := call()
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}
	return diagutil.HandleStatusResponse(statusCode, body, http.StatusOK, http.StatusNotFound)
}
