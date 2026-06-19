package moreos

import (
	"errors"
	"io"
)

var (
	ErrReadCanceled   = errors.New("read canceled")
	ErrNotImplemented = errors.New("not implemented")
)

// CancelReader is returned by [NewCancelReader].
type CancelReader = cancelReader

// NewCancelReader returns a new [CancelReader].
//
// CancelReader allows breaking out of blocking reads
// from fast pollable readers like pipes and consoles.
//
// will cause a timely return from a blocked [CancelReader.Read]
// (which will either return successfully or with ErrCanceled)
// and any subsequent [CancelReader.Read] calls
// will return ErrCanceled.
//
// Reading from src while [CancelReader.Read] is running
// can lead to it blocking in a manner that [CancelReader.Cancel] cannot unblock.
func NewCancelReader(src io.Reader, srcFd uintptr) (*CancelReader, error) {
	return newCancelReader(src, srcFd)
}

func (c *CancelReader) Read(b []byte) (int, error) {
	return c.read(b)
}

func (c *CancelReader) Cancel() {
	c.cancel()
}
