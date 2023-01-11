package models

type User struct {
	Username     string                 `json:"-"`
	FullName     string                 `json:"full_name,omitempty"`
	Email        string                 `json:"email,omitempty"`
	Roles        []string               `json:"roles"`
	Password     *string                `json:"password,omitempty"`
	PasswordHash *string                `json:"password_hash,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	Enabled      bool                   `json:"enabled"`
}

func (u *User) IsSystemUser() bool {
	if reserved := u.Metadata["_reserved"]; reserved != nil {
		isReserved, ok := reserved.(bool)
		return ok && isReserved
	}
	return false
}

type UserPassword struct {
	Password     *string `json:"password,omitempty"`
	PasswordHash *string `json:"password_hash,omitempty"`
}

type Role struct {
	Name         string                 `json:"-"`
	Applications []Application          `json:"applications,omitempty"`
	Global       map[string]interface{} `json:"global,omitempty"`
	Cluster      []string               `json:"cluster,omitempty"`
	Indices      []IndexPerms           `json:"indices,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	RusAs        []string               `json:"run_as,omitempty"`
}

type RoleMapping struct {
	Name          string                   `json:"-"`
	Enabled       bool                     `json:"enabled"`
	Roles         []string                 `json:"roles,omitempty"`
	RoleTemplates []map[string]interface{} `json:"role_templates,omitempty"`
	Rules         map[string]interface{}   `json:"rules"`
	Metadata      interface{}              `json:"metadata"`
}

type ApiKey struct {
	Name             string                 `json:"name"`
	RolesDescriptors map[string]Role        `json:"role_descriptors,omitempty"`
	Expiration       string                 `json:"expiration,omitempty"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
}

type ApiKeyResponse struct {
	ApiKey
	RolesDescriptors map[string]Role `json:"role_descriptors,omitempty"`
	Expiration       int64           `json:"expiration,omitempty"`
	Id               string          `json:"id,omitempty"`
	Key              string          `json:"api_key,omitempty"`
	EncodedKey       string          `json:"encoded,omitempty"`
	Invalidated      bool            `json:"invalidated,omitempty"`
}

type IndexPerms struct {
	FieldSecurity          *FieldSecurity `json:"field_security,omitempty"`
	Names                  []string       `json:"names"`
	Privileges             []string       `json:"privileges"`
	Query                  *string        `json:"query,omitempty"`
	AllowRestrictedIndices *bool          `json:"allow_restricted_indices,omitempty"`
}

type FieldSecurity struct {
	Grant  []string `json:"grant,omitempty"`
	Except []string `json:"except,omitempty"`
}

type Application struct {
	Name       string   `json:"application"`
	Privileges []string `json:"privileges,omitempty"`
	Resources  []string `json:"resources"`
}

type IndexTemplate struct {
	Name          string                 `json:"-"`
	Create        bool                   `json:"-"`
	Timeout       string                 `json:"-"`
	ComposedOf    []string               `json:"composed_of"`
	DataStream    *DataStreamSettings    `json:"data_stream,omitempty"`
	IndexPatterns []string               `json:"index_patterns"`
	Meta          map[string]interface{} `json:"_meta,omitempty"`
	Priority      *int                   `json:"priority,omitempty"`
	Template      *Template              `json:"template,omitempty"`
	Version       *int                   `json:"version,omitempty"`
}

type DataStreamSettings struct {
	Hidden             *bool `json:"hidden,omitempty"`
	AllowCustomRouting *bool `json:"allow_custom_routing,omitempty"`
}

type Template struct {
	Aliases  map[string]IndexAlias  `json:"aliases,omitempty"`
	Mappings map[string]interface{} `json:"mappings,omitempty"`
	Settings map[string]interface{} `json:"settings,omitempty"`
}

type IndexTemplatesResponse struct {
	IndexTemplates []IndexTemplateResponse `json:"index_templates"`
}

type IndexTemplateResponse struct {
	Name          string        `json:"name"`
	IndexTemplate IndexTemplate `json:"index_template"`
}

type ComponentTemplate struct {
	Name     string                 `json:"-"`
	Meta     map[string]interface{} `json:"_meta,omitempty"`
	Template *Template              `json:"template,omitempty"`
	Version  *int                   `json:"version,omitempty"`
}

type ComponentTemplatesResponse struct {
	ComponentTemplates []ComponentTemplateResponse `json:"component_templates"`
}

type ComponentTemplateResponse struct {
	Name              string            `json:"name"`
	ComponentTemplate ComponentTemplate `json:"component_template"`
}

type PolicyDefinition struct {
	Policy   Policy `json:"policy"`
	Modified string `json:"modified_date"`
}

type Policy struct {
	Name     string                 `json:"-"`
	Metadata map[string]interface{} `json:"_meta,omitempty"`
	Phases   map[string]Phase       `json:"phases"`
}

type Phase struct {
	MinAge  string            `json:"min_age,omitempty"`
	Actions map[string]Action `json:"actions"`
}

type Action map[string]interface{}

type SnapshotRepository struct {
	Name     string                 `json:"-"`
	Type     string                 `json:"type"`
	Settings map[string]interface{} `json:"settings"`
	Verify   bool                   `json:"verify"`
}

type SnapshotPolicy struct {
	Id         string                `json:"-"`
	Config     *SnapshotPolicyConfig `json:"config,omitempty"`
	Name       string                `json:"name"`
	Repository string                `json:"repository"`
	Retention  *SnapshortRetention   `json:"retention,omitempty"`
	Schedule   string                `json:"schedule"`
}

type SnapshortRetention struct {
	ExpireAfter *string `json:"expire_after,omitempty"`
	MaxCount    *int    `json:"max_count,omitempty"`
	MinCount    *int    `json:"min_count,omitempty"`
}

type SnapshotPolicyConfig struct {
	ExpandWildcards    *string                `json:"expand_wildcards,omitempty"`
	IgnoreUnavailable  *bool                  `json:"ignore_unavailable,omitempty"`
	IncludeGlobalState *bool                  `json:"include_global_state,omitempty"`
	Indices            []string               `json:"indices,omitempty"`
	FeatureStates      []string               `json:"feature_states,omitempty"`
	Metadata           map[string]interface{} `json:"metadata,omitempty"`
	Partial            *bool                  `json:"partial,omitempty"`
}

type Index struct {
	Name     string                 `json:"-"`
	Aliases  map[string]IndexAlias  `json:"aliases,omitempty"`
	Mappings map[string]interface{} `json:"mappings,omitempty"`
	Settings map[string]interface{} `json:"settings,omitempty"`
}

type IndexAlias struct {
	Name          string                 `json:"-"`
	Filter        map[string]interface{} `json:"filter,omitempty"`
	IndexRouting  string                 `json:"index_routing,omitempty"`
	IsHidden      bool                   `json:"is_hidden,omitempty"`
	IsWriteIndex  bool                   `json:"is_write_index,omitempty"`
	Routing       string                 `json:"routing,omitempty"`
	SearchRouting string                 `json:"search_routing,omitempty"`
}

type DataStream struct {
	Name           string                 `json:"name"`
	TimestampField TimestampField         `json:"timestamp_field"`
	Indices        []DataStreamIndex      `json:"indices"`
	Generation     uint64                 `json:"generation"`
	Meta           map[string]interface{} `json:"_meta"`
	Status         string                 `json:"status"`
	Template       string                 `json:"template"`
	IlmPolicy      string                 `json:"ilm_policy"`
	Hidden         bool                   `json:"hidden"`
	System         bool                   `json:"system"`
	Replicated     bool                   `json:"replicated"`
}

type DataStreamIndex struct {
	IndexName string `json:"index_name"`
	IndexUUID string `json:"index_uuid"`
}

type TimestampField struct {
	Name string `json:"name"`
}

type LogstashPipeline struct {
	PipelineID       string                 `json:"-"`
	Description      string                 `json:"description,omitempty"`
	LastModified     string                 `json:"last_modified"`
	Pipeline         string                 `json:"pipeline"`
	PipelineMetadata map[string]interface{} `json:"pipeline_metadata"`
	PipelineSettings map[string]interface{} `json:"pipeline_settings"`
	Username         string                 `json:"username"`
}

type Script struct {
	ID       string                 `json:"-"`
	Language string                 `json:"lang"`
	Source   string                 `json:"source"`
	Params   map[string]interface{} `json:"params"`
	Context  string                 `json:"-"`
}
