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
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateSpace_409Conflict(t *testing.T) {
	t.Parallel()

	const spaceID = "default"

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/api/spaces/space", r.URL.Path)
		w.WriteHeader(http.StatusConflict)
	}))
	t.Cleanup(srv.Close)

	client := newTestClient(t, srv)
	_, diags := CreateSpace(context.Background(), client, kbapi.PostSpacesSpaceJSONRequestBody{
		Id:   spaceID,
		Name: "Default",
	})

	require.True(t, diags.HasError())
	assert.Equal(t, "Kibana space already exists", diags.Errors()[0].Summary())
	detail := diags.Errors()[0].Detail()
	assert.Contains(t, detail, spaceID)
	assert.Contains(t, detail, "terraform import elasticstack_kibana_space.<NAME> "+spaceID)
}
