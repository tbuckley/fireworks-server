package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/tbuckley/fireworks-server/lib"
)

var (
	errGameNotFound = errors.New("game not found")
)

type Server struct {
	games map[string]*lib.Game
}

func (s *Server) WithMessage(w http.ResponseWriter, r *http.Request, handler func(*lib.Message)) {
	m, ok := lib.DecodeMove(r.PostFormValue("data"))
	if !ok {
		log.Printf("Malformed message, Discarding.")
		http.Error(w, "Error: malformed JSON message", http.StatusInternalServerError)
		return
	}
	handler(m)
}

func (s *Server) WithGame(w http.ResponseWriter, m *lib.Message, handler func(*lib.Game)) {
	game, err := s.getGame(m.Game)
	if err != nil {
		log.Printf("Error getting/retrieving game: %v", err.Error())
		http.Error(w, "Error: problem retrieving game", http.StatusInternalServerError)
		return
	}
	handler(game)
}

func (s *Server) WithPlayer(w http.ResponseWriter, m *lib.Message, g *lib.Game, handler func(*lib.Player)) {
	player := g.GetPlayerByID(m.Player)
	if player == nil {
		fmt.Printf("Attempting to make a move with nonexistent player.")
		fmt.Fprintf(w, "Error: Attempting to make a move with nonexistent player.")
		return
	}
	handler(player)
}

// getGame returns the game if it exists, and errGameNotFound if not.
func (s *Server) getGame(id string) (*lib.Game, error) {
	game, ok := s.games[id]
	if !ok {
		log.Printf("Could not find existing game: %v", id)
		return nil, errGameNotFound
	}

	log.Printf("Found existing game: %v", id)
	return game, nil
}

// getOrCreateGame returns the game if it exists, creating a new one if not.
// An error may be returned if a new game cannot be created.
func (s *Server) getOrCreateGame(id string) (*lib.Game, error) {
	game, err := s.getGame(id)

	if err == errGameNotFound {
		log.Printf("Creating new game: %v", id)
		game = lib.NewGame(id)
		s.games[id] = game
	} else if err != nil {
		return nil, err
	}

	return game, nil
}

// handleJoin handles a new player joining the game.
func (s *Server) handleJoin(w http.ResponseWriter, r *http.Request) {
	s.WithMessage(w, r, func(m *lib.Message) {
		s.WithGame(w, m, func(g *lib.Game) {
			player := g.GetPlayerByID(m.Player)

			// add player if it doesn't exist
			if player == nil {
				g.AddPlayer(m.Player)
			} else {
				log.Printf("Player already exists: %v", player.ID)
			}

			fmt.Fprintf(w, lib.EncodeGame(g.CreateState(m.Player)))
		})
	})
}

// handleStart begins the game if it has not already begun.
func (s *Server) handleStart(w http.ResponseWriter, r *http.Request) {
	s.WithMessage(w, r, func(m *lib.Message) {
		s.WithGame(w, m, func(g *lib.Game) {
			s.WithPlayer(w, m, g, func(p *lib.Player) {
				if g.Started {
					fmt.Printf("Attempting to start already started game.")
					fmt.Fprintf(w, "Error: Attempting to start already started game.")
					return
				}

				g.Start()
				fmt.Fprintf(w, lib.EncodeGame(g.CreateState(m.Player)))
			})
		})
	})
}

// handleMove processes a player's move.
func (s *Server) handleMove(w http.ResponseWriter, r *http.Request) {
	s.WithMessage(w, r, func(m *lib.Message) {
		s.WithGame(w, m, func(g *lib.Game) {
			s.WithPlayer(w, m, g, func(p *lib.Player) {
				g.ProcessMove(m)
				log.Printf("Global game state: %#v", g)
				fmt.Fprintf(w, lib.EncodeGame(g.CreateState(m.Player)))
			})
		})
	})
}

func main() {
	// TODO: check to make sure no other server is running

	// initialize server
	s := new(Server)
	s.games = make(map[string]*lib.Game)

	// create router
	r := mux.NewRouter()
	r.HandleFunc("/api/join", s.handleJoin)
	r.HandleFunc("/api/start", s.handleStart)
	r.HandleFunc("/api/move", s.handleMove)
	r.Handle("/", http.FileServer(http.Dir(lib.ClientDirectory)))

	// listen for connections
	http.Handle("/", r)
	http.ListenAndServe(":"+lib.Port, nil)
}
