package flagsmithapi

import (
	"fmt"
)

func (c *Client) GetOrganisationByUUID(orgUUID string) (*Organisation, error) {
	url := fmt.Sprintf("%s/organisations/get-by-uuid/%s/", c.baseURL, orgUUID)
	organisation := Organisation{}
	resp, err := c.client.R().
		SetResult(&organisation).
		Get(url)

	if err != nil {
		return nil, err
	}
	if !resp.IsSuccess() {
		return nil, fmt.Errorf("flagsmithapi: Error getting organisation: %s", resp)
	}
	return &organisation, nil

}

func (c *Client) GetOrganisationUsers(orgID int64) ([]User, error) {
	url := fmt.Sprintf("%s/organisations/%d/users/", c.baseURL, orgID)
	users := []User{}
	resp, err := c.client.R().
		SetResult(&users).
		Get(url)

	if err != nil {
		return nil, err
	}
	if !resp.IsSuccess() {
		return nil, fmt.Errorf("flagsmithapi: Error getting organisation users: %s", resp)
	}
	return users, nil
}

func (c *Client) GetOrganisationUserByEmail(orgID int64, email string) (*User, error) {
	users, err := c.GetOrganisationUsers(orgID)
	if err != nil {
		return nil, err
	}
	for i := range users {
		if users[i].Email == email {
			return &users[i], nil
		}
	}
	return nil, UserNotFoundError{email: email}
}
