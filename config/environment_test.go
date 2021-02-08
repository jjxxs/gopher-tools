package config

import (
	"os"
	"testing"
)

func TestGetEnvironmentOrDefault(t *testing.T) {
	if err := os.Setenv("testEnvKey", "testEnvValue"); err != nil {
		t.Fail()
	}
	if val := GetEnvironmentOrDefault("testEnvKey", "testEnvDefault"); val != "testEnvValue" {
		t.Fail()
	}
	if val := GetEnvironmentOrDefault("testEnvKeyNoExist", "testEnvDefault"); val != "testEnvDefault" {
		t.Fail()
	}
}

func TestGetEnvironmentOrPanic(t *testing.T) {
	if err := os.Setenv("testEnvKey", "testEnvValue"); err != nil {
		t.Fail()
	}
	if val := GetEnvironmentOrPanic("testEnvKey"); val != "testEnvValue" {
		t.Fail()
	}
	defer func() {
		if r := recover(); r == nil { // if didn't panic, fail the test
			t.Fail()
		}
	}()
	_ = GetEnvironmentOrPanic("testEnvKeyNoExist") // this should panic
}
