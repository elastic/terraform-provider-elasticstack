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
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

const alertingV2DisabledCode = "ALERTING_DISABLED"

// alertingV2DisabledDiagnostic is the single owner of the diagnostic returned
// whenever an Alerting v2 route answers HTTP 503 with code ALERTING_DISABLED.
// serverMessage is the message from the error body and is appended so the
// server's wording is not lost.
func alertingV2DisabledDiagnostic(serverMessage string) diag.Diagnostic {
	return diag.NewErrorDiagnostic(
		"Alerting v2 is disabled",
		"Kibana rejected the request because the Alerting v2 engine is disabled in the target space. "+
			"Enable the `alerting:v2:enabled` advanced setting in Kibana and retry. Server message: "+serverMessage,
	)
}

// alertingV2ErrorDiagnostics maps a non-success Alerting v2 response to
// diagnostics. Bodies that decode as KibanaHTTPAPIsAlertingErrorResponse
// surface their stable code and message; a 503 ALERTING_DISABLED becomes
// alertingV2DisabledDiagnostic. Anything else falls back to the raw body.
func alertingV2ErrorDiagnostics(statusCode int, body []byte) diag.Diagnostics {
	var apiErr kbapi.KibanaHTTPAPIsAlertingErrorResponse
	if err := json.Unmarshal(body, &apiErr); err != nil || apiErr.Code == "" {
		return diagutil.ReportUnknownHTTPError(statusCode, body)
	}

	if statusCode == http.StatusServiceUnavailable && apiErr.Code == alertingV2DisabledCode {
		return diag.Diagnostics{alertingV2DisabledDiagnostic(apiErr.Message)}
	}

	return diag.Diagnostics{diag.NewErrorDiagnostic(
		fmt.Sprintf("Alerting v2 API error: HTTP %d %s", statusCode, apiErr.Code),
		apiErr.Message,
	)}
}
