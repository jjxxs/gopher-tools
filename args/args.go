package args

import (
	"os"
	"strings"
)

func GetArgumentOrDefault(arg string, def string) string {
	r := ""

	// see if argument exists
	for _, arg := range os.Args[1:] {
		if strings.Contains(arg, arg) {
			r = strings.TrimLeft(arg, arg)
			break
		}
	}

	if len(r) == 0 {
		r = def
	}

	return r
}
