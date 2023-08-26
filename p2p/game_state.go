package p2p

import (
	"fmt"
	"sort"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

type AtomicInt struct {
	value int32
}

func NewAtomicInt(value int32) *AtomicInt {
	return &AtomicInt{
		value: value,
	}
}

func (a *AtomicInt) String() string {
	return fmt.Sprintf("%d", a.value)
}

func (a *AtomicInt) Set(value int32) {
	atomic.StoreInt32(&a.value, value)
}

func (a *AtomicInt) Get() int32 {
	return atomic.LoadInt32(&a.value)
}

func (a *AtomicInt) Inc() {
	currentValue := a.Get()
	a.Set(currentValue + 1)
}

type PlayerActionsRecv struct {
	mu          sync.RWMutex
	recvActions map[string]MessagePlayerAction
}

func NewPlayerActionRecv() *PlayerActionsRecv {
	return &PlayerActionsRecv{
		recvActions: make(map[string]MessagePlayerAction),
	}
}

func (pa *PlayerActionsRecv) addAction(from string, action MessagePlayerAction) {
	pa.mu.Lock()
	defer pa.mu.Unlock()

	pa.recvActions[from] = action
}

type PlayersReady struct {
	mu         sync.RWMutex
	recvStatus map[string]bool
}

func NewPlayersReady() *PlayersReady {
	return &PlayersReady{
		recvStatus: make(map[string]bool),
	}
}

func (pr *PlayersReady) addRecvStatus(from string) {
	pr.mu.Lock()
	defer pr.mu.Unlock()

	pr.recvStatus[from] = true
}

func (pr *PlayersReady) len() int {
	pr.mu.RLock()
	defer pr.mu.RUnlock()

	return len(pr.recvStatus)
}

func (pr *PlayersReady) clear() {
	pr.mu.Lock()
	defer pr.mu.Unlock()

	pr.recvStatus = make(map[string]bool)
}

type Game struct {
	listenAddr  string
	broadcastTo chan BroadcastTo

	playersReady *PlayersReady
	playersList  PlayersList

	// currentStatus and currentDealer shoud be automatically accessable
	currentStatus *AtomicInt
	// NOTE: this will be -1 when the game is in a bootstrapped state
	currentDealer *AtomicInt
	// currentPlayerTurn should be atomically accessable
	currentPlayerTurn *AtomicInt

	currentPlayerAction *AtomicInt

	recvPlayerActions *PlayerActionsRecv
}

func NewGame(addr string, bc chan BroadcastTo) *Game {
	g := &Game{
		listenAddr:          addr,
		broadcastTo:         bc,
		playersReady:        NewPlayersReady(),
		currentStatus:       NewAtomicInt(int32(GameStatusConnected)),
		currentDealer:       NewAtomicInt(0),
		recvPlayerActions:   NewPlayerActionRecv(),
		currentPlayerTurn:   NewAtomicInt(0),
		currentPlayerAction: NewAtomicInt(0),
	}

	g.playersList = append(g.playersList, addr)

	go g.loop()

	return g
}

func (g *Game) canTakeAction(from string) bool {
	currentPlayerAddr := g.playersList[g.currentPlayerTurn.Get()]
	return currentPlayerAddr == from
}

func (g *Game) handlePlayerAction(from string, action MessagePlayerAction) error {
	if !g.canTakeAction(from) {
		return fmt.Errorf("player (%s) taking action before hist turn", from)
	}

	g.recvPlayerActions.addAction(from, action)
	g.currentPlayerTurn.Inc()

	fmt.Printf("recieved player action: we %s, from %s, action %v\n", g.listenAddr, from, action)
	return nil
}

func (g *Game) TakeAction(action PlayerAction, value int) (err error) {
	if !g.canTakeAction(g.listenAddr) {
		err = fmt.Errorf("Cannot take action before ur turn %s", g.listenAddr)
		return
	}

	g.currentPlayerAction.Set(int32(action))

	g.currentPlayerTurn.Inc()

	g.sendToPlayers(MessagePlayerAction{
		Action:            action,
		CurrentGameStatus: GameStatus(g.currentStatus.Get()),
		Value:             value,
	}, g.getOtherPlayers()...)

	return
}

func (g *Game) ShuffleAndEncrypt(from string, deck [][]byte) error {

	prevPlayer := g.playersList[g.getPrevPositionOnTable()]

	if from != prevPlayer {
		return fmt.Errorf("recieved encrypted deck from the wrong player (%s) should be (%s)", from, prevPlayer)
	}

	_, isDealer := g.getCurrentDealerAddr()

	if isDealer && from == prevPlayer {
		fmt.Print("Dearler ", isDealer, g.listenAddr, from, prevPlayer)
		g.setStatus(GameStatusPreFlop)
		g.sendToPlayers(MessagePreFlop{}, g.getOtherPlayers()...)
		return nil
	}

	dealToPlayer := g.playersList[g.getNextPositionOnTable()]
	// dealToPlayer := g.getNextReadyPlayer(g.getNextPositionOnTable())
	//encyrption and shuffle

	g.setStatus(GameStatusDealing)
	g.sendToPlayers(MessageEncDeck{Deck: [][]byte{}}, dealToPlayer)

	fmt.Printf("received cards and going to shuffle: we %s, from %s, dealing %s\n", g.listenAddr, from, dealToPlayer)
	return nil
}

// InitiateShuffleAndDeal is only used for the "real" dealer. The actual "button player"
func (g *Game) InitiateShuffleAndDeal() {
	dealToPlayer := g.playersList[g.getNextPositionOnTable()]

	g.setStatus(GameStatusDealing)

	g.sendToPlayers(MessageEncDeck{Deck: [][]byte{}}, dealToPlayer)

	fmt.Printf("Dealing cards: we %s, to %s\n", g.listenAddr, dealToPlayer)
}

func (g *Game) SetPlayerReady(from string) {
	g.playersReady.addRecvStatus(from)
	fmt.Printf("setting player satus to ready we: %s, player: %s\n", g.listenAddr, from)

	// Check if the round can be started
	if g.playersReady.len() < 2 {
		return
	}

	// In this case we have engough players, hence, the round can be started
	// g.playersReady.clear()

	if _, ok := g.getCurrentDealerAddr(); ok {
		g.InitiateShuffleAndDeal()
	}
}

func (g *Game) SetReady() {
	g.playersReady.addRecvStatus(g.listenAddr)
	g.sendToPlayers(MessageReady{}, g.getOtherPlayers()...)

	g.setStatus(GameStatusReady)
}

func (g *Game) getCurrentDealerAddr() (string, bool) {
	currentDealer := g.playersList[g.currentDealer.Get()]

	return currentDealer, g.listenAddr == currentDealer
}

func (g *Game) SetStatus(s GameStatus) {
	g.setStatus(s)
}

func (g *Game) setStatus(s GameStatus) {
	if s == GameStatusPreFlop {
		g.currentPlayerTurn.Inc()
	}

	if GameStatus(g.currentStatus.Get()) != s {
		g.currentStatus.Set(int32(s))
		// atomic.StoreInt32((*int32)(&g.currentStatus), (int32)(s))
	}
}

func (g *Game) sendToPlayers(payload any, addr ...string) {
	g.broadcastTo <- BroadcastTo{
		To:      addr,
		Payload: payload,
	}

	fmt.Printf("sending to players payload: %v, players: %v, we %s\n", payload, addr, g.listenAddr)
}

func (g *Game) AddPlayer(from string) {
	g.playersList = append(g.playersList, from)
	sort.Sort(g.playersList)
}

func (g *Game) loop() {
	ticker := time.NewTicker(time.Second * 5)
	for {
		<-ticker.C

		currentDealer, _ := g.getCurrentDealerAddr()
		fmt.Printf("players: we: %s, gameStatus: %s, dealer %s, nextPlayerTurn %s, playerStatus %s\n", g.listenAddr, GameStatus(g.currentStatus.Get()), currentDealer, g.currentPlayerTurn, PlayerAction(g.currentPlayerAction.Get()))
		fmt.Printf("playersList: %s | playersAction %v\n", g.playersList, g.recvPlayerActions.recvActions)
	}
}

// index of our own position on the table
func (g *Game) getPositionOnTable() int {
	for index, player := range g.playersList {
		if g.listenAddr == player {
			return index
		}
	}

	return -1
}

func (g *Game) getPrevPositionOnTable() int {
	ourPosition := g.getPositionOnTable()

	// if we are in the first position of the table
	// we should return the last index of the player lists
	if ourPosition == 0 {
		return len(g.playersList) - 1
	}

	return ourPosition - 1
}

// getNextPosition return the index of the next player in teh PlayersList.
func (g *Game) getNextPositionOnTable() int {
	ourPosition := g.getPositionOnTable()
	if ourPosition == -1 {
		panic("getNextPositionOnTable errror")
	}

	if ourPosition == len(g.playersList)-1 {
		return 0
	}

	return ourPosition + 1

}

func (g *Game) getOtherPlayers() []string {
	players := []string{}

	for _, addr := range g.playersList {
		if addr == g.listenAddr {
			continue
		}
		players = append(players, addr)
	}
	return players
}

// func (g *Game) getNextReadyPlayer(nextPos int) string {
// 	if nextPos >= len(g.playersList) {
// 		return ""
// 	}
// 	nextPlayer := g.playersList[nextPos]

// 	if g.playersReady.haveRecv(nextPlayer) {
// 		return nextPlayer
// 	}

// 	return g.getNextReadyPlayer(nextPos + 1)
// }

type PlayersList []string

func (list PlayersList) Len() int { return len(list) }
func (list PlayersList) Swap(i, j int) {
	list[i], list[j] = list[j], list[i]
}
func (list PlayersList) Less(i, j int) bool {
	port1, _ := strconv.Atoi(list[i][1:])
	port2, _ := strconv.Atoi(list[j][1:])

	return port1 < port2
}
