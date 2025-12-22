package stealth

import (
	"math/rand"
	"strings"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/input"
)

type TypingOptions struct {
	MinDelay            int
	MaxDelay            int
	TypoProbability     float64
	WordPauseMultiplier float64
	BurstProbability    float64
}

func DefaultTypingOptions() *TypingOptions {
	return &TypingOptions{MinDelay: 50, MaxDelay: 200, TypoProbability: 0.05, WordPauseMultiplier: 2.0, BurstProbability: 0.15}
}

func TypeHumanLike(page *rod.Page, text string, opts *TypingOptions) error {
	if opts == nil {
		opts = DefaultTypingOptions()
	}
	runes := []rune(text)
	inBurst := false
	for i := 0; i < len(runes); i++ {
		char := runes[i]
		if !inBurst && rand.Float64() < opts.BurstProbability {
			inBurst = true
		} else if inBurst && (char == ' ' || rand.Float64() < 0.3) {
			inBurst = false
		}
		var delay time.Duration
		if inBurst {
			delay = time.Duration(opts.MinDelay/2+rand.Intn(opts.MinDelay)) * time.Millisecond
		} else {
			delay = time.Duration(opts.MinDelay+rand.Intn(opts.MaxDelay-opts.MinDelay)) * time.Millisecond
		}
		if char == ' ' {
			delay = time.Duration(float64(delay) * opts.WordPauseMultiplier)
		}
		if rand.Float64() < opts.TypoProbability && char != ' ' {
			wrongChar := getTypoChar(char)
			if err := page.Keyboard.Type(input.Key(wrongChar)); err != nil {
				return err
			}
			time.Sleep(time.Duration(100+rand.Intn(300)) * time.Millisecond)
			if err := page.Keyboard.Type(input.Backspace); err != nil {
				return err
			}
			time.Sleep(time.Duration(50+rand.Intn(100)) * time.Millisecond)
		}
		if err := page.Keyboard.Type(input.Key(char)); err != nil {
			return err
		}
		time.Sleep(delay)
	}
	return nil
}

func getTypoChar(char rune) rune {
	typoMap := map[rune][]rune{
		'a': {'s', 'q', 'w', 'z'}, 'b': {'v', 'g', 'h', 'n'}, 'c': {'x', 'd', 'f', 'v'}, 'd': {'s', 'e', 'r', 'f'},
		'e': {'w', 'r', 'd', 's'}, 'f': {'d', 'r', 't', 'g'}, 'g': {'f', 't', 'y', 'h'}, 'h': {'g', 'y', 'u', 'j'},
		'i': {'u', 'o', 'k', 'j'}, 'j': {'h', 'u', 'i', 'k'}, 'k': {'j', 'i', 'o', 'l'}, 'l': {'k', 'o', 'p'},
		'm': {'n', 'j', 'k'}, 'n': {'b', 'h', 'j', 'm'}, 'o': {'i', 'p', 'l', 'k'}, 'p': {'o', 'l'},
		'q': {'w', 'a'}, 'r': {'e', 't', 'f', 'd'}, 's': {'a', 'w', 'e', 'd'}, 't': {'r', 'y', 'g', 'f'},
		'u': {'y', 'i', 'j', 'h'}, 'v': {'c', 'f', 'g', 'b'}, 'w': {'q', 'e', 's', 'a'}, 'x': {'z', 's', 'd', 'c'},
		'y': {'t', 'u', 'h', 'g'}, 'z': {'a', 's', 'x'},
	}
	lowerChar := rune(strings.ToLower(string(char))[0])
	if typos, exists := typoMap[lowerChar]; exists {
		typo := typos[rand.Intn(len(typos))]
		if char >= 'A' && char <= 'Z' {
			return rune(strings.ToUpper(string(typo))[0])
		}
		return typo
	}
	nearby := []rune{'a', 'e', 'i', 'o', 'u'}
	return nearby[rand.Intn(len(nearby))]
}

func ClearAndType(page *rod.Page, element *rod.Element, text string, opts *TypingOptions) error {
	shape, err := element.Shape()
	if err != nil {
		return err
	}
	if len(shape.Quads) > 0 {
		quad := shape.Quads[0]
		centerX := (quad[0] + quad[2]) / 2
		centerY := (quad[1] + quad[5]) / 2
		if err := MoveAndClick(page, centerX, centerY, DefaultMoveMouseOptions()); err != nil {
			return err
		}
	} else {
		element.MustClick()
	}
	time.Sleep(time.Duration(100+rand.Intn(200)) * time.Millisecond)
	if err := page.Keyboard.Press(input.ControlLeft); err != nil {
		return err
	}
	if err := page.Keyboard.Type('a'); err != nil {
		return err
	}
	if err := page.Keyboard.Release(input.ControlLeft); err != nil {
		return err
	}
	time.Sleep(time.Duration(50+rand.Intn(100)) * time.Millisecond)
	return TypeHumanLike(page, text, opts)
}

func PressEnter(page *rod.Page) error {
	time.Sleep(time.Duration(100+rand.Intn(200)) * time.Millisecond)
	return page.Keyboard.Type(input.Enter)
}
