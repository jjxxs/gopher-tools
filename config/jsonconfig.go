package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"sync"
)

/*
 * JSON Config
 */

type jsonConfigImpl struct {
	config interface{}
	path   string
	mutex  *sync.Mutex
}

func NewJsonFileConfig(path string) (Config, error) {
	c := jsonConfigImpl{
		config: nil,
		path:   path,
		mutex:  &sync.Mutex{},
	}

	// load the config
	if err := c.loadConfig(); err != nil {
		return nil, err
	}

	return &c, nil
}

func (c *jsonConfigImpl) GetConfig() interface{} {
	return c.config
}

func (c *jsonConfigImpl) Transaction(mutator func(config interface{})) error {
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

func (c *jsonConfigImpl) Exit() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	_ = c.saveConfig()
}

func (c *jsonConfigImpl) loadConfig() error {
	if bytes, err := ioutil.ReadFile(c.path); err != nil {
		return errors.New(fmt.Sprintf("failed to read config %s, err=%s", c.path, err))
	} else if err := json.Unmarshal(bytes, &c.config); err != nil {
		return errors.New(fmt.Sprintf("failed to unmarshal config, err=%s", err))
	}

	return nil
}

func (c *jsonConfigImpl) saveConfig() error {
	if bytes, err := json.MarshalIndent(c.config, "", "\t"); err != nil {
		return errors.New(fmt.Sprintf("failed to marshal config, err=%s", err))
	} else if err := ioutil.WriteFile(c.path, bytes, 0660); err != nil {
		return errors.New(fmt.Sprintf("failed to write config %s, err=%s", c.path, err))
	}

	return nil
}
