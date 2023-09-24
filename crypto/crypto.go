package crypto

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"github.com/ZZMarquis/gm/sm3"
	"github.com/api0-work/plugin"
	"github.com/api0-work/plugins/crypto/crypt"
	"github.com/ssgo/s"
	"hash"
)

var gmCrypto = crypt.GMCrypt{}
var cmCrypto = crypt.CMCrypt{}

func init() {
	plugin.Register(plugin.Plugin{
		Id:   "github.com/api0-work/plugins/crypto",
		Name: "crypto",
		Objects: s.Map{
			// base64 Base64编码
			"base64": s.Map{
				// encode 将字符串编码
				// encode data 用于编码的数据，string类型
				// encode return 编码后的结果
				"encode": func(data string) string {
					return base64.StdEncoding.EncodeToString([]byte(data))
				},
				// encodeBytes 将二进制数据编码
				// encodeBytes data 用于编码的数据，二进制的字节数组
				// encodeBytes return 编码后的结果
				"encodeBytes": func(data []byte) string {
					return base64.StdEncoding.EncodeToString(data)
				},
				// decode 解码为字符串，如果发生错误将抛出异常
				// decode data 编码后的结果
				// decode return 解码后的数据，string类型
				"decode": func(data string) (string, error) {
					return makeStringResult(base64.StdEncoding.DecodeString(data))
				},
				// decodeBytes 解码为二进制数据，如果发生错误将抛出异常
				// decodeBytes data 编码后的结果
				// decodeBytes return 解码后的数据，二进制的字节数组
				"decodeBytes": func(data string) ([]byte, error) {
					return base64.StdEncoding.DecodeString(data)
				},
			},
			// urlBase64 Base64编码（兼容URL）
			"urlBase64": s.Map{
				"encode": func(data string) string {
					return base64.URLEncoding.EncodeToString([]byte(data))
				},
				"encodeBytes": func(data []byte) string {
					return base64.URLEncoding.EncodeToString(data)
				},
				"decode": func(data string) (string, error) {
					return makeStringResult(base64.URLEncoding.DecodeString(data))
				},
				"decodeBytes": func(data string) ([]byte, error) {
					return base64.URLEncoding.DecodeString(data)
				},
			},
			// hex hex编码
			"hex": s.Map{
				"encode": func(data string) string {
					return hex.EncodeToString([]byte(data))
				},
				"encodeBytes": func(data []byte) string {
					return hex.EncodeToString(data)
				},
				"decode": func(data string) (string, error) {
					return makeStringResult(hex.DecodeString(data))
				},
				"decodeBytes": func(data string) ([]byte, error) {
					return hex.DecodeString(data)
				},
			},
			// md5 将字符串进行MD5编码
			// md5 data 用于编码的数据，string类型
			// md5 return 编码后的结果，16进制字符串
			"md5": func(data string) string {
				return hex.EncodeToString(makeHash(md5.New(), []byte(data)))
			},
			// md5 将二进制数据进行MD5编码
			// md5 data 用于编码的数据，用于编码的数据，二进制的字节数组
			// md5 return 编码后的结果，二进制的字节数组，如果需要转换为字符串可以使用hex.encode或者base64.encode等编码工具进行编码
			"md5Bytes": func(data []byte) []byte {
				return makeHash(md5.New(), data)
			},
			// sha1 将字符串进行SHA1编码
			// sha1 data 用于编码的数据，string类型
			// sha1 return 编码后的结果，16进制字符串
			"sha1": func(data string) string {
				return hex.EncodeToString(makeHash(sha1.New(), []byte(data)))
			},
			// sha1Bytes 将二进制数据进行SHA1编码
			// sha1Bytes data 用于编码的数据，用于编码的数据，二进制的字节数组
			// sha1Bytes return 编码后的结果，二进制的字节数组，如果需要转换为字符串可以使用hex.encode或者base64.encode等编码工具进行编码
			"sha1Bytes": func(data []byte) []byte {
				return makeHash(sha1.New(), data)
			},
			// sha256 将字符串进行SHA256编码
			// sha256 data 用于编码的数据，string类型
			// sha256 return 编码后的结果，16进制字符串
			"sha256": func(data string) string {
				return hex.EncodeToString(makeHash(sha256.New(), []byte(data)))
			},
			// sha256Bytes 将二进制数据进行SHA256编码
			// sha256Bytes data 用于编码的数据，用于编码的数据，二进制的字节数组
			// sha256Bytes return 编码后的结果，二进制的字节数组，如果需要转换为字符串可以使用hex.encode或者base64.encode等编码工具进行编码
			"sha256Bytes": func(data []byte) []byte {
				return makeHash(sha256.New(), data)
			},
			// sha512 将字符串进行SHA512编码
			// sha512 data 用于编码的数据，string类型
			// sha512 return 编码后的结果，16进制字符串
			"sha512": func(data string) string {
				return hex.EncodeToString(makeHash(sha512.New(), []byte(data)))
			},
			// sha512Bytes 将二进制数据进行SHA512编码
			// sha512Bytes data 用于编码的数据，用于编码的数据，二进制的字节数组
			// sha512Bytes return 编码后的结果，二进制的字节数组，如果需要转换为字符串可以使用hex.encode或者base64.encode等编码工具进行编码
			"sha512Bytes": func(data []byte) []byte {
				return makeHash(sha512.New(), data)
			},
			// sm3 将字符串进行SM3编码
			// sm3 data 用于编码的数据，string类型
			// sm3 return 编码后的结果，16进制字符串
			"sm3": func(data string) string {
				return hex.EncodeToString(makeHash(sm3.New(), []byte(data)))
			},
			// sm3Bytes 将二进制数据进行SM3编码
			// sm3Bytes data 用于编码的数据，用于编码的数据，二进制的字节数组
			// sm3Bytes return 编码后的结果，二进制的字节数组，如果需要转换为字符串可以使用hex.encode或者base64.encode等编码工具进行编码
			"sm3Bytes": func(data []byte) []byte {
				return makeHash(sm3.New(), data)
			},
			// aes 使用AES算法(CBC)进行加解密
			"aes": &Aes{crypto: &cmCrypto},
			// sm4 使用SM4算法进行加解密
			"sm4": &Aes{crypto: &gmCrypto},
			// ecdsa 使用ECDSA算法进行签名或加解密
			"ecdsa": &Ecdsa{crypto: &cmCrypto},
			// sm2 使用SM2算法进行签名或加解密
			"sm2": &Ecdsa{crypto: &gmCrypto},
		},
	})
}

type Aes struct {
	crypto crypt.Crypt
}

// Encrypt 将字符串加密
// Encrypt data 用于加密的数据，string类型
// Encrypt key hex编码的16位或32位密钥
// Encrypt iv hex编码的16位或32位向量
// Encrypt return 返回hex编码的加密结果
func (ae *Aes)Encrypt(data, key, iv string) (string, error) {
	if keyD, err1 := hex.DecodeString(key); err1 == nil {
		if ivD, err2 := hex.DecodeString(iv); err2 == nil {
			return makeHexStringResult(cmCrypto.Encrypt([]byte(data), keyD, ivD))
		} else {
			return "", err2
		}
	} else {
		return "", err1
	}
}
// EncryptBytes 将二进制数据加密
// EncryptBytes data 用于加密的数据，二进制的字节数组
// EncryptBytes key 二进制的16位或32位密钥
// EncryptBytes iv 二进制的16位或32位向量
// EncryptBytes return 返回二进制的加密结果
func (ae *Aes)EncryptBytes(data, key, iv []byte) ([]byte, error) {
	return cmCrypto.Encrypt(data, key, iv)
}
// Decrypt 解密为字符串
// Decrypt data hex编码的加密后结果
// Decrypt key hex编码的16位或32位密钥
// Decrypt iv hex编码的16位或32位向量
// Decrypt return 解密后的数据，string类型
func (ae *Aes)Decrypt(data string, key, iv string) (string, error) {
	if keyD, err1 := hex.DecodeString(key); err1 == nil {
		if ivD, err2 := hex.DecodeString(iv); err2 == nil {
			if dataD, err3 := hex.DecodeString(data); err3 == nil {
				return makeStringResult(cmCrypto.Decrypt(dataD, keyD, ivD))
			} else {
				return "", err3
			}
		} else {
			return "", err2
		}
	} else {
		return "", err1
	}
}

// DecryptBytes 解密为二进制数据
// DecryptBytes data 二进制的加密后结果
// DecryptBytes key 二进制的16位或32位密钥
// DecryptBytes iv 二进制的16位或32位向量
// DecryptBytes return 解密后的数据，二进制的字节数组
func (ae *Aes)DecryptBytes(data, key, iv []byte) ([]byte, error) {
	return cmCrypto.Decrypt(data, key, iv)
}


type Ecdsa struct {
	crypto crypt.Crypt
}

// GenKey 生成hex编码的公钥私钥
// GenKey priKey hex编码的私钥信息
// GenKey pubKey hex编码的公钥信息
func (ec *Ecdsa) GenKey() (priKey string, pubKey string, error error) {
	buf1, buf2, err := ec.crypto.GenKey()
	return hex.EncodeToString(buf1), hex.EncodeToString(buf2), err
}

// GenKeyBytes 生成二进制的公钥私钥
// GenKeyBytes priKey 二进制的私钥信息
// GenKeyBytes pubKey 二进制的公钥信息
func (ec *Ecdsa) GenKeyBytes() (priKey []byte, pubKey []byte, error error) {
	return ec.crypto.GenKey()
}

// Sign 对数据进行签名，返回hex编码格式的数据
// Sign data 将要签名的数据
// Sign priKey hex编码的私钥信息
func (ec *Ecdsa) Sign(data string, priKey string) (string, error) {
	if priKeyD, err2 := hex.DecodeString(priKey); err2 == nil {
		return makeHexStringResult(ec.crypto.Sign([]byte(data), priKeyD))
	} else {
		return "", err2
	}
}

// SignBytes 对数据进行签名，返回二进制的数据
// SignBytes data 将要签名的二进制数据
// SignBytes priKey 二进制的私钥信息
func (ec *Ecdsa) SignBytes(data, priKey []byte) ([]byte, error) {
	return ec.crypto.Sign(data, priKey)
}

// Verify 校验签名
// Verify data 将要签名的数据
// Verify signature hex编码的签名信息
// Verify pubKey hex编码的公钥信息
// Verify return 校验结果
func (ec *Ecdsa) Verify(data string, signature string, pubKey string) (bool, error) {
	if signatureD, err1 := hex.DecodeString(signature); err1 == nil {
		if pubKeyD, err2 := hex.DecodeString(pubKey); err2 == nil {
			return ec.crypto.Verify([]byte(data), signatureD, pubKeyD), nil
		} else {
			return false, err2
		}
	} else {
		return false, err1
	}
}

// VerifyBytes 校验签名
// VerifyBytes data 将要签名的二进制数据
// VerifyBytes signature 二进制的签名信息
// VerifyBytes pubKey 二进制的公钥信息
// VerifyBytes return 校验结果
func (ec *Ecdsa) VerifyBytes(data, signature, pubKey []byte) bool {
	return ec.crypto.Verify(data, signature, pubKey)
}

// Encrypt pubKey hex编码的公钥信息
func (ec *Ecdsa) Encrypt(data, pubKey string) (string, error) {
	if pubKeyD, err := hex.DecodeString(pubKey); err == nil {
		return makeHexStringResult(ec.crypto.EncryptE([]byte(data), pubKeyD))
	} else {
		return "", err
	}
}

// EncryptBytes pubKey 二进制的公钥信息
func (ec *Ecdsa) EncryptBytes(data, pubKey []byte) ([]byte, error) {
	return ec.crypto.EncryptE(data, pubKey)
}

// Decrypt priKey hex编码的私钥信息
func (ec *Ecdsa) Decrypt(data string, priKey string) (string, error) {
	if dataD, err1 := hex.DecodeString(data); err1 == nil {
		if priKeyD, err2 := hex.DecodeString(priKey); err2 == nil {
			return makeStringResult(ec.crypto.DecryptE(dataD, priKeyD))
		} else {
			return "", err2
		}
	} else {
		return "", err1
	}
}

// DecryptBytes priKey 二进制的私钥信息
func (ec *Ecdsa) DecryptBytes(data, priKey []byte) ([]byte, error) {
	return cmCrypto.DecryptE(data, priKey)
}

func makeHash(h hash.Hash, data []byte) []byte {
	h.Write(data)
	return h.Sum(nil)
}

func makeStringResult(buf []byte, err error) (string, error) {
	if err == nil {
		return string(buf), nil
	} else {
		return "", err
	}
}

func makeHexStringResult(buf []byte, err error) (string, error) {
	if err == nil {
		return hex.EncodeToString(buf), nil
	} else {
		return "", err
	}
}
