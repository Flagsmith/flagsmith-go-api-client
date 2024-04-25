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
		return nil, fmt.Errorf("flagsmithapi: Error fetching identity: %v", resp)
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
func (c *Client) GetTraits(environmentKey string, identityID int64) ([]Trait, error) {
	url := fmt.Sprintf("%s/environments/%s/identities/%d/traits/", c.baseURL, environmentKey, identityID)
	result := struct {
		Traits []Trait `json:"results"`
	}{}
	resp, err := c.client.R().SetResult(&result).Get(url)
	if err != nil {
		return nil, err
	}

	if !resp.IsSuccess() {
		return nil, fmt.Errorf("flagsmithapi: Error fetching traits: %v", resp)
	}
	return result.Traits, nil
}

func (c *Client) CreateTrait(environmentKey string, identityID int64, trait *Trait) error {
	url := fmt.Sprintf("%s/environments/%s/identities/%d/traits/", c.baseURL, environmentKey, identityID)

	resp, err := c.client.R().
		SetBody(trait).
		SetResult(&trait).
		Post(url)

	if err != nil {
		return err
	}

	if !resp.IsSuccess() {
		return fmt.Errorf("flagsmithapi: Error creating trait: %v", resp)
	}

	return nil
}

func (c *Client) UpdateTrait(environmentKey string, identityID int64, trait *Trait) error {
	url := fmt.Sprintf("%s/environments/%s/identities/%d/traits/%d/", c.baseURL, environmentKey, identityID, trait.ID)

	resp, err := c.client.R().
		SetBody(trait).
		Put(url)

	if err != nil {
		return err
	}

	if !resp.IsSuccess() {
		return fmt.Errorf("flagsmithapi: Error updating trait: %v", resp)
	}

	return nil
}

func (c *Client) DeleteTrait(environmentKey string, identityID int64, traitID int64) error {
	url := fmt.Sprintf("%s/environments/%s/identities/%d/traits/%d/", c.baseURL, environmentKey, identityID, traitID)

	resp, err := c.client.R().
		Delete(url)

	if err != nil {
		return err
	}

	if !resp.IsSuccess() {
		return fmt.Errorf("flagsmithapi: Error deleting trait: %v", resp)
	}

	return nil
}
