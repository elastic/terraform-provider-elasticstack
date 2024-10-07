package models

import (
	"encoding/json"
	"time"
)

type Transform struct {
	Id              string                    `json:"id,omitempty"`
	Name            string                    `json:"-"`
	Description     string                    `json:"description,omitempty"`
	Source          *TransformSource          `json:"source"`
	Destination     *TransformDestination     `json:"dest"`
	Pivot           interface{}               `json:"pivot,omitempty"`
	Latest          interface{}               `json:"latest,omitempty"`
	Frequency       string                    `json:"frequency,omitempty"`
	RetentionPolicy *TransformRetentionPolicy `json:"retention_policy,omitempty"`
	Sync            *TransformSync            `json:"sync,omitempty"`
	Meta            interface{}               `json:"_meta,omitempty"`
	Settings        *TransformSettings        `json:"settings,omitempty"`
}

type TransformSource struct {
	Indices         []string    `json:"index"`
	Query           interface{} `json:"query,omitempty"`
	RuntimeMappings interface{} `json:"runtime_mappings,omitempty"`
}

type TransformAlias struct {
	Alias          string `json:"alias"`
	MoveOnCreation bool   `json:"move_on_creation,omitempty"`
}

type TransformDestination struct {
	Index    string           `json:"index"`
	Aliases  []TransformAlias `json:"aliases,omitempty"`
	Pipeline string           `json:"pipeline,omitempty"`
}

type TransformRetentionPolicy struct {
	Time TransformRetentionPolicyTime `json:"time"`
}

type TransformRetentionPolicyTime struct {
	Field  string `json:"field"`
	MaxAge string `json:"max_age"`
}

type TransformSync struct {
	Time TransformSyncTime `json:"time"`
}

type TransformSyncTime struct {
	Field string `json:"field"`
	Delay string `json:"delay,omitempty"`
}

type TransformSettings struct {
	AlignCheckpoints   *bool    `json:"align_checkpoints,omitempty"`
	DatesAsEpochMillis *bool    `json:"dates_as_epoch_millis,omitempty"`
	DeduceMappings     *bool    `json:"deduce_mappings,omitempty"`
	DocsPerSecond      *float64 `json:"docs_per_second,omitempty"`
	MaxPageSearchSize  *int     `json:"max_page_search_size,omitempty"`
	NumFailureRetries  *int     `json:"num_failure_retries,omitempty"`
	Unattended         *bool    `json:"unattended,omitempty"`
}

type PutTransformParams struct {
	DeferValidation bool
	Timeout         time.Duration
	Enabled         bool
}

type UpdateTransformParams struct {
	DeferValidation bool
	Timeout         time.Duration
	Enabled         bool
	ApplyEnabled    bool
}

type GetTransformResponse struct {
	Count      json.Number `json:"count"`
	Transforms []Transform `json:"transforms"`
}

type TransformStats struct {
	Id    string `json:"id"`
	State string `json:"state"`
}

type GetTransformStatsResponse struct {
	Count          json.Number      `json:"count"`
	TransformStats []TransformStats `json:"transforms"`
}

func (ts *TransformStats) IsStarted() bool {
	return ts.State == "started" || ts.State == "indexing"
}
