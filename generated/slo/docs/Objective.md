# Objective

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Target** | **float64** | the target objective between 0 and 1 excluded | 
**TimesliceTarget** | Pointer to **float64** | the target objective for each slice when using a timeslices budgeting method | [optional] 
**TimesliceWindow** | Pointer to **string** | the duration of each slice when using a timeslices budgeting method, as {duraton}{unit} | [optional] 

## Methods

### NewObjective

`func NewObjective(target float64, ) *Objective`

NewObjective instantiates a new Objective object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewObjectiveWithDefaults

`func NewObjectiveWithDefaults() *Objective`

NewObjectiveWithDefaults instantiates a new Objective object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetTarget

`func (o *Objective) GetTarget() float64`

GetTarget returns the Target field if non-nil, zero value otherwise.

### GetTargetOk

`func (o *Objective) GetTargetOk() (*float64, bool)`

GetTargetOk returns a tuple with the Target field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTarget

`func (o *Objective) SetTarget(v float64)`

SetTarget sets Target field to given value.


### GetTimesliceTarget

`func (o *Objective) GetTimesliceTarget() float64`

GetTimesliceTarget returns the TimesliceTarget field if non-nil, zero value otherwise.

### GetTimesliceTargetOk

`func (o *Objective) GetTimesliceTargetOk() (*float64, bool)`

GetTimesliceTargetOk returns a tuple with the TimesliceTarget field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTimesliceTarget

`func (o *Objective) SetTimesliceTarget(v float64)`

SetTimesliceTarget sets TimesliceTarget field to given value.

### HasTimesliceTarget

`func (o *Objective) HasTimesliceTarget() bool`

HasTimesliceTarget returns a boolean if a field has been set.

### GetTimesliceWindow

`func (o *Objective) GetTimesliceWindow() string`

GetTimesliceWindow returns the TimesliceWindow field if non-nil, zero value otherwise.

### GetTimesliceWindowOk

`func (o *Objective) GetTimesliceWindowOk() (*string, bool)`

GetTimesliceWindowOk returns a tuple with the TimesliceWindow field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTimesliceWindow

`func (o *Objective) SetTimesliceWindow(v string)`

SetTimesliceWindow sets TimesliceWindow field to given value.

### HasTimesliceWindow

`func (o *Objective) HasTimesliceWindow() bool`

HasTimesliceWindow returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


