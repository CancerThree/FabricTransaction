package transaction

import (
	"encoding/json"
	"errors"

	"github.com/hyperledger/fabric/core/chaincode/shim"
)

//资产池地址-资产地址对应结构体，存储加密后资产池内的资产地址
type PoolAsset struct {
	PoolID         string `json:"poolId"`         //资产池地址
	EncryptAssetID string `json:"encryptAssetId"` //加密后资产地址
	HasSpent       string `json:"hasSpent"`       //资产有效标识
	TypeID         string `json:"typeId"`         //资产类型ID
	ObjectType     string `json:"objectType"`
}

//资产结构体，存储资产
type Asset struct {
	Addr       string  `json:"addr,omitempty"`       //资产地址，使用GUID，GUID由外部传入
	Value      float64 `json:"value"`                //资产值
	TypeID     string  `json:"typeId,omitempty"`     //资产类型ID，一般对应产品证券ID
	AttachHash string  `json:"attachHash,omitempty"` //MSPID+Addr的哈希值，用于验证资产归属
	HasSpent   string  `json:"hasSpent,omitempty"`   //资产有效标识
	ObjectType string  `json:"objectType"`
}

// 判断asset是否可进行转让
func (asset *Asset) CanBeTransfer(accountId string) (bool, error) {
	if asset.Value <= 0.0 {
		return false, errors.New("invalid asset, value lte 0.0")
	}

	if asset.HasSpent == ASSET_HAS_SPENT {
		return false, nil
	}

	hashSign, err := getShaBase64Str(accountId + asset.Addr)
	if err != nil {
		return false, errors.New("calc asset's hash failed: " + err.Error())
	}

	if hashSign != asset.AttachHash {
		return false, errors.New("asset's hash cannot match")
	}

	return true, nil
}

//将asset按照Value值从小到大排序，并且返回asset的Value总和
func sortAndCountByAmount(assets *[]Asset) float64 {
	if assets == nil {
		return 0
	}

	sum := 0.0

	assetArray := *assets
	for i := 0; i < len(assetArray); i++ {
		for j := len(assetArray) - 1; j > i; j-- {
			if assetArray[i].Value > assetArray[j].Value {
				tempAsset := assetArray[i]
				assetArray[i] = assetArray[j]
				assetArray[j] = tempAsset
			}
		}
		if assetArray[i].HasSpent == ASSET_HAS_NOT_SPENT {
			sum += assetArray[i].Value
		}
	}
	return sum
}

//查询地址下的所有资产
func GetAssetsByAddrs(stub shim.ChaincodeStubInterface, addrs []string) ([]Asset, error) {
	// assets := make([]Asset, len(addrs))
	var assets []Asset

	for i := 0; i < len(addrs); i++ {
		var asset Asset
		val, err := stub.GetState(addrs[i])
		if err != nil {
			return nil, errors.New("get asset failed:" + addrs[i])
		}

		err = json.Unmarshal(val, &asset)
		if err != nil {
			return nil, errors.New("unmarshal asset failed:" + addrs[i])
		}

		assets = append(assets, asset)
	}

	return assets, nil
}

func addAsset(stub shim.ChaincodeStubInterface, accId string, assetAddr string, amount float64, typeId string) error {
	// val, err := stub.GetState(accId)
	// if err != nil {
	// 	return errors.New("accountId Invalid:" + accId)
	// }
	val, err := stub.GetState(assetAddr)
	if err != nil {
		return errors.New("assetAddr Invalid:" + assetAddr)
	}
	if val != nil {
		return errors.New("asset already exists.")
	}

	if amount <= 0 {
		return errors.New("asset amount invalid")
	}

	hashSign, err := getShaBase64Str(accId + assetAddr)
	if err != nil {
		return err
	}

	asset := Asset{
		Addr:       assetAddr,
		Value:      amount,
		TypeID:     typeId,
		AttachHash: hashSign,
		HasSpent:   ASSET_HAS_NOT_SPENT}
	assetBytes, err := json.Marshal(asset)
	if err != nil {
		return errors.New("marshal new asset failed:" + err.Error())
	}

	stub.PutState(assetAddr, assetBytes)
	return nil
}
