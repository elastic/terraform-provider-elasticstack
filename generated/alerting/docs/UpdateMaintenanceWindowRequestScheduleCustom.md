# UpdateMaintenanceWindowRequestScheduleCustom

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Duration** | **string** | The duration of the schedule. It allows values in &#x60;&lt;integer&gt;&lt;unit&gt;&#x60; format. &#x60;&lt;unit&gt;&#x60; is one of &#x60;d&#x60;, &#x60;h&#x60;, &#x60;m&#x60;, or &#x60;s&#x60; for hours, minutes, seconds. For example: &#x60;1d&#x60;, &#x60;5h&#x60;, &#x60;30m&#x60;, &#x60;5000s&#x60;. | 
**Recurring** | Pointer to [**UpdateMaintenanceWindowRequestScheduleCustomRecurring**](UpdateMaintenanceWindowRequestScheduleCustomRecurring.md) |  | [optional] 
**Start** | **string** | The start date and time of the schedule, provided in ISO 8601 format and set to the UTC timezone. For example: &#x60;2025-03-12T12:00:00.000Z&#x60;. | 
**Timezone** | Pointer to **string** | The timezone of the schedule. The default timezone is UTC. | [optional] 

## Methods

### NewUpdateMaintenanceWindowRequestScheduleCustom

`func NewUpdateMaintenanceWindowRequestScheduleCustom(duration string, start string, ) *UpdateMaintenanceWindowRequestScheduleCustom`

NewUpdateMaintenanceWindowRequestScheduleCustom instantiates a new UpdateMaintenanceWindowRequestScheduleCustom object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewUpdateMaintenanceWindowRequestScheduleCustomWithDefaults

`func NewUpdateMaintenanceWindowRequestScheduleCustomWithDefaults() *UpdateMaintenanceWindowRequestScheduleCustom`

NewUpdateMaintenanceWindowRequestScheduleCustomWithDefaults instantiates a new UpdateMaintenanceWindowRequestScheduleCustom object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetDuration

`func (o *UpdateMaintenanceWindowRequestScheduleCustom) GetDuration() string`

GetDuration returns the Duration field if non-nil, zero value otherwise.

### GetDurationOk

`func (o *UpdateMaintenanceWindowRequestScheduleCustom) GetDurationOk() (*string, bool)`

GetDurationOk returns a tuple with the Duration field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDuration

`func (o *UpdateMaintenanceWindowRequestScheduleCustom) SetDuration(v string)`

SetDuration sets Duration field to given value.


### GetRecurring

`func (o *UpdateMaintenanceWindowRequestScheduleCustom) GetRecurring() UpdateMaintenanceWindowRequestScheduleCustomRecurring`

GetRecurring returns the Recurring field if non-nil, zero value otherwise.

### GetRecurringOk

`func (o *UpdateMaintenanceWindowRequestScheduleCustom) GetRecurringOk() (*UpdateMaintenanceWindowRequestScheduleCustomRecurring, bool)`

GetRecurringOk returns a tuple with the Recurring field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRecurring

`func (o *UpdateMaintenanceWindowRequestScheduleCustom) SetRecurring(v UpdateMaintenanceWindowRequestScheduleCustomRecurring)`

SetRecurring sets Recurring field to given value.

### HasRecurring

`func (o *UpdateMaintenanceWindowRequestScheduleCustom) HasRecurring() bool`

HasRecurring returns a boolean if a field has been set.

### GetStart

`func (o *UpdateMaintenanceWindowRequestScheduleCustom) GetStart() string`

GetStart returns the Start field if non-nil, zero value otherwise.

### GetStartOk

`func (o *UpdateMaintenanceWindowRequestScheduleCustom) GetStartOk() (*string, bool)`

GetStartOk returns a tuple with the Start field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStart

`func (o *UpdateMaintenanceWindowRequestScheduleCustom) SetStart(v string)`

SetStart sets Start field to given value.


### GetTimezone

`func (o *UpdateMaintenanceWindowRequestScheduleCustom) GetTimezone() string`

GetTimezone returns the Timezone field if non-nil, zero value otherwise.

### GetTimezoneOk

`func (o *UpdateMaintenanceWindowRequestScheduleCustom) GetTimezoneOk() (*string, bool)`

GetTimezoneOk returns a tuple with the Timezone field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTimezone

`func (o *UpdateMaintenanceWindowRequestScheduleCustom) SetTimezone(v string)`

SetTimezone sets Timezone field to given value.

### HasTimezone

`func (o *UpdateMaintenanceWindowRequestScheduleCustom) HasTimezone() bool`

HasTimezone returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


