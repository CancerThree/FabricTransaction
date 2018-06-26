package transaction

import (
	"encoding/json"
	"errors"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

type OrganizationRequest struct {
	OrgId         string `json:"orgId,omitempty"`         //机构ID
	PublicKey     string `json:"publicKey,omitempty"`     //公钥
	SignPublicKey string `json:"signPublicKey,omitempty"` //签名公钥

	ReqOrgId string `json:"reqOrgId,omitempty"` //发送机构
	ReqSign  string `json:"reqSign,omitempty"`  //发送机构
}

type Organization struct {
	OrgId         string `json:"orgId,omitempty"`         //机构ID 主键
	PublicKey     string `json:"publicKey,omitempty"`     //公钥
	SignPublicKey string `json:"signPublicKey,omitempty"` //签名公钥

	ObjectType string `json:"objectType,omitempty"` //结构体类型
	MspId      string `json:"mspId,omitempty"`      //创建的节点机构
}

//添加机构
func (t *AbsChaincode) addOrganization(stub shim.ChaincodeStubInterface, args []string) error {
	var request OrganizationRequest
	//验证参数
	if len(args) != 1 {
		return errors.New("addOrganization所需参数个数：1")
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
	publicKey := request.SignPublicKey
	err = CheckJSONObjectSignature(&request, publicKey)
	if err != nil {
		return errors.New("验签失败：" + err.Error())
	}
	//TODO 权限控制

	//判重
	key, err := stub.CreateCompositeKey(ObjectTypeOrganization, []string{request.OrgId})
	if err != nil {
		return errors.New("create key failed:" + err.Error())
	}
	val, err := stub.GetState(key)
	if err != nil {
		return errors.New("query org failed:" + err.Error())
	}
	if val != nil {
		return errors.New("交易机构已存在")
	}

	//保存
	mspId, _ := getMspId(stub)
	record := Organization{
		OrgId:         request.OrgId,
		PublicKey:     request.PublicKey,
		SignPublicKey: request.SignPublicKey,
		ObjectType:    ObjectTypeOrganization,
		MspId:         mspId,
	}

	val, err = json.Marshal(record)
	if err != nil {
		return shim.Error(err.Error())
	}
	stub.PutState(key, val)

	return shim.Success(nil)
}

func (t *AbsChaincode) queryOrganization(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var record Organization

	if len(args) != 1 {
		return shim.Error("queryOrganization所需参数个数：1")
	}
	orgId := args[0]

	// generate compositeKey
	key, err := stub.CreateCompositeKey(ObjectTypeOrganization, []string{orgId})
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

func (request OrganizationRequest) verifyField() error {
	if isEmptyStr(request.OrgId) <= 1 {
		return errors.New("OrgId为空")
	}
	if isEmptyStr(request.PublicKey) <= 1 {
		return errors.New("PublicKey为空")
	}
	if isEmptyStr(request.SignPublicKey) <= 1 {
		return errors.New("SignPublicKey为空")
	}
	if isEmptyStr(request.ReqOrgId) <= 1 {
		return errors.New("ReqOrgId为空")
	}
	if isEmptyStr(request.ReqSign) <= 1 {
		return errors.New("ReqSign为空")
	}
	return nil
}

//根据orgId查询机构信息，供其他模块查询公钥使用
func GetOrganization(stub shim.ChaincodeStubInterface, orgId string) (*Organization, error) {
	var record Organization
	key, err := stub.CreateCompositeKey(ObjectTypeOrganization, []string{orgId})
	if err != nil {
		return nil, err
	}

	val, err := stub.GetState(key)
	if err != nil {
		return nil, err
	}
	if val == nil {
		return nil, errors.New("记录不存在")
	}
	err = json.Unmarshal(val, &record)
	if err != nil {
		return nil, err
	}
	return &record, nil
}
