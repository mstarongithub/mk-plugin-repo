package util

import "io"

type BodyOverwrite struct {
	NewBody io.Reader
	OldBody io.ReadCloser

	closed bool
}

func (b *BodyOverwrite) Read(p []byte) (n int, err error) {
	if b.closed {
		return 0, io.EOF
	}
	return b.NewBody.Read(p)
}

func (b *BodyOverwrite) Close() error {
	b.closed = true
	return b.OldBody.Close()
}
