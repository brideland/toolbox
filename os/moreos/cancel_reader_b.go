//go:build !unix && !windows

package moreos

import (
	"io"
)

type cancelReader struct{}

func newCancelReader(_ io.Reader, _ uintptr) (*CancelReader, error) {
	return nil, ErrNotImplemented
}

func (c *cancelReader) read(_ []byte) (int, error) {
	return 0, nil
}

func (c *cancelReader) cancel() {}
