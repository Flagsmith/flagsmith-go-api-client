package flagsmithapi

import (
	"fmt"
	"net/http"
	"strconv"

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
func (c *Client) GetEnvironmentFeatureState(environmentKey string, featureID int64) (*FeatureState, error) {
	url := fmt.Sprintf("%s/environments/%s/featurestates/", c.baseURL, environmentKey)
	result := struct {
		Results []*FeatureState `json:"results"`
	}{}
	resp, err := c.client.R().
		SetQueryParams(map[string]string{
			"feature": strconv.FormatInt(featureID, 10),
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
func (c *Client) GetFeatureState(featureStateUUID string) (*FeatureState, error) {
	url := fmt.Sprintf("%s/features/featurestates/get-by-uuid/%s/", c.baseURL, featureStateUUID)
	featureState := FeatureState{}
	resp, err := c.client.R().
		SetResult(&featureState).Get(url)

	if err != nil {
		return nil, err
	}

	if !resp.IsSuccess() {
		if resp.StatusCode() == http.StatusNotFound {
			return nil, FeatureStateNotFoundError{featureStateUUID: featureStateUUID}
		}
		return nil, fmt.Errorf("flagsmithapi: Error getting feature state: %s", resp)
	}
	if featureState.FeatureSegment != nil {
		// load feature segment data
		featureSegment, err := c.GetFeatureSegmentByID(*featureState.FeatureSegment)
		if err != nil {
			return nil, err
		}
		featureState.Segment = featureSegment.Segment
		featureState.SegmentPriority = featureSegment.Priority
	}

	return &featureState, nil
}

// Update the feature state
func (c *Client) UpdateFeatureState(featureState *FeatureState, updateSegmentPriority bool) error {
	url := fmt.Sprintf("%s/features/featurestates/%d/", c.baseURL, featureState.ID)
	resp, err := c.client.R().SetBody(featureState).SetResult(&featureState).Put(url)
	if err != nil {
		return err
	}
	if !resp.IsSuccess() {
		return fmt.Errorf("flagsmithapi: Error updating feature state: %s", resp.Status())
	}

	if updateSegmentPriority && featureState.FeatureSegment != nil {
		SegmentPriority := featureState.SegmentPriority
		Segment := featureState.Segment

		// Update segment priority
		err := c.UpdateFeatureSegmentPriority(*featureState.FeatureSegment, *SegmentPriority)
		if err != nil {
			return err
		}
		featureState.SegmentPriority = SegmentPriority
		featureState.Segment = Segment

	}

	return nil
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
		if resp.StatusCode() == http.StatusNotFound {
			return nil, FeatureNotFoundError{featureUUID: featureUUID}
		}
		return nil, fmt.Errorf("flagsmithapi: Error getting feature: %s", resp)
	}
	project, err := c.GetProjectByID(*feature.ProjectID)
	if err != nil {
		return nil, err
	}
	feature.ProjectUUID = project.UUID
	return &feature, nil
}

func (c *Client) CreateFeature(feature *Feature) error {
	if feature.ProjectID == nil {
		projectID, err := c.getProjectID(feature.ProjectUUID)
		if err != nil {
			return err
		}
		feature.ProjectID = &projectID
	}

	url := fmt.Sprintf("%s/projects/%d/features/", c.baseURL, *feature.ProjectID)

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

func (c *Client) getProjectID(projectUUID string) (int64, error) {
	project, err := c.GetProject(projectUUID)

	if err != nil {
		return 0, err
	}
	return project.ID, nil
}

func (c *Client) manageFeatureOwners(feature *Feature, ownerIDs []int64, endpoint string) (*resty.Response, error) {
	url := fmt.Sprintf("%s/projects/%d/features/%d/%s/", c.baseURL, *feature.ProjectID, *feature.ID, endpoint)
	body := struct {
		UserIDs []int64 `json:"user_ids"`
	}{
		UserIDs: ownerIDs,
	}
	resp, err := c.client.R().SetBody(body).Post(url)
	return resp, err

}

func (c *Client) AddFeatureOwners(feature *Feature, ownerIDs []int64) error {
	resp, err := c.manageFeatureOwners(feature, ownerIDs, "add-owners")
	if err != nil {
		return err
	}
	if !resp.IsSuccess() {
		return fmt.Errorf("flagsmithapi: Error adding feature owners: %s", resp)
	}
	return nil

}

func (c *Client) RemoveFeatureOwners(feature *Feature, ownerIDs []int64) error {
	resp, err := c.manageFeatureOwners(feature, ownerIDs, "remove-owners")
	if err != nil {
		return nil
	}
	if !resp.IsSuccess() {
		return fmt.Errorf("flagsmithapi: Error removing feature owners: %s", resp)
	}
	return nil
}

func (c *Client) GetFeatureMVOption(featureUUID, mvOptionUUID string) (*FeatureMultivariateOption, error) {
	url := fmt.Sprintf("%s/multivariate/options/get-by-uuid/%s/", c.baseURL, mvOptionUUID)
	featureMVOption := FeatureMultivariateOption{}
	resp, err := c.client.R().
		SetResult(&featureMVOption).
		Get(url)

	if err != nil {
		return nil, err
	}

	if !resp.IsSuccess() {
		if resp.StatusCode() == http.StatusNotFound {
			return nil, FeatureMVOptionNotFoundError{featureMVOptionUUID: mvOptionUUID}
		}
		return nil, fmt.Errorf("flagsmithapi: Error getting feature MV option: %s", resp)
	}
	feature, err := c.GetFeature(featureUUID)
	if err != nil {
		return nil, err
	}
	featureMVOption.FeatureUUID = featureUUID
	featureMVOption.ProjectID = feature.ProjectID

	return &featureMVOption, nil
}

func (c *Client) DeleteFeatureMVOption(projectID, featureID, mvOptionID int64) error {
	url := fmt.Sprintf("%s/projects/%d/features/%d/mv-options/%d/", c.baseURL, projectID, featureID, mvOptionID)

	resp, err := c.client.R().Delete(url)

	if err != nil {
		return err
	}

	if !resp.IsSuccess() {
		return fmt.Errorf("flagsmithapi: Error deleting feature MV option: %s", resp)
	}
	return nil
}

func (c *Client) UpdateFeatureMVOption(featureMVOption *FeatureMultivariateOption) error {
	url := fmt.Sprintf("%s/projects/%d/features/%d/mv-options/%d/", c.baseURL, *featureMVOption.ProjectID,
		*featureMVOption.FeatureID, featureMVOption.ID)
	resp, err := c.client.R().SetBody(featureMVOption).SetResult(featureMVOption).Put(url)

	if err != nil {
		return err
	}

	if !resp.IsSuccess() {
		return fmt.Errorf("flagsmithapi: Error updating feature MV option: %s", resp)
	}

	return nil
}

func (c *Client) CreateFeatureMVOption(featureMVOption *FeatureMultivariateOption) error {
	if featureMVOption.FeatureID == nil {
		feature, err := c.GetFeature(featureMVOption.FeatureUUID)
		if err != nil {
			return err
		}
		featureMVOption.FeatureID = feature.ID
		featureMVOption.ProjectID = feature.ProjectID
	}

	url := fmt.Sprintf("%s/projects/%d/features/%d/mv-options/", c.baseURL, *featureMVOption.ProjectID,
		*featureMVOption.FeatureID)

	resp, err := c.client.R().SetBody(featureMVOption).SetResult(&featureMVOption).Post(url)

	if err != nil {
		return err
	}

	if !resp.IsSuccess() {
		return fmt.Errorf("flagsmithapi: Error creating feature MV option: %s", resp)
	}

	return nil
}

func (c *Client) GetSegment(segmentUUID string) (*Segment, error) {
	url := fmt.Sprintf("%s/segments/get-by-uuid/%s/", c.baseURL, segmentUUID)
	segment := Segment{}
	resp, err := c.client.R().
		SetResult(&segment).Get(url)

	if err != nil {
		return nil, err
	}

	if !resp.IsSuccess() {
		if resp.StatusCode() == http.StatusNotFound {
			return nil, SegmentNotFoundError{segmentUUID: segmentUUID}
		}
		return nil, fmt.Errorf("flagsmithapi: Error getting segment: %s", resp)
	}
	project, err := c.GetProjectByID(*segment.ProjectID)
	if err != nil {
		return nil, err
	}
	segment.ProjectUUID = project.UUID
	return &segment, nil
}
func (c *Client) DeleteSegment(projectID, segmentID int64) error {
	url := fmt.Sprintf("%s/projects/%d/segments/%d/", c.baseURL, projectID, segmentID)

	resp, err := c.client.R().Delete(url)
	if err != nil {
		return err
	}

	if !resp.IsSuccess() {
		return fmt.Errorf("flagsmithapi: Error deleting segment: %s", resp)
	}
	return nil
}
func (c *Client) CreateSegment(segment *Segment) error {
	projectID := segment.ProjectID
	if projectID == nil {
		project, err := c.GetProject(segment.ProjectUUID)
		if err != nil {
			return err
		}
		projectID = &project.ID
	}
	segment.ProjectID = projectID

	url := fmt.Sprintf("%s/projects/%d/segments/", c.baseURL, *projectID)
	resp, err := c.client.R().SetBody(segment).SetResult(segment).Post(url)

	if err != nil {
		return err
	}

	if !resp.IsSuccess() {
		return fmt.Errorf("flagsmithapi: Error creating segment: %s", resp)
	}

	return nil
}
func (c *Client) UpdateSegment(segment *Segment) error {
	projectID := segment.ProjectID
	if projectID == nil {
		project, err := c.GetProject(segment.ProjectUUID)
		if err != nil {
			return err
		}
		projectID = &project.ID
	}
	segment.ProjectID = projectID

	url := fmt.Sprintf("%s/projects/%d/segments/%d/", c.baseURL, *projectID, *segment.ID)
	resp, err := c.client.R().SetBody(segment).SetResult(segment).Put(url)

	if err != nil {
		return err
	}

	if !resp.IsSuccess() {
		return fmt.Errorf("flagsmithapi: Error updating segment: %s", resp)
	}

	return nil
}

func (c *Client) GetFeatureSegmentByID(featureSegmentID int64) (*FeatureSegment, error) {
	url := fmt.Sprintf("%s/features/feature-segments/%d/", c.baseURL, featureSegmentID)
	featureSegment := FeatureSegment{}
	resp, err := c.client.R().
		SetResult(&featureSegment).Get(url)

	if err != nil {
		return nil, err
	}

	if !resp.IsSuccess() {
		return nil, fmt.Errorf("flagsmithapi: Error getting feature segment: %s", resp)
	}
	return &featureSegment, nil
}

func (c *Client) UpdateFeatureSegmentPriority(featureSegmentID, priority int64) error {
	body := []struct {
		Priority int64 `json:"priority"`
		ID       int64 `json:"id"`
	}{
		{
			Priority: priority,
			ID:       featureSegmentID,
		},
	}
	url := fmt.Sprintf("%s/features/feature-segments/update-priorities/", c.baseURL)
	resp, err := c.client.R().SetBody(body).Post(url)

	if err != nil {
		return err
	}

	if !resp.IsSuccess() {
		return fmt.Errorf("flagsmithapi: Error updating feature segment priority: %s", resp)
	}

	return nil
}

func (c *Client) DeleteFeatureSegment(featureSegmentID int64) error {
	url := fmt.Sprintf("%s/features/feature-segments/%d/", c.baseURL, featureSegmentID)

	resp, err := c.client.R().Delete(url)
	if err != nil {
		return err
	}

	if !resp.IsSuccess() {
		return fmt.Errorf("flagsmithapi: Error deleting feature segment: %s", resp)
	}
	return nil
}

func (c *Client) CreateFeatureSegment(featureSegment *FeatureSegment) error {
	url := fmt.Sprintf("%s/features/feature-segments/", c.baseURL)
	resp, err := c.client.R().SetBody(featureSegment).SetResult(featureSegment).Post(url)

	if err != nil {
		return err
	}

	if !resp.IsSuccess() {
		return fmt.Errorf("flagsmithapi: Error creating feature segment: %s", resp)
	}

	return nil
}

func (c *Client) CreateSegmentOverride(featureState *FeatureState) error {
	// fetch and set environment
	environnmetKey := featureState.EnvironmentKey
	environment, err := c.GetEnvironment(environnmetKey)
	if err != nil {
		return err
	}
	featureState.Environment = &environment.ID

	// Create and set feature segment
	featureSegment := FeatureSegment{
		Feature:     featureState.Feature,
		Environment: environment.ID,
		Segment:     featureState.Segment,
		Priority:    featureState.SegmentPriority,
	}

	err = c.CreateFeatureSegment(&featureSegment)
	if err != nil {
		return err
	}
	featureState.FeatureSegment = featureSegment.ID

	// Finally, create the feature state
	url := fmt.Sprintf("%s/features/featurestates/", c.baseURL)
	resp, err := c.client.R().SetBody(featureState).SetResult(&featureState).Post(url)
	if err != nil {
		return err
	}

	if !resp.IsSuccess() {
		return fmt.Errorf("flagsmithapi: Error creating segment override feature state: %s", resp.Status())
	}

	return nil

}

func (c *Client) GetTag(projectUUID string, tagUUID string) (*Tag, error) {
	projectID, err := c.getProjectID(projectUUID)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/projects/%d/tags/get-by-uuid/%s/", c.baseURL, projectID, tagUUID)
	tag := Tag{}
	resp, err := c.client.R().
		SetResult(&tag).Get(url)

	if err != nil {
		return nil, err
	}

	if !resp.IsSuccess() {
		return nil, fmt.Errorf("flagsmithapi: Error getting tag: %s", resp)
	}
	tag.ProjectUUID = projectUUID
	return &tag, nil
}
func (c *Client) CreateTag(tag *Tag) error {
	if tag.ProjectID == nil {
		projectID, err := c.getProjectID(tag.ProjectUUID)
		if err != nil {
			return err
		}
		tag.ProjectID = &projectID
	}
	url := fmt.Sprintf("%s/projects/%d/tags/", c.baseURL, *tag.ProjectID)
	resp, err := c.client.R().SetBody(tag).SetResult(tag).Post(url)

	if err != nil {
		return err
	}

	if !resp.IsSuccess() {
		return fmt.Errorf("flagsmithapi: Error creating tag: %s", resp)
	}

	return nil
}

func (c *Client) UpdateTag(tag *Tag) error {
	url := fmt.Sprintf("%s/projects/%d/tags/%d/", c.baseURL, *tag.ProjectID, *tag.ID)
	resp, err := c.client.R().SetBody(tag).SetResult(tag).Put(url)

	if err != nil {
		return err
	}

	if !resp.IsSuccess() {
		return fmt.Errorf("flagsmithapi: Error updating tag: %s", resp)
	}

	return nil
}

func (c *Client) DeleteTag(projectID, tagID int64) error {
	url := fmt.Sprintf("%s/projects/%d/tags/%d/", c.baseURL, projectID, tagID)

	resp, err := c.client.R().Delete(url)
	if err != nil {
		return err
	}

	if !resp.IsSuccess() {
		return fmt.Errorf("flagsmithapi: Error deleting tag: %s", resp)
	}
	return nil
}
