// Copyright 2025 the toolbox authors.
// SPDX-License-Identifier: Apache-2.0

package streamcrypt_test

import (
	"bytes"
	"io"
	"testing"

	"github.com/google/go-cmp/cmp"
	sc "github.com/brideland/toolbox/crypto/streamcrypt"
	"github.com/brideland/toolbox/must"
)

func TestStreamCrypt(t *testing.T) {

	tests := []struct {
		mode sc.Mode
	}{
		{mode: sc.ModeXChaCha20},
		{mode: sc.ModeAES256CTR},
	}

	input := []byte("hello world")

	passwordString := "mypass123"
	password := []byte(passwordString)
	passFunc := func() ([]byte, error) {
		return []byte(passwordString), nil
	}

	for _, test := range tests {
		t.Run(test.mode.String(), func(t *testing.T) {

			ciphertext := sc.Encrypt(input, password, sc.WithMode(test.mode))
			output, err := sc.Decrypt(ciphertext, passFunc)

			if err != nil {
				t.Fatalf("could not decrypt: %s", err)
			}

			if diff := cmp.Diff(input, output); diff != "" {
				t.Errorf("incorrect result (-want +got):\n%s", diff)
			}
		})
	}
}

func BenchmarkWrite(b *testing.B) {

	benches := []struct {
		mode sc.Mode
	}{
		{mode: sc.ModeXChaCha20},
		{mode: sc.ModeAES256CTR},
	}

	password := []byte("mypass123")
	plaintext := []byte("hello world")

	for _, bench := range benches {
		b.Run(bench.mode.String(), func(b *testing.B) {

			w := sc.NewEncryptor(io.Discard, password, sc.WithMode(bench.mode))

			b.ResetTimer()

			for b.Loop() {
				w.Write(plaintext)
			}
		})
	}
}

func BenchmarkRead(b *testing.B) {

	benches := []struct {
		mode sc.Mode
	}{
		{mode: sc.ModeXChaCha20},
		{mode: sc.ModeAES256CTR},
	}

	passwordString := "mypass123"
	password := []byte(passwordString)
	passFunc := func() ([]byte, error) {
		return []byte(passwordString), nil
	}

	for _, bench := range benches {
		b.Run(bench.mode.String(), func(b *testing.B) {

			ciphertextBuf := new(bytes.Buffer)

			w := sc.NewEncryptor(ciphertextBuf, password, sc.WithMode(bench.mode))
			w.Write([]byte("hello world"))

			rr := &repeatReader{
				b: ciphertextBuf.Bytes(),
			}

			r := sc.NewDecryptor(rr, passFunc)
			readBuf := make([]byte, ciphertextBuf.Len())

			b.ResetTimer()

			for b.Loop() {
				must.Get(r.Read(readBuf))
			}
		})
	}
}

var global []byte

func BenchmarkEncrypt(b *testing.B) {

	benches := []struct {
		mode sc.Mode
	}{
		{mode: sc.ModeXChaCha20},
		{mode: sc.ModeAES256CTR},
	}

	password := []byte("mypass123")
	plaintext := []byte("hello world")

	for _, bench := range benches {
		b.Run(bench.mode.String(), func(b *testing.B) {

			for b.Loop() {
				global = sc.Encrypt(plaintext, password, sc.WithMode(bench.mode))
			}

			b.ReportMetric(
				float64(b.Elapsed().Milliseconds())/float64(b.N),
				"ms/op",
			)
		})
	}
}

func BenchmarkDecrypt(b *testing.B) {

	benches := []struct {
		mode sc.Mode
	}{
		{mode: sc.ModeXChaCha20},
		{mode: sc.ModeAES256CTR},
	}

	plaintext := []byte("hello world")

	passwordString := "mypass123"
	password := []byte(passwordString)
	passFunc := func() ([]byte, error) {
		return []byte(passwordString), nil
	}

	var err error

	for _, bench := range benches {
		b.Run(bench.mode.String(), func(b *testing.B) {

			ciphertext := sc.Encrypt(plaintext, password, sc.WithMode(bench.mode))

			b.ResetTimer()

			for b.Loop() {
				global, err = sc.Decrypt(ciphertext, passFunc)
				if err != nil {
					b.Fatalf("could not decrypt: %s", err)
				}
			}

			b.ReportMetric(
				float64(b.Elapsed().Milliseconds())/float64(b.N),
				"ms/op",
			)
		})
	}
}

type repeatReader struct {
	b []byte
}

func (r *repeatReader) Read(b []byte) (int, error) {
	return copy(b, r.b), nil
}
