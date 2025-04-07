# CreateMaintenanceWindowRequestScheduleCustom

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Duration** | **string** | The duration of the schedule. It allows values in &#x60;&lt;integer&gt;&lt;unit&gt;&#x60; format. &#x60;&lt;unit&gt;&#x60; is one of &#x60;d&#x60;, &#x60;h&#x60;, &#x60;m&#x60;, or &#x60;s&#x60; for hours, minutes, seconds. For example: &#x60;1d&#x60;, &#x60;5h&#x60;, &#x60;30m&#x60;, &#x60;5000s&#x60;. | 
**Recurring** | Pointer to [**CreateMaintenanceWindowRequestScheduleCustomRecurring**](CreateMaintenanceWindowRequestScheduleCustomRecurring.md) |  | [optional] 
**Start** | **string** | The start date and time of the schedule, provided in ISO 8601 format and set to the UTC timezone. For example: &#x60;2025-03-12T12:00:00.000Z&#x60;. | 
**Timezone** | Pointer to **string** | The timezone of the schedule. The default timezone is UTC. | [optional] 

## Methods

### NewCreateMaintenanceWindowRequestScheduleCustom

`func NewCreateMaintenanceWindowRequestScheduleCustom(duration string, start string, ) *CreateMaintenanceWindowRequestScheduleCustom`

NewCreateMaintenanceWindowRequestScheduleCustom instantiates a new CreateMaintenanceWindowRequestScheduleCustom object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewCreateMaintenanceWindowRequestScheduleCustomWithDefaults

`func NewCreateMaintenanceWindowRequestScheduleCustomWithDefaults() *CreateMaintenanceWindowRequestScheduleCustom`

NewCreateMaintenanceWindowRequestScheduleCustomWithDefaults instantiates a new CreateMaintenanceWindowRequestScheduleCustom object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetDuration

`func (o *CreateMaintenanceWindowRequestScheduleCustom) GetDuration() string`

GetDuration returns the Duration field if non-nil, zero value otherwise.

### GetDurationOk

`func (o *CreateMaintenanceWindowRequestScheduleCustom) GetDurationOk() (*string, bool)`

GetDurationOk returns a tuple with the Duration field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDuration

`func (o *CreateMaintenanceWindowRequestScheduleCustom) SetDuration(v string)`

SetDuration sets Duration field to given value.


### GetRecurring

`func (o *CreateMaintenanceWindowRequestScheduleCustom) GetRecurring() CreateMaintenanceWindowRequestScheduleCustomRecurring`

GetRecurring returns the Recurring field if non-nil, zero value otherwise.

### GetRecurringOk

`func (o *CreateMaintenanceWindowRequestScheduleCustom) GetRecurringOk() (*CreateMaintenanceWindowRequestScheduleCustomRecurring, bool)`

GetRecurringOk returns a tuple with the Recurring field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRecurring

`func (o *CreateMaintenanceWindowRequestScheduleCustom) SetRecurring(v CreateMaintenanceWindowRequestScheduleCustomRecurring)`

SetRecurring sets Recurring field to given value.

### HasRecurring

`func (o *CreateMaintenanceWindowRequestScheduleCustom) HasRecurring() bool`

HasRecurring returns a boolean if a field has been set.

### GetStart

`func (o *CreateMaintenanceWindowRequestScheduleCustom) GetStart() string`

GetStart returns the Start field if non-nil, zero value otherwise.

### GetStartOk

`func (o *CreateMaintenanceWindowRequestScheduleCustom) GetStartOk() (*string, bool)`

GetStartOk returns a tuple with the Start field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStart

`func (o *CreateMaintenanceWindowRequestScheduleCustom) SetStart(v string)`

SetStart sets Start field to given value.


### GetTimezone

`func (o *CreateMaintenanceWindowRequestScheduleCustom) GetTimezone() string`

GetTimezone returns the Timezone field if non-nil, zero value otherwise.

### GetTimezoneOk

`func (o *CreateMaintenanceWindowRequestScheduleCustom) GetTimezoneOk() (*string, bool)`

GetTimezoneOk returns a tuple with the Timezone field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTimezone

`func (o *CreateMaintenanceWindowRequestScheduleCustom) SetTimezone(v string)`

SetTimezone sets Timezone field to given value.

### HasTimezone

`func (o *CreateMaintenanceWindowRequestScheduleCustom) HasTimezone() bool`

HasTimezone returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


