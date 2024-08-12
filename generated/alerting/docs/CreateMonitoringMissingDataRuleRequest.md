# CreateMonitoringMissingDataRuleRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Actions** | Pointer to [**[]ActionsInner**](ActionsInner.md) |  | [optional] 
**Consumer** | **string** | The name of the application or feature that owns the rule. For example: &#x60;alerts&#x60;, &#x60;apm&#x60;, &#x60;discover&#x60;, &#x60;infrastructure&#x60;, &#x60;logs&#x60;, &#x60;metrics&#x60;, &#x60;ml&#x60;, &#x60;monitoring&#x60;, &#x60;securitySolution&#x60;, &#x60;siem&#x60;, &#x60;stackAlerts&#x60;, or &#x60;uptime&#x60;.  | 
**Enabled** | Pointer to **bool** | Indicates whether you want to run the rule on an interval basis after it is created. | [optional] 
**Name** | **string** | The name of the rule. While this name does not have to be unique, a distinctive name can help you identify a rule.  | 
**NotifyWhen** | Pointer to [**NotifyWhen**](NotifyWhen.md) |  | [optional] 
**Params** | **map[string]interface{}** | The parameters for a missing monitoring data rule. | 
**RuleTypeId** | **string** | The ID of the rule type that you want to call when the rule is scheduled to run. | 
**Schedule** | [**Schedule**](Schedule.md) |  | 
**Tags** | Pointer to **[]string** |  | [optional] 
**Throttle** | Pointer to **NullableString** | Deprecated in 8.13.0. Use the &#x60;throttle&#x60; property in the action &#x60;frequency&#x60; object instead. The throttle interval, which defines how often an alert generates repeated actions. NOTE: You cannot specify the throttle interval at both the rule and action level. If you set it at the rule level then update the rule in Kibana, it is automatically changed to use action-specific values.  | [optional] 

## Methods

### NewCreateMonitoringMissingDataRuleRequest

`func NewCreateMonitoringMissingDataRuleRequest(consumer string, name string, params map[string]interface{}, ruleTypeId string, schedule Schedule, ) *CreateMonitoringMissingDataRuleRequest`

NewCreateMonitoringMissingDataRuleRequest instantiates a new CreateMonitoringMissingDataRuleRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewCreateMonitoringMissingDataRuleRequestWithDefaults

`func NewCreateMonitoringMissingDataRuleRequestWithDefaults() *CreateMonitoringMissingDataRuleRequest`

NewCreateMonitoringMissingDataRuleRequestWithDefaults instantiates a new CreateMonitoringMissingDataRuleRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetActions

`func (o *CreateMonitoringMissingDataRuleRequest) GetActions() []ActionsInner`

GetActions returns the Actions field if non-nil, zero value otherwise.

### GetActionsOk

`func (o *CreateMonitoringMissingDataRuleRequest) GetActionsOk() (*[]ActionsInner, bool)`

GetActionsOk returns a tuple with the Actions field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetActions

`func (o *CreateMonitoringMissingDataRuleRequest) SetActions(v []ActionsInner)`

SetActions sets Actions field to given value.

### HasActions

`func (o *CreateMonitoringMissingDataRuleRequest) HasActions() bool`

HasActions returns a boolean if a field has been set.

### GetConsumer

`func (o *CreateMonitoringMissingDataRuleRequest) GetConsumer() string`

GetConsumer returns the Consumer field if non-nil, zero value otherwise.

### GetConsumerOk

`func (o *CreateMonitoringMissingDataRuleRequest) GetConsumerOk() (*string, bool)`

GetConsumerOk returns a tuple with the Consumer field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetConsumer

`func (o *CreateMonitoringMissingDataRuleRequest) SetConsumer(v string)`

SetConsumer sets Consumer field to given value.


### GetEnabled

`func (o *CreateMonitoringMissingDataRuleRequest) GetEnabled() bool`

GetEnabled returns the Enabled field if non-nil, zero value otherwise.

### GetEnabledOk

`func (o *CreateMonitoringMissingDataRuleRequest) GetEnabledOk() (*bool, bool)`

GetEnabledOk returns a tuple with the Enabled field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEnabled

`func (o *CreateMonitoringMissingDataRuleRequest) SetEnabled(v bool)`

SetEnabled sets Enabled field to given value.

### HasEnabled

`func (o *CreateMonitoringMissingDataRuleRequest) HasEnabled() bool`

HasEnabled returns a boolean if a field has been set.

### GetName

`func (o *CreateMonitoringMissingDataRuleRequest) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *CreateMonitoringMissingDataRuleRequest) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *CreateMonitoringMissingDataRuleRequest) SetName(v string)`

SetName sets Name field to given value.


### GetNotifyWhen

`func (o *CreateMonitoringMissingDataRuleRequest) GetNotifyWhen() NotifyWhen`

GetNotifyWhen returns the NotifyWhen field if non-nil, zero value otherwise.

### GetNotifyWhenOk

`func (o *CreateMonitoringMissingDataRuleRequest) GetNotifyWhenOk() (*NotifyWhen, bool)`

GetNotifyWhenOk returns a tuple with the NotifyWhen field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNotifyWhen

`func (o *CreateMonitoringMissingDataRuleRequest) SetNotifyWhen(v NotifyWhen)`

SetNotifyWhen sets NotifyWhen field to given value.

### HasNotifyWhen

`func (o *CreateMonitoringMissingDataRuleRequest) HasNotifyWhen() bool`

HasNotifyWhen returns a boolean if a field has been set.

### GetParams

`func (o *CreateMonitoringMissingDataRuleRequest) GetParams() map[string]interface{}`

GetParams returns the Params field if non-nil, zero value otherwise.

### GetParamsOk

`func (o *CreateMonitoringMissingDataRuleRequest) GetParamsOk() (*map[string]interface{}, bool)`

GetParamsOk returns a tuple with the Params field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetParams

`func (o *CreateMonitoringMissingDataRuleRequest) SetParams(v map[string]interface{})`

SetParams sets Params field to given value.


### GetRuleTypeId

`func (o *CreateMonitoringMissingDataRuleRequest) GetRuleTypeId() string`

GetRuleTypeId returns the RuleTypeId field if non-nil, zero value otherwise.

### GetRuleTypeIdOk

`func (o *CreateMonitoringMissingDataRuleRequest) GetRuleTypeIdOk() (*string, bool)`

GetRuleTypeIdOk returns a tuple with the RuleTypeId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRuleTypeId

`func (o *CreateMonitoringMissingDataRuleRequest) SetRuleTypeId(v string)`

SetRuleTypeId sets RuleTypeId field to given value.


### GetSchedule

`func (o *CreateMonitoringMissingDataRuleRequest) GetSchedule() Schedule`

GetSchedule returns the Schedule field if non-nil, zero value otherwise.

### GetScheduleOk

`func (o *CreateMonitoringMissingDataRuleRequest) GetScheduleOk() (*Schedule, bool)`

GetScheduleOk returns a tuple with the Schedule field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSchedule

`func (o *CreateMonitoringMissingDataRuleRequest) SetSchedule(v Schedule)`

SetSchedule sets Schedule field to given value.


### GetTags

`func (o *CreateMonitoringMissingDataRuleRequest) GetTags() []string`

GetTags returns the Tags field if non-nil, zero value otherwise.

### GetTagsOk

`func (o *CreateMonitoringMissingDataRuleRequest) GetTagsOk() (*[]string, bool)`

GetTagsOk returns a tuple with the Tags field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTags

`func (o *CreateMonitoringMissingDataRuleRequest) SetTags(v []string)`

SetTags sets Tags field to given value.

### HasTags

`func (o *CreateMonitoringMissingDataRuleRequest) HasTags() bool`

HasTags returns a boolean if a field has been set.

### GetThrottle

`func (o *CreateMonitoringMissingDataRuleRequest) GetThrottle() string`

GetThrottle returns the Throttle field if non-nil, zero value otherwise.

### GetThrottleOk

`func (o *CreateMonitoringMissingDataRuleRequest) GetThrottleOk() (*string, bool)`

GetThrottleOk returns a tuple with the Throttle field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetThrottle

`func (o *CreateMonitoringMissingDataRuleRequest) SetThrottle(v string)`

SetThrottle sets Throttle field to given value.

### HasThrottle

`func (o *CreateMonitoringMissingDataRuleRequest) HasThrottle() bool`

HasThrottle returns a boolean if a field has been set.

### SetThrottleNil

`func (o *CreateMonitoringMissingDataRuleRequest) SetThrottleNil(b bool)`

 SetThrottleNil sets the value for Throttle to be an explicit nil

### UnsetThrottle
`func (o *CreateMonitoringMissingDataRuleRequest) UnsetThrottle()`

UnsetThrottle ensures that no value is present for Throttle, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


