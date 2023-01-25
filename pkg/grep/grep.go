package grep

import (
	"fmt"
	"log"
	"strings"
	"unicode"
	"unicode/utf8"
)

func Run(input, pattern string) (matched bool) {
	g := grep{}
	matched, err := g.matchLine(input, pattern)
	if err != nil {
		log.Println("ERROR:", err)
	}
	return matched
}

const (
	tokChar = iota
	tokBracket
	tokRune
	tokAnchor
	tokPlus
)

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
		var token *Token
		var err error
		if prev != nil && prev.op != nil {
			token = prev
		} else {
			token, err = g.nextToken(patterns)
			if err != nil {
				return false, err
			}
			if token.typ == tokPlus {
				token = prev
				token.op = &Operator{typ: opPlus}
			}
			prev = token
		}

		last, err = g.parse(*token, r)
		if err != nil {
			return false, err
		}
		if token.op != nil && token.cnt > 0 && !last.ok {
			log.Printf("unmatched: r=%q, token=%+v (continue)", r, token)
			token.op = nil
			g.cursor += len(token.s)
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
		g.cursor += len(token.s)
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

func (g *grep) nextToken(patterns string) (*Token, error) {
	s := string(patterns[g.cursor])
	switch {
	case s == "+":
		return &Token{s: s, typ: tokPlus}, nil
	case s == "^":
		return &Token{s: s, typ: tokAnchor}, nil
	case s == `\`:
		t := &Token{
			s:   patterns[g.cursor : g.cursor+2],
			typ: tokChar,
		}
		if s := string(patterns[g.cursor+1]); s == "d" || s == "w" {
			return t, nil
		}
		return nil, fmt.Errorf("unsupported charactor class: %s", t.s)
	case s == "[":
		idx := strings.Index(patterns[g.cursor:], "]")
		if idx == -1 {
			return nil, fmt.Errorf("unmatched bracket: %s", patterns[g.cursor:])
		}
		return &Token{s: patterns[g.cursor : idx+1], typ: tokBracket}, nil
	case utf8.RuneCountInString(s) == 1:
		return &Token{s: s, typ: tokRune}, nil
	}
	return nil, fmt.Errorf("unknown token: %s", s)
}

func (g *grep) parse(t Token, r rune) (*Result, error) {
	var rlt Result
	switch t.typ {
	case tokChar:
		rlt.ok = (t.s == `\d` && unicode.IsDigit(r)) || (t.s == `\w` && unicode.IsLetter(r))
	case tokBracket:
		if string(t.s[1]) != "^" {
			rlt.ok = strings.ContainsAny(string(r), t.s[1:len(t.s)-1])
			break
		}
		rlt.ok = !strings.ContainsAny(string(r), t.s[2:len(t.s)-1])
		if !rlt.ok {
			rlt.exit = true
		}
	case tokRune:
		rlt.ok = strings.ContainsRune(t.s, r)
	default:
		return nil, fmt.Errorf("unknown token type: id=%d", t.typ)
	}
	return &rlt, nil
}
