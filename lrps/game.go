package lrps

import "errors"

type Hand int

const (
	HandNil Hand = iota
	HandRock
	HandPaper
	HandScissor
)

var ErrInvalidHand = errors.New("Invalid hand")

func ParseHand(s string) (Hand, error) {
	switch s {
	case "rock":
		return HandRock, nil
	case "paper":
		return HandPaper, nil
	case "scissor":
		return HandScissor, nil
	}
	return HandNil, ErrInvalidHand
}

func (h Hand) String() string {
	switch h {
	case HandRock:
		return "rock"
	case HandPaper:
		return "paper"
	case HandScissor:
		return "scissor"
	}
	return "nil"
}

func CompareHands(a Hand, b Hand) int {
	if a == b || a == HandNil || b == HandNil {
		return 0
	}

	n := int(b - a)
	if n == 1 || n == -1 {
		return -n
	}
	return n / 2
}
