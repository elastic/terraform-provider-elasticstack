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
	"net/http"
	"slices"

	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

func handleGetResponse[T any](statusCode int, body []byte) (*T, diag.Diagnostics) {
	switch statusCode {
	case http.StatusOK:
		var result T
		if err := json.Unmarshal(body, &result); err != nil {
			return nil, diagutil.FrameworkDiagFromError(err)
		}
		return &result, nil
	case http.StatusNotFound:
		return nil, nil
	default:
		return nil, diagutil.ReportUnknownHTTPError(statusCode, body)
	}
}

func handleMutateResponse[T any](statusCode int, body []byte) (*T, diag.Diagnostics) {
	switch statusCode {
	case http.StatusOK:
		var result T
		if err := json.Unmarshal(body, &result); err != nil {
			return nil, diagutil.FrameworkDiagFromError(err)
		}
		return &result, nil
	default:
		return nil, diagutil.ReportUnknownHTTPError(statusCode, body)
	}
}

func handleStatusResponse(statusCode int, body []byte, successStatusCodes ...int) diag.Diagnostics {
	if slices.Contains(successStatusCodes, statusCode) {
		return nil
	}

	return diagutil.ReportUnknownHTTPError(statusCode, body)
}

func clientError(err error) diag.Diagnostics {
	return diagutil.FrameworkDiagFromError(err)
}
