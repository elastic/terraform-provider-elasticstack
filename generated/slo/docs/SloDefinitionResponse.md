# SloDefinitionResponse

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | **string** | The identifier of the SLO. | 
**Name** | **string** | The name of the SLO. | 
**Description** | **string** | The description of the SLO. | 
**Indicator** | [**SloWithSummaryResponseIndicator**](SloWithSummaryResponseIndicator.md) |  | 
**TimeWindow** | [**TimeWindow**](TimeWindow.md) |  | 
**BudgetingMethod** | [**BudgetingMethod**](BudgetingMethod.md) |  | 
**Objective** | [**Objective**](Objective.md) |  | 
**Settings** | [**Settings**](Settings.md) |  | 
**Revision** | **float32** | The SLO revision | 
**Enabled** | **bool** | Indicate if the SLO is enabled | 
**GroupBy** | [**GroupBy**](GroupBy.md) |  | 
**Tags** | **[]string** | List of tags | 
**CreatedAt** | **string** | The creation date | 
**UpdatedAt** | **string** | The last update date | 
**Version** | **float32** | The internal SLO version | 

## Methods

### NewSloDefinitionResponse

`func NewSloDefinitionResponse(id string, name string, description string, indicator SloWithSummaryResponseIndicator, timeWindow TimeWindow, budgetingMethod BudgetingMethod, objective Objective, settings Settings, revision float32, enabled bool, groupBy GroupBy, tags []string, createdAt string, updatedAt string, version float32, ) *SloDefinitionResponse`

NewSloDefinitionResponse instantiates a new SloDefinitionResponse object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewSloDefinitionResponseWithDefaults

`func NewSloDefinitionResponseWithDefaults() *SloDefinitionResponse`

NewSloDefinitionResponseWithDefaults instantiates a new SloDefinitionResponse object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *SloDefinitionResponse) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *SloDefinitionResponse) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *SloDefinitionResponse) SetId(v string)`

SetId sets Id field to given value.


### GetName

`func (o *SloDefinitionResponse) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *SloDefinitionResponse) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *SloDefinitionResponse) SetName(v string)`

SetName sets Name field to given value.


### GetDescription

`func (o *SloDefinitionResponse) GetDescription() string`

GetDescription returns the Description field if non-nil, zero value otherwise.

### GetDescriptionOk

`func (o *SloDefinitionResponse) GetDescriptionOk() (*string, bool)`

GetDescriptionOk returns a tuple with the Description field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDescription

`func (o *SloDefinitionResponse) SetDescription(v string)`

SetDescription sets Description field to given value.


### GetIndicator

`func (o *SloDefinitionResponse) GetIndicator() SloWithSummaryResponseIndicator`

GetIndicator returns the Indicator field if non-nil, zero value otherwise.

### GetIndicatorOk

`func (o *SloDefinitionResponse) GetIndicatorOk() (*SloWithSummaryResponseIndicator, bool)`

GetIndicatorOk returns a tuple with the Indicator field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetIndicator

`func (o *SloDefinitionResponse) SetIndicator(v SloWithSummaryResponseIndicator)`

SetIndicator sets Indicator field to given value.


### GetTimeWindow

`func (o *SloDefinitionResponse) GetTimeWindow() TimeWindow`

GetTimeWindow returns the TimeWindow field if non-nil, zero value otherwise.

### GetTimeWindowOk

`func (o *SloDefinitionResponse) GetTimeWindowOk() (*TimeWindow, bool)`

GetTimeWindowOk returns a tuple with the TimeWindow field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTimeWindow

`func (o *SloDefinitionResponse) SetTimeWindow(v TimeWindow)`

SetTimeWindow sets TimeWindow field to given value.


### GetBudgetingMethod

`func (o *SloDefinitionResponse) GetBudgetingMethod() BudgetingMethod`

GetBudgetingMethod returns the BudgetingMethod field if non-nil, zero value otherwise.

### GetBudgetingMethodOk

`func (o *SloDefinitionResponse) GetBudgetingMethodOk() (*BudgetingMethod, bool)`

GetBudgetingMethodOk returns a tuple with the BudgetingMethod field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetBudgetingMethod

`func (o *SloDefinitionResponse) SetBudgetingMethod(v BudgetingMethod)`

SetBudgetingMethod sets BudgetingMethod field to given value.


### GetObjective

`func (o *SloDefinitionResponse) GetObjective() Objective`

GetObjective returns the Objective field if non-nil, zero value otherwise.

### GetObjectiveOk

`func (o *SloDefinitionResponse) GetObjectiveOk() (*Objective, bool)`

GetObjectiveOk returns a tuple with the Objective field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetObjective

`func (o *SloDefinitionResponse) SetObjective(v Objective)`

SetObjective sets Objective field to given value.


### GetSettings

`func (o *SloDefinitionResponse) GetSettings() Settings`

GetSettings returns the Settings field if non-nil, zero value otherwise.

### GetSettingsOk

`func (o *SloDefinitionResponse) GetSettingsOk() (*Settings, bool)`

GetSettingsOk returns a tuple with the Settings field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSettings

`func (o *SloDefinitionResponse) SetSettings(v Settings)`

SetSettings sets Settings field to given value.


### GetRevision

`func (o *SloDefinitionResponse) GetRevision() float32`

GetRevision returns the Revision field if non-nil, zero value otherwise.

### GetRevisionOk

`func (o *SloDefinitionResponse) GetRevisionOk() (*float32, bool)`

GetRevisionOk returns a tuple with the Revision field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRevision

`func (o *SloDefinitionResponse) SetRevision(v float32)`

SetRevision sets Revision field to given value.


### GetEnabled

`func (o *SloDefinitionResponse) GetEnabled() bool`

GetEnabled returns the Enabled field if non-nil, zero value otherwise.

### GetEnabledOk

`func (o *SloDefinitionResponse) GetEnabledOk() (*bool, bool)`

GetEnabledOk returns a tuple with the Enabled field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEnabled

`func (o *SloDefinitionResponse) SetEnabled(v bool)`

SetEnabled sets Enabled field to given value.


### GetGroupBy

`func (o *SloDefinitionResponse) GetGroupBy() GroupBy`

GetGroupBy returns the GroupBy field if non-nil, zero value otherwise.

### GetGroupByOk

`func (o *SloDefinitionResponse) GetGroupByOk() (*GroupBy, bool)`

GetGroupByOk returns a tuple with the GroupBy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGroupBy

`func (o *SloDefinitionResponse) SetGroupBy(v GroupBy)`

SetGroupBy sets GroupBy field to given value.


### GetTags

`func (o *SloDefinitionResponse) GetTags() []string`

GetTags returns the Tags field if non-nil, zero value otherwise.

### GetTagsOk

`func (o *SloDefinitionResponse) GetTagsOk() (*[]string, bool)`

GetTagsOk returns a tuple with the Tags field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTags

`func (o *SloDefinitionResponse) SetTags(v []string)`

SetTags sets Tags field to given value.


### GetCreatedAt

`func (o *SloDefinitionResponse) GetCreatedAt() string`

GetCreatedAt returns the CreatedAt field if non-nil, zero value otherwise.

### GetCreatedAtOk

`func (o *SloDefinitionResponse) GetCreatedAtOk() (*string, bool)`

GetCreatedAtOk returns a tuple with the CreatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedAt

`func (o *SloDefinitionResponse) SetCreatedAt(v string)`

SetCreatedAt sets CreatedAt field to given value.


### GetUpdatedAt

`func (o *SloDefinitionResponse) GetUpdatedAt() string`

GetUpdatedAt returns the UpdatedAt field if non-nil, zero value otherwise.

### GetUpdatedAtOk

`func (o *SloDefinitionResponse) GetUpdatedAtOk() (*string, bool)`

GetUpdatedAtOk returns a tuple with the UpdatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUpdatedAt

`func (o *SloDefinitionResponse) SetUpdatedAt(v string)`

SetUpdatedAt sets UpdatedAt field to given value.


### GetVersion

`func (o *SloDefinitionResponse) GetVersion() float32`

GetVersion returns the Version field if non-nil, zero value otherwise.

### GetVersionOk

`func (o *SloDefinitionResponse) GetVersionOk() (*float32, bool)`

GetVersionOk returns a tuple with the Version field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetVersion

`func (o *SloDefinitionResponse) SetVersion(v float32)`

SetVersion sets Version field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


