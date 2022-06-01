package utils

import "strings"

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
