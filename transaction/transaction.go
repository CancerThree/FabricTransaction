package transaction

import (
	"encoding/json"
	"errors"

	"github.com/hyperledger/fabric/core/chaincode/shim"
)

type Transaction struct {
	ToOrgID   string `json:"toOrgId"` //放链下存放机构日志
	ToPool    string `json:"toPool"`
	FromPool  string `json:"fromPool"`
	TimeStamp string `json:"timeStamp"`

	OrgID         string   `json:"orgId"`
	AssetTypeID   string   `json:"assetTypeId"`
	Amount        float64  `json:"amount"`
	TxType        string   `json:"txType"`
	NewAssetAddrs []string `json:"newAssetAddrs"` //UUID
	AssetAddrs    []string `json:"assetAddrs"`
	LogInfo       string   `json:"logInfo,omitempty"`
	ModUser       string   `json:"modUser"`

	ReqSign string `json:"reqSign,omitempty"` //发送机构
	ReqSign string `json:"reqSign,omitempty"` //发送机构签名
}

func TransferByJsonStr(stub shim.ChaincodeStubInterface, str string) error {
	var tx Transaction

	err := json.Unmarshal([]byte(str), &tx)
	if err != nil {
		return errors.New("unmarshal transaction str failed" + err.Error())
	}

	return Transfer(stub, tx)
}

func Transfer(stub shim.ChaincodeStubInterface, tx Transaction) error {
	// var tx Transaction

	// //TODO 签名验证 && 验证OrgId
	// data, err := unsignEncryptData(stub, args[0])
	// if err != nil {
	// 	return errors.New(err.Error())
	// }

	//验证空字段
	if isEmptyStr(tx.AssetTypeID) || isEmptyStr(tx.ModUser) || isEmptyStr(tx.OrgID) ||
		isEmptyStr(tx.TimeStamp) || isEmptyStr(tx.ToOrgID) {
		return errors.New("assetType, timestamp, moduser, orgId cannot be null")
	}

	org, err := GetOrganizationByKey(stub, tx.OrgID)
	if err != nil {
		return errors.New("query org " + tx.OrgID + " failed:" + err.Error())
	}

	// 验签
	err = CheckJSONObjectSignature(tx, org.SignPublicKey)
	if err != nil {
		return errors.New("check signature failed:" + err.Error())
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
		return errors.New("invalid amount")
	}
	if balance < tx.Amount {
		return errors.New("poor balance")
	}

	//验证所用地址是否合法
	val, err := stub.GetState(tx.NewAddr[0])
	if err != nil {
		return errors.New("GetState new addr[0]" + tx.NewAddr[0] + " failed: " + err.Error())
	}
	if val != nil {
		return errors.New("addr has been used:" + tx.NewAddr[0])
	}
	val, err = stub.GetState(tx.NewAddr[1])
	if err != nil {
		return errors.New("GetState new addr[1]" + tx.NewAddr[1] + " failed: " + err.Error())
	}
	if val != nil {
		return errors.New("addr has been used:" + tx.NewAddr[1])
	}

	//验证资金池
	fromPool, err := getAssetPoolById(tx.FromPool)
	if err != nil {
		return errors.New("fromPool verified failed:" + err.Error())
	}
	err = verifyAccountOfOrg(tx.OrgID, fromPool)
	if err != nil {
		return err
	}

	toPool, err := getAssetPoolById(tx.ToPool)
	if err != nil {
		return errors.New("toPool verified failed:" + err.Error())
	}

	if err = transferAssets(stub, assets, tx, fromPool, toPool); err != nil {
		return errors.New("transfer failed.\r\n" + err.Error())
	}

	return nil
}

func IssueAssetByJsonStr(stub shim.ChaincodeStubInterface, str string) error {
	var tx Transaction
	//TODO 签名验证

	// if len(args) != 1 {
	// 	return errors.New("issue only need 1 argument")
	// }
	err := json.Unmarshal([]byte(args[0]), &tx)
	if err != nil {
		return errors.New("unmarshal tx failed:" + err.Error())
	}

}

func IssueAsset(stub shim.ChaincodeStubInterface, tx Transaction) error {
	//验证字段是否为空
	if isEmptyStr(tx.OrgID) || isEmptyStr(tx.AssetTypeID) || isEmptyStr(tx.ModUser) ||
		isEmptyStr(tx.TimeStamp) {
		return errors.New("orgID, assetTypeId, modUser, timestamp, accountAddr cannot be empty")
	}

	//验证数额
	if tx.Amount < 0.0 {
		return errors.New("issue amount cannot less than 0")
	}

	// 验证所用地址
	if len(tx.NewAssetAddrs) < 1 || isEmptyStr(tx.NewAssetAddrs[0]) {
		return errors.New("new asset addr is empty")
	}
	val, err := stub.GetState(tx.NewAssetAddrs[0])
	if err != nil {
		return errors.New("verify asset addr failed:" + err.Error())
	}
	if val != nil {
		return errors.New("assets already exist:" + tx.NewAssetAddrs[0])
	}

	//验证账户
	account, err := getAccountById(tx.ToPool)
	if err != nil {
		return errors.New("verify account id failed:" + err.Error())
	}

	err = addAssetUnderAccount(stub, account, tx.AssetAddrs[0], tx.Amount, tx.AssetTypeID)
	if err != nil {
		return errors.New("add asset failed:" + err.Error())
	}

	return nil
}
