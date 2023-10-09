package p2p

import (
	"fmt"
	"sort"
	"time"

	"github.com/habibbushira/ggpoker/proto"
)

type GameState struct {
	listenAddr  string
	broadcastTo chan BroadcastTo

	playersList *PlayersList

	// currentStatus and currentDealer shoud be automatically accessable
	currentStatus *AtomicInt
	// NOTE: this will be -1 when the game is in a bootstrapped state
	currentDealer *AtomicInt
	// currentPlayerTurn should be atomically accessable
	currentPlayerTurn *AtomicInt

	currentPlayerAction *AtomicInt

	table *Table
}

func NewGameState(addr string, bc chan BroadcastTo) *GameState {
	g := &GameState{
		listenAddr:          addr,
		broadcastTo:         bc,
		currentStatus:       NewAtomicInt(int32(GameStatusConnected)),
		currentDealer:       NewAtomicInt(0),
		currentPlayerTurn:   NewAtomicInt(0),
		currentPlayerAction: NewAtomicInt(0),

		playersList: NewPlayersList(),
		table:       NewTable(6),
	}

	g.playersList.add(addr)

	go g.loop()

	return g
}

func (g *GameState) canTakeAction(from string) bool {
	// currentPlayerAddr := g.playersList[g.currentPlayerTurn.Get()]
	currentPlayerAddr := g.playersList.get(int(g.currentPlayerTurn.Get()))
	return currentPlayerAddr == from
}

func (g *GameState) isFromCurrentDealer(from string) bool {
	return g.playersList.get(int(g.currentDealer.Get())) == from
}

func (g *GameState) handlePlayerAction(from string, action MessagePlayerAction) error {
	if !g.canTakeAction(from) {
		return fmt.Errorf("player (%s) taking action before hist turn", from)
	}

	if action.CurrentGameStatus != GameStatus(g.currentStatus.Get()) && !g.isFromCurrentDealer(from) {
		return fmt.Errorf("player (%s) has not the correct game status (%s)", from, action.CurrentGameStatus)
	}

	if g.playersList.get(int(g.currentDealer.Get())) == from {
		g.advanceToNextRound()
	}

	// g.recvPlayerActions.addAction(from, action)
	g.incNextPlayer()

	fmt.Printf("recieved player action: we %s, from %s, action %v\n", g.listenAddr, from, action)
	return nil
}

func (g *GameState) TakeAction(action PlayerAction, value int) (err error) {
	if !g.canTakeAction(g.listenAddr) {
		err = fmt.Errorf("Cannot take action before ur turn %s", g.listenAddr)
		return
	}

	g.currentPlayerAction.Set(int32(action))

	g.incNextPlayer()

	// _, isDealer := g.getCurrentDealerAddr(); isDealer
	if g.listenAddr == g.playersList.get(int(g.currentDealer.Get())) {
		g.advanceToNextRound()
	}

	g.sendToPlayers(MessagePlayerAction{
		Action:            action,
		CurrentGameStatus: GameStatus(g.currentStatus.Get()),
		Value:             value,
	}, g.getOtherPlayers()...)

	return
}

func (g *GameState) ShuffleAndEncrypt(from string, deck [][]byte) error {

	// prevPlayer := g.playersList.get(int(g.getPrevPositionOnTable()))
	prevPlayer, err := g.table.GetPlayerBefore(g.listenAddr)
	if err != nil {
		panic(err)
	}

	if from != prevPlayer.addr {
		return fmt.Errorf("recieved encrypted deck from the wrong player (%s) should be (%s)", from, prevPlayer.addr)
	}

	_, isDealer := g.getCurrentDealerAddr()

	if isDealer && from == prevPlayer.addr {
		g.setStatus(GameStatusPreFlop)
		g.table.SetPlayerStatus(g.listenAddr, GameStatusPreFlop)
		g.sendToPlayers(MessagePreFlop{}, g.getOtherPlayers()...)
		return nil
	}

	// dealToPlayer := g.playersList.get(int(g.getNextPositionOnTable()))
	dealToPlayer, err := g.table.GetPlayerAfter(g.listenAddr)
	if err != nil {
		panic(err)
	}
	//encyrption and shuffle

	g.setStatus(GameStatusDealing)
	g.sendToPlayers(MessageEncDeck{Deck: [][]byte{}}, dealToPlayer.addr)

	fmt.Printf("received cards and going to shuffle: we %s, from %s, dealing %s\n", g.listenAddr, from, dealToPlayer.addr)
	return nil
}

// InitiateShuffleAndDeal is only used for the "real" dealer. The actual "button player"
func (g *GameState) InitiateShuffleAndDeal() {
	// dealToPlayer := g.playersList.get(int(g.getNextPositionOnTable()))
	dealToPlayer, err := g.table.GetPlayerAfter(g.listenAddr)
	if err != nil {
		panic(err)
	}

	g.setStatus(GameStatusDealing)

	g.sendToPlayers(MessageEncDeck{Deck: [][]byte{}}, dealToPlayer.addr)

	fmt.Printf("Dealing cards: we %s, to %s\n", g.listenAddr, dealToPlayer.addr)
}

func (g *GameState) mayBeDeal() {
	return
	if GameStatus(g.currentStatus.Get()) == GameStatusReady {
		g.InitiateShuffleAndDeal()
	}
}

// This is called when we receive a ready message from
// a player in the network
func (g *GameState) SetPlayerAtTable(addr string) {
	tablePos := g.playersList.getIndex(addr)
	g.table.AddPlayerOnPosition(addr, tablePos)

	//TODO: check if we really need this

	// fmt.Printf("setting player satus to ready we: %s, player: %s\n", g.listenAddr, from)

	// Check if the round can be started
	if g.table.LenPlayers() < 2 {
		return
	}

	// In this case we have engough players, hence, the round can be started
	//

	if _, areWeDealer := g.getCurrentDealerAddr(); areWeDealer {
		go func() {
			time.Sleep(time.Second * 5)
			g.mayBeDeal()
		}()
	}
}

// This is called when we set ourselfs as ready.
func (g *GameState) TakeSeatAtTable() {
	tablePos := g.playersList.getIndex(g.listenAddr)
	g.table.AddPlayerOnPosition(g.listenAddr, tablePos)

	g.sendToPlayers(&proto.TakeSeat{
		Addr: g.listenAddr,
	}, g.getOtherPlayers()...)

	g.setStatus(GameStatusReady)
}

func (g *GameState) getCurrentDealerAddr() (string, bool) {
	currentDealer := g.playersList.get(int(g.currentDealer.Get()))

	return currentDealer, g.listenAddr == currentDealer
}

func (g *GameState) SetStatus(s GameStatus) {
	g.setStatus(s)
	g.table.SetPlayerStatus(g.listenAddr, s)
}

func (g *GameState) setStatus(s GameStatus) {
	if s == GameStatusPreFlop {
		g.incNextPlayer()
	}

	if GameStatus(g.currentStatus.Get()) != s {
		g.currentStatus.Set(int32(s))
	}
}

func (g *GameState) sendToPlayers(payload any, addr ...string) {
	g.broadcastTo <- BroadcastTo{
		To:      addr,
		Payload: payload,
	}

	// fmt.Printf("sending to players payload: %v, players: %v, we %s\n", payload, addr, g.listenAddr)
}

func (g *GameState) AddPlayer(from string) {
	g.playersList.add(from)
	sort.Sort(g.playersList)
}

func (g *GameState) loop() {
	ticker := time.NewTicker(time.Second * 5)
	for {
		<-ticker.C

		currentDealer, _ := g.getCurrentDealerAddr()
		fmt.Printf("players: we: %s, gameState %s, dealer %s, nextPlayerTurn %s\n", g.listenAddr, GameStatus(g.currentStatus.Get()), currentDealer, g.currentPlayerTurn)
		fmt.Printf("playersList: %s \n", g.playersList.List())
		fmt.Printf("we: %s, table: %s \n", g.listenAddr, g.table)
		// fmt.Printf("table: %s\n", g.table)
	}
}

func (g *GameState) getNextGameStatus() GameStatus {
	switch GameStatus(g.currentStatus.Get()) {
	case GameStatusPreFlop:
		return GameStatusFlop
	case GameStatusFlop:
		return GameStatusTurn
	case GameStatusTurn:
		return GameStatusRiver
	case GameStatusRiver:
		return GameStatusReady
	default:
		panic("invalid game status")
	}
}

func (g *GameState) advanceToNextRound() {
	// g.recvPlayerActions.clear()
	g.currentPlayerAction.Set(int32(PlayerActionNone))

	if GameStatus(g.currentStatus.Get()) == GameStatusRiver {
		g.TakeSeatAtTable()
	} else {
		g.currentStatus.Set(int32(g.getNextGameStatus()))
	}
}

func (g *GameState) incNextPlayer() {
	player, err := g.table.GetPlayerAfter(g.listenAddr)
	if err != nil {
		panic(err)
	}

	if g.playersList.Len()-1 == int(g.currentPlayerTurn.Get()) {
		g.currentPlayerTurn.Set(0)
	} else {
		g.currentPlayerTurn.Inc()
	}

	fmt.Println("next player on the table: ", player.tablePos)
	fmt.Println("old wrong value => : ", g.currentPlayerTurn)
	panic("dd")
}

// index of our own position on the table
func (g *GameState) getPositionOnTable() int {
	for index, player := range g.playersList.List() {
		if g.listenAddr == player {
			return index
		}
	}

	return -1
}

func (g *GameState) getNextDealer() int {
	panic("TODO: getNextDealer not implemented")
}

func (g *GameState) getOtherPlayers() []string {
	players := []string{}

	for _, addr := range g.playersList.List() {
		if addr == g.listenAddr {
			continue
		}
		players = append(players, addr)
	}
	return players
}
