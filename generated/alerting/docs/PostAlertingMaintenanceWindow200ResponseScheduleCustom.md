# PostAlertingMaintenanceWindow200ResponseScheduleCustom

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Duration** | **interface{}** | The duration of the schedule. It allows values in &#x60;&lt;integer&gt;&lt;unit&gt;&#x60; format. &#x60;&lt;unit&gt;&#x60; is one of &#x60;d&#x60;, &#x60;h&#x60;, &#x60;m&#x60;, or &#x60;s&#x60; for hours, minutes, seconds. For example: &#x60;1d&#x60;, &#x60;5h&#x60;, &#x60;30m&#x60;, &#x60;5000s&#x60;. | 
**Recurring** | Pointer to [**PostAlertingMaintenanceWindow200ResponseScheduleCustomRecurring**](PostAlertingMaintenanceWindow200ResponseScheduleCustomRecurring.md) |  | [optional] 
**Start** | **interface{}** | The start date and time of the schedule, provided in ISO 8601 format and set to the UTC timezone. For example: &#x60;2025-03-12T12:00:00.000Z&#x60;. | 
**Timezone** | Pointer to **interface{}** | The timezone of the schedule. The default timezone is UTC. | [optional] 

## Methods

### NewPostAlertingMaintenanceWindow200ResponseScheduleCustom

`func NewPostAlertingMaintenanceWindow200ResponseScheduleCustom(duration interface{}, start interface{}, ) *PostAlertingMaintenanceWindow200ResponseScheduleCustom`

NewPostAlertingMaintenanceWindow200ResponseScheduleCustom instantiates a new PostAlertingMaintenanceWindow200ResponseScheduleCustom object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewPostAlertingMaintenanceWindow200ResponseScheduleCustomWithDefaults

`func NewPostAlertingMaintenanceWindow200ResponseScheduleCustomWithDefaults() *PostAlertingMaintenanceWindow200ResponseScheduleCustom`

NewPostAlertingMaintenanceWindow200ResponseScheduleCustomWithDefaults instantiates a new PostAlertingMaintenanceWindow200ResponseScheduleCustom object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetDuration

`func (o *PostAlertingMaintenanceWindow200ResponseScheduleCustom) GetDuration() interface{}`

GetDuration returns the Duration field if non-nil, zero value otherwise.

### GetDurationOk

`func (o *PostAlertingMaintenanceWindow200ResponseScheduleCustom) GetDurationOk() (*interface{}, bool)`

GetDurationOk returns a tuple with the Duration field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDuration

`func (o *PostAlertingMaintenanceWindow200ResponseScheduleCustom) SetDuration(v interface{})`

SetDuration sets Duration field to given value.


### SetDurationNil

`func (o *PostAlertingMaintenanceWindow200ResponseScheduleCustom) SetDurationNil(b bool)`

 SetDurationNil sets the value for Duration to be an explicit nil

### UnsetDuration
`func (o *PostAlertingMaintenanceWindow200ResponseScheduleCustom) UnsetDuration()`

UnsetDuration ensures that no value is present for Duration, not even an explicit nil
### GetRecurring

`func (o *PostAlertingMaintenanceWindow200ResponseScheduleCustom) GetRecurring() PostAlertingMaintenanceWindow200ResponseScheduleCustomRecurring`

GetRecurring returns the Recurring field if non-nil, zero value otherwise.

### GetRecurringOk

`func (o *PostAlertingMaintenanceWindow200ResponseScheduleCustom) GetRecurringOk() (*PostAlertingMaintenanceWindow200ResponseScheduleCustomRecurring, bool)`

GetRecurringOk returns a tuple with the Recurring field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRecurring

`func (o *PostAlertingMaintenanceWindow200ResponseScheduleCustom) SetRecurring(v PostAlertingMaintenanceWindow200ResponseScheduleCustomRecurring)`

SetRecurring sets Recurring field to given value.

### HasRecurring

`func (o *PostAlertingMaintenanceWindow200ResponseScheduleCustom) HasRecurring() bool`

HasRecurring returns a boolean if a field has been set.

### GetStart

`func (o *PostAlertingMaintenanceWindow200ResponseScheduleCustom) GetStart() interface{}`

GetStart returns the Start field if non-nil, zero value otherwise.

### GetStartOk

`func (o *PostAlertingMaintenanceWindow200ResponseScheduleCustom) GetStartOk() (*interface{}, bool)`

GetStartOk returns a tuple with the Start field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStart

`func (o *PostAlertingMaintenanceWindow200ResponseScheduleCustom) SetStart(v interface{})`

SetStart sets Start field to given value.


### SetStartNil

`func (o *PostAlertingMaintenanceWindow200ResponseScheduleCustom) SetStartNil(b bool)`

 SetStartNil sets the value for Start to be an explicit nil

### UnsetStart
`func (o *PostAlertingMaintenanceWindow200ResponseScheduleCustom) UnsetStart()`

UnsetStart ensures that no value is present for Start, not even an explicit nil
### GetTimezone

`func (o *PostAlertingMaintenanceWindow200ResponseScheduleCustom) GetTimezone() interface{}`

GetTimezone returns the Timezone field if non-nil, zero value otherwise.

### GetTimezoneOk

`func (o *PostAlertingMaintenanceWindow200ResponseScheduleCustom) GetTimezoneOk() (*interface{}, bool)`

GetTimezoneOk returns a tuple with the Timezone field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTimezone

`func (o *PostAlertingMaintenanceWindow200ResponseScheduleCustom) SetTimezone(v interface{})`

SetTimezone sets Timezone field to given value.

### HasTimezone

`func (o *PostAlertingMaintenanceWindow200ResponseScheduleCustom) HasTimezone() bool`

HasTimezone returns a boolean if a field has been set.

### SetTimezoneNil

`func (o *PostAlertingMaintenanceWindow200ResponseScheduleCustom) SetTimezoneNil(b bool)`

 SetTimezoneNil sets the value for Timezone to be an explicit nil

### UnsetTimezone
`func (o *PostAlertingMaintenanceWindow200ResponseScheduleCustom) UnsetTimezone()`

UnsetTimezone ensures that no value is present for Timezone, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


