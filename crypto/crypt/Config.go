package crypt

import "hash"

type Crypt interface {
	Hash(data ...[]byte) []byte
	NewHash() hash.Hash
	Encrypt(data []byte, key []byte, iv []byte) (enData []byte, err error)
	Decrypt(enData []byte, key []byte, iv []byte) (data []byte, err error)
	GenKey() (priKey []byte, pubKey []byte, err error)
	Sign(data []byte, priKey []byte) (signature []byte, err error)
	Verify(data []byte, signature []byte, pubKey []byte) bool
	EncryptE(data []byte, pubKey []byte) (enData []byte, err error)
	DecryptE(enData []byte, priKey []byte) (data []byte, err error)
}

type CMCrypt struct {
}

type GMCrypt struct {
}
