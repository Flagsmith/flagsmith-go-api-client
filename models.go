package flagsmithapi

import (
	"encoding/json"
	"log"
)

type Project struct {
	ID           int64  `json:"id"`
	UUID         string `json:"uuid"`
	Name         string `json:"name"`
	Organisation int64  `json:"organisation"`
}

type FeatureStateValue struct {
	Type         string  `json:"type"`
	StringValue  *string `json:"string_value"`
	IntegerValue *int64  `json:"integer_value"`
	BooleanValue *bool   `json:"boolean_value"`
}

type FeatureState struct {
	ID                int64              `json:"id"`
	FeatureStateValue *FeatureStateValue `json:"feature_state_value"`
	Enabled           bool               `json:"enabled"`
	Feature           int64              `json:"feature"`
	Environment       int64              `json:"environment"`
}

func (fs *FeatureState) UnmarshalJSON(data []byte) error {
	var obj struct {
		ID                int64           `json:"id"`
		FeatureStateValue json.RawMessage `json:"feature_state_value"`
		Enabled           bool            `json:"enabled"`
		Feature           int64           `json:"feature"`
		Environment       int64           `json:"environment"`
	}

	err := json.Unmarshal(data, &obj)

	if err != nil {
		return err
	}
	fs.ID = obj.ID
	fs.Enabled = obj.Enabled
	fs.Feature = obj.Feature
	fs.Environment = obj.Environment

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
