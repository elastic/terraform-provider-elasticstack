# FilterMeta

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Alias** | Pointer to **NullableString** |  | [optional] 
**ControlledBy** | Pointer to **string** |  | [optional] 
**Disabled** | Pointer to **bool** |  | [optional] 
**Field** | Pointer to **string** |  | [optional] 
**Group** | Pointer to **string** |  | [optional] 
**Index** | Pointer to **string** |  | [optional] 
**IsMultiIndex** | Pointer to **bool** |  | [optional] 
**Key** | Pointer to **string** |  | [optional] 
**Negate** | Pointer to **bool** |  | [optional] 
**Params** | Pointer to **map[string]interface{}** |  | [optional] 
**Type** | Pointer to **string** |  | [optional] 
**Value** | Pointer to **string** |  | [optional] 

## Methods

### NewFilterMeta

`func NewFilterMeta() *FilterMeta`

NewFilterMeta instantiates a new FilterMeta object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewFilterMetaWithDefaults

`func NewFilterMetaWithDefaults() *FilterMeta`

NewFilterMetaWithDefaults instantiates a new FilterMeta object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetAlias

`func (o *FilterMeta) GetAlias() string`

GetAlias returns the Alias field if non-nil, zero value otherwise.

### GetAliasOk

`func (o *FilterMeta) GetAliasOk() (*string, bool)`

GetAliasOk returns a tuple with the Alias field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAlias

`func (o *FilterMeta) SetAlias(v string)`

SetAlias sets Alias field to given value.

### HasAlias

`func (o *FilterMeta) HasAlias() bool`

HasAlias returns a boolean if a field has been set.

### SetAliasNil

`func (o *FilterMeta) SetAliasNil(b bool)`

 SetAliasNil sets the value for Alias to be an explicit nil

### UnsetAlias
`func (o *FilterMeta) UnsetAlias()`

UnsetAlias ensures that no value is present for Alias, not even an explicit nil
### GetControlledBy

`func (o *FilterMeta) GetControlledBy() string`

GetControlledBy returns the ControlledBy field if non-nil, zero value otherwise.

### GetControlledByOk

`func (o *FilterMeta) GetControlledByOk() (*string, bool)`

GetControlledByOk returns a tuple with the ControlledBy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetControlledBy

`func (o *FilterMeta) SetControlledBy(v string)`

SetControlledBy sets ControlledBy field to given value.

### HasControlledBy

`func (o *FilterMeta) HasControlledBy() bool`

HasControlledBy returns a boolean if a field has been set.

### GetDisabled

`func (o *FilterMeta) GetDisabled() bool`

GetDisabled returns the Disabled field if non-nil, zero value otherwise.

### GetDisabledOk

`func (o *FilterMeta) GetDisabledOk() (*bool, bool)`

GetDisabledOk returns a tuple with the Disabled field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDisabled

`func (o *FilterMeta) SetDisabled(v bool)`

SetDisabled sets Disabled field to given value.

### HasDisabled

`func (o *FilterMeta) HasDisabled() bool`

HasDisabled returns a boolean if a field has been set.

### GetField

`func (o *FilterMeta) GetField() string`

GetField returns the Field field if non-nil, zero value otherwise.

### GetFieldOk

`func (o *FilterMeta) GetFieldOk() (*string, bool)`

GetFieldOk returns a tuple with the Field field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetField

`func (o *FilterMeta) SetField(v string)`

SetField sets Field field to given value.

### HasField

`func (o *FilterMeta) HasField() bool`

HasField returns a boolean if a field has been set.

### GetGroup

`func (o *FilterMeta) GetGroup() string`

GetGroup returns the Group field if non-nil, zero value otherwise.

### GetGroupOk

`func (o *FilterMeta) GetGroupOk() (*string, bool)`

GetGroupOk returns a tuple with the Group field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGroup

`func (o *FilterMeta) SetGroup(v string)`

SetGroup sets Group field to given value.

### HasGroup

`func (o *FilterMeta) HasGroup() bool`

HasGroup returns a boolean if a field has been set.

### GetIndex

`func (o *FilterMeta) GetIndex() string`

GetIndex returns the Index field if non-nil, zero value otherwise.

### GetIndexOk

`func (o *FilterMeta) GetIndexOk() (*string, bool)`

GetIndexOk returns a tuple with the Index field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetIndex

`func (o *FilterMeta) SetIndex(v string)`

SetIndex sets Index field to given value.

### HasIndex

`func (o *FilterMeta) HasIndex() bool`

HasIndex returns a boolean if a field has been set.

### GetIsMultiIndex

`func (o *FilterMeta) GetIsMultiIndex() bool`

GetIsMultiIndex returns the IsMultiIndex field if non-nil, zero value otherwise.

### GetIsMultiIndexOk

`func (o *FilterMeta) GetIsMultiIndexOk() (*bool, bool)`

GetIsMultiIndexOk returns a tuple with the IsMultiIndex field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetIsMultiIndex

`func (o *FilterMeta) SetIsMultiIndex(v bool)`

SetIsMultiIndex sets IsMultiIndex field to given value.

### HasIsMultiIndex

`func (o *FilterMeta) HasIsMultiIndex() bool`

HasIsMultiIndex returns a boolean if a field has been set.

### GetKey

`func (o *FilterMeta) GetKey() string`

GetKey returns the Key field if non-nil, zero value otherwise.

### GetKeyOk

`func (o *FilterMeta) GetKeyOk() (*string, bool)`

GetKeyOk returns a tuple with the Key field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetKey

`func (o *FilterMeta) SetKey(v string)`

SetKey sets Key field to given value.

### HasKey

`func (o *FilterMeta) HasKey() bool`

HasKey returns a boolean if a field has been set.

### GetNegate

`func (o *FilterMeta) GetNegate() bool`

GetNegate returns the Negate field if non-nil, zero value otherwise.

### GetNegateOk

`func (o *FilterMeta) GetNegateOk() (*bool, bool)`

GetNegateOk returns a tuple with the Negate field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNegate

`func (o *FilterMeta) SetNegate(v bool)`

SetNegate sets Negate field to given value.

### HasNegate

`func (o *FilterMeta) HasNegate() bool`

HasNegate returns a boolean if a field has been set.

### GetParams

`func (o *FilterMeta) GetParams() map[string]interface{}`

GetParams returns the Params field if non-nil, zero value otherwise.

### GetParamsOk

`func (o *FilterMeta) GetParamsOk() (*map[string]interface{}, bool)`

GetParamsOk returns a tuple with the Params field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetParams

`func (o *FilterMeta) SetParams(v map[string]interface{})`

SetParams sets Params field to given value.

### HasParams

`func (o *FilterMeta) HasParams() bool`

HasParams returns a boolean if a field has been set.

### GetType

`func (o *FilterMeta) GetType() string`

GetType returns the Type field if non-nil, zero value otherwise.

### GetTypeOk

`func (o *FilterMeta) GetTypeOk() (*string, bool)`

GetTypeOk returns a tuple with the Type field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetType

`func (o *FilterMeta) SetType(v string)`

SetType sets Type field to given value.

### HasType

`func (o *FilterMeta) HasType() bool`

HasType returns a boolean if a field has been set.

### GetValue

`func (o *FilterMeta) GetValue() string`

GetValue returns the Value field if non-nil, zero value otherwise.

### GetValueOk

`func (o *FilterMeta) GetValueOk() (*string, bool)`

GetValueOk returns a tuple with the Value field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetValue

`func (o *FilterMeta) SetValue(v string)`

SetValue sets Value field to given value.

### HasValue

`func (o *FilterMeta) HasValue() bool`

HasValue returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


