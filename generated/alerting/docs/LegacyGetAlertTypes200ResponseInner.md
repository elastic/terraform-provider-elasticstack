# LegacyGetAlertTypes200ResponseInner

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ActionGroups** | Pointer to [**[]LegacyGetAlertTypes200ResponseInnerActionGroupsInner**](LegacyGetAlertTypes200ResponseInnerActionGroupsInner.md) |  | [optional] 
**ActionVariables** | Pointer to [**LegacyGetAlertTypes200ResponseInnerActionVariables**](LegacyGetAlertTypes200ResponseInnerActionVariables.md) |  | [optional] 
**AuthorizedConsumers** | Pointer to **map[string]interface{}** | The list of the plugins IDs that have access to the alert type. | [optional] 
**DefaultActionGroupId** | Pointer to **string** | The default identifier for the alert type group. | [optional] 
**EnabledInLicense** | Pointer to **bool** | Indicates whether the rule type is enabled based on the subscription. | [optional] 
**Id** | Pointer to **string** | The unique identifier for the alert type. | [optional] 
**IsExportable** | Pointer to **bool** | Indicates whether the alert type is exportable in Saved Objects Management UI. | [optional] 
**MinimumLicenseRequired** | Pointer to **string** | The subscriptions required to use the alert type. | [optional] 
**Name** | Pointer to **string** | The descriptive name of the alert type. | [optional] 
**Producer** | Pointer to **string** | An identifier for the application that produces this alert type. | [optional] 
**RecoveryActionGroup** | Pointer to [**LegacyGetAlertTypes200ResponseInnerRecoveryActionGroup**](LegacyGetAlertTypes200ResponseInnerRecoveryActionGroup.md) |  | [optional] 

## Methods

### NewLegacyGetAlertTypes200ResponseInner

`func NewLegacyGetAlertTypes200ResponseInner() *LegacyGetAlertTypes200ResponseInner`

NewLegacyGetAlertTypes200ResponseInner instantiates a new LegacyGetAlertTypes200ResponseInner object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewLegacyGetAlertTypes200ResponseInnerWithDefaults

`func NewLegacyGetAlertTypes200ResponseInnerWithDefaults() *LegacyGetAlertTypes200ResponseInner`

NewLegacyGetAlertTypes200ResponseInnerWithDefaults instantiates a new LegacyGetAlertTypes200ResponseInner object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetActionGroups

`func (o *LegacyGetAlertTypes200ResponseInner) GetActionGroups() []LegacyGetAlertTypes200ResponseInnerActionGroupsInner`

GetActionGroups returns the ActionGroups field if non-nil, zero value otherwise.

### GetActionGroupsOk

`func (o *LegacyGetAlertTypes200ResponseInner) GetActionGroupsOk() (*[]LegacyGetAlertTypes200ResponseInnerActionGroupsInner, bool)`

GetActionGroupsOk returns a tuple with the ActionGroups field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetActionGroups

`func (o *LegacyGetAlertTypes200ResponseInner) SetActionGroups(v []LegacyGetAlertTypes200ResponseInnerActionGroupsInner)`

SetActionGroups sets ActionGroups field to given value.

### HasActionGroups

`func (o *LegacyGetAlertTypes200ResponseInner) HasActionGroups() bool`

HasActionGroups returns a boolean if a field has been set.

### GetActionVariables

`func (o *LegacyGetAlertTypes200ResponseInner) GetActionVariables() LegacyGetAlertTypes200ResponseInnerActionVariables`

GetActionVariables returns the ActionVariables field if non-nil, zero value otherwise.

### GetActionVariablesOk

`func (o *LegacyGetAlertTypes200ResponseInner) GetActionVariablesOk() (*LegacyGetAlertTypes200ResponseInnerActionVariables, bool)`

GetActionVariablesOk returns a tuple with the ActionVariables field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetActionVariables

`func (o *LegacyGetAlertTypes200ResponseInner) SetActionVariables(v LegacyGetAlertTypes200ResponseInnerActionVariables)`

SetActionVariables sets ActionVariables field to given value.

### HasActionVariables

`func (o *LegacyGetAlertTypes200ResponseInner) HasActionVariables() bool`

HasActionVariables returns a boolean if a field has been set.

### GetAuthorizedConsumers

`func (o *LegacyGetAlertTypes200ResponseInner) GetAuthorizedConsumers() map[string]interface{}`

GetAuthorizedConsumers returns the AuthorizedConsumers field if non-nil, zero value otherwise.

### GetAuthorizedConsumersOk

`func (o *LegacyGetAlertTypes200ResponseInner) GetAuthorizedConsumersOk() (*map[string]interface{}, bool)`

GetAuthorizedConsumersOk returns a tuple with the AuthorizedConsumers field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAuthorizedConsumers

`func (o *LegacyGetAlertTypes200ResponseInner) SetAuthorizedConsumers(v map[string]interface{})`

SetAuthorizedConsumers sets AuthorizedConsumers field to given value.

### HasAuthorizedConsumers

`func (o *LegacyGetAlertTypes200ResponseInner) HasAuthorizedConsumers() bool`

HasAuthorizedConsumers returns a boolean if a field has been set.

### GetDefaultActionGroupId

`func (o *LegacyGetAlertTypes200ResponseInner) GetDefaultActionGroupId() string`

GetDefaultActionGroupId returns the DefaultActionGroupId field if non-nil, zero value otherwise.

### GetDefaultActionGroupIdOk

`func (o *LegacyGetAlertTypes200ResponseInner) GetDefaultActionGroupIdOk() (*string, bool)`

GetDefaultActionGroupIdOk returns a tuple with the DefaultActionGroupId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDefaultActionGroupId

`func (o *LegacyGetAlertTypes200ResponseInner) SetDefaultActionGroupId(v string)`

SetDefaultActionGroupId sets DefaultActionGroupId field to given value.

### HasDefaultActionGroupId

`func (o *LegacyGetAlertTypes200ResponseInner) HasDefaultActionGroupId() bool`

HasDefaultActionGroupId returns a boolean if a field has been set.

### GetEnabledInLicense

`func (o *LegacyGetAlertTypes200ResponseInner) GetEnabledInLicense() bool`

GetEnabledInLicense returns the EnabledInLicense field if non-nil, zero value otherwise.

### GetEnabledInLicenseOk

`func (o *LegacyGetAlertTypes200ResponseInner) GetEnabledInLicenseOk() (*bool, bool)`

GetEnabledInLicenseOk returns a tuple with the EnabledInLicense field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEnabledInLicense

`func (o *LegacyGetAlertTypes200ResponseInner) SetEnabledInLicense(v bool)`

SetEnabledInLicense sets EnabledInLicense field to given value.

### HasEnabledInLicense

`func (o *LegacyGetAlertTypes200ResponseInner) HasEnabledInLicense() bool`

HasEnabledInLicense returns a boolean if a field has been set.

### GetId

`func (o *LegacyGetAlertTypes200ResponseInner) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *LegacyGetAlertTypes200ResponseInner) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *LegacyGetAlertTypes200ResponseInner) SetId(v string)`

SetId sets Id field to given value.

### HasId

`func (o *LegacyGetAlertTypes200ResponseInner) HasId() bool`

HasId returns a boolean if a field has been set.

### GetIsExportable

`func (o *LegacyGetAlertTypes200ResponseInner) GetIsExportable() bool`

GetIsExportable returns the IsExportable field if non-nil, zero value otherwise.

### GetIsExportableOk

`func (o *LegacyGetAlertTypes200ResponseInner) GetIsExportableOk() (*bool, bool)`

GetIsExportableOk returns a tuple with the IsExportable field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetIsExportable

`func (o *LegacyGetAlertTypes200ResponseInner) SetIsExportable(v bool)`

SetIsExportable sets IsExportable field to given value.

### HasIsExportable

`func (o *LegacyGetAlertTypes200ResponseInner) HasIsExportable() bool`

HasIsExportable returns a boolean if a field has been set.

### GetMinimumLicenseRequired

`func (o *LegacyGetAlertTypes200ResponseInner) GetMinimumLicenseRequired() string`

GetMinimumLicenseRequired returns the MinimumLicenseRequired field if non-nil, zero value otherwise.

### GetMinimumLicenseRequiredOk

`func (o *LegacyGetAlertTypes200ResponseInner) GetMinimumLicenseRequiredOk() (*string, bool)`

GetMinimumLicenseRequiredOk returns a tuple with the MinimumLicenseRequired field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMinimumLicenseRequired

`func (o *LegacyGetAlertTypes200ResponseInner) SetMinimumLicenseRequired(v string)`

SetMinimumLicenseRequired sets MinimumLicenseRequired field to given value.

### HasMinimumLicenseRequired

`func (o *LegacyGetAlertTypes200ResponseInner) HasMinimumLicenseRequired() bool`

HasMinimumLicenseRequired returns a boolean if a field has been set.

### GetName

`func (o *LegacyGetAlertTypes200ResponseInner) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *LegacyGetAlertTypes200ResponseInner) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *LegacyGetAlertTypes200ResponseInner) SetName(v string)`

SetName sets Name field to given value.

### HasName

`func (o *LegacyGetAlertTypes200ResponseInner) HasName() bool`

HasName returns a boolean if a field has been set.

### GetProducer

`func (o *LegacyGetAlertTypes200ResponseInner) GetProducer() string`

GetProducer returns the Producer field if non-nil, zero value otherwise.

### GetProducerOk

`func (o *LegacyGetAlertTypes200ResponseInner) GetProducerOk() (*string, bool)`

GetProducerOk returns a tuple with the Producer field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProducer

`func (o *LegacyGetAlertTypes200ResponseInner) SetProducer(v string)`

SetProducer sets Producer field to given value.

### HasProducer

`func (o *LegacyGetAlertTypes200ResponseInner) HasProducer() bool`

HasProducer returns a boolean if a field has been set.

### GetRecoveryActionGroup

`func (o *LegacyGetAlertTypes200ResponseInner) GetRecoveryActionGroup() LegacyGetAlertTypes200ResponseInnerRecoveryActionGroup`

GetRecoveryActionGroup returns the RecoveryActionGroup field if non-nil, zero value otherwise.

### GetRecoveryActionGroupOk

`func (o *LegacyGetAlertTypes200ResponseInner) GetRecoveryActionGroupOk() (*LegacyGetAlertTypes200ResponseInnerRecoveryActionGroup, bool)`

GetRecoveryActionGroupOk returns a tuple with the RecoveryActionGroup field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRecoveryActionGroup

`func (o *LegacyGetAlertTypes200ResponseInner) SetRecoveryActionGroup(v LegacyGetAlertTypes200ResponseInnerRecoveryActionGroup)`

SetRecoveryActionGroup sets RecoveryActionGroup field to given value.

### HasRecoveryActionGroup

`func (o *LegacyGetAlertTypes200ResponseInner) HasRecoveryActionGroup() bool`

HasRecoveryActionGroup returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


