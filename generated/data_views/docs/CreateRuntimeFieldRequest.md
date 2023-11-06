# CreateRuntimeFieldRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | **interface{}** | The name for a runtime field.  | 
**RuntimeField** | **interface{}** | The runtime field definition object.  | 

## Methods

### NewCreateRuntimeFieldRequest

`func NewCreateRuntimeFieldRequest(name interface{}, runtimeField interface{}, ) *CreateRuntimeFieldRequest`

NewCreateRuntimeFieldRequest instantiates a new CreateRuntimeFieldRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewCreateRuntimeFieldRequestWithDefaults

`func NewCreateRuntimeFieldRequestWithDefaults() *CreateRuntimeFieldRequest`

NewCreateRuntimeFieldRequestWithDefaults instantiates a new CreateRuntimeFieldRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetName

`func (o *CreateRuntimeFieldRequest) GetName() interface{}`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *CreateRuntimeFieldRequest) GetNameOk() (*interface{}, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *CreateRuntimeFieldRequest) SetName(v interface{})`

SetName sets Name field to given value.


### SetNameNil

`func (o *CreateRuntimeFieldRequest) SetNameNil(b bool)`

 SetNameNil sets the value for Name to be an explicit nil

### UnsetName
`func (o *CreateRuntimeFieldRequest) UnsetName()`

UnsetName ensures that no value is present for Name, not even an explicit nil
### GetRuntimeField

`func (o *CreateRuntimeFieldRequest) GetRuntimeField() interface{}`

GetRuntimeField returns the RuntimeField field if non-nil, zero value otherwise.

### GetRuntimeFieldOk

`func (o *CreateRuntimeFieldRequest) GetRuntimeFieldOk() (*interface{}, bool)`

GetRuntimeFieldOk returns a tuple with the RuntimeField field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRuntimeField

`func (o *CreateRuntimeFieldRequest) SetRuntimeField(v interface{})`

SetRuntimeField sets RuntimeField field to given value.


### SetRuntimeFieldNil

`func (o *CreateRuntimeFieldRequest) SetRuntimeFieldNil(b bool)`

 SetRuntimeFieldNil sets the value for RuntimeField to be an explicit nil

### UnsetRuntimeField
`func (o *CreateRuntimeFieldRequest) UnsetRuntimeField()`

UnsetRuntimeField ensures that no value is present for RuntimeField, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


