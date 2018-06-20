package transaction

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"strings"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/hyperledger/fabric/protos/msp"
)

//计算字符串的SHA256哈希值，并且将哈希值转为Base64编码的字符串返回
func getShaBase64Str(str string) (string, error) {
	hash := sha256.New()
	if _, err := hash.Write([]byte(str)); err != nil {
		return "", err
	}

	hashByte := hash.Sum(nil)
	return base64.StdEncoding.EncodeToString(hashByte), nil
}

//解签名交易数据
func unsignEncryptData(decryptKey []byte, signedStr string) ([]byte, error) {
	bytesData, err := decryptBase64Str(signedStr)
	if err != nil {
		return nil, err
	}

	decryptedData, err := decryptData(decryptKey, bytesData)
	if err != nil {
		return nil, err
	}
	return decryptedData, nil
}

func decryptBase64Str(encryptStr string) ([]byte, error) {
	decryptedData, err := base64.StdEncoding.DecodeString(encryptStr)
	if err != nil {
		return nil, errors.New("base64 decode string failed:" + err.Error())
	}

	return decryptedData, nil
}

func decryptData(decryptKey []byte, encryptedData []byte) ([]byte, error) {
	//获取解密密钥
	block, _ := pem.Decode(decryptKey)
	if block == nil {
		return nil, errors.New("invalid decrypt key")
	}
	priv, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, errors.New("parse key failed:" + err.Error())
	}

	//解密
	data, err := rsa.DecryptPKCS1v15(rand.Reader, priv, encryptedData)
	if data == nil {
		return nil, errors.New("decrypt failed:" + err.Error())
	}

	return data, nil
}

func isEmptyStr(str string) bool {
	if str == nil {
		return true
	}

	return strings.Trim(str, " ") == ""
}

/*
	获取发起交易机构的MspId
*/
func getMspId(stub shim.ChaincodeStubInterface) (string, error) {
	crtOrg, err := getCrtOrg(stub)
	if err != nil {
		return "undefined", err
	}
	return crtOrg.MspId, nil
}

/*
	获取发起交易机构
*/
func getCrtOrg(stub shim.ChaincodeStubInterface) (*OrgAuth, error) {
	_, cert, err := getCreatorInfo(stub)
	if err != nil {
		return nil, err
	}

	resultsIterator, err := stub.GetStateByPartialCompositeKey(ObjectTypeOrgAuth, []string{})
	if err != nil {
		return nil, err
	}

	for i := 0; resultsIterator.HasNext(); i++ {
		response, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}
		key := response.Key
		val, err := stub.GetState(key)
		if err != nil {
			return nil, err
		}
		var orgAuth OrgAuth
		err = json.Unmarshal(val, &orgAuth)
		if err != nil {
			return nil, err
		}
		if orgAuth.Cert == cert {
			return &orgAuth, nil
		}
	}
	return nil, errors.New("机构未初始化")
}

/*
	获取发起交易机构的父机构mspId和发起交易机构的证书
*/
func getCreatorInfo(stub shim.ChaincodeStubInterface) (string, string, error) {

	creator, err := stub.GetCreator()
	if err != nil {
		return "undefined", "undefined", err
	}
	sId := &msp.SerializedIdentity{}
	err = proto.Unmarshal(creator, sId)
	if err != nil {
		return "undefined", "undefined", err
	}

	cert := string(sId.IdBytes)

	return sId.Mspid, cert, nil
}
