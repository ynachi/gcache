package frame

// Null frame.
type Null struct{}

func (n *Null) Serialize() []byte {
	return []byte(n.String())
}

// String provides a text representation of a Null frame.
func (n *Null) String() string {
	return "_\r\n"
}
