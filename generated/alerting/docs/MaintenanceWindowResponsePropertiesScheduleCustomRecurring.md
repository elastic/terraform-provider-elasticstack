# MaintenanceWindowResponsePropertiesScheduleCustomRecurring

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

### NewMaintenanceWindowResponsePropertiesScheduleCustomRecurring

`func NewMaintenanceWindowResponsePropertiesScheduleCustomRecurring() *MaintenanceWindowResponsePropertiesScheduleCustomRecurring`

NewMaintenanceWindowResponsePropertiesScheduleCustomRecurring instantiates a new MaintenanceWindowResponsePropertiesScheduleCustomRecurring object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewMaintenanceWindowResponsePropertiesScheduleCustomRecurringWithDefaults

`func NewMaintenanceWindowResponsePropertiesScheduleCustomRecurringWithDefaults() *MaintenanceWindowResponsePropertiesScheduleCustomRecurring`

NewMaintenanceWindowResponsePropertiesScheduleCustomRecurringWithDefaults instantiates a new MaintenanceWindowResponsePropertiesScheduleCustomRecurring object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetEnd

`func (o *MaintenanceWindowResponsePropertiesScheduleCustomRecurring) GetEnd() string`

GetEnd returns the End field if non-nil, zero value otherwise.

### GetEndOk

`func (o *MaintenanceWindowResponsePropertiesScheduleCustomRecurring) GetEndOk() (*string, bool)`

GetEndOk returns a tuple with the End field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEnd

`func (o *MaintenanceWindowResponsePropertiesScheduleCustomRecurring) SetEnd(v string)`

SetEnd sets End field to given value.

### HasEnd

`func (o *MaintenanceWindowResponsePropertiesScheduleCustomRecurring) HasEnd() bool`

HasEnd returns a boolean if a field has been set.

### GetEvery

`func (o *MaintenanceWindowResponsePropertiesScheduleCustomRecurring) GetEvery() string`

GetEvery returns the Every field if non-nil, zero value otherwise.

### GetEveryOk

`func (o *MaintenanceWindowResponsePropertiesScheduleCustomRecurring) GetEveryOk() (*string, bool)`

GetEveryOk returns a tuple with the Every field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEvery

`func (o *MaintenanceWindowResponsePropertiesScheduleCustomRecurring) SetEvery(v string)`

SetEvery sets Every field to given value.

### HasEvery

`func (o *MaintenanceWindowResponsePropertiesScheduleCustomRecurring) HasEvery() bool`

HasEvery returns a boolean if a field has been set.

### GetOccurrences

`func (o *MaintenanceWindowResponsePropertiesScheduleCustomRecurring) GetOccurrences() float32`

GetOccurrences returns the Occurrences field if non-nil, zero value otherwise.

### GetOccurrencesOk

`func (o *MaintenanceWindowResponsePropertiesScheduleCustomRecurring) GetOccurrencesOk() (*float32, bool)`

GetOccurrencesOk returns a tuple with the Occurrences field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOccurrences

`func (o *MaintenanceWindowResponsePropertiesScheduleCustomRecurring) SetOccurrences(v float32)`

SetOccurrences sets Occurrences field to given value.

### HasOccurrences

`func (o *MaintenanceWindowResponsePropertiesScheduleCustomRecurring) HasOccurrences() bool`

HasOccurrences returns a boolean if a field has been set.

### GetOnMonth

`func (o *MaintenanceWindowResponsePropertiesScheduleCustomRecurring) GetOnMonth() []float32`

GetOnMonth returns the OnMonth field if non-nil, zero value otherwise.

### GetOnMonthOk

`func (o *MaintenanceWindowResponsePropertiesScheduleCustomRecurring) GetOnMonthOk() (*[]float32, bool)`

GetOnMonthOk returns a tuple with the OnMonth field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOnMonth

`func (o *MaintenanceWindowResponsePropertiesScheduleCustomRecurring) SetOnMonth(v []float32)`

SetOnMonth sets OnMonth field to given value.

### HasOnMonth

`func (o *MaintenanceWindowResponsePropertiesScheduleCustomRecurring) HasOnMonth() bool`

HasOnMonth returns a boolean if a field has been set.

### GetOnMonthDay

`func (o *MaintenanceWindowResponsePropertiesScheduleCustomRecurring) GetOnMonthDay() []float32`

GetOnMonthDay returns the OnMonthDay field if non-nil, zero value otherwise.

### GetOnMonthDayOk

`func (o *MaintenanceWindowResponsePropertiesScheduleCustomRecurring) GetOnMonthDayOk() (*[]float32, bool)`

GetOnMonthDayOk returns a tuple with the OnMonthDay field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOnMonthDay

`func (o *MaintenanceWindowResponsePropertiesScheduleCustomRecurring) SetOnMonthDay(v []float32)`

SetOnMonthDay sets OnMonthDay field to given value.

### HasOnMonthDay

`func (o *MaintenanceWindowResponsePropertiesScheduleCustomRecurring) HasOnMonthDay() bool`

HasOnMonthDay returns a boolean if a field has been set.

### GetOnWeekDay

`func (o *MaintenanceWindowResponsePropertiesScheduleCustomRecurring) GetOnWeekDay() []string`

GetOnWeekDay returns the OnWeekDay field if non-nil, zero value otherwise.

### GetOnWeekDayOk

`func (o *MaintenanceWindowResponsePropertiesScheduleCustomRecurring) GetOnWeekDayOk() (*[]string, bool)`

GetOnWeekDayOk returns a tuple with the OnWeekDay field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOnWeekDay

`func (o *MaintenanceWindowResponsePropertiesScheduleCustomRecurring) SetOnWeekDay(v []string)`

SetOnWeekDay sets OnWeekDay field to given value.

### HasOnWeekDay

`func (o *MaintenanceWindowResponsePropertiesScheduleCustomRecurring) HasOnWeekDay() bool`

HasOnWeekDay returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


