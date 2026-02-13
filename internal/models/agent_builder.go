package models

type Agent struct {
	ID            string              `json:"id"`
	Name          string              `json:"name"`
	Description   *string             `json:"description,omitempty"`
	AvatarColor   *string             `json:"avatar_color,omitempty"`
	AvatarSymbol  *string             `json:"avatar_symbol,omitempty"`
	Labels        []string            `json:"labels,omitempty"`
	Configuration AgentConfiguration  `json:"configuration"`
}

type AgentConfiguration struct {
	Instructions *string            `json:"instructions,omitempty"`
	Tools        []AgentToolsConfig `json:"tools,omitempty"`
}

type AgentToolsConfig struct {
	ToolIds []string `json:"tool_ids"`
}

type Tool struct {
	ID            string                 `json:"id"`
	Type          string                 `json:"type"`
	Description   *string                `json:"description,omitempty"`
	Tags          []string               `json:"tags,omitempty"`
	Configuration map[string]interface{} `json:"configuration"`
}
