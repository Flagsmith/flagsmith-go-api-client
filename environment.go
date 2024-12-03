package flagsmithapi

import (
	"fmt"
)

func (c *Client) GetEnvironment(apiKey string) (*Environment, error) {
	url := fmt.Sprintf("%s/environments/%s/", c.baseURL, apiKey)
	environment := Environment{}
	resp, err := c.client.R().
		SetResult(&environment).Get(url)

	if err != nil {
		return nil, err
	}

	if !resp.IsSuccess() {
		return nil, fmt.Errorf("flagsmithapi: Error getting environment: %s", resp)
	}

	return &environment, nil
}
func (c *Client) GetEnvironmentByUUID(uuid string) (*Environment, error) {
	url := fmt.Sprintf("%s/environments/get-by-uuid/%s/", c.baseURL, uuid)
	environment := Environment{}
	resp, err := c.client.R().
		SetResult(&environment).Get(url)

	if err != nil {
		return nil, err
	}

	if !resp.IsSuccess() {
		return nil, fmt.Errorf("flagsmithapi: Error getting environment: %s", resp)
	}

	return &environment, nil
}
func (c *Client) CreateEnvironment(environment *Environment) error {
	url := fmt.Sprintf("%s/environments/", c.baseURL)
	resp, err := c.client.R().SetBody(environment).SetResult(environment).Post(url)

	if err != nil {
		return err
	}

	if !resp.IsSuccess() {
		return fmt.Errorf("flagsmithapi: Error creating environment: %s", resp)
	}

	return nil
}
func (c *Client) UpdateEnvironment(environment *Environment) error {
	url := fmt.Sprintf("%s/environments/%s/", c.baseURL, environment.APIKey)
	resp, err := c.client.R().SetBody(environment).SetResult(environment).Put(url)

	if err != nil {
		return err
	}

	if !resp.IsSuccess() {
		return fmt.Errorf("flagsmithapi: Error creating environment: %s", resp)
	}

	return nil
}
func (c *Client) DeleteEnvironment(apiKey string) error {
	url := fmt.Sprintf("%s/environments/%s/", c.baseURL, apiKey)

	resp, err := c.client.R().Delete(url)
	if err != nil {
		return err
	}

	if !resp.IsSuccess() {
		return fmt.Errorf("flagsmithapi: Error deleting environment: %s", resp)
	}

	return nil
}
