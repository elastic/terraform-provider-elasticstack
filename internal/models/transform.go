package models

import (
	"encoding/json"
	"time"
)

type Transform struct {
	Id          string                 `json:"id,omitempty"`
	Name        string                 `json:"-"`
	Description string                 `json:"description,omitempty"`
	Source      TransformSource        `json:"source"`
	Destination TransformDestination   `json:"dest"`
	Pivot       interface{}            `json:"pivot,omitempty"`
	Latest      interface{}            `json:"latest,omitempty"`
	Frequency   string                 `json:"frequency,omitempty"`
	Meta        map[string]interface{} `json:"_meta,omitempty"`
}

type TransformSource struct {
	Indices         []string    `json:"index"`
	Query           interface{} `json:"query,omitempty"`
	RuntimeMappings interface{} `json:"runtime_mappings,omitempty"`
}

type TransformDestination struct {
	Index    string `json:"index"`
	Pipeline string `json:"pipeline,omitempty"`
}

type PutTransformParams struct {
	DeferValidation bool
	Timeout         time.Duration
}

type GetTransformResponse struct {
	Count      json.Number `json:"count,omitempty"`
	Transforms []Transform `json:"transforms"`
}
