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

// KibanaEphemeralModel is the type constraint for models passed to
// [NewKibanaEphemeralResource]. Concrete types must provide GetKibanaConnection,
// typically by embedding [KibanaConnectionField].
type KibanaEphemeralModel interface {
	GetKibanaConnection() types.List
}

type KibanaEphemeralOpenFunc[T KibanaEphemeralModel, S any] func(
	context.Context,
	*clients.KibanaScopedClient,
	OpenRequest[T],
) (OpenResult[T, S], diag.Diagnostics)

type KibanaEphemeralCloseFunc[S any] func(
	context.Context,
	*clients.KibanaScopedClient,
	CloseRequest[S],
) (CloseResponse, diag.Diagnostics)

// KibanaEphemeralOptions configures [NewKibanaEphemeralResource].
// Schema, Open, and Close must be non-nil or the constructor panics.
type KibanaEphemeralOptions[T KibanaEphemeralModel, S any] struct {
	Schema func(context.Context) eschema.Schema
	Open   KibanaEphemeralOpenFunc[T, S]
	Close  KibanaEphemeralCloseFunc[S]
}

// KibanaEphemeralResource implements [ephemeral.EphemeralResource] and related
// interfaces for Kibana-backed ephemeral resources.
type KibanaEphemeralResource[T KibanaEphemeralModel, S any] = genericEphemeralResource[T, S, *clients.KibanaScopedClient]

// NewKibanaEphemeralResource returns an [ephemeral.EphemeralResource] that
// owns Metadata, Configure, Schema (with kibana_connection block injection),
// Open, and Close for the Kibana namespace.
func NewKibanaEphemeralResource[T KibanaEphemeralModel, S any](
	name string,
	opts KibanaEphemeralOptions[T, S],
) ephemeral.EphemeralResource {
	if opts.Schema == nil {
		panic("entitycore: KibanaEphemeralOptions.Schema must not be nil")
	}
	if opts.Open == nil {
		panic("entitycore: KibanaEphemeralOptions.Open must not be nil")
	}
	if opts.Close == nil {
		panic("entitycore: KibanaEphemeralOptions.Close must not be nil")
	}
	mustBePlainGoCloseState[S]()

	return &genericEphemeralResource[T, S, *clients.KibanaScopedClient]{
		EphemeralBase: NewEphemeralBase(ComponentKibana, name),
		schemaFactory: opts.Schema,
		openFunc:      opts.Open,
		closeFunc:     opts.Close,
		adapter: ephemeralAdapter[T, *clients.KibanaScopedClient]{
			getConnection: func(model T) types.List { return model.GetKibanaConnection() },
			getClient: func(ctx context.Context, factory *clients.ProviderClientFactory, connection types.List) (*clients.KibanaScopedClient, diag.Diagnostics) {
				return factory.GetKibanaClient(ctx, connection)
			},
			encodeConn:         encodeKibanaConnection,
			decodeConn:         decodeKibanaConnection,
			schemaBlockKey:     "kibana_connection",
			schemaBlockFactory: providerschema.GetKbEphemeralConnectionBlock,
			errorSummary:       "Kibana ephemeral envelope internal error",
		},
	}
}

var (
	_ ephemeral.EphemeralResource              = (*KibanaEphemeralResource[KibanaConnectionField, struct{}])(nil)
	_ ephemeral.EphemeralResourceWithConfigure = (*KibanaEphemeralResource[KibanaConnectionField, struct{}])(nil)
	_ ephemeral.EphemeralResourceWithClose     = (*KibanaEphemeralResource[KibanaConnectionField, struct{}])(nil)
)
