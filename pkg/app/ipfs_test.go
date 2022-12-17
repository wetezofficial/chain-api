package app

import (
	"testing"
)

func TestClusterApiDemo(t *testing.T) {
	// clusterApiDemo()
	err := NewIPFSClient()
	if err != nil {
		t.Log(err)
	}
}

func TestIpfsApiDemo(t *testing.T) {
	ipfsApiDemo()
}
