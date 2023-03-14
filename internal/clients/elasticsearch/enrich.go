package elasticsearch

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"net/http"
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
	return "", fmt.Errorf("Did not find expected policy type.")
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
	if diags := utils.CheckError(res, fmt.Sprintf("Unable to get requested EnrichPolicy: %s", policyName)); diags.HasError() {
		return nil, diags
	}

	var policies enrichPoliciesResponse
	if err := json.NewDecoder(res.Body).Decode(&policies); err != nil {
		return nil, diag.FromErr(err)
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
