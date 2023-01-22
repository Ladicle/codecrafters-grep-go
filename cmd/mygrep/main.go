package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Usage: echo <input_text> | your_grep.sh -E <pattern>
func main() {
	if len(os.Args) < 3 || os.Args[1] != "-E" {
		fmt.Fprintf(os.Stderr, "usage: mygrep -E <pattern>\n")
		os.Exit(2) // 1 means no lines were selected, >1 means error
	}

	pattern := os.Args[2]

	line, err := io.ReadAll(os.Stdin) // assume we're only dealing with a single line
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: read input text: %v\n", err)
		os.Exit(2)
	}

	ok, err := matchLine(line, pattern)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
	}
	if !ok {
		os.Exit(1)
	}

	// default exit code is 0 which means success
}

const (
	TokenCharactorClass = iota
	TokenBracket
	TokenRune
	TokenAnchor
	TokenPlus
)

type Token struct {
	s   string
	typ int

	matched int
	plus    bool
}

type Result struct {
	ok   bool
	exit bool
}

func matchLine(line []byte, patterns string) (bool, error) {
	var cursor int

	tokenize := func() (*Token, error) {
		s := string(patterns[cursor])
		switch {
		case s == "+":
			return &Token{s: s, typ: TokenPlus}, nil
		case s == "^":
			return &Token{s: s, typ: TokenAnchor}, nil
		case s == `\`:
			t := &Token{
				s:   patterns[cursor : cursor+2],
				typ: TokenCharactorClass,
			}
			if s := string(patterns[cursor+1]); s == "d" || s == "w" {
				return t, nil
			}
			return nil, fmt.Errorf("unsupported charactor class: %s", t.s)
		case s == "[":
			idx := strings.Index(patterns[cursor:], "]")
			if idx == -1 {
				return nil, fmt.Errorf("unmatched bracket: %s", patterns[cursor:])
			}
			return &Token{s: patterns[cursor : idx+1], typ: TokenBracket}, nil
		case utf8.RuneCountInString(s) == 1:
			return &Token{s: s, typ: TokenRune}, nil
		}
		return nil, fmt.Errorf("unknown token: %s", s)
	}

	parse := func(t Token, r rune) (*Result, error) {
		var rlt Result
		switch t.typ {
		case TokenCharactorClass:
			rlt.ok = (t.s == `\d` && unicode.IsDigit(r)) || (t.s == `\w` && unicode.IsLetter(r))
		case TokenBracket:
			if string(t.s[1]) != "^" {
				rlt.ok = strings.ContainsAny(string(r), t.s[1:len(t.s)-1])
				break
			}
			rlt.ok = !strings.ContainsAny(string(r), t.s[2:len(t.s)-1])
			if !rlt.ok {
				rlt.exit = true
			}
		case TokenRune:
			rlt.ok = strings.ContainsRune(t.s, r)
		default:
			return nil, fmt.Errorf("unknown token type: id=%d", t.typ)
		}
		return &rlt, nil
	}

	s := string(patterns[0])
	hasAnchor := s == "^"
	if hasAnchor {
		cursor++
	}
	s = string(patterns[len(patterns)-1])
	hasEndAnchor := s == "$"
	if hasEndAnchor {
		patterns = patterns[:len(patterns)-1]
	}

	inputs := bytes.Runes(line)

	var idx int
	var prev *Token
	var last *Result
	for idx < len(inputs) {
		r := inputs[idx]
		var token *Token
		var err error
		if prev != nil && prev.plus {
			token = prev
		} else {
			token, err = tokenize()
			if err != nil {
				return false, err
			}
			if token.typ == TokenPlus {
				token = prev
				token.plus = true
			}
			prev = token
		}

		last, err = parse(*token, r)
		if err != nil {
			return false, err
		}
		if token.plus && token.matched > 0 && !last.ok {
			log.Printf("unmatched: r=%q, token=%+v (continue)", r, token)
			token.plus = false
			cursor += len(token.s)
			if len(patterns) == cursor {
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
		token.matched++
		cursor += len(token.s)
		if len(patterns) == cursor {
			break
		}
		idx++
	}
	ok := last != nil && last.ok && len(patterns) == cursor
	if ok && hasEndAnchor {
		return idx == len(inputs)-1, nil
	}
	return ok, nil
}
