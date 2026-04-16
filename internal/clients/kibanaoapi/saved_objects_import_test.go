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

package kibanaoapi_test

import (
	"encoding/json"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	kibanaoapi "github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestImportSavedObjects_MultipartFormat(t *testing.T) {
	const ndjsonContent = `{"id":"obj-1","type":"config"}` + "\n" + `{"exportedCount":1}`

	var capturedContentType string
	var capturedBody []byte

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		capturedContentType = r.Header.Get("Content-Type")
		body, err := io.ReadAll(r.Body)
		assert.NoError(t, err)
		capturedBody = body

		rw.Header().Set("Content-Type", "application/json")
		rw.WriteHeader(http.StatusOK)
		resp := map[string]any{
			"success":        true,
			"successCount":   float32(1),
			"errors":         []any{},
			"successResults": []any{},
		}
		_ = json.NewEncoder(rw).Encode(resp)
	}))
	defer server.Close()

	t.Setenv("ELASTICSEARCH_URL", server.URL)
	t.Setenv("KIBANA_ENDPOINT", server.URL)

	apiClient, err := clients.NewAcceptanceTestingKibanaScopedClient()
	require.NoError(t, err)

	oapiClient, err := apiClient.GetKibanaOapiClient()
	require.NoError(t, err)

	params := kbapi.PostSavedObjectsImportParams{}
	result, diags := kibanaoapi.ImportSavedObjects(t.Context(), oapiClient, "", []byte(ndjsonContent), params)
	require.Nil(t, diags)
	require.NotNil(t, result)
	assert.True(t, result.Success)
	assert.Equal(t, int64(1), result.SuccessCount)

	// Verify multipart format
	mediaType, params2, err := mime.ParseMediaType(capturedContentType)
	require.NoError(t, err)
	assert.Equal(t, "multipart/form-data", mediaType)

	boundary := params2["boundary"]
	require.NotEmpty(t, boundary)

	mr := multipart.NewReader(strings.NewReader(string(capturedBody)), boundary)
	part, err := mr.NextPart()
	require.NoError(t, err)
	assert.Equal(t, "file", part.FormName())
	assert.Equal(t, "export.ndjson", part.FileName())

	partContent, err := io.ReadAll(part)
	require.NoError(t, err)
	assert.Equal(t, ndjsonContent, string(partContent)) //nolint:testifylint // NDJSON is not a single JSON document; JSONEq does not apply

	// Verify no more parts
	_, err = mr.NextPart()
	assert.Equal(t, io.EOF, err)
}

func TestImportSavedObjects_SpaceAwarePath(t *testing.T) {
	var capturedPath string

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		capturedPath = req.URL.Path
		rw.Header().Set("Content-Type", "application/json")
		rw.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(rw).Encode(map[string]any{
			"success":        true,
			"successCount":   float32(0),
			"errors":         []any{},
			"successResults": []any{},
		})
	}))
	defer server.Close()

	t.Setenv("ELASTICSEARCH_URL", server.URL)
	t.Setenv("KIBANA_ENDPOINT", server.URL)

	apiClient, err := clients.NewAcceptanceTestingKibanaScopedClient()
	require.NoError(t, err)

	oapiClient, err := apiClient.GetKibanaOapiClient()
	require.NoError(t, err)

	params := kbapi.PostSavedObjectsImportParams{}
	_, diags := kibanaoapi.ImportSavedObjects(t.Context(), oapiClient, "my-space", []byte("{}"), params)
	require.Nil(t, diags)
	assert.Contains(t, capturedPath, "/s/my-space/")
}

func TestImportSavedObjects_400Response(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
		rw.Header().Set("Content-Type", "application/json")
		rw.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(rw).Encode(map[string]any{
			"error":      "Bad Request",
			"message":    "invalid file format",
			"statusCode": 400,
		})
	}))
	defer server.Close()

	t.Setenv("ELASTICSEARCH_URL", server.URL)
	t.Setenv("KIBANA_ENDPOINT", server.URL)

	apiClient, err := clients.NewAcceptanceTestingKibanaScopedClient()
	require.NoError(t, err)

	oapiClient, err := apiClient.GetKibanaOapiClient()
	require.NoError(t, err)

	params := kbapi.PostSavedObjectsImportParams{}
	result, diags := kibanaoapi.ImportSavedObjects(t.Context(), oapiClient, "", []byte("bad data"), params)
	require.NotNil(t, diags)
	require.True(t, diags.HasError())
	assert.Nil(t, result)
}

func TestImportSavedObjects_UnexpectedStatusCode(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
		rw.WriteHeader(http.StatusInternalServerError)
		_, _ = rw.Write([]byte("internal server error"))
	}))
	defer server.Close()

	t.Setenv("ELASTICSEARCH_URL", server.URL)
	t.Setenv("KIBANA_ENDPOINT", server.URL)

	apiClient, err := clients.NewAcceptanceTestingKibanaScopedClient()
	require.NoError(t, err)

	oapiClient, err := apiClient.GetKibanaOapiClient()
	require.NoError(t, err)

	params := kbapi.PostSavedObjectsImportParams{}
	result, diags := kibanaoapi.ImportSavedObjects(t.Context(), oapiClient, "", []byte("{}"), params)
	require.NotNil(t, diags)
	require.True(t, diags.HasError())
	assert.Nil(t, result)
}

func TestImportSavedObjects_SuccessCountConvertedToInt64(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
		rw.Header().Set("Content-Type", "application/json")
		rw.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(rw).Encode(map[string]any{
			"success":        true,
			"successCount":   float32(42),
			"errors":         []any{},
			"successResults": []any{},
		})
	}))
	defer server.Close()

	t.Setenv("ELASTICSEARCH_URL", server.URL)
	t.Setenv("KIBANA_ENDPOINT", server.URL)

	apiClient, err := clients.NewAcceptanceTestingKibanaScopedClient()
	require.NoError(t, err)

	oapiClient, err := apiClient.GetKibanaOapiClient()
	require.NoError(t, err)

	params := kbapi.PostSavedObjectsImportParams{}
	result, diags := kibanaoapi.ImportSavedObjects(t.Context(), oapiClient, "", []byte("{}"), params)
	require.Nil(t, diags)
	require.NotNil(t, result)
	assert.Equal(t, int64(42), result.SuccessCount)
}

func TestImportSavedObjects_QueryParamWiring(t *testing.T) {
	var capturedQuery map[string][]string

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		capturedQuery = map[string][]string(req.URL.Query())
		rw.Header().Set("Content-Type", "application/json")
		rw.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(rw).Encode(map[string]any{
			"success":        true,
			"successCount":   float32(0),
			"errors":         []any{},
			"successResults": []any{},
		})
	}))
	defer server.Close()

	t.Setenv("ELASTICSEARCH_URL", server.URL)
	t.Setenv("KIBANA_ENDPOINT", server.URL)

	apiClient, err := clients.NewAcceptanceTestingKibanaScopedClient()
	require.NoError(t, err)

	oapiClient, err := apiClient.GetKibanaOapiClient()
	require.NoError(t, err)

	trueVal := true
	params := kbapi.PostSavedObjectsImportParams{
		Overwrite:         &trueVal,
		CreateNewCopies:   &trueVal,
		CompatibilityMode: &trueVal,
	}
	_, _ = kibanaoapi.ImportSavedObjects(t.Context(), oapiClient, "", []byte("{}"), params)

	assert.Equal(t, []string{"true"}, capturedQuery["overwrite"], "overwrite query param should be set")
	assert.Equal(t, []string{"true"}, capturedQuery["createNewCopies"], "createNewCopies query param should be set")
	assert.Equal(t, []string{"true"}, capturedQuery["compatibilityMode"], "compatibilityMode query param should be set")
}
