package FabricTransaction

import (
	"encoding/json"
	"errors"

	"github.com/FabricTransaction/asset"
	"github.com/FabricTransaction/assetPool"
	"github.com/FabricTransaction/common"
	"github.com/FabricTransaction/common/securityTool"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

type SignStruct struct {
	Sign string `json:"sign"`
}

type AssetPoolReq struct {
	assetPool.AssetPool
	SignStruct
}

type TransferReq struct {
	ToPool   string  `json:"toPool"`   //转入资产池ID
	FromPool string  `json:"fromPool"` //转出资产池ID
	Amount   float64 `json:"amount"`   //转让量
	TxType   string  `json:"txType"`   //交易类型：发行/转让
	asset.AssetInfo
	SignStruct
}

type SignVerifyStruct struct {
	AssetPoolID string `json:"assetPoolId"`
	SignStruct
}

func AbsTxInvoke(stub shim.ChaincodeStubInterface) pb.Response {
	_, args := stub.GetFunctionAndParameters()
	if len(args) < 1 {
		return shim.Error("no abstx invoke function")
	}
	switch args[0] {
	case "issue", "transfer":
		err := VerifyReq(stub, args[1])
		if err != nil {
			return shim.Error(err.Error())
		}

		tx := TransferReq{}
		err = json.Unmarshal([]byte(args[1]), &tx)
		if err != nil {
			return shim.Error(err.Error())
		}
		err = doAbsTx(stub, tx)
		if err != nil {
			return shim.Error(err.Error())
		}
		return shim.Success(nil)
	case "addAssetPool":
		pool := AssetPoolReq{}
		err := json.Unmarshal([]byte(args[1]), &pool)
		if err != nil {
			return shim.Error(err.Error())
		}
		err = AddAssetPool(stub, pool)
		if err != nil {
			return shim.Error(err.Error())
		}
		return shim.Success(nil)
	}

	return shim.Success(nil)
}

func doAbsTx(stub shim.ChaincodeStubInterface, tx TransferReq) error {
	if tx.TxType == common.TX_TYPE_TRANSFER {
		var from, to assetPool.AssetPool
		if err := common.GetDataByKey(stub, common.OBJECT_TYPE_ASEETPOOL, []string{tx.FromPool}, &from); err != nil {
			return err
		}
		if err := common.GetDataByKey(stub, common.OBJECT_TYPE_ASEETPOOL, []string{tx.ToPool}, &to); err != nil {
			return err
		}
		ok, err := from.Transfer(stub, tx.AssetTypeID, to, tx.Amount)
		if err != nil {
			return err
		}
		if !ok {
			return errors.New("transfer failed")
		}
	} else if tx.TxType == common.TX_TYPE_ISSUE {
		var issuePool assetPool.AssetPool
		if err := common.GetDataByKey(stub, common.OBJECT_TYPE_ASEETPOOL, []string{tx.ToPool}, &issuePool); err != nil {
			return err
		}
		return issuePool.Issue(stub, tx.Amount, tx.AssetInfo)
	}
	return errors.New("Invalid tx Type")
}

func VerifyReq(stub shim.ChaincodeStubInterface, reqStr string) error {
	verifyStruct := SignVerifyStruct{}
	err := json.Unmarshal([]byte(reqStr), &verifyStruct)
	if err != nil {
		return err
	}

	pool := assetPool.AssetPool{}
	err = common.GetDataByKey(stub, common.OBJECT_TYPE_ASEETPOOL, []string{verifyStruct.AssetPoolID}, &pool)
	if err != nil {
		return err
	}

	valid, err := securityTool.CheckJSONObjectSignatureString(reqStr, pool.PublicKey)
	if err != nil {
		return err
	}
	if !valid {
		return errors.New("verify sign failed")
	}

	return nil
}

func AddAssetPool(stub shim.ChaincodeStubInterface, pool AssetPoolReq) error {
	p := new(assetPool.AssetPool)
	return p.Init(stub, pool.AssetPoolAddr, pool.PublicKey, pool.AssetPoolType)
}
