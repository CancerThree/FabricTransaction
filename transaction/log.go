package transaction

import (
	"encoding/json"
	"errors"

	"github.com/hyperledger/fabric/core/chaincode/shim"
)

type OrgLogAddr struct {
	OrgID        string `json:"orgId"`
	EncryptLogID string `json:"encryptLogId"`
	ObjectType   string `json:"objectType"`
	TimeStamp    string `json:"timestamp"`
}

type OrgPrivateLog struct {
	FromOrg   string  `json:"fromOrg"`
	ToOrg     string  `json:"toOrg"`
	FromPool  string  `json:"fromPool"`
	ToPool    string  `json:"toPool"`
	Amount    float64 `json:"amount"`
	AssetType string  `json:"assetType"`
	Timestamp string  `json:"timestamp"`
	OpeType   string  `json:"opeType"`
	LogInfo   string  `json:"logInfo"`
	TxId      string  `json:"txId"`
	ModUser   string  `json:"modUser"`
}

type ChainLog struct {
	FromPool   string  `json:"fromPool"`
	ToPool     string  `json:"toPool"`
	Value      float64 `json:"value"`
	AssetType  string  `json:"assetType"`
	Timestamp  string  `json:"timestamp"`
	ObjectType string  `json:"objectType"`
	TxID       string  `json:"txId"`
}

func GetPrivateLogByAddrJsonStr(stub shim.ChaincodeStubInterface, arrayJsonStr string) ([]OrgPrivateLog, error) {
	var addrs []string

	err := json.Unmarshal([]byte(arrayJsonStr), &addrs)
	if err != nil {
		return nil, errors.New("unmarshal array str failed:" + err.Error())
	}

	return GetPrivateLogByAddrs(stub, addrs)
}

func GetPrivateLogByAddrs(stub shim.ChaincodeStubInterface, addrs []string) ([]OrgPrivateLog, error) {
	logs := []OrgPrivateLog{}

	for _, addr := range addrs {
		var log OrgPrivateLog
		val, err := stub.GetState(addr)
		if err != nil {
			return nil, errors.New("query log" + addr + " failed:" + err.Error())
		}
		err = json.Unmarshal(val, &log)
		if err != nil {
			return nil, errors.New("unmarshal log" + addr + " failed:" + err.Error())
		}
		logs = append(logs, log)
	}
	return logs, nil
}

func addOrgPrivateLog(stub shim.ChaincodeStubInterface, tx Transaction, assetPool AssetPool, opeType string) error {
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
		FromOrg:   tx.OrgID,
		ToOrg:     tx.ToOrgID,
		FromPool:  tx.FromPool,
		ToPool:    tx.ToPool,
		Amount:    tx.Amount,
		AssetType: tx.AssetTypeID,
		Timestamp: tx.TimeStamp,
		OpeType:   opeType,
		LogInfo:   tx.LogInfo,
		TxId:      txId,
		ModUser:   modUser}
	logKey := ACCOUNT_LOG_PREFIX + txId

	logBytes, err := json.Marshal(log)
	if err != nil {
		return errors.New("marshal log failed:" + err.Error())
	}

	logStr, err := encryptData2Base64Str([]byte(assetPool.EncryptKey), logBytes)
	if err != nil {
		return err
	}

	if err = stub.PutState(logKey, []byte(logStr)); err != nil {
		return errors.New("putState log failed:" + err.Error())
	}

	encryptLogKeyStr, err := encryptData2Base64Str([]byte(assetPool.EncryptKey), []byte(logKey))
	if err != nil {
		return err
	}

	logAddr := OrgLogAddr{
		OrgID:        assetPool.OrgID,
		EncryptLogID: encryptLogKeyStr,
		ObjectType:   OBJECT_TYPE_LOG_ADDR,
		TimeStamp:    tx.TimeStamp}
	logAddrKey, err := stub.CreateCompositeKey(OBJECT_TYPE_LOG_ADDR, []string{assetPool.OrgID, encryptLogKeyStr})
	if err != nil {
		return errors.New(err.Error())
	}

	logAddrBytes, err := json.Marshal(logAddr)
	if err != nil {
		return err
	}

	err = stub.PutState(logAddrKey, logAddrBytes)
	if err != nil {
		return errors.New("store logAddr failed:" + err.Error())
	}

	return nil
}

func addChainLog(stub shim.ChaincodeStubInterface, tx Transaction, opeType string) error {
	txId := stub.GetTxID()
	key := CHAIN_LOG_PREFIX + txId

	chainLog := ChainLog{
		FromPool:   tx.FromPool,
		ToPool:     tx.ToPool,
		Value:      tx.Amount,
		AssetType:  tx.AssetTypeID,
		Timestamp:  tx.TimeStamp,
		ObjectType: OBJECT_TYPE_CHAIN_LOG,
		TxID:       txId}
	chainLogBytes, err := json.Marshal(chainLog)
	if err != nil {
		return errors.New("marshal chainlog failed:" + err.Error())
	}

	err = stub.PutState(key, chainLogBytes)
	return err
}
