package flagsmithapi

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
	Identity          *int64             `json:"identity"`
	Feature_segment   *int64             `json:"feature_segment"`
}
