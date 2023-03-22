package models

type AlertingRule struct {
	ID         string                 `json:"id"`
	SpaceID    string                 `json:"-"`
	Name       string                 `json:"name"`
	Consumer   string                 `json:"consumer"`
	NotifyWhen string                 `json:"notify_when"`
	Params     map[string]interface{} `json:"params"`
	RuleTypeID string                 `json:"rule_type_id"`
	Schedule   AlertingRuleSchedule   `json:"schedule"`
	Actions    []AlertingRuleAction   `json:"actions"`
	Enabled    *bool                  `json:"enabled"`
	Tags       []string               `json:"tags"`
	Throttle   *string                `json:"throttle"`

	ScheduledTaskID string                      `json:"scheduled_task_id"`
	ExecutionStatus AlertingRuleExecutionStatus `json:"execution_status"`
}

type AlertingRuleSchedule struct {
	Interval string `json:"interval"`
}

type AlertingRuleAction struct {
	Group  string                 `json:"group"`
	ID     string                 `json:"id"`
	Params map[string]interface{} `json:"actions"`
}

type AlertingRuleExecutionStatus struct {
	LastExecutionDate string `json:"last_execution_date"`
	Status            string `json:"status"`
}
