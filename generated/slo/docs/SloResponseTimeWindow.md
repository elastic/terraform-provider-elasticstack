# SloResponseTimeWindow

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Duration** | **string** | the duration formatted as {duration}{unit} | 
**IsRolling** | **bool** | Indicates a rolling time window | 
**Calendar** | [**TimeWindowCalendarAlignedCalendar**](TimeWindowCalendarAlignedCalendar.md) |  | 

## Methods

### NewSloResponseTimeWindow

`func NewSloResponseTimeWindow(duration string, isRolling bool, calendar TimeWindowCalendarAlignedCalendar, ) *SloResponseTimeWindow`

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


### GetCalendar

`func (o *SloResponseTimeWindow) GetCalendar() TimeWindowCalendarAlignedCalendar`

GetCalendar returns the Calendar field if non-nil, zero value otherwise.

### GetCalendarOk

`func (o *SloResponseTimeWindow) GetCalendarOk() (*TimeWindowCalendarAlignedCalendar, bool)`

GetCalendarOk returns a tuple with the Calendar field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCalendar

`func (o *SloResponseTimeWindow) SetCalendar(v TimeWindowCalendarAlignedCalendar)`

SetCalendar sets Calendar field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


