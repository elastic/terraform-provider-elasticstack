# SetDefaultDatailViewRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**DataViewId** | **interface{}** | The data view identifier. NOTE: The API does not validate whether it is a valid identifier. Use &#x60;null&#x60; to unset the default data view.  | 
**Force** | Pointer to **interface{}** | Update an existing default data view identifier. | [optional] [default to false]

## Methods

### NewSetDefaultDatailViewRequest

`func NewSetDefaultDatailViewRequest(dataViewId interface{}, ) *SetDefaultDatailViewRequest`

NewSetDefaultDatailViewRequest instantiates a new SetDefaultDatailViewRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewSetDefaultDatailViewRequestWithDefaults

`func NewSetDefaultDatailViewRequestWithDefaults() *SetDefaultDatailViewRequest`

NewSetDefaultDatailViewRequestWithDefaults instantiates a new SetDefaultDatailViewRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetDataViewId

`func (o *SetDefaultDatailViewRequest) GetDataViewId() interface{}`

GetDataViewId returns the DataViewId field if non-nil, zero value otherwise.

### GetDataViewIdOk

`func (o *SetDefaultDatailViewRequest) GetDataViewIdOk() (*interface{}, bool)`

GetDataViewIdOk returns a tuple with the DataViewId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDataViewId

`func (o *SetDefaultDatailViewRequest) SetDataViewId(v interface{})`

SetDataViewId sets DataViewId field to given value.


### SetDataViewIdNil

`func (o *SetDefaultDatailViewRequest) SetDataViewIdNil(b bool)`

 SetDataViewIdNil sets the value for DataViewId to be an explicit nil

### UnsetDataViewId
`func (o *SetDefaultDatailViewRequest) UnsetDataViewId()`

UnsetDataViewId ensures that no value is present for DataViewId, not even an explicit nil
### GetForce

`func (o *SetDefaultDatailViewRequest) GetForce() interface{}`

GetForce returns the Force field if non-nil, zero value otherwise.

### GetForceOk

`func (o *SetDefaultDatailViewRequest) GetForceOk() (*interface{}, bool)`

GetForceOk returns a tuple with the Force field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetForce

`func (o *SetDefaultDatailViewRequest) SetForce(v interface{})`

SetForce sets Force field to given value.

### HasForce

`func (o *SetDefaultDatailViewRequest) HasForce() bool`

HasForce returns a boolean if a field has been set.

### SetForceNil

`func (o *SetDefaultDatailViewRequest) SetForceNil(b bool)`

 SetForceNil sets the value for Force to be an explicit nil

### UnsetForce
`func (o *SetDefaultDatailViewRequest) UnsetForce()`

UnsetForce ensures that no value is present for Force, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


