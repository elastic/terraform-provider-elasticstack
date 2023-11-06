# CreateDataViewRequestObject

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**DataView** | [**CreateDataViewRequestObjectDataView**](CreateDataViewRequestObjectDataView.md) |  | 
**Override** | Pointer to **interface{}** | Override an existing data view if a data view with the provided title already exists. | [optional] [default to false]

## Methods

### NewCreateDataViewRequestObject

`func NewCreateDataViewRequestObject(dataView CreateDataViewRequestObjectDataView, ) *CreateDataViewRequestObject`

NewCreateDataViewRequestObject instantiates a new CreateDataViewRequestObject object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewCreateDataViewRequestObjectWithDefaults

`func NewCreateDataViewRequestObjectWithDefaults() *CreateDataViewRequestObject`

NewCreateDataViewRequestObjectWithDefaults instantiates a new CreateDataViewRequestObject object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetDataView

`func (o *CreateDataViewRequestObject) GetDataView() CreateDataViewRequestObjectDataView`

GetDataView returns the DataView field if non-nil, zero value otherwise.

### GetDataViewOk

`func (o *CreateDataViewRequestObject) GetDataViewOk() (*CreateDataViewRequestObjectDataView, bool)`

GetDataViewOk returns a tuple with the DataView field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDataView

`func (o *CreateDataViewRequestObject) SetDataView(v CreateDataViewRequestObjectDataView)`

SetDataView sets DataView field to given value.


### GetOverride

`func (o *CreateDataViewRequestObject) GetOverride() interface{}`

GetOverride returns the Override field if non-nil, zero value otherwise.

### GetOverrideOk

`func (o *CreateDataViewRequestObject) GetOverrideOk() (*interface{}, bool)`

GetOverrideOk returns a tuple with the Override field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOverride

`func (o *CreateDataViewRequestObject) SetOverride(v interface{})`

SetOverride sets Override field to given value.

### HasOverride

`func (o *CreateDataViewRequestObject) HasOverride() bool`

HasOverride returns a boolean if a field has been set.

### SetOverrideNil

`func (o *CreateDataViewRequestObject) SetOverrideNil(b bool)`

 SetOverrideNil sets the value for Override to be an explicit nil

### UnsetOverride
`func (o *CreateDataViewRequestObject) UnsetOverride()`

UnsetOverride ensures that no value is present for Override, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


