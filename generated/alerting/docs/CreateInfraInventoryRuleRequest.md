# CreateInfraInventoryRuleRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Actions** | Pointer to [**[]ActionsInner**](ActionsInner.md) |  | [optional] 
**AlertDelay** | Pointer to [**AlertDelay**](AlertDelay.md) |  | [optional] 
**Consumer** | **string** | The name of the application or feature that owns the rule. For example: &#x60;alerts&#x60;, &#x60;apm&#x60;, &#x60;discover&#x60;, &#x60;infrastructure&#x60;, &#x60;logs&#x60;, &#x60;metrics&#x60;, &#x60;ml&#x60;, &#x60;monitoring&#x60;, &#x60;securitySolution&#x60;, &#x60;siem&#x60;, &#x60;stackAlerts&#x60;, or &#x60;uptime&#x60;.  | 
**Enabled** | Pointer to **bool** | Indicates whether you want to run the rule on an interval basis after it is created. | [optional] 
**Name** | **string** | The name of the rule. While this name does not have to be unique, a distinctive name can help you identify a rule.  | 
**NotifyWhen** | Pointer to [**NotifyWhen**](NotifyWhen.md) |  | [optional] 
**Params** | [**ParamsPropertyInfraInventory**](ParamsPropertyInfraInventory.md) |  | 
**RuleTypeId** | **string** | The ID of the rule type that you want to call when the rule is scheduled to run. | 
**Schedule** | [**Schedule**](Schedule.md) |  | 
**Tags** | Pointer to **[]string** |  | [optional] 
**Throttle** | Pointer to **NullableString** | Deprecated in 8.13.0. Use the &#x60;throttle&#x60; property in the action &#x60;frequency&#x60; object instead. The throttle interval, which defines how often an alert generates repeated actions. NOTE: You cannot specify the throttle interval at both the rule and action level. If you set it at the rule level then update the rule in Kibana, it is automatically changed to use action-specific values.  | [optional] 

## Methods

### NewCreateInfraInventoryRuleRequest

`func NewCreateInfraInventoryRuleRequest(consumer string, name string, params ParamsPropertyInfraInventory, ruleTypeId string, schedule Schedule, ) *CreateInfraInventoryRuleRequest`

NewCreateInfraInventoryRuleRequest instantiates a new CreateInfraInventoryRuleRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewCreateInfraInventoryRuleRequestWithDefaults

`func NewCreateInfraInventoryRuleRequestWithDefaults() *CreateInfraInventoryRuleRequest`

NewCreateInfraInventoryRuleRequestWithDefaults instantiates a new CreateInfraInventoryRuleRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetActions

`func (o *CreateInfraInventoryRuleRequest) GetActions() []ActionsInner`

GetActions returns the Actions field if non-nil, zero value otherwise.

### GetActionsOk

`func (o *CreateInfraInventoryRuleRequest) GetActionsOk() (*[]ActionsInner, bool)`

GetActionsOk returns a tuple with the Actions field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetActions

`func (o *CreateInfraInventoryRuleRequest) SetActions(v []ActionsInner)`

SetActions sets Actions field to given value.

### HasActions

`func (o *CreateInfraInventoryRuleRequest) HasActions() bool`

HasActions returns a boolean if a field has been set.

### GetAlertDelay

`func (o *CreateInfraInventoryRuleRequest) GetAlertDelay() AlertDelay`

GetAlertDelay returns the AlertDelay field if non-nil, zero value otherwise.

### GetAlertDelayOk

`func (o *CreateInfraInventoryRuleRequest) GetAlertDelayOk() (*AlertDelay, bool)`

GetAlertDelayOk returns a tuple with the AlertDelay field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAlertDelay

`func (o *CreateInfraInventoryRuleRequest) SetAlertDelay(v AlertDelay)`

SetAlertDelay sets AlertDelay field to given value.

### HasAlertDelay

`func (o *CreateInfraInventoryRuleRequest) HasAlertDelay() bool`

HasAlertDelay returns a boolean if a field has been set.

### GetConsumer

`func (o *CreateInfraInventoryRuleRequest) GetConsumer() string`

GetConsumer returns the Consumer field if non-nil, zero value otherwise.

### GetConsumerOk

`func (o *CreateInfraInventoryRuleRequest) GetConsumerOk() (*string, bool)`

GetConsumerOk returns a tuple with the Consumer field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetConsumer

`func (o *CreateInfraInventoryRuleRequest) SetConsumer(v string)`

SetConsumer sets Consumer field to given value.


### GetEnabled

`func (o *CreateInfraInventoryRuleRequest) GetEnabled() bool`

GetEnabled returns the Enabled field if non-nil, zero value otherwise.

### GetEnabledOk

`func (o *CreateInfraInventoryRuleRequest) GetEnabledOk() (*bool, bool)`

GetEnabledOk returns a tuple with the Enabled field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEnabled

`func (o *CreateInfraInventoryRuleRequest) SetEnabled(v bool)`

SetEnabled sets Enabled field to given value.

### HasEnabled

`func (o *CreateInfraInventoryRuleRequest) HasEnabled() bool`

HasEnabled returns a boolean if a field has been set.

### GetName

`func (o *CreateInfraInventoryRuleRequest) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *CreateInfraInventoryRuleRequest) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *CreateInfraInventoryRuleRequest) SetName(v string)`

SetName sets Name field to given value.


### GetNotifyWhen

`func (o *CreateInfraInventoryRuleRequest) GetNotifyWhen() NotifyWhen`

GetNotifyWhen returns the NotifyWhen field if non-nil, zero value otherwise.

### GetNotifyWhenOk

`func (o *CreateInfraInventoryRuleRequest) GetNotifyWhenOk() (*NotifyWhen, bool)`

GetNotifyWhenOk returns a tuple with the NotifyWhen field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNotifyWhen

`func (o *CreateInfraInventoryRuleRequest) SetNotifyWhen(v NotifyWhen)`

SetNotifyWhen sets NotifyWhen field to given value.

### HasNotifyWhen

`func (o *CreateInfraInventoryRuleRequest) HasNotifyWhen() bool`

HasNotifyWhen returns a boolean if a field has been set.

### GetParams

`func (o *CreateInfraInventoryRuleRequest) GetParams() ParamsPropertyInfraInventory`

GetParams returns the Params field if non-nil, zero value otherwise.

### GetParamsOk

`func (o *CreateInfraInventoryRuleRequest) GetParamsOk() (*ParamsPropertyInfraInventory, bool)`

GetParamsOk returns a tuple with the Params field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetParams

`func (o *CreateInfraInventoryRuleRequest) SetParams(v ParamsPropertyInfraInventory)`

SetParams sets Params field to given value.


### GetRuleTypeId

`func (o *CreateInfraInventoryRuleRequest) GetRuleTypeId() string`

GetRuleTypeId returns the RuleTypeId field if non-nil, zero value otherwise.

### GetRuleTypeIdOk

`func (o *CreateInfraInventoryRuleRequest) GetRuleTypeIdOk() (*string, bool)`

GetRuleTypeIdOk returns a tuple with the RuleTypeId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRuleTypeId

`func (o *CreateInfraInventoryRuleRequest) SetRuleTypeId(v string)`

SetRuleTypeId sets RuleTypeId field to given value.


### GetSchedule

`func (o *CreateInfraInventoryRuleRequest) GetSchedule() Schedule`

GetSchedule returns the Schedule field if non-nil, zero value otherwise.

### GetScheduleOk

`func (o *CreateInfraInventoryRuleRequest) GetScheduleOk() (*Schedule, bool)`

GetScheduleOk returns a tuple with the Schedule field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSchedule

`func (o *CreateInfraInventoryRuleRequest) SetSchedule(v Schedule)`

SetSchedule sets Schedule field to given value.


### GetTags

`func (o *CreateInfraInventoryRuleRequest) GetTags() []string`

GetTags returns the Tags field if non-nil, zero value otherwise.

### GetTagsOk

`func (o *CreateInfraInventoryRuleRequest) GetTagsOk() (*[]string, bool)`

GetTagsOk returns a tuple with the Tags field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTags

`func (o *CreateInfraInventoryRuleRequest) SetTags(v []string)`

SetTags sets Tags field to given value.

### HasTags

`func (o *CreateInfraInventoryRuleRequest) HasTags() bool`

HasTags returns a boolean if a field has been set.

### GetThrottle

`func (o *CreateInfraInventoryRuleRequest) GetThrottle() string`

GetThrottle returns the Throttle field if non-nil, zero value otherwise.

### GetThrottleOk

`func (o *CreateInfraInventoryRuleRequest) GetThrottleOk() (*string, bool)`

GetThrottleOk returns a tuple with the Throttle field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetThrottle

`func (o *CreateInfraInventoryRuleRequest) SetThrottle(v string)`

SetThrottle sets Throttle field to given value.

### HasThrottle

`func (o *CreateInfraInventoryRuleRequest) HasThrottle() bool`

HasThrottle returns a boolean if a field has been set.

### SetThrottleNil

`func (o *CreateInfraInventoryRuleRequest) SetThrottleNil(b bool)`

 SetThrottleNil sets the value for Throttle to be an explicit nil

### UnsetThrottle
`func (o *CreateInfraInventoryRuleRequest) UnsetThrottle()`

UnsetThrottle ensures that no value is present for Throttle, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


