package securityTool

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
)

type RSATool struct {
}

//转换公钥
func (rsaTool RSATool) ParsePublicKey(publicKey string) (interface{}, error) {
	publicKey = "-----BEGIN PUBLIC KEY-----\n" + publicKey + "\n" + "-----END PUBLIC KEY-----"
	block, _ := pem.Decode([]byte(publicKey))
	if block == nil {
		return nil, errors.New("公钥解析出错")
	}
	cert, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	key := cert.(*rsa.PublicKey)
	return key, nil
}

func (rsaTool RSATool) EncryptByPoolPublicKey(publicKey []byte, data []byte) (string, error) {
	key, err := RSATool.ParsePublicKey(RSATool{}, string(publicKey))
	if err != nil {
		return "", err
	}
	pubKey := key.(*rsa.PublicKey)
	encryptedData, err := rsa.EncryptPKCS1v15(rand.Reader, pubKey, data)
	if err != nil {
		return "", errors.New("encrypt failed:" + err.Error())
	}
	dataStr := base64.StdEncoding.EncodeToString(encryptedData)
	return dataStr, nil
}

func (rsaTool RSATool) VerifySignByPoolPublicKey(data []byte, signature, publicKey string) (bool, error) {
	//公钥
	pub, err := RSATool.ParsePublicKey(RSATool{}, publicKey)
	if err != nil {
		return false, err
	}
	pubKey := pub.(*rsa.PublicKey)

	sign, err := base64.StdEncoding.DecodeString(signature)
	if err != nil {
		return false, err
	}
	//计算hash
	h := sha256.New()
	h.Write(data)
	hashByte := h.Sum(nil)
	//验签
	err = rsa.VerifyPKCS1v15(pubKey, crypto.SHA256, hashByte, sign)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (rsaTool RSATool) VerifyTxByPoolPublicKey(publicKey string, obj interface{}) (bool, error) {
	return false, nil
}
