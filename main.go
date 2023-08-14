package main

import (
	"fmt"
	"time"

	"github.com/habibbushira/ggpoker/p2p"
)

func makeServer(addr string) *p2p.Server {
	cfg := p2p.ServerConfig{
		ListenAddr:  addr,
		Version:     "GGPOKER V0.1-alpha",
		GameVariant: p2p.TexasHoldem,
	}
	server := p2p.NewServer(cfg)
	go server.Start()

	time.Sleep(200 * time.Millisecond)

	return server
}

func main() {
	playerA := makeServer(":3000")
	playerB := makeServer(":4000")
	playerC := makeServer(":5000")
	// playerD := makeServer(":6000")
	// playerE := makeServer(":7000")
	// playerF := makeServer(":8000")

	fmt.Println()

	playerB.Connect(playerA.ListenAddr)
	time.Sleep(200 * time.Millisecond)

	playerC.Connect(playerB.ListenAddr)
	time.Sleep(200 * time.Millisecond)

	// playerD.Connect(playerC.ListenAddr)
	// time.Sleep(200 * time.Millisecond)

	// playerE.Connect(playerD.ListenAddr)
	// time.Sleep(200 * time.Millisecond)

	// playerF.Connect(playerE.ListenAddr)
	// time.Sleep(200 * time.Millisecond)

	select {}

}
