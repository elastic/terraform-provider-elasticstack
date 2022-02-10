package models

type IngestPipeline struct {
	Name        string                   `json:"-"`
	Description *string                  `json:"description,omitempty"`
	OnFailure   []map[string]interface{} `json:"on_failure,omitempty"`
	Processors  []map[string]interface{} `json:"processors"`
	Metadata    map[string]interface{}   `json:"_meta,omitempty"`
}

type CommonProcessor struct {
	Description   string                   `json:"description,omitempty"`
	If            string                   `json:"if,omitempty"`
	IgnoreFailure bool                     `json:"ignore_failure"`
	OnFailure     []map[string]interface{} `json:"on_failure,omitempty"`
	Tag           string                   `json:"tag,omitempty"`
}

type ProcessorAppend struct {
	CommonProcessor

	Field           string   `json:"field"`
	Value           []string `json:"value"`
	AllowDuplicates bool     `json:"allow_duplicates"`
	MediaType       string   `json:"media_type"`
}

type ProcessorBytes struct {
	CommonProcessor

	Field         string `json:"field"`
	TargetField   string `json:"target_field,omitempty"`
	IgnoreMissing bool   `json:"ignore_missing"`
}
