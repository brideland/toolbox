//go:build windows

package moreos

import (
	"fmt"

	"golang.org/x/sys/windows"
)

func poll(srcFd, cancelFd uintptr) error {

	handles := []windows.Handle{
		windows.Handle(srcFd),
		windows.Handle(cancelFd),
	}

	for {
		ev, err := windows.WaitForMultipleObjects(handles, false, windows.INFINITE)
		if err != nil {
			return fmt.Errorf("WaitForMultipleObjects() failed: %w", err)
		}
		if ev == windows.WAIT_FAILED {
			return fmt.Errorf("WaitForMultipleObjects() indicated failure")
		}
		if ev == windows.WAIT_OBJECT_0 {
			return nil
		}
		if ev == windows.WAIT_OBJECT_0+1 {
			return ErrReadCanceled
		}
		if testingPollWindowsEvChan != nil {
			testingPollWindowsEvChan <- ev
		}
	}
}
