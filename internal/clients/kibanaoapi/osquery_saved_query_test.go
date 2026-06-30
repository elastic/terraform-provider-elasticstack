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
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOsquerySavedQueryCreateEntityFrom(t *testing.T) {
	t.Run("returns nil for nil response", func(t *testing.T) {
		assert.Nil(t, osquerySavedQueryCreateEntityFrom(nil))
	})

	t.Run("unwraps data payload with union interval and version", func(t *testing.T) {
		var resp kbapi.SecurityOsqueryAPICreateSavedQueryResponse
		require.NoError(t, json.Unmarshal([]byte(`{
			"data": {
				"id": "list_processes",
				"query": "SELECT * FROM processes",
				"saved_object_id": "osquery-saved-query/list_processes",
				"interval": 3600,
				"version": "1.0.0",
				"created_by_profile_uid": "profile-1",
				"updated_by_profile_uid": "profile-2"
			}
		}`), &resp))

		entity := osquerySavedQueryCreateEntityFrom(&resp)
		require.NotNil(t, entity)
		assert.Equal(t, kbapi.SecurityOsqueryAPISavedQueryId("list_processes"), entity.ID)
		assert.Equal(t, "osquery-saved-query/list_processes", entity.SavedObjectID)
		require.NotNil(t, entity.Query)
		assert.Equal(t, kbapi.SecurityOsqueryAPIQuery("SELECT * FROM processes"), *entity.Query)
		require.NotNil(t, entity.CreatedByProfileUID)
		assert.Equal(t, "profile-1", *entity.CreatedByProfileUID)
		require.NotNil(t, entity.UpdatedByProfileUID)
		assert.Equal(t, "profile-2", *entity.UpdatedByProfileUID)

		interval, err := entity.Interval.AsSecurityOsqueryAPICreateSavedQueryResponseDataInterval0()
		require.NoError(t, err)
		assert.Equal(t, 3600, interval)

		version, err := entity.Version.AsSecurityOsqueryAPICreateSavedQueryResponseDataVersion1()
		require.NoError(t, err)
		assert.Equal(t, "1.0.0", version)
	})
}

func TestOsquerySavedQueryGetEntityFrom(t *testing.T) {
	t.Run("returns nil for nil response", func(t *testing.T) {
		assert.Nil(t, osquerySavedQueryGetEntityFrom(nil))
	})

	t.Run("unwraps data payload with string interval arm", func(t *testing.T) {
		var resp kbapi.SecurityOsqueryAPIFindSavedQueryDetailResponse
		require.NoError(t, json.Unmarshal([]byte(`{
			"data": {
				"id": "list_processes",
				"query": "SELECT * FROM processes",
				"saved_object_id": "osquery-saved-query/list_processes",
				"interval": "7200",
				"version": 2
			}
		}`), &resp))

		entity := osquerySavedQueryGetEntityFrom(&resp)
		require.NotNil(t, entity)
		assert.Equal(t, kbapi.SecurityOsqueryAPISavedQueryId("list_processes"), entity.ID)

		interval, err := entity.Interval.AsSecurityOsqueryAPIFindSavedQueryDetailResponseDataInterval1()
		require.NoError(t, err)
		assert.Equal(t, "7200", interval)

		version, err := entity.Version.AsSecurityOsqueryAPIFindSavedQueryDetailResponseDataVersion0()
		require.NoError(t, err)
		assert.Equal(t, 2, version)
	})
}

func TestOsquerySavedQueryUpdateEntityFrom(t *testing.T) {
	t.Run("returns nil for nil response", func(t *testing.T) {
		assert.Nil(t, osquerySavedQueryUpdateEntityFrom(nil))
	})

	t.Run("unwraps data payload with plain string version", func(t *testing.T) {
		var resp kbapi.SecurityOsqueryAPIUpdateSavedQueryResponse
		require.NoError(t, json.Unmarshal([]byte(`{
			"data": {
				"id": "list_processes",
				"query": "SELECT pid FROM processes",
				"saved_object_id": "osquery-saved-query/list_processes",
				"interval": 1800,
				"version": "2.1.0"
			}
		}`), &resp))

		entity := osquerySavedQueryUpdateEntityFrom(&resp)
		require.NotNil(t, entity)
		assert.Equal(t, kbapi.SecurityOsqueryAPISavedQueryId("list_processes"), entity.ID)
		require.NotNil(t, entity.Version)
		assert.Equal(t, "2.1.0", *entity.Version)

		interval, err := entity.Interval.AsSecurityOsqueryAPIUpdateSavedQueryResponseDataInterval0()
		require.NoError(t, err)
		assert.Equal(t, 1800, interval)
	})
}
