package models

type EnrichPolicy struct {
	Type         string   `json:policy_type`
	Name         string   `json:name`
	Indices      []string `json:"indices"`
	MatchField   string   `json:"match_field"`
	EnrichFields []string `json:"enrich_fields"`
	Query        string   `json:query,omitempty`
}
