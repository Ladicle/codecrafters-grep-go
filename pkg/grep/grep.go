package grep

import (
	"errors"
	"log"
	"strings"
)

func Run(input, pattern string) (matched bool) {
	grep := newGrep(pattern)
	matched, err := grep.matchLine(input)
	if err != nil {
		log.Println("ERROR:", err)
	}
	return matched
}

func newGrep(pattern string) grep {
	anchor := strings.HasPrefix(pattern, "^")
	if anchor {
		pattern = pattern[1:]
	}
	endAnchor := strings.HasSuffix(pattern, "$")
	if endAnchor {
		pattern = pattern[:len(pattern)-1]
	}
	return grep{
		pattern:   pattern,
		anchor:    anchor,
		endAnchor: endAnchor,
	}
}

type grep struct {
	pattern string

	anchor    bool
	endAnchor bool
}

func (g *grep) matchLine(line string) (bool, error) {
	inputs := []rune(line)

	token, ok := g.nextToken()
	if !ok {
		return false, errors.New("no pattern")
	}

	var matched bool
	var idx int
	for idx < len(inputs) {
		var err error
		matched, err = token.match(inputs[idx])
		if err != nil {
			return false, err
		}

		if matched {
			idx++
			if !token.canReuse() {
				token, ok = g.nextToken()
				if !ok {
					break
				}
			}
			continue
		}

		if g.anchor || token.typ == tokNegative {
			break
		}
		if !token.canIgnore() {
			idx++
			continue
		}
		token, ok = g.nextToken()
		if !ok {
			break
		}
	}

	ok = matched &&
		!g.hasNextToken() &&
		(token == emptyToken || token.cnt > 0 || token.canIgnore())

	if ok && g.endAnchor {
		return idx >= len(inputs), nil
	}
	log.Printf("matched=%v, g=%+v, token=%+v", matched, g, token)
	return ok, nil
}

func (g *grep) hasNextToken() bool {
	return len(g.pattern) > 0
}

func (g *grep) nextToken() (tok Token, ok bool) {
	if !g.hasNextToken() {
		return emptyToken, false
	}

	defer func() {
		if !g.hasNextToken() {
			return
		}
		switch string(g.pattern[0]) {
		case opPlus:
			tok.op = &Operator{typ: opPlus}
		case opQuestion:
			tok.op = &Operator{typ: opQuestion}
		default:
			return
		}
		g.pattern = g.pattern[1:]
	}()

	switch {
	case strings.HasPrefix(g.pattern, tokWildcard):
		tok = NewToken(tokWildcard, "")
	case strings.HasPrefix(g.pattern, tokAlnum):
		tok = NewToken(tokAlnum, "")
	case strings.HasPrefix(g.pattern, tokDigit):
		tok = NewToken(tokDigit, "")
	case strings.HasPrefix(g.pattern, tokNegative):
		idx := strings.Index(g.pattern, "]")
		tok = NewToken(tokNegative, g.pattern[2:idx])
	case strings.HasPrefix(g.pattern, tokPositive):
		idx := strings.Index(g.pattern, "]")
		tok = NewToken(tokPositive, g.pattern[1:idx])
	default:
		tok = NewToken(tokRune, string(g.pattern[0]))
	}

	g.pattern = g.pattern[tok.cursor():]
	return tok, true
}
