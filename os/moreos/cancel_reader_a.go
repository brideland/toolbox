//go:build unix || windows

package moreos

import (
	"fmt"
	"io"
	"os"
	"sync/atomic"
)

type cancelReader struct {
	src       io.Reader
	srcFd     uintptr
	readPipe  *os.File
	writePipe *os.File
	canceled  atomic.Bool
}

func newCancelReader(src io.Reader, srcFd uintptr) (*cancelReader, error) {

	readPipe, writePipe, err := os.Pipe()
	if err != nil {
		return nil, fmt.Errorf(
			"could not create pipe: %w",
			err,
		)
	}

	c := &CancelReader{
		src:       src,
		srcFd:     srcFd,
		readPipe:  readPipe,
		writePipe: writePipe,
	}

	return c, nil
}

func (c *cancelReader) read(b []byte) (int, error) {
	if c.canceled.Load() {
		return 0, ErrReadCanceled
	}
	err := poll(c.srcFd, c.readPipe.Fd())
	if err != nil {
		return 0, err
	}
	return c.src.Read(b)
}

func (c *cancelReader) cancel() {
	alreadyCanceled := c.canceled.Swap(true)
	if !alreadyCanceled {
		var b [1]byte
		c.writePipe.Write(b[:])
	}
}
