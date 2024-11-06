package models

import (
	"time"
)

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
	AlertDelay      *float32
}

type AlertingRuleSchedule struct {
	Interval string
}

type AlertingRuleAction struct {
	Group        string
	ID           string
	Params       map[string]interface{}
	Frequency    *ActionFrequency
	AlertsFilter *ActionAlertsFilter
}

type AlertingRuleExecutionStatus struct {
	LastExecutionDate *time.Time
	Status            *string
}

type ActionFrequency struct {
	Summary    bool
	NotifyWhen string
	Throttle   *string
}

type ActionAlertsFilter struct {
	Kql       *string
	Timeframe *AlertsFilterTimeframe
}

type AlertsFilterTimeframe struct {
	Days       []int32
	Timezone   string
	HoursStart string
	HoursEnd   string
}
