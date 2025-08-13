package env

import (
	"bufio"
	"os"
	"reflect"
	"strconv"
	"strings"
)

const envFile = ".env"

var values = make(map[string]string) // store everything as string

func init() {
	file, err := os.Open(envFile)
	if err == nil {
		defer file.Close()
		lr := bufio.NewScanner(file)

		for lr.Scan() {
			line := strings.TrimSpace(lr.Text())
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			kv := strings.SplitN(line, "=", 2)
			if len(kv) == 2 {
				values[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
			}
		}
	}
}

func Get[T string | int](key string, defaultValue T) T {
	// First check .env values
	if v, ok := values[key]; ok {
		return castValue(v, defaultValue)
	}

	// Then check real environment variables
	if envVal := os.Getenv(key); envVal != "" {
		return castValue(envVal, defaultValue)
	}

	// Fallback
	return defaultValue
}

func castValue[T string | int](raw string, defaultValue T) T {
	if reflect.TypeOf(defaultValue).Kind() == reflect.String {
		return any(raw).(T)
	}
	intVal, err := strconv.Atoi(raw)
	if err != nil {
		panic(err)
	}
	return any(intVal).(T)
}
