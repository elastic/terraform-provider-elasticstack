# RunConnectorSubactionCreatealertSubActionParams

## Properties
Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Actions** | **[]string** | The custom actions available to the alert. | [optional] [default to null]
**Alias** | **string** | The unique identifier used for alert deduplication in Opsgenie. | [optional] [default to null]
**Description** | **string** | A description that provides detailed information about the alert. | [optional] [default to null]
**Details** | [**ModelMap**](interface{}.md) | The custom properties of the alert. | [optional] [default to null]
**Entity** | **string** | The domain of the alert. For example, the application or server name. | [optional] [default to null]
**Message** | **string** | The alert message. | [default to null]
**Note** | **string** | Additional information for the alert. | [optional] [default to null]
**Priority** | **string** | The priority level for the alert. | [optional] [default to null]
**Responders** | [**[]RunConnectorSubactionCreatealertSubActionParamsResponders**](run_connector_subaction_createalert_subActionParams_responders.md) | The entities to receive notifications about the alert. If &#x60;type&#x60; is &#x60;user&#x60;, either &#x60;id&#x60; or &#x60;username&#x60; is required. If &#x60;type&#x60; is &#x60;team&#x60;, either &#x60;id&#x60; or &#x60;name&#x60; is required.  | [optional] [default to null]
**Source** | **string** | The display name for the source of the alert. | [optional] [default to null]
**Tags** | **[]string** | The tags for the alert. | [optional] [default to null]
**User** | **string** | The display name for the owner. | [optional] [default to null]
**VisibleTo** | [**[]RunConnectorSubactionCreatealertSubActionParamsVisibleTo**](run_connector_subaction_createalert_subActionParams_visibleTo.md) | The teams and users that the alert will be visible to without sending a notification. Only one of &#x60;id&#x60;, &#x60;name&#x60;, or &#x60;username&#x60; is required. | [optional] [default to null]

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)

