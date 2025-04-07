# UpdateMaintenanceWindowRequestScheduleCustomRecurring

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**End** | Pointer to **string** | The end date of a recurring schedule, provided in ISO 8601 format and set to the UTC timezone. For example: &#x60;2025-04-01T00:00:00.000Z&#x60;. | [optional] 
**Every** | Pointer to **string** | The interval and frequency of a recurring schedule. It allows values in &#x60;&lt;integer&gt;&lt;unit&gt;&#x60; format. &#x60;&lt;unit&gt;&#x60; is one of &#x60;d&#x60;, &#x60;w&#x60;, &#x60;M&#x60;, or &#x60;y&#x60; for days, weeks, months, years. For example: &#x60;15d&#x60;, &#x60;2w&#x60;, &#x60;3m&#x60;, &#x60;1y&#x60;. | [optional] 
**Occurrences** | Pointer to **float32** | The total number of recurrences of the schedule. | [optional] 
**OnMonth** | Pointer to **interface{}** | The specific months for a recurring schedule. Valid values are 1-12. | [optional] 
**OnMonthDay** | Pointer to **interface{}** | The specific days of the month for a recurring schedule. Valid values are 1-31. | [optional] 
**OnWeekDay** | Pointer to **interface{}** | The specific days of the week (&#x60;[MO,TU,WE,TH,FR,SA,SU]&#x60;) or nth day of month (&#x60;[+1MO, -3FR, +2WE, -4SA, -5SU]&#x60;) for a recurring schedule. | [optional] 

## Methods

### NewUpdateMaintenanceWindowRequestScheduleCustomRecurring

`func NewUpdateMaintenanceWindowRequestScheduleCustomRecurring() *UpdateMaintenanceWindowRequestScheduleCustomRecurring`

NewUpdateMaintenanceWindowRequestScheduleCustomRecurring instantiates a new UpdateMaintenanceWindowRequestScheduleCustomRecurring object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewUpdateMaintenanceWindowRequestScheduleCustomRecurringWithDefaults

`func NewUpdateMaintenanceWindowRequestScheduleCustomRecurringWithDefaults() *UpdateMaintenanceWindowRequestScheduleCustomRecurring`

NewUpdateMaintenanceWindowRequestScheduleCustomRecurringWithDefaults instantiates a new UpdateMaintenanceWindowRequestScheduleCustomRecurring object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetEnd

`func (o *UpdateMaintenanceWindowRequestScheduleCustomRecurring) GetEnd() string`

GetEnd returns the End field if non-nil, zero value otherwise.

### GetEndOk

`func (o *UpdateMaintenanceWindowRequestScheduleCustomRecurring) GetEndOk() (*string, bool)`

GetEndOk returns a tuple with the End field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEnd

`func (o *UpdateMaintenanceWindowRequestScheduleCustomRecurring) SetEnd(v string)`

SetEnd sets End field to given value.

### HasEnd

`func (o *UpdateMaintenanceWindowRequestScheduleCustomRecurring) HasEnd() bool`

HasEnd returns a boolean if a field has been set.

### GetEvery

`func (o *UpdateMaintenanceWindowRequestScheduleCustomRecurring) GetEvery() string`

GetEvery returns the Every field if non-nil, zero value otherwise.

### GetEveryOk

`func (o *UpdateMaintenanceWindowRequestScheduleCustomRecurring) GetEveryOk() (*string, bool)`

GetEveryOk returns a tuple with the Every field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEvery

`func (o *UpdateMaintenanceWindowRequestScheduleCustomRecurring) SetEvery(v string)`

SetEvery sets Every field to given value.

### HasEvery

`func (o *UpdateMaintenanceWindowRequestScheduleCustomRecurring) HasEvery() bool`

HasEvery returns a boolean if a field has been set.

### GetOccurrences

`func (o *UpdateMaintenanceWindowRequestScheduleCustomRecurring) GetOccurrences() float32`

GetOccurrences returns the Occurrences field if non-nil, zero value otherwise.

### GetOccurrencesOk

`func (o *UpdateMaintenanceWindowRequestScheduleCustomRecurring) GetOccurrencesOk() (*float32, bool)`

GetOccurrencesOk returns a tuple with the Occurrences field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOccurrences

`func (o *UpdateMaintenanceWindowRequestScheduleCustomRecurring) SetOccurrences(v float32)`

SetOccurrences sets Occurrences field to given value.

### HasOccurrences

`func (o *UpdateMaintenanceWindowRequestScheduleCustomRecurring) HasOccurrences() bool`

HasOccurrences returns a boolean if a field has been set.

### GetOnMonth

`func (o *UpdateMaintenanceWindowRequestScheduleCustomRecurring) GetOnMonth() interface{}`

GetOnMonth returns the OnMonth field if non-nil, zero value otherwise.

### GetOnMonthOk

`func (o *UpdateMaintenanceWindowRequestScheduleCustomRecurring) GetOnMonthOk() (*interface{}, bool)`

GetOnMonthOk returns a tuple with the OnMonth field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOnMonth

`func (o *UpdateMaintenanceWindowRequestScheduleCustomRecurring) SetOnMonth(v interface{})`

SetOnMonth sets OnMonth field to given value.

### HasOnMonth

`func (o *UpdateMaintenanceWindowRequestScheduleCustomRecurring) HasOnMonth() bool`

HasOnMonth returns a boolean if a field has been set.

### SetOnMonthNil

`func (o *UpdateMaintenanceWindowRequestScheduleCustomRecurring) SetOnMonthNil(b bool)`

 SetOnMonthNil sets the value for OnMonth to be an explicit nil

### UnsetOnMonth
`func (o *UpdateMaintenanceWindowRequestScheduleCustomRecurring) UnsetOnMonth()`

UnsetOnMonth ensures that no value is present for OnMonth, not even an explicit nil
### GetOnMonthDay

`func (o *UpdateMaintenanceWindowRequestScheduleCustomRecurring) GetOnMonthDay() interface{}`

GetOnMonthDay returns the OnMonthDay field if non-nil, zero value otherwise.

### GetOnMonthDayOk

`func (o *UpdateMaintenanceWindowRequestScheduleCustomRecurring) GetOnMonthDayOk() (*interface{}, bool)`

GetOnMonthDayOk returns a tuple with the OnMonthDay field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOnMonthDay

`func (o *UpdateMaintenanceWindowRequestScheduleCustomRecurring) SetOnMonthDay(v interface{})`

SetOnMonthDay sets OnMonthDay field to given value.

### HasOnMonthDay

`func (o *UpdateMaintenanceWindowRequestScheduleCustomRecurring) HasOnMonthDay() bool`

HasOnMonthDay returns a boolean if a field has been set.

### SetOnMonthDayNil

`func (o *UpdateMaintenanceWindowRequestScheduleCustomRecurring) SetOnMonthDayNil(b bool)`

 SetOnMonthDayNil sets the value for OnMonthDay to be an explicit nil

### UnsetOnMonthDay
`func (o *UpdateMaintenanceWindowRequestScheduleCustomRecurring) UnsetOnMonthDay()`

UnsetOnMonthDay ensures that no value is present for OnMonthDay, not even an explicit nil
### GetOnWeekDay

`func (o *UpdateMaintenanceWindowRequestScheduleCustomRecurring) GetOnWeekDay() interface{}`

GetOnWeekDay returns the OnWeekDay field if non-nil, zero value otherwise.

### GetOnWeekDayOk

`func (o *UpdateMaintenanceWindowRequestScheduleCustomRecurring) GetOnWeekDayOk() (*interface{}, bool)`

GetOnWeekDayOk returns a tuple with the OnWeekDay field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOnWeekDay

`func (o *UpdateMaintenanceWindowRequestScheduleCustomRecurring) SetOnWeekDay(v interface{})`

SetOnWeekDay sets OnWeekDay field to given value.

### HasOnWeekDay

`func (o *UpdateMaintenanceWindowRequestScheduleCustomRecurring) HasOnWeekDay() bool`

HasOnWeekDay returns a boolean if a field has been set.

### SetOnWeekDayNil

`func (o *UpdateMaintenanceWindowRequestScheduleCustomRecurring) SetOnWeekDayNil(b bool)`

 SetOnWeekDayNil sets the value for OnWeekDay to be an explicit nil

### UnsetOnWeekDay
`func (o *UpdateMaintenanceWindowRequestScheduleCustomRecurring) UnsetOnWeekDay()`

UnsetOnWeekDay ensures that no value is present for OnWeekDay, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


