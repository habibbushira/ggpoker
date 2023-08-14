package p2p

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/habibbushira/ggpoker/deck"
)

type GameStatus int32

func (g GameStatus) String() string {
	switch g {
	case GameStatusWaiting:
		return "WAITING"
	case GameStatusReceiving:
		return "RECEIVING"
	case GameStatusDealing:
		return "DEALING"
	case GameStatusPreFlop:
		return "PRE FLOP"
	case GameStatusFlop:
		return "FLOP"
	case GameStatusTurn:
		return "TURN"
	case GameStatusRiver:
		return "RIVER"
	default:
		return "unknown"
	}
}

const (
	GameStatusWaiting GameStatus = iota
	GameStatusReceiving
	GameStatusDealing
	GameStatusPreFlop
	GameStatusFlop
	GameStatusTurn
	GameStatusRiver
)

type Player struct {
	Status GameStatus
}

type GameState struct {
	listenAddr  string
	broadcastch chan any
	isDealer    bool       // atomic accessable
	gameStatus  GameStatus // atomic accessable

	playersWaiting int32
	playersLock    sync.RWMutex
	players        map[string]*Player
}

func NewGameState(addr string, broadcastch chan any) *GameState {
	g := &GameState{
		listenAddr:  addr,
		broadcastch: broadcastch,
		isDealer:    false,
		gameStatus:  GameStatusWaiting,
		players:     make(map[string]*Player),
	}

	go g.loop()
	return g
}

// TODO: (@habib) Check other read and write occurences of the GameStatus
func (g *GameState) SetStatus(s GameStatus) {
	oldStatus := atomic.LoadInt32((*int32)(&g.gameStatus))
	atomic.StoreInt32((*int32)(&oldStatus), (int32)(s))
	g.gameStatus = s
}

func (g *GameState) AddPlayerWaiting() {
	atomic.AddInt32(&g.playersWaiting, 1)
}

func (g *GameState) CheckNeedDealCards() {
	playersWaiting := atomic.LoadInt32(&g.playersWaiting)

	if playersWaiting == int32(len(g.players)) &&
		g.isDealer &&
		g.gameStatus == GameStatusWaiting {
		fmt.Printf("need to deal cards %s", g.listenAddr)

		g.DealCards()
	}
}

func (g *GameState) DealCards() {
	g.broadcastch <- MessageCards{Deck: deck.New()}
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

	g.players[addr] = &Player{
		Status: status,
	}

	// Seth the player status also when we add the player
	g.SetPlayerStatus(addr, status)

	fmt.Printf("new player joined: addr - %s, status - %s\n", addr, status)
}

func (g *GameState) loop() {
	ticker := time.NewTicker(time.Second * 5)
	for {
		select {
		case <-ticker.C:
			fmt.Printf("players connected: %d, status: %s\n", g.lenPlayersConnectedWithLock(), g.gameStatus)
		}
	}
}
