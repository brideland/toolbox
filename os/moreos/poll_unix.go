//go:build unix

package moreos

import (
	"fmt"
	"syscall"

	"golang.org/x/sys/unix"
)

func poll(srcFd, cancelFd uintptr) error {

	pollFds := []unix.PollFd{
		{Events: unix.POLLIN, Fd: int32(srcFd)},
		{Events: unix.POLLIN, Fd: int32(cancelFd)},
	}

	for {
		n, err := unix.Poll(pollFds, -1)
		if err == syscall.EINTR {
			continue
		}
		if err != nil {
			return fmt.Errorf("poll() failed: %w", err)
		}
		if n == 0 {
			continue
		}
		if pollFds[0].Revents&unix.POLLIN != 0 {
			return nil
		}
		if pollFds[1].Revents&unix.POLLIN != 0 {
			return ErrReadCanceled
		}
		return fmt.Errorf("poll() returned unexpected event")
	}
}
