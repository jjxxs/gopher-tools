package config

import (
	"encoding/json"
	"io/ioutil"
	"reflect"
	"testing"
)

var testCfg = Config{
	This: 42,
	Is:   "Test",
	A:    43.133,
	Test: []string{"1", "2", "3", "4"},
	Configuration: map[string]int{
		"1": 1,
		"2": 2,
		"3": 3,
	},
	Change: "Change",
	Me:     133,
}

func TestJsonConfig(t *testing.T) {
	cfgFile := createTemporaryConfig()

	// creating the provider should succeed
	var provider Provider
	provider, err := NewJsonFileConfigProvider(cfgFile)
	if err != nil {
		t.Fatal(err)
	}

	// getting the config from the provider should return an
	// object that is identical to the previously created testCfg
	actualCfg := provider.GetConfig()
	if !reflect.DeepEqual(testCfg, actualCfg) {
		t.Fatal("configs are not equal")
	}

	// make a change
	testCfg.This = 42
	testCfg.Test[2] = "this is a test"
	err = provider.Transaction(func(config Config) Config {
		config.This = 42
		config.Test[2] = "this is a test"
		return config
	})
	if err != nil {
		t.Fatal(err)
	}

	// config should've been changed
	actualCfg = provider.GetConfig()
	if !reflect.DeepEqual(testCfg, actualCfg) {
		t.Fatal("configs are not equal")
	}

	// exit provider & reload
	provider.Exit()
	provider, err = NewJsonFileConfigProvider(cfgFile)
	if err != nil {
		t.Fatal(err)
	}

	// get cfg again, should be equal
	actualCfg = provider.GetConfig()
	if !reflect.DeepEqual(testCfg, actualCfg) {
		t.Fatal("configs are not equal")
	}
}

// creates a temporary file and writes a json-formatted test-config
// to the file. closes the file and returns the path to the file.
func createTemporaryConfig() string {
	file, _ := ioutil.TempFile("", "gopherTools_jsonConfigTest")
	bytes, _ := json.Marshal(&testCfg)
	_, _ = file.WriteAt(bytes, 0)
	_ = file.Close()
	return file.Name()
}
