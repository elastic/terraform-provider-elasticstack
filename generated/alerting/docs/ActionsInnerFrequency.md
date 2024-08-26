# ActionsInnerFrequency

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**NotifyWhen** | [**NotifyWhenAction**](NotifyWhenAction.md) |  | 
**Summary** | **bool** | Indicates whether the action is a summary. | 
**Throttle** | Pointer to **NullableString** | The throttle interval, which defines how often an alert generates repeated actions. It is specified in seconds, minutes, hours, or days and is applicable only if &#x60;notify_when&#x60; is set to &#x60;onThrottleInterval&#x60;. NOTE: You cannot specify the throttle interval at both the rule and action level. The recommended method is to set it for each action. If you set it at the rule level then update the rule in Kibana, it is automatically changed to use action-specific values.  | [optional] 

## Methods

### NewActionsInnerFrequency

`func NewActionsInnerFrequency(notifyWhen NotifyWhenAction, summary bool, ) *ActionsInnerFrequency`

NewActionsInnerFrequency instantiates a new ActionsInnerFrequency object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewActionsInnerFrequencyWithDefaults

`func NewActionsInnerFrequencyWithDefaults() *ActionsInnerFrequency`

NewActionsInnerFrequencyWithDefaults instantiates a new ActionsInnerFrequency object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetNotifyWhen

`func (o *ActionsInnerFrequency) GetNotifyWhen() NotifyWhenAction`

GetNotifyWhen returns the NotifyWhen field if non-nil, zero value otherwise.

### GetNotifyWhenOk

`func (o *ActionsInnerFrequency) GetNotifyWhenOk() (*NotifyWhenAction, bool)`

GetNotifyWhenOk returns a tuple with the NotifyWhen field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNotifyWhen

`func (o *ActionsInnerFrequency) SetNotifyWhen(v NotifyWhenAction)`

SetNotifyWhen sets NotifyWhen field to given value.


### GetSummary

`func (o *ActionsInnerFrequency) GetSummary() bool`

GetSummary returns the Summary field if non-nil, zero value otherwise.

### GetSummaryOk

`func (o *ActionsInnerFrequency) GetSummaryOk() (*bool, bool)`

GetSummaryOk returns a tuple with the Summary field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSummary

`func (o *ActionsInnerFrequency) SetSummary(v bool)`

SetSummary sets Summary field to given value.


### GetThrottle

`func (o *ActionsInnerFrequency) GetThrottle() string`

GetThrottle returns the Throttle field if non-nil, zero value otherwise.

### GetThrottleOk

`func (o *ActionsInnerFrequency) GetThrottleOk() (*string, bool)`

GetThrottleOk returns a tuple with the Throttle field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetThrottle

`func (o *ActionsInnerFrequency) SetThrottle(v string)`

SetThrottle sets Throttle field to given value.

### HasThrottle

`func (o *ActionsInnerFrequency) HasThrottle() bool`

HasThrottle returns a boolean if a field has been set.

### SetThrottleNil

`func (o *ActionsInnerFrequency) SetThrottleNil(b bool)`

 SetThrottleNil sets the value for Throttle to be an explicit nil

### UnsetThrottle
`func (o *ActionsInnerFrequency) UnsetThrottle()`

UnsetThrottle ensures that no value is present for Throttle, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


