package frame

// Bool implements Framer interface.
type Bool struct {
	value bool
}

// Serialize turns the frame into a slice of byte for transfer over a network stream.
func (b *Bool) Serialize() []byte {
	return []byte(b.String())
}

// String provides a text representation of a Bool frame.
func (b *Bool) String() string {
	if b.value {
		return "#t\r\n"
	}
	return "#f\r\n"
}
