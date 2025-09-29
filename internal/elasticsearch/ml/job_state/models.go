package job_state

import (
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type MLJobStateData struct {
	Id                      types.String         `tfsdk:"id"`
	ElasticsearchConnection types.List           `tfsdk:"elasticsearch_connection"`
	JobId                   types.String         `tfsdk:"job_id"`
	State                   types.String         `tfsdk:"state"`
	Force                   types.Bool           `tfsdk:"force"`
	Timeout                 customtypes.Duration `tfsdk:"job_timeout"`
	Timeouts                timeouts.Value       `tfsdk:"timeouts"`
}

// MLJobStats represents the statistics structure for an ML job
type MLJobStats struct {
	Jobs []MLJob `json:"jobs"`
}

// MLJob represents a single ML job in the stats response
type MLJob struct {
	JobId string     `json:"job_id"`
	State string     `json:"state"`
	Node  *MLJobNode `json:"node,omitempty"`
}

// MLJobNode represents the node information for an ML job
type MLJobNode struct {
	Id         string                 `json:"id"`
	Name       string                 `json:"name"`
	Attributes map[string]interface{} `json:"attributes"`
}
