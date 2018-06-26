package transaction

import (
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"reflect"
)

//验证签名，默认签名字段为ReqSign
func CheckJSONObjectSignature(obj interface{}, publicKey string) error {
	refV := reflect.ValueOf(obj).Elem()
	fieldV := refV.FieldByName("ReqSign")
	if fieldV.IsValid() {
		sign := fieldV.String()
		fieldV.SetString("")
		data, err := json.Marshal(obj)
		if err != nil {
			return fmt.Errorf("检查签名时序列化对象出错[%s]", err.Error())
		}
		// fieldV.SetString(sign)
		return CheckSignature(data, sign, publicKey)
	}
	return fmt.Errorf("对象签名属性名[ReqSign]无效")
}

func CheckSignature(data []byte, signature, publicKey string) error {
	//公钥
	pub, err := ParsePublicKey(publicKey)
	if err != nil {
		return err
	}

	sign, err := base64.StdEncoding.DecodeString(signature)
	if err != nil {
		return err
	}
	//计算hash
	h := sha256.New()
	h.Write(data)
	hashByte := h.Sum(nil)
	//验签
	return rsa.VerifyPKCS1v15(pub, crypto.SHA256, hashByte, sign)
}

//转换公钥
func ParsePublicKey(publicKey string) (key *rsa.PublicKey, err error) {
	publicKey = "-----BEGIN PUBLIC KEY-----\n" + publicKey + "\n" + "-----END PUBLIC KEY-----"
	block, _ := pem.Decode([]byte(publicKey))
	if block == nil {
		return nil, errors.New("公钥解析出错")
	}
	cert, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	key = cert.(*rsa.PublicKey)
	return key, nil
}

//转换私钥
func ParsePrivateKey(privateKey string) (key *rsa.PrivateKey, err error) {
	privateKey = "-----BEGIN RSA PRIVATE KEY-----\n" + privateKey + "\n" + "-----END RSA PRIVATE KEY-----"
	block, _ := pem.Decode([]byte(privateKey))
	if block == nil {
		return nil, errors.New("私钥解析出错")
	}
	privt, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	key = privt.(*rsa.PrivateKey)
	return key, nil
}
