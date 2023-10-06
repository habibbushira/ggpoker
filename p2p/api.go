package p2p

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

type apiFunc func(w http.ResponseWriter, r *http.Request) error

func makeHttpHandleFunc(f apiFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := f(w, r); err != nil {
			JSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
		}
	}
}

func JSON(w http.ResponseWriter, status int, v any) error {
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(v)
}

type APIServer struct {
	listenAddr string
	game       *GameState
}

func NewAPIServer(listenAddr string, game *GameState) *APIServer {
	return &APIServer{
		game:       game,
		listenAddr: listenAddr,
	}
}

func (s *APIServer) Run() {
	r := mux.NewRouter()

	r.HandleFunc("/ready", makeHttpHandleFunc(s.handlePlayerReady))
	r.HandleFunc("/fold", makeHttpHandleFunc(s.handlePlayerFold))
	r.HandleFunc("/check", makeHttpHandleFunc(s.handlePlayerCheck))
	r.HandleFunc("/bet/{value}", makeHttpHandleFunc(s.handlePlayerBet))
	http.ListenAndServe(s.listenAddr, r)
}

func (s *APIServer) handlePlayerReady(w http.ResponseWriter, r *http.Request) error {
	s.game.SetReady()
	return JSON(w, http.StatusOK, []byte("READY"))
}

func (s *APIServer) handlePlayerFold(w http.ResponseWriter, r *http.Request) error {
	if err := s.game.TakeAction(PlayerActionFold, 0); err != nil {
		return err
	}
	return JSON(w, http.StatusOK, []byte("FOLD"))
}

func (s *APIServer) handlePlayerCheck(w http.ResponseWriter, r *http.Request) error {
	if err := s.game.TakeAction(PlayerActionCheck, 0); err != nil {
		return err
	}
	return JSON(w, http.StatusOK, []byte("CHECK"))
}

func (s *APIServer) handlePlayerBet(w http.ResponseWriter, r *http.Request) error {
	valueStr := mux.Vars(r)["value"]
	value, err := strconv.Atoi(valueStr)

	if err != nil {
		return err
	}

	if err := s.game.TakeAction(PlayerActionBet, value); err != nil {
		return err
	}

	return JSON(w, http.StatusOK, []byte("BET"))
}
