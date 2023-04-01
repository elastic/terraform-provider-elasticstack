# RuleResponseProperties

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Actions** | [**[]ActionsInner**](ActionsInner.md) |  | [default to []]
**ApiKeyOwner** | **NullableString** |  | 
**Consumer** | **string** | The application or feature that owns the rule. For example, &#x60;alerts&#x60;, &#x60;apm&#x60;, &#x60;discover&#x60;, &#x60;infrastructure&#x60;, &#x60;logs&#x60;, &#x60;metrics&#x60;, &#x60;ml&#x60;, &#x60;monitoring&#x60;, &#x60;securitySolution&#x60;, &#x60;siem&#x60;, &#x60;stackAlerts&#x60;, or &#x60;uptime&#x60;. | 
**CreatedAt** | **time.Time** | The date and time that the rule was created. | 
**CreatedBy** | **NullableString** | The identifier for the user that created the rule. | 
**Enabled** | **bool** | Indicates whether the rule is currently enabled. | 
**ExecutionStatus** | [**RuleResponsePropertiesExecutionStatus**](RuleResponsePropertiesExecutionStatus.md) |  | 
**Id** | **string** | The identifier for the rule. | 
**LastRun** | Pointer to [**RuleResponsePropertiesLastRun**](RuleResponsePropertiesLastRun.md) |  | [optional] 
**MutedAlertIds** | **[]string** |  | 
**MuteAll** | **bool** |  | 
**Name** | **string** | The name of the rule. | 
**NextRun** | Pointer to **NullableTime** |  | [optional] 
**NotifyWhen** | Pointer to [**NotifyWhen**](NotifyWhen.md) |  | [optional] 
**Params** | **map[string]interface{}** | The parameters for the rule. | 
**RuleTypeId** | **string** | The identifier for the type of rule. For example, &#x60;.es-query&#x60;, &#x60;.index-threshold&#x60;, &#x60;logs.alert.document.count&#x60;, &#x60;monitoring_alert_cluster_health&#x60;, &#x60;siem.thresholdRule&#x60;, or &#x60;xpack.ml.anomaly_detection_alert&#x60;.  | 
**Running** | Pointer to **bool** | Indicates whether the rule is running. | [optional] 
**Schedule** | [**Schedule**](Schedule.md) |  | 
**ScheduledTaskId** | Pointer to **string** |  | [optional] 
**Tags** | **[]string** | The tags for the rule. | [default to []]
**Throttle** | **NullableString** | The throttle interval, which defines how often an alert generates repeated actions. It is applicable only if &#x60;notify_when&#x60; is set to &#x60;onThrottleInterval&#x60;. It is specified in seconds, minutes, hours, or days. | 
**UpdatedAt** | **string** | The date and time that the rule was updated most recently. | 
**UpdatedBy** | **NullableString** | The identifier for the user that updated this rule most recently. | 

## Methods

### NewRuleResponseProperties

`func NewRuleResponseProperties(actions []ActionsInner, apiKeyOwner NullableString, consumer string, createdAt time.Time, createdBy NullableString, enabled bool, executionStatus RuleResponsePropertiesExecutionStatus, id string, mutedAlertIds []string, muteAll bool, name string, params map[string]interface{}, ruleTypeId string, schedule Schedule, tags []string, throttle NullableString, updatedAt string, updatedBy NullableString, ) *RuleResponseProperties`

NewRuleResponseProperties instantiates a new RuleResponseProperties object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewRuleResponsePropertiesWithDefaults

`func NewRuleResponsePropertiesWithDefaults() *RuleResponseProperties`

NewRuleResponsePropertiesWithDefaults instantiates a new RuleResponseProperties object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetActions

`func (o *RuleResponseProperties) GetActions() []ActionsInner`

GetActions returns the Actions field if non-nil, zero value otherwise.

### GetActionsOk

`func (o *RuleResponseProperties) GetActionsOk() (*[]ActionsInner, bool)`

GetActionsOk returns a tuple with the Actions field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetActions

`func (o *RuleResponseProperties) SetActions(v []ActionsInner)`

SetActions sets Actions field to given value.


### SetActionsNil

`func (o *RuleResponseProperties) SetActionsNil(b bool)`

 SetActionsNil sets the value for Actions to be an explicit nil

### UnsetActions
`func (o *RuleResponseProperties) UnsetActions()`

UnsetActions ensures that no value is present for Actions, not even an explicit nil
### GetApiKeyOwner

`func (o *RuleResponseProperties) GetApiKeyOwner() string`

GetApiKeyOwner returns the ApiKeyOwner field if non-nil, zero value otherwise.

### GetApiKeyOwnerOk

`func (o *RuleResponseProperties) GetApiKeyOwnerOk() (*string, bool)`

GetApiKeyOwnerOk returns a tuple with the ApiKeyOwner field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetApiKeyOwner

`func (o *RuleResponseProperties) SetApiKeyOwner(v string)`

SetApiKeyOwner sets ApiKeyOwner field to given value.


### SetApiKeyOwnerNil

`func (o *RuleResponseProperties) SetApiKeyOwnerNil(b bool)`

 SetApiKeyOwnerNil sets the value for ApiKeyOwner to be an explicit nil

### UnsetApiKeyOwner
`func (o *RuleResponseProperties) UnsetApiKeyOwner()`

UnsetApiKeyOwner ensures that no value is present for ApiKeyOwner, not even an explicit nil
### GetConsumer

`func (o *RuleResponseProperties) GetConsumer() string`

GetConsumer returns the Consumer field if non-nil, zero value otherwise.

### GetConsumerOk

`func (o *RuleResponseProperties) GetConsumerOk() (*string, bool)`

GetConsumerOk returns a tuple with the Consumer field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetConsumer

`func (o *RuleResponseProperties) SetConsumer(v string)`

SetConsumer sets Consumer field to given value.


### GetCreatedAt

`func (o *RuleResponseProperties) GetCreatedAt() time.Time`

GetCreatedAt returns the CreatedAt field if non-nil, zero value otherwise.

### GetCreatedAtOk

`func (o *RuleResponseProperties) GetCreatedAtOk() (*time.Time, bool)`

GetCreatedAtOk returns a tuple with the CreatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedAt

`func (o *RuleResponseProperties) SetCreatedAt(v time.Time)`

SetCreatedAt sets CreatedAt field to given value.


### GetCreatedBy

`func (o *RuleResponseProperties) GetCreatedBy() string`

GetCreatedBy returns the CreatedBy field if non-nil, zero value otherwise.

### GetCreatedByOk

`func (o *RuleResponseProperties) GetCreatedByOk() (*string, bool)`

GetCreatedByOk returns a tuple with the CreatedBy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedBy

`func (o *RuleResponseProperties) SetCreatedBy(v string)`

SetCreatedBy sets CreatedBy field to given value.


### SetCreatedByNil

`func (o *RuleResponseProperties) SetCreatedByNil(b bool)`

 SetCreatedByNil sets the value for CreatedBy to be an explicit nil

### UnsetCreatedBy
`func (o *RuleResponseProperties) UnsetCreatedBy()`

UnsetCreatedBy ensures that no value is present for CreatedBy, not even an explicit nil
### GetEnabled

`func (o *RuleResponseProperties) GetEnabled() bool`

GetEnabled returns the Enabled field if non-nil, zero value otherwise.

### GetEnabledOk

`func (o *RuleResponseProperties) GetEnabledOk() (*bool, bool)`

GetEnabledOk returns a tuple with the Enabled field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEnabled

`func (o *RuleResponseProperties) SetEnabled(v bool)`

SetEnabled sets Enabled field to given value.


### GetExecutionStatus

`func (o *RuleResponseProperties) GetExecutionStatus() RuleResponsePropertiesExecutionStatus`

GetExecutionStatus returns the ExecutionStatus field if non-nil, zero value otherwise.

### GetExecutionStatusOk

`func (o *RuleResponseProperties) GetExecutionStatusOk() (*RuleResponsePropertiesExecutionStatus, bool)`

GetExecutionStatusOk returns a tuple with the ExecutionStatus field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetExecutionStatus

`func (o *RuleResponseProperties) SetExecutionStatus(v RuleResponsePropertiesExecutionStatus)`

SetExecutionStatus sets ExecutionStatus field to given value.


### GetId

`func (o *RuleResponseProperties) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *RuleResponseProperties) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *RuleResponseProperties) SetId(v string)`

SetId sets Id field to given value.


### GetLastRun

`func (o *RuleResponseProperties) GetLastRun() RuleResponsePropertiesLastRun`

GetLastRun returns the LastRun field if non-nil, zero value otherwise.

### GetLastRunOk

`func (o *RuleResponseProperties) GetLastRunOk() (*RuleResponsePropertiesLastRun, bool)`

GetLastRunOk returns a tuple with the LastRun field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLastRun

`func (o *RuleResponseProperties) SetLastRun(v RuleResponsePropertiesLastRun)`

SetLastRun sets LastRun field to given value.

### HasLastRun

`func (o *RuleResponseProperties) HasLastRun() bool`

HasLastRun returns a boolean if a field has been set.

### GetMutedAlertIds

`func (o *RuleResponseProperties) GetMutedAlertIds() []string`

GetMutedAlertIds returns the MutedAlertIds field if non-nil, zero value otherwise.

### GetMutedAlertIdsOk

`func (o *RuleResponseProperties) GetMutedAlertIdsOk() (*[]string, bool)`

GetMutedAlertIdsOk returns a tuple with the MutedAlertIds field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMutedAlertIds

`func (o *RuleResponseProperties) SetMutedAlertIds(v []string)`

SetMutedAlertIds sets MutedAlertIds field to given value.


### SetMutedAlertIdsNil

`func (o *RuleResponseProperties) SetMutedAlertIdsNil(b bool)`

 SetMutedAlertIdsNil sets the value for MutedAlertIds to be an explicit nil

### UnsetMutedAlertIds
`func (o *RuleResponseProperties) UnsetMutedAlertIds()`

UnsetMutedAlertIds ensures that no value is present for MutedAlertIds, not even an explicit nil
### GetMuteAll

`func (o *RuleResponseProperties) GetMuteAll() bool`

GetMuteAll returns the MuteAll field if non-nil, zero value otherwise.

### GetMuteAllOk

`func (o *RuleResponseProperties) GetMuteAllOk() (*bool, bool)`

GetMuteAllOk returns a tuple with the MuteAll field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMuteAll

`func (o *RuleResponseProperties) SetMuteAll(v bool)`

SetMuteAll sets MuteAll field to given value.


### GetName

`func (o *RuleResponseProperties) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *RuleResponseProperties) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *RuleResponseProperties) SetName(v string)`

SetName sets Name field to given value.


### GetNextRun

`func (o *RuleResponseProperties) GetNextRun() time.Time`

GetNextRun returns the NextRun field if non-nil, zero value otherwise.

### GetNextRunOk

`func (o *RuleResponseProperties) GetNextRunOk() (*time.Time, bool)`

GetNextRunOk returns a tuple with the NextRun field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNextRun

`func (o *RuleResponseProperties) SetNextRun(v time.Time)`

SetNextRun sets NextRun field to given value.

### HasNextRun

`func (o *RuleResponseProperties) HasNextRun() bool`

HasNextRun returns a boolean if a field has been set.

### SetNextRunNil

`func (o *RuleResponseProperties) SetNextRunNil(b bool)`

 SetNextRunNil sets the value for NextRun to be an explicit nil

### UnsetNextRun
`func (o *RuleResponseProperties) UnsetNextRun()`

UnsetNextRun ensures that no value is present for NextRun, not even an explicit nil
### GetNotifyWhen

`func (o *RuleResponseProperties) GetNotifyWhen() NotifyWhen`

GetNotifyWhen returns the NotifyWhen field if non-nil, zero value otherwise.

### GetNotifyWhenOk

`func (o *RuleResponseProperties) GetNotifyWhenOk() (*NotifyWhen, bool)`

GetNotifyWhenOk returns a tuple with the NotifyWhen field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNotifyWhen

`func (o *RuleResponseProperties) SetNotifyWhen(v NotifyWhen)`

SetNotifyWhen sets NotifyWhen field to given value.

### HasNotifyWhen

`func (o *RuleResponseProperties) HasNotifyWhen() bool`

HasNotifyWhen returns a boolean if a field has been set.

### GetParams

`func (o *RuleResponseProperties) GetParams() map[string]interface{}`

GetParams returns the Params field if non-nil, zero value otherwise.

### GetParamsOk

`func (o *RuleResponseProperties) GetParamsOk() (*map[string]interface{}, bool)`

GetParamsOk returns a tuple with the Params field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetParams

`func (o *RuleResponseProperties) SetParams(v map[string]interface{})`

SetParams sets Params field to given value.


### GetRuleTypeId

`func (o *RuleResponseProperties) GetRuleTypeId() string`

GetRuleTypeId returns the RuleTypeId field if non-nil, zero value otherwise.

### GetRuleTypeIdOk

`func (o *RuleResponseProperties) GetRuleTypeIdOk() (*string, bool)`

GetRuleTypeIdOk returns a tuple with the RuleTypeId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRuleTypeId

`func (o *RuleResponseProperties) SetRuleTypeId(v string)`

SetRuleTypeId sets RuleTypeId field to given value.


### GetRunning

`func (o *RuleResponseProperties) GetRunning() bool`

GetRunning returns the Running field if non-nil, zero value otherwise.

### GetRunningOk

`func (o *RuleResponseProperties) GetRunningOk() (*bool, bool)`

GetRunningOk returns a tuple with the Running field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRunning

`func (o *RuleResponseProperties) SetRunning(v bool)`

SetRunning sets Running field to given value.

### HasRunning

`func (o *RuleResponseProperties) HasRunning() bool`

HasRunning returns a boolean if a field has been set.

### GetSchedule

`func (o *RuleResponseProperties) GetSchedule() Schedule`

GetSchedule returns the Schedule field if non-nil, zero value otherwise.

### GetScheduleOk

`func (o *RuleResponseProperties) GetScheduleOk() (*Schedule, bool)`

GetScheduleOk returns a tuple with the Schedule field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSchedule

`func (o *RuleResponseProperties) SetSchedule(v Schedule)`

SetSchedule sets Schedule field to given value.


### GetScheduledTaskId

`func (o *RuleResponseProperties) GetScheduledTaskId() string`

GetScheduledTaskId returns the ScheduledTaskId field if non-nil, zero value otherwise.

### GetScheduledTaskIdOk

`func (o *RuleResponseProperties) GetScheduledTaskIdOk() (*string, bool)`

GetScheduledTaskIdOk returns a tuple with the ScheduledTaskId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetScheduledTaskId

`func (o *RuleResponseProperties) SetScheduledTaskId(v string)`

SetScheduledTaskId sets ScheduledTaskId field to given value.

### HasScheduledTaskId

`func (o *RuleResponseProperties) HasScheduledTaskId() bool`

HasScheduledTaskId returns a boolean if a field has been set.

### GetTags

`func (o *RuleResponseProperties) GetTags() []string`

GetTags returns the Tags field if non-nil, zero value otherwise.

### GetTagsOk

`func (o *RuleResponseProperties) GetTagsOk() (*[]string, bool)`

GetTagsOk returns a tuple with the Tags field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTags

`func (o *RuleResponseProperties) SetTags(v []string)`

SetTags sets Tags field to given value.


### GetThrottle

`func (o *RuleResponseProperties) GetThrottle() string`

GetThrottle returns the Throttle field if non-nil, zero value otherwise.

### GetThrottleOk

`func (o *RuleResponseProperties) GetThrottleOk() (*string, bool)`

GetThrottleOk returns a tuple with the Throttle field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetThrottle

`func (o *RuleResponseProperties) SetThrottle(v string)`

SetThrottle sets Throttle field to given value.


### SetThrottleNil

`func (o *RuleResponseProperties) SetThrottleNil(b bool)`

 SetThrottleNil sets the value for Throttle to be an explicit nil

### UnsetThrottle
`func (o *RuleResponseProperties) UnsetThrottle()`

UnsetThrottle ensures that no value is present for Throttle, not even an explicit nil
### GetUpdatedAt

`func (o *RuleResponseProperties) GetUpdatedAt() string`

GetUpdatedAt returns the UpdatedAt field if non-nil, zero value otherwise.

### GetUpdatedAtOk

`func (o *RuleResponseProperties) GetUpdatedAtOk() (*string, bool)`

GetUpdatedAtOk returns a tuple with the UpdatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUpdatedAt

`func (o *RuleResponseProperties) SetUpdatedAt(v string)`

SetUpdatedAt sets UpdatedAt field to given value.


### GetUpdatedBy

`func (o *RuleResponseProperties) GetUpdatedBy() string`

GetUpdatedBy returns the UpdatedBy field if non-nil, zero value otherwise.

### GetUpdatedByOk

`func (o *RuleResponseProperties) GetUpdatedByOk() (*string, bool)`

GetUpdatedByOk returns a tuple with the UpdatedBy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUpdatedBy

`func (o *RuleResponseProperties) SetUpdatedBy(v string)`

SetUpdatedBy sets UpdatedBy field to given value.


### SetUpdatedByNil

`func (o *RuleResponseProperties) SetUpdatedByNil(b bool)`

 SetUpdatedByNil sets the value for UpdatedBy to be an explicit nil

### UnsetUpdatedBy
`func (o *RuleResponseProperties) UnsetUpdatedBy()`

UnsetUpdatedBy ensures that no value is present for UpdatedBy, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


