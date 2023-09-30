package p2p

type GameStatus int32

func (g GameStatus) String() string {
	switch g {
	case GameStatusConnected:
		return "Connected"
	case GameStatusReady:
		return "READY"
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
	GameStatusDealing
	GameStatusPreFlop
	GameStatusFlop
	GameStatusTurn
	GameStatusRiver
)

type PlayerAction byte

func (pa PlayerAction) String() string {
	switch pa {
	case PlayerActionNone:
		return "NONE"
	case PlayerActionFold:
		return "FOLD"
	case PlayerActionCheck:
		return "CHECK"
	case PlayerActionBet:
		return "BET"
	default:
		return "INVALID"
	}
}

const (
	PlayerActionNone PlayerAction = iota
	PlayerActionFold
	PlayerActionCheck
	PlayerActionBet
)
