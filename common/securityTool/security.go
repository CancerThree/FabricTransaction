package securityTool

import (
	"crypto/sha256"
	"encoding/base64"
)

type SecurityTool interface {
	ParsePublicKey(publicKey string) (interface{}, error)
	EncryptByPoolPublicKey(publicKey []byte, data []byte) (string, error)
	VerifySignStrByPoolPublicKey(data []byte, signature, publicKey string) (bool, error)
	VerifyTxByPoolPublicKey(publicKey string, obj interface{}) (bool, error)
}

func CalcSHA256Base64Str(str string) (string, error) {
	hash := sha256.New()
	if _, err := hash.Write([]byte(str)); err != nil {
		return "", err
	}

	hashByte := hash.Sum(nil)
	return base64.StdEncoding.EncodeToString(hashByte), nil
}
