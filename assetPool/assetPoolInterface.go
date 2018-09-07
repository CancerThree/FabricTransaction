package assetPool

import (
	"encoding/json"
	"errors"
	"log"

	ast "github.com/FabricTransaction/asset"
	"github.com/FabricTransaction/common"
	"github.com/hyperledger/fabric/core/chaincode/shim"
)

type AssetPool struct {
	AssetPoolAddr string `json:"assetPoolAddr"`
	AssetPoolType string `json:"assetPoolType"`
	PublicKey     string `json:"publicKey"`
	// Hash          string `json:"hash"`
}

type AssetPoolInterface interface {
	Transfer(stub shim.ChaincodeStubInterface, _to AssetPool, _value float64) (bool, error)

	Approve(stub shim.ChaincodeStubInterface, _spender string, _value float64) (bool, error)

	Allowance(stub shim.ChaincodeStubInterface, _owner string)

	SetTransferEvent(stub shim.ChaincodeStubInterface, _from string, _to string, _value float64) error

	SetApprovalEvent(stub shim.ChaincodeStubInterface, _to string, _value float64) error

	Issue(stub shim.ChaincodeStubInterface, _value float64, assetTypeInfo ast.AssetInfo) error
}

func (pool *AssetPool) Init(stub shim.ChaincodeStubInterface, addr string, publicKey string, poolType string) error {
	pool.AssetPoolAddr = addr
	pool.AssetPoolType = poolType
	pool.PublicKey = publicKey

	exist, _, _, err := common.CheckExistByKey(stub, common.OBJECT_TYPE_ASEETPOOL, []string{addr})
	if err != nil {
		return err
	}
	if exist {
		return errors.New("asset pool " + addr + " already exists")
	}

	return pool.Store(stub)
}

func (pool *AssetPool) Store(stub shim.ChaincodeStubInterface) error {
	err := pool.VerifyFields()
	if err != nil {
		return err
	}

	bytes, err := json.Marshal(pool)
	if err != nil {
		return err
	}
	key, err := stub.CreateCompositeKey(common.OBJECT_TYPE_ASEETPOOL, []string{pool.AssetPoolAddr})
	if err != nil {
		return err
	}

	return stub.PutState(key, bytes)
}

func (pool *AssetPool) Transfer(stub shim.ChaincodeStubInterface, assetType string, _to AssetPool, _value float64) (bool, error) {
	priData, err := stub.GetTransient()
	if err != nil {
		return false, err
	}

	addrsBytes, ok := priData["assetAddrs"]
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
	assets, balance, err := ast.GetSortedAssetsByAddrs(stub, decryptAddrs)
	if err != nil {
		return false, err
	}
	if balance < _value {
		return false, errors.New("poor balance")
	}

	change, err := pool.BurnAssets(stub, assetType, *assets, _value, common.TX_TYPE_TRANSFER)
	if err != nil {
		return false, err
	}

	newAssetAddr, ok := priData["newAssetAddr"]
	if !ok {
		log.Println("get newAssetAddr failed, transient is:")
		log.Println(priData)
		return false, errors.New("cannot get newAssetAddr data")
	}
	err = _to.GenerateAndAddAsset(stub, string(newAssetAddr), _value, assetType)
	if err != nil {
		return false, err
	}

	if change != nil {
		err := pool.AddAsset(stub, change)
		if err != nil {
			return false, err
		}
	}

	return true, nil
}

func (pool *AssetPool) Issue(stub shim.ChaincodeStubInterface, _value float64, assetTypeInfo ast.AssetInfo) error {
	priData, err := stub.GetTransient()
	if err != nil {
		return err
	}

	assetAddr, ok := priData["assetAddr"]
	if !ok {
		log.Printf("get assetAddr failed, priData is %s", priData)
		return errors.New("get assetAddr failed: " + err.Error())
	}

	return pool.GenerateAndAddAsset(stub, string(assetAddr), _value, assetTypeInfo.AssetTypeID)
}

func (pool *AssetPool) AddAsset(stub shim.ChaincodeStubInterface, asset *ast.Asset) error {
	if asset == nil {
		return errors.New("AddAsset: empty asset")
	}

	exist, _, _, err := common.CheckExistByKey(stub, common.OBJECT_TYPE_ASSET, []string{asset.AssetAddr})
	if err != nil {
		return err
	}
	if exist {
		return errors.New("Invalid addr: asset addr exists")
	}

	err = asset.AddSign(stub, pool.AssetPoolAddr)
	if err != nil {
		return err
	}

	if err := asset.Store(stub); err != nil {
		return err
	}
	return nil
}

func (pool *AssetPool) GenerateAndAddAsset(stub shim.ChaincodeStubInterface, addr string, value float64, assetType string) error {
	asset := ast.Asset{
		AssetAddr:     addr,
		Value:         value,
		AssetTypeID:   assetType,
		HasTransfered: false,
	}
	err := pool.AddAsset(stub, &asset)
	return err
}

func (pool *AssetPool) BurnAssets(stub shim.ChaincodeStubInterface, assetType string, assets []ast.Asset, _value float64, burnType string) (*ast.Asset, error) {
	priData, err := stub.GetTransient()
	if err != nil {
		return nil, err
	}

	var burnAssetAddrArray []string
	encryptedAddrsBytes, ok := priData["encryptedAddrs"]
	if !ok {
		return nil, errors.New("no encryptedAddrs")
	}
	encryptedAddrs := []string{}
	err = json.Unmarshal(encryptedAddrsBytes, &encryptedAddrs)
	if err != nil {
		log.Println("unmarhsal encryptedAddrs failed: %s", encryptedAddrsBytes)
		return nil, err
	}

	if len(assets) < 1 {
		log.Println("no asset transfered")
		return nil, nil
	}

	for i, v := range assets {
		if !v.CanTransfer(stub, pool.AssetPoolAddr, assetType) {
			continue
		}
		if _value <= 0 {
			break
		}
		_value -= v.Value
		v.HasTransfered = true
		v.AddLogInfo()

		if err := v.Store(stub); err != nil {
			return nil, err
		}
		burnAssetAddrArray = append(burnAssetAddrArray, encryptedAddrs[i])
	}
	if _value > 0 {
		return nil, errors.New("poor balance")
	}

	err = pool.BurnAssetAddr(stub, burnAssetAddrArray)
	if err != nil {
		return nil, err
	}

	if _value < 0 {
		changeAssetAddr, ok := priData["changeAddr"]
		if !ok {
			return nil, errors.New("no changeAddr data")
		}
		changeAsset := ast.Asset{
			AssetAddr:     string(changeAssetAddr),
			Value:         -_value,
			AssetTypeID:   assetType,
			HasTransfered: false,
		}
		return &changeAsset, nil
	}
	return nil, nil
}

func (pool *AssetPool) BurnAssetAddr(stub shim.ChaincodeStubInterface, encryptAddrs []string) error {
	for _, v := range encryptAddrs {
		exists, _, val, err := common.CheckExistByKey(stub, common.OBJECT_TYPE_ASSET_ADDR, []string{pool.AssetPoolAddr, v})
		if err != nil {
			return errors.New("getState " + v + " failed:" + err.Error())
		}
		if !exists {
			continue
		}

		addr := AssetAddr{}
		err = json.Unmarshal(val, &addr)
		if err != nil {
			return errors.New("unmarshal " + v + " failed:" + err.Error())
		}
		if err = addr.Burn(stub); err != nil {
			return errors.New("burn " + v + " failed:" + err.Error())
		}
	}
	return nil
}

func (pool *AssetPool) VerifyFields() error {
	if common.IsEmptyStr(pool.AssetPoolAddr) {
		return errors.New("assetPool's addr is empty")
	}
	if common.IsEmptyStr(pool.AssetPoolType) {
		return errors.New("assetPool's type is empty")
	}
	if common.IsEmptyStr(pool.PublicKey) {
		return errors.New("assetPool's publicKey is empty")
	}
	return nil
}
