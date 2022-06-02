package flagsmithapi

type FeatureStateValue struct {
	Type string `json:"type"`
	StringValue string `json:"string_value"`
	IntegerValue int `json:"integer_value"`
	BooleanValue bool `json:"boolean_value"`

}
type FeatureState struct {
	ID int `json:"id"`
	FeatureStateValue *FeatureStateValue `json:"feature_state_value"`
	Enabled bool `json:"enabled"`
	Feature int `json:"feature"`
	Environment int `json:"environment"`
	Identity int `json:"identity"`
	Feature_segment int `json:"feature_segment"`
}
