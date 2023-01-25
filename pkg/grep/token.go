package grep

type opType string

const (
	opPlus opType = "+"
)

type Token struct {
	s   string
	typ int
	len int

	op  *Operator
	cnt int
}

type Operator struct {
	typ   opType
	count int
}
