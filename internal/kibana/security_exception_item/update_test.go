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

package securityexceptionitem

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func newUpdatePathTestClient(t *testing.T, readStatus int) (*clients.KibanaScopedClient, *updatePathRequests) {
	t.Helper()

	item := kbapi.SecurityExceptionsAPIExceptionListItem{
		Id:            "item-id",
		ItemId:        "item-id",
		ListId:        "list-id",
		Name:          "name",
		NamespaceType: "agnostic",
		Type:          "simple",
		Description:   "desc",
		TieBreakerId:  "tie-breaker-id",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
		CreatedBy:     "analyst",
		UpdatedBy:     "analyst",
		Comments: kbapi.SecurityExceptionsAPIExceptionListItemCommentArray{
			{Id: "id-alpha", Comment: "alpha", CreatedBy: "analyst"},
		},
	}
	itemJSON, err := json.Marshal(item)
	require.NoError(t, err)

	reqs := &updatePathRequests{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/status":
			reqs.status++
			w.WriteHeader(http.StatusInternalServerError)
		case r.Method == http.MethodGet:
			reqs.read++
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(readStatus)
			if readStatus == http.StatusOK {
				_, _ = w.Write(itemJSON)
			}
		case r.Method == http.MethodPut:
			reqs.put++
			w.WriteHeader(http.StatusInternalServerError)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(srv.Close)

	kibOapi, err := kibanaoapi.NewClient(kibanaoapi.Config{URL: srv.URL})
	require.NoError(t, err)
	return clients.NewKibanaScopedClientForTest(kibOapi), reqs
}

type updatePathRequests struct {
	read   int
	put    int
	status int
}

func emptyRequiredPlan() ExceptionItemModel {
	return ExceptionItemModel{
		RequiredComments: types.ListValueMust(types.StringType, nil),
		NamespaceType:    types.StringValue("agnostic"),
	}
}

func TestReconcileRequiredComments_EmptyRequired(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("failed fresh read aborts before any PUT", func(t *testing.T) {
		t.Parallel()
		client, reqs := newUpdatePathTestClient(t, http.StatusInternalServerError)

		body, diags := reconcileRequiredComments(ctx, client, entitycore.KibanaWriteRequest[ExceptionItemModel]{
			Plan:    emptyRequiredPlan(),
			WriteID: "item-id",
			SpaceID: "default",
		})

		require.Nil(t, body)
		require.True(t, diags.HasError(), "a failed fresh read must abort the update: %s", diags)
		require.Equal(t, 1, reqs.read, "the update path must perform a fresh read")
		require.Zero(t, reqs.put, "no PUT may be sent when the fresh read fails")
		require.Zero(t, reqs.status, "no version discovery may happen before the fresh read succeeds")
	})

	t.Run("fresh read with nothing missing omits comments and skips version discovery", func(t *testing.T) {
		t.Parallel()
		client, reqs := newUpdatePathTestClient(t, http.StatusOK)

		body, diags := reconcileRequiredComments(ctx, client, entitycore.KibanaWriteRequest[ExceptionItemModel]{
			Plan:    emptyRequiredPlan(),
			WriteID: "item-id",
			SpaceID: "default",
		})

		require.False(t, diags.HasError(), "%s", diags)
		require.Nil(t, body, "the update body must omit comments when nothing is missing")
		require.Equal(t, 1, reqs.read, "the update path must perform a fresh read")
		require.Zero(t, reqs.put, "reconciliation itself must not PUT")
		require.Zero(t, reqs.status, "no version discovery is needed when nothing is missing")
	})
}
