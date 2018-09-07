package securityTool

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"regexp"
	"strings"
)

type SecurityTool interface {
	ParsePublicKey(publicKey string) (interface{}, error)
	EncryptByPoolPublicKey(publicKey []byte, data []byte) (string, error)
	VerifySignByPoolPublicKey(data []byte, signature, publicKey string) (bool, error)
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

//验证签名，默认签名字段为ReqSign
func CheckJSONObjectSignatureString(requestString, publicKey string) (bool, error) {
	reg, _ := regexp.Compile(",\"sign\":\"(.+)\"")

	match := reg.MatchString(requestString)
	if match {
		signString := reg.FindString(requestString)

		requestString = strings.Replace(requestString, signString, "", -1)
		signString = strings.TrimRight(signString, "\"")
		signString = strings.TrimLeft(signString, ",\"sign\":\"")
		return SecurityTool.VerifySignByPoolPublicKey(RSATool{}, []byte(requestString), signString, publicKey)
	}
	return false, fmt.Errorf("对象签名属性名[sign]无效")
}
