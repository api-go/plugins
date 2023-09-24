package crypt

import (
	"errors"
	"github.com/ZZMarquis/gm/sm4"
	"github.com/ZZMarquis/gm/util"
	"github.com/ssgo/u"
)

func (c *CMCrypt) Encrypt(data []byte, key []byte, iv []byte) (enData []byte, err error) {
	return u.EncryptAesBytes(data, key, iv)
}

func (c *CMCrypt) Decrypt(enData []byte, key []byte, iv []byte) (data []byte, err error) {
	return u.DecryptAesBytes(enData, key, iv)
}

func (c *GMCrypt) Encrypt(data []byte, key []byte, iv []byte) (enData []byte, err error) {
	return sm4.CBCEncrypt(key[0:16], iv[0:16], util.PKCS5Padding(data, 16))
}

func (c *GMCrypt) Decrypt(enData []byte, key []byte, iv []byte) (data []byte, err error) {
	data, err = sm4.CBCDecrypt(key[0:16], iv[0:16], enData)
	if err == nil {
		length := len(data)
		if length > 0 {
			unpadding := int(data[length-1])
			if length-unpadding >= 0 {
				data = util.PKCS5UnPadding(data)
			} else {
				return nil, errors.New("failed to decrypt by gm4")
			}
		} else {
			return nil, errors.New("failed to decrypt by gm4")
		}
	}
	return data, err
}
