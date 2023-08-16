package p2p

type GameStatus int32

func (g GameStatus) String() string {
	switch g {
	case GameStatusWaiting:
		return "WAITING"
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
	GameStatusWaiting GameStatus = iota
	GameStatusShuffleAndDeal
	GameStatusReceiving
	GameStatusDealing
	GameStatusPreFlop
	GameStatusFlop
	GameStatusTurn
	GameStatusRiver
)
