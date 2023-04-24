package homeassistant

type AutoConfiguration struct {
	StateTopic          string `json:"state_topic"`
	Name                string `json:"name"`
	AvailabilityTopic   string `json:"availability_topic"`
	JSONAttributesTopic string `json:"json_attributes_topic"`
	Icon                string `json:"icon,omitempty"`
	SourceType          string `json:"source_type"`
}
