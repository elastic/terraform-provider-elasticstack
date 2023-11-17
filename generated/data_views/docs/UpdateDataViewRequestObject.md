# UpdateDataViewRequestObject

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**DataView** | [**UpdateDataViewRequestObjectDataView**](UpdateDataViewRequestObjectDataView.md) |  | 
**RefreshFields** | Pointer to **bool** | Reloads the data view fields after the data view is updated. | [optional] [default to false]

## Methods

### NewUpdateDataViewRequestObject

`func NewUpdateDataViewRequestObject(dataView UpdateDataViewRequestObjectDataView, ) *UpdateDataViewRequestObject`

NewUpdateDataViewRequestObject instantiates a new UpdateDataViewRequestObject object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewUpdateDataViewRequestObjectWithDefaults

`func NewUpdateDataViewRequestObjectWithDefaults() *UpdateDataViewRequestObject`

NewUpdateDataViewRequestObjectWithDefaults instantiates a new UpdateDataViewRequestObject object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetDataView

`func (o *UpdateDataViewRequestObject) GetDataView() UpdateDataViewRequestObjectDataView`

GetDataView returns the DataView field if non-nil, zero value otherwise.

### GetDataViewOk

`func (o *UpdateDataViewRequestObject) GetDataViewOk() (*UpdateDataViewRequestObjectDataView, bool)`

GetDataViewOk returns a tuple with the DataView field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDataView

`func (o *UpdateDataViewRequestObject) SetDataView(v UpdateDataViewRequestObjectDataView)`

SetDataView sets DataView field to given value.


### GetRefreshFields

`func (o *UpdateDataViewRequestObject) GetRefreshFields() bool`

GetRefreshFields returns the RefreshFields field if non-nil, zero value otherwise.

### GetRefreshFieldsOk

`func (o *UpdateDataViewRequestObject) GetRefreshFieldsOk() (*bool, bool)`

GetRefreshFieldsOk returns a tuple with the RefreshFields field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRefreshFields

`func (o *UpdateDataViewRequestObject) SetRefreshFields(v bool)`

SetRefreshFields sets RefreshFields field to given value.

### HasRefreshFields

`func (o *UpdateDataViewRequestObject) HasRefreshFields() bool`

HasRefreshFields returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


