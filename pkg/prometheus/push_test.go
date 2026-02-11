package prometheus

import (
	"testing"
)

func TestPush(t *testing.T) {
	network := "test_network"
	metrics := []ErrorNumMetric{
		{"wetez", 3},
		{"official", 2},
	}
	PushMetrics(network, metrics)
}
