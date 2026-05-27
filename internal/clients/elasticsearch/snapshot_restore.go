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

// RestoreSnapshotRequest holds the body fields for POST /_snapshot/{repository}/{snapshot}/_restore.
type RestoreSnapshotRequest struct {
	Indices             []string
	IgnoreUnavailable   *bool
	IncludeGlobalState  *bool
	IncludeAliases      *bool
	FeatureStates         []string
	RenamePattern       *string
	RenameReplacement   *string
	Partial               *bool
	IndexSettings         json.RawMessage
	IgnoreIndexSettings []string
}

// RestoreSnapshot invokes the Elasticsearch snapshot restore API.
func RestoreSnapshot(ctx context.Context, client *clients.ElasticsearchScopedClient, repo, snapshot string, body *RestoreSnapshotRequest, waitForCompletion bool) fwdiag.Diagnostics {
	typedClient := client.GetESClient()

	req := typedClient.Snapshot.Restore(repo, snapshot).WaitForCompletion(waitForCompletion)

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
		if body.IncludeAliases != nil {
			payload["include_aliases"] = *body.IncludeAliases
		}
		if len(body.FeatureStates) > 0 {
			payload["feature_states"] = body.FeatureStates
		}
		if body.RenamePattern != nil {
			payload["rename_pattern"] = *body.RenamePattern
		}
		if body.RenameReplacement != nil {
			payload["rename_replacement"] = *body.RenameReplacement
		}
		if body.Partial != nil {
			payload["partial"] = *body.Partial
		}
		if len(body.IgnoreIndexSettings) > 0 {
			payload["ignore_index_settings"] = body.IgnoreIndexSettings
		}
		if len(body.IndexSettings) > 0 {
			var indexSettings map[string]any
			if err := json.Unmarshal(body.IndexSettings, &indexSettings); err != nil {
				return diagutil.FrameworkDiagFromError(err)
			}
			payload["index_settings"] = indexSettings
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
