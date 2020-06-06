package config

// Config provides means to access a configuration-object
type Config interface {
	// Retrieves the configuration
	GetConfig() interface{}
	// Performs an atomic mutation on the configuration
	Transaction(mutator func(config interface{})) error
	// Closes the Config
	Exit()
}
