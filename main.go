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

	time.Sleep(1 * time.Second)

	return server
}

func main() {
	playerA := makeServer(":3000")
	playerB := makeServer(":4000")
	playerC := makeServer(":5000")
	playerD := makeServer(":6000")
	playerE := makeServer(":7000")
	playerF := makeServer(":8000")

	fmt.Println()

	playerB.Connect(playerA.ListenAddr)
	playerC.Connect(playerB.ListenAddr)
	playerD.Connect(playerC.ListenAddr)
	playerE.Connect(playerD.ListenAddr)
	playerF.Connect(playerE.ListenAddr)

	select {}

}
