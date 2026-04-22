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
	"bytes"
	"context"
	"fmt"
	"mime/multipart"
	"net/http"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// ImportSavedObjectsResult holds the parsed result of a saved objects import.
type ImportSavedObjectsResult struct {
	Success        bool
	SuccessCount   int64
	Errors         []map[string]any
	SuccessResults []map[string]any
}

// ImportSavedObjects calls the Kibana Saved Objects Import API with the provided
// NDJSON file contents and parameters. It builds a multipart/form-data body with
// the file contents as a single part named "file" with filename "export.ndjson".
func ImportSavedObjects(ctx context.Context, client *Client, spaceID string, fileContents []byte, params kbapi.PostSavedObjectsImportParams) (*ImportSavedObjectsResult, diag.Diagnostics) {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	part, err := writer.CreateFormFile("file", "export.ndjson")
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(fmt.Errorf("failed to create multipart form file: %w", err))
	}

	if _, err := part.Write(fileContents); err != nil {
		return nil, diagutil.FrameworkDiagFromError(fmt.Errorf("failed to write file contents to multipart form: %w", err))
	}

	if err := writer.Close(); err != nil {
		return nil, diagutil.FrameworkDiagFromError(fmt.Errorf("failed to close multipart writer: %w", err))
	}

	contentType := writer.FormDataContentType()

	resp, err := client.API.PostSavedObjectsImportWithBodyWithResponse(ctx, &params, contentType, &buf, SpaceAwarePathRequestEditor(spaceID))
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		if resp.JSON200 == nil {
			return nil, diag.Diagnostics{
				diag.NewErrorDiagnostic("Failed to parse import response", "API returned 200 but JSON200 is nil"),
			}
		}

		result := &ImportSavedObjectsResult{
			Success:        resp.JSON200.Success,
			SuccessCount:   int64(resp.JSON200.SuccessCount),
			Errors:         resp.JSON200.Errors,
			SuccessResults: resp.JSON200.SuccessResults,
		}
		return result, nil

	case http.StatusBadRequest:
		if resp.JSON400 != nil {
			return nil, diag.Diagnostics{
				diag.NewErrorDiagnostic(
					fmt.Sprintf("Bad request (HTTP 400): %s", resp.JSON400.Error),
					resp.JSON400.Message,
				),
			}
		}
		return nil, reportUnknownError(resp.StatusCode(), resp.Body)

	default:
		return nil, reportUnknownError(resp.StatusCode(), resp.Body)
	}
}
