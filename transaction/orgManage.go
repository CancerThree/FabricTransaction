package transaction

import (
	"encoding/json"
	"errors"

	"github.com/hyperledger/fabric/core/chaincode/shim"
)

type Transaction struct {
	// ToOrgID	string	`json:"toOrgId"`	//放链下存放机构日志
	ToAccount   string `json:"toAccount"`
	FromAccount string `json:"fromAccount"`
	TimeStamp   string `json:"timeStamp"`

	OrgID         string   `json:"orgId"`
	AssetTypeID   string   `json:"assetTypeId"`
	Amount        float64  `json:"amount"`
	TxType        string   `json:"txType"`
	NewAssetAddrs []string `json:"newAssetAddrs"` //UUID
	AssetAddrs    []string `json:"assetAddrs"`
}

func Transfer(stub shim.ChaincodeStubInterface, args []string) error {
	var tx Transaction

	//TODO 签名验证
	data, err := unsignEncryptData(stub, args[0])
	if err != nil {
		return errors.New(err.Error())
	}

	err = json.Unmarshal(data, &tx)
	if err != nil {
		return errors.New("unmarshal transaction str failed" + err.Error())
	}

	assetsAddrs := tx.AssetAddrs[:]
	// 从解密后地址中查询资产
	assets, err := getAssetsByAddrs(stub, assetsAddrs)
	if err != nil {
		return errors.New("find assets failed")
	}
	balance := sortAndCountByAmount(&assets)

	//交易参数合法性验证

	if tx.Amount <= 0 {
		return errors.New("")
	}
	if balance < tx.Amount {
		return errors.New("poor balance")
	}

	//验证所用地址是否合法
	val, err := stub.GetState(tx.NewAddr[0])
	if val != nil {
		return errors.New("addr has been used:" + tx.NewAddr[0])
	}
	val, err = stub.GetState(tx.NewAddr[1])
	if val != nil {
		return errors.New("addr has been used:" + tx.NewAddr[1])
	}
}
