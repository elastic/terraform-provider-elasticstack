# ActionsInner

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ConnectorTypeId** | Pointer to **string** | The type of connector. This property appears in responses but cannot be set in requests. | [optional] [readonly] 
**Frequency** | Pointer to [**ActionsInnerFrequency**](ActionsInnerFrequency.md) |  | [optional] 
**Group** | Pointer to **string** | The group name for the actions. If you don&#39;t need to group actions, set to &#x60;default&#x60;. | [optional] 
**Id** | Pointer to **string** | The identifier for the connector saved object. | [optional] 
**Params** | Pointer to **map[string]interface{}** | The parameters for the action, which are sent to the connector. The &#x60;params&#x60; are handled as Mustache templates and passed a default set of context. | [optional] 

## Methods

### NewActionsInner

`func NewActionsInner() *ActionsInner`

NewActionsInner instantiates a new ActionsInner object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewActionsInnerWithDefaults

`func NewActionsInnerWithDefaults() *ActionsInner`

NewActionsInnerWithDefaults instantiates a new ActionsInner object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetConnectorTypeId

`func (o *ActionsInner) GetConnectorTypeId() string`

GetConnectorTypeId returns the ConnectorTypeId field if non-nil, zero value otherwise.

### GetConnectorTypeIdOk

`func (o *ActionsInner) GetConnectorTypeIdOk() (*string, bool)`

GetConnectorTypeIdOk returns a tuple with the ConnectorTypeId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetConnectorTypeId

`func (o *ActionsInner) SetConnectorTypeId(v string)`

SetConnectorTypeId sets ConnectorTypeId field to given value.

### HasConnectorTypeId

`func (o *ActionsInner) HasConnectorTypeId() bool`

HasConnectorTypeId returns a boolean if a field has been set.

### GetFrequency

`func (o *ActionsInner) GetFrequency() ActionsInnerFrequency`

GetFrequency returns the Frequency field if non-nil, zero value otherwise.

### GetFrequencyOk

`func (o *ActionsInner) GetFrequencyOk() (*ActionsInnerFrequency, bool)`

GetFrequencyOk returns a tuple with the Frequency field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFrequency

`func (o *ActionsInner) SetFrequency(v ActionsInnerFrequency)`

SetFrequency sets Frequency field to given value.

### HasFrequency

`func (o *ActionsInner) HasFrequency() bool`

HasFrequency returns a boolean if a field has been set.

### GetGroup

`func (o *ActionsInner) GetGroup() string`

GetGroup returns the Group field if non-nil, zero value otherwise.

### GetGroupOk

`func (o *ActionsInner) GetGroupOk() (*string, bool)`

GetGroupOk returns a tuple with the Group field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGroup

`func (o *ActionsInner) SetGroup(v string)`

SetGroup sets Group field to given value.

### HasGroup

`func (o *ActionsInner) HasGroup() bool`

HasGroup returns a boolean if a field has been set.

### GetId

`func (o *ActionsInner) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *ActionsInner) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *ActionsInner) SetId(v string)`

SetId sets Id field to given value.

### HasId

`func (o *ActionsInner) HasId() bool`

HasId returns a boolean if a field has been set.

### GetParams

`func (o *ActionsInner) GetParams() map[string]interface{}`

GetParams returns the Params field if non-nil, zero value otherwise.

### GetParamsOk

`func (o *ActionsInner) GetParamsOk() (*map[string]interface{}, bool)`

GetParamsOk returns a tuple with the Params field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetParams

`func (o *ActionsInner) SetParams(v map[string]interface{})`

SetParams sets Params field to given value.

### HasParams

`func (o *ActionsInner) HasParams() bool`

HasParams returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


