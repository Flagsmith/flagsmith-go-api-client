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

func (c *Client) GetFeatureStates(environmentID int) *[]FeatureState {
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
	// TODO: Add error handling
	// if response.IsSuccess{

	// }

	error := json.Unmarshal(resp.Body(), &getFeatureStatesResponse)

	fmt.Println("error -> ", error)
	featureStates := getFeatureStatesResponse.Results

	fmt.Println(" response -> ", getFeatureStatesResponse)
	fmt.Println("feature State-> :?", featureStates)

	return featureStates

}

//TODO: remove environment_id
func (c *Client) GetFeatureState(featureStateID int, environmentID int) *FeatureState {
	url := fmt.Sprintf("%s/features/featurestates/%d/", c.baseURL, featureStateID)
	fmt.Println("making request with", url)

	resp, err := c.client.R().
		SetQueryParams(map[string]string{
			"environment": strconv.Itoa(environmentID),
		}).
		SetResult(FeatureState{}).
		Get(url)
	// TODO: Add error handling
	// if response.IsSuccess{

	// }
	featureState := resp.Result().(*FeatureState)
	fmt.Println(" response -> ", resp)

	fmt.Println("error -> ", err)
	fmt.Println("feature State-> :?", featureState.ID)
	fmt.Println("feature State-> :?", featureState.FeatureStateValue)
	return featureState

}

//TODO: remove environment_id
func (c *Client) DeleteFeatureState(featureStateID int, environmentID int) error {
	url := fmt.Sprintf("%s/features/featurestates/%d/", c.baseURL, featureStateID)
	fmt.Println("making request with", url)

	resp, err := c.client.R().
		SetQueryParams(map[string]string{
			"environment": strconv.Itoa(environmentID),
		}).
		Delete(url)
	// TODO: Add error handling
	// if response.IsSuccess{

	// }
	// featureState := resp.Result().(*FeatureState)

	fmt.Println(" response -> ", resp)
	fmt.Println("error -> ", err)
	return err

	// fmt.Println("feature State-> :?", featureState.ID)
	// fmt.Println("feature State-> :?", featureState.FeatureStateValue)
	// return featureState

}
