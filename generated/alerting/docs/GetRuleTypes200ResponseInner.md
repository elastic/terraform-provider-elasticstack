# GetRuleTypes200ResponseInner

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ActionGroups** | Pointer to [**[]GetRuleTypes200ResponseInnerActionGroupsInner**](GetRuleTypes200ResponseInnerActionGroupsInner.md) | An explicit list of groups for which the rule type can schedule actions, each with the action group&#39;s unique ID and human readable name. Rule actions validation uses this configuration to ensure that groups are valid.  | [optional] 
**ActionVariables** | Pointer to [**GetRuleTypes200ResponseInnerActionVariables**](GetRuleTypes200ResponseInnerActionVariables.md) |  | [optional] 
**AuthorizedConsumers** | Pointer to [**GetRuleTypes200ResponseInnerAuthorizedConsumers**](GetRuleTypes200ResponseInnerAuthorizedConsumers.md) |  | [optional] 
**DefaultActionGroupId** | Pointer to **string** | The default identifier for the rule type group. | [optional] 
**DoesSetRecoveryContext** | Pointer to **bool** | Indicates whether the rule passes context variables to its recovery action. | [optional] 
**EnabledInLicense** | Pointer to **bool** | Indicates whether the rule type is enabled or disabled based on the subscription. | [optional] 
**Id** | Pointer to **string** | The unique identifier for the rule type. | [optional] 
**IsExportable** | Pointer to **bool** | Indicates whether the rule type is exportable in **Stack Management &gt; Saved Objects**. | [optional] 
**MinimumLicenseRequired** | Pointer to **string** | The subscriptions required to use the rule type. | [optional] 
**Name** | Pointer to **string** | The descriptive name of the rule type. | [optional] 
**Producer** | Pointer to **string** | An identifier for the application that produces this rule type. | [optional] 
**RecoveryActionGroup** | Pointer to [**GetRuleTypes200ResponseInnerRecoveryActionGroup**](GetRuleTypes200ResponseInnerRecoveryActionGroup.md) |  | [optional] 
**RuleTaskTimeout** | Pointer to **string** |  | [optional] 

## Methods

### NewGetRuleTypes200ResponseInner

`func NewGetRuleTypes200ResponseInner() *GetRuleTypes200ResponseInner`

NewGetRuleTypes200ResponseInner instantiates a new GetRuleTypes200ResponseInner object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewGetRuleTypes200ResponseInnerWithDefaults

`func NewGetRuleTypes200ResponseInnerWithDefaults() *GetRuleTypes200ResponseInner`

NewGetRuleTypes200ResponseInnerWithDefaults instantiates a new GetRuleTypes200ResponseInner object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetActionGroups

`func (o *GetRuleTypes200ResponseInner) GetActionGroups() []GetRuleTypes200ResponseInnerActionGroupsInner`

GetActionGroups returns the ActionGroups field if non-nil, zero value otherwise.

### GetActionGroupsOk

`func (o *GetRuleTypes200ResponseInner) GetActionGroupsOk() (*[]GetRuleTypes200ResponseInnerActionGroupsInner, bool)`

GetActionGroupsOk returns a tuple with the ActionGroups field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetActionGroups

`func (o *GetRuleTypes200ResponseInner) SetActionGroups(v []GetRuleTypes200ResponseInnerActionGroupsInner)`

SetActionGroups sets ActionGroups field to given value.

### HasActionGroups

`func (o *GetRuleTypes200ResponseInner) HasActionGroups() bool`

HasActionGroups returns a boolean if a field has been set.

### GetActionVariables

`func (o *GetRuleTypes200ResponseInner) GetActionVariables() GetRuleTypes200ResponseInnerActionVariables`

GetActionVariables returns the ActionVariables field if non-nil, zero value otherwise.

### GetActionVariablesOk

`func (o *GetRuleTypes200ResponseInner) GetActionVariablesOk() (*GetRuleTypes200ResponseInnerActionVariables, bool)`

GetActionVariablesOk returns a tuple with the ActionVariables field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetActionVariables

`func (o *GetRuleTypes200ResponseInner) SetActionVariables(v GetRuleTypes200ResponseInnerActionVariables)`

SetActionVariables sets ActionVariables field to given value.

### HasActionVariables

`func (o *GetRuleTypes200ResponseInner) HasActionVariables() bool`

HasActionVariables returns a boolean if a field has been set.

### GetAuthorizedConsumers

`func (o *GetRuleTypes200ResponseInner) GetAuthorizedConsumers() GetRuleTypes200ResponseInnerAuthorizedConsumers`

GetAuthorizedConsumers returns the AuthorizedConsumers field if non-nil, zero value otherwise.

### GetAuthorizedConsumersOk

`func (o *GetRuleTypes200ResponseInner) GetAuthorizedConsumersOk() (*GetRuleTypes200ResponseInnerAuthorizedConsumers, bool)`

GetAuthorizedConsumersOk returns a tuple with the AuthorizedConsumers field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAuthorizedConsumers

`func (o *GetRuleTypes200ResponseInner) SetAuthorizedConsumers(v GetRuleTypes200ResponseInnerAuthorizedConsumers)`

SetAuthorizedConsumers sets AuthorizedConsumers field to given value.

### HasAuthorizedConsumers

`func (o *GetRuleTypes200ResponseInner) HasAuthorizedConsumers() bool`

HasAuthorizedConsumers returns a boolean if a field has been set.

### GetDefaultActionGroupId

`func (o *GetRuleTypes200ResponseInner) GetDefaultActionGroupId() string`

GetDefaultActionGroupId returns the DefaultActionGroupId field if non-nil, zero value otherwise.

### GetDefaultActionGroupIdOk

`func (o *GetRuleTypes200ResponseInner) GetDefaultActionGroupIdOk() (*string, bool)`

GetDefaultActionGroupIdOk returns a tuple with the DefaultActionGroupId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDefaultActionGroupId

`func (o *GetRuleTypes200ResponseInner) SetDefaultActionGroupId(v string)`

SetDefaultActionGroupId sets DefaultActionGroupId field to given value.

### HasDefaultActionGroupId

`func (o *GetRuleTypes200ResponseInner) HasDefaultActionGroupId() bool`

HasDefaultActionGroupId returns a boolean if a field has been set.

### GetDoesSetRecoveryContext

`func (o *GetRuleTypes200ResponseInner) GetDoesSetRecoveryContext() bool`

GetDoesSetRecoveryContext returns the DoesSetRecoveryContext field if non-nil, zero value otherwise.

### GetDoesSetRecoveryContextOk

`func (o *GetRuleTypes200ResponseInner) GetDoesSetRecoveryContextOk() (*bool, bool)`

GetDoesSetRecoveryContextOk returns a tuple with the DoesSetRecoveryContext field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDoesSetRecoveryContext

`func (o *GetRuleTypes200ResponseInner) SetDoesSetRecoveryContext(v bool)`

SetDoesSetRecoveryContext sets DoesSetRecoveryContext field to given value.

### HasDoesSetRecoveryContext

`func (o *GetRuleTypes200ResponseInner) HasDoesSetRecoveryContext() bool`

HasDoesSetRecoveryContext returns a boolean if a field has been set.

### GetEnabledInLicense

`func (o *GetRuleTypes200ResponseInner) GetEnabledInLicense() bool`

GetEnabledInLicense returns the EnabledInLicense field if non-nil, zero value otherwise.

### GetEnabledInLicenseOk

`func (o *GetRuleTypes200ResponseInner) GetEnabledInLicenseOk() (*bool, bool)`

GetEnabledInLicenseOk returns a tuple with the EnabledInLicense field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEnabledInLicense

`func (o *GetRuleTypes200ResponseInner) SetEnabledInLicense(v bool)`

SetEnabledInLicense sets EnabledInLicense field to given value.

### HasEnabledInLicense

`func (o *GetRuleTypes200ResponseInner) HasEnabledInLicense() bool`

HasEnabledInLicense returns a boolean if a field has been set.

### GetId

`func (o *GetRuleTypes200ResponseInner) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *GetRuleTypes200ResponseInner) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *GetRuleTypes200ResponseInner) SetId(v string)`

SetId sets Id field to given value.

### HasId

`func (o *GetRuleTypes200ResponseInner) HasId() bool`

HasId returns a boolean if a field has been set.

### GetIsExportable

`func (o *GetRuleTypes200ResponseInner) GetIsExportable() bool`

GetIsExportable returns the IsExportable field if non-nil, zero value otherwise.

### GetIsExportableOk

`func (o *GetRuleTypes200ResponseInner) GetIsExportableOk() (*bool, bool)`

GetIsExportableOk returns a tuple with the IsExportable field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetIsExportable

`func (o *GetRuleTypes200ResponseInner) SetIsExportable(v bool)`

SetIsExportable sets IsExportable field to given value.

### HasIsExportable

`func (o *GetRuleTypes200ResponseInner) HasIsExportable() bool`

HasIsExportable returns a boolean if a field has been set.

### GetMinimumLicenseRequired

`func (o *GetRuleTypes200ResponseInner) GetMinimumLicenseRequired() string`

GetMinimumLicenseRequired returns the MinimumLicenseRequired field if non-nil, zero value otherwise.

### GetMinimumLicenseRequiredOk

`func (o *GetRuleTypes200ResponseInner) GetMinimumLicenseRequiredOk() (*string, bool)`

GetMinimumLicenseRequiredOk returns a tuple with the MinimumLicenseRequired field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMinimumLicenseRequired

`func (o *GetRuleTypes200ResponseInner) SetMinimumLicenseRequired(v string)`

SetMinimumLicenseRequired sets MinimumLicenseRequired field to given value.

### HasMinimumLicenseRequired

`func (o *GetRuleTypes200ResponseInner) HasMinimumLicenseRequired() bool`

HasMinimumLicenseRequired returns a boolean if a field has been set.

### GetName

`func (o *GetRuleTypes200ResponseInner) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *GetRuleTypes200ResponseInner) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *GetRuleTypes200ResponseInner) SetName(v string)`

SetName sets Name field to given value.

### HasName

`func (o *GetRuleTypes200ResponseInner) HasName() bool`

HasName returns a boolean if a field has been set.

### GetProducer

`func (o *GetRuleTypes200ResponseInner) GetProducer() string`

GetProducer returns the Producer field if non-nil, zero value otherwise.

### GetProducerOk

`func (o *GetRuleTypes200ResponseInner) GetProducerOk() (*string, bool)`

GetProducerOk returns a tuple with the Producer field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProducer

`func (o *GetRuleTypes200ResponseInner) SetProducer(v string)`

SetProducer sets Producer field to given value.

### HasProducer

`func (o *GetRuleTypes200ResponseInner) HasProducer() bool`

HasProducer returns a boolean if a field has been set.

### GetRecoveryActionGroup

`func (o *GetRuleTypes200ResponseInner) GetRecoveryActionGroup() GetRuleTypes200ResponseInnerRecoveryActionGroup`

GetRecoveryActionGroup returns the RecoveryActionGroup field if non-nil, zero value otherwise.

### GetRecoveryActionGroupOk

`func (o *GetRuleTypes200ResponseInner) GetRecoveryActionGroupOk() (*GetRuleTypes200ResponseInnerRecoveryActionGroup, bool)`

GetRecoveryActionGroupOk returns a tuple with the RecoveryActionGroup field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRecoveryActionGroup

`func (o *GetRuleTypes200ResponseInner) SetRecoveryActionGroup(v GetRuleTypes200ResponseInnerRecoveryActionGroup)`

SetRecoveryActionGroup sets RecoveryActionGroup field to given value.

### HasRecoveryActionGroup

`func (o *GetRuleTypes200ResponseInner) HasRecoveryActionGroup() bool`

HasRecoveryActionGroup returns a boolean if a field has been set.

### GetRuleTaskTimeout

`func (o *GetRuleTypes200ResponseInner) GetRuleTaskTimeout() string`

GetRuleTaskTimeout returns the RuleTaskTimeout field if non-nil, zero value otherwise.

### GetRuleTaskTimeoutOk

`func (o *GetRuleTypes200ResponseInner) GetRuleTaskTimeoutOk() (*string, bool)`

GetRuleTaskTimeoutOk returns a tuple with the RuleTaskTimeout field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRuleTaskTimeout

`func (o *GetRuleTypes200ResponseInner) SetRuleTaskTimeout(v string)`

SetRuleTaskTimeout sets RuleTaskTimeout field to given value.

### HasRuleTaskTimeout

`func (o *GetRuleTypes200ResponseInner) HasRuleTaskTimeout() bool`

HasRuleTaskTimeout returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


