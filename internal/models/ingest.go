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

type ProcessortFields struct {
	Field         string `json:"field"`
	TargetField   string `json:"target_field,omitempty"`
	IgnoreMissing bool   `json:"ignore_missing"`
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
	ProcessortFields
}

type ProcessorCircle struct {
	CommonProcessor
	ProcessortFields

	ErrorDistance float64 `json:"error_distance"`
	ShapeType     string  `json:"shape_type"`
}

type ProcessorCommunityId struct {
	CommonProcessor

	SourceIp        string `json:"source_ip,omitempty"`
	SourcePort      *int   `json:"source_port,omitempty"`
	DestinationIp   string `json:"destination_ip,omitempty"`
	DestinationPort *int   `json:"destination_port,omitempty"`
	IanaNumber      string `json:"iana_number,omitempty"`
	IcmpType        *int   `json:"icmp_type,omitempty"`
	IcmpCode        *int   `json:"icmp_code,omitempty"`
	Transport       string `json:"transport,omitempty"`
	TargetField     string `json:"target_field,omitempty"`
	Seed            *int   `json:"seed"`
	IgnoreMissing   bool   `json:"ignore_missing"`
}

type ProcessorConvert struct {
	CommonProcessor
	ProcessortFields

	Type string `json:"type"`
}

type ProcessorCSV struct {
	CommonProcessor

	Field         string   `json:"field"`
	TargetFields  []string `json:"target_fields"`
	IgnoreMissing bool     `json:"ignore_missing"`
	Separator     string   `json:"separator"`
	Quote         string   `json:"quote"`
	Trim          bool     `json:"trim"`
	EmptyValue    string   `json:"empty_value,omitempty"`
}

type ProcessorDate struct {
	CommonProcessor

	Field        string   `json:"field"`
	TargetField  string   `json:"target_field,omitempty"`
	Formats      []string `json:"formats"`
	Timezone     string   `json:"timezone,omitempty"`
	Locale       string   `json:"locale,omitempty"`
	OutputFormat string   `json:"output_format,omitempty"`
}

type ProcessorDateIndexName struct {
	CommonProcessor

	Field           string   `json:"field"`
	IndexNamePrefix string   `json:"index_name_prefix,omitempty"`
	DateRounding    string   `json:"date_rounding"`
	DateFormats     []string `json:"date_formats,omitempty"`
	Timezone        string   `json:"timezone,omitempty"`
	Locale          string   `json:"locale,omitempty"`
	IndexNameFormat string   `json:"index_name_format,omitempty"`
}

type ProcessorDissect struct {
	CommonProcessor

	Field           string `json:"field"`
	Pattern         string `json:"pattern"`
	AppendSeparator string `json:"append_separator"`
	IgnoreMissing   bool   `json:"ignore_missing"`
}

type ProcessorDotExpander struct {
	CommonProcessor

	Field    string `json:"field"`
	Path     string `json:"path,omitempty"`
	Override bool   `json:"override"`
}

type ProcessorDrop struct {
	CommonProcessor
}

type ProcessorEnrich struct {
	CommonProcessor
	ProcessortFields

	PolicyName    string `json:"policy_name"`
	Override      bool   `json:"override"`
	MaxMatches    int    `json:"max_matches"`
	ShapeRelation string `json:"shape_relation,omitempty"`
}

type ProcessorFail struct {
	CommonProcessor

	Message string `json:"message"`
}
