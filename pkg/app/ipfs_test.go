package app

import (
	"starnet/chain-api/pkg/initapp"
	"testing"
)

func TestClusterApiDemo(t *testing.T) {
	// clusterApiDemo()
	err := initapp.NewIPFSClient()
	if err != nil {
		t.Log(err)
	}
}
