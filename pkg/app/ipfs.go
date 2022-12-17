package app

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/ipfs-cluster/ipfs-cluster/api"
	"github.com/ipfs-cluster/ipfs-cluster/api/rest/client"
	shell "github.com/ipfs/go-ipfs-api"
	ma "github.com/multiformats/go-multiaddr"
)

var filePath = ""

func NewIPFSClient() error {
	ctx := context.Background()
	addr, err := ma.NewMultiaddr(fmt.Sprintf("/ip4/127.0.0.1/tcp/%d", 9094))
	if err != nil {
		return err
	}

	cfg := &client.Config{
		APIAddr:           addr,
		DisableKeepAlives: true,
	}
	c, err := client.NewDefaultClient(cfg)
	if err != nil {
		return err
	}
	out := make(chan api.AddedOutput)
	go func() {
		err = c.Add(ctx, []string{filePath}, api.AddParams{}, out)
		if err != nil {
			return
		}
	}()
	result := <-out
	fmt.Println(result.Cid)

	return nil
}

func clusterApiDemo() {
	// apiMAddr, err := ma.NewMultiaddr("/ip4/127.0.0.1/tcp/0")
	// if err!=nil{
	// 	log.Fatalln(err)
	// }
	cfg := client.Config{
		Host: "127.0.0.1",
	}
	ipfsClient, err := client.NewDefaultClient(&cfg)
	if err != nil {
		log.Fatalln(err)
	}
	out := make(chan api.AddedOutput)
	go func() {
		err = ipfsClient.Add(context.Background(), []string{filePath}, api.AddParams{}, out)
		if err != nil {
			log.Fatalln(err)
		}
	}()
	result := <-out
	fmt.Println(result.Cid)
}

func ipfsApiDemo() {
	var err error
	sh := shell.NewShell("localhost:9095")
	cid, err := sh.Add(strings.NewReader("hello world!"))
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "error: %s", err)
		os.Exit(1)
	}
	fmt.Printf("added %s", cid)
	err = sh.Pin(cid)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "error: %s", err)
		os.Exit(1)
	}
}
