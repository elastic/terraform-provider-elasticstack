package elasticsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
)

type enrichPolicyResponse struct {
	Name         string         `json:"name"`
	Indices      []string       `json:"indices"`
	MatchField   string         `json:"match_field"`
	EnrichFields []string       `json:"enrich_fields"`
	Query        map[string]any `json:"query,omitempty"`
}

type enrichPoliciesResponse struct {
	Policies []struct {
		Config map[string]enrichPolicyResponse `json:"config"`
	} `json:"policies"`
}

var policyTypes = []string{"range", "match", "geo_match"}

func getPolicyType(m map[string]enrichPolicyResponse) (string, error) {
	for _, policyType := range policyTypes {
		if _, ok := m[policyType]; ok {
			return policyType, nil
		}
	}
	return "", fmt.Errorf("did not find expected policy type")
}

func GetEnrichPolicy(ctx context.Context, apiClient *clients.ApiClient, policyName string) (*models.EnrichPolicy, diag.Diagnostics) {
	var diags diag.Diagnostics
	esClient, err := apiClient.GetESClient()
	if err != nil {
		return nil, diag.FromErr(err)
	}
	req := esClient.EnrichGetPolicy.WithName(policyName)
	res, err := esClient.EnrichGetPolicy(req, esClient.EnrichGetPolicy.WithContext(ctx))
	if err != nil {
		return nil, diag.FromErr(err)
	}
	defer res.Body.Close()
	if res.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if diags := diagutil.CheckError(res, fmt.Sprintf("Unable to get requested EnrichPolicy: %s", policyName)); diags.HasError() {
		return nil, diags
	}

	var policies enrichPoliciesResponse
	if err := json.NewDecoder(res.Body).Decode(&policies); err != nil {
		return nil, diag.FromErr(err)
	}

	if len(policies.Policies) == 0 {
		return nil, diags
	}

	if len(policies.Policies) > 1 {
		tflog.Warn(ctx, fmt.Sprintf(`Somehow found more than one policy for policy named %s`, policyName))
	}
	config := policies.Policies[0].Config
	policyType, err := getPolicyType(config)
	if err != nil {
		return nil, diag.FromErr(err)
	}
	policy := config[policyType]
	queryJSON, err := json.Marshal(policy.Query)
	if err != nil {
		return nil, diag.FromErr(err)
	}

	return &models.EnrichPolicy{
		Type:         policyType,
		Name:         policy.Name,
		Indices:      policy.Indices,
		MatchField:   policy.MatchField,
		EnrichFields: policy.EnrichFields,
		Query:        string(queryJSON),
	}, diags
}

func tryJSONUnmarshalString(s string) (any, bool) {
	var data any
	if err := json.Unmarshal([]byte(s), &data); err != nil {
		return s, false
	}
	return data, true
}

func PutEnrichPolicy(ctx context.Context, apiClient *clients.ApiClient, policy *models.EnrichPolicy) diag.Diagnostics {
	var diags diag.Diagnostics
	payloadPolicy := map[string]any{
		"indices":       policy.Indices,
		"enrich_fields": policy.EnrichFields,
		"match_field":   policy.MatchField,
	}

	if query, ok := tryJSONUnmarshalString(policy.Query); ok {
		payloadPolicy["query"] = query
	} else if policy.Query != "" {
		tflog.Error(ctx, fmt.Sprintf("JAW: query did not unmarshall %s", policy.Query))
	}

	payload := map[string]any{}
	payload[policy.Type] = payloadPolicy
	policyBytes, err := json.Marshal(payload)
	if err != nil {
		return diag.FromErr(err)
	}

	esClient, err := apiClient.GetESClient()
	if err != nil {
		return diag.FromErr(err)
	}

	res, err := esClient.EnrichPutPolicy(policy.Name, bytes.NewReader(policyBytes), esClient.EnrichPutPolicy.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}
	defer res.Body.Close()

	if diags := diagutil.CheckError(res, "Unable to create enrich policy"); diags.HasError() {
		return diags
	}
	return diags
}

func DeleteEnrichPolicy(ctx context.Context, apiClient *clients.ApiClient, policyName string) diag.Diagnostics {
	var diags diag.Diagnostics

	esClient, err := apiClient.GetESClient()
	if err != nil {
		return diag.FromErr(err)
	}
	res, err := esClient.EnrichDeletePolicy(policyName, esClient.EnrichDeletePolicy.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}
	defer res.Body.Close()
	if diags := diagutil.CheckError(res, fmt.Sprintf("Unable to delete enrich policy: %s", policyName)); diags.HasError() {
		return diags
	}

	return diags
}

func ExecuteEnrichPolicy(ctx context.Context, apiClient *clients.ApiClient, policyName string) diag.Diagnostics {
	var diags diag.Diagnostics
	esClient, err := apiClient.GetESClient()
	if err != nil {
		return diag.FromErr(err)
	}
	res, err := esClient.EnrichExecutePolicy(
		policyName, esClient.EnrichExecutePolicy.WithContext(ctx), esClient.EnrichExecutePolicy.WithWaitForCompletion(true),
	)
	if err != nil {
		return diag.FromErr(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return diag.Errorf(`Executing policy "%s" failed with http status %d`, policyName, res.StatusCode)
	}
	var response struct {
		Status struct {
			Phase string `json:"phase"`
		} `json:"status"`
	}
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return diag.FromErr(err)
	}
	if response.Status.Phase != "COMPLETE" {
		return diag.Errorf(`Unexpected response to executing enrich policy: %s`, response.Status.Phase)
	}
	return diags
}
