package transaction

import (
	"encoding/json"
	"errors"

	"github.com/hyperledger/fabric/core/chaincode/shim"
)

type Account struct {
	Addr       string `json:"addr"`
	UnsignKey  string `json:"unsignKey"`  //验签用-生成的privateKey
	EncryptKey string `json:"encryptKey"` //加密用-publicKey
	AttachHash string `json:"attachHash"`
	OrgID      string `json:"orgId"`
}

func InitAccount(stub shim.ChaincodeStubInterface, args []string) error {
	var account Account

	if len(args) != 1 {
		return errors.New("invalid argument: only need 1 argument")
	}

	//TODO: add org unsign data

	err := json.Unmarshal([]byte(args[0]), &account)
	if err != nil {
		return errors.New("unmarshal account failed")
	}

	if isEmptyStr(account.Addr) || isEmptyStr(account.UnsignKey) ||
		isEmptyStr(account.EncryptKey) {
		return errors.New("account has empty field")
	}

	hash, err := getShaBase64Str(account.OrgID + account.Addr)
	if err != nil {
		return errors.New("calc hash failed:" + err.Error())
	}
	account.AttachHash = hash

	accountByte, _ := json.Marshal(account)
	err = stub.PutState(account.Addr, accountByte)
	if err != nil {
		return err
	}

	return nil
}

func addAsset(stub shim.ChaincodeStubInterface, acc Account, assetAddr string, amount float64, typeId string) error {
	err := sotreAsset(stub, acc.Addr, assetAddr, amount, typeId)
	if err != nil {
		return err
	}

	encryptedData, err := encryptData([]byte(acc.EncryptKey), []byte(assetAddr))
	encryptedStr, err := getShaBase64Str(string(encryptedData))
	accountAsset := AccountsAsset{
		AccountID:      acc.Addr,
		EncryptAssetID: encryptedStr,
		HasSpent:       ASSET_HAS_NOT_SPENT,
		TypeID:         typeId}

	key, err := stub.CreateCompositeKey()
}

func transferAssets(stub shim.ChaincodeStubInterface, assets []Asset, tx Transaction, fromAcc Account, toAcc Account) error {
	transferAmount := tx.Amount

	// 从小到大修改asset
	for i := 0; i < len(assets); i++ {
		if transferAmount <= 0 {
			break
		}

		if canTrans, err := assets[i].CanBeTransfer(tx.FromAccount); !canTrans {
			if err == nil {
				continue
			} else {
				return err
			}
		}

		assets[i].HasSpent = "Y"
		transferAmount -= assets[i].Value

		assetBytes, err := json.Marshal(assets[i])
		if err != nil {
			return errors.New("marshal asset failed:" + err.Error())
		}

		stub.PutState(assets[i].Addr, assetBytes)
	}
}

// func (acc *Account) getUnsignKey() string {
// 	return acc.UnsignKey
// }
