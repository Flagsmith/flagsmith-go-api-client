package flagsmithapi

import (
	"fmt"
)

func (c *Client) GetServerSideEnvKeys(environmentKey string) ([]ServerSideEnvKey, error) {
	url := fmt.Sprintf("%s/environments/%s/api-keys/", c.baseURL, environmentKey)
	keys := []ServerSideEnvKey{}
	resp, err := c.client.R().SetResult(&keys).Get(url)
	if err != nil {
		return nil, err
	}

	if !resp.IsSuccess() {
		return nil, fmt.Errorf("flagsmithapi: Error fetching server side keys: %v", resp)
	}
	return keys, nil
}
func (c *Client) CreateServerSideEnvKey(environmentKey string, key *ServerSideEnvKey) error {
	url := fmt.Sprintf("%s/environments/%s/api-keys/", c.baseURL, environmentKey)

	resp, err := c.client.R().SetBody(key).SetResult(&key).Post(url)
	if err != nil {
		return err
	}
	if !resp.IsSuccess() {
		return fmt.Errorf("flagsmithapi: Error creating server side environment key: %s", resp)
	}
	return nil

}
func (c *Client) UpdateServerSideEnvKey(environmentKey string, key *ServerSideEnvKey) error {
	url := fmt.Sprintf("%s/environments/%s/api-keys/%d/", c.baseURL, environmentKey, key.ID)
	resp, err := c.client.R().SetBody(key).SetResult(&key).Put(url)
	if err != nil {
		return nil
	}

	if !resp.IsSuccess() {
		return fmt.Errorf("flagsmithapi: Error updating server side environment key: %s", resp)
	}
	return nil
}
func (c *Client) DeleteServerSideEnvKey(environmentKey string, keyID int64) error {
	url := fmt.Sprintf("%s/environments/%s/api-keys/%d/", c.baseURL, environmentKey, keyID)
	resp, err := c.client.R().Delete(url)
	if err != nil {
		return err
	}

	if !resp.IsSuccess() {
		return fmt.Errorf("flagsmithapi: Error deleting server side environment key: %s", resp)
	}
	return nil

}
