# TimeWindowCalendarAligned

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Duration** | **string** | the duration formatted as {duration}{unit} | 
**Calendar** | [**TimeWindowCalendarAlignedCalendar**](TimeWindowCalendarAlignedCalendar.md) |  | 

## Methods

### NewTimeWindowCalendarAligned

`func NewTimeWindowCalendarAligned(duration string, calendar TimeWindowCalendarAlignedCalendar, ) *TimeWindowCalendarAligned`

NewTimeWindowCalendarAligned instantiates a new TimeWindowCalendarAligned object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewTimeWindowCalendarAlignedWithDefaults

`func NewTimeWindowCalendarAlignedWithDefaults() *TimeWindowCalendarAligned`

NewTimeWindowCalendarAlignedWithDefaults instantiates a new TimeWindowCalendarAligned object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetDuration

`func (o *TimeWindowCalendarAligned) GetDuration() string`

GetDuration returns the Duration field if non-nil, zero value otherwise.

### GetDurationOk

`func (o *TimeWindowCalendarAligned) GetDurationOk() (*string, bool)`

GetDurationOk returns a tuple with the Duration field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDuration

`func (o *TimeWindowCalendarAligned) SetDuration(v string)`

SetDuration sets Duration field to given value.


### GetCalendar

`func (o *TimeWindowCalendarAligned) GetCalendar() TimeWindowCalendarAlignedCalendar`

GetCalendar returns the Calendar field if non-nil, zero value otherwise.

### GetCalendarOk

`func (o *TimeWindowCalendarAligned) GetCalendarOk() (*TimeWindowCalendarAlignedCalendar, bool)`

GetCalendarOk returns a tuple with the Calendar field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCalendar

`func (o *TimeWindowCalendarAligned) SetCalendar(v TimeWindowCalendarAlignedCalendar)`

SetCalendar sets Calendar field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


