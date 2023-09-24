package crypt

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"github.com/emmansun/gmsm/sm2"
	"github.com/obscuren/ecies"
	"github.com/ssgo/u"
	"math/big"
)

func makePriKey(priKey []byte, curve elliptic.Curve) *ecdsa.PrivateKey {
	x, y := curve.ScalarBaseMult(priKey)
	return &ecdsa.PrivateKey{
		D: new(big.Int).SetBytes(priKey),
		PublicKey: ecdsa.PublicKey{
			Curve: curve,
			X:     x,
			Y:     y,
		},
	}
}

func makePubKey(pubKey []byte, curve elliptic.Curve) *ecdsa.PublicKey {
	keyLen := pubKey[0]
	x := new(big.Int)
	y := new(big.Int)
	x.SetBytes(pubKey[1 : keyLen+1])
	y.SetBytes(pubKey[keyLen+1:])
	return &ecdsa.PublicKey{Curve: curve, X: x, Y: y}
}

func (c *CMCrypt) GenKey() (priKey []byte, pubKey []byte, err error) {
	pri, err := ecdsa.GenerateKey(elliptic.P256(), u.GlobalRand2)
	if err != nil {
		return nil, nil, err
	}
	var buf bytes.Buffer
	buf.WriteByte(byte(len(pri.X.Bytes())))
	buf.Write(pri.X.Bytes())
	buf.Write(pri.Y.Bytes())
	return pri.D.Bytes(), buf.Bytes(), nil
}

func (c *CMCrypt) Sign(data []byte, priKey []byte) (signature []byte, err error) {
	r, s, err := ecdsa.Sign(u.GlobalRand1, makePriKey(priKey, elliptic.P256()), u.Sha256(data))
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	buf.WriteByte(byte(len(r.Bytes())))
	buf.Write(r.Bytes())
	buf.Write(s.Bytes())
	return buf.Bytes(), nil
}

func (c *CMCrypt) Verify(data []byte, signature []byte, pubKey []byte) bool {
	byteLen := signature[0]
	r := new(big.Int)
	s := new(big.Int)
	r.SetBytes(signature[1 : byteLen+1])
	s.SetBytes(signature[byteLen+1:])
	return ecdsa.Verify(makePubKey(pubKey, elliptic.P256()), u.Sha256(data), r, s)
}

func (c *CMCrypt) EncryptE(data []byte, pubKey []byte) (enData []byte, err error) {
	pub := ecies.ImportECDSAPublic(makePubKey(pubKey, elliptic.P256()))
	if r, err := ecies.Encrypt(u.GlobalRand1, pub, data, nil, nil); err == nil {
		return r, nil
	} else {
		return nil, err
	}
}

func (c *CMCrypt) DecryptE(enData []byte, priKey []byte) (data []byte, err error) {
	pri := ecies.ImportECDSA(makePriKey(priKey, elliptic.P256()))
	if r, err := pri.Decrypt(u.GlobalRand1, enData, nil, nil); err == nil {
		return r, nil
	} else {
		return nil, err
	}
}

// SM2

func (c *GMCrypt) GenKey() (priKey []byte, pubKey []byte, err error) {
	pri, err := sm2.GenerateKey(u.GlobalRand2)
	if err != nil {
		return nil, nil, err
	}

	var buf bytes.Buffer
	buf.WriteByte(byte(len(pri.X.Bytes())))
	buf.Write(pri.X.Bytes())
	buf.Write(pri.Y.Bytes())
	return pri.D.Bytes(), buf.Bytes(), nil
}

func (c *GMCrypt) Sign(data []byte, priKey []byte) (signature []byte, err error) {
	r, s, err := sm2.SignWithSM2(u.GlobalRand1, makePriKey(priKey, sm2.P256()), nil, data)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	buf.WriteByte(byte(len(r.Bytes())))
	buf.Write(r.Bytes())
	buf.Write(s.Bytes())
	return buf.Bytes(), nil
}

func (c *GMCrypt) Verify(data []byte, signature []byte, pubKey []byte) bool {
	byteLen := signature[0]
	r := new(big.Int)
	s := new(big.Int)
	r.SetBytes(signature[1 : byteLen+1])
	s.SetBytes(signature[byteLen+1:])
	return sm2.VerifyWithSM2(makePubKey(pubKey, sm2.P256()), nil, data, r, s)
}

func (c *GMCrypt) EncryptE(data []byte, pubKey []byte) (enData []byte, err error) {
	if r, err := sm2.Encrypt(u.GlobalRand1, makePubKey(pubKey, sm2.P256()), data, nil); err == nil {
		return r, nil
	} else {
		return nil, err
	}
}

func (c *GMCrypt) DecryptE(enData []byte, priKey []byte) (data []byte, err error) {
	if r, err := sm2.Decrypt(&sm2.PrivateKey{PrivateKey: *makePriKey(priKey, sm2.P256())}, enData); err == nil {
		return r, nil
	} else {
		return nil, err
	}
}

//func (c *GMCrypt) GenKey() (priKey []byte, pubKey []byte, err error) {
//	pri, pub, err := sm2.GenerateKey(u.GlobalRand2)
//	if err != nil {
//		return nil, nil, err
//	}
//	return pri.GetRawBytes(), pub.GetRawBytes(), nil
//}
//
//func (c *GMCrypt) Sign(data []byte, priKey []byte) (signature []byte, err error) {
//	pri, err := sm2.RawBytesToPrivateKey(priKey)
//	if err != nil {
//		return nil, err
//	}
//	signature, err = sm2.Sign(pri, nil, u.Sha256(data))
//	if err != nil {
//		return nil, err
//	}
//	return signature, nil
//}
//
//func (c *GMCrypt) Verify(data []byte, signature []byte, pubKey []byte) bool {
//	pub, err := sm2.RawBytesToPublicKey(pubKey)
//	if err != nil {
//		return false
//	}
//	return sm2.Verify(pub, nil, u.Sha256(data), signature)
//}
