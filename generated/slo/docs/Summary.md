# Summary

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Status** | Pointer to **string** |  | [optional] 
**SliValue** | Pointer to **float32** |  | [optional] 
**ErrorBudget** | Pointer to [**ErrorBudget**](ErrorBudget.md) |  | [optional] 

## Methods

### NewSummary

`func NewSummary() *Summary`

NewSummary instantiates a new Summary object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewSummaryWithDefaults

`func NewSummaryWithDefaults() *Summary`

NewSummaryWithDefaults instantiates a new Summary object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetStatus

`func (o *Summary) GetStatus() string`

GetStatus returns the Status field if non-nil, zero value otherwise.

### GetStatusOk

`func (o *Summary) GetStatusOk() (*string, bool)`

GetStatusOk returns a tuple with the Status field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStatus

`func (o *Summary) SetStatus(v string)`

SetStatus sets Status field to given value.

### HasStatus

`func (o *Summary) HasStatus() bool`

HasStatus returns a boolean if a field has been set.

### GetSliValue

`func (o *Summary) GetSliValue() float32`

GetSliValue returns the SliValue field if non-nil, zero value otherwise.

### GetSliValueOk

`func (o *Summary) GetSliValueOk() (*float32, bool)`

GetSliValueOk returns a tuple with the SliValue field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSliValue

`func (o *Summary) SetSliValue(v float32)`

SetSliValue sets SliValue field to given value.

### HasSliValue

`func (o *Summary) HasSliValue() bool`

HasSliValue returns a boolean if a field has been set.

### GetErrorBudget

`func (o *Summary) GetErrorBudget() ErrorBudget`

GetErrorBudget returns the ErrorBudget field if non-nil, zero value otherwise.

### GetErrorBudgetOk

`func (o *Summary) GetErrorBudgetOk() (*ErrorBudget, bool)`

GetErrorBudgetOk returns a tuple with the ErrorBudget field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetErrorBudget

`func (o *Summary) SetErrorBudget(v ErrorBudget)`

SetErrorBudget sets ErrorBudget field to given value.

### HasErrorBudget

`func (o *Summary) HasErrorBudget() bool`

HasErrorBudget returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


