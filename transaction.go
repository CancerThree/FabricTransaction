package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"strconv"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

const (
	AccountPreifx       = "AccountOf"
	LOG_ID_PREFIX       = "LOG_"
	ASSET_HAS_SPENT     = "Y"
	ASSET_HAS_NOT_SPENT = "N"
)

type AccountLog struct { //私钥加密
	LogSerialNum string `json:"logSerialNum"`
	ModUser      string `json:"modUser,omitempty"`
	Timestamp    string `json:"timestamp,omitempty"`
	TxId         string `json:"txId,omitempty"`
	RemoteAddr   string `json:"inAddr,omitempty"`
	Amount       string `json:"amount,omitempty"`
	OpeType      string `json:"opeType,omitempty"`
	LogInfo      string `json:"logInfo,omitempty"`
}

type Account struct {
	Addr       string   `json:"addr,omitempty"`
	Assets     []string `json:"assets,omitempty"`
	Log        []string `json:"log,omitempty"`
	UnsignKey  string   `json:"unsignKey"`
	EncryptKey string   `json:"encryptKey"`
}

type Asset struct {
	Addr       string  `json:"addr,omitempty"`
	Value      float64 `json:"value"`
	AttachHash string  `json:"attachHash,omitempty"`
	HasSpent   string  `json:"hasSpent,omitempty"`
}

type Transaction struct {
	MspId      string   `json:"mspId,omitempty"`
	From       string   `json:"from,omitempty"`
	To         string   `json:"to,omitempty"`
	Timestamp  string   `json:"timestamp,omitempty"`
	Amount     float64  `json:"amount,omitempty"`
	TxType     string   `json:"txType,omitempty"`
	NewAddr    []string `json:"newAddr,omitempty"`
	Info       string   `json:"info,omitempty"`
	AssetAddrs []string `json:"assetAddrs,omitempty"`
}

//获取签名解密密钥
func (t *AbsChaincode) getUnsignKey(stub shim.ChaincodeStubInterface) ([]byte, error) {
	var account Account
	val, err := stub.GetState(t.getAccoutAddr(stub))
	if err != nil {
		return nil, errors.New("cannot find account")
	}

	err = json.Unmarshal(val, &account)
	if err != nil {
		return nil, errors.New("account unmarshal failed")
	}

	return []byte(account.UnsignKey), nil
}

//解签名交易数据
func (t *AbsChaincode) unsignEncryptData(stub shim.ChaincodeStubInterface, signedStr string) ([]byte, error) {
	decryptKey, err := t.getUnsignKey(stub)
	if err != nil {
		return nil, err
	}

	//获取解密密钥
	block, _ := pem.Decode(decryptKey)
	if block == nil {
		return nil, errors.New("invalid decrypt key")
	}
	priv, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, errors.New("parse key failed:" + err.Error())
	}

	//base64解码加密字符串
	encryptedData, err := base64.StdEncoding.DecodeString(signedStr)
	if err != nil {
		return nil, errors.New("base64 decode string failed:" + err.Error())
	}

	//解密
	data, err := rsa.DecryptPKCS1v15(rand.Reader, priv, encryptedData)
	if data == nil {
		return nil, errors.New("decrypt failed:" + err.Error())
	}

	return data, nil
}

func (t *AbsChaincode) encryptStrData(stub shim.ChaincodeStubInterface, mspId string, data string) (string, error) {
	str, err := t.encryptByteData(stub, mspId, []byte(data))
	if err != nil {
		return nil, err
	}
	return str, nil
}

//使用机构加密密钥进行私有信息加密
func (t *AbsChaincode) encryptByteData(stub shim.ChaincodeStubInterface, mspId string, data []byte) (string, error) {
	// 查询机构加密密钥
	encryptKey, err := t.getEncryptKey(stub, mspId)
	if err != nil {
		return nil, err
	}

	//解析提取密钥
	block, _ := pem.Decode(encryptKey)
	if block == nil {
		return nil, errors.New("decode encrypt key failed, encryptKey bytes:" + string(encryptKey))
	}
	pubInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, errors.New("parse key failed:" + err.Error())
	}
	pub := pubInterface.(*rsa.PublicKey)

	//加密
	encryptedData, err := rsa.EncryptPKCS1v15(rand.Reader, pub, data)
	if err != nil {
		return nil, errors.New("encrypt failed:" + err.Error())
	}

	//base64编码
	encryptStr := base64.StdEncoding.EncodeToString(encryptedData)

	return encryptedStr, nil
}

//获取机构私有信息加密密钥
func (t *AbsChaincode) getEncryptKey(stub shim.ChaincodeStubInterface, accountId string) ([]byte, error) {
	var account Account
	val, err := stub.GetState(AccountPreifx + accountId)
	if err != nil {
		return nil, errors.New("cannot find account")
	}

	err = json.Unmarshal(val, &account)
	if err != nil {
		return nil, errors.New("account unmarshal failed")
	}

	return []byte(account.EncryptKey), nil
}

//计算字符串的SHA256哈希值，并且将哈希值转为Base64编码的字符串返回
func getShaBase64Str(str string) (string, error) {
	hash := sha256.New()
	_, err := hash.Write([]byte(str))
	if err != nil {
		return "", err
	}
	hashByte := hash.Sum(nil)
	return base64.StdEncoding.EncodeToString(hashByte)
}

/* Name: InitAccount
 * Description: 初始化账户，在链上创建调用该接口的机构的账户
 */
func (t *AbsChaincode) InitAccount(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var account Account

	if len(args) != 1 {
		return shim.Error("invalid argument")
	}

	operator, err := getMspId(stub)
	if err != nil {
		return shim.Error("get mspid failed:" + err.Error())
	}

	err = json.Unmarshal([]byte(args[0]), &account)
	if err != nil {
		return shim.Error("transfer to json failed")
	}

	account.Addr = t.getAccoutAddr(stub)
	account.Assets = []string{}
	if len(account.UnsignKey) == 0 {
		return shim.Error("unsignkey is missing")
	}
	if len(account.EncryptKey) == 0 {
		return shim.Error("EncryptKey is missing")
	}

	accountLog := AccountLog{
		ModUser:    operator,
		Timestamp:  "",
		TxId:       stub.GetTxID(),
		RemoteAddr: "",
		Amount:     "",
		OpeType:    "new account",
		LogInfo:    ""}
	account.Log = append(account.Log, &accountLog)
	accountByte, _ := json.Marshal(account)
	err = stub.PutState(account.Addr, accountByte)
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(nil)
}

// 判断asset是否可进行转让
func canTransfer(asset Asset, mspId string) (bool, error) {
	if asset.Amount <= 0 {
		return false, errors.New("invalid asset, system error")
	}

	if asset.HasSpent == ASSET_HAS_SPENT {
		return false, errors.New("asset has spent")
	}

	hashSign, err := getShaBase64Str(asset.addr + mspId)
	if err != nil {
		return false, errors.New("calc hash failed:" + err.Error())
	}

	if hashSign != asset.AttachHash {
		return false, errors.New("asset's hash cannot match")
	}

	return true, nil
}

func (t *AbsChaincode) getAccoutAddr(stub shim.ChaincodeStubInterface) string {
	mspId, _ := getMspId(stub)

	return AccountPreifx + mspId
}

func (t *AbsChaincode) storeAccountLog(stub shim.ChaincodeStubInterface, account *Account, log AccountLog) error {
	txId := stub.GetTxID()
	mspId := getMspId(stub)
	// 填公共获取方法的字段
	log.LogSerialNum = LOG_ID_PREFIX + txId
	log.ModUser = mspId
	log.TxId = txId

	//加密日志信息
	logBytes, err := json.Marshal(log)
	if err != nil {
		return err
	}
	encryptedStr, err := t.encryptByteData(stub, mspId, logBytes)
	if err != nil {
		return err
	}

	stub.PutState(log.LogSerialNum, []byte(encryptedStr))

	encryptedLogId, err := t.encryptStrData(stub, mspId, log.LogSerialNum)
	if err != nil {
		return err
	}

	account.Log = append(account.Log, encryptedLogId)
	return nil
}

//将asset按照Value值从小到大排序，并且返回asset的Value总和
func sortAndCountByAmount(assets *[]Asset) float64 {
	if assets == nil {
		return 0
	}

	sum := 0.0

	assetArray := *assets
	for i := 0; i < len(assetArray); i++ {
		for j := len(assetArray) - 1; j > i; j-- {
			if assetArray[i].Value > assetArray[j].Value {
				tempAsset := assetArray[i]
				assetArray[i] = assetArray[j]
				assetArray[j] = tempAsset
			}
		}
		sum += assetArray[i].Value
	}
	return sum
}

//查询地址下的所有资产
func getAssetsByAddrs(stub shim.ChaincodeStubInterface, addrs []string) ([]Asset, error) {
	// assets := make([]Asset, len(addrs))
	var assets []Asset

	for i := 0; i < len(addrs); i++ {
		var asset Asset
		val, err := stub.GetState(addrs[i])
		if err != nil {
			return nil, errors.New("get asset failed:" + addrs[i])
		}

		err = json.Unmarshal(val, &asset)
		if err != nil {
			return nil, errors.New("unmarshal asset failed:" + addrs[i])
		}

		assets = append(assets, asset)
	}

	return assets, nil
}

//进行asset的转移
func (t *AbsChaincode) transferAssets(stub shim.ChaincodeStubInterface, assets []Asset, amount float64, newAssetAddrs []string, remoteMspId string) error {
	mspId, err := getMspId(stub)
	if err != nil {
		return errors.New("get mspId failed:" + err.Error())
	}

	transferAmount := amount
	if amount <= 0 {
		return errors.New("invalid amount: amount <=0: ")
	}

	// 从小到大修改asset
	for i := 0; i < len(assets); i++ {
		if amount <= 0 {
			break
		}

		_, err := canTransfer(assets[i], mspId)
		if err != nil {
			return err
		}

		assets[i].HasSpent = "Y"
		amount -= assets[i].Value

		assetBytes, err := json.Marshal(assets[i])
		if err != nil {
			return errors.New("marshal asset failed:" + err.Error())
		}

		stub.PutState(assets[i].Addr, assetBytes)
	}
	val, err := stub.GetState(newAssetAddrs[0])
	if val != nil {
		return errors.New("addr has been used:" + newAssetAddrs[0])
	}
	val, err = stub.GetState(newAssetAddrs[1])
	if val != nil {
		return errors.New("addr has been used:" + newAssetAddrs[1])
	}

	attchHash := sha256.New()
	attchHash.Write([]byte(newAssetAddrs[0] + remoteMspId))
	sign := attchHash.Sum(nil)
	hashSign := base64.StdEncoding.EncodeToString(sign)

	transferAsset := Asset{
		Addr:       newAssetAddrs[0],
		Value:      transferAmount,
		AttachHash: hashSign,
		HasSpent:   "N"}
	transferAssetByte, err := json.Marshal(transferAsset)
	if err != nil {
		return errors.New("marshal new asset failed")
	}
	stub.PutState(newAssetAddrs[0], transferAssetByte)

	encryptKey, err := t.getEncryptKey(stub, remoteMspId)
	block, _ := pem.Decode(encryptKey)
	if block == nil {
		return errors.New("get encrypt key failed, err:" + err.Error())
	}
	pubInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return err
	}
	pub := pubInterface.(*rsa.PublicKey)
	data, err := rsa.EncryptPKCS1v15(rand.Reader, pub, []byte(newAssetAddrs[0]))
	if err != nil {
		return err
	}

	afterData := base64.StdEncoding.EncodeToString(data)
	var account Account
	accountByte, _ := stub.GetState(AccountPreifx + remoteMspId)
	if accountByte != nil {
		json.Unmarshal(accountByte, &account)
	}

	account.Assets = append(account.Assets, afterData)
	accountByte, _ = json.Marshal(account)
	stub.PutState(AccountPreifx+remoteMspId, accountByte)

	if amount < 0 {
		attchHash = sha256.New()
		attchHash.Write([]byte(newAssetAddrs[1] + remoteMspId))
		sign = attchHash.Sum(nil)
		hashSign := base64.StdEncoding.EncodeToString(sign)

		transferAsset = Asset{
			Addr:       newAssetAddrs[1],
			Value:      0 - amount,
			AttachHash: hashSign,
			HasSpent:   "N"}
		transferAssetByte, err := json.Marshal(transferAsset)
		if err != nil {
			return errors.New("marshal new asset failed")
		}
		stub.PutState(newAssetAddrs[1], transferAssetByte)
	}
	return nil
}

//转账交易
func (t *AbsChaincode) transfer(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var tx Transaction
	err := json.Unmarshal([]byte(args[0]), &tx)

	// 签名验证
	decryptKey, err := t.getUnsignKey(stub)
	if err != nil {
		return shim.Error(err.Error())
	}

	block, _ := pem.Decode(decryptKey)
	if block == nil {
		return shim.Error("invalid key")
	}
	priv, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return shim.Error("parse peivate key failed")
	}
	encryptData, err := base64.StdEncoding.DecodeString(args[0])
	if err != nil {
		return shim.Error(err.Error())
	}
	data, _ := rsa.DecryptPKCS1v15(rand.Reader, priv, encryptData)

	if data != nil {
		err = json.Unmarshal([]byte(data), &tx)
	}

	assetsAddrs := tx.AssetAddrs[:]

	assets, err := getAssetsByAddrs(stub, assetsAddrs)
	if err != nil {
		return shim.Error("find assets failed")
	}
	balance := sortAndCountByAmount(&assets)

	if balance < tx.Amount {
		return shim.Error("poor balance")
	}

	err = t.transferAssets(stub, assets, tx.Amount, tx.NewAddr, tx.To)
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(nil)
}

func (t *AbsChaincode) coinBase(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 2 {
		return shim.Error("need 2 argument: amount, id")
	}

	var amount float64
	amount, err := strconv.ParseFloat(args[0], 64)

	mspId, _ := getMspId(stub)
	attchHash := sha256.New()
	attchHash.Write([]byte(args[1] + mspId))
	sign := attchHash.Sum(nil)
	hashSign := base64.StdEncoding.EncodeToString(sign)

	transferAsset := Asset{
		Addr:       args[1],
		Value:      amount,
		AttachHash: hashSign,
		HasSpent:   "N"}

	bytes, _ := json.Marshal(transferAsset)
	stub.PutState(args[1], bytes)

	encryptKey, _ := t.getEncryptKey(stub, mspId)
	block, _ := pem.Decode(encryptKey)
	if block == nil {
		return shim.Error("get encryptKey failed:" + string(encryptKey))
	}
	pubInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return shim.Error("get publicKey failed")
	}
	pub := pubInterface.(*rsa.PublicKey)
	data, _ := rsa.EncryptPKCS1v15(rand.Reader, pub, []byte(args[1]))
	afterData := base64.StdEncoding.EncodeToString(data)
	var account Account
	accountByte, _ := stub.GetState(AccountPreifx + mspId)
	if accountByte != nil {
		json.Unmarshal(accountByte, &account)
	}
	account.Assets = append(account.Assets, afterData)
	accountByte, _ = json.Marshal(account)
	stub.PutState(AccountPreifx+mspId, accountByte)
	return shim.Success(nil)
}

func (t *AbsChaincode) queryById(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		return shim.Error("need 1 args: id")
	}

	val, err := stub.GetState(args[0])
	if err != nil {
		return shim.Error("getstate failed:" + err.Error())
	}

	return shim.Success(val)
}

/* Description: 根据解密后地址查询所有可交易资产
 */
func (t *AbsChaincode) getAssetsByAddrs(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var assetAddrs []string
	var assets []Asset

	if len(args) != 1 {
		return shim.Error("only need addrs")
	}
	err := json.Unmarshal([]byte(args[0]), &assetAddrs)
	if err != nil {
		return shim.Error("unmarshal addrs failed:" + err.Error())
	}

	for i := 0; i < len(assetAddrs); i++ {
		var asset Asset
		val, err := stub.GetState(assetAddrs[0])
		if err != nil {
			return shim.Error("get asset failed:" + err.Error())
		}
		err = json.Unmarshal(val, &asset)
		if err != nil {
			return shim.Error("unmarshal asset failed:" + err.Error())
		}

		assets = append(assets, asset)
	}

	val, err := json.Marshal(assets)
	if err != nil {
		return shim.Error("marshal assets failed:" + err.Error())
	}

	return shim.Success(val)
}
