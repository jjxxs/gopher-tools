package config

/*
 * Config
 */

type Config interface {
	GetConfig() interface{}
	Transaction(mutator func(config interface{})) error
	Exit()
}