// Copyright 2025 the toolbox authors.
// SPDX-License-Identifier: Apache-2.0

package streamcrypt

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"io"

	"github.com/brideland/toolbox/must"
	"golang.org/x/crypto/chacha20"
)

var (
	ErrUnsupportedMode        = errors.New("incorrect or unsupported encryption mode")
	ErrHeaderParamsOutOfRange = errors.New("header params out of range")
)

const (
	argonSaltSize = 16
	aesKeyLen     = 32
)

type bin struct {
	Mode         Mode
	ArgonTime    uint32
	ArgonMemory  uint32
	ArgonThreads uint8
	ArgonSalt    [argonSaltSize]byte
	ChachaNonce  [chacha20.NonceSizeX]byte
	AesIV        [aes.BlockSize]byte
}

type header struct {
	bin
	argonTimeMax    uint32
	argonMemoryMax  uint32
	argonThreadsMax uint8
}

func newHeader(c *config) (h header) {

	h = header{
		bin: bin{
			Mode:         c.mode,
			ArgonTime:    c.argonTime,
			ArgonMemory:  c.argonMemory,
			ArgonThreads: c.argonThreads,
		},
		argonTimeMax:    c.argonTimeMax,
		argonMemoryMax:  c.argonMemoryMax,
		argonThreadsMax: c.argonThreadsMax,
	}

	must.Get(rand.Read(h.ArgonSalt[:]))

	switch c.mode {
	case ModeXChaCha20:
		must.Get(rand.Read(h.ChachaNonce[:]))
	case ModeAES256CTR:
		must.Get(rand.Read(h.AesIV[:]))
	}

	return h
}

func newHeaderForDecryptor(c *config) (h header) {
	h = newHeader(c)
	h.bin = bin{}
	return h
}

func (h header) writeTo(w io.Writer) error {
	err := h.check()
	if err != nil {
		return err
	}
	return binary.Write(w, binary.BigEndian, h.bin)
}

func (h *header) readFrom(r io.Reader) error {
	err := binary.Read(r, binary.BigEndian, &h.bin)
	if err != nil {
		return err
	}
	return h.check()
}

func (h header) check() error {
	if h.Mode <= modeBegin || h.Mode >= modeEnd {
		return fmt.Errorf(
			"%w: want %d < Mode < %d, got %d",
			ErrUnsupportedMode, modeBegin, modeEnd, h.Mode,
		)
	}
	if h.ArgonTime <= 0 || h.ArgonTime > h.argonTimeMax {
		return fmt.Errorf(
			"%w: want 0 < ArgonTime < %d, got %d",
			ErrHeaderParamsOutOfRange, h.argonTimeMax, h.ArgonTime,
		)
	}
	if h.ArgonMemory <= 0 || h.ArgonMemory > h.argonMemoryMax {
		return fmt.Errorf(
			"%w: want 0 < ArgonMemory < %d, got %d",
			ErrHeaderParamsOutOfRange, h.argonMemoryMax, h.ArgonMemory,
		)
	}
	if h.ArgonThreads <= 0 || h.ArgonThreads > h.argonThreadsMax {
		return fmt.Errorf(
			"%w: want 0 < ArgonThreads < %d, got %d",
			ErrHeaderParamsOutOfRange, h.argonThreadsMax, h.ArgonThreads,
		)
	}
	return nil
}

func (h header) keyLen() uint32 {
	switch h.Mode {
	case ModeXChaCha20:
		return chacha20.KeySize
	case ModeAES256CTR:
		return aesKeyLen
	default:
		panic(fmt.Sprintf("symmetric: unknown mode %d", h.Mode))
	}
}

func (h header) getStream(key []byte) cipher.Stream {
	switch h.Mode {
	case ModeXChaCha20:
		return must.Get(chacha20.NewUnauthenticatedCipher(key, h.ChachaNonce[:]))
	case ModeAES256CTR:
		block := must.Get(aes.NewCipher(key))
		return cipher.NewCTR(block, h.AesIV[:])
	default:
		panic(fmt.Sprintf("symmetric: unknown mode %d", h.Mode))
	}
}
