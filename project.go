package flagsmithapi

import (
	"fmt"
)

func (c *Client) GetProject(projectUUID string) (*Project, error) {
	url := fmt.Sprintf("%s/projects/get-by-uuid/%s/", c.baseURL, projectUUID)
	project := Project{}
	resp, err := c.client.R().
		SetResult(&project).
		Get(url)

	if err != nil {
		return nil, err
	}
	if !resp.IsSuccess() {
		return nil, fmt.Errorf("flagsmithapi: Error getting project: %s", resp)
	}
	return &project, nil

}
func (c *Client) GetProjectByID(projectID int64) (*Project, error) {
	url := fmt.Sprintf("%s/projects/%d/", c.baseURL, projectID)
	project := Project{}
	resp, err := c.client.R().
		SetResult(&project).
		Get(url)

	if err != nil {
		return nil, err
	}
	if !resp.IsSuccess() {
		return nil, fmt.Errorf("flagsmithapi: Error getting project: %s", resp)
	}
	return &project, nil

}

func (c *Client) CreateProject(project *Project) error {
	url := fmt.Sprintf("%s/projects/", c.baseURL)
	resp, err := c.client.R().SetBody(project).SetResult(project).Post(url)

	if err != nil {
		return err
	}

	if !resp.IsSuccess() {
		return fmt.Errorf("flagsmithapi: Error creating project: %s", resp)
	}

	return nil
}

func (c *Client) UpdateProject(project *Project) error {
	url := fmt.Sprintf("%s/projects/%d/", c.baseURL, project.ID)
	resp, err := c.client.R().SetBody(project).SetResult(project).Put(url)

	if err != nil {
		return err
	}

	if !resp.IsSuccess() {
		return fmt.Errorf("flagsmithapi: Error updating project: %s", resp)
	}

	return nil
}

func (c *Client) DeleteProject(projectID int64) error {
	url := fmt.Sprintf("%s/projects/%d/", c.baseURL, projectID)

	resp, err := c.client.R().Delete(url)
	if err != nil {
		return err
	}

	if !resp.IsSuccess() {
		return fmt.Errorf("flagsmithapi: Error deleting project: %s", resp)
	}

	return nil
}
