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

type ProcessorFingerprint struct {
	CommonProcessor

	Fields        []string `json:"fields"`
	TargetField   string   `json:"target_field,omitempty"`
	IgnoreMissing bool     `json:"ignore_missing"`
	Salt          string   `json:"salt,omitempty"`
	Method        string   `json:"method,omitempty"`
}

type ProcessorForeach struct {
	CommonProcessor

	Field         string                 `json:"field"`
	IgnoreMissing bool                   `json:"ignore_missing"`
	Processor     map[string]interface{} `json:"processor"`
}

type ProcessorGeoip struct {
	ProcessortFields

	DatabaseFile string   `json:"database_file,omitempty"`
	Properties   []string `json:"properties,omitempty"`
	FirstOnly    bool     `json:"first_only"`
}

type ProcessorGrok struct {
	CommonProcessor

	Field              string            `json:"field"`
	Patterns           []string          `json:"patterns"`
	PatternDefinitions map[string]string `json:"pattern_definitions,omitempty"`
	EcsCompatibility   string            `json:"ecs_compatibility,omitempty"`
	TraceMatch         bool              `json:"trace_match"`
	IgnoreMissing      bool              `json:"ignore_missing"`
}

type ProcessorGsub struct {
	CommonProcessor
	ProcessortFields

	Pattern     string `json:"pattern"`
	Replacement string `json:"replacement"`
}

type ProcessorHtmlStrip struct {
	CommonProcessor
	ProcessortFields
}

type ProcessorJoin struct {
	CommonProcessor

	Field       string `json:"field"`
	Separator   string `json:"separator"`
	TargetField string `json:"target_field,omitempty"`
}

type ProcessorJson struct {
	CommonProcessor

	Field                     string `json:"field"`
	TargetField               string `json:"target_field,omitempty"`
	AddToRoot                 *bool  `json:"add_to_root,omitempty"`
	AddToRootConflictStrategy string `json:"add_to_root_conflict_strategy,omitempty"`
	AllowDuplicateKeys        *bool  `json:"allow_duplicate_keys,omitempty"`
}

type ProcessorKV struct {
	CommonProcessor
	ProcessortFields

	FieldSplit    string   `json:"field_split"`
	ValueSplit    string   `json:"value_split"`
	IncludeKeys   []string `json:"include_keys,omitempty"`
	ExcludeKeys   []string `json:"exclude_keys,omitempty"`
	Prefix        string   `json:"prefix,omitempty"`
	TrimKey       string   `json:"trim_key,omitempty"`
	TrimValue     string   `json:"trim_value,omitempty"`
	StripBrackets bool     `json:"strip_brackets"`
}

type ProcessorLowercase struct {
	CommonProcessor
	ProcessortFields
}

type ProcessorNetworkDirection struct {
	CommonProcessor

	SourceIp              string   `json:"source_ip,omitempty"`
	DestinationIp         string   `json:"destination_ip,omitempty"`
	TargetField           string   `json:"target_field,omitempty"`
	InternalNetworks      []string `json:"internal_networks,omitempty"`
	InternalNetworksField string   `json:"internal_networks_field,omitempty"`
	IgnoreMissing         bool     `json:"ignore_missing"`
}

type ProcessorPipeline struct {
	CommonProcessor

	Name string `json:"name"`
}

type ProcessorRegisteredDomain struct {
	CommonProcessor
	ProcessortFields
}

type ProcessorRemove struct {
	CommonProcessor

	Field         []string `json:"field"`
	IgnoreMissing bool     `json:"ignore_missing"`
}

type ProcessorRename struct {
	CommonProcessor
	ProcessortFields
}

type ProcessorScript struct {
	CommonProcessor

	Lang     string                 `json:"lang,omitempty"`
	ScriptId string                 `json:"id,omitempty"`
	Source   string                 `json:"source,omitempty"`
	Params   map[string]interface{} `json:"params,omitempty"`
}
