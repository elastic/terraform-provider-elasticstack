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
	"fmt"
	"io"
	"net/http"
	"slices"

	"github.com/elastic/go-elasticsearch/v8/esapi"
	fwdiag "github.com/hashicorp/terraform-plugin-framework/diag"
)

func CheckErrorFromFW(res *esapi.Response, errMsg string) fwdiag.Diagnostics {
	var diags fwdiag.Diagnostics

	if res.IsError() {
		body, err := io.ReadAll(res.Body)
		if err != nil {
			diags.AddError(errMsg, err.Error())
			return diags
		}
		diags.AddError(errMsg, fmt.Sprintf("Failed with: %s", body))
		return diags
	}
	return diags
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

// HandleStatusResponse returns nil when statusCode is one of successCodes, and
// an error diagnostic otherwise.
func HandleStatusResponse(statusCode int, body []byte, successCodes ...int) fwdiag.Diagnostics {
	if slices.Contains(successCodes, statusCode) {
		return nil
	}
	return ReportUnknownHTTPError(statusCode, body)
}
