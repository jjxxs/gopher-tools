package config

import (
	"os"
	"strings"
)

var argSeparator = "="

// GetArgumentOrDefault - Tests if a given argument was passed to the process. If so,
// returns the arguments value. Otherwise falls back to a given default.
func GetArgumentOrDefault(argName string, def string) string {
	for _, arg := range os.Args[1:] {
		if len(arg) <= 1 {
			continue
		}

		// remove first character, usually a '-' for program-arguments
		arg = arg[1:]

		// split after the separator
		splits := strings.SplitN(arg, argSeparator, 2)
		if len(splits) != 2 {
			continue
		}

		if splits[0] == argName {
			return splits[1]
		}
	}

	return def
}
