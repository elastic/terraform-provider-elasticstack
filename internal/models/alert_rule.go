package models

import "time"

type AlertingRule struct {
	RuleID     string
	SpaceID    string
	Name       string
	Consumer   string
	NotifyWhen string
	Params     map[string]interface{}
	RuleTypeID string
	Schedule   AlertingRuleSchedule
	Actions    []AlertingRuleAction
	Enabled    *bool
	Tags       []string
	Throttle   *string

	ScheduledTaskID *string
	ExecutionStatus AlertingRuleExecutionStatus
}

type AlertingRuleSchedule struct {
	Interval string
}

type AlertingRuleAction struct {
	Group  string
	ID     string
	Params map[string]interface{}
}

type AlertingRuleExecutionStatus struct {
	LastExecutionDate *time.Time
	Status            *string
}
