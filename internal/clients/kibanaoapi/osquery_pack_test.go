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
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOsqueryPackShardsFromMap(t *testing.T) {
	t.Run("returns nil for nil input", func(t *testing.T) {
		assert.Nil(t, osqueryPackShardsFromMap(nil))
	})

	t.Run("returns nil for empty map", func(t *testing.T) {
		empty := kbapi.SecurityOsqueryAPIShards{}
		assert.Nil(t, osqueryPackShardsFromMap(&empty))
	})

	t.Run("converts float32 map to float64", func(t *testing.T) {
		shards := kbapi.SecurityOsqueryAPIShards{
			"policy-a": 50,
			"policy-b": 100,
		}

		result := osqueryPackShardsFromMap(&shards)

		assert.Equal(t, OsqueryPackShards{"policy-a": 50, "policy-b": 100}, result)
	})
}

func TestOsqueryPackShardsFromCreateArray(t *testing.T) {
	t.Run("returns nil for nil input", func(t *testing.T) {
		assert.Nil(t, osqueryPackShardsFromCreateArray(nil))
	})

	t.Run("returns nil for empty array", func(t *testing.T) {
		empty := []struct {
			Key   *string  `json:"key,omitempty"`
			Value *float32 `json:"value,omitempty"`
		}{}
		assert.Nil(t, osqueryPackShardsFromCreateArray(&empty))
	})

	t.Run("converts key-value array to map", func(t *testing.T) {
		keyA := "policy-a"
		keyB := "policy-b"
		valA := float32(25)
		valB := float32(75)
		shards := []struct {
			Key   *string  `json:"key,omitempty"`
			Value *float32 `json:"value,omitempty"`
		}{
			{Key: &keyA, Value: &valA},
			{Key: &keyB, Value: &valB},
		}

		result := osqueryPackShardsFromCreateArray(&shards)

		assert.Equal(t, OsqueryPackShards{"policy-a": 25, "policy-b": 75}, result)
	})

	t.Run("skips entries with nil key", func(t *testing.T) {
		val := float32(50)
		shards := []struct {
			Key   *string  `json:"key,omitempty"`
			Value *float32 `json:"value,omitempty"`
		}{
			{Key: nil, Value: &val},
		}

		assert.Nil(t, osqueryPackShardsFromCreateArray(&shards))
	})

	t.Run("uses zero when key is present but value is nil", func(t *testing.T) {
		key := "policy-a"
		shards := []struct {
			Key   *string  `json:"key,omitempty"`
			Value *float32 `json:"value,omitempty"`
		}{
			{Key: &key, Value: nil},
		}

		result := osqueryPackShardsFromCreateArray(&shards)

		assert.Equal(t, OsqueryPackShards{"policy-a": 0}, result)
	})

	t.Run("keeps valid entries when mixed with nil-key rows", func(t *testing.T) {
		validKey := "policy-b"
		validVal := float32(60)
		droppedVal := float32(99)
		shards := []struct {
			Key   *string  `json:"key,omitempty"`
			Value *float32 `json:"value,omitempty"`
		}{
			{Key: nil, Value: &droppedVal},
			{Key: &validKey, Value: &validVal},
		}

		result := osqueryPackShardsFromCreateArray(&shards)

		assert.Equal(t, OsqueryPackShards{"policy-b": 60}, result)
	})
}

func TestOsqueryPackDetailFromFindResponse(t *testing.T) {
	readOnly := true
	packType := "osquery-pack"
	namespaces := []string{"default", "security"}
	shards := kbapi.SecurityOsqueryAPIShards{"policy-1": 50}

	resp := &kbapi.SecurityOsqueryAPIFindPackResponse{}
	resp.Data.Name = "find-pack"
	resp.Data.SavedObjectId = "find-id"
	resp.Data.ReadOnly = &readOnly
	resp.Data.Namespaces = &namespaces
	resp.Data.Type = &packType
	resp.Data.Shards = &shards

	detail := osqueryPackDetailFromFindResponse(resp)

	require.NotNil(t, detail)
	assert.Equal(t, kbapi.SecurityOsqueryAPIPackName("find-pack"), detail.Name)
	assert.Equal(t, "find-id", detail.SavedObjectID)
	assert.Equal(t, &readOnly, detail.ReadOnly)
	assert.Equal(t, &namespaces, detail.Namespaces)
	assert.Equal(t, &packType, detail.Type)
	assert.Equal(t, OsqueryPackShards{"policy-1": 50}, detail.Shards)
}

func TestOsqueryPackDetailFromCreateResponse(t *testing.T) {
	key := "policy-a"
	val := float32(33)
	shards := []struct {
		Key   *string  `json:"key,omitempty"`
		Value *float32 `json:"value,omitempty"`
	}{
		{Key: &key, Value: &val},
	}

	resp := &kbapi.SecurityOsqueryAPICreatePacksResponse{}
	resp.Data.Name = "create-pack"
	resp.Data.SavedObjectId = "create-id"
	resp.Data.Shards = &shards

	detail := osqueryPackDetailFromCreateResponse(resp)

	require.NotNil(t, detail)
	assert.Equal(t, kbapi.SecurityOsqueryAPIPackName("create-pack"), detail.Name)
	assert.Equal(t, "create-id", detail.SavedObjectID)
	assert.Nil(t, detail.ReadOnly)
	assert.Nil(t, detail.Namespaces)
	assert.Nil(t, detail.Type)
	assert.Equal(t, OsqueryPackShards{"policy-a": 33}, detail.Shards)
}

func TestOsqueryPackDetailFromUpdateResponse(t *testing.T) {
	t.Run("maps optional name and saved_object_id pointers", func(t *testing.T) {
		var resp kbapi.SecurityOsqueryAPIUpdatePacksResponse
		require.NoError(t, json.Unmarshal([]byte(`{
			"data": {
				"name": "updated-pack",
				"saved_object_id": "update-id",
				"shards": {"policy-x": 80}
			}
		}`), &resp))

		detail := osqueryPackDetailFromUpdateResponse(&resp)

		require.NotNil(t, detail)
		assert.Equal(t, kbapi.SecurityOsqueryAPIPackName("updated-pack"), detail.Name)
		assert.Equal(t, "update-id", detail.SavedObjectID)
		assert.Equal(t, OsqueryPackShards{"policy-x": 80}, detail.Shards)
	})

	t.Run("leaves name and saved_object_id empty when pointers absent", func(t *testing.T) {
		var resp kbapi.SecurityOsqueryAPIUpdatePacksResponse
		require.NoError(t, json.Unmarshal([]byte(`{"data":{"shards":{"policy-y": 10}}}`), &resp))

		detail := osqueryPackDetailFromUpdateResponse(&resp)

		require.NotNil(t, detail)
		assert.Empty(t, detail.Name)
		assert.Empty(t, detail.SavedObjectID)
		assert.Equal(t, OsqueryPackShards{"policy-y": 10}, detail.Shards)
	})

	t.Run("returns nil when data is nil", func(t *testing.T) {
		assert.Nil(t, osqueryPackDetailFromUpdateResponse(&kbapi.SecurityOsqueryAPIUpdatePacksResponse{}))
	})
}

func TestUpdateOsqueryPackNilDataDiagnostic(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method)
		assert.Equal(t, "/api/osquery/packs/pack-id", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	}))
	t.Cleanup(srv.Close)

	client := newTestClient(t, srv)
	_, diags := UpdateOsqueryPack(context.Background(), client, "default", "pack-id", kbapi.OsqueryUpdatePacksJSONRequestBody{})

	require.True(t, diags.HasError())
	assert.Equal(t, "Failed to parse response", diags[0].Summary())
	assert.Contains(t, diags[0].Detail(), "update response data was nil")
}

func TestGetOsqueryPack404(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/api/osquery/packs/missing-pack", r.URL.Path)
		w.WriteHeader(http.StatusNotFound)
	}))
	t.Cleanup(srv.Close)

	client := newTestClient(t, srv)
	result, diags := GetOsqueryPack(context.Background(), client, "default", "missing-pack")

	assert.False(t, diags.HasError(), diags)
	assert.Nil(t, result)
}

func TestDeleteOsqueryPack404(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)
		assert.Equal(t, "/api/osquery/packs/missing-pack", r.URL.Path)
		w.WriteHeader(http.StatusNotFound)
	}))
	t.Cleanup(srv.Close)

	client := newTestClient(t, srv)
	diags := DeleteOsqueryPack(context.Background(), client, "default", "missing-pack")

	assert.False(t, diags.HasError(), diags)
}
