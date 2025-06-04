# PostAlertingMaintenanceWindowRequestScheduleCustomRecurring

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**End** | Pointer to **interface{}** | The end date of a recurring schedule, provided in ISO 8601 format and set to the UTC timezone. For example: &#x60;2025-04-01T00:00:00.000Z&#x60;. | [optional] 
**Every** | Pointer to **interface{}** | The interval and frequency of a recurring schedule. It allows values in &#x60;&lt;integer&gt;&lt;unit&gt;&#x60; format. &#x60;&lt;unit&gt;&#x60; is one of &#x60;d&#x60;, &#x60;w&#x60;, &#x60;M&#x60;, or &#x60;y&#x60; for days, weeks, months, years. For example: &#x60;15d&#x60;, &#x60;2w&#x60;, &#x60;3m&#x60;, &#x60;1y&#x60;. | [optional] 
**Occurrences** | Pointer to **interface{}** | The total number of recurrences of the schedule. | [optional] 
**OnMonth** | Pointer to **interface{}** | The specific months for a recurring schedule. Valid values are 1-12. | [optional] 
**OnMonthDay** | Pointer to **interface{}** | The specific days of the month for a recurring schedule. Valid values are 1-31. | [optional] 
**OnWeekDay** | Pointer to **interface{}** | The specific days of the week (&#x60;[MO,TU,WE,TH,FR,SA,SU]&#x60;) or nth day of month (&#x60;[+1MO, -3FR, +2WE, -4SA, -5SU]&#x60;) for a recurring schedule. | [optional] 

## Methods

### NewPostAlertingMaintenanceWindowRequestScheduleCustomRecurring

`func NewPostAlertingMaintenanceWindowRequestScheduleCustomRecurring() *PostAlertingMaintenanceWindowRequestScheduleCustomRecurring`

NewPostAlertingMaintenanceWindowRequestScheduleCustomRecurring instantiates a new PostAlertingMaintenanceWindowRequestScheduleCustomRecurring object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewPostAlertingMaintenanceWindowRequestScheduleCustomRecurringWithDefaults

`func NewPostAlertingMaintenanceWindowRequestScheduleCustomRecurringWithDefaults() *PostAlertingMaintenanceWindowRequestScheduleCustomRecurring`

NewPostAlertingMaintenanceWindowRequestScheduleCustomRecurringWithDefaults instantiates a new PostAlertingMaintenanceWindowRequestScheduleCustomRecurring object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetEnd

`func (o *PostAlertingMaintenanceWindowRequestScheduleCustomRecurring) GetEnd() interface{}`

GetEnd returns the End field if non-nil, zero value otherwise.

### GetEndOk

`func (o *PostAlertingMaintenanceWindowRequestScheduleCustomRecurring) GetEndOk() (*interface{}, bool)`

GetEndOk returns a tuple with the End field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEnd

`func (o *PostAlertingMaintenanceWindowRequestScheduleCustomRecurring) SetEnd(v interface{})`

SetEnd sets End field to given value.

### HasEnd

`func (o *PostAlertingMaintenanceWindowRequestScheduleCustomRecurring) HasEnd() bool`

HasEnd returns a boolean if a field has been set.

### SetEndNil

`func (o *PostAlertingMaintenanceWindowRequestScheduleCustomRecurring) SetEndNil(b bool)`

 SetEndNil sets the value for End to be an explicit nil

### UnsetEnd
`func (o *PostAlertingMaintenanceWindowRequestScheduleCustomRecurring) UnsetEnd()`

UnsetEnd ensures that no value is present for End, not even an explicit nil
### GetEvery

`func (o *PostAlertingMaintenanceWindowRequestScheduleCustomRecurring) GetEvery() interface{}`

GetEvery returns the Every field if non-nil, zero value otherwise.

### GetEveryOk

`func (o *PostAlertingMaintenanceWindowRequestScheduleCustomRecurring) GetEveryOk() (*interface{}, bool)`

GetEveryOk returns a tuple with the Every field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEvery

`func (o *PostAlertingMaintenanceWindowRequestScheduleCustomRecurring) SetEvery(v interface{})`

SetEvery sets Every field to given value.

### HasEvery

`func (o *PostAlertingMaintenanceWindowRequestScheduleCustomRecurring) HasEvery() bool`

HasEvery returns a boolean if a field has been set.

### SetEveryNil

`func (o *PostAlertingMaintenanceWindowRequestScheduleCustomRecurring) SetEveryNil(b bool)`

 SetEveryNil sets the value for Every to be an explicit nil

### UnsetEvery
`func (o *PostAlertingMaintenanceWindowRequestScheduleCustomRecurring) UnsetEvery()`

UnsetEvery ensures that no value is present for Every, not even an explicit nil
### GetOccurrences

`func (o *PostAlertingMaintenanceWindowRequestScheduleCustomRecurring) GetOccurrences() interface{}`

GetOccurrences returns the Occurrences field if non-nil, zero value otherwise.

### GetOccurrencesOk

`func (o *PostAlertingMaintenanceWindowRequestScheduleCustomRecurring) GetOccurrencesOk() (*interface{}, bool)`

GetOccurrencesOk returns a tuple with the Occurrences field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOccurrences

`func (o *PostAlertingMaintenanceWindowRequestScheduleCustomRecurring) SetOccurrences(v interface{})`

SetOccurrences sets Occurrences field to given value.

### HasOccurrences

`func (o *PostAlertingMaintenanceWindowRequestScheduleCustomRecurring) HasOccurrences() bool`

HasOccurrences returns a boolean if a field has been set.

### SetOccurrencesNil

`func (o *PostAlertingMaintenanceWindowRequestScheduleCustomRecurring) SetOccurrencesNil(b bool)`

 SetOccurrencesNil sets the value for Occurrences to be an explicit nil

### UnsetOccurrences
`func (o *PostAlertingMaintenanceWindowRequestScheduleCustomRecurring) UnsetOccurrences()`

UnsetOccurrences ensures that no value is present for Occurrences, not even an explicit nil
### GetOnMonth

`func (o *PostAlertingMaintenanceWindowRequestScheduleCustomRecurring) GetOnMonth() interface{}`

GetOnMonth returns the OnMonth field if non-nil, zero value otherwise.

### GetOnMonthOk

`func (o *PostAlertingMaintenanceWindowRequestScheduleCustomRecurring) GetOnMonthOk() (*interface{}, bool)`

GetOnMonthOk returns a tuple with the OnMonth field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOnMonth

`func (o *PostAlertingMaintenanceWindowRequestScheduleCustomRecurring) SetOnMonth(v interface{})`

SetOnMonth sets OnMonth field to given value.

### HasOnMonth

`func (o *PostAlertingMaintenanceWindowRequestScheduleCustomRecurring) HasOnMonth() bool`

HasOnMonth returns a boolean if a field has been set.

### SetOnMonthNil

`func (o *PostAlertingMaintenanceWindowRequestScheduleCustomRecurring) SetOnMonthNil(b bool)`

 SetOnMonthNil sets the value for OnMonth to be an explicit nil

### UnsetOnMonth
`func (o *PostAlertingMaintenanceWindowRequestScheduleCustomRecurring) UnsetOnMonth()`

UnsetOnMonth ensures that no value is present for OnMonth, not even an explicit nil
### GetOnMonthDay

`func (o *PostAlertingMaintenanceWindowRequestScheduleCustomRecurring) GetOnMonthDay() interface{}`

GetOnMonthDay returns the OnMonthDay field if non-nil, zero value otherwise.

### GetOnMonthDayOk

`func (o *PostAlertingMaintenanceWindowRequestScheduleCustomRecurring) GetOnMonthDayOk() (*interface{}, bool)`

GetOnMonthDayOk returns a tuple with the OnMonthDay field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOnMonthDay

`func (o *PostAlertingMaintenanceWindowRequestScheduleCustomRecurring) SetOnMonthDay(v interface{})`

SetOnMonthDay sets OnMonthDay field to given value.

### HasOnMonthDay

`func (o *PostAlertingMaintenanceWindowRequestScheduleCustomRecurring) HasOnMonthDay() bool`

HasOnMonthDay returns a boolean if a field has been set.

### SetOnMonthDayNil

`func (o *PostAlertingMaintenanceWindowRequestScheduleCustomRecurring) SetOnMonthDayNil(b bool)`

 SetOnMonthDayNil sets the value for OnMonthDay to be an explicit nil

### UnsetOnMonthDay
`func (o *PostAlertingMaintenanceWindowRequestScheduleCustomRecurring) UnsetOnMonthDay()`

UnsetOnMonthDay ensures that no value is present for OnMonthDay, not even an explicit nil
### GetOnWeekDay

`func (o *PostAlertingMaintenanceWindowRequestScheduleCustomRecurring) GetOnWeekDay() interface{}`

GetOnWeekDay returns the OnWeekDay field if non-nil, zero value otherwise.

### GetOnWeekDayOk

`func (o *PostAlertingMaintenanceWindowRequestScheduleCustomRecurring) GetOnWeekDayOk() (*interface{}, bool)`

GetOnWeekDayOk returns a tuple with the OnWeekDay field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOnWeekDay

`func (o *PostAlertingMaintenanceWindowRequestScheduleCustomRecurring) SetOnWeekDay(v interface{})`

SetOnWeekDay sets OnWeekDay field to given value.

### HasOnWeekDay

`func (o *PostAlertingMaintenanceWindowRequestScheduleCustomRecurring) HasOnWeekDay() bool`

HasOnWeekDay returns a boolean if a field has been set.

### SetOnWeekDayNil

`func (o *PostAlertingMaintenanceWindowRequestScheduleCustomRecurring) SetOnWeekDayNil(b bool)`

 SetOnWeekDayNil sets the value for OnWeekDay to be an explicit nil

### UnsetOnWeekDay
`func (o *PostAlertingMaintenanceWindowRequestScheduleCustomRecurring) UnsetOnWeekDay()`

UnsetOnWeekDay ensures that no value is present for OnWeekDay, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


