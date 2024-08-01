package kbapi

import (
	"encoding/json"
	"fmt"
	"github.com/go-resty/resty/v2"
	log "github.com/sirupsen/logrus"
	"time"
)

const (
	basePathKibanaSynthetics = "/api/synthetics" // Base URL to access on Kibana save objects
	privateLocationsSuffix   = "/private_locations"
	monitorsSuffix           = "/monitors"
)

type MonitorID string
type MonitorType string
type MonitorLocation string
type MonitorSchedule int
type HttpMonitorMode string

type KibanaError struct {
	Code    int    `json:"statusCode,omitempty"`
	Error   string `json:"error,omitempty"`
	Message string `json:"message,omitempty"`
}

type JsonObject map[string]interface{}

const (
	Http    MonitorType = "http"
	Tcp     MonitorType = "tcp"
	Icmp    MonitorType = "icmp"
	Browser MonitorType = "browser"

	Every1Minute    MonitorSchedule = 1
	Every2Minutes                   = 2
	Every3Minutes                   = 3
	Every5Minutes                   = 5
	Every10Minutes                  = 10
	Every15Minutes                  = 15
	Every20Minutes                  = 20
	Every30Minutes                  = 30
	Every60Minutes                  = 60
	Every120Minutes                 = 120
	Every240Minutes                 = 240

	Japan         MonitorLocation = "japan"
	India                         = "india"
	Singapore                     = "singapore"
	AustraliaEast                 = "australia_east"
	UnitedKingdom                 = "united_kingdom"
	Germany                       = "germany"
	CanadaEast                    = "canada_east"
	Brazil                        = "brazil"
	USEast                        = "us_east"
	USWest                        = "us_west"

	ModeAll HttpMonitorMode = "all"
	ModeAny                 = "any"
)

type KibanaSyntheticsMonitorAPI struct {
	Add    KibanaSyntheticsMonitorAdd
	Delete KibanaSyntheticsMonitorDelete
	Get    KibanaSyntheticsMonitorGet
}

type KibanaSyntheticsPrivateLocationAPI struct {
	Create KibanaSyntheticsPrivateLocationCreate
	Delete KibanaSyntheticsPrivateLocationDelete
	Get    KibanaSyntheticsPrivateLocationGet
}

type SyntheticsStatusConfig struct {
	Enabled *bool `json:"enabled,omitempty"`
}

type MonitorAlertConfig struct {
	Status SyntheticsStatusConfig `json:"status,omitempty"`
	Tls    SyntheticsStatusConfig `json:"tls,omitempty"`
}

type HTTPMonitorFields struct {
	Url                 string          `json:"url"`
	SslSetting          JsonObject      `json:"ssl,omitempty"` //https://www.elastic.co/guide/en/beats/heartbeat/current/configuration-ssl.html
	MaxRedirects        int             `json:"max_redirects,omitempty"`
	Mode                HttpMonitorMode `json:"mode,omitempty"`
	Ipv4                *bool           `json:"ipv4,omitempty"`
	Ipv6                *bool           `json:"ipv6,omitempty"`
	Username            string          `json:"username,omitempty"`
	Password            string          `json:"password,omitempty"`
	ProxyHeader         JsonObject      `json:"proxy_headers,omitempty"`
	ProxyUrl            string          `json:"proxy_url,omitempty"`
	Response            JsonObject      `json:"response,omitempty"`
	ResponseIncludeBody *bool           `json:"response.include_body,omitempty"` //TODO: test with Response
	Check               JsonObject      `json:"check,omitempty"`
}

type SyntheticsMonitorConfig struct {
	Name             string              `json:"name"`
	Type             MonitorType         `json:"type"`
	Schedule         MonitorSchedule     `json:"schedule,omitempty"`
	Locations        []MonitorLocation   `json:"locations,omitempty"`
	PrivateLocations []string            `json:"private_locations,omitempty"`
	Enabled          *bool               `json:"enabled,omitempty"`
	Tags             []string            `json:"tags,omitempty"`
	Alert            *MonitorAlertConfig `json:"alert,omitempty"`
	APMServiceName   string              `json:"service.name,omitempty"`
	TimeoutSeconds   int                 `json:"timeout,omitempty"`
	Namespace        string              `json:"namespace,omitempty"`
	Params           string              `json:"params,omitempty"`
	RetestOnFailure  *bool               `json:"retest_on_failure,omitempty"`
}

type MonitorScheduleConfig struct {
	Number string `json:"number"`
	Unit   string `json:"unit"`
}

type SyntheticGeoConfig struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}

type MonitorLocationConfig struct {
	Id               string              `json:"id"`
	Label            string              `json:"label"`
	Geo              *SyntheticGeoConfig `json:"geo,omitempty"`
	IsServiceManaged bool                `json:"isServiceManaged"`
}

type SyntheticsMonitor struct {
	Name                        string                  `json:"name"`
	Type                        MonitorType             `json:"type"`
	ConfigId                    MonitorID               `json:"config_id"`
	Id                          MonitorID               `json:"id"`
	Mode                        HttpMonitorMode         `json:"mode"`
	CreatedAt                   time.Time               `json:"created_at"`
	UpdatedAt                   time.Time               `json:"updated_at"`
	Namespace                   string                  `json:"namespace"`
	Enabled                     *bool                   `json:"enabled,omitempty"`
	Alert                       *MonitorAlertConfig     `json:"alert,omitempty"`
	Schedule                    *MonitorScheduleConfig  `json:"schedule,omitempty"`
	Timeout                     string                  `json:"timeout,omitempty"`
	Locations                   []MonitorLocationConfig `json:"locations,omitempty"`
	Origin                      string                  `json:"origin,omitempty"`
	MaxAttempts                 int                     `json:"max_attempts"`
	MaxRedirects                string                  `json:"max_redirects"`
	ResponseIncludeBody         string                  `json:"response.include_body"`
	ResponseIncludeHeaders      bool                    `json:"response.include_headers"`
	CheckRequestMethod          string                  `json:"check.request.method"`
	ResponseIncludeBodyMaxBytes string                  `json:"response.include_body_max_bytes,omitempty"`
	Ipv4                        bool                    `json:"ipv4,omitempty"`
	Ipv6                        bool                    `json:"ipv6,omitempty"`
	SslVerificationMode         string                  `json:"ssl.verification_mode,omitempty"`
	SslSupportedProtocols       []string                `json:"ssl.supported_protocols,omitempty"`
	Revision                    int                     `json:"revision,omitempty"`
	Url                         string                  `json:"url,omitempty"`
	Ui                          struct {
		IsTlsEnabled bool `json:"is_tls_enabled"`
	} `json:"__ui,omitempty"`
}

type PrivateLocationConfig struct {
	Label         string              `json:"label"`
	AgentPolicyId string              `json:"agentPolicyId"`
	Tags          []string            `json:"tags,omitempty"`
	Geo           *SyntheticGeoConfig `json:"geo,omitempty"`
}

type PrivateLocation struct {
	Id string `json:"id"`
	PrivateLocationConfig
}

type MonitorDeleteStatus struct {
	Id      MonitorID `json:"id"`
	Deleted bool      `json:"deleted"`
}

type KibanaSyntheticsMonitorAdd func(config SyntheticsMonitorConfig, fields HTTPMonitorFields, namespace string) (*SyntheticsMonitor, error)

type KibanaSyntheticsMonitorGet func(id MonitorID, namespace string) (*SyntheticsMonitor, error)

type KibanaSyntheticsMonitorDelete func(namespace string, ids ...MonitorID) ([]MonitorDeleteStatus, error)

type KibanaSyntheticsPrivateLocationCreate func(pLoc PrivateLocationConfig, namespace string) (*PrivateLocation, error)

type KibanaSyntheticsPrivateLocationGet func(idOrLabel string, namespace string) (*PrivateLocation, error)

type KibanaSyntheticsPrivateLocationDelete func(id string, namespace string) error

func newKibanaSyntheticsPrivateLocationGetFunc(c *resty.Client) KibanaSyntheticsPrivateLocationGet {
	return func(idOrLabel string, namespace string) (*PrivateLocation, error) {

		path := fmt.Sprintf("%s/%s", basePath(namespace, privateLocationsSuffix), idOrLabel)
		log.Debugf("URL to get private locations: %s", path)
		resp, err := c.R().Get(path)
		if err = handleKibanaError(err, resp); err != nil {
			return nil, err
		}
		return unmarshal(resp, PrivateLocation{})
	}
}

func newKibanaSyntheticsPrivateLocationDeleteFunc(c *resty.Client) KibanaSyntheticsPrivateLocationDelete {
	return func(id string, namespace string) error {
		path := fmt.Sprintf("%s/%s", basePath(namespace, privateLocationsSuffix), id)
		log.Debugf("URL to delete private locations: %s", path)
		resp, err := c.R().Delete(path)
		err = handleKibanaError(err, resp)
		return err
	}
}

func newKibanaSyntheticsMonitorGetFunc(c *resty.Client) KibanaSyntheticsMonitorGet {
	return func(id MonitorID, namespace string) (*SyntheticsMonitor, error) {
		path := fmt.Sprintf("%s/%s", basePath(namespace, monitorsSuffix), id)
		log.Debugf("URL to create monitor: %s", path)

		resp, err := c.R().Get(path)
		if err := handleKibanaError(err, resp); err != nil {
			return nil, err
		}
		return unmarshal(resp, SyntheticsMonitor{})
	}
}

func newKibanaSyntheticsMonitorDeleteFunc(c *resty.Client) KibanaSyntheticsMonitorDelete {
	return func(namespace string, ids ...MonitorID) ([]MonitorDeleteStatus, error) {
		path := basePath(namespace, monitorsSuffix)
		log.Debugf("URL to delete monitors: %s", path)

		resp, err := c.R().SetBody(map[string]interface{}{
			"ids": ids,
		}).Delete(path)
		if err = handleKibanaError(err, resp); err != nil {
			return nil, err
		}

		result, err := unmarshal(resp, []MonitorDeleteStatus{})
		return *result, err
	}
}

func newKibanaSyntheticsPrivateLocationCreateFunc(c *resty.Client) KibanaSyntheticsPrivateLocationCreate {
	return func(pLoc PrivateLocationConfig, namespace string) (*PrivateLocation, error) {

		path := basePath(namespace, privateLocationsSuffix)
		log.Debugf("URL to create private locations: %s", path)
		resp, err := c.R().SetBody(pLoc).Post(path)
		if err = handleKibanaError(err, resp); err != nil {
			return nil, err
		}
		return unmarshal(resp, PrivateLocation{})
	}
}

func newKibanaSyntheticsMonitorAddFunc(c *resty.Client) KibanaSyntheticsMonitorAdd {
	return func(config SyntheticsMonitorConfig, fields HTTPMonitorFields, namespace string) (*SyntheticsMonitor, error) {

		path := basePath(namespace, monitorsSuffix)
		log.Debugf("URL to create monitor: %s", path)

		data := struct {
			SyntheticsMonitorConfig
			HTTPMonitorFields
		}{
			config,
			fields,
		}

		resp, err := c.R().SetBody(data).Post(path)
		if err := handleKibanaError(err, resp); err != nil {
			return nil, err
		}
		return unmarshal(resp, SyntheticsMonitor{})
	}
}

func unmarshal[T interface{}](resp *resty.Response, result T) (*T, error) {
	respBody := resp.Body()
	err := json.Unmarshal(respBody, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func handleKibanaError(err error, resp *resty.Response) error {
	if err != nil {
		return err
	}
	log.Debug("Response: ", resp)
	if resp.StatusCode() >= 300 {
		kibanaErr := KibanaError{}
		err := json.Unmarshal(resp.Body(), &kibanaErr)
		if err != nil {
			return NewAPIError(resp.StatusCode(), resp.Status(), err)
		}
		return NewAPIError(resp.StatusCode(), kibanaErr.Message, kibanaErr.Error)
	}
	return nil
}

func basePath(namespace, suffix string) string {
	return namespaceBasesPath(namespace, basePathKibanaSynthetics, suffix)
}

func namespaceBasesPath(namespace, basePath, suffix string) string {
	if namespace == "" || namespace == "default" {
		return fmt.Sprintf("%s%s", basePath, suffix)
	}

	return fmt.Sprintf("/s/%s%s%s", namespace, basePath, suffix)
}

//TODO: Monitor - Update https://www.elastic.co/guide/en/kibana/current/synthetics-apis.html
