# SloResponseTimeWindow

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Duration** | **string** | the duration formatted as {duration}{unit}, accept &#39;1w&#39; (weekly calendar) or &#39;1M&#39; (monthly calendar) only | 
**IsRolling** | **bool** | Indicates a rolling time window | 
**IsCalendar** | **bool** | Indicates a calendar aligned time window | 

## Methods

### NewSloResponseTimeWindow

`func NewSloResponseTimeWindow(duration string, isRolling bool, isCalendar bool, ) *SloResponseTimeWindow`

NewSloResponseTimeWindow instantiates a new SloResponseTimeWindow object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewSloResponseTimeWindowWithDefaults

`func NewSloResponseTimeWindowWithDefaults() *SloResponseTimeWindow`

NewSloResponseTimeWindowWithDefaults instantiates a new SloResponseTimeWindow object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetDuration

`func (o *SloResponseTimeWindow) GetDuration() string`

GetDuration returns the Duration field if non-nil, zero value otherwise.

### GetDurationOk

`func (o *SloResponseTimeWindow) GetDurationOk() (*string, bool)`

GetDurationOk returns a tuple with the Duration field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDuration

`func (o *SloResponseTimeWindow) SetDuration(v string)`

SetDuration sets Duration field to given value.


### GetIsRolling

`func (o *SloResponseTimeWindow) GetIsRolling() bool`

GetIsRolling returns the IsRolling field if non-nil, zero value otherwise.

### GetIsRollingOk

`func (o *SloResponseTimeWindow) GetIsRollingOk() (*bool, bool)`

GetIsRollingOk returns a tuple with the IsRolling field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetIsRolling

`func (o *SloResponseTimeWindow) SetIsRolling(v bool)`

SetIsRolling sets IsRolling field to given value.


### GetIsCalendar

`func (o *SloResponseTimeWindow) GetIsCalendar() bool`

GetIsCalendar returns the IsCalendar field if non-nil, zero value otherwise.

### GetIsCalendarOk

`func (o *SloResponseTimeWindow) GetIsCalendarOk() (*bool, bool)`

GetIsCalendarOk returns a tuple with the IsCalendar field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetIsCalendar

`func (o *SloResponseTimeWindow) SetIsCalendar(v bool)`

SetIsCalendar sets IsCalendar field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


