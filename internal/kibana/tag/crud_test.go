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

package tag

import (
	"context"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	kibanaoapi "github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func restoreTagAPIs(t *testing.T) {
	t.Helper()
	origGetTagAPI := getTagAPI
	origUpsertTagAPI := upsertTagAPI
	origDeleteTagAPI := deleteTagAPI
	t.Cleanup(func() {
		getTagAPI = origGetTagAPI
		upsertTagAPI = origUpsertTagAPI
		deleteTagAPI = origDeleteTagAPI
	})
}

func TestReadTag_managedTagReturnsDiagnostic(t *testing.T) {
	restoreTagAPIs(t)
	managed := true
	getTagAPI = func(context.Context, *kibanaoapi.Client, string, string) (*kibanaoapi.TagDetail, diag.Diagnostics) {
		return &kibanaoapi.TagDetail{ID: "tag-123", Managed: &managed}, nil
	}

	_, found, diags := readTag(context.Background(), &clients.KibanaScopedClient{}, "tag-123", "default", tagModel{})

	assert.False(t, found)
	require.True(t, diags.HasError())
	assert.Contains(t, diags[0].Detail(), "tag-123")
}

func TestUpdateTag_managedTagDoesNotUpsert(t *testing.T) {
	restoreTagAPIs(t)
	managed := true
	upsertCalled := false
	getTagAPI = func(context.Context, *kibanaoapi.Client, string, string) (*kibanaoapi.TagDetail, diag.Diagnostics) {
		return &kibanaoapi.TagDetail{ID: "tag-123", Managed: &managed}, nil
	}
	upsertTagAPI = func(context.Context, *kibanaoapi.Client, string, string, kbapi.PutTagsIdJSONRequestBody) (*kibanaoapi.TagDetail, diag.Diagnostics) {
		upsertCalled = true
		return &kibanaoapi.TagDetail{}, nil
	}

	_, diags := updateTag(context.Background(), &clients.KibanaScopedClient{}, entitycore.KibanaWriteRequest[tagModel]{
		Plan:    tagModel{tagBaseModel: tagBaseModel{Name: types.StringValue("updated")}},
		WriteID: "tag-123",
		SpaceID: "default",
	})

	require.True(t, diags.HasError())
	assert.False(t, upsertCalled)
}

func TestUpdateTag_missingTagUpserts(t *testing.T) {
	restoreTagAPIs(t)
	upsertCalled := false
	getTagAPI = func(context.Context, *kibanaoapi.Client, string, string) (*kibanaoapi.TagDetail, diag.Diagnostics) {
		return nil, nil
	}
	upsertTagAPI = func(context.Context, *kibanaoapi.Client, string, string, kbapi.PutTagsIdJSONRequestBody) (*kibanaoapi.TagDetail, diag.Diagnostics) {
		upsertCalled = true
		return &kibanaoapi.TagDetail{ID: "tag-123"}, nil
	}

	_, diags := updateTag(context.Background(), &clients.KibanaScopedClient{}, entitycore.KibanaWriteRequest[tagModel]{
		Plan:    tagModel{tagBaseModel: tagBaseModel{Name: types.StringValue("updated")}},
		WriteID: "tag-123",
		SpaceID: "default",
	})

	require.False(t, diags.HasError())
	assert.True(t, upsertCalled)
}

func TestDeleteTag_managedTagDoesNotDelete(t *testing.T) {
	restoreTagAPIs(t)
	managed := true
	deleteCalled := false
	getTagAPI = func(context.Context, *kibanaoapi.Client, string, string) (*kibanaoapi.TagDetail, diag.Diagnostics) {
		return &kibanaoapi.TagDetail{ID: "tag-123", Managed: &managed}, nil
	}
	deleteTagAPI = func(context.Context, *kibanaoapi.Client, string, string) diag.Diagnostics {
		deleteCalled = true
		return nil
	}

	diags := deleteTag(context.Background(), &clients.KibanaScopedClient{}, "tag-123", "default", tagModel{})

	require.True(t, diags.HasError())
	assert.False(t, deleteCalled)
}

func TestDeleteTag_missingTagIsNoop(t *testing.T) {
	restoreTagAPIs(t)
	deleteCalled := false
	getTagAPI = func(context.Context, *kibanaoapi.Client, string, string) (*kibanaoapi.TagDetail, diag.Diagnostics) {
		return nil, nil
	}
	deleteTagAPI = func(context.Context, *kibanaoapi.Client, string, string) diag.Diagnostics {
		deleteCalled = true
		return nil
	}

	diags := deleteTag(context.Background(), &clients.KibanaScopedClient{}, "tag-123", "default", tagModel{})

	require.False(t, diags.HasError())
	assert.False(t, deleteCalled)
}
