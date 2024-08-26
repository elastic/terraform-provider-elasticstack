# GetRuleTypes200ResponseInnerAlerts

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Context** | Pointer to **string** | The namespace for this rule type.  | [optional] 
**Dynamic** | Pointer to **string** | Indicates whether new fields are added dynamically. | [optional] 
**IsSpaceAware** | Pointer to **bool** | Indicates whether the alerts are space-aware. If true, space-specific alert indices are used.  | [optional] 
**Mappings** | Pointer to [**GetRuleTypes200ResponseInnerAlertsMappings**](GetRuleTypes200ResponseInnerAlertsMappings.md) |  | [optional] 
**SecondaryAlias** | Pointer to **string** | A secondary alias. It is typically used to support the signals alias for detection rules.  | [optional] 
**ShouldWrite** | Pointer to **bool** | Indicates whether the rule should write out alerts as data.  | [optional] 
**UseEcs** | Pointer to **bool** | Indicates whether to include the ECS component template for the alerts.  | [optional] 
**UseLegacyAlerts** | Pointer to **bool** | Indicates whether to include the legacy component template for the alerts.  | [optional] [default to false]

## Methods

### NewGetRuleTypes200ResponseInnerAlerts

`func NewGetRuleTypes200ResponseInnerAlerts() *GetRuleTypes200ResponseInnerAlerts`

NewGetRuleTypes200ResponseInnerAlerts instantiates a new GetRuleTypes200ResponseInnerAlerts object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewGetRuleTypes200ResponseInnerAlertsWithDefaults

`func NewGetRuleTypes200ResponseInnerAlertsWithDefaults() *GetRuleTypes200ResponseInnerAlerts`

NewGetRuleTypes200ResponseInnerAlertsWithDefaults instantiates a new GetRuleTypes200ResponseInnerAlerts object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetContext

`func (o *GetRuleTypes200ResponseInnerAlerts) GetContext() string`

GetContext returns the Context field if non-nil, zero value otherwise.

### GetContextOk

`func (o *GetRuleTypes200ResponseInnerAlerts) GetContextOk() (*string, bool)`

GetContextOk returns a tuple with the Context field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetContext

`func (o *GetRuleTypes200ResponseInnerAlerts) SetContext(v string)`

SetContext sets Context field to given value.

### HasContext

`func (o *GetRuleTypes200ResponseInnerAlerts) HasContext() bool`

HasContext returns a boolean if a field has been set.

### GetDynamic

`func (o *GetRuleTypes200ResponseInnerAlerts) GetDynamic() string`

GetDynamic returns the Dynamic field if non-nil, zero value otherwise.

### GetDynamicOk

`func (o *GetRuleTypes200ResponseInnerAlerts) GetDynamicOk() (*string, bool)`

GetDynamicOk returns a tuple with the Dynamic field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDynamic

`func (o *GetRuleTypes200ResponseInnerAlerts) SetDynamic(v string)`

SetDynamic sets Dynamic field to given value.

### HasDynamic

`func (o *GetRuleTypes200ResponseInnerAlerts) HasDynamic() bool`

HasDynamic returns a boolean if a field has been set.

### GetIsSpaceAware

`func (o *GetRuleTypes200ResponseInnerAlerts) GetIsSpaceAware() bool`

GetIsSpaceAware returns the IsSpaceAware field if non-nil, zero value otherwise.

### GetIsSpaceAwareOk

`func (o *GetRuleTypes200ResponseInnerAlerts) GetIsSpaceAwareOk() (*bool, bool)`

GetIsSpaceAwareOk returns a tuple with the IsSpaceAware field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetIsSpaceAware

`func (o *GetRuleTypes200ResponseInnerAlerts) SetIsSpaceAware(v bool)`

SetIsSpaceAware sets IsSpaceAware field to given value.

### HasIsSpaceAware

`func (o *GetRuleTypes200ResponseInnerAlerts) HasIsSpaceAware() bool`

HasIsSpaceAware returns a boolean if a field has been set.

### GetMappings

`func (o *GetRuleTypes200ResponseInnerAlerts) GetMappings() GetRuleTypes200ResponseInnerAlertsMappings`

GetMappings returns the Mappings field if non-nil, zero value otherwise.

### GetMappingsOk

`func (o *GetRuleTypes200ResponseInnerAlerts) GetMappingsOk() (*GetRuleTypes200ResponseInnerAlertsMappings, bool)`

GetMappingsOk returns a tuple with the Mappings field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMappings

`func (o *GetRuleTypes200ResponseInnerAlerts) SetMappings(v GetRuleTypes200ResponseInnerAlertsMappings)`

SetMappings sets Mappings field to given value.

### HasMappings

`func (o *GetRuleTypes200ResponseInnerAlerts) HasMappings() bool`

HasMappings returns a boolean if a field has been set.

### GetSecondaryAlias

`func (o *GetRuleTypes200ResponseInnerAlerts) GetSecondaryAlias() string`

GetSecondaryAlias returns the SecondaryAlias field if non-nil, zero value otherwise.

### GetSecondaryAliasOk

`func (o *GetRuleTypes200ResponseInnerAlerts) GetSecondaryAliasOk() (*string, bool)`

GetSecondaryAliasOk returns a tuple with the SecondaryAlias field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSecondaryAlias

`func (o *GetRuleTypes200ResponseInnerAlerts) SetSecondaryAlias(v string)`

SetSecondaryAlias sets SecondaryAlias field to given value.

### HasSecondaryAlias

`func (o *GetRuleTypes200ResponseInnerAlerts) HasSecondaryAlias() bool`

HasSecondaryAlias returns a boolean if a field has been set.

### GetShouldWrite

`func (o *GetRuleTypes200ResponseInnerAlerts) GetShouldWrite() bool`

GetShouldWrite returns the ShouldWrite field if non-nil, zero value otherwise.

### GetShouldWriteOk

`func (o *GetRuleTypes200ResponseInnerAlerts) GetShouldWriteOk() (*bool, bool)`

GetShouldWriteOk returns a tuple with the ShouldWrite field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetShouldWrite

`func (o *GetRuleTypes200ResponseInnerAlerts) SetShouldWrite(v bool)`

SetShouldWrite sets ShouldWrite field to given value.

### HasShouldWrite

`func (o *GetRuleTypes200ResponseInnerAlerts) HasShouldWrite() bool`

HasShouldWrite returns a boolean if a field has been set.

### GetUseEcs

`func (o *GetRuleTypes200ResponseInnerAlerts) GetUseEcs() bool`

GetUseEcs returns the UseEcs field if non-nil, zero value otherwise.

### GetUseEcsOk

`func (o *GetRuleTypes200ResponseInnerAlerts) GetUseEcsOk() (*bool, bool)`

GetUseEcsOk returns a tuple with the UseEcs field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUseEcs

`func (o *GetRuleTypes200ResponseInnerAlerts) SetUseEcs(v bool)`

SetUseEcs sets UseEcs field to given value.

### HasUseEcs

`func (o *GetRuleTypes200ResponseInnerAlerts) HasUseEcs() bool`

HasUseEcs returns a boolean if a field has been set.

### GetUseLegacyAlerts

`func (o *GetRuleTypes200ResponseInnerAlerts) GetUseLegacyAlerts() bool`

GetUseLegacyAlerts returns the UseLegacyAlerts field if non-nil, zero value otherwise.

### GetUseLegacyAlertsOk

`func (o *GetRuleTypes200ResponseInnerAlerts) GetUseLegacyAlertsOk() (*bool, bool)`

GetUseLegacyAlertsOk returns a tuple with the UseLegacyAlerts field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUseLegacyAlerts

`func (o *GetRuleTypes200ResponseInnerAlerts) SetUseLegacyAlerts(v bool)`

SetUseLegacyAlerts sets UseLegacyAlerts field to given value.

### HasUseLegacyAlerts

`func (o *GetRuleTypes200ResponseInnerAlerts) HasUseLegacyAlerts() bool`

HasUseLegacyAlerts returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


