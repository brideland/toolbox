package moreos

import (
	"fmt"
	"os"
	"testing"
	"time"
)

func TestCancelReader_Read(t *testing.T) {

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}

	c, err := NewCancelReader(r, r.Fd())
	if err != nil {
		t.Fatal(err)
	}

	in := "myInput"
	go func() {
		time.Sleep(100 * time.Millisecond)
		w.WriteString(in)
	}()

	b := make([]byte, 32)

	n, err := c.Read(b)
	if err != nil {
		t.Fatalf("read failed: %s", err)
	}
	out := string(b[:n])

	if in != out {
		t.Fatalf("incorrect read: want %q, got %q", in, out)
	}
}

func TestCancelReader_CancelBefore(t *testing.T) {

	r, _, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}

	c, err := NewCancelReader(r, r.Fd())
	if err != nil {
		t.Fatal(err)
	}

	c.Cancel()

	ch := make(chan any)

	time.AfterFunc(time.Second, func() {
		ch <- "timeout"
	})

	go func() {
		b := make([]byte, 1)
		for i := range 2 {
			if i > 0 {
				time.Sleep(100 * time.Millisecond)
			}
			_, err = c.Read(b)
			if err != ErrReadCanceled {
				ch <- fmt.Errorf(
					"incorrect error: want ErrReadCanceled, got %q",
					err,
				)
			}
		}
		ch <- "success"
	}()

	switch v := (<-ch).(type) {
	case error:
		t.Fatal(v)
	case string:
		if v == "success" {
			return
		}
		if v == "timeout" {
			t.Fatal("test timed out")
		}
		panic("unreachable")
	default:
		panic("unreachable")
	}
}

func TestCancelReader_CancelDuring(t *testing.T) {

	testingPollWindowsEvChan = make(chan uint32, 16)
	defer func() {
		testingPollWindowsEvChan = nil
	}()

	r, _, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}

	c, err := NewCancelReader(r, r.Fd())
	if err != nil {
		t.Fatal(err)
	}

	go func() {
		time.Sleep(100 * time.Millisecond)
		c.Cancel()
	}()

	ch := make(chan any)

	time.AfterFunc(time.Second, func() {
		ch <- "timeout"
	})

	go func() {
		b := make([]byte, 1)
		for i := range 2 {
			if i > 0 {
				time.Sleep(100 * time.Millisecond)
			}
			_, err = c.Read(b)
			if err != ErrReadCanceled {
				ch <- fmt.Errorf(
					"incorrect error: want ErrReadCanceled, got %q",
					err,
				)
			}
		}
		ch <- "success"
	}()

	switch v := (<-ch).(type) {
	case error:
		t.Fatal(v)
	case string:
		if v == "success" {
			return
		}
		if v == "timeout" {
			t.Fatalf(
				"test timed out with poll values %v",
				dumpChan(testingPollWindowsEvChan),
			)
		}
		panic("unreachable")
	default:
		panic("unreachable")
	}
}

func dumpChan[T any](ch chan T) (s []T) {
	for {
		select {
		case v := <-ch:
			s = append(s, v)
		default:
			return s
		}
	}
}
