package config

// Config - Add your custom configuration type here.
type Config struct {
	This          int            `json:"this"`
	Is            string         `json:"is"`
	A             float64        `json:"a"`
	Test          []string       `json:"test"`
	Configuration map[string]int `json:"configuration"`
	Change        string         `json:"change"`
	Me            int            `json:"me"`
}

// Provider - Provides access to a configuration-object.
type Provider interface {
	// Retrieves the configuration
	GetConfig() Config
	// Performs an atomic mutation on the configuration
	Transaction(mutator func(config Config) Config) error
	// Closes the Config
	Exit()
}
