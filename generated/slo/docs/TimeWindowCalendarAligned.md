# TimeWindowCalendarAligned

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Duration** | **string** | the duration formatted as {duration}{unit}, accept &#39;1w&#39; (weekly calendar) or &#39;1M&#39; (monthly calendar) only | 
**IsCalendar** | **bool** | Indicates a calendar aligned time window | 

## Methods

### NewTimeWindowCalendarAligned

`func NewTimeWindowCalendarAligned(duration string, isCalendar bool, ) *TimeWindowCalendarAligned`

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


### GetIsCalendar

`func (o *TimeWindowCalendarAligned) GetIsCalendar() bool`

GetIsCalendar returns the IsCalendar field if non-nil, zero value otherwise.

### GetIsCalendarOk

`func (o *TimeWindowCalendarAligned) GetIsCalendarOk() (*bool, bool)`

GetIsCalendarOk returns a tuple with the IsCalendar field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetIsCalendar

`func (o *TimeWindowCalendarAligned) SetIsCalendar(v bool)`

SetIsCalendar sets IsCalendar field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


