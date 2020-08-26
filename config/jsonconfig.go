package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"sync"
)

// Represents a json-formatted file-based configuration
type jsonFileConfigProvider struct {
	config Config
	path   string
	file   *os.File
	mutex  *sync.Mutex
}

// Provides a configuration from a json-formatted file located
// at the specified path, returns error on failure
func NewJsonFileConfigProvider(path string) (Provider, error) {
	c := jsonFileConfigProvider{
		config: Config{},
		path:   path,
		file:   nil,
		mutex:  &sync.Mutex{},
	}

	if file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0600); err != nil {
		return nil, err
	} else {
		c.file = file
	}

	if err := c.loadConfig(); err != nil {
		return nil, err
	}

	return &c, nil
}

func (c *jsonFileConfigProvider) GetConfig() Config {
	return c.config
}

func (c *jsonFileConfigProvider) Transaction(mutator func(config Config) Config) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	mutator(c.config)

	if err := c.saveConfig(); err != nil {
		return err
	} else if err = c.loadConfig(); err != nil {
		return err
	}

	return nil
}

func (c *jsonFileConfigProvider) Exit() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	_ = c.saveConfig()
}

func (c *jsonFileConfigProvider) loadConfig() error {
	if _, err := c.file.Seek(0, 0); err != nil {
		return fmt.Errorf("failed to seek config %s, err=%s", c.path, err)
	} else if bytes, err := ioutil.ReadAll(c.file); err != nil {
		return fmt.Errorf("failed to read config %s, err=%s", c.path, err)
	} else if err := json.Unmarshal(bytes, &c.config); err != nil {
		return fmt.Errorf("failed to unmarshal config %s, err=%s", c.path, err)
	}

	return nil
}

func (c *jsonFileConfigProvider) saveConfig() error {
	if bytes, err := json.MarshalIndent(c.config, "", "\t"); err != nil {
		return fmt.Errorf("failed to marshal config, err=%s", err)
	} else if err = c.file.Truncate(int64(len(bytes))); err != nil {
		return fmt.Errorf("failed to truncate config %s, err=%s", c.path, err)
	} else if _, err = c.file.Seek(0, 0); err != nil {
		return fmt.Errorf("failed to seek config %s, err=%s", c.path, err)
	} else if _, err = c.file.Write(bytes); err != nil {
		return fmt.Errorf("failed to write config %s, err=%s", c.path, err)
	}

	return nil
}
