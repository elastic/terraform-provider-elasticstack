# LegacyUpdateAlertRequestProperties

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Actions** | Pointer to [**[]LegacyUpdateAlertRequestPropertiesActionsInner**](LegacyUpdateAlertRequestPropertiesActionsInner.md) |  | [optional] 
**Name** | **string** | A name to reference and search. | 
**NotifyWhen** | **string** | The condition for throttling the notification. | 
**Params** | **map[string]interface{}** | The parameters to pass to the alert type executor &#x60;params&#x60; value. This will also validate against the alert type params validator, if defined. | 
**Schedule** | [**LegacyUpdateAlertRequestPropertiesSchedule**](LegacyUpdateAlertRequestPropertiesSchedule.md) |  | 
**Tags** | Pointer to **[]string** |  | [optional] 
**Throttle** | Pointer to **string** | How often this alert should fire the same actions. This will prevent the alert from sending out the same notification over and over. For example, if an alert with a schedule of 1 minute stays in a triggered state for 90 minutes, setting a throttle of &#x60;10m&#x60; or &#x60;1h&#x60; will prevent it from sending 90 notifications during this period.  | [optional] 

## Methods

### NewLegacyUpdateAlertRequestProperties

`func NewLegacyUpdateAlertRequestProperties(name string, notifyWhen string, params map[string]interface{}, schedule LegacyUpdateAlertRequestPropertiesSchedule, ) *LegacyUpdateAlertRequestProperties`

NewLegacyUpdateAlertRequestProperties instantiates a new LegacyUpdateAlertRequestProperties object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewLegacyUpdateAlertRequestPropertiesWithDefaults

`func NewLegacyUpdateAlertRequestPropertiesWithDefaults() *LegacyUpdateAlertRequestProperties`

NewLegacyUpdateAlertRequestPropertiesWithDefaults instantiates a new LegacyUpdateAlertRequestProperties object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetActions

`func (o *LegacyUpdateAlertRequestProperties) GetActions() []LegacyUpdateAlertRequestPropertiesActionsInner`

GetActions returns the Actions field if non-nil, zero value otherwise.

### GetActionsOk

`func (o *LegacyUpdateAlertRequestProperties) GetActionsOk() (*[]LegacyUpdateAlertRequestPropertiesActionsInner, bool)`

GetActionsOk returns a tuple with the Actions field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetActions

`func (o *LegacyUpdateAlertRequestProperties) SetActions(v []LegacyUpdateAlertRequestPropertiesActionsInner)`

SetActions sets Actions field to given value.

### HasActions

`func (o *LegacyUpdateAlertRequestProperties) HasActions() bool`

HasActions returns a boolean if a field has been set.

### GetName

`func (o *LegacyUpdateAlertRequestProperties) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *LegacyUpdateAlertRequestProperties) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *LegacyUpdateAlertRequestProperties) SetName(v string)`

SetName sets Name field to given value.


### GetNotifyWhen

`func (o *LegacyUpdateAlertRequestProperties) GetNotifyWhen() string`

GetNotifyWhen returns the NotifyWhen field if non-nil, zero value otherwise.

### GetNotifyWhenOk

`func (o *LegacyUpdateAlertRequestProperties) GetNotifyWhenOk() (*string, bool)`

GetNotifyWhenOk returns a tuple with the NotifyWhen field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNotifyWhen

`func (o *LegacyUpdateAlertRequestProperties) SetNotifyWhen(v string)`

SetNotifyWhen sets NotifyWhen field to given value.


### GetParams

`func (o *LegacyUpdateAlertRequestProperties) GetParams() map[string]interface{}`

GetParams returns the Params field if non-nil, zero value otherwise.

### GetParamsOk

`func (o *LegacyUpdateAlertRequestProperties) GetParamsOk() (*map[string]interface{}, bool)`

GetParamsOk returns a tuple with the Params field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetParams

`func (o *LegacyUpdateAlertRequestProperties) SetParams(v map[string]interface{})`

SetParams sets Params field to given value.


### GetSchedule

`func (o *LegacyUpdateAlertRequestProperties) GetSchedule() LegacyUpdateAlertRequestPropertiesSchedule`

GetSchedule returns the Schedule field if non-nil, zero value otherwise.

### GetScheduleOk

`func (o *LegacyUpdateAlertRequestProperties) GetScheduleOk() (*LegacyUpdateAlertRequestPropertiesSchedule, bool)`

GetScheduleOk returns a tuple with the Schedule field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSchedule

`func (o *LegacyUpdateAlertRequestProperties) SetSchedule(v LegacyUpdateAlertRequestPropertiesSchedule)`

SetSchedule sets Schedule field to given value.


### GetTags

`func (o *LegacyUpdateAlertRequestProperties) GetTags() []string`

GetTags returns the Tags field if non-nil, zero value otherwise.

### GetTagsOk

`func (o *LegacyUpdateAlertRequestProperties) GetTagsOk() (*[]string, bool)`

GetTagsOk returns a tuple with the Tags field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTags

`func (o *LegacyUpdateAlertRequestProperties) SetTags(v []string)`

SetTags sets Tags field to given value.

### HasTags

`func (o *LegacyUpdateAlertRequestProperties) HasTags() bool`

HasTags returns a boolean if a field has been set.

### GetThrottle

`func (o *LegacyUpdateAlertRequestProperties) GetThrottle() string`

GetThrottle returns the Throttle field if non-nil, zero value otherwise.

### GetThrottleOk

`func (o *LegacyUpdateAlertRequestProperties) GetThrottleOk() (*string, bool)`

GetThrottleOk returns a tuple with the Throttle field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetThrottle

`func (o *LegacyUpdateAlertRequestProperties) SetThrottle(v string)`

SetThrottle sets Throttle field to given value.

### HasThrottle

`func (o *LegacyUpdateAlertRequestProperties) HasThrottle() bool`

HasThrottle returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


