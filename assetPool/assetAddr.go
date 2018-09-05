package assetPool

import (
	"encoding/json"
	"errors"

	ast "github.com/FabricTransaction/asset"
	"github.com/FabricTransaction/common"
	"github.com/FabricTransaction/common/securityTool"
	"github.com/hyperledger/fabric/core/chaincode/shim"
)

type AssetAddr struct {
	AssetPoolAddr    string `json:"assetPoolAddr"`
	EncryptAssetAddr string `json:"encryptAssetAddr"`
	AssetTypeID      string `json:"assetTypeId"`
	BurnTime         string `json:"burnTime"`
	HasTransfered    bool   `json:"hasTransfered"`
}

func GenerateAndStoreAssetAddr(stub shim.ChaincodeStubInterface, asset ast.Asset, pool AssetPool) error {
	cryptSuite := securityTool.RSATool{}

	ctStr, err := cryptSuite.EncryptByPoolPublicKey([]byte(pool.PublicKey), []byte(asset.AssetAddr))
	if err != nil {
		return err
	}
	addr := AssetAddr{
		AssetPoolAddr:    pool.AssetPoolAddr,
		EncryptAssetAddr: ctStr,
		AssetTypeID:      asset.AssetTypeID,
		HasTransfered:    false,
	}

	exist, key, _, err := common.CheckExistByKey(stub, common.OBJECT_TYPE_ASSET_ADDR, []string{addr.AssetPoolAddr, addr.EncryptAssetAddr})
	if exist {
		return errors.New("asset addr already exists")
	}
	val, err := json.Marshal(addr)
	if err != nil {
		return err
	}

	return stub.PutState(key, val)
}

func (assetAddr *AssetAddr) StoreAssetAddr(stub shim.ChaincodeStubInterface) error {
	key, err := stub.CreateCompositeKey(common.OBJECT_TYPE_ASSET_ADDR, []string{assetAddr.AssetPoolAddr, assetAddr.EncryptAssetAddr})
	if err != nil {
		return err
	}
	val, err := json.Marshal(assetAddr)
	if err != nil {
		return err
	}
	return stub.PutState(key, val)
}

func (assetAddr *AssetAddr) Burn(stub shim.ChaincodeStubInterface) error {
	assetAddr.HasTransfered = true
	return assetAddr.StoreAssetAddr(stub)
}
