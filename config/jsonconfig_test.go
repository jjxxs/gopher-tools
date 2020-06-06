package config

import (
	"encoding/json"
	"io/ioutil"
	"reflect"
	"testing"
)

type conf struct {
	Prop1 int     `json:"int"`
	Prop2 string  `json:"string"`
	Prop3 float64 `json:"float64"`
	Prop4 []int   `json:"array"`
	Prop5 conf2   `json:"struct"`
}

type conf2 struct {
	Prop1 int     `json:"int"`
	Prop2 string  `json:"string"`
	Prop3 float64 `json:"float64"`
}

func TestJsonConfig(t *testing.T) {
	// creates a temporary file and writes a json-formatted test-config
	// to the file. closes the file and gets the path to instantiate
	// the subject under test (JsonFileConfig) with
	file, _ := ioutil.TempFile("", "gopherTools_jsonConfigTest")
	testCfg := conf{
		Prop1: 1,
		Prop2: "2",
		Prop3: 3.3,
		Prop4: []int{1, 2, 3, 4, 5},
		Prop5: conf2{
			Prop1: 1,
			Prop2: "2",
			Prop3: 3.3,
		},
	}
	bytes, _ := json.Marshal(&testCfg)
	_, _ = file.WriteAt(bytes, 0)
	_ = file.Close()
	p := file.Name()

	// opening the config should succeed
	var config Config
	config, err := NewJsonFileConfig(p)
	if err != nil {
		t.Fatal(err)
	}

	// getting the config should return an object that is identical
	// to the previously created testCfg
	var actualCfg conf
	cfg := config.GetConfig()
	if ok := convertInterfaceToConf(cfg, &actualCfg); !ok {
		t.Fatal("config could not be converted")
	}

	// check for equality
	if !reflect.DeepEqual(testCfg, actualCfg) {
		t.Fatal("configs are not equal")
	}

	// make change
	testCfg.Prop1 = 42
	testCfg.Prop5.Prop2 = "this is a test"
	err = config.Transaction(func(config interface{}) {
		is, _ := config.(map[string]interface{})
		is["int"] = 42
		str := is["struct"].(map[string]interface{})
		str["string"] = "this is a test"
	})
	if err != nil {
		t.Fatal(err)
	}

	// config should have changed
	cfg = config.GetConfig()
	if ok := convertInterfaceToConf(cfg, &actualCfg); !ok {
		t.Fatal("config could not be converted")
	}
	if !reflect.DeepEqual(testCfg, actualCfg) {
		t.Fatal("configs are not equal")
	}

	// exit config & reload
	config.Exit()
	config, err = NewJsonFileConfig(p)
	if err != nil {
		t.Fatal(err)
	}

	// get cfg again, should be equal
	cfg = config.GetConfig()
	if ok := convertInterfaceToConf(cfg, &actualCfg); !ok {
		t.Fatal("config could not be converted")
	}
	if !reflect.DeepEqual(testCfg, actualCfg) {
		t.Fatal("configs are not equal")
	}
}

func convertInterfaceToConf(in interface{}, out *conf) bool {
	is, ok := in.(map[string]interface{})
	if !ok {
		return false
	}

	out.Prop1 = int(is["int"].(float64))
	out.Prop2 = is["string"].(string)
	out.Prop3 = is["float64"].(float64)
	out.Prop4 = make([]int, 0)
	arr := is["array"].([]interface{})
	out.Prop4 = append(out.Prop4, int(arr[0].(float64)))
	out.Prop4 = append(out.Prop4, int(arr[1].(float64)))
	out.Prop4 = append(out.Prop4, int(arr[2].(float64)))
	out.Prop4 = append(out.Prop4, int(arr[3].(float64)))
	out.Prop4 = append(out.Prop4, int(arr[4].(float64)))
	str := is["struct"].(map[string]interface{})
	out.Prop5 = conf2{
		Prop1: int(str["int"].(float64)),
		Prop2: str["string"].(string),
		Prop3: str["float64"].(float64),
	}

	return true
}
