package transaction

import (
	"encoding/json"
	"errors"

	"github.com/hyperledger/fabric/core/chaincode/shim"
)

type AssetPoolRequest struct {
	AssetpoolId   string `json:"assetpoolId"`    //资产库ID 主键
	AssetpoolType string `json:"assetpoolType"`  //资产库类型
	OrgID         string `json:"assetpoolOwner"` //资产库所属机构
	EncryptKey    string `json:"publicKey"`      //加密用-publicKey
	UnsignKey     string `json:"signPublicKey"`  //验签用-生成的privateKey

	ReqOrgId string `json:"reqOrgId,omitempty"` //发送机构
	ReqSign  string `json:"reqSign,omitempty"`  //发送机构签名
}

type AssetPool struct {
	AssetpoolId   string `json:"assetpoolId"`
	AssetpoolType string `json:"assetpoolType"` //资产库类型
	UnsignKey     string `json:"signPublicKey"` //验签用-生成的privateKey
	EncryptKey    string `json:"publicKey"`     //加密用-publicKey
	AttachHash    string `json:"attachHash"`
	OrgID         string `json:"assetpoolOwner"`
	ObjectType    string `json:"objectType"`
	// MspId          string `json:"mspId,omitempty"`          //所属机构 维护机构
}

func AddAssetPoolByJsonStr(stub shim.ChaincodeStubInterface, str string) error {
	// if len(args) != 1 {
	// 	return errors.New("invalid argument: only need 1 argument")
	// }
	// return errors.New(str)
	var assetPoolReq AssetPoolRequest
	err := json.Unmarshal([]byte(str), &assetPoolReq)
	if err != nil {
		return errors.New("unmarshal assetPool request str failed:" + err.Error())
	}

	return InitAssetPool(stub, assetPoolReq)
}

func InitAssetPool(stub shim.ChaincodeStubInterface, assetPoolReq AssetPoolRequest) error {
	if err := assetPoolReq.verifyField(); err != nil {
		return err
	}

	// unsignKey := assetPoolReq.UnsignKey
	org, err := GetOrganizationByKey(stub, assetPoolReq.OrgID)
	if err != nil {
		return errors.New("get org failed:" + err.Error())
	}
	err = CheckJSONObjectSignature(&assetPoolReq, org.SignPublicKey)
	if err != nil {
		return errors.New("verify sign failed:" + err.Error())
	}

	assetPool := AssetPool{
		AssetpoolId:   assetPoolReq.AssetpoolId,
		AssetpoolType: assetPoolReq.AssetpoolType,
		OrgID:         assetPoolReq.OrgID,
		UnsignKey:     assetPoolReq.UnsignKey,
		EncryptKey:    assetPoolReq.EncryptKey,
		ObjectType:    OBJECT_TYPE_ASEETPOOL}

	hash, err := getShaBase64Str(assetPool.OrgID + assetPool.AssetpoolId)
	if err != nil {
		return errors.New("calc hash failed:" + err.Error())
	}
	assetPool.AttachHash = hash

	assetPoolByte, _ := json.Marshal(assetPool)

	key, err := stub.CreateCompositeKey(OBJECT_TYPE_ASEETPOOL, []string{assetPool.AssetpoolId})
	if err != nil {
		return errors.New(err.Error())
	}

	err = stub.PutState(key, assetPoolByte)
	if err != nil {
		return err
	}

	return nil
}

func verifyAssetPoolOfOrg(orgId string, assetPool AssetPool) error {
	hash, err := getShaBase64Str(orgId + assetPool.AssetpoolId)
	if err != nil {
		return errors.New("calc hash failed:" + err.Error())
	}
	if assetPool.AttachHash != hash {
		return errors.New("hash verified failed:" + orgId + "," + assetPool.AssetpoolId)
	}

	return nil
}

func GetAssetPoolById(stub shim.ChaincodeStubInterface, id string) (*AssetPool, error) {
	var assetPool AssetPool
	if isEmptyStr(id) {
		return nil, errors.New("addr cannot be empty")
	}

	key, err := stub.CreateCompositeKey(OBJECT_TYPE_ASEETPOOL, []string{id})
	if err != nil {
		return nil, err
	}

	val, err := stub.GetState(key)
	if err != nil {
		return nil, errors.New("query assetPool failed:" + key + ",detail:" + err.Error())
	}
	if val == nil {
		return nil, errors.New("assetPool cannot find:" + key)
	}
	err = json.Unmarshal(val, &assetPool)
	if err != nil {
		return nil, errors.New("unmarshal assetPool failed:" + err.Error())
	}
	return &assetPool, nil
}

func addAssetToPool(stub shim.ChaincodeStubInterface, assetPool AssetPool, assetAddr string, amount float64, typeId string) error {
	err := addAsset(stub, assetPool.AssetpoolId, assetAddr, amount, typeId)
	if err != nil {
		return err
	}

	encryptedData, err := encryptData([]byte(assetPool.EncryptKey), []byte(assetAddr))
	encryptedStr, err := getShaBase64Str(string(encryptedData))
	accountAsset := PoolAsset{
		PoolID:         assetPool.AssetpoolId,
		EncryptAssetID: encryptedStr,
		HasSpent:       ASSET_HAS_NOT_SPENT,
		TypeID:         typeId,
		ObjectType:     OBJECT_TYPE_ASSET}

	key, err := stub.CreateCompositeKey(OBJECT_TYPE_ASSET, []string{accountAsset.PoolID, encryptedStr})
	if err != nil {
		return errors.New("create PoolAsset Key failed:" + err.Error())
	}
	val, err := stub.GetState(key)
	if err != nil {
		return errors.New("get PoolAsset failed:" + err.Error())
	}
	if val != nil {
		return errors.New("PoolAsset already exist")
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

func transferAssets(stub shim.ChaincodeStubInterface, assets []Asset, tx Transaction, fromPool AssetPool, toPool AssetPool) error {
	transferAmount := tx.Amount

	// 从小到大修改asset
	for i := 0; i < len(assets); i++ {
		if transferAmount <= 0 {
			break
		}

		if canTrans, err := assets[i].CanBeTransfer(tx.FromPool); !canTrans {
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

	//新建Asset作为转让以及“找零”的资产
	if err := addAssetToPool(stub, toPool, tx.AssetAddrs[0], tx.Amount, tx.AssetTypeID); err != nil {
		return err
	}
	if transferAmount < 0 {
		if err := addAssetToPool(stub, fromPool, tx.AssetAddrs[1], 0-transferAmount, tx.AssetTypeID); err != nil {
			return err
		}
	}

	//填log
	err := addOrgPrivateLog(stub, tx, fromPool, TX_TYPE_TRANSFER_OUT)
	if err != nil {
		return err
	}

	err = addOrgPrivateLog(stub, tx, toPool, TX_TYPE_TRANSFER_IN)
	if err != nil {
		return err
	}

	err = addChainLog(stub, tx, TX_TYPE_TRANSFER)
	if err != nil {
		return err
	}

	return nil
}

func (request *AssetPoolRequest) verifyField() error {
	if isEmptyStr(request.AssetpoolId) {
		return errors.New("AssetpoolId为空")
	}
	if isEmptyStr(request.AssetpoolType) {
		return errors.New("AssetpoolType为空")
	}
	if isEmptyStr(request.OrgID) {
		return errors.New("AssetpoolOwner为空")
	}
	if isEmptyStr(request.EncryptKey) {
		return errors.New("EncryptKey为空")
	}
	if isEmptyStr(request.UnsignKey) {
		return errors.New("UnsignKey为空")
	}
	if isEmptyStr(request.ReqOrgId) {
		return errors.New("ReqOrgId为空")
	}
	if isEmptyStr(request.ReqSign) {
		return errors.New("ReqSign为空")
	}
	return nil
}

// func (acc *AssetPool) getUnsignKey() string {
// 	return acc.UnsignKey
// }
