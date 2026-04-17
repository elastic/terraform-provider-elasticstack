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

package kbapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// BulkDeleteSyntheticsParametersBody is the request body for BulkDeleteSyntheticsParameters.
type BulkDeleteSyntheticsParametersBody struct {
	// Ids is the list of parameter ids to delete.
	Ids []string `json:"ids"`
}

// BulkDeleteSyntheticsParametersResponse is the response from BulkDeleteSyntheticsParameters.
type BulkDeleteSyntheticsParametersResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *[]BulkDeleteSyntheticsParameterResult
}

// BulkDeleteSyntheticsParameterResult holds the per-id result of a bulk delete.
type BulkDeleteSyntheticsParameterResult struct {
	ID      string `json:"id"`
	Deleted bool   `json:"deleted"`
}

// Status returns HTTPResponse.Status.
func (r BulkDeleteSyntheticsParametersResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode.
func (r BulkDeleteSyntheticsParametersResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

// NewBulkDeleteSyntheticsParametersRequestWithBody creates a DELETE /api/synthetics/params
// request with the given content type and body reader.
//
// This endpoint (DELETE /api/synthetics/params with {"ids":[...]} body) is supported on
// Kibana >= 8.12.0. It differs from the generated DeleteParameters operation which uses
// POST /api/synthetics/params/_bulk_delete (supported only on >= 8.17.0).
func NewBulkDeleteSyntheticsParametersRequestWithBody(server string, contentType string, body io.Reader) (*http.Request, error) {
	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := "./api/synthetics/params"

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodDelete, queryURL.String(), body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", contentType)

	return req, nil
}

// NewBulkDeleteSyntheticsParametersRequest creates a DELETE /api/synthetics/params request.
func NewBulkDeleteSyntheticsParametersRequest(server string, body BulkDeleteSyntheticsParametersBody) (*http.Request, error) {
	buf, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	return NewBulkDeleteSyntheticsParametersRequestWithBody(server, "application/json", bytes.NewReader(buf))
}

// BulkDeleteSyntheticsParametersWithBody performs a DELETE /api/synthetics/params request.
func (c *Client) BulkDeleteSyntheticsParametersWithBody(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewBulkDeleteSyntheticsParametersRequestWithBody(c.Server, contentType, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

// BulkDeleteSyntheticsParameters performs a DELETE /api/synthetics/params request.
func (c *Client) BulkDeleteSyntheticsParameters(ctx context.Context, body BulkDeleteSyntheticsParametersBody, reqEditors ...RequestEditorFn) (*http.Response, error) {
	buf, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	return c.BulkDeleteSyntheticsParametersWithBody(ctx, "application/json", bytes.NewReader(buf), reqEditors...)
}

// ParseBulkDeleteSyntheticsParametersResponse parses the response from BulkDeleteSyntheticsParameters.
func ParseBulkDeleteSyntheticsParametersResponse(rsp *http.Response) (*BulkDeleteSyntheticsParametersResponse, error) {
	bodyBytes, err := io.ReadAll(rsp.Body)
	if err != nil {
		return nil, err
	}

	response := &BulkDeleteSyntheticsParametersResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	if rsp.StatusCode == http.StatusOK {
		var dest []BulkDeleteSyntheticsParameterResult
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest
	}

	return response, nil
}

// BulkDeleteSyntheticsParametersWithBodyWithResponse performs a DELETE /api/synthetics/params
// request and returns a parsed response.
func (c *ClientWithResponses) BulkDeleteSyntheticsParametersWithBodyWithResponse(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*BulkDeleteSyntheticsParametersResponse, error) {
	client, ok := c.ClientInterface.(*Client)
	if !ok {
		return nil, fmt.Errorf("ClientWithResponses.ClientInterface is not a *Client")
	}
	rsp, err := client.BulkDeleteSyntheticsParametersWithBody(ctx, contentType, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseBulkDeleteSyntheticsParametersResponse(rsp)
}

// BulkDeleteSyntheticsParametersWithResponse performs a DELETE /api/synthetics/params
// request and returns a parsed response.
func (c *ClientWithResponses) BulkDeleteSyntheticsParametersWithResponse(ctx context.Context, body BulkDeleteSyntheticsParametersBody, reqEditors ...RequestEditorFn) (*BulkDeleteSyntheticsParametersResponse, error) {
	client, ok := c.ClientInterface.(*Client)
	if !ok {
		return nil, fmt.Errorf("ClientWithResponses.ClientInterface is not a *Client")
	}
	rsp, err := client.BulkDeleteSyntheticsParameters(ctx, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseBulkDeleteSyntheticsParametersResponse(rsp)
}
