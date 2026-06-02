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

package elasticsearch

import (
	"bytes"
	"context"
	"encoding/json"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	fwdiag "github.com/hashicorp/terraform-plugin-framework/diag"
)

// CreateSnapshotRequest holds the body fields for POST /_snapshot/{repository}/{snapshot}.
type CreateSnapshotRequest struct {
	Indices            []string
	IgnoreUnavailable  *bool
	IncludeGlobalState *bool
	FeatureStates      []string
	ExpandWildcards    *string
	Metadata           json.RawMessage
	Partial            *bool
}

// CreateSnapshot invokes the Elasticsearch snapshot create API.
func CreateSnapshot(ctx context.Context, client *clients.ElasticsearchScopedClient, repo, snapshot string, body *CreateSnapshotRequest, waitForCompletion bool) fwdiag.Diagnostics {
	typedClient := client.GetESClient()

	req := typedClient.Snapshot.Create(repo, snapshot).WaitForCompletion(waitForCompletion)

	if body != nil {
		payload := map[string]any{}
		if len(body.Indices) > 0 {
			payload["indices"] = body.Indices
		}
		if body.IgnoreUnavailable != nil {
			payload["ignore_unavailable"] = *body.IgnoreUnavailable
		}
		if body.IncludeGlobalState != nil {
			payload["include_global_state"] = *body.IncludeGlobalState
		}
		if len(body.FeatureStates) > 0 {
			payload["feature_states"] = body.FeatureStates
		}
		if body.ExpandWildcards != nil {
			payload["expand_wildcards"] = *body.ExpandWildcards
		}
		if body.Partial != nil {
			payload["partial"] = *body.Partial
		}
		if len(body.Metadata) > 0 {
			var metadata map[string]any
			if err := json.Unmarshal(body.Metadata, &metadata); err != nil {
				return diagutil.FrameworkDiagFromError(err)
			}
			payload["metadata"] = metadata
		}

		if len(payload) > 0 {
			bodyBytes, err := json.Marshal(payload)
			if err != nil {
				return diagutil.FrameworkDiagFromError(err)
			}
			req.Raw(bytes.NewReader(bodyBytes))
		}
	}

	_, err := req.Do(ctx)
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}
	return nil
}
