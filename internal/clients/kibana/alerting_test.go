package kibana

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/stretchr/testify/require"
)

func Test_ruleResponseToModel(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	nowStr := now.Format(time.RFC3339)

	tests := []struct {
		name          string
		spaceId       string
		ruleResponse  *kbapi.GetAlertingRuleIdResponse
		expectedModel *models.AlertingRule
	}{
		{
			name:          "nil response should return a nil model",
			spaceId:       "space-id",
			ruleResponse:  nil,
			expectedModel: nil,
		},
		{
			name:    "nil optional fields should not blow up the transform",
			spaceId: "space-id",
			ruleResponse: func() *kbapi.GetAlertingRuleIdResponse {
				jsonData := `{
					"id": "id",
					"name": "name",
					"consumer": "consumer",
					"params": {},
					"rule_type_id": "rule-type-id",
					"enabled": true,
					"tags": ["hello"],
					"actions": [],
					"execution_status": {
						"last_execution_date": "` + nowStr + `",
						"status": "ok"
					},
					"schedule": {
						"interval": "1m"
					},
					"created_at": "` + nowStr + `",
					"updated_at": "` + nowStr + `",
					"mute_all": false,
					"muted_alert_ids": [],
					"revision": 0
				}`
				var resp kbapi.GetAlertingRuleIdResponse
				resp.JSON200 = new(struct {
					Actions []struct {
						AlertsFilter *struct {
							Query *struct {
								Dsl     *string `json:"dsl,omitempty"`
								Filters []struct {
									State *struct {
										Store kbapi.GetAlertingRuleId200ActionsAlertsFilterQueryFiltersStateStore `json:"store"`
									} `json:"$state,omitempty"`
									Meta  map[string]interface{}  `json:"meta"`
									Query *map[string]interface{} `json:"query,omitempty"`
								} `json:"filters"`
								Kql string `json:"kql"`
							} `json:"query,omitempty"`
							Timeframe *struct {
								Days  []kbapi.GetAlertingRuleId200ActionsAlertsFilterTimeframeDays `json:"days"`
								Hours struct {
									End   string `json:"end"`
									Start string `json:"start"`
								} `json:"hours"`
								Timezone string `json:"timezone"`
							} `json:"timeframe,omitempty"`
						} `json:"alerts_filter,omitempty"`
						ConnectorTypeId string `json:"connector_type_id"`
						Frequency       *struct {
							NotifyWhen kbapi.GetAlertingRuleId200ActionsFrequencyNotifyWhen `json:"notify_when"`
							Summary    bool                                                 `json:"summary"`
							Throttle   *string                                              `json:"throttle"`
						} `json:"frequency,omitempty"`
						Group                   *string                `json:"group,omitempty"`
						Id                      string                 `json:"id"`
						Params                  map[string]interface{} `json:"params"`
						UseAlertDataForTemplate *bool                  `json:"use_alert_data_for_template,omitempty"`
						Uuid                    *string                `json:"uuid,omitempty"`
					} `json:"actions"`
					ActiveSnoozes *[]string `json:"active_snoozes,omitempty"`
					AlertDelay    *struct {
						Active float32 `json:"active"`
					} `json:"alert_delay,omitempty"`
					ApiKeyCreatedByUser *bool   `json:"api_key_created_by_user"`
					ApiKeyOwner         *string `json:"api_key_owner"`
					Artifacts           *struct {
						Dashboards *[]struct {
							Id string `json:"id"`
						} `json:"dashboards,omitempty"`
						InvestigationGuide *struct {
							Blob string `json:"blob"`
						} `json:"investigation_guide,omitempty"`
					} `json:"artifacts,omitempty"`
					Consumer        string  `json:"consumer"`
					CreatedAt       string  `json:"created_at"`
					CreatedBy       *string `json:"created_by"`
					Enabled         bool    `json:"enabled"`
					ExecutionStatus struct {
						Error *struct {
							Message string                                               `json:"message"`
							Reason  kbapi.GetAlertingRuleId200ExecutionStatusErrorReason `json:"reason"`
						} `json:"error,omitempty"`
						LastDuration      *float32                                        `json:"last_duration,omitempty"`
						LastExecutionDate string                                          `json:"last_execution_date"`
						Status            kbapi.GetAlertingRuleId200ExecutionStatusStatus `json:"status"`
						Warning           *struct {
							Message string                                                 `json:"message"`
							Reason  kbapi.GetAlertingRuleId200ExecutionStatusWarningReason `json:"reason"`
						} `json:"warning,omitempty"`
					} `json:"execution_status"`
					Flapping *struct {
						Enabled               *bool   `json:"enabled,omitempty"`
						LookBackWindow        float32 `json:"look_back_window"`
						StatusChangeThreshold float32 `json:"status_change_threshold"`
					} `json:"flapping"`
					Id             string  `json:"id"`
					IsSnoozedUntil *string `json:"is_snoozed_until"`
					LastRun        *struct {
						AlertsCount struct {
							Active    *float32 `json:"active"`
							Ignored   *float32 `json:"ignored"`
							New       *float32 `json:"new"`
							Recovered *float32 `json:"recovered"`
						} `json:"alerts_count"`
						Outcome      kbapi.GetAlertingRuleId200LastRunOutcome  `json:"outcome"`
						OutcomeMsg   *[]string                                 `json:"outcome_msg"`
						OutcomeOrder *float32                                  `json:"outcome_order,omitempty"`
						Warning      *kbapi.GetAlertingRuleId200LastRunWarning `json:"warning"`
					} `json:"last_run"`
					MappedParams *map[string]interface{} `json:"mapped_params,omitempty"`
					Monitoring   *struct {
						Run struct {
							CalculatedMetrics struct {
								P50          *float32 `json:"p50,omitempty"`
								P95          *float32 `json:"p95,omitempty"`
								P99          *float32 `json:"p99,omitempty"`
								SuccessRatio float32  `json:"success_ratio"`
							} `json:"calculated_metrics"`
							History []struct {
								Duration  *float32                                               `json:"duration,omitempty"`
								Outcome   *kbapi.GetAlertingRuleId200MonitoringRunHistoryOutcome `json:"outcome,omitempty"`
								Success   bool                                                   `json:"success"`
								Timestamp float32                                                `json:"timestamp"`
							} `json:"history"`
							LastRun struct {
								Metrics struct {
									Duration     *float32 `json:"duration,omitempty"`
									GapDurationS *float32 `json:"gap_duration_s"`
									GapRange     *struct {
										Gte string `json:"gte"`
										Lte string `json:"lte"`
									} `json:"gap_range"`
									TotalAlertsCreated      *float32 `json:"total_alerts_created"`
									TotalAlertsDetected     *float32 `json:"total_alerts_detected"`
									TotalIndexingDurationMs *float32 `json:"total_indexing_duration_ms"`
									TotalSearchDurationMs   *float32 `json:"total_search_duration_ms"`
								} `json:"metrics"`
								Timestamp string `json:"timestamp"`
							} `json:"last_run"`
						} `json:"run"`
					} `json:"monitoring,omitempty"`
					MuteAll       bool                                  `json:"mute_all"`
					MutedAlertIds []string                              `json:"muted_alert_ids"`
					Name          string                                `json:"name"`
					NextRun       *string                               `json:"next_run"`
					NotifyWhen    *kbapi.GetAlertingRuleId200NotifyWhen `json:"notify_when"`
					Params        map[string]interface{}                `json:"params"`
					Revision      float32                               `json:"revision"`
					RuleTypeId    string                                `json:"rule_type_id"`
					Running       *bool                                 `json:"running"`
					Schedule      struct {
						Interval string `json:"interval"`
					} `json:"schedule"`
					ScheduledTaskId *string `json:"scheduled_task_id,omitempty"`
					SnoozeSchedule  *[]struct {
						Duration float32 `json:"duration"`
						Id       *string `json:"id,omitempty"`
						RRule    struct {
							Byhour     *[]float32                                                         `json:"byhour"`
							Byminute   *[]float32                                                         `json:"byminute"`
							Bymonth    *[]float32                                                         `json:"bymonth"`
							Bymonthday *[]float32                                                         `json:"bymonthday"`
							Bysecond   *[]float32                                                         `json:"bysecond"`
							Bysetpos   *[]float32                                                         `json:"bysetpos"`
							Byweekday  *[]kbapi.GetAlertingRuleId_200_SnoozeSchedule_RRule_Byweekday_Item `json:"byweekday"`
							Byweekno   *[]float32                                                         `json:"byweekno"`
							Byyearday  *[]float32                                                         `json:"byyearday"`
							Count      *float32                                                           `json:"count,omitempty"`
							Dtstart    string                                                             `json:"dtstart"`
							Freq       *kbapi.GetAlertingRuleId200SnoozeScheduleRRuleFreq                 `json:"freq,omitempty"`
							Interval   *float32                                                           `json:"interval,omitempty"`
							Tzid       string                                                             `json:"tzid"`
							Until      *string                                                            `json:"until,omitempty"`
							Wkst       *kbapi.GetAlertingRuleId200SnoozeScheduleRRuleWkst                 `json:"wkst,omitempty"`
						} `json:"rRule"`
						SkipRecurrences *[]string `json:"skipRecurrences,omitempty"`
					} `json:"snooze_schedule,omitempty"`
					Tags                 []string `json:"tags"`
					Throttle             *string  `json:"throttle"`
					UpdatedAt            string   `json:"updated_at"`
					UpdatedBy            *string  `json:"updated_by"`
					ViewInAppRelativeUrl *string  `json:"view_in_app_relative_url"`
				})
				err := json.Unmarshal([]byte(jsonData), resp.JSON200)
				if err != nil {
					panic(err)
				}
				return &resp
			}(),
			expectedModel: &models.AlertingRule{
				RuleID:     "id",
				SpaceID:    "space-id",
				Name:       "name",
				Consumer:   "consumer",
				Params:     map[string]interface{}{},
				RuleTypeID: "rule-type-id",
				Enabled:    utils.Pointer(true),
				Tags:       []string{"hello"},
				Schedule:   models.AlertingRuleSchedule{Interval: "1m"},
				Actions:    []models.AlertingRuleAction{},
				ExecutionStatus: models.AlertingRuleExecutionStatus{
					LastExecutionDate: &now,
					Status:            utils.Pointer("ok"),
				},
			},
		},
		{
			name:    "a full response should be successfully transformed",
			spaceId: "space-id",
			ruleResponse: func() *kbapi.GetAlertingRuleIdResponse {
				jsonData := `{
					"id": "id",
					"name": "name",
					"consumer": "consumer",
					"params": {},
					"rule_type_id": "rule-type-id",
					"enabled": true,
					"tags": ["hello"],
					"notify_when": "broken",
					"actions": [
						{
							"group": "group-1",
							"id": "id",
							"params": {},
							"connector_type_id": "connector-type",
							"frequency": {
								"summary": true,
								"notify_when": "onThrottleInterval",
								"throttle": "10s"
							},
							"alerts_filter": {
								"query": {
									"kql": "foobar"
								},
								"timeframe": {
									"days": [3, 5, 7],
									"timezone": "UTC+1",
									"hours": {
										"start": "00:00",
										"end": "08:00"
									}
								}
							}
						},
						{
							"group": "group-2",
							"id": "id",
							"params": {},
							"connector_type_id": "connector-type",
							"frequency": {
								"summary": true,
								"notify_when": "onActionGroupChange"
							}
						},
						{
							"group": "group-3",
							"id": "id",
							"params": {},
							"connector_type_id": "connector-type"
						}
					],
					"execution_status": {
						"status": "firing",
						"last_execution_date": "` + nowStr + `"
					},
					"scheduled_task_id": "scheduled-task-id",
					"schedule": {
						"interval": "1m"
					},
					"throttle": "throttle",
					"alert_delay": {
						"active": 4
					},
					"created_at": "` + nowStr + `",
					"updated_at": "` + nowStr + `",
					"mute_all": false,
					"muted_alert_ids": [],
					"revision": 0
				}`
				var resp kbapi.GetAlertingRuleIdResponse
				resp.JSON200 = new(struct {
					Actions []struct {
						AlertsFilter *struct {
							Query *struct {
								Dsl     *string `json:"dsl,omitempty"`
								Filters []struct {
									State *struct {
										Store kbapi.GetAlertingRuleId200ActionsAlertsFilterQueryFiltersStateStore `json:"store"`
									} `json:"$state,omitempty"`
									Meta  map[string]interface{}  `json:"meta"`
									Query *map[string]interface{} `json:"query,omitempty"`
								} `json:"filters"`
								Kql string `json:"kql"`
							} `json:"query,omitempty"`
							Timeframe *struct {
								Days  []kbapi.GetAlertingRuleId200ActionsAlertsFilterTimeframeDays `json:"days"`
								Hours struct {
									End   string `json:"end"`
									Start string `json:"start"`
								} `json:"hours"`
								Timezone string `json:"timezone"`
							} `json:"timeframe,omitempty"`
						} `json:"alerts_filter,omitempty"`
						ConnectorTypeId string `json:"connector_type_id"`
						Frequency       *struct {
							NotifyWhen kbapi.GetAlertingRuleId200ActionsFrequencyNotifyWhen `json:"notify_when"`
							Summary    bool                                                 `json:"summary"`
							Throttle   *string                                              `json:"throttle"`
						} `json:"frequency,omitempty"`
						Group                   *string                `json:"group,omitempty"`
						Id                      string                 `json:"id"`
						Params                  map[string]interface{} `json:"params"`
						UseAlertDataForTemplate *bool                  `json:"use_alert_data_for_template,omitempty"`
						Uuid                    *string                `json:"uuid,omitempty"`
					} `json:"actions"`
					ActiveSnoozes *[]string `json:"active_snoozes,omitempty"`
					AlertDelay    *struct {
						Active float32 `json:"active"`
					} `json:"alert_delay,omitempty"`
					ApiKeyCreatedByUser *bool   `json:"api_key_created_by_user"`
					ApiKeyOwner         *string `json:"api_key_owner"`
					Artifacts           *struct {
						Dashboards *[]struct {
							Id string `json:"id"`
						} `json:"dashboards,omitempty"`
						InvestigationGuide *struct {
							Blob string `json:"blob"`
						} `json:"investigation_guide,omitempty"`
					} `json:"artifacts,omitempty"`
					Consumer        string  `json:"consumer"`
					CreatedAt       string  `json:"created_at"`
					CreatedBy       *string `json:"created_by"`
					Enabled         bool    `json:"enabled"`
					ExecutionStatus struct {
						Error *struct {
							Message string                                               `json:"message"`
							Reason  kbapi.GetAlertingRuleId200ExecutionStatusErrorReason `json:"reason"`
						} `json:"error,omitempty"`
						LastDuration      *float32                                        `json:"last_duration,omitempty"`
						LastExecutionDate string                                          `json:"last_execution_date"`
						Status            kbapi.GetAlertingRuleId200ExecutionStatusStatus `json:"status"`
						Warning           *struct {
							Message string                                                 `json:"message"`
							Reason  kbapi.GetAlertingRuleId200ExecutionStatusWarningReason `json:"reason"`
						} `json:"warning,omitempty"`
					} `json:"execution_status"`
					Flapping *struct {
						Enabled               *bool   `json:"enabled,omitempty"`
						LookBackWindow        float32 `json:"look_back_window"`
						StatusChangeThreshold float32 `json:"status_change_threshold"`
					} `json:"flapping"`
					Id             string  `json:"id"`
					IsSnoozedUntil *string `json:"is_snoozed_until"`
					LastRun        *struct {
						AlertsCount struct {
							Active    *float32 `json:"active"`
							Ignored   *float32 `json:"ignored"`
							New       *float32 `json:"new"`
							Recovered *float32 `json:"recovered"`
						} `json:"alerts_count"`
						Outcome      kbapi.GetAlertingRuleId200LastRunOutcome  `json:"outcome"`
						OutcomeMsg   *[]string                                 `json:"outcome_msg"`
						OutcomeOrder *float32                                  `json:"outcome_order,omitempty"`
						Warning      *kbapi.GetAlertingRuleId200LastRunWarning `json:"warning"`
					} `json:"last_run"`
					MappedParams *map[string]interface{} `json:"mapped_params,omitempty"`
					Monitoring   *struct {
						Run struct {
							CalculatedMetrics struct {
								P50          *float32 `json:"p50,omitempty"`
								P95          *float32 `json:"p95,omitempty"`
								P99          *float32 `json:"p99,omitempty"`
								SuccessRatio float32  `json:"success_ratio"`
							} `json:"calculated_metrics"`
							History []struct {
								Duration  *float32                                               `json:"duration,omitempty"`
								Outcome   *kbapi.GetAlertingRuleId200MonitoringRunHistoryOutcome `json:"outcome,omitempty"`
								Success   bool                                                   `json:"success"`
								Timestamp float32                                                `json:"timestamp"`
							} `json:"history"`
							LastRun struct {
								Metrics struct {
									Duration     *float32 `json:"duration,omitempty"`
									GapDurationS *float32 `json:"gap_duration_s"`
									GapRange     *struct {
										Gte string `json:"gte"`
										Lte string `json:"lte"`
									} `json:"gap_range"`
									TotalAlertsCreated      *float32 `json:"total_alerts_created"`
									TotalAlertsDetected     *float32 `json:"total_alerts_detected"`
									TotalIndexingDurationMs *float32 `json:"total_indexing_duration_ms"`
									TotalSearchDurationMs   *float32 `json:"total_search_duration_ms"`
								} `json:"metrics"`
								Timestamp string `json:"timestamp"`
							} `json:"last_run"`
						} `json:"run"`
					} `json:"monitoring,omitempty"`
					MuteAll       bool                                  `json:"mute_all"`
					MutedAlertIds []string                              `json:"muted_alert_ids"`
					Name          string                                `json:"name"`
					NextRun       *string                               `json:"next_run"`
					NotifyWhen    *kbapi.GetAlertingRuleId200NotifyWhen `json:"notify_when"`
					Params        map[string]interface{}                `json:"params"`
					Revision      float32                               `json:"revision"`
					RuleTypeId    string                                `json:"rule_type_id"`
					Running       *bool                                 `json:"running"`
					Schedule      struct {
						Interval string `json:"interval"`
					} `json:"schedule"`
					ScheduledTaskId *string `json:"scheduled_task_id,omitempty"`
					SnoozeSchedule  *[]struct {
						Duration float32 `json:"duration"`
						Id       *string `json:"id,omitempty"`
						RRule    struct {
							Byhour     *[]float32                                                         `json:"byhour"`
							Byminute   *[]float32                                                         `json:"byminute"`
							Bymonth    *[]float32                                                         `json:"bymonth"`
							Bymonthday *[]float32                                                         `json:"bymonthday"`
							Bysecond   *[]float32                                                         `json:"bysecond"`
							Bysetpos   *[]float32                                                         `json:"bysetpos"`
							Byweekday  *[]kbapi.GetAlertingRuleId_200_SnoozeSchedule_RRule_Byweekday_Item `json:"byweekday"`
							Byweekno   *[]float32                                                         `json:"byweekno"`
							Byyearday  *[]float32                                                         `json:"byyearday"`
							Count      *float32                                                           `json:"count,omitempty"`
							Dtstart    string                                                             `json:"dtstart"`
							Freq       *kbapi.GetAlertingRuleId200SnoozeScheduleRRuleFreq                 `json:"freq,omitempty"`
							Interval   *float32                                                           `json:"interval,omitempty"`
							Tzid       string                                                             `json:"tzid"`
							Until      *string                                                            `json:"until,omitempty"`
							Wkst       *kbapi.GetAlertingRuleId200SnoozeScheduleRRuleWkst                 `json:"wkst,omitempty"`
						} `json:"rRule"`
						SkipRecurrences *[]string `json:"skipRecurrences,omitempty"`
					} `json:"snooze_schedule,omitempty"`
					Tags                 []string `json:"tags"`
					Throttle             *string  `json:"throttle"`
					UpdatedAt            string   `json:"updated_at"`
					UpdatedBy            *string  `json:"updated_by"`
					ViewInAppRelativeUrl *string  `json:"view_in_app_relative_url"`
				})
				err := json.Unmarshal([]byte(jsonData), resp.JSON200)
				if err != nil {
					panic(err)
				}
				return &resp
			}(),
			expectedModel: &models.AlertingRule{
				RuleID:          "id",
				SpaceID:         "space-id",
				Name:            "name",
				Consumer:        "consumer",
				Params:          map[string]interface{}{},
				RuleTypeID:      "rule-type-id",
				Enabled:         utils.Pointer(true),
				Tags:            []string{"hello"},
				NotifyWhen:      utils.Pointer("broken"),
				Schedule:        models.AlertingRuleSchedule{Interval: "1m"},
				Throttle:        utils.Pointer("throttle"),
				ScheduledTaskID: utils.Pointer("scheduled-task-id"),
				ExecutionStatus: models.AlertingRuleExecutionStatus{
					LastExecutionDate: &now,
					Status:            utils.Pointer("firing"),
				},
				Actions: []models.AlertingRuleAction{
					{
						Group:  "group-1",
						ID:     "id",
						Params: map[string]interface{}{},
						Frequency: &models.ActionFrequency{
							Summary:    true,
							NotifyWhen: "onThrottleInterval",
							Throttle:   utils.Pointer("10s"),
						},
						AlertsFilter: &models.ActionAlertsFilter{
							Kql: utils.Pointer("foobar"),
							Timeframe: &models.AlertsFilterTimeframe{
								Days:       []int32{3, 5, 7},
								Timezone:   "UTC+1",
								HoursStart: "00:00",
								HoursEnd:   "08:00",
							},
						},
					},
					{
						Group:  "group-2",
						ID:     "id",
						Params: map[string]interface{}{},
						Frequency: &models.ActionFrequency{
							Summary:    true,
							NotifyWhen: "onActionGroupChange",
						},
					},
					{
						Group:  "group-3",
						ID:     "id",
						Params: map[string]interface{}{},
					},
				},
				AlertDelay: utils.Pointer(float32(4)),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := ruleResponseToModel(tt.spaceId, tt.ruleResponse)

			require.Equal(t, tt.expectedModel, model)
		})
	}
}

// Test_CreateUpdateAlertingRule tests have been removed as they now test kibana_oapi layer
// Error handling for Create/Update is tested at the kibana_oapi level
// The ruleResponseToModel function above provides sufficient coverage for the transformation logic
