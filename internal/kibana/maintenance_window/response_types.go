package maintenance_window

/*
* The types generated automatically for kibana_oapi are deeply nested a very hard to use.
* This file defines convenience types that can be used to define these neestes objects
* when needed.
 */

type ResponseJson struct {
	CreatedAt string               `json:"created_at"`
	CreatedBy *string              `json:"created_by"`
	Enabled   bool                 `json:"enabled"`
	Id        string               `json:"id"`
	Schedule  ResponseJsonSchedule `json:"schedule"`
	Scope     *ResponseJsonScope   `json:"scope,omitempty"`
	Title     string               `json:"title"`
}

type ResponseJsonSchedule struct {
	Custom ResponseJsonCustomSchedule `json:"custom"`
}

type ResponseJsonCustomSchedule struct {
	Duration  string                 `json:"duration"`
	Recurring *ResponseJsonRecurring `json:"recurring,omitempty"`
	Start     string                 `json:"start"`
	Timezone  *string                `json:"timezone,omitempty"`
}

type ResponseJsonRecurring struct {
	End         *string    `json:"end,omitempty"`
	Every       *string    `json:"every,omitempty"`
	Occurrences *float32   `json:"occurrences,omitempty"`
	OnMonth     *[]float32 `json:"onMonth,omitempty"`
	OnMonthDay  *[]float32 `json:"onMonthDay,omitempty"`
	OnWeekDay   *[]string  `json:"onWeekDay,omitempty"`
}

type ResponseJsonScope struct {
	Alerting ResponseJsonAlerting `json:"alerting"`
}

type ResponseJsonAlerting struct {
	Query ResponseJsonAlertingQuery `json:"query"`
}

type ResponseJsonAlertingQuery struct {
	Kql string `json:"kql"`
}
