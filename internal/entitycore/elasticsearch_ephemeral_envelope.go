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

package entitycore

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	providerschema "github.com/elastic/terraform-provider-elasticstack/internal/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	eschema "github.com/hashicorp/terraform-plugin-framework/ephemeral/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// ElasticsearchEphemeralModel is the type constraint for models passed to
// [NewElasticsearchEphemeralResource]. Concrete types must provide
// GetElasticsearchConnection, typically by embedding [ElasticsearchConnectionField].
type ElasticsearchEphemeralModel interface {
	GetElasticsearchConnection() types.List
}

type ElasticsearchEphemeralOpenFunc[T ElasticsearchEphemeralModel, S any] func(
	context.Context,
	*clients.ElasticsearchScopedClient,
	OpenRequest[T],
) (OpenResult[T, S], diag.Diagnostics)

type ElasticsearchEphemeralCloseFunc[S any] func(
	context.Context,
	*clients.ElasticsearchScopedClient,
	CloseRequest[S],
) (CloseResponse, diag.Diagnostics)

// ElasticsearchEphemeralOptions configures [NewElasticsearchEphemeralResource].
// Schema, Open, and Close must be non-nil or the constructor panics.
type ElasticsearchEphemeralOptions[T ElasticsearchEphemeralModel, S any] struct {
	Schema func(context.Context) eschema.Schema
	Open   ElasticsearchEphemeralOpenFunc[T, S]
	Close  ElasticsearchEphemeralCloseFunc[S]
}

// ElasticsearchEphemeralResource implements [ephemeral.EphemeralResource] and
// related interfaces for Elasticsearch-backed ephemeral resources.
type ElasticsearchEphemeralResource[T ElasticsearchEphemeralModel, S any] = genericEphemeralResource[T, S, *clients.ElasticsearchScopedClient]

// NewElasticsearchEphemeralResource returns an [ephemeral.EphemeralResource]
// that owns Metadata, Configure, Schema (with elasticsearch_connection block
// injection), Open, and Close for the Elasticsearch namespace.
func NewElasticsearchEphemeralResource[T ElasticsearchEphemeralModel, S any](
	name string,
	opts ElasticsearchEphemeralOptions[T, S],
) ephemeral.EphemeralResource {
	if opts.Schema == nil {
		panic("entitycore: ElasticsearchEphemeralOptions.Schema must not be nil")
	}
	if opts.Open == nil {
		panic("entitycore: ElasticsearchEphemeralOptions.Open must not be nil")
	}
	if opts.Close == nil {
		panic("entitycore: ElasticsearchEphemeralOptions.Close must not be nil")
	}
	mustBePlainGoCloseState[S]()

	return &genericEphemeralResource[T, S, *clients.ElasticsearchScopedClient]{
		EphemeralBase: NewEphemeralBase(ComponentElasticsearch, name),
		schemaFactory: opts.Schema,
		openFunc:      opts.Open,
		closeFunc:     opts.Close,
		adapter: ephemeralAdapter[T, *clients.ElasticsearchScopedClient]{
			getConnection: func(model T) types.List { return model.GetElasticsearchConnection() },
			getClient: func(ctx context.Context, factory *clients.ProviderClientFactory, connection types.List) (*clients.ElasticsearchScopedClient, diag.Diagnostics) {
				return factory.GetElasticsearchClient(ctx, connection)
			},
			encodeConn:         encodeElasticsearchConnection,
			decodeConn:         decodeElasticsearchConnection,
			schemaBlockKey:     blockElasticsearchConnection,
			schemaBlockFactory: providerschema.GetEsEphemeralConnectionBlock,
			errorSummary:       "Elasticsearch ephemeral envelope internal error",
		},
	}
}

var (
	_ ephemeral.EphemeralResource              = (*ElasticsearchEphemeralResource[ElasticsearchConnectionField, struct{}])(nil)
	_ ephemeral.EphemeralResourceWithConfigure = (*ElasticsearchEphemeralResource[ElasticsearchConnectionField, struct{}])(nil)
	_ ephemeral.EphemeralResourceWithClose     = (*ElasticsearchEphemeralResource[ElasticsearchConnectionField, struct{}])(nil)
)
