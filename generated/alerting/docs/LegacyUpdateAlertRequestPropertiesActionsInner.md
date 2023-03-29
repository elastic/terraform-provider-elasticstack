# LegacyUpdateAlertRequestPropertiesActionsInner

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ActionTypeId** | **string** | The identifier for the action type. | 
**Group** | **string** | Grouping actions is recommended for escalations for different types of alert instances. If you don&#39;t need this functionality, set it to &#x60;default&#x60;.  | 
**Id** | **string** | The ID of the action saved object to execute. | 
**Params** | **map[string]interface{}** | The map to the &#x60;params&#x60; that the action type will receive. &#x60;params&#x60; are handled as Mustache templates and passed a default set of context.  | 

## Methods

### NewLegacyUpdateAlertRequestPropertiesActionsInner

`func NewLegacyUpdateAlertRequestPropertiesActionsInner(actionTypeId string, group string, id string, params map[string]interface{}, ) *LegacyUpdateAlertRequestPropertiesActionsInner`

NewLegacyUpdateAlertRequestPropertiesActionsInner instantiates a new LegacyUpdateAlertRequestPropertiesActionsInner object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewLegacyUpdateAlertRequestPropertiesActionsInnerWithDefaults

`func NewLegacyUpdateAlertRequestPropertiesActionsInnerWithDefaults() *LegacyUpdateAlertRequestPropertiesActionsInner`

NewLegacyUpdateAlertRequestPropertiesActionsInnerWithDefaults instantiates a new LegacyUpdateAlertRequestPropertiesActionsInner object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetActionTypeId

`func (o *LegacyUpdateAlertRequestPropertiesActionsInner) GetActionTypeId() string`

GetActionTypeId returns the ActionTypeId field if non-nil, zero value otherwise.

### GetActionTypeIdOk

`func (o *LegacyUpdateAlertRequestPropertiesActionsInner) GetActionTypeIdOk() (*string, bool)`

GetActionTypeIdOk returns a tuple with the ActionTypeId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetActionTypeId

`func (o *LegacyUpdateAlertRequestPropertiesActionsInner) SetActionTypeId(v string)`

SetActionTypeId sets ActionTypeId field to given value.


### GetGroup

`func (o *LegacyUpdateAlertRequestPropertiesActionsInner) GetGroup() string`

GetGroup returns the Group field if non-nil, zero value otherwise.

### GetGroupOk

`func (o *LegacyUpdateAlertRequestPropertiesActionsInner) GetGroupOk() (*string, bool)`

GetGroupOk returns a tuple with the Group field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGroup

`func (o *LegacyUpdateAlertRequestPropertiesActionsInner) SetGroup(v string)`

SetGroup sets Group field to given value.


### GetId

`func (o *LegacyUpdateAlertRequestPropertiesActionsInner) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *LegacyUpdateAlertRequestPropertiesActionsInner) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *LegacyUpdateAlertRequestPropertiesActionsInner) SetId(v string)`

SetId sets Id field to given value.


### GetParams

`func (o *LegacyUpdateAlertRequestPropertiesActionsInner) GetParams() map[string]interface{}`

GetParams returns the Params field if non-nil, zero value otherwise.

### GetParamsOk

`func (o *LegacyUpdateAlertRequestPropertiesActionsInner) GetParamsOk() (*map[string]interface{}, bool)`

GetParamsOk returns a tuple with the Params field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetParams

`func (o *LegacyUpdateAlertRequestPropertiesActionsInner) SetParams(v map[string]interface{})`

SetParams sets Params field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


