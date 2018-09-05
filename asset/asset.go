package asset

import (
	"encoding/json"
	"errors"

	"github.com/FabricTransaction/transaction/common"
	"github.com/hyperledger/fabric/core/chaincode/shim"
)

type Asset struct {
	AssetAddr       string  `json:"assetAddr"`
	Value           float64 `json:"value"`
	AssetTypeID     string  `json:"assetTypeId"`
	HasTransfered   bool    `json:"hasTransfered"`
	LogInfo         string  `json:"logInfo,omitempty"`
	AuthedAssetPool string  `json:"authedAssetPool,omitempty"`
	Sign            string  `json:"sign"`
	// GenerateTime    string  `json:"generateTime"`
}

type AssetInfo struct {
	AssetName   string  `json:"assetName"`   //资产类型名称
	AssetSymbol string  `json:"assetSymbol"` //资产简称
	Decimals    string  `json:"decimals"`    //支持的小数点位数
	TotalSupply float64 `json:"totalSupply"` //总发行金额
}

func (asset *Asset) AddLogInfo() {
	// TODO
}

func (asset *Asset) CanTransfer() bool {
	return asset.Value >= 0 && asset.HasTransfered == false
}

func (asset *Asset) GetAssetInfo(stub shim.ChaincodeStubInterface) (*AssetInfo, error) {
	if common.IsEmptyStr(asset.AssetTypeID) {
		return nil, errors.New("No TypeID")
	}
	key, err := stub.CreateCompositeKey(common.OBJECT_TYPE_ASSET_INFO, []string{asset.AssetTypeID})
	if err != nil {
		return nil, err
	}
	val, err := stub.GetState(key)
	if err != nil {
		return nil, err
	}

	info := &AssetInfo{}
	if err := json.Unmarshal(val, info); err != nil {
		return nil, err
	}
	return info, nil
}

func GetAssetsByAddrs(stub shim.ChaincodeStubInterface, addrs []string) (*[]Asset, error) {
	var assets []Asset

	for _, addr := range addrs {
		key, err := stub.CreateCompositeKey(common.OBJECT_TYPE_ASSET, []string{addr})
		if err != nil {
			return nil, err
		}
		val, err := stub.GetState(key)
		if err != nil {
			return nil, err
		}

		var asset Asset
		if err = json.Unmarshal(val, &asset); err != nil {
			return nil, err
		}
		assets = append(assets, asset)
	}

	return &assets, nil
}

func GetSortedAssetsByAddrs(stub shim.ChaincodeStubInterface, addrs []string) (*[]Asset, float64, error) {
	var assets []Asset
	sum := float64(0)

	for _, addr := range addrs {
		key, err := stub.CreateCompositeKey(common.OBJECT_TYPE_ASSET, []string{addr})
		if err != nil {
			return nil, 0, err
		}
		val, err := stub.GetState(key)
		if err != nil {
			return nil, 0, err
		}

		var asset Asset
		if err = json.Unmarshal(val, &asset); err != nil {
			return nil, 0, err
		}
		assets = ascInsert(&assets, asset)
		sum += asset.Value
	}

	return &assets, sum, nil
}

func (asset *Asset) checkFields() error {
	// AssetAddr       string  `json:"assetAddr"`
	// Value           float64 `json:"value"`
	// AssetTypeID     string  `json:"assetTypeId"`
	// HasTransfered   bool    `json:"hasTransfered"`
	// LogInfo         string  `json:"logInfo,omitempty"`
	// AuthedAssetPool string  `json:"authedAssetPool,omitempty"`
	// Sign            string  `json:"sign"`
	if common.IsEmptyStr(asset.AssetAddr) {
		return errors.New("assetAddr is empty")
	}
	if common.IsEmptyStr(asset.AssetTypeID) {
		return errors.New("AssetTypeID is empty")
	}
}

func ascInsert(assets *[]Asset, asset Asset) []Asset {
	tmpAssets := *assets
	// var sortedAssets []Asset
	if assets == nil {
		return nil
	}
	for i, v := range tmpAssets {
		if v.Value > asset.Value {
			sortedAssets := make([]Asset, len(tmpAssets)+1)
			copy(sortedAssets, tmpAssets[:i])
			copy(sortedAssets[i:], []Asset{asset})
			copy(sortedAssets[i+1:], tmpAssets[i:])
			return sortedAssets
		}
	}
	sortedAssets := append(tmpAssets, asset)
	return sortedAssets
}
