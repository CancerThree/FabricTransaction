package assetPool

import (
	"encoding/json"
	"errors"
	"log"

	"github.com/FabricTransaction/transaction/asset"
	"github.com/FabricTransaction/transaction/common"
	"github.com/hyperledger/fabric/core/chaincode/shim"
)

type AssetPool struct {
	AssetPoolAddr string `json:"assetPoolAddr"`
	AssetPoolType string `json:"assetPoolType"`
	PublicKey     string `json:"publicKey"`
}

type AssetPoolInterface interface {
	Transfer(stub shim.ChaincodeStubInterface, _to AssetPool, _value float64) (bool, error)

	Approve(stub shim.ChaincodeStubInterface, _spender string, _value float64) (bool, error)

	Allowance(stub shim.ChaincodeStubInterface, _owner string)

	SetTransferEvent(stub shim.ChaincodeStubInterface, _from string, _to string, _value float64) error

	SetApprovalEvent(stub shim.ChaincodeStubInterface, _to string, _value float64) error
}

func (pool *AssetPool) Transfer(stub shim.ChaincodeStubInterface, assetType string, _to AssetPool, _value float64) (bool, error) {
	priData, err := stub.GetTransient()
	if err != nil {
		return false, err
	}

	addrsBytes, ok := priData["addrs"]
	if !ok {
		log.Println("get addrs failed, transient is:")
		log.Println(priData)
		return false, errors.New("cannot get addrs data")
	}

	var decryptAddrs []string
	if err = json.Unmarshal(addrsBytes, &decryptAddrs); err != nil {
		log.Printf("unmarshal addrs failed, bytes is: %s", string(addrsBytes))
		return false, err
	}
	assets, balance, err := asset.GetSortedAssetsByAddrs(stub, decryptAddrs)
	if err != nil {
		return false, err
	}
	if balance < _value {
		return false, errors.New("poor balance")
	}

	change, err := pool.BurnAssets(stub, assetType, *assets, _value, common.TX_TYPE_TRANSFER)

	return true, nil
}

func (pool *AssetPool) BurnAssets(stub shim.ChaincodeStubInterface, assetType string, assets []asset.Asset, _value float64, burnType string) (*asset.Asset, error) {
	priData, err := stub.GetTransient()
	if err != nil {
		return nil, err
	}

	if len(assets) < 1 {
		log.Println()
		return nil, nil
	}

	for i, v := range assets {
		if !v.CanTransfer() {
			continue
		}
		if _value <= 0 {
			break
		}
		_value -= v.Value
		v.HasTransfered = true
		v.AddLogInfo()

		bytes, err := json.Marshal(v)
		if err != nil {
			return nil, err
		}

		key, err := stub.CreateCompositeKey(common.OBJECT_TYPE_ASSET, []string{v.AssetAddr})
		if err != nil {
			return nil, err
		}
		err = stub.PutState(key, bytes)
		if err != nil {
			return nil, err
		}
	}
	if _value > 0 {
		return nil, errors.New("poor balance")
	}

	if _value < 0 {
		changeAssetAddr, ok := priData["changeAddr"]
		if !ok {
			return nil, errors.New("no changeAddr data")
		}
		//TODO check addr exists
		changeAssetSign, ok := priData["changeAssetSign"]
		if !ok {
			return nil, errors.New("no changeAssetSign")
		}
		changeAsset := asset.Asset{
			AssetAddr:     string(changeAssetAddr),
			Value:         -_value,
			AssetTypeID:   assetType,
			HasTransfered: false,
			Sign:          string(changeAssetSign),
		}
		return &changeAsset, nil
	}
	return nil, nil
}
