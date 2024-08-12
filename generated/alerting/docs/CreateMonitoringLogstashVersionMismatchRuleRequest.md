# CreateMonitoringLogstashVersionMismatchRuleRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Actions** | Pointer to [**[]ActionsInner**](ActionsInner.md) |  | [optional] 
**Consumer** | **string** | The name of the application or feature that owns the rule. For example: &#x60;alerts&#x60;, &#x60;apm&#x60;, &#x60;discover&#x60;, &#x60;infrastructure&#x60;, &#x60;logs&#x60;, &#x60;metrics&#x60;, &#x60;ml&#x60;, &#x60;monitoring&#x60;, &#x60;securitySolution&#x60;, &#x60;siem&#x60;, &#x60;stackAlerts&#x60;, or &#x60;uptime&#x60;.  | 
**Enabled** | Pointer to **bool** | Indicates whether you want to run the rule on an interval basis after it is created. | [optional] 
**Name** | **string** | The name of the rule. While this name does not have to be unique, a distinctive name can help you identify a rule.  | 
**NotifyWhen** | Pointer to [**NotifyWhen**](NotifyWhen.md) |  | [optional] 
**Params** | **map[string]interface{}** | The parameters for a Logstash version mismatch rule. | 
**RuleTypeId** | **string** | The ID of the rule type that you want to call when the rule is scheduled to run. | 
**Schedule** | [**Schedule**](Schedule.md) |  | 
**Tags** | Pointer to **[]string** |  | [optional] 
**Throttle** | Pointer to **NullableString** | Deprecated in 8.13.0. Use the &#x60;throttle&#x60; property in the action &#x60;frequency&#x60; object instead. The throttle interval, which defines how often an alert generates repeated actions. NOTE: You cannot specify the throttle interval at both the rule and action level. If you set it at the rule level then update the rule in Kibana, it is automatically changed to use action-specific values.  | [optional] 

## Methods

### NewCreateMonitoringLogstashVersionMismatchRuleRequest

`func NewCreateMonitoringLogstashVersionMismatchRuleRequest(consumer string, name string, params map[string]interface{}, ruleTypeId string, schedule Schedule, ) *CreateMonitoringLogstashVersionMismatchRuleRequest`

NewCreateMonitoringLogstashVersionMismatchRuleRequest instantiates a new CreateMonitoringLogstashVersionMismatchRuleRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewCreateMonitoringLogstashVersionMismatchRuleRequestWithDefaults

`func NewCreateMonitoringLogstashVersionMismatchRuleRequestWithDefaults() *CreateMonitoringLogstashVersionMismatchRuleRequest`

NewCreateMonitoringLogstashVersionMismatchRuleRequestWithDefaults instantiates a new CreateMonitoringLogstashVersionMismatchRuleRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetActions

`func (o *CreateMonitoringLogstashVersionMismatchRuleRequest) GetActions() []ActionsInner`

GetActions returns the Actions field if non-nil, zero value otherwise.

### GetActionsOk

`func (o *CreateMonitoringLogstashVersionMismatchRuleRequest) GetActionsOk() (*[]ActionsInner, bool)`

GetActionsOk returns a tuple with the Actions field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetActions

`func (o *CreateMonitoringLogstashVersionMismatchRuleRequest) SetActions(v []ActionsInner)`

SetActions sets Actions field to given value.

### HasActions

`func (o *CreateMonitoringLogstashVersionMismatchRuleRequest) HasActions() bool`

HasActions returns a boolean if a field has been set.

### GetConsumer

`func (o *CreateMonitoringLogstashVersionMismatchRuleRequest) GetConsumer() string`

GetConsumer returns the Consumer field if non-nil, zero value otherwise.

### GetConsumerOk

`func (o *CreateMonitoringLogstashVersionMismatchRuleRequest) GetConsumerOk() (*string, bool)`

GetConsumerOk returns a tuple with the Consumer field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetConsumer

`func (o *CreateMonitoringLogstashVersionMismatchRuleRequest) SetConsumer(v string)`

SetConsumer sets Consumer field to given value.


### GetEnabled

`func (o *CreateMonitoringLogstashVersionMismatchRuleRequest) GetEnabled() bool`

GetEnabled returns the Enabled field if non-nil, zero value otherwise.

### GetEnabledOk

`func (o *CreateMonitoringLogstashVersionMismatchRuleRequest) GetEnabledOk() (*bool, bool)`

GetEnabledOk returns a tuple with the Enabled field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEnabled

`func (o *CreateMonitoringLogstashVersionMismatchRuleRequest) SetEnabled(v bool)`

SetEnabled sets Enabled field to given value.

### HasEnabled

`func (o *CreateMonitoringLogstashVersionMismatchRuleRequest) HasEnabled() bool`

HasEnabled returns a boolean if a field has been set.

### GetName

`func (o *CreateMonitoringLogstashVersionMismatchRuleRequest) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *CreateMonitoringLogstashVersionMismatchRuleRequest) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *CreateMonitoringLogstashVersionMismatchRuleRequest) SetName(v string)`

SetName sets Name field to given value.


### GetNotifyWhen

`func (o *CreateMonitoringLogstashVersionMismatchRuleRequest) GetNotifyWhen() NotifyWhen`

GetNotifyWhen returns the NotifyWhen field if non-nil, zero value otherwise.

### GetNotifyWhenOk

`func (o *CreateMonitoringLogstashVersionMismatchRuleRequest) GetNotifyWhenOk() (*NotifyWhen, bool)`

GetNotifyWhenOk returns a tuple with the NotifyWhen field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNotifyWhen

`func (o *CreateMonitoringLogstashVersionMismatchRuleRequest) SetNotifyWhen(v NotifyWhen)`

SetNotifyWhen sets NotifyWhen field to given value.

### HasNotifyWhen

`func (o *CreateMonitoringLogstashVersionMismatchRuleRequest) HasNotifyWhen() bool`

HasNotifyWhen returns a boolean if a field has been set.

### GetParams

`func (o *CreateMonitoringLogstashVersionMismatchRuleRequest) GetParams() map[string]interface{}`

GetParams returns the Params field if non-nil, zero value otherwise.

### GetParamsOk

`func (o *CreateMonitoringLogstashVersionMismatchRuleRequest) GetParamsOk() (*map[string]interface{}, bool)`

GetParamsOk returns a tuple with the Params field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetParams

`func (o *CreateMonitoringLogstashVersionMismatchRuleRequest) SetParams(v map[string]interface{})`

SetParams sets Params field to given value.


### GetRuleTypeId

`func (o *CreateMonitoringLogstashVersionMismatchRuleRequest) GetRuleTypeId() string`

GetRuleTypeId returns the RuleTypeId field if non-nil, zero value otherwise.

### GetRuleTypeIdOk

`func (o *CreateMonitoringLogstashVersionMismatchRuleRequest) GetRuleTypeIdOk() (*string, bool)`

GetRuleTypeIdOk returns a tuple with the RuleTypeId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRuleTypeId

`func (o *CreateMonitoringLogstashVersionMismatchRuleRequest) SetRuleTypeId(v string)`

SetRuleTypeId sets RuleTypeId field to given value.


### GetSchedule

`func (o *CreateMonitoringLogstashVersionMismatchRuleRequest) GetSchedule() Schedule`

GetSchedule returns the Schedule field if non-nil, zero value otherwise.

### GetScheduleOk

`func (o *CreateMonitoringLogstashVersionMismatchRuleRequest) GetScheduleOk() (*Schedule, bool)`

GetScheduleOk returns a tuple with the Schedule field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSchedule

`func (o *CreateMonitoringLogstashVersionMismatchRuleRequest) SetSchedule(v Schedule)`

SetSchedule sets Schedule field to given value.


### GetTags

`func (o *CreateMonitoringLogstashVersionMismatchRuleRequest) GetTags() []string`

GetTags returns the Tags field if non-nil, zero value otherwise.

### GetTagsOk

`func (o *CreateMonitoringLogstashVersionMismatchRuleRequest) GetTagsOk() (*[]string, bool)`

GetTagsOk returns a tuple with the Tags field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTags

`func (o *CreateMonitoringLogstashVersionMismatchRuleRequest) SetTags(v []string)`

SetTags sets Tags field to given value.

### HasTags

`func (o *CreateMonitoringLogstashVersionMismatchRuleRequest) HasTags() bool`

HasTags returns a boolean if a field has been set.

### GetThrottle

`func (o *CreateMonitoringLogstashVersionMismatchRuleRequest) GetThrottle() string`

GetThrottle returns the Throttle field if non-nil, zero value otherwise.

### GetThrottleOk

`func (o *CreateMonitoringLogstashVersionMismatchRuleRequest) GetThrottleOk() (*string, bool)`

GetThrottleOk returns a tuple with the Throttle field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetThrottle

`func (o *CreateMonitoringLogstashVersionMismatchRuleRequest) SetThrottle(v string)`

SetThrottle sets Throttle field to given value.

### HasThrottle

`func (o *CreateMonitoringLogstashVersionMismatchRuleRequest) HasThrottle() bool`

HasThrottle returns a boolean if a field has been set.

### SetThrottleNil

`func (o *CreateMonitoringLogstashVersionMismatchRuleRequest) SetThrottleNil(b bool)`

 SetThrottleNil sets the value for Throttle to be an explicit nil

### UnsetThrottle
`func (o *CreateMonitoringLogstashVersionMismatchRuleRequest) UnsetThrottle()`

UnsetThrottle ensures that no value is present for Throttle, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


