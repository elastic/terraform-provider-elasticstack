package models

import "time"

type AlertingRule struct {
	RuleID     string
	SpaceID    string
	Name       string
	Consumer   string
	NotifyWhen *string
	Params     map[string]interface{}
	RuleTypeID string
	Schedule   AlertingRuleSchedule
	Actions    []AlertingRuleAction
	Enabled    *bool
	Tags       []string
	Throttle   *string

	ScheduledTaskID *string
	ExecutionStatus AlertingRuleExecutionStatus
	AlertDelay      AlertingRuleAlertDelay
}

type AlertingRuleSchedule struct {
	Interval string
}

type AlertingRuleAlertDelay struct {
	Active float32
}

type AlertingRuleAction struct {
	Group     string
	ID        string
	Params    map[string]interface{}
	Frequency AlertingRuleActionFrequency
}

type AlertingRuleExecutionStatus struct {
	LastExecutionDate *time.Time
	Status            *string
}

type AlertingRuleActionFrequency struct {
	NotifyWhen string
	Summary    bool
	Throttle   *string
}
