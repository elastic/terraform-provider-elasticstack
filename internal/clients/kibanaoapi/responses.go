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

// Package kibanaoapi provides HTTP response helpers for the Kibana API client.
//
// There are two handler families:
//   - Raw handlers (HandleGetRawResponse, HandleMutateRawResponse): use json.Unmarshal
//     directly on the response body. Use these when the kbapi generated struct cannot
//     unmarshal the API response correctly, or when you need a custom struct type.
//   - Typed handlers (HandleGetTypedResponse, HandleMutateTypedResponse): use a callback
//     to extract a pre-parsed struct pointer from the kbapi response (e.g. resp.JSON200).
//     Use these when the kbapi generated types correctly represent the API response.
package kibanaoapi

import (
	"encoding/json"
	"net/http"
	"slices"

	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// HandleGetRawResponse handles a read response by unmarshaling the body into T.
// Use this when the kbapi generated struct cannot unmarshal the API response correctly.
// Returns (nil, nil) on HTTP 404.
func HandleGetRawResponse[T any](statusCode int, body []byte) (*T, diag.Diagnostics) {
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

// HandleMutateRawResponse handles a create/update response by unmarshaling the body into T.
// Use this when the kbapi generated struct cannot unmarshal the API response correctly.
func HandleMutateRawResponse[T any](statusCode int, body []byte) (*T, diag.Diagnostics) {
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

// HandleGetTypedResponse handles a read response for kbapi typed-struct responses.
// The extract callback is called only on HTTP 200 and should return the pre-parsed
// struct pointer from the response (e.g. resp.JSON200). Returns (nil, nil) on 404.
func HandleGetTypedResponse[T any](statusCode int, body []byte, extract func() *T) (*T, diag.Diagnostics) {
	switch statusCode {
	case http.StatusOK:
		result := extract()
		if result == nil {
			return nil, diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Failed to parse response",
					"API returned success status but response body was nil or not JSON",
				),
			}
		}
		return result, nil
	case http.StatusNotFound:
		return nil, nil
	default:
		return nil, diagutil.ReportUnknownHTTPError(statusCode, body)
	}
}

// HandleMutateTypedResponse handles a create/update response for kbapi typed-struct responses.
// The extract callback is called only on success status and should return the pre-parsed struct pointer.
// Defaults to HTTP 200 when successCodes is omitted.
func HandleMutateTypedResponse[T any](statusCode int, body []byte, extract func() *T, successCodes ...int) (*T, diag.Diagnostics) {
	if len(successCodes) == 0 {
		successCodes = []int{http.StatusOK}
	}

	if slices.Contains(successCodes, statusCode) {
		result := extract()
		if result == nil {
			return nil, diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Failed to parse response",
					"API returned success status but response body was nil or not JSON",
				),
			}
		}
		return result, nil
	}

	return nil, diagutil.ReportUnknownHTTPError(statusCode, body)
}
