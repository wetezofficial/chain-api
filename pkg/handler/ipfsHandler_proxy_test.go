package handler

import (
	"fmt"
	"strings"
	"testing"
)

func TestIPFSHandler_bindApiKey(t *testing.T) {
	str := "/xxx/stats/bw"
	fmt.Println(strings.Split(str, "/")[1])
}
