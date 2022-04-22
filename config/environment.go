package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// GetEnvironmentOrDefault returns the environment-variable named by key
// or default if a variable with the key does not exist.
func GetEnvironmentOrDefault(key, def string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return def
}

// GetEnvironmentOrPanic returns the environment-variable named by key
// or panics if a variable with the key does not exist.
func GetEnvironmentOrPanic(key string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	panic(fmt.Errorf("environment variable for %s not set", key))
}

type lookup struct {
	key string
	val string
	ok  bool
}

func LookupEnv(key string) *lookup {
	l := &lookup{key: key}
	l.val, l.ok = os.LookupEnv(key)
	return l
}

func (l *lookup) String() *string {
	if !l.ok {
		return nil
	}
	return &l.val
}

func (l *lookup) Bool() *bool {
	if !l.ok {
		return nil
	} else if i, err := strconv.ParseBool(l.val); err != nil {
		return nil
	} else {
		return &i
	}
}

func (l *lookup) Int() *int {
	if !l.ok {
		return nil
	} else if i, err := strconv.Atoi(l.val); err != nil {
		return nil
	} else {
		return &i
	}
}

func (l *lookup) Int32() *int32 {
	if !l.ok {
		return nil
	} else if i, err := strconv.ParseInt(l.val, 10, 32); err != nil {
		return nil
	} else {
		j := int32(i)
		return &j
	}
}

func (l *lookup) Int64() *int64 {
	if !l.ok {
		return nil
	} else if i, err := strconv.ParseInt(l.val, 10, 64); err != nil {
		return nil
	} else {
		return &i
	}
}

func (l *lookup) Uint32() *uint32 {
	if !l.ok {
		return nil
	} else if i, err := strconv.ParseUint(l.val, 10, 32); err != nil {
		return nil
	} else {
		j := uint32(i)
		return &j
	}
}

func (l *lookup) Uint64() *uint64 {
	if !l.ok {
		return nil
	} else if i, err := strconv.ParseUint(l.val, 10, 64); err != nil {
		return nil
	} else {
		return &i
	}
}

func (l *lookup) Float32() *float32 {
	if !l.ok {
		return nil
	} else if i, err := strconv.ParseFloat(l.val, 32); err != nil {
		return nil
	} else {
		j := float32(i)
		return &j
	}
}

func (l *lookup) Float64() *float64 {
	if !l.ok {
		return nil
	} else if i, err := strconv.ParseFloat(l.val, 64); err != nil {
		return nil
	} else {
		return &i
	}
}

func (l *lookup) IntSlice() (is []int) {
	if !l.ok {
		return nil
	} else {
		for _, s := range strings.Split(l.val, ",") {
			if i, err := strconv.Atoi(s); err != nil {
				return nil
			} else {
				is = append(is, i)
			}
		}
	}
	return is
}
