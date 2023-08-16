package p2p

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

type Player struct {
	Status     GameStatus
	ListenAddr string
}

type GameState struct {
	listenAddr  string
	broadcastch chan BroadcastTo
	isDealer    bool       // atomic accessable
	gameStatus  GameStatus // atomic accessable

	playersWaiting int32
	playersLock    sync.RWMutex
	players        map[string]*Player
	playersList    []*Player

	decksRecievedLock sync.RWMutex
	decksRecieved     map[string]bool
}

func NewGameState(addr string, broadcastch chan BroadcastTo) *GameState {
	g := &GameState{
		listenAddr:    addr,
		broadcastch:   broadcastch,
		isDealer:      false,
		gameStatus:    GameStatusWaiting,
		players:       make(map[string]*Player),
		decksRecieved: make(map[string]bool),
	}

	go g.loop()
	return g
}

// TODO: (@habib) Check other read and write occurences of the GameStatus
func (g *GameState) SetStatus(s GameStatus) {
	if g.gameStatus != s {
		atomic.StoreInt32((*int32)(&g.gameStatus), (int32)(s))
	}
}

func (g *GameState) AddPlayerWaiting() {
	atomic.AddInt32(&g.playersWaiting, 1)
}

func (g *GameState) CheckNeedDealCards() {
	playersWaiting := atomic.LoadInt32(&g.playersWaiting)

	if playersWaiting == int32(len(g.players)) &&
		g.isDealer &&
		g.gameStatus == GameStatusWaiting {
		fmt.Printf("need to deal cards %s\n", g.listenAddr)

		g.InitiateShuffleAndDeal()
	}
}

func (g *GameState) GetPlayersWithStatus(s GameStatus) []string {
	players := []string{}
	for addr, player := range g.players {
		if player.Status == s {
			players = append(players, addr)
		}
	}

	return players

}

func (g *GameState) SetDeckRecievedToPlayer(from string) {
	g.decksRecievedLock.Lock()
	g.decksRecieved[from] = true
	g.decksRecievedLock.Unlock()
}

func (g *GameState) ShuffleAndEncrypt(from string, deck [][]byte) error {
	dealToPlayer := g.playersList[1]
	//encyrption and shuffle

	g.SendToPlayer(dealToPlayer.ListenAddr, MessageEncDeck{Deck: [][]byte{}})
	g.SetStatus(GameStatusShuffleAndDeal)

	return nil
}

// InitiateShuffleAndDeal is only used for the "real" dealer. The actual "button player"
func (g *GameState) InitiateShuffleAndDeal() {
	dealToPlayer := g.playersList[0]

	g.SendToPlayer(dealToPlayer.ListenAddr, MessageEncDeck{Deck: [][]byte{}})
	g.SetStatus(GameStatusShuffleAndDeal)
}

func (g *GameState) SendToPlayer(addr string, payload any) {
	g.broadcastch <- BroadcastTo{
		To:      []string{addr},
		Payload: payload,
	}

	fmt.Printf("sending payload: %v to player: %s\n", payload, addr)
}

func (g *GameState) SendToPlayersWithStatus(payload any, s GameStatus) {
	players := g.GetPlayersWithStatus(s)

	g.broadcastch <- BroadcastTo{
		To:      players,
		Payload: payload,
	}

	fmt.Printf("sendign to players payload: %v, players: %v", payload, players)
}

func (g *GameState) DealCards() {
	// g.broadcastch <- MessageEncDeck{Deck: deck.New()}
}

func (g *GameState) SetPlayerStatus(addr string, status GameStatus) {

	player, ok := g.players[addr]
	if !ok {
		panic("player could not be found")
	}

	player.Status = status

	g.CheckNeedDealCards()
}

func (g *GameState) lenPlayersConnectedWithLock() int {
	g.playersLock.RLock()
	defer g.playersLock.RUnlock()

	return len(g.players)
}

func (g *GameState) AddPlayer(addr string, status GameStatus) {
	g.playersLock.Lock()
	defer g.playersLock.Unlock()

	if status == GameStatusWaiting {
		g.AddPlayerWaiting()
	}

	player := &Player{
		ListenAddr: addr,
	}
	g.players[addr] = player
	g.playersList = append(g.playersList, player)

	// Seth the player status also when we add the player
	g.SetPlayerStatus(addr, status)

	fmt.Printf("new player joined: addr - %s, status - %s\n", addr, status)
}

func (g *GameState) loop() {
	ticker := time.NewTicker(time.Second * 5)
	for {
		select {
		case <-ticker.C:
			fmt.Printf("players connected: we: %s, %d, status: %s\n", g.listenAddr, g.lenPlayersConnectedWithLock(), g.gameStatus)
		}
	}
}
