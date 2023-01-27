package grep

import (
	"fmt"
	"strings"
	"unicode"
)

const (
	opPlus     = "+"
	opQuestion = "?"
)

type Operator struct {
	typ   string
	count int
}

const (
	tokPositive = "["
	tokNegative = "[^"
	tokDigit    = `\d`
	tokAlnum    = `\w`
	tokRune     = "rune"
	tokWildcard = "."
	tokOr       = "("
)

var emptyToken = Token{}

func NewToken(typ string, val ...string) Token {
	return Token{typ: typ, val: val}
}

type Token struct {
	val []string
	typ string
	op  *Operator
	cnt int
}

func (t *Token) match(s string) (ok bool, _ error) {
	r := rune(s[0])
	switch t.typ {
	case tokAlnum:
		ok = unicode.IsLetter(r)
	case tokDigit:
		ok = unicode.IsDigit(r)
	case tokPositive:
		ok = strings.ContainsAny(string(r), t.val[0])
	case tokNegative:
		ok = !strings.ContainsAny(string(r), t.val[0])
	case tokRune:
		ok = strings.HasPrefix(s, t.val[0])
	case tokWildcard:
		ok = true
	case tokOr:
		for _, val := range t.val {
			if strings.HasPrefix(s, val) {
				ok = true
				break
			}
		}
	default:
		return false, fmt.Errorf("unknown token type: id=%s", t.typ)
	}
	t.cnt++
	return ok, nil
}

func (t *Token) cursor() int {
	switch t.typ {
	case tokPositive, tokNegative:
		return len(t.typ) + len(t.val[0]) + 1
	case tokDigit, tokAlnum:
		return 2
	case tokWildcard, tokRune:
		return 1
	case tokOr:
		cnt := len(t.val) - 1 // for separator
		for _, val := range t.val {
			cnt += len(val)
		}
		return 1 + cnt + 1
	default:
		return -1
	}
}

func (t *Token) canReuse() bool {
	if t.op == nil {
		return false
	}
	switch t.op.typ {
	case opPlus:
		return true
	case opQuestion:
		return t.cnt == 1
	}
	return false
}

func (t *Token) canIgnore() bool {
	if t.op == nil {
		return false
	}
	switch t.op.typ {
	case opPlus:
		return t.cnt > 1
	case opQuestion:
		return true
	}
	return false
}
