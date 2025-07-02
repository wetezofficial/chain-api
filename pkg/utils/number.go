package utils

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

func ToUint64(s string) (uint64, error) {
	if strings.HasPrefix(s, "0x") {
		return strconv.ParseUint(s[2:], 16, 64)
	} else if strings.Contains(s, ".") {
		if v, err := strconv.ParseFloat(s, 64); err == nil {
			// Convert float64 to int64, checking for overflow
			if v > float64(math.MaxUint64) || v < 0 {
				return 0, fmt.Errorf("value %v overflows uint64", v)
			}
			return uint64(v), nil
		}
	}

	if number, err := strconv.ParseUint(s, 10, 64); err == nil {
		return number, nil
	}

	return strconv.ParseUint(s, 10, 64)
}
