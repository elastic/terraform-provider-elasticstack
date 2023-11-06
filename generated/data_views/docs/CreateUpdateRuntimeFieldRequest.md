# CreateUpdateRuntimeFieldRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | **string** | The name for a runtime field.  | 
**RuntimeField** | **map[string]interface{}** | The runtime field definition object.  | 

## Methods

### NewCreateUpdateRuntimeFieldRequest

`func NewCreateUpdateRuntimeFieldRequest(name string, runtimeField map[string]interface{}, ) *CreateUpdateRuntimeFieldRequest`

NewCreateUpdateRuntimeFieldRequest instantiates a new CreateUpdateRuntimeFieldRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewCreateUpdateRuntimeFieldRequestWithDefaults

`func NewCreateUpdateRuntimeFieldRequestWithDefaults() *CreateUpdateRuntimeFieldRequest`

NewCreateUpdateRuntimeFieldRequestWithDefaults instantiates a new CreateUpdateRuntimeFieldRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetName

`func (o *CreateUpdateRuntimeFieldRequest) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *CreateUpdateRuntimeFieldRequest) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *CreateUpdateRuntimeFieldRequest) SetName(v string)`

SetName sets Name field to given value.


### GetRuntimeField

`func (o *CreateUpdateRuntimeFieldRequest) GetRuntimeField() map[string]interface{}`

GetRuntimeField returns the RuntimeField field if non-nil, zero value otherwise.

### GetRuntimeFieldOk

`func (o *CreateUpdateRuntimeFieldRequest) GetRuntimeFieldOk() (*map[string]interface{}, bool)`

GetRuntimeFieldOk returns a tuple with the RuntimeField field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRuntimeField

`func (o *CreateUpdateRuntimeFieldRequest) SetRuntimeField(v map[string]interface{})`

SetRuntimeField sets RuntimeField field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


