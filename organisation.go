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
