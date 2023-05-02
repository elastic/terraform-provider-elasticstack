# Objective

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Target** | **float32** | the target objective between 0 and 1 excluded | 
**TimeslicesTarget** | Pointer to **float32** | the target objective for each slice when using a timeslices budgeting method | [optional] 
**TimeslicesWindow** | Pointer to **string** | the duration of each slice when using a timeslices budgeting method, as {duraton}{unit} | [optional] 

## Methods

### NewObjective

`func NewObjective(target float32, ) *Objective`

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

`func (o *Objective) GetTarget() float32`

GetTarget returns the Target field if non-nil, zero value otherwise.

### GetTargetOk

`func (o *Objective) GetTargetOk() (*float32, bool)`

GetTargetOk returns a tuple with the Target field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTarget

`func (o *Objective) SetTarget(v float32)`

SetTarget sets Target field to given value.


### GetTimeslicesTarget

`func (o *Objective) GetTimeslicesTarget() float32`

GetTimeslicesTarget returns the TimeslicesTarget field if non-nil, zero value otherwise.

### GetTimeslicesTargetOk

`func (o *Objective) GetTimeslicesTargetOk() (*float32, bool)`

GetTimeslicesTargetOk returns a tuple with the TimeslicesTarget field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTimeslicesTarget

`func (o *Objective) SetTimeslicesTarget(v float32)`

SetTimeslicesTarget sets TimeslicesTarget field to given value.

### HasTimeslicesTarget

`func (o *Objective) HasTimeslicesTarget() bool`

HasTimeslicesTarget returns a boolean if a field has been set.

### GetTimeslicesWindow

`func (o *Objective) GetTimeslicesWindow() string`

GetTimeslicesWindow returns the TimeslicesWindow field if non-nil, zero value otherwise.

### GetTimeslicesWindowOk

`func (o *Objective) GetTimeslicesWindowOk() (*string, bool)`

GetTimeslicesWindowOk returns a tuple with the TimeslicesWindow field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTimeslicesWindow

`func (o *Objective) SetTimeslicesWindow(v string)`

SetTimeslicesWindow sets TimeslicesWindow field to given value.

### HasTimeslicesWindow

`func (o *Objective) HasTimeslicesWindow() bool`

HasTimeslicesWindow returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


