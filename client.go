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

func (c *Client) GetProject(projectUUID string) (*Project, error) {
	url := fmt.Sprintf("%s/projects/", c.baseURL)
	result := []*Project{}
	resp, err := c.client.R().
		SetQueryParams(map[string]string{
			"uuid": projectUUID,
		}).
		SetResult(&result).
		Get(url)

	if err != nil {
		return nil, err
	}
	if !resp.IsSuccess() || len(result) != 1 {
		return nil, fmt.Errorf("flagsmithapi: Error getting project: %s", resp)
	}
	project := result[0]
	return project, nil

}

func (c *Client) GetFeature(featureUUID string) (*Feature, error) {
	url := fmt.Sprintf("%s/features/get-by-uuid/%s/", c.baseURL, featureUUID)
	feature := Feature{}
	resp, err := c.client.R().
		SetResult(&feature).Get(url)

	if err != nil {
		return nil, err
	}

	if !resp.IsSuccess() {
		return nil, fmt.Errorf("flagsmithapi: Error getting feature: %s", resp)
	}
	return &feature, nil
}

func (c *Client) CreateFeature(feature *Feature) error {
	projectID := feature.ProjectID
	if projectID == nil {
		project, err := c.GetProject(feature.ProjectUUID)
		if err != nil {
			return err
		}
		projectID = &project.ID
	}

	url := fmt.Sprintf("%s/projects/%d/features/", c.baseURL, *projectID)

	resp, err := c.client.R().SetBody(feature).SetResult(&feature).Post(url)

	if err != nil {
		return err
	}

	if !resp.IsSuccess() {
		return fmt.Errorf("flagsmithapi: Error creating feature: %s", resp)
	}

	return nil
}

func (c *Client) DeleteFeature(projectID, featureID int64) error {
	url := fmt.Sprintf("%s/projects/%d/features/%d/", c.baseURL, projectID, featureID)

	resp, err := c.client.R().Delete(url)

	if err != nil {
		return err
	}

	if !resp.IsSuccess() {
		return fmt.Errorf("flagsmithapi: Error deleting feature: %s", resp)
	}
	return nil
}

func (c *Client) UpdateFeature(feature *Feature) error {
	url := fmt.Sprintf("%s/projects/%d/features/%d/", c.baseURL, *feature.ProjectID, *feature.ID)
	resp, err := c.client.R().SetBody(feature).SetResult(feature).Put(url)

	if err != nil {
		return err
	}

	if !resp.IsSuccess() {
		return fmt.Errorf("flagsmithapi: Error updating feature: %s", resp)
	}

	return nil
}
