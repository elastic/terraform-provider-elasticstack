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

package fleet

import (
	"context"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// OutputUnionHandlers holds one callback per kbapi.OutputUnion discriminated
// variant. Callers with different receiver types (e.g. the output resource
// and the outputs data source) supply their own per-type fromAPI* methods as
// handlers to DispatchOutputUnion, so the discriminator dispatch logic is
// defined once and both call sites are guaranteed to handle the same set of
// variants.
type OutputUnionHandlers struct {
	Elasticsearch       func(ctx context.Context, data *kbapi.KibanaHTTPAPIsOutputResponseElasticsearch) diag.Diagnostics
	Logstash            func(ctx context.Context, data *kbapi.KibanaHTTPAPIsOutputResponseLogstash) diag.Diagnostics
	Kafka               func(ctx context.Context, data *kbapi.KibanaHTTPAPIsOutputResponseKafka) diag.Diagnostics
	RemoteElasticsearch func(ctx context.Context, data *kbapi.KibanaHTTPAPIsOutputResponseRemoteElasticsearch) diag.Diagnostics
}

// DispatchOutputUnion resolves the discriminated variant held by union and
// invokes the matching handler in handlers. It appends an error diagnostic if
// union is unset, the discriminator can't be resolved, or the resolved
// variant has no matching handler.
func DispatchOutputUnion(ctx context.Context, union *kbapi.OutputUnion, handlers OutputUnionHandlers) (diags diag.Diagnostics) {
	if union == nil {
		return
	}

	output, err := union.ValueByDiscriminator()
	if err != nil {
		diags.AddError(err.Error(), "")
		return
	}

	switch output := output.(type) {
	case kbapi.KibanaHTTPAPIsOutputResponseElasticsearch:
		diags.Append(handlers.Elasticsearch(ctx, &output)...)
	case kbapi.KibanaHTTPAPIsOutputResponseLogstash:
		diags.Append(handlers.Logstash(ctx, &output)...)
	case kbapi.KibanaHTTPAPIsOutputResponseKafka:
		diags.Append(handlers.Kafka(ctx, &output)...)
	case kbapi.KibanaHTTPAPIsOutputResponseRemoteElasticsearch:
		diags.Append(handlers.RemoteElasticsearch(ctx, &output)...)
	default:
		diags.AddError(fmt.Sprintf("unhandled output type: %T", output), "")
	}

	return
}
