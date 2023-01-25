package grep

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"unicode"
)

func Run(input, pattern string) (matched bool) {
	g := grep{}
	matched, err := g.matchLine(input, pattern)
	if err != nil {
		log.Println("ERROR:", err)
	}
	return matched
}

type Result struct {
	ok   bool
	exit bool
}

type grep struct {
	cursor int
}

func (g *grep) matchLine(line, patterns string) (bool, error) {
	hasAnchor := strings.HasPrefix(patterns, "^")
	if hasAnchor {
		g.cursor++
	}
	hasEndAnchor := strings.HasSuffix(patterns, "$")
	if hasEndAnchor {
		patterns = patterns[:len(patterns)-1]
	}

	inputs := []rune(line)

	var idx int
	var prev *Token
	var last *Result
	for idx < len(inputs) {
		r := inputs[idx]
		var token Token
		var err error
		if prev != nil && prev.op != nil {
			token = *prev
		} else {
			var ok bool
			token, ok = g.next(patterns)
			if !ok {
				return false, errors.New("failed to get token")
			}
			if token.typ == tokPlus {
				token = *prev
				token.op = &Operator{typ: opPlus}
			}
			prev = &token
		}

		last, err = g.parse(token, r)
		if err != nil {
			return false, err
		}
		if token.op != nil && token.cnt > 0 && !last.ok {
			log.Printf("unmatched: r=%q, token=%+v (continue)", r, token)
			token.op = nil
			g.cursor += token.cursor()
			if len(patterns) == g.cursor {
				break
			}
			continue
		}
		if last.exit {
			return last.ok, nil
		}
		if !last.ok {
			log.Printf("unmatched: r=%q, token=%+v", r, token)
			if hasAnchor {
				return false, nil
			}
			idx++
			continue
		}
		token.cnt++
		g.cursor += token.cursor()
		if len(patterns) == g.cursor {
			break
		}
		idx++
	}
	ok := last != nil && last.ok && len(patterns) == g.cursor
	if ok && hasEndAnchor {
		return idx == len(inputs)-1, nil
	}
	return ok, nil
}

func (g *grep) next(p string) (_ Token, ok bool) {
	if g.cursor >= len(p) {
		return emptyToken, false
	}
	var tok Token
	p = p[g.cursor:]
	switch {
	case strings.HasPrefix(p, tokPlus):
		tok = NewToken(tokPlus, "")
	case strings.HasPrefix(p, tokAlnum):
		tok = NewToken(tokAlnum, "")
	case strings.HasPrefix(p, tokDigit):
		tok = NewToken(tokDigit, "")
	case strings.HasPrefix(p, tokNegative):
		idx := strings.Index(p, "]")
		tok = NewToken(tokNegative, p[2:idx])
	case strings.HasPrefix(p, tokPositive):
		idx := strings.Index(p, "]")
		tok = NewToken(tokPositive, p[1:idx])
	default:
		tok = NewToken(tokRune, string(p[0]))
	}
	return tok, true
}

func (g *grep) parse(t Token, r rune) (*Result, error) {
	var rlt Result
	switch t.typ {
	case tokAlnum:
		rlt.ok = unicode.IsLetter(r)
	case tokDigit:
		rlt.ok = unicode.IsDigit(r)
	case tokPositive:
		rlt.ok = strings.ContainsAny(string(r), t.val)
	case tokNegative:
		rlt.ok = !strings.ContainsAny(string(r), t.val)
		rlt.exit = !rlt.ok
	case tokRune:
		rlt.ok = strings.ContainsRune(t.val, r)
	default:
		return nil, fmt.Errorf("unknown token type: id=%s", t.typ)
	}
	return &rlt, nil
}
