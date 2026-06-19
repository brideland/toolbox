package moreos_test

import (
	"os"
	"testing"
	"time"

	"github.com/brideland/toolbox/os/moreos"
)

func TestCancelReader_Read(t *testing.T) {

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}

	c, err := moreos.NewCancelReader(r, r.Fd())
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

	c, err := moreos.NewCancelReader(r, r.Fd())
	if err != nil {
		t.Fatal(err)
	}

	c.Cancel()

	b := make([]byte, 1)

	for i := range 2 {
		if i > 0 {
			time.Sleep(100 * time.Millisecond)
		}
		_, err = c.Read(b)
		if err != moreos.ErrReadCanceled {
			t.Fatalf(
				"incorrect error: want ErrReadCanceled, got %q",
				err,
			)
		}
	}
}

func TestCancelReader_CancelDuring(t *testing.T) {

	r, _, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}

	c, err := moreos.NewCancelReader(r, r.Fd())
	if err != nil {
		t.Fatal(err)
	}

	go func() {
		time.Sleep(100 * time.Millisecond)
		c.Cancel()
	}()

	b := make([]byte, 1)

	for i := range 2 {
		if i > 0 {
			time.Sleep(100 * time.Millisecond)
		}
		_, err = c.Read(b)
		if err != moreos.ErrReadCanceled {
			t.Fatalf(
				"incorrect error: want ErrReadCanceled, got %q",
				err,
			)
		}
	}
}
