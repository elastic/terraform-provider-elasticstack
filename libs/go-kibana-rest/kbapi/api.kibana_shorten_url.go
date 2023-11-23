package kbapi

import (
	"encoding/json"

	"github.com/go-resty/resty/v2"
	log "github.com/sirupsen/logrus"
)

const (
	basePathKibanaShortenURL = "/api/short_url" // Base URL to access on Kibana shorten URL
)

// ShortenURL is the shorten URL object
type ShortenURL struct {
	LocatorId         string         `json:"locatorId"`
	Params            map[string]any `json:"params"`
	Slug              string         `json:"slug,omitempty"`
	HumanReadableSlug bool           `json:"humanReadableSlug,omitempty"`
}

// ShortenURLResponse is the shorten URL object response
type ShortenURLResponse struct {
	ID      string      `json:"id"`
	Locator *ShortenURL `json:"locator"`
}

// KibanaShortenURLCreate permit to create new shorten URL
type KibanaShortenURLCreate func(shortenURL *ShortenURL) (*ShortenURLResponse, error)

// String permit to return ShortenURL object as JSON string
func (o *ShortenURL) String() string {
	json, _ := json.Marshal(o)
	return string(json)
}

// String permit to return ShortenURLResponse object as JSON string
func (o *ShortenURLResponse) String() string {
	json, _ := json.Marshal(o)
	return string(json)
}

// newKibanaShortenURLCreateFunc permit to create new shorten URL
func newKibanaShortenURLCreateFunc(c *resty.Client) KibanaShortenURLCreate {
	return func(shortenURL *ShortenURL) (*ShortenURLResponse, error) {

		if shortenURL == nil {
			return nil, NewAPIError(600, "You must provide shorten URL object")
		}
		log.Debug("Shorten URL: ", shortenURL)

		jsonData, err := json.Marshal(shortenURL)
		if err != nil {
			return nil, err
		}

		log.Debugf("Shorten URL payload: %s", jsonData)

		resp, err := c.R().SetBody(jsonData).Post(basePathKibanaShortenURL)
		if err != nil {
			return nil, err
		}

		log.Debug("Response: ", resp)
		if resp.StatusCode() >= 300 {
			return nil, NewAPIError(resp.StatusCode(), resp.Status())
		}

		shortenURLResponse := &ShortenURLResponse{}
		err = json.Unmarshal(resp.Body(), shortenURLResponse)
		if err != nil {
			return nil, err
		}
		log.Debug("ShortenURLResponse: ", shortenURLResponse)

		return shortenURLResponse, nil
	}
}
