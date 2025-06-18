package utils

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/itchyny/gojq"
	"github.com/pkg/errors"
)

func CamelCase(name string) string {
	in := strings.Split(name, "_")
	if len(in) == 0 {
		return strings.Title(name)
	}
	out := make([]string, 0, len(in))
	for _, word := range in {
		out = append(out, strings.Title(word))
	}
	return strings.TrimSpace(strings.Join(out, ""))
}

func JqQueryFirst(input []byte, query *gojq.Query) (string, error) {
	var jsonData interface{}
	if err := json.Unmarshal([]byte(input), &jsonData); err != nil {
		return "", errors.Wrap(err, "failed to unmarshal JSON")
	}

	iter := query.Run(jsonData)
	for {
		v, ok := iter.Next()
		if !ok {
			break
		}
		if err, ok := v.(error); ok {
			if err, ok := err.(*gojq.HaltError); ok && err.Value() == nil {
				break
			}
			return "", err
		}
		if s, ok := v.(string); ok {
			return s, nil
		} else {
			return fmt.Sprintf("%v", v), nil
		}
	}
	return "", fmt.Errorf("no result found")
}
