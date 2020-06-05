package common

import (
	"log"
	"os"
	"path"
	"strings"
)

func GetParameterOrDefault(parameter string, defaultPath string) string {
	p := ""

	// see if user passed parameter
	for _, arg := range os.Args[1:] {
		if strings.Contains(arg, parameter) {
			p = strings.TrimLeft(arg, parameter)
			break
		}
	}

	// use default if no parameter was passed
	if len(p) == 0 {
		if dir, err := os.Getwd(); err != nil {
			log.Fatal(err)
		} else {
			p = path.Join(dir, defaultPath)
		}
	}

	return p
}
