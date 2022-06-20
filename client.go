package flagsmithapi

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/go-resty/resty/v2"
)

type Client struct {
	master_api_key string
	baseURL        string
	client         *resty.Client
}

func NewClient(masterAPIKey string, baseURL string) *Client {
	c := &Client{master_api_key: masterAPIKey, baseURL: baseURL, client: resty.New()}
	c.client.SetHeaders(map[string]string{
		"Accept":        "application/json",
		"Authorization": "Api-Key " + c.master_api_key,
	})
	return c

}

func (c *Client) GetFeatureStates(environmentID int) (*[]FeatureState, error) {
	url := fmt.Sprintf("%s/features/featurestates/", c.baseURL)
	var getFeatureStatesResponse struct {
		Results *[]FeatureState `json:"results"`
		Count   int             `json:"count"`
	}
	resp, err := c.client.R().
		SetQueryParams(map[string]string{
			"environment": strconv.Itoa(environmentID),
		}).
		Get(url)
	fmt.Println("error -> ", err)
	if err != nil {
		return nil, err
	}
	error := json.Unmarshal(resp.Body(), &getFeatureStatesResponse)
	if error != nil {
		fmt.Println("error reading body to json -> ", error)
		return nil, error
	}

	featureStates := getFeatureStatesResponse.Results

	fmt.Println(" response -> ", getFeatureStatesResponse)
	fmt.Println("feature State-> :?", featureStates)

	return featureStates, nil

}

func (c *Client) GetFeatureState(featureStateID int64) (*FeatureState, error) {
	url := fmt.Sprintf("%s/features/featurestates/%d/", c.baseURL, featureStateID)
	fmt.Println("making request with", url)

	resp, err := c.client.R().
		SetResult(FeatureState{}).
		Get(url)
	if err != nil {
		fmt.Println("error -> ", err)
		return nil, err
	}
	if resp.IsSuccess() {
		return nil, fmt.Errorf("Error getting feature state: %s", resp.Status())

	}
	featureState := resp.Result().(*FeatureState)
	return featureState, nil

}

func (c *Client) DeleteFeatureState(featureStateID int) error {
	url := fmt.Sprintf("%s/features/featurestates/%d/", c.baseURL, featureStateID)
	fmt.Println("making request with", url)

	resp, err := c.client.R().Delete(url)
	// TODO: Add error handling
	// if response.IsSuccess{

	// }
	// featureState := resp.Result().(*FeatureState)

	fmt.Println(" response -> ", resp)
	fmt.Println("error -> ", err)
	return err

}

// Update Feature State
func (c *Client) UpdateFeatureState(featureState *FeatureState) (*FeatureState, error) {
	url := fmt.Sprintf("%s/features/featurestates/%d/", c.baseURL, featureState.ID)

	resp, err := c.client.R().SetBody(featureState).SetResult(FeatureState{}).Put(url)

	if err != nil {
		return nil, err
	}
	if !resp.IsSuccess() {
		return nil, fmt.Errorf("Error updating feature state: %s", resp.Status())
	}
	updatedFeatureState, ok := resp.Result().(*FeatureState)
	if !ok {
		return nil, fmt.Errorf("unable to cast result to FeatureState")
	}
	return updatedFeatureState, nil
}
