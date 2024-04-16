package flagsmithapi

import (
	"encoding/json"
	"log"
	"time"
)

type Project struct {
	ID                             int64  `json:"id,omitempty"`
	UUID                           string `json:"uuid,omitempty"`
	Name                           string `json:"name"`
	Organisation                   int64  `json:"organisation"`
	HideDisabledFlags              bool   `json:"hide_disabled_flags,omitempty"`
	PreventFlagDefaults            bool   `json:"prevent_flag_defaults,omitempty"`
	OnlyAllowLowerCaseFeatureNames bool   `json:"only_allow_lower_case_feature_names,omitempty"`
	FeatureNameRegex               bool   `json:"feature_name_regex,omitempty"`
}

type FeatureMultivariateOption struct {
	ID                          int64   `json:"id,omitempty"`
	Type                        string  `json:"type"`
	UUID                        string  `json:"uuid,omitempty"`
	FeatureID                   *int64  `json:"feature,omitempty"`
	IntegerValue                *int64  `json:"integer_value,omitempty"`
	StringValue                 *string `json:"string_value,omitempty"`
	BooleanValue                *bool   `json:"boolean_value,omitempty"`
	DefaultPercentageAllocation float64 `json:"default_percentage_allocation"`

	FeatureUUID string `json:"-"`
	ProjectID   *int64 `json:"-"`
}

type Feature struct {
	Name           string   `json:"name"`
	UUID           string   `json:"uuid,omitempty"`
	ID             *int64   `json:"id,omitempty"`
	Type           *string  `json:"type,omitempty"`
	Description    *string  `json:"description,omitempty"`
	InitialValue   string   `json:"initial_value,omitempty"`
	DefaultEnabled bool     `json:"default_enabled,omitempty"`
	IsArchived     bool     `json:"is_archived,omitempty"`
	Owners         *[]int64 `json:"owners,omitempty"`
	Tags           []int64  `json:"tags"`

	ProjectUUID string `json:"-"`
	ProjectID   *int64 `json:"project,omitempty"`
}

func (f *Feature) UnmarshalJSON(data []byte) error {
	type owner struct {
		ID int64 `json:"id"`
	}
	var obj struct {
		Name           string  `json:"name"`
		UUID           string  `json:"uuid,omitempty"`
		ID             *int64  `json:"id,omitempty"`
		Type           *string `json:"type,omitempty"`
		Description    *string `json:"description,omitempty"`
		InitialValue   string  `json:"initial_value,omitempty"`
		DefaultEnabled bool    `json:"default_enabled,omitempty"`
		IsArchived     bool    `json:"is_archived,omitempty"`
		Owners         []owner `json:"owners,omitempty"`
		ProjectID      *int64  `json:"project,omitempty"`
		Tags           []int64 `json:"tags"`
	}

	err := json.Unmarshal(data, &obj)

	if err != nil {
		return err
	}

	f.Name = obj.Name
	f.UUID = obj.UUID
	f.ID = obj.ID
	f.Type = obj.Type
	f.Description = obj.Description
	f.InitialValue = obj.InitialValue
	f.DefaultEnabled = obj.DefaultEnabled
	f.IsArchived = obj.IsArchived
	f.ProjectID = obj.ProjectID
	f.Tags = obj.Tags
	if obj.Owners != nil {
		f.Owners = &[]int64{}
		for _, o := range obj.Owners {
			*f.Owners = append(*f.Owners, o.ID)
		}
	}
	return nil
}

type FeatureStateValue struct {
	Type         string  `json:"type"`
	StringValue  *string `json:"string_value"`
	IntegerValue *int64  `json:"integer_value"`
	BooleanValue *bool   `json:"boolean_value"`
}

type FeatureState struct {
	ID                int64              `json:"id,omitempty"`
	UUID              string             `json:"uuid,omitempty"`
	FeatureStateValue *FeatureStateValue `json:"feature_state_value"`
	Enabled           bool               `json:"enabled"`
	Feature           int64              `json:"feature"`
	Environment       *int64             `json:"environment"`
	FeatureSegment    *int64             `json:"feature_segment,omitempty"`

	EnvironmentKey  string `json:"-"`
	Segment         *int64 `json:"-"`
	SegmentPriority *int64 `json:"-"`
}

func (fs *FeatureState) UnmarshalJSON(data []byte) error {
	var obj struct {
		ID                int64           `json:"id"`
		UUID              string          `json:"uuid"`
		FeatureStateValue json.RawMessage `json:"feature_state_value"`
		Enabled           bool            `json:"enabled"`
		Feature           int64           `json:"feature"`
		Environment       *int64          `json:"environment"`
		FeatureSegment    *int64          `json:"feature_segment"`
	}

	err := json.Unmarshal(data, &obj)

	if err != nil {
		return err
	}
	fs.ID = obj.ID
	fs.Enabled = obj.Enabled
	fs.Feature = obj.Feature
	fs.Environment = obj.Environment
	fs.FeatureSegment = obj.FeatureSegment
	fs.UUID = obj.UUID

	// If the feature state value is a struct(i.e: we are using `/features/featurestates/` endpoint) then unmarshal, set
	// and exit
	featureStateValue := FeatureStateValue{}
	err = json.Unmarshal(obj.FeatureStateValue, &featureStateValue)
	if err == nil {
		fs.FeatureStateValue = &featureStateValue
		return nil

	}
	// else(i.e: we are using `/environments/<env_key>/featurestates/` endpoint) convert the feature state value to struct, set and exit
	var fsValueRaw interface{}
	err = json.Unmarshal(obj.FeatureStateValue, &fsValueRaw)
	if err != nil {
		log.Println("flagsmithapi: Error unmarshalling FeatureStateValue: ", err)
	}

	switch fsv := fsValueRaw.(type) {
	case int64:
		fs.FeatureStateValue = &FeatureStateValue{
			Type:         "int",
			IntegerValue: &fsv,
		}
	case string:
		fs.FeatureStateValue = &FeatureStateValue{
			Type:        "unicode",
			StringValue: &fsv,
		}
	case bool:
		fs.FeatureStateValue = &FeatureStateValue{
			Type:         "bool",
			BooleanValue: &fsv,
		}
	case float64:
		int64Fsv := int64(fsv)
		fs.FeatureStateValue = &FeatureStateValue{
			Type:         "int",
			IntegerValue: &int64Fsv,
		}
	default:
		fs.FeatureStateValue = nil

	}
	return nil
}

type Condition struct {
	Operator string `json:"operator"`
	Property string `json:"property"`
	Value    string `json:"value"`
}

type Rule struct {
	Type       string      `json:"type"`
	Rules      []Rule      `json:"rules,omitempty"`
	Conditions []Condition `json:"conditions,omitempty"`
}

type Segment struct {
	ID          *int64  `json:"id,omitempty"`
	UUID        string  `json:"uuid,omitempty"`
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	ProjectID   *int64  `json:"project"`
	ProjectUUID string  `json:"-"`
	FeatureID   *int64  `json:"feature,omitempty"`
	Rules       []Rule  `json:"rules"`
}

type FeatureSegment struct {
	ID          *int64 `json:"id,omitempty"`
	Feature     int64  `json:"feature"`
	Segment     *int64 `json:"segment"`
	Environment int64  `json:"environment"`
	Priority    *int64 `json:"priority"`
}

type Environment struct {
	ID          int64  `json:"id,omitempty"`
	Name        string `json:"name"`
	APIKey      string `json:"api_key"`
	Description string `json:"description"`
	Project     int64  `json:"project"`
}

type Tag struct {
	ID          *int64  `json:"id,omitempty"`
	UUID        string  `json:"uuid,omitempty"`
	Name        string  `json:"label"`
	Description *string `json:"description"`
	Colour      string  `json:"color"`

	ProjectUUID string `json:"-"`
	ProjectID   *int64 `json:"project,omitempty"`
}

type Identity struct {
	ID         *int64 `json:"id,omitempty"`
	Identifier string `json:"identifier"`
}

type ServerSideEnvKey struct {
	ID        *int64    `json:"id,omitempty"`
	Active    bool      `json:"active,omitempty"`
	Name      string    `json:"name,omitempty"`
	Key       string    `json:"key,omitempty"`
	ExpiresAt time.Time `json:"expires_at,omitempty"`
}
