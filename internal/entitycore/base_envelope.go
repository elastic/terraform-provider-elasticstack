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
	"maps"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

// baseResourceEnvelope holds common wiring shared by KibanaResource and
// ElasticsearchResource: the provider client base, schema factory, and the
// connection block that each envelope injects. hasReadFunc records whether a
// read callback was supplied so requireReadFunc can return a diagnostic instead
// of panicking.
type baseResourceEnvelope struct {
	*ResourceBase
	schemaFactory   func(context.Context) rschema.Schema
	connectionKey   string
	connectionBlock rschema.Block
	hasReadFunc     bool
}

// Schema implements [resource.Resource], injecting the connection block into
// the schema returned by the concrete schema factory.
func (b *baseResourceEnvelope) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	schema := b.schemaFactory(ctx)
	blocks := make(map[string]rschema.Block, len(schema.Blocks)+1)
	maps.Copy(blocks, schema.Blocks)
	blocks[b.connectionKey] = b.connectionBlock
	schema.Blocks = blocks
	resp.Schema = schema
}

// requireReadFunc returns an error diagnostic when no read callback was
// supplied at construction time.
func (b *baseResourceEnvelope) requireReadFunc() diag.Diagnostics {
	if !b.hasReadFunc {
		return requireReadFuncDiag(b.component)
	}
	return nil
}

// applyReadFoundResult handles the terminal branch of a Read call: when the
// resource is found, postRead (if non-nil) is invoked and the result is
// persisted; when not found the resource is removed from state.
func applyReadFoundResult[T any](
	ctx context.Context,
	resp *resource.ReadResponse,
	found bool,
	result T,
	postRead func(T) (T, diag.Diagnostics),
) {
	if found {
		if postRead != nil {
			var prDiags diag.Diagnostics
			result, prDiags = postRead(result)
			resp.Diagnostics.Append(prDiags...)
			if resp.Diagnostics.HasError() {
				return
			}
		}
		resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
	} else {
		resp.State.RemoveResource(ctx)
	}
}
