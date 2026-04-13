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

package agentdownloadsource

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func TestModelToAPICreateModel(t *testing.T) {
	ctx := context.Background()
	state := model{
		Name:     types.StringValue("download-source"),
		Host:     types.StringValue("https://artifacts.example.com/elastic-agent"),
		Default:  types.BoolValue(true),
		ProxyID:  types.StringValue("proxy-123"),
		SourceID: types.StringValue("source-123"),
	}

	body := state.toAPICreateModel(ctx)
	require.Equal(t, "download-source", body.Name)
	require.Equal(t, "https://artifacts.example.com/elastic-agent", body.Host)
	require.NotNil(t, body.IsDefault)
	require.True(t, *body.IsDefault)
	require.NotNil(t, body.ProxyId)
	require.Equal(t, "proxy-123", *body.ProxyId)
	require.NotNil(t, body.Id)
	require.Equal(t, "source-123", *body.Id)
}

func TestModelToAPICreateModelWithoutSourceID(t *testing.T) {
	ctx := context.Background()
	state := model{
		Name:     types.StringValue("download-source"),
		Host:     types.StringValue("https://artifacts.example.com/elastic-agent"),
		Default:  types.BoolValue(false),
		ProxyID:  types.StringNull(),
		SourceID: types.StringNull(),
	}

	body := state.toAPICreateModel(ctx)
	require.Equal(t, "download-source", body.Name)
	require.Equal(t, "https://artifacts.example.com/elastic-agent", body.Host)
	require.NotNil(t, body.IsDefault)
	require.False(t, *body.IsDefault)
	require.Nil(t, body.ProxyId)
	require.Nil(t, body.Id)
}

func TestModelToAPIUpdateModel(t *testing.T) {
	ctx := context.Background()
	state := model{
		Name:    types.StringValue("updated-name"),
		Host:    types.StringValue("https://artifacts.example.com/elastic-agent-updated"),
		Default: types.BoolValue(true),
		ProxyID: types.StringValue("proxy-456"),
	}

	body := state.toAPIUpdateModel(ctx)
	require.Equal(t, "updated-name", body.Name)
	require.Equal(t, "https://artifacts.example.com/elastic-agent-updated", body.Host)
	require.NotNil(t, body.IsDefault)
	require.True(t, *body.IsDefault)
	require.NotNil(t, body.ProxyId)
	require.Equal(t, "proxy-456", *body.ProxyId)
}
