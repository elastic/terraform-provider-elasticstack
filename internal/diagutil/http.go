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

package diagutil

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"slices"

	fwdiag "github.com/hashicorp/terraform-plugin-framework/diag"
)

type kibanaBoomError struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

func CheckHTTPErrorFromFW(res *http.Response, errMsg string) fwdiag.Diagnostics {
	var diags fwdiag.Diagnostics

	if res.StatusCode >= 400 {
		body, err := io.ReadAll(res.Body)
		if err != nil {
			diags.AddError(errMsg, err.Error())
			return diags
		}
		diags.AddError(errMsg, fmt.Sprintf("Failed with: %s", body))
	}
	return diags
}

func ReportUnknownHTTPError(statusCode int, body []byte) fwdiag.Diagnostics {
	return fwdiag.Diagnostics{
		fwdiag.NewErrorDiagnostic(
			fmt.Sprintf("Unexpected status code from server: got HTTP %d", statusCode),
			string(body),
		),
	}
}

// ReportKibanaBoomHTTPError attempts to parse body as a Kibana Boom error
// envelope (`{"statusCode": N, "error": "...", "message": "..."}`) by
// unmarshalling it as JSON and reading the `message` field. When unmarshal
// succeeds and `message` is a non-empty string, it returns a single error
// diagnostic using the caller-supplied summary and the extracted message as
// the detail. Otherwise (invalid JSON, missing or empty `message`, or any
// other shape) it falls back to ReportUnknownHTTPError so callers still
// receive the raw response body.
func ReportKibanaBoomHTTPError(statusCode int, summary string, body []byte) fwdiag.Diagnostics {
	var boom kibanaBoomError
	if err := json.Unmarshal(body, &boom); err == nil && boom.Message != "" {
		return fwdiag.Diagnostics{
			fwdiag.NewErrorDiagnostic(summary, boom.Message),
		}
	}
	return ReportUnknownHTTPError(statusCode, body)
}

// HandleStatusResponse returns nil when statusCode is one of successCodes, and
// an error diagnostic otherwise.
func HandleStatusResponse(statusCode int, body []byte, successCodes ...int) fwdiag.Diagnostics {
	if slices.Contains(successCodes, statusCode) {
		return nil
	}
	return ReportUnknownHTTPError(statusCode, body)
}

// UnwrapJSON200 returns val when non-nil, or an error diagnostic when val is nil.
// entityName is used in the error message (e.g. "list", "list item").
func UnwrapJSON200[T any](val *T, entityName string) (*T, fwdiag.Diagnostics) {
	if val == nil {
		return nil, fwdiag.Diagnostics{
			fwdiag.NewErrorDiagnostic(
				"Failed to parse "+entityName+" response",
				"API returned 200 but response body was nil",
			),
		}
	}
	return val, nil
}
