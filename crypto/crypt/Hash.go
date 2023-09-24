package crypt

import (
	"crypto/sha512"
	"github.com/ZZMarquis/gm/sm3"
	"hash"
)

func (c *CMCrypt) NewHash() hash.Hash {
	return sha512.New()
}

func (c *GMCrypt) NewHash() hash.Hash {
	return sm3.New()
}

func (c *CMCrypt) Hash(data ...[]byte) []byte {
	return makeHash(c.NewHash(), data...)
}

func (c *GMCrypt) Hash(data ...[]byte) []byte {
	return makeHash(c.NewHash(), data...)
}

func makeHash(h hash.Hash, data ...[]byte) []byte {
	for _, b := range data {
		h.Write(b)
	}
	return h.Sum(nil)
}