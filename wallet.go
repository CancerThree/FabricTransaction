package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

type WalletLog struct {
	ModUser    string `json:"modUser,omitempty"`
	Timestamp  string `json:"timestamp,omitempty"`
	TxId       string `json:"txId,omitempty"`
	RemoteAddr string `json:"inAddr,omitempty"`
	Amount     string `json:"amount,omitempty"`
	OpeType    string `json:"opeType,omitempty"`
	LogInfo    string `json:"logInfo,omitempty"`
}

type Wallet struct {
	MspId   string `json:"mspId,omitempty"`
	Addr    string `json:"addr,omitempty"`
	Balance string `json:"balance,omitempty"`
	// Timestamp string `json:"timestamp,omitempty"`
	Log []*WalletLog `json:"log,omitempty"`
}

type Signature struct {
	From       string `json:"from,omitempty"`
	To         string `json:"to,omitempty"`
	Timestamp  string `json:"timestamp,omitempty"`
	Amount     string `json:"amount,omitempty"`
	TxType     string `json:"txType,omitempty"`	
}

type Transaction struct {
	TxId       string `json:"txId,omitempty"`
	Nonce      string `json:"nonce,omitempty"` //nonce由客户端保证唯一性
	From       string `json:"from,omitempty"`
	To         string `json:"to,omitempty"`
	Timestamp  string `json:"timestamp,omitempty"`
	Amount     string `json:"amount,omitempty"`
	TxType     string `json:"txType,omitempty"`
	Sign	string	`json:"sign,omitempty"`
	DecryptKey string `json:"decryptKey,omitempty"`
}

func (t *AbsChaicode) InitWallet(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var wallet Wallet

	if len(args) != 1 {
		return shim.Error("invalid argument")
	}

	err := json.Unmarshal([]byte(args[0]), &wallet)
	if err != nil {
		return shim.Error("transfer to json failed")
	}

	if wallet.MspId == "" || wallet.Addr == "" || wallet.Balance == "" {
		return shim.Error("invalide arguments")
	}

	walletLog := WalletLog{
		ModUser:    wallet.MspId,
		Timestamp:  "",
		TxId:       stub.GetTxID(),
		RemoteAddr: "",
		Amount:     "",
		OpeType:    "new wallet account",
		LogInfo:    ""}
	wallet.WalletLog = append(wallet.WalletLog, walletLog)
	walletByte, _ := json.Marshal(wallet)
	stub.PutState(wallet.MspId+wallet.Addr, walletByte)
}

func (t *AbsChaincode) sendTransaction(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var tx Transaction
	err := json.Unmarshal([]byte(args[0]), &tx)

	decryptKey := tx.DecryptKey

	block, _ := pem.Decode(decryptKey)
	if block == nil {
		return shim.Error("invalid key")
	}
	priv, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return shim.Error("")
	}
	data, _ := rsa.DecryptPKCS1v15(rand.Reader, priv, []byte(args[0]))

	if data != nil {
		err = json.Unmarshal([]byte(data, &tx)
	}
}
