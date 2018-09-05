package common

import (
	"errors"
	"log"
	"strings"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/hyperledger/fabric/protos/msp"
)

func IsEmptyStr(str string) bool {
	return strings.Trim(str, " ") == ""
}

func GetMspID(stub shim.ChaincodeStubInterface) (string, error) {
	mspId, _, err := getCreatorInfo(stub)
	if err != nil {
		return "", errors.New("get mspid failed:" + err.Error())
	}
	return mspId, nil
}

func CheckExistByKey(stub shim.ChaincodeStubInterface, objType string, keys []string) (bool, string, []byte, error) {
	key, err := stub.CreateCompositeKey(objType, keys)
	if err != nil {
		log.Printf("[CheckExistByKey-createKey] objType=%s, keys=%s\n", objType, keys)
		return false, "", nil, err
	}
	val, err := stub.GetState(key)
	if err != nil {
		log.Printf("[CheckExistByKey-getstate] objType=%s, keys=%s\n", objType, keys)
		return false, key, nil, err
	}

	if val == nil {
		return false, key, nil, nil
	}
	return true, key, val, nil
}

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
