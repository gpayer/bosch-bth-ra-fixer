package rewriter

type Availability struct {
	Topic         string `json:"topic"`
	ValueTemplate string `json:"value_template"`
}

type Device struct {
	Identifiers  []string `json:"identifiers"`
	Manufacturer string   `json:"manufacturer"`
	Model        string   `json:"model"`
	Name         string   `json:"name"`
	SwVersion    string   `json:"sw_version"`
	ViaDevice    string   `json:"via_device"`
}

type Origin struct {
	Name string `json:"name"`
	Sw   string `json:"sw"`
	Url  string `json:"url"`
}

type Config struct {
	ActionTemplate             string         `json:"action_template"`
	ActionTopic                string         `json:"action_topic"`
	Availability               []Availability `json:"availability"`
	CurrentTemperatureTemplate string         `json:"current_temperature_template"`
	CurrentTemperatureTopic    string         `json:"current_temperature_topic"`
	Device                     Device         `json:"device"`
	MaxTemp                    string         `json:"max_temp"`
	MinTemp                    string         `json:"min_temp"`
	ModeCommandTopic           string         `json:"mode_command_topic"`
	ModeStateTemplate          string         `json:"mode_state_template"`
	ModeStateTopic             string         `json:"mode_state_topic"`
	Modes                      []string       `json:"modes"`
	Name                       *string        `json:"name"`
	ObjectID                   string         `json:"object_id"`
	Origin                     Origin         `json:"origin"`
	TempStep                   float64        `json:"temp_step"`
	TemperatureCommandTopic    string         `json:"temperature_command_topic"`
	TemperatureStateTemplate   string         `json:"temperature_state_template"`
	TemperatureStateTopic      string         `json:"temperature_state_topic"`
	TemperatureUnit            string         `json:"temperature_unit"`
	UniqueID                   string         `json:"unique_id"`
}
