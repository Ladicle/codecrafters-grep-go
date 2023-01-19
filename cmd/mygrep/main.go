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
)

type token struct {
	str string
	typ int
}

type result struct {
	ok   bool
	exit bool
}

func matchLine(line []byte, patterns string) (bool, error) {
	var cursor int

	tokenize := func() (*token, error) {
		s := string(patterns[cursor])
		if s == `\` {
			t := &token{
				str: patterns[cursor : cursor+2],
				typ: TokenCharactorClass,
			}
			if s := string(patterns[cursor+1]); s == "d" || s == "w" {
				return t, nil
			}
			return nil, fmt.Errorf("unsupported charactor class: %s", t.str)
		}
		if s == "[" {
			idx := strings.Index(patterns[cursor:], "]")
			if idx == -1 {
				return nil, fmt.Errorf("unmatched bracket: %s", patterns[cursor:])
			}
			return &token{str: patterns[cursor+1 : idx+1], typ: TokenBracket}, nil
		}
		if utf8.RuneCountInString(s) == 1 {
			return &token{str: s, typ: TokenRune}, nil
		}
		return nil, fmt.Errorf("unknown token: %s", s)
	}

	parse := func(t token, r rune) (*result, error) {
		var rlt result
		switch t.typ {
		case TokenCharactorClass:
			rlt.ok = (t.str == `\d` && unicode.IsDigit(r)) || (t.str == `\w` && unicode.IsLetter(r))
		case TokenBracket:
			if string(t.str[1]) != "^" {
				rlt.ok = strings.ContainsAny(string(r), t.str[1:len(t.str)-1])
				break
			}
			rlt.ok = !strings.ContainsAny(string(r), t.str[2:len(t.str)-1])
			if !rlt.ok {
				rlt.exit = true
			}
		case TokenRune:
			rlt.ok = strings.ContainsRune(t.str, r)
		default:
			return nil, fmt.Errorf("unknown token type: id=%d", t.typ)
		}
		return &rlt, nil
	}

	var last *result
	for _, r := range bytes.Runes(line) {
		token, err := tokenize()
		if err != nil {
			return false, err
		}
		last, err = parse(*token, r)
		if err != nil {
			return false, err
		}
		if last.exit {
			return last.ok, nil
		}
		if !last.ok {
			log.Printf("unmatched: r=%q, token=%+v", r, token)
			continue
		}
		cursor += len(token.str)
		if len(patterns) == cursor {
			break
		}
	}
	return last != nil && last.ok && len(patterns) == cursor, nil
}
