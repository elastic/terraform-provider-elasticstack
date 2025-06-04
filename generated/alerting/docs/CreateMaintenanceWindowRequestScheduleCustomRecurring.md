# CreateMaintenanceWindowRequestScheduleCustomRecurring

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**End** | Pointer to **string** | The end date of a recurring schedule, provided in ISO 8601 format and set to the UTC timezone. For example: &#x60;2025-04-01T00:00:00.000Z&#x60;. | [optional] 
**Every** | Pointer to **string** | The interval and frequency of a recurring schedule. It allows values in &#x60;&lt;integer&gt;&lt;unit&gt;&#x60; format. &#x60;&lt;unit&gt;&#x60; is one of &#x60;d&#x60;, &#x60;w&#x60;, &#x60;M&#x60;, or &#x60;y&#x60; for days, weeks, months, years. For example: &#x60;15d&#x60;, &#x60;2w&#x60;, &#x60;3m&#x60;, &#x60;1y&#x60;. | [optional] 
**Occurrences** | Pointer to **float32** | The total number of recurrences of the schedule. | [optional] 
**OnMonth** | Pointer to **[]float32** |  | [optional] 
**OnMonthDay** | Pointer to **[]float32** |  | [optional] 
**OnWeekDay** | Pointer to **[]string** |  | [optional] 

## Methods

### NewCreateMaintenanceWindowRequestScheduleCustomRecurring

`func NewCreateMaintenanceWindowRequestScheduleCustomRecurring() *CreateMaintenanceWindowRequestScheduleCustomRecurring`

NewCreateMaintenanceWindowRequestScheduleCustomRecurring instantiates a new CreateMaintenanceWindowRequestScheduleCustomRecurring object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewCreateMaintenanceWindowRequestScheduleCustomRecurringWithDefaults

`func NewCreateMaintenanceWindowRequestScheduleCustomRecurringWithDefaults() *CreateMaintenanceWindowRequestScheduleCustomRecurring`

NewCreateMaintenanceWindowRequestScheduleCustomRecurringWithDefaults instantiates a new CreateMaintenanceWindowRequestScheduleCustomRecurring object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetEnd

`func (o *CreateMaintenanceWindowRequestScheduleCustomRecurring) GetEnd() string`

GetEnd returns the End field if non-nil, zero value otherwise.

### GetEndOk

`func (o *CreateMaintenanceWindowRequestScheduleCustomRecurring) GetEndOk() (*string, bool)`

GetEndOk returns a tuple with the End field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEnd

`func (o *CreateMaintenanceWindowRequestScheduleCustomRecurring) SetEnd(v string)`

SetEnd sets End field to given value.

### HasEnd

`func (o *CreateMaintenanceWindowRequestScheduleCustomRecurring) HasEnd() bool`

HasEnd returns a boolean if a field has been set.

### GetEvery

`func (o *CreateMaintenanceWindowRequestScheduleCustomRecurring) GetEvery() string`

GetEvery returns the Every field if non-nil, zero value otherwise.

### GetEveryOk

`func (o *CreateMaintenanceWindowRequestScheduleCustomRecurring) GetEveryOk() (*string, bool)`

GetEveryOk returns a tuple with the Every field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEvery

`func (o *CreateMaintenanceWindowRequestScheduleCustomRecurring) SetEvery(v string)`

SetEvery sets Every field to given value.

### HasEvery

`func (o *CreateMaintenanceWindowRequestScheduleCustomRecurring) HasEvery() bool`

HasEvery returns a boolean if a field has been set.

### GetOccurrences

`func (o *CreateMaintenanceWindowRequestScheduleCustomRecurring) GetOccurrences() float32`

GetOccurrences returns the Occurrences field if non-nil, zero value otherwise.

### GetOccurrencesOk

`func (o *CreateMaintenanceWindowRequestScheduleCustomRecurring) GetOccurrencesOk() (*float32, bool)`

GetOccurrencesOk returns a tuple with the Occurrences field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOccurrences

`func (o *CreateMaintenanceWindowRequestScheduleCustomRecurring) SetOccurrences(v float32)`

SetOccurrences sets Occurrences field to given value.

### HasOccurrences

`func (o *CreateMaintenanceWindowRequestScheduleCustomRecurring) HasOccurrences() bool`

HasOccurrences returns a boolean if a field has been set.

### GetOnMonth

`func (o *CreateMaintenanceWindowRequestScheduleCustomRecurring) GetOnMonth() []float32`

GetOnMonth returns the OnMonth field if non-nil, zero value otherwise.

### GetOnMonthOk

`func (o *CreateMaintenanceWindowRequestScheduleCustomRecurring) GetOnMonthOk() (*[]float32, bool)`

GetOnMonthOk returns a tuple with the OnMonth field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOnMonth

`func (o *CreateMaintenanceWindowRequestScheduleCustomRecurring) SetOnMonth(v []float32)`

SetOnMonth sets OnMonth field to given value.

### HasOnMonth

`func (o *CreateMaintenanceWindowRequestScheduleCustomRecurring) HasOnMonth() bool`

HasOnMonth returns a boolean if a field has been set.

### GetOnMonthDay

`func (o *CreateMaintenanceWindowRequestScheduleCustomRecurring) GetOnMonthDay() []float32`

GetOnMonthDay returns the OnMonthDay field if non-nil, zero value otherwise.

### GetOnMonthDayOk

`func (o *CreateMaintenanceWindowRequestScheduleCustomRecurring) GetOnMonthDayOk() (*[]float32, bool)`

GetOnMonthDayOk returns a tuple with the OnMonthDay field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOnMonthDay

`func (o *CreateMaintenanceWindowRequestScheduleCustomRecurring) SetOnMonthDay(v []float32)`

SetOnMonthDay sets OnMonthDay field to given value.

### HasOnMonthDay

`func (o *CreateMaintenanceWindowRequestScheduleCustomRecurring) HasOnMonthDay() bool`

HasOnMonthDay returns a boolean if a field has been set.

### GetOnWeekDay

`func (o *CreateMaintenanceWindowRequestScheduleCustomRecurring) GetOnWeekDay() []string`

GetOnWeekDay returns the OnWeekDay field if non-nil, zero value otherwise.

### GetOnWeekDayOk

`func (o *CreateMaintenanceWindowRequestScheduleCustomRecurring) GetOnWeekDayOk() (*[]string, bool)`

GetOnWeekDayOk returns a tuple with the OnWeekDay field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOnWeekDay

`func (o *CreateMaintenanceWindowRequestScheduleCustomRecurring) SetOnWeekDay(v []string)`

SetOnWeekDay sets OnWeekDay field to given value.

### HasOnWeekDay

`func (o *CreateMaintenanceWindowRequestScheduleCustomRecurring) HasOnWeekDay() bool`

HasOnWeekDay returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


