package grep

const (
	opPlus = "+"
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
	tokPlus     = "+"
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

func (t *Token) cursor() int {
	switch t.typ {
	case tokPositive, tokNegative:
		return len(t.typ) + len(t.val) + 1
	case tokDigit, tokAlnum:
		return 2
	case tokRune, tokPlus:
		return 1
	default:
		return -1
	}
}
