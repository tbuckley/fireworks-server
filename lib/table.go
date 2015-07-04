package lib

// game.go was getting big, so I separated some stuff into here

import (
	rand "math/rand"
)

func (g *Game) PopulateDeck() {
	i := 0
	for number, count := range numbers {
		for _, color := range colors {
			for j := 0; j < count; j++ {
				g.deck[i].Color = color
				g.deck[i].Number = number
				i++
			}
		}
	}
}

func (g *Game) DrawCard() Card {
	index := rand.Intn(len(g.deck))
	card := g.deck[index]
	g.deck = append(g.deck[:index], g.deck[index+1:]...)
	return card
}

func (g *Game) PlayCard(c Card) bool {
	for index, count := range g.piles {
		if colors[index] == c.Color {
			if count+1 == c.Number {
				// good play!
				g.piles[index]++
				return true
			} else {
				return false
			}
		}
	}
	return false
}

func (g *Game) PilesComplete() bool {
	for _, count := range g.piles {
		if count != len(numbers)-1 {
			return false
		}
	}
	return true
}
