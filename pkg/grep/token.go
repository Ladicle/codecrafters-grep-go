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
)

var emptyToken = Token{}

func NewToken(typ, val string) Token {
	return Token{typ: typ, val: val}
}

type Token struct {
	val string
	typ string
	op  *Operator
	cnt int
}

func (t *Token) match(r rune) (ok bool, _ error) {
	switch t.typ {
	case tokAlnum:
		ok = unicode.IsLetter(r)
	case tokDigit:
		ok = unicode.IsDigit(r)
	case tokPositive:
		ok = strings.ContainsAny(string(r), t.val)
	case tokNegative:
		ok = !strings.ContainsAny(string(r), t.val)
	case tokRune:
		ok = strings.ContainsRune(t.val, r)
	default:
		return false, fmt.Errorf("unknown token type: id=%s", t.typ)
	}
	t.cnt++
	return ok, nil
}

func (t *Token) cursor() int {
	switch t.typ {
	case tokPositive, tokNegative:
		return len(t.typ) + len(t.val) + 1
	case tokDigit, tokAlnum:
		return 2
	case tokRune:
		return 1
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
