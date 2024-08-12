# ActionsInnerAlertsFilterTimeframe

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Days** | Pointer to **[]int32** |  | [optional] 
**Hours** | Pointer to [**ActionsInnerAlertsFilterTimeframeHours**](ActionsInnerAlertsFilterTimeframeHours.md) |  | [optional] 
**Timezone** | Pointer to **string** | The ISO time zone for the &#x60;hours&#x60; values. Values such as &#x60;UTC&#x60; and &#x60;UTC+1&#x60; also work but lack built-in daylight savings time support and are not recommended.  | [optional] 

## Methods

### NewActionsInnerAlertsFilterTimeframe

`func NewActionsInnerAlertsFilterTimeframe() *ActionsInnerAlertsFilterTimeframe`

NewActionsInnerAlertsFilterTimeframe instantiates a new ActionsInnerAlertsFilterTimeframe object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewActionsInnerAlertsFilterTimeframeWithDefaults

`func NewActionsInnerAlertsFilterTimeframeWithDefaults() *ActionsInnerAlertsFilterTimeframe`

NewActionsInnerAlertsFilterTimeframeWithDefaults instantiates a new ActionsInnerAlertsFilterTimeframe object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetDays

`func (o *ActionsInnerAlertsFilterTimeframe) GetDays() []int32`

GetDays returns the Days field if non-nil, zero value otherwise.

### GetDaysOk

`func (o *ActionsInnerAlertsFilterTimeframe) GetDaysOk() (*[]int32, bool)`

GetDaysOk returns a tuple with the Days field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDays

`func (o *ActionsInnerAlertsFilterTimeframe) SetDays(v []int32)`

SetDays sets Days field to given value.

### HasDays

`func (o *ActionsInnerAlertsFilterTimeframe) HasDays() bool`

HasDays returns a boolean if a field has been set.

### GetHours

`func (o *ActionsInnerAlertsFilterTimeframe) GetHours() ActionsInnerAlertsFilterTimeframeHours`

GetHours returns the Hours field if non-nil, zero value otherwise.

### GetHoursOk

`func (o *ActionsInnerAlertsFilterTimeframe) GetHoursOk() (*ActionsInnerAlertsFilterTimeframeHours, bool)`

GetHoursOk returns a tuple with the Hours field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetHours

`func (o *ActionsInnerAlertsFilterTimeframe) SetHours(v ActionsInnerAlertsFilterTimeframeHours)`

SetHours sets Hours field to given value.

### HasHours

`func (o *ActionsInnerAlertsFilterTimeframe) HasHours() bool`

HasHours returns a boolean if a field has been set.

### GetTimezone

`func (o *ActionsInnerAlertsFilterTimeframe) GetTimezone() string`

GetTimezone returns the Timezone field if non-nil, zero value otherwise.

### GetTimezoneOk

`func (o *ActionsInnerAlertsFilterTimeframe) GetTimezoneOk() (*string, bool)`

GetTimezoneOk returns a tuple with the Timezone field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTimezone

`func (o *ActionsInnerAlertsFilterTimeframe) SetTimezone(v string)`

SetTimezone sets Timezone field to given value.

### HasTimezone

`func (o *ActionsInnerAlertsFilterTimeframe) HasTimezone() bool`

HasTimezone returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


