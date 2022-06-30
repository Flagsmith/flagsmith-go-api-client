package flagsmithapi

import (
	"fmt"

	"github.com/go-resty/resty/v2"
)

const BaseAPIURL = "https://api.flagsmith.com/api/v1"

type Client struct {
	master_api_key string
	baseURL        string
	client         *resty.Client
}

func NewClient(masterAPIKey string, baseURL string) *Client {
	if baseURL == "" {
		baseURL = BaseAPIURL
	}
	c := &Client{master_api_key: masterAPIKey, baseURL: baseURL, client: resty.New()}
	c.client.SetHeaders(map[string]string{
		"Accept":        "application/json",
		"Content-type":  "application/json",
		"Authorization": "Api-Key " + c.master_api_key,
	})
	return c

}

// Get the feature state associated with the environment for a given feature
func (c *Client) GetEnvironmentFeatureState(environmentAPIKey string, featureName string) (*FeatureState, error) {
	url := fmt.Sprintf("%s/environments/%s/featurestates/", c.baseURL, environmentAPIKey)
	result := struct {
		Results []*FeatureState `json:"results"`
	}{}
	resp, err := c.client.R().
		SetQueryParams(map[string]string{
			"feature_name": featureName,
		}).
		SetResult(&result).
		Get(url)

	if err != nil {
		return nil, err
	}
	if !resp.IsSuccess() || len(result.Results) != 1 {
		return nil, fmt.Errorf("flagsmithapi: Failed to get feature state")

	}
	featureState := result.Results[0]
	return featureState, nil

}

// Update the feature state
func (c *Client) UpdateFeatureState(featureState *FeatureState) (*FeatureState, error) {
	url := fmt.Sprintf("%s/features/featurestates/%d/", c.baseURL, featureState.ID)
	updatedFeatureState := FeatureState{}
	resp, err := c.client.R().SetBody(featureState).SetResult(&updatedFeatureState).Put(url)
	if err != nil {
		return nil, err
	}
	if !resp.IsSuccess() {
		return nil, fmt.Errorf("flagsmithapi: Error updating feature state: %s", resp.Status())
	}
	return &updatedFeatureState, nil
}
