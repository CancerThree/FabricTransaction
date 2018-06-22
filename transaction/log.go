package transaction

import (
	"errors"

	"github.com/hyperledger/fabric/core/chaincode/shim"
)

type OrgLogAddr struct {
	OrgID        string `json:"orgId"`
	EncryptLogID string `json:"encryptLogId"`
	ObjectType   string `json:"objectType"`
}

type OrgPrivateLog struct {
	FromOrg   string  `json:"fromOrg"`
	ToOrg     string  `json:"toOrg"`
	Amount    float64 `json:"amount"`
	AssetType string  `json:"assetType"`
	Timestamp string  `json:"timestamp"`
	OpeType   string  `json:"opeType"`
	LogInfo   string  `json:"logInfo"`
	TxId      string  `json:"txId"`
	ModUser   string  `json:"modUser"`
}

type ChainLog struct {
	FromAcc   string  `json:"fromAcc"`
	ToAcc     string  `json:"toAcc"`
	Value     float64 `json:"value"`
	AssetType string  `json:"assetType"`
	Timestamp	string	`json:"timestamp"`
}

func addOrgPrivateLog(stub shim.ChaincodeStubInterface, tx Transaction, acc Account, opeType string) error {
	txId := stub.GetTxID()
	key := ACCOUNT_LOG_PREFIX + txId
	val, err := stub.GetState(key)
	if err != nil {
		return errors.New("find log failed:" + err.Error())
	}
	if val != nil {
		return errors.New("system error, log has already existed")
	}
	modUser := ""
	if opeType != TX_TYPE_TRANSFER_IN {
		modUser = tx.ModUser
	}

	log := OrgPrivateLog{
		FromOrg:   tx.FromAccount,
		ToOrg:     tx.ToAccount,
		Amount:    tx.Amount,
		AssetType: tx.AssetTypeID,
		Timestamp: tx.TimeStamp,
		OpeType:   opeType,
		LogInfo:   tx.LogInfo,
		TxId:      txId,
		ModUser:   modUser}
	
	logAddr := OrgLogAddr {
		OrgID:
	}

}
