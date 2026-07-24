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

package parameter

import (
	"net/http"
	"testing"

	kboapi "github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/stretchr/testify/require"
)

func TestParseCreateParameterResponse_nonOKStatus(t *testing.T) {
	t.Parallel()

	resp := &kboapi.PostParametersResponse{
		Body:         []byte(`{"statusCode":409,"error":"Conflict","message":"Parameter already exists"}`),
		HTTPResponse: &http.Response{StatusCode: http.StatusConflict, Status: "409 Conflict"},
	}

	_, diags := parseCreateParameterResponse(resp, "my-key")
	require.True(t, diags.HasError())
	require.Contains(t, diags.Errors()[0].Summary(), "409")
}

func TestParseCreateParameterResponse_successWithoutJSON200(t *testing.T) {
	t.Parallel()

	resp := &kboapi.PostParametersResponse{
		Body:         []byte(`ok`),
		HTTPResponse: &http.Response{StatusCode: http.StatusOK, Status: "200 OK"},
	}

	_, diags := parseCreateParameterResponse(resp, "my-key")
	require.True(t, diags.HasError())
	require.Contains(t, diags.Errors()[0].Summary(), "Failed to parse response")
}

func TestParseCreateParameterResponse_successWithID(t *testing.T) {
	t.Parallel()

	union := kboapi.CreateParamResponse{}
	require.NoError(t, union.FromSyntheticsPostParameterResponse(kboapi.SyntheticsPostParameterResponse{
		Id: new("param-uuid"),
	}))

	resp := &kboapi.PostParametersResponse{
		Body:         []byte(`{"id":"param-uuid"}`),
		HTTPResponse: &http.Response{StatusCode: http.StatusOK, Status: "200 OK"},
		JSON200:      &union,
	}

	got, diags := parseCreateParameterResponse(resp, "my-key")
	require.False(t, diags.HasError())
	require.NotNil(t, got.Id)
	require.Equal(t, "param-uuid", *got.Id)
}

func TestParseCreateParameterResponse_successWithNilID(t *testing.T) {
	t.Parallel()

	union := kboapi.CreateParamResponse{}
	require.NoError(t, union.FromSyntheticsPostParameterResponse(kboapi.SyntheticsPostParameterResponse{}))

	resp := &kboapi.PostParametersResponse{
		Body:         []byte(`{}`),
		HTTPResponse: &http.Response{StatusCode: http.StatusOK, Status: "200 OK"},
		JSON200:      &union,
	}

	_, diags := parseCreateParameterResponse(resp, "my-key")
	require.True(t, diags.HasError())
	require.Contains(t, diags.Errors()[0].Summary(), "Unexpected nil id")
	require.Contains(t, diags.Errors()[0].Detail(), "did not include a parameter id")
}

func TestParseCreateParameterResponse_nilResponse(t *testing.T) {
	t.Parallel()

	_, diags := parseCreateParameterResponse(nil, "my-key")
	require.True(t, diags.HasError())
	require.Contains(t, diags.Errors()[0].Summary(), "Failed to create parameter")
}
