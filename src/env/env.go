package env

import (
	"bufio"
	"os"
	"reflect"
	"strconv"
	"strings"
)

const envFile = ".env"

var values = make(map[string]any)

func init() {
	file, err := os.Open(envFile)
	if err != nil {
		return
	}
	defer file.Close()

	lr := bufio.NewScanner(file)

	for lr.Scan() {
		line := lr.Text()
		kv := strings.Split(line, "=")

		if len(kv) < 2 {
			continue
		}

		if len(kv) == 2 {
			values[kv[0]] = kv[1]
			continue
		}

		values[kv[0]] = strings.TrimSpace(strings.Join(kv[1:], "="))
	}
}

func Get[T string | int](key string, defaultValue T) T {
	var value T = defaultValue

	if v, ok := values[key]; ok {
		value = v.(T)
	}

	if reflect.TypeOf(value).Kind() == reflect.String {
		return value
	}

	v, err := strconv.Atoi((string)(value))

	if err != nil {
		panic(err)
	}

	return (T)(v)
}
