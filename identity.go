package flagsmithapi

import (
	"fmt"
)

func (c *Client) GetIdentity(environmentKey string, identityID int64) (*Identity, error) {
	url := fmt.Sprintf("%s/environments/%s/identities/%d/", c.baseURL, environmentKey, identityID)
	identity := Identity{}
	resp, err := c.client.R().SetResult(&identity).Get(url)
	if err != nil {
		return nil, err
	}

	if !resp.IsSuccess() {
		return nil, fmt.Errorf("flagsmithapi: Error deleting identity: %v", resp)
	}
	return &identity, nil
}
func (c *Client) CreateIdentity(environmentKey string, identity *Identity) error {
	url := fmt.Sprintf("%s/environments/%s/identities/", c.baseURL, environmentKey)

	resp, err := c.client.R().SetBody(identity).SetResult(&identity).Post(url)
	if err != nil {
		return err
	}
	if !resp.IsSuccess() {
		return fmt.Errorf("flagsmithapi: Error creating identity: %s", resp)
	}
	return nil

}
func (c *Client) DeleteIdentity(environmentKey string, identityID int64) error {
	url := fmt.Sprintf("%s/environments/%s/identities/%d/", c.baseURL, environmentKey, identityID)
	resp, err := c.client.R().Delete(url)
	if err != nil {
		return err
	}

	if !resp.IsSuccess() {
		return fmt.Errorf("flagsmithapi: Error deleting identity: %s", resp)
	}
	return nil

}
