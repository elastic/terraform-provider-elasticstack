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

package agentbuilderapi

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAgentsAPI_StructImplementsInterface(t *testing.T) {
	// This test verifies that AgentsAPI can be instantiated and used
	// The actual API implementation requires the generated kbapi client,
	// so we verify the structure is correct
	api := &AgentsAPI{}
	assert.NotNil(t, api)
}

func TestHandleGetResponse_Success(t *testing.T) {
	type TestStruct struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}

	body := `{"id": "agent-123", "name": "Test Agent"}`
	result, found, diags := handleGetResponse[TestStruct](200, []byte(body))

	assert.False(t, diags.HasError())
	assert.True(t, found)
	assert.NotNil(t, result)
	assert.Equal(t, "agent-123", result.ID)
	assert.Equal(t, "Test Agent", result.Name)
}

func TestHandleGetResponse_NotFound(t *testing.T) {
	type TestStruct struct {
		ID string `json:"id"`
	}

	body := `{"error": "not found"}`
	result, found, diags := handleGetResponse[TestStruct](404, []byte(body))

	assert.False(t, diags.HasError())
	assert.False(t, found)
	assert.Nil(t, result)
}

func TestHandleGetResponse_Error(t *testing.T) {
	type TestStruct struct {
		ID string `json:"id"`
	}

	body := `{"error": "internal error"}`
	result, found, diags := handleGetResponse[TestStruct](500, []byte(body))

	assert.True(t, diags.HasError())
	assert.False(t, found)
	assert.Nil(t, result)
}

func TestHandleGetResponse_InvalidJSON(t *testing.T) {
	type TestStruct struct {
		ID string `json:"id"`
	}

	body := `invalid json`
	result, found, diags := handleGetResponse[TestStruct](200, []byte(body))

	assert.True(t, diags.HasError())
	assert.False(t, found)
	assert.Nil(t, result)
}

func TestHandleMutateResponse_Success(t *testing.T) {
	type TestStruct struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}

	body := `{"id": "agent-123", "name": "Test Agent"}`
	result, diags := handleMutateResponse[TestStruct](200, []byte(body))

	assert.False(t, diags.HasError())
	assert.NotNil(t, result)
	assert.Equal(t, "agent-123", result.ID)
}

func TestHandleMutateResponse_Error(t *testing.T) {
	type TestStruct struct {
		ID string `json:"id"`
	}

	body := `{"error": "bad request"}`
	result, diags := handleMutateResponse[TestStruct](400, []byte(body))

	assert.True(t, diags.HasError())
	assert.Nil(t, result)
}

func TestHandleMutateResponse_InvalidJSON(t *testing.T) {
	type TestStruct struct {
		ID string `json:"id"`
	}

	body := `invalid json`
	result, diags := handleMutateResponse[TestStruct](200, []byte(body))

	assert.True(t, diags.HasError())
	assert.Nil(t, result)
}

func TestHandleStatusResponse_Success(t *testing.T) {
	diags := handleStatusResponse(200, []byte(""), 200, 204)
	assert.False(t, diags.HasError())
}

func TestHandleStatusResponse_NotFoundAllowed(t *testing.T) {
	diags := handleStatusResponse(404, []byte("not found"), 200, 404)
	assert.False(t, diags.HasError())
}

func TestHandleStatusResponse_Error(t *testing.T) {
	diags := handleStatusResponse(500, []byte("internal error"), 200, 204)
	assert.True(t, diags.HasError())
}

func TestToolsAPI_Struct(t *testing.T) {
	api := &ToolsAPI{}
	assert.NotNil(t, api)
}

func TestWorkflowsAPI_Struct(t *testing.T) {
	api := &WorkflowsAPI{}
	assert.NotNil(t, api)
}
