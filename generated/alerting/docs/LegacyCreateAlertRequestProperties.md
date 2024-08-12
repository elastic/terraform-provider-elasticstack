# LegacyCreateAlertRequestProperties

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Actions** | Pointer to [**[]LegacyUpdateAlertRequestPropertiesActionsInner**](LegacyUpdateAlertRequestPropertiesActionsInner.md) |  | [optional] 
**AlertTypeId** | **string** | The ID of the alert type that you want to call when the alert is scheduled to run. | 
**Consumer** | **string** | The name of the application that owns the alert. This name has to match the Kibana feature name, as that dictates the required role-based access control privileges. | 
**Enabled** | Pointer to **bool** | Indicates if you want to run the alert on an interval basis after it is created. | [optional] 
**Name** | **string** | A name to reference and search. | 
**NotifyWhen** | **string** | The condition for throttling the notification. | 
**Params** | **map[string]interface{}** | The parameters to pass to the alert type executor &#x60;params&#x60; value. This will also validate against the alert type params validator, if defined. | 
**Schedule** | [**LegacyUpdateAlertRequestPropertiesSchedule**](LegacyUpdateAlertRequestPropertiesSchedule.md) |  | 
**Tags** | Pointer to **[]string** |  | [optional] 
**Throttle** | Pointer to **string** | How often this alert should fire the same actions. This will prevent the alert from sending out the same notification over and over. For example, if an alert with a schedule of 1 minute stays in a triggered state for 90 minutes, setting a throttle of &#x60;10m&#x60; or &#x60;1h&#x60; will prevent it from sending 90 notifications during this period.  | [optional] 

## Methods

### NewLegacyCreateAlertRequestProperties

`func NewLegacyCreateAlertRequestProperties(alertTypeId string, consumer string, name string, notifyWhen string, params map[string]interface{}, schedule LegacyUpdateAlertRequestPropertiesSchedule, ) *LegacyCreateAlertRequestProperties`

NewLegacyCreateAlertRequestProperties instantiates a new LegacyCreateAlertRequestProperties object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewLegacyCreateAlertRequestPropertiesWithDefaults

`func NewLegacyCreateAlertRequestPropertiesWithDefaults() *LegacyCreateAlertRequestProperties`

NewLegacyCreateAlertRequestPropertiesWithDefaults instantiates a new LegacyCreateAlertRequestProperties object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetActions

`func (o *LegacyCreateAlertRequestProperties) GetActions() []LegacyUpdateAlertRequestPropertiesActionsInner`

GetActions returns the Actions field if non-nil, zero value otherwise.

### GetActionsOk

`func (o *LegacyCreateAlertRequestProperties) GetActionsOk() (*[]LegacyUpdateAlertRequestPropertiesActionsInner, bool)`

GetActionsOk returns a tuple with the Actions field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetActions

`func (o *LegacyCreateAlertRequestProperties) SetActions(v []LegacyUpdateAlertRequestPropertiesActionsInner)`

SetActions sets Actions field to given value.

### HasActions

`func (o *LegacyCreateAlertRequestProperties) HasActions() bool`

HasActions returns a boolean if a field has been set.

### GetAlertTypeId

`func (o *LegacyCreateAlertRequestProperties) GetAlertTypeId() string`

GetAlertTypeId returns the AlertTypeId field if non-nil, zero value otherwise.

### GetAlertTypeIdOk

`func (o *LegacyCreateAlertRequestProperties) GetAlertTypeIdOk() (*string, bool)`

GetAlertTypeIdOk returns a tuple with the AlertTypeId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAlertTypeId

`func (o *LegacyCreateAlertRequestProperties) SetAlertTypeId(v string)`

SetAlertTypeId sets AlertTypeId field to given value.


### GetConsumer

`func (o *LegacyCreateAlertRequestProperties) GetConsumer() string`

GetConsumer returns the Consumer field if non-nil, zero value otherwise.

### GetConsumerOk

`func (o *LegacyCreateAlertRequestProperties) GetConsumerOk() (*string, bool)`

GetConsumerOk returns a tuple with the Consumer field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetConsumer

`func (o *LegacyCreateAlertRequestProperties) SetConsumer(v string)`

SetConsumer sets Consumer field to given value.


### GetEnabled

`func (o *LegacyCreateAlertRequestProperties) GetEnabled() bool`

GetEnabled returns the Enabled field if non-nil, zero value otherwise.

### GetEnabledOk

`func (o *LegacyCreateAlertRequestProperties) GetEnabledOk() (*bool, bool)`

GetEnabledOk returns a tuple with the Enabled field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEnabled

`func (o *LegacyCreateAlertRequestProperties) SetEnabled(v bool)`

SetEnabled sets Enabled field to given value.

### HasEnabled

`func (o *LegacyCreateAlertRequestProperties) HasEnabled() bool`

HasEnabled returns a boolean if a field has been set.

### GetName

`func (o *LegacyCreateAlertRequestProperties) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *LegacyCreateAlertRequestProperties) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *LegacyCreateAlertRequestProperties) SetName(v string)`

SetName sets Name field to given value.


### GetNotifyWhen

`func (o *LegacyCreateAlertRequestProperties) GetNotifyWhen() string`

GetNotifyWhen returns the NotifyWhen field if non-nil, zero value otherwise.

### GetNotifyWhenOk

`func (o *LegacyCreateAlertRequestProperties) GetNotifyWhenOk() (*string, bool)`

GetNotifyWhenOk returns a tuple with the NotifyWhen field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNotifyWhen

`func (o *LegacyCreateAlertRequestProperties) SetNotifyWhen(v string)`

SetNotifyWhen sets NotifyWhen field to given value.


### GetParams

`func (o *LegacyCreateAlertRequestProperties) GetParams() map[string]interface{}`

GetParams returns the Params field if non-nil, zero value otherwise.

### GetParamsOk

`func (o *LegacyCreateAlertRequestProperties) GetParamsOk() (*map[string]interface{}, bool)`

GetParamsOk returns a tuple with the Params field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetParams

`func (o *LegacyCreateAlertRequestProperties) SetParams(v map[string]interface{})`

SetParams sets Params field to given value.


### GetSchedule

`func (o *LegacyCreateAlertRequestProperties) GetSchedule() LegacyUpdateAlertRequestPropertiesSchedule`

GetSchedule returns the Schedule field if non-nil, zero value otherwise.

### GetScheduleOk

`func (o *LegacyCreateAlertRequestProperties) GetScheduleOk() (*LegacyUpdateAlertRequestPropertiesSchedule, bool)`

GetScheduleOk returns a tuple with the Schedule field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSchedule

`func (o *LegacyCreateAlertRequestProperties) SetSchedule(v LegacyUpdateAlertRequestPropertiesSchedule)`

SetSchedule sets Schedule field to given value.


### GetTags

`func (o *LegacyCreateAlertRequestProperties) GetTags() []string`

GetTags returns the Tags field if non-nil, zero value otherwise.

### GetTagsOk

`func (o *LegacyCreateAlertRequestProperties) GetTagsOk() (*[]string, bool)`

GetTagsOk returns a tuple with the Tags field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTags

`func (o *LegacyCreateAlertRequestProperties) SetTags(v []string)`

SetTags sets Tags field to given value.

### HasTags

`func (o *LegacyCreateAlertRequestProperties) HasTags() bool`

HasTags returns a boolean if a field has been set.

### GetThrottle

`func (o *LegacyCreateAlertRequestProperties) GetThrottle() string`

GetThrottle returns the Throttle field if non-nil, zero value otherwise.

### GetThrottleOk

`func (o *LegacyCreateAlertRequestProperties) GetThrottleOk() (*string, bool)`

GetThrottleOk returns a tuple with the Throttle field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetThrottle

`func (o *LegacyCreateAlertRequestProperties) SetThrottle(v string)`

SetThrottle sets Throttle field to given value.

### HasThrottle

`func (o *LegacyCreateAlertRequestProperties) HasThrottle() bool`

HasThrottle returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


