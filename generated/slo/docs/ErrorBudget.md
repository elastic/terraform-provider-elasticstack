# ErrorBudget

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Initial** | Pointer to **float32** | The initial error budget, as 1 - objective | [optional] 
**Consumed** | Pointer to **float32** | The error budget consummed, as a percentage of the initial value. | [optional] 
**Remaining** | Pointer to **float32** | The error budget remaining, as a percentage of the initial value. | [optional] 
**IsEstimated** | Pointer to **bool** | Only for SLO defined with occurrences budgeting method and calendar aligned time window. | [optional] 

## Methods

### NewErrorBudget

`func NewErrorBudget() *ErrorBudget`

NewErrorBudget instantiates a new ErrorBudget object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewErrorBudgetWithDefaults

`func NewErrorBudgetWithDefaults() *ErrorBudget`

NewErrorBudgetWithDefaults instantiates a new ErrorBudget object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetInitial

`func (o *ErrorBudget) GetInitial() float32`

GetInitial returns the Initial field if non-nil, zero value otherwise.

### GetInitialOk

`func (o *ErrorBudget) GetInitialOk() (*float32, bool)`

GetInitialOk returns a tuple with the Initial field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetInitial

`func (o *ErrorBudget) SetInitial(v float32)`

SetInitial sets Initial field to given value.

### HasInitial

`func (o *ErrorBudget) HasInitial() bool`

HasInitial returns a boolean if a field has been set.

### GetConsumed

`func (o *ErrorBudget) GetConsumed() float32`

GetConsumed returns the Consumed field if non-nil, zero value otherwise.

### GetConsumedOk

`func (o *ErrorBudget) GetConsumedOk() (*float32, bool)`

GetConsumedOk returns a tuple with the Consumed field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetConsumed

`func (o *ErrorBudget) SetConsumed(v float32)`

SetConsumed sets Consumed field to given value.

### HasConsumed

`func (o *ErrorBudget) HasConsumed() bool`

HasConsumed returns a boolean if a field has been set.

### GetRemaining

`func (o *ErrorBudget) GetRemaining() float32`

GetRemaining returns the Remaining field if non-nil, zero value otherwise.

### GetRemainingOk

`func (o *ErrorBudget) GetRemainingOk() (*float32, bool)`

GetRemainingOk returns a tuple with the Remaining field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRemaining

`func (o *ErrorBudget) SetRemaining(v float32)`

SetRemaining sets Remaining field to given value.

### HasRemaining

`func (o *ErrorBudget) HasRemaining() bool`

HasRemaining returns a boolean if a field has been set.

### GetIsEstimated

`func (o *ErrorBudget) GetIsEstimated() bool`

GetIsEstimated returns the IsEstimated field if non-nil, zero value otherwise.

### GetIsEstimatedOk

`func (o *ErrorBudget) GetIsEstimatedOk() (*bool, bool)`

GetIsEstimatedOk returns a tuple with the IsEstimated field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetIsEstimated

`func (o *ErrorBudget) SetIsEstimated(v bool)`

SetIsEstimated sets IsEstimated field to given value.

### HasIsEstimated

`func (o *ErrorBudget) HasIsEstimated() bool`

HasIsEstimated returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


