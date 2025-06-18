package utils

import (
	"strconv"
	"strings"
)

func ToUint64(s string) (uint64, error) {
	if strings.HasPrefix(s, "0x") {
		return strconv.ParseUint(s[2:], 16, 64)
	}
	return strconv.ParseUint(s, 10, 64)
}
