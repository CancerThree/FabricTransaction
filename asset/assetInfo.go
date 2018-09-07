package asset

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/hyperledger/fabric/core/chaincode/shim"

	"github.com/FabricTransaction/common"
)

type AssetInfo struct {
	AssetTypeID string  `json:"assetTypeId"`
	AssetName   string  `json:"assetName"`   //资产类型名称
	AssetSymbol string  `json:"assetSymbol"` //资产简称
	Decimals    string  `json:"decimals"`    //支持的小数点位数
	TotalSupply float64 `json:"totalSupply"` //总发行金额
}

func (ai *AssetInfo) VerifyFields() error {
	if common.IsEmptyStr(ai.AssetTypeID) {
		return errors.New("AssetTypeID is empty")
	}
	if common.IsEmptyStr(ai.AssetName) {
		return errors.New("AssetName is empty")
	}
	if common.IsEmptyStr(ai.AssetSymbol) {
		return errors.New("AssetSymbol is empty")
	}
	if ai.TotalSupply < 0 {
		return errors.New("invalid supply amount")
	}

	return nil
}

func (ai *AssetInfo) Store(stub shim.ChaincodeStubInterface) error {
	err := ai.VerifyFields()
	if err != nil {
		return err
	}

	key, err := stub.CreateCompositeKey(common.OBJECT_TYPE_ASSET_INFO, []string{ai.AssetTypeID})
	if err != nil {
		return err
	}
	bytes, err := json.Marshal(ai)
	if err != nil {
		return nil
	}
	return stub.PutState(key, bytes)
}

func (ai *AssetInfo) Init(stub shim.ChaincodeStubInterface, info AssetInfo) error {
	ai.AssetTypeID = info.AssetTypeID
	ai.AssetName = info.AssetName
	ai.AssetSymbol = info.AssetSymbol
	ai.Decimals = "0"
	ai.TotalSupply = info.TotalSupply

	exist, _, _, err := common.CheckExistByKey(stub, common.OBJECT_TYPE_ASSET_INFO, []string{ai.AssetTypeID})
	if err != nil {
		return err
	}
	if exist {
		return fmt.Errorf("asset %s already exists, id = %s", ai.AssetName, ai.AssetTypeID)
	}

	return ai.Store(stub)
}
