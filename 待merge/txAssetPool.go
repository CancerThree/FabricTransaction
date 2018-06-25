package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

type TxAssetPoolRequest struct {
	AssetpoolId    string `json:"assetpoolId,omitempty"`    //资产库ID 主键
	AssetpoolType  string `json:"assetpoolType,omitempty"`  //资产库类型
	AssetpoolOwner string `json:"assetpoolOwner,omitempty"` //资产库所属机构
	PublicKey      string `json:"publicKey,omitempty"`      //公钥
	SignPublicKey  string `json:"signPublicKey,omitempty"`  //签名公钥
	ReqOrgId       string `json:"reqOrgId,omitempty"`       //发送机构
	ReqSign        string `json:"reqSign,omitempty"`        //发送机构签名
}

type TxAssetPool struct {
	AssetpoolId    string `json:"assetpoolId,omitempty"`    //资产库ID 主键
	AssetpoolType  string `json:"assetpoolType,omitempty"`  //资产库类型
	AssetpoolOwner string `json:"assetpoolOwner,omitempty"` //资产库所属机构
	PublicKey      string `json:"publicKey,omitempty"`      //公钥
	SignPublicKey  string `json:"signPublicKey,omitempty"`  //签名公钥
	ObjectType     string `json:"objectType,omitempty"`     //结构体类型
	MspId          string `json:"mspId,omitempty"`          //所属机构 维护机构
}

//添加资产库
func (t *AbsChaincode) addTxAssetPool(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var request TxAssetPoolRequest
	//验证参数
	if len(args) != 1 {
		return shim.Error("addTxAssetPool所需参数个数：1")
	}
	err := json.Unmarshal([]byte(args[0]), &request)
	if err != nil {
		return shim.Error(err.Error())
	}
	// 验证字段
	err = request.verifyField()
	if err != nil {
		return shim.Error(err.Error())
	}
	//验证签名
	fmt.Println("开始延签：")
	//查询机构公钥，验证机构公钥签名
	//var orgRecord TXOrganization
	orgRecord, err := GetTXOrganization(stub, request.ReqOrgId)
	if err != nil {
		return shim.Error("机构信息获取失败：" + err.Error())
	}
	err = CheckJSONObjectSignature(&request, orgRecord.SignPublicKey)
	if err != nil {
		return shim.Error("延签失败：" + err.Error())
	}
	fmt.Println("结束延签")

	//判重
	key, err := stub.CreateCompositeKey(ObjectTypeTxAssetPool, []string{request.AssetpoolId})
	if err != nil {
		return shim.Error(err.Error())
	}
	val, err := stub.GetState(key)
	if err != nil {
		return shim.Error(err.Error())
	}
	if val != nil {
		return shim.Error("资产库已存在")
	}

	//保存
	mspId, _ := getMspId(stub)
	record := TxAssetPool{
		AssetpoolId:    request.AssetpoolId,
		AssetpoolType:  request.AssetpoolType,
		AssetpoolOwner: request.AssetpoolOwner,
		PublicKey:      request.PublicKey,
		SignPublicKey:  request.SignPublicKey,
		ObjectType:     ObjectTypeTxAssetPool,
		MspId:          mspId,
	}

	val, err = json.Marshal(record)
	if err != nil {
		return shim.Error(err.Error())
	}
	stub.PutState(key, val)

	return shim.Success(nil)
}

func (t *AbsChaincode) queryTxAssetPool(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var record TxAssetPool

	if len(args) != 1 {
		return shim.Error("queryTxAssetPool所需参数个数：1")
	}
	assetpoolId := args[0]

	// generate compositeKey
	key, err := stub.CreateCompositeKey(ObjectTypeTxAssetPool, []string{assetpoolId})
	if err != nil {
		return shim.Error(err.Error())
	}

	val, err := stub.GetState(key)
	if err != nil {
		return shim.Error(err.Error())
	}
	if val == nil {
		return shim.Error("该记录不存在")
	}
	err = json.Unmarshal(val, &record)
	if err != nil {
		return shim.Error(err.Error())
	}

	//TODO 权限

	val, err = json.Marshal(record)
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(val)
}

func (request TxAssetPoolRequest) verifyField() error {
	if strings.Count(request.AssetpoolId, "") <= 1 {
		return errors.New("AssetpoolId为空")
	}
	if strings.Count(request.AssetpoolType, "") <= 1 {
		return errors.New("AssetpoolType为空")
	}
	if strings.Count(request.AssetpoolOwner, "") <= 1 {
		return errors.New("AssetpoolOwner为空")
	}
	if strings.Count(request.PublicKey, "") <= 1 {
		return errors.New("PublicKey为空")
	}
	if strings.Count(request.SignPublicKey, "") <= 1 {
		return errors.New("SignPublicKey为空")
	}
	if strings.Count(request.ReqOrgId, "") <= 1 {
		return errors.New("ReqOrgId为空")
	}
	if strings.Count(request.ReqSign, "") <= 1 {
		return errors.New("ReqSign为空")
	}
	return nil
}
