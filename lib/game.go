package lib

import (
	"fmt"
	"math/rand"
	"os"
)

type Game struct {
	GameID      string
	Players     []Player
	Initialized bool
	Started     bool

	Hints         int
	Bombs         int
	Deck          []Card
	Discard       []Card
	Piles         []int
	CurrentPlayer int
	StartingTime  int
	Finished      bool
	Won           bool
	TurnsLeft     int
}

func NewGame(id string) *Game {
	g := new(Game)
	g.GameID = id

	// figure out how many cards are in the Deck
	maxCards := 0
	for _, count := range numbers {
		maxCards += count * len(colors)
	}

	// populate the Deck, Discard, and Piles
	g.Deck = make([]Card, maxCards, maxCards)
	g.PopulateDeck()
	g.Discard = make([]Card, 0, maxCards)
	g.Piles = make([]int, len(colors), len(colors))

	// set starting values
	g.Hints = startingHints
	g.Bombs = startingBombs
	g.TurnsLeft = -1

	// start with no Players
	g.Players = make([]Player, 0, len(cardsInHand)-1)
	g.Initialized = true
	fmt.Printf("Game '%s' initialized.\n", g.GameID)

	return g
}

func (g *Game) AddPlayer(id string) bool {
	if !g.Initialized {
		fmt.Printf("Attempting to add player before game has been fully Initialized.")
		return false
	}
	if g.Started {
		fmt.Printf("Attempting to add player after game has Started.")
		return false
	}
	if g.Finished {
		fmt.Printf("Attempting to add Players after game has ended.")
		return false
	}
	if len(g.Players) >= len(cardsInHand)-1 {
		fmt.Printf("Attempted to add a player to a full game.")
		return false
	}

	g.Players = append(g.Players, Player{ID: id})
	fmt.Printf("Player '%s' joined game '%s'.\n", id, g.GameID)
	return true
}

func (g *Game) Start() {
	if !g.Initialized {
		fmt.Printf("Attempting to start before game has been fully Initialized.")
		os.Exit(1)
	}
	if g.Started {
		fmt.Printf("Attempting to start a game already in progress.")
		os.Exit(1)
	}
	if g.Finished {
		fmt.Printf("Attempting to start a game after it has ended.")
		os.Exit(1)
	}
	numPlayers := len(g.Players)
	if numPlayers >= len(cardsInHand) || cardsInHand[numPlayers] == 0 {
		fmt.Printf("Attempted to start game with invalid number of Players.")
		os.Exit(1)
	}

	// create hands
	for index, _ := range g.Players {
		g.Players[index].Initialize(cardsInHand[numPlayers])
		for i := 0; i < cardsInHand[numPlayers]; i++ {
			fmt.Printf("Adding card to player's hand.")
			g.Players[index].AddCard(g.DrawCard())
		}
	}

	// let's do it
	g.CurrentPlayer = rand.Intn(numPlayers)
	g.Started = true
}

func (g *Game) ProcessMove(m *Message) bool {
	if !g.Started {
		fmt.Printf("Attempting to process move for a game that hasn't Started yet.")
		return false
	}
	if g.Finished {
		fmt.Printf("Attempting to process a move for a Finished game.")
		return false
	}
	p := g.GetPlayerByID(m.Player)
	if p == nil {
		fmt.Printf("Attempting to process a move for a nonexistent player.")
		return false
	}

	// TODO: more checking that it's a valid move

	card := p.GetCard(m.CardIndex)

	if m.MoveType == MovePlay {
		p.RemoveCard(m.CardIndex)

		if g.PlayCard(card) {
			// play was successful!
			if g.PilesComplete() {
				g.Win()
			}
		} else {
			// play was unsuccessful :(
			g.Bombs--
			if g.Bombs == 0 {
				g.Lose()
			}
			g.Discard = append(g.Discard, card)
		}
	} else if m.MoveType == MoveDiscard {
		p.RemoveCard(m.CardIndex)
		g.Discard = append(g.Discard, card)
		g.Hints++
		if g.Hints > maxHints {
			g.Hints = maxHints
		}
	} else if m.MoveType == MoveHint {
		if g.Hints <= 0 {
			return false
		}
		hintReceiver := g.GetPlayerByID(m.HintPlayer)
		if hintReceiver == nil {
			fmt.Printf("Attempting to give hint to a nonexistent player.")
			return false
		}
		hintReceiver.ReceiveHint(m.CardIndex, m.HintInfoType)
		g.Hints--
	} else {
		fmt.Printf("Attempting to process unknown move type.")
		return false
	}

	if m.MoveType == MovePlay || m.MoveType == MoveDiscard {
		if len(g.Deck) > 0 {
			p.AddCard(g.DrawCard())
		} else if g.TurnsLeft == -1 {
			// Deck is empty, start the countdown
			g.TurnsLeft = len(g.Players)
		}
	}

	if g.TurnsLeft == 0 {
		g.Lose()
	}

	// TODO: log move (if it's valid)
	return true
}

func (g *Game) CreateState(playerid string) Game {
	p := g.GetPlayerByID(playerid)

	gCopy := Game{}
	gCopy = *g

	// clear the Deck (could be used to determine your hand)
	gCopy.Deck = make([]Card, 0, 0)

	// clear your hand, except for revealed info
	newPlayers := make([]Player, len(g.Players), len(g.Players))
	for playerIndex, player := range gCopy.Players {
		if p.ID == player.ID {
			newHand := make([]Card, len(player.Cards), len(player.Cards))
			for cardIndex, card := range player.Cards {
				card.Color = ""
				card.Number = 0
				newHand[cardIndex] = card
			}
			player.Cards = newHand
		}
		newPlayers[playerIndex] = player
	}
	gCopy.Players = newPlayers
	return gCopy
}

func (g *Game) Win() {
	g.Finished = true
	g.Won = true
}

func (g *Game) Lose() {
	g.Finished = true
	g.Won = false
}
