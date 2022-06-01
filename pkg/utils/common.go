package utils

import (
	"github.com/davecgh/go-spew/spew"
	"regexp"
	"unicode"
)

func SanitizeBackticks(s string) string {
	reg, err := regexp.Compile("`")
	if err != nil {
		panic(err.Error())
	}
	s = reg.ReplaceAllString(s, "'")
	return s
}

func Slice(val []interface{}, index int) interface{} {
	return val[index]
}

func Inpect(val interface{}) string {
	return spew.Sdump(val)
}

func LowerFirst(name string) string {
	for i, v := range name {
		return string(unicode.ToLower(v)) + name[i+1:]
	}
	return ""
}

func FirstOf(opts ...string) string {
	for _, opt := range opts {
		if opt != "" {
			return opt
		}
	}
	return ""
}
