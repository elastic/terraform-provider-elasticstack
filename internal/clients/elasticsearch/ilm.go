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
	"fmt"
	"io"
	"strings"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	fwdiags "github.com/hashicorp/terraform-plugin-framework/diag"
)

func PutIlm(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, policy *models.Policy) fwdiags.Diagnostics {
	policyBytes, err := json.Marshal(map[string]any{"policy": policy})
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}

	typedClient := apiClient.GetESClient()
	// Submit the marshaled policy JSON as the raw request body rather than
	// unmarshaling into putlifecycle.Request. The typed SearchableSnapshotAction
	// struct does not model force_merge_on_clone, and its UnmarshalJSON would
	// silently drop the field. Tracked upstream as
	// https://github.com/elastic/elasticsearch-specification/issues/6575
	// (terraform-provider-elasticstack#4606).
	_, err = typedClient.Ilm.PutLifecycle(policy.Name).Raw(bytes.NewReader(policyBytes)).Do(ctx)
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}
	return nil
}

// IlmPolicy is the subset of a GET _ilm/policy response the provider needs.
// Actions are left as generic maps so fields the typed client does not model
// (for example searchable_snapshot.force_merge_on_clone) survive the read path.
// Metadata is kept as raw JSON so integer values are not rounded via float64.
type IlmPolicy struct {
	ModifiedDate string
	Metadata     json.RawMessage
	Phases       map[string]IlmPhase
}

// IlmPhase is one phase of an [IlmPolicy].
type IlmPhase struct {
	MinAge  string
	Actions map[string]map[string]any
}

type ilmLifecycleResponseEntry struct {
	ModifiedDate any `json:"modified_date"`
	Policy       struct {
		Meta   json.RawMessage                      `json:"_meta"`
		Phases map[string]ilmLifecycleResponsePhase `json:"phases"`
	} `json:"policy"`
}

type ilmLifecycleResponsePhase struct {
	MinAge  any                       `json:"min_age"`
	Actions map[string]map[string]any `json:"actions"`
}

func GetIlm(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, policyName string) (*IlmPolicy, fwdiags.Diagnostics) {
	typedClient := apiClient.GetESClient()

	// Use Perform rather than Do so the response is not decoded through the
	// typed Lifecycle/SearchableSnapshotAction structs, which would silently
	// drop fields the generated client does not model (e.g. force_merge_on_clone).
	res, err := typedClient.Ilm.GetLifecycle().Policy(policyName).Perform(ctx)
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}
	defer res.Body.Close()

	if notFound, d := diagutil.CheckHTTPErrorOrNotFound(res, "Unable to fetch ILM policy"); notFound || d.HasError() {
		if notFound {
			return nil, nil
		}
		return nil, d
	}

	return decodeGetIlmResponse(policyName, res.Body)
}

func decodeGetIlmResponse(policyName string, body io.Reader) (*IlmPolicy, fwdiags.Diagnostics) {
	var decoded map[string]ilmLifecycleResponseEntry
	if err := json.NewDecoder(body).Decode(&decoded); err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}

	entry, ok := decoded[policyName]
	if !ok {
		return nil, fwdiags.Diagnostics{
			fwdiags.NewErrorDiagnostic(
				"Unable to find a ILM policy in the cluster",
				fmt.Sprintf(`Unable to find "%s" ILM policy in the cluster`, policyName),
			),
		}
	}

	out := &IlmPolicy{
		ModifiedDate: anyToString(entry.ModifiedDate),
		Metadata:     entry.Policy.Meta,
		Phases:       make(map[string]IlmPhase, len(entry.Policy.Phases)),
	}
	for name, phase := range entry.Policy.Phases {
		out.Phases[name] = IlmPhase{
			MinAge:  anyToString(phase.MinAge),
			Actions: phase.Actions,
		}
	}
	return out, nil
}

func anyToString(v any) string {
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return fmt.Sprint(v)
}

// GetIndicesWithILMPolicy returns the names of all indices currently using
// the given ILM policy.
//
// It queries GET /_ilm/policy/<policyName> and reads the
// `<policy>.in_use_by.indices` field, which Elasticsearch maintains per
// policy. This is a single targeted lookup keyed by the policy and avoids
// scanning indices cluster-wide.
//
// The typed client's generated `Lifecycle` struct does not expose
// `in_use_by`, so this function uses Perform to obtain the raw HTTP response
// and decodes the relevant subset of the body itself.
func GetIndicesWithILMPolicy(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, policyName string) ([]string, fwdiags.Diagnostics) {
	typedClient := apiClient.GetESClient()

	res, err := typedClient.Ilm.GetLifecycle().Policy(policyName).Perform(ctx)
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}
	defer res.Body.Close()

	if notFound, d := diagutil.CheckHTTPErrorOrNotFound(res, "Unable to fetch ILM policy"); notFound || d.HasError() {
		return nil, d
	}

	// The response is shaped as:
	//   { "<policy_name>": { "in_use_by": { "indices": [...], ... }, ... } }
	var decoded map[string]struct {
		InUseBy struct {
			Indices []string `json:"indices"`
		} `json:"in_use_by"`
	}
	if err := json.NewDecoder(res.Body).Decode(&decoded); err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}

	entry, ok := decoded[policyName]
	if !ok {
		return nil, nil
	}
	return entry.InUseBy.Indices, nil
}

// ClearILMPolicyFromIndices removes the ILM policy reference from the
// provided indices by setting index.lifecycle.name to null.
// It issues PUT /{indices}/_settings.
func ClearILMPolicyFromIndices(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, indices []string) fwdiags.Diagnostics {
	if len(indices) == 0 {
		return nil
	}

	settingsBytes, err := json.Marshal(map[string]any{"index.lifecycle.name": nil})
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}

	typedClient := apiClient.GetESClient()

	_, err = typedClient.Indices.PutSettings().Indices(strings.Join(indices, ",")).Raw(bytes.NewReader(settingsBytes)).Do(ctx)
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}
	return nil
}

func DeleteIlm(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, policyName string) fwdiags.Diagnostics {
	typedClient := apiClient.GetESClient()
	_, err := typedClient.Ilm.DeleteLifecycle(policyName).Do(ctx)
	return DiagsOrNotFound(err)
}
