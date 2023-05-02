# TimeWindowRolling

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Duration** | **string** | the duration formatted as {duration}{unit} | 
**IsRolling** | **bool** | Indicates a rolling time window | 

## Methods

### NewTimeWindowRolling

`func NewTimeWindowRolling(duration string, isRolling bool, ) *TimeWindowRolling`

NewTimeWindowRolling instantiates a new TimeWindowRolling object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewTimeWindowRollingWithDefaults

`func NewTimeWindowRollingWithDefaults() *TimeWindowRolling`

NewTimeWindowRollingWithDefaults instantiates a new TimeWindowRolling object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetDuration

`func (o *TimeWindowRolling) GetDuration() string`

GetDuration returns the Duration field if non-nil, zero value otherwise.

### GetDurationOk

`func (o *TimeWindowRolling) GetDurationOk() (*string, bool)`

GetDurationOk returns a tuple with the Duration field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDuration

`func (o *TimeWindowRolling) SetDuration(v string)`

SetDuration sets Duration field to given value.


### GetIsRolling

`func (o *TimeWindowRolling) GetIsRolling() bool`

GetIsRolling returns the IsRolling field if non-nil, zero value otherwise.

### GetIsRollingOk

`func (o *TimeWindowRolling) GetIsRollingOk() (*bool, bool)`

GetIsRollingOk returns a tuple with the IsRolling field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetIsRolling

`func (o *TimeWindowRolling) SetIsRolling(v bool)`

SetIsRolling sets IsRolling field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


