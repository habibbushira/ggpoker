package p2p

type GameStatus int32

func (g GameStatus) String() string {
	switch g {
	case GameStatusConnected:
		return "Connected"
	case GameStatusReady:
		return "READY"
	case GameStatusShuffleAndDeal:
		return "SHUFFLE AND DEAL"
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
	GameStatusConnected GameStatus = iota
	GameStatusReady
	GameStatusShuffleAndDeal
	GameStatusReceiving
	GameStatusDealing
	GameStatusPreFlop
	GameStatusFlop
	GameStatusTurn
	GameStatusRiver
)
