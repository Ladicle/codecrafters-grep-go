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

func matchLine(line []byte, patterns string) (bool, error) {
	var cursor int
	var ok bool
	for _, r := range bytes.Runes(line) {
		log.Printf("len=%d, cursor=%d", len(patterns), cursor)
		if len(patterns) == cursor {
			return true, nil
		}
		if r == '\n' {
			log.Println("END of Line")
			break
		}
		s := string(patterns[cursor])
		switch {
		case s == `\`:
			cursor++
			s = string(patterns[cursor])
			if (s == "d" && unicode.IsDigit(r)) || (s == "w" && unicode.IsLetter(r)) {
				cursor++
				ok = true
				continue
			} else {
				cursor--
				s = string(patterns[cursor])
			}
		case s == "[":
			parts := strings.SplitN(patterns[cursor+1:], "]", 2)
			if len(parts) != 2 {
				return false, fmt.Errorf("invalid input: unmatched bracket: %s", patterns)
			}
			cursor += 1 + len(parts[0]) + 1
			if parts[0][0] == '^' {
				if !strings.ContainsAny(string(r), parts[0][1:]) {
					ok = true
					continue
				}
				log.Println("unmatched negative charactor groups")
				return false, nil
			} else if strings.ContainsAny(string(r), parts[0]) {
				ok = true
				continue
			}
			cursor -= 1 + len(parts[0]) + 1
			s = string(patterns[cursor])
		case utf8.RuneCountInString(s) == 1 && strings.ContainsRune(s, r):
			cursor++
			ok = true
			continue
		}
		// unmatched
		log.Printf("invalid argument: input=%q, pattern=%q", string(r), s)
		ok = false
	}
	if ok {
		return len(patterns) == cursor, nil
	}
	log.Printf("unmatched: lines=%q, pattern=%q", line, patterns)
	return false, nil
}
