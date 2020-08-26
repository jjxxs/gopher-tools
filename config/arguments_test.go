package config

import (
	"fmt"
	"os"
	"testing"
)

func TestGetArgumentOrDefault(t *testing.T) {
	args := map[string]string{
		"integer": "42",
		"string":  "Hallo Welt",
		"path":    "/dev/zero",
	}

	// set arguments, as if the process was actually called with these
	for k, v := range args {
		s := fmt.Sprintf("-%s=%s", k, v)
		os.Args = append(os.Args, s)
	}

	// for arguments that exist
	for k, v := range args {
		s := GetArgumentOrDefault(k, "")
		if s != v {
			t.Fail()
		}
	}

	// default should be returned for args that don't exist
	const defVal = "myDefault"
	s := GetArgumentOrDefault("thisShouldNotExist-Right?*", defVal)
	if s != defVal {
		t.Fail()
	}
}
