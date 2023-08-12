package main

import (
	"fmt"
	"time"

	"github.com/habibbushira/ggpoker/p2p"
)

func main() {
	// rand.New(rand.NewSource(time.Now().UnixNano()))
	// for i := 0; i < 10; i++ {
	// 	d := deck.New()
	// 	fmt.Println(d)
	// 	fmt.Println()
	// }

	cfg := p2p.ServerConfig{
		ListenAddr: ":3000",
		Version:    "GGPOKER V0.1-alpha",
	}
	server := p2p.NewServer(cfg)
	go server.Start()
	time.Sleep(1 * time.Second)

	remoteCfg := p2p.ServerConfig{
		ListenAddr: ":4000",
		Version:    "GGPOKER V0.1-alpha",
	}
	remoteServer := p2p.NewServer(remoteCfg)
	go remoteServer.Start()
	if err := remoteServer.Connect(":3000"); err != nil {
		fmt.Println(err)
	}

	select {}

}
