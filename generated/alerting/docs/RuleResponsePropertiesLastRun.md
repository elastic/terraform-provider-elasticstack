# RuleResponsePropertiesLastRun

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**AlertsCount** | Pointer to [**RuleResponsePropertiesLastRunAlertsCount**](RuleResponsePropertiesLastRunAlertsCount.md) |  | [optional] 
**Outcome** | Pointer to **string** |  | [optional] 
**OutcomeMsg** | Pointer to **NullableString** |  | [optional] 
**Warning** | Pointer to **NullableString** |  | [optional] 

## Methods

### NewRuleResponsePropertiesLastRun

`func NewRuleResponsePropertiesLastRun() *RuleResponsePropertiesLastRun`

NewRuleResponsePropertiesLastRun instantiates a new RuleResponsePropertiesLastRun object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewRuleResponsePropertiesLastRunWithDefaults

`func NewRuleResponsePropertiesLastRunWithDefaults() *RuleResponsePropertiesLastRun`

NewRuleResponsePropertiesLastRunWithDefaults instantiates a new RuleResponsePropertiesLastRun object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetAlertsCount

`func (o *RuleResponsePropertiesLastRun) GetAlertsCount() RuleResponsePropertiesLastRunAlertsCount`

GetAlertsCount returns the AlertsCount field if non-nil, zero value otherwise.

### GetAlertsCountOk

`func (o *RuleResponsePropertiesLastRun) GetAlertsCountOk() (*RuleResponsePropertiesLastRunAlertsCount, bool)`

GetAlertsCountOk returns a tuple with the AlertsCount field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAlertsCount

`func (o *RuleResponsePropertiesLastRun) SetAlertsCount(v RuleResponsePropertiesLastRunAlertsCount)`

SetAlertsCount sets AlertsCount field to given value.

### HasAlertsCount

`func (o *RuleResponsePropertiesLastRun) HasAlertsCount() bool`

HasAlertsCount returns a boolean if a field has been set.

### GetOutcome

`func (o *RuleResponsePropertiesLastRun) GetOutcome() string`

GetOutcome returns the Outcome field if non-nil, zero value otherwise.

### GetOutcomeOk

`func (o *RuleResponsePropertiesLastRun) GetOutcomeOk() (*string, bool)`

GetOutcomeOk returns a tuple with the Outcome field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOutcome

`func (o *RuleResponsePropertiesLastRun) SetOutcome(v string)`

SetOutcome sets Outcome field to given value.

### HasOutcome

`func (o *RuleResponsePropertiesLastRun) HasOutcome() bool`

HasOutcome returns a boolean if a field has been set.

### GetOutcomeMsg

`func (o *RuleResponsePropertiesLastRun) GetOutcomeMsg() string`

GetOutcomeMsg returns the OutcomeMsg field if non-nil, zero value otherwise.

### GetOutcomeMsgOk

`func (o *RuleResponsePropertiesLastRun) GetOutcomeMsgOk() (*string, bool)`

GetOutcomeMsgOk returns a tuple with the OutcomeMsg field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOutcomeMsg

`func (o *RuleResponsePropertiesLastRun) SetOutcomeMsg(v string)`

SetOutcomeMsg sets OutcomeMsg field to given value.

### HasOutcomeMsg

`func (o *RuleResponsePropertiesLastRun) HasOutcomeMsg() bool`

HasOutcomeMsg returns a boolean if a field has been set.

### SetOutcomeMsgNil

`func (o *RuleResponsePropertiesLastRun) SetOutcomeMsgNil(b bool)`

 SetOutcomeMsgNil sets the value for OutcomeMsg to be an explicit nil

### UnsetOutcomeMsg
`func (o *RuleResponsePropertiesLastRun) UnsetOutcomeMsg()`

UnsetOutcomeMsg ensures that no value is present for OutcomeMsg, not even an explicit nil
### GetWarning

`func (o *RuleResponsePropertiesLastRun) GetWarning() string`

GetWarning returns the Warning field if non-nil, zero value otherwise.

### GetWarningOk

`func (o *RuleResponsePropertiesLastRun) GetWarningOk() (*string, bool)`

GetWarningOk returns a tuple with the Warning field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetWarning

`func (o *RuleResponsePropertiesLastRun) SetWarning(v string)`

SetWarning sets Warning field to given value.

### HasWarning

`func (o *RuleResponsePropertiesLastRun) HasWarning() bool`

HasWarning returns a boolean if a field has been set.

### SetWarningNil

`func (o *RuleResponsePropertiesLastRun) SetWarningNil(b bool)`

 SetWarningNil sets the value for Warning to be an explicit nil

### UnsetWarning
`func (o *RuleResponsePropertiesLastRun) UnsetWarning()`

UnsetWarning ensures that no value is present for Warning, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


