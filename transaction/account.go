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
	ObjectType string `json:"objectType"`
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

func addAssetUnderAccount(stub shim.ChaincodeStubInterface, acc Account, assetAddr string, amount float64, typeId string) error {
	err := addAsset(stub, acc.Addr, assetAddr, amount, typeId)
	if err != nil {
		return err
	}

	encryptedData, err := encryptData([]byte(acc.EncryptKey), []byte(assetAddr))
	encryptedStr, err := getShaBase64Str(string(encryptedData))
	accountAsset := AccountsAsset{
		AccountID:      acc.Addr,
		EncryptAssetID: encryptedStr,
		HasSpent:       ASSET_HAS_NOT_SPENT,
		TypeID:         typeId,
		ObjectType:		OBJECT_TYPE_ASSET}

	key, err := stub.CreateCompositeKey(OBJECT_TYPE_ASSET, []string{acc.Addr, encryptedStr})
	if err != nil {
		return errors.New("create accoutAsset Key failed:" + err.Error())
	}
	val, err := stub.GetState(key)
	if err != nil {
		return errors.New("get accoutAsset failed:" + err.Error())
	}
	if val != nil {
		return errors.New("asset already exist")
	}

	bytes, err := json.Marshal(accountAsset)
	if err != nil {
		return errors.New("marshal accountAsset failed:" + err.Error())
	}
	if err = stub.PutState(key, bytes); err != nil {
		return errors.New("sotre accountAsset failed:" + err.Error())
	}
	return nil
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

		[i].HasSpent = "Y"
		transferAmount -= assets[i].Value

		assetBytes, err := json.Marshal(assets[i])
		if err != nil {
			return errors.New("marshal asset failed:" + err.Error())
		}

		stub.PutState(assets[i].Addr, assetBytes)
	}

	//新建Asset作为转让以及“找零”的资产
	if err := addAssetUnderAccount(stub, toAcc, tx.AssetAddrs[0], tx.Amount, tx.AssetTypeID); err != nil {
		return err
	}
	if transferAmount < 0 {
		if err := addAssetUnderAccount(stub, fromAcc, tx.AssetAddrs[1], 0-transferAmount, tx.AssetTypeID); err != nil {
			return err
		}
	}

	//填log
	err = addOrgPrivateLog(stub, tx, fromAcc, TX_TYPE_TRANSFER_OUT)
	if err != nil {
		return err
	}

	err = addOrgPrivateLog(stub, tx, toAcc, TX_TYPE_TRANSFER_IN)
	if err != nil {
		return err
	}

	err = addChainLog(stub, tx, TX_TYPE_TRANSFER)
	if err != nil {
		return err
	}

	return nil
}

// func (acc *Account) getUnsignKey() string {
// 	return acc.UnsignKey
// }
