package frame

import "io"

// Null frame.
type Null struct{}

func (n *Null) Serialize() []byte {
	return []byte(n.String())
}

// String provides a text representation of a Null frame.
func (n *Null) String() string {
	return "_\r\n"
}

// WriteTo writes a frame to an io.reader.
func (n *Null) WriteTo(w io.Writer) (int64, error) {
	frameToBytes := n.Serialize()
	count, err := w.Write(frameToBytes)
	return int64(count), err
}
