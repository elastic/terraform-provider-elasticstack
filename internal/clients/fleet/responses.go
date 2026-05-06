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
	"net/http"
	"slices"

	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// clientError converts an API transport error into diagnostics.
func clientError(err error) diag.Diagnostics {
	return diagutil.FrameworkDiagFromError(err)
}

// handleDeleteResponse handles responses from delete operations. Both 200 and
// 404 are treated as success (idempotent delete). Any other status code is
// reported as an error.
func handleDeleteResponse(statusCode int, body []byte) diag.Diagnostics {
	switch statusCode {
	case http.StatusOK, http.StatusNotFound:
		return nil
	default:
		return diagutil.ReportUnknownHTTPError(statusCode, body)
	}
}

// handleStatusResponse returns nil diagnostics when statusCode is one of the
// provided successCodes, and an error diagnostic otherwise.
func handleStatusResponse(statusCode int, body []byte, successCodes ...int) diag.Diagnostics {
	if slices.Contains(successCodes, statusCode) {
		return nil
	}
	return diagutil.ReportUnknownHTTPError(statusCode, body)
}
