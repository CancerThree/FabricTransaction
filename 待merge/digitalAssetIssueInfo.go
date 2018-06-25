package main

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

/*数字资产发行*/
type DigitalAssetIssue struct {
	DigitalAssetId   string       `json:"digitalAssetId,omitempty"`   //数字资产ID
	DigitalAssetName string       `json:"digitalAssetName,omitempty"` //数字资产名称
	DigitalAssetType string       `json:"digitalAssetType,omitempty"` //数字资产类型
	IssueScale       float64      `json:"issueScale"`                 //发行规模（元）
	CreateOrg        string       `json:"createOrg,omitempty"`        //创建机构
	ReqSign          string       `json:"reqSign,omitempty"`          //创建机构签名
	AssetPoolId      string       `json:"assetPoolId,omitempty"`      //资产库ID
	EffectSts        string       `json:"effectSts,omitempty"`        //生效状态
	SignData         string       `json:"signData,omitempty"`         //签名报文数据
	ObjectType       string       `json:"objectType,omitempty"`       //结构体类型
	MspId            string       `json:"mspId,omitempty"`            //所属机构
	ModMspId         string       `json:"modMspId,omitempty"`         //维护机构
	ModTime          string       `json:"modTime,omitempty"`          //维护时间
	ModUser          string       `json:"modUser,omitempty"`          //维护用户
	ModRmk           string       `json:"modRmk,omitempty"`           //维护原因
	NewAssetAddrs    []string     `json:"newAssetAddrs,omitempty"`    //资产地址
	OpeLogList       []*AbsOpelog `json:"opeLogList,omitempty"`
}

//------开始：智能合约-数字资产发行实现------

/*Name: createDigitalAssetIssue
**Args: string digitalAssetIssue json数据
**Description:
 */
func (t *AbsChaincode) createDigitalAssetIssue(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var digitalAssetIssue DigitalAssetIssue

	if len(args) != 1 {
		return shim.Error("createDigitalAssetIssue所需参数个数：1")
	}

	err := json.Unmarshal([]byte(args[0]), &digitalAssetIssue)
	if err != nil {
		return shim.Error("unmarshal args[0] failed, detail:" + err.Error())
	}

	//验签
	record, err := GetTXOrganization(stub, digitalAssetIssue.CreateOrg)
	if err != nil {
		return shim.Error("获取机构公钥信息失败:" + err.Error())
	}
	publicKey := record.SignPublicKey

	err = CheckJSONObjectSignature(&digitalAssetIssue, publicKey)
	if err != nil {
		return shim.Error("验证签名失败:" + err.Error())
	}

	//检查字段
	err = digitalAssetIssue.verifyField()
	if err != nil {
		return shim.Error(err.Error() + string(args[0]))
	}

	// err = checkMspId(stub, digitalAssetIssue.MspId)
	// if err != nil {
	// 	return shim.Error(err.Error())
	// }

	// err = checkMspId(stub, digitalAssetIssue.ModMspId)
	// if err != nil {
	// 	return shim.Error(err.Error())
	// }

	// err = checkAuth(stub, ObjectTypeDigitalAssetIssue, AUTHCOD_TYP_EDT, digitalAssetIssue.MspId)
	// if err != nil {
	// 	return shim.Error(err.Error())
	// }

	// check the product is already exist or not
	isExist, _, err := t.checkExistByKey(stub, ObjectTypeDigitalAssetIssue, []string{digitalAssetIssue.DigitalAssetId})
	if err != nil {
		return shim.Error(err.Error())
	}
	if isExist {
		return shim.Error("该数字资产发行信息已存在")
	}

	// add log info
	digitalAssetIssue.ObjectType = ObjectTypeDigitalAssetIssue
	opeLog := AbsOpelog{
		ModMspId: digitalAssetIssue.ModMspId,
		ModTime:  digitalAssetIssue.ModTime,
		ModUser:  digitalAssetIssue.ModUser,
		ModRmk:   digitalAssetIssue.ModRmk,
		ModTyp:   ModTyp_ADD,
		TrxId:    stub.GetTxID()}
	digitalAssetIssue.OpeLogList = append(digitalAssetIssue.OpeLogList, &opeLog)
	// digitalAssetIssue.ModRmk = ""

	val, err := json.Marshal(digitalAssetIssue)
	if err != nil {
		return shim.Error("transfer digitalAssetIssue to json failed, detail:" + err.Error())
	}

	key, err := stub.CreateCompositeKey(ObjectTypeDigitalAssetIssue, []string{digitalAssetIssue.DigitalAssetId})
	if err != nil {
		return shim.Error("Create Key Failed, Detail:" + err.Error())
	}

	err = stub.PutState(key, val)
	if err != nil {
		return shim.Error("PutState Failed. Detail:" + err.Error())
	}
	return shim.Success(nil)
}

/*Name: modifyDigitalAssetIssue
**Args: string digitalAssetIssue json数据
**Description:
 */
func (t *AbsChaincode) modifyDigitalAssetIssue(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var newDigitalAssetIssue DigitalAssetIssue

	if len(args) != 1 {
		return shim.Error("modifyDigitalAssetIssue所需参数个数:1")
	}

	err := json.Unmarshal([]byte(args[0]), &newDigitalAssetIssue)
	if err != nil {
		return shim.Error("unmarshal args[0] failed, detail:" + err.Error())
	}

	//验签

	// check argument is valid or not
	// if isEmptyStr(newProdInfo.ProdId) || isEmptyStr(newProdInfo.PkgId) || isEmptyStr(newProdInfo.ProdName) ||
	// 	isEmptyStr(newProdInfo.ProdType) {
	// 	return shim.Error("以下参数不能为空：ProdId/PkgId/ProName/ProType")
	// }
	// if newProdInfo.IssueScale < 0 {
	// 	return shim.Error("发行规模不可为负")
	// }

	//TODO 权限
	// err = checkMspId(stub, newDigitalAssetIssue.ModMspId)
	// if err != nil {
	// 	return shim.Error(err.Error())
	// }

	// check is exist by key
	isExist, oldVal, err := t.checkExistByKey(stub, ObjectTypeDigitalAssetIssue, []string{newDigitalAssetIssue.DigitalAssetId})
	if err != nil {
		return shim.Error(err.Error())
	}
	if !isExist {
		return shim.Error("该数字资产发行信息不存在")
	}

	var oldDigitalAssetIssue DigitalAssetIssue
	err = json.Unmarshal(oldVal, &oldDigitalAssetIssue)
	if err != nil {
		return shim.Error("unmarshal json failed, detail:" + err.Error())
	}

	//TODO 权限
	// err = checkAuth(stub, ObjectTypeProductInitial, AUTHCOD_TYP_EDT, oldProductInitial.MspId)
	// if err != nil {
	// 	return shim.Error(err.Error())
	// }

	// 非全量更新产品信息
	opeLog := AbsOpelog{
		ModMspId: newDigitalAssetIssue.ModMspId,
		ModTime:  newDigitalAssetIssue.ModTime,
		ModUser:  newDigitalAssetIssue.ModUser,
		ModRmk:   newDigitalAssetIssue.ModRmk,
		ModTyp:   ModTyp_MOD,
		TrxId:    stub.GetTxID()}

	oldDigitalAssetIssue.DigitalAssetName = newDigitalAssetIssue.DigitalAssetName
	oldDigitalAssetIssue.DigitalAssetType = newDigitalAssetIssue.DigitalAssetType
	oldDigitalAssetIssue.IssueScale = newDigitalAssetIssue.IssueScale
	oldDigitalAssetIssue.CreateOrg = newDigitalAssetIssue.CreateOrg
	oldDigitalAssetIssue.ReqSign = newDigitalAssetIssue.ReqSign
	oldDigitalAssetIssue.AssetPoolId = newDigitalAssetIssue.AssetPoolId
	oldDigitalAssetIssue.EffectSts = newDigitalAssetIssue.EffectSts

	oldDigitalAssetIssue.OpeLogList = append(oldDigitalAssetIssue.OpeLogList, &opeLog)
	oldDigitalAssetIssue.ModRmk = ""

	val, err := json.Marshal(oldDigitalAssetIssue)
	if err != nil {
		return shim.Error(err.Error())
	}

	key, err := stub.CreateCompositeKey(ObjectTypeDigitalAssetIssue, []string{oldDigitalAssetIssue.DigitalAssetId})
	if err != nil {
		return shim.Error("Create Key Failed, Detail:" + err.Error())
	}

	err = stub.PutState(key, val)
	if err != nil {
		return shim.Error("putState failed, detail:" + err.Error())
	}

	return shim.Success(nil)
}

/*Name: queryDigitalAssetIssueByKey
**Args: string digitalAssetId
**Description:
**/
func (t *AbsChaincode) queryDigitalAssetIssueByKey(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		return shim.Error("queryDigitalAssetIssueByKey所需参数个数：1")
	}

	key, err := stub.CreateCompositeKey(ObjectTypeDigitalAssetIssue, []string{args[0]})
	if err != nil {
		return shim.Error(err.Error())
	}
	val, err := stub.GetState(key)
	if err != nil {
		return shim.Error("query digitalAssetIssue failed, detail:" + err.Error())
	}

	if val == nil {
		return shim.Error("该记录不存在")
	}

	var digitalAssetIssue DigitalAssetIssue
	err = json.Unmarshal(val, &digitalAssetIssue)
	if err != nil {
		return shim.Error("unmarshal json failed, detail:" + err.Error())
	}

	//TODO add auth
	// err = checkAuth(stub, ObjectTypeDigitalAssetIssue, AUTHCOD_TYP_QRY, digitalAssetIssue.MspId)
	// if err != nil {
	// 	return shim.Error(err.Error())
	// }

	val, err = json.Marshal(digitalAssetIssue)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(val)
}

// 检查字段
func (digitalAssetIssue DigitalAssetIssue) verifyField() error {
	if strings.Count(digitalAssetIssue.DigitalAssetId, "") <= 1 {
		return errors.New("数字资产ID不能为空")
	}
	if strings.Count(digitalAssetIssue.DigitalAssetName, "") <= 1 {
		return errors.New("数字资产名称不能为空")
	}
	if strings.Count(digitalAssetIssue.DigitalAssetType, "") <= 1 {
		return errors.New("数字资产类型不能为空")
	}
	if digitalAssetIssue.IssueScale < 0 {
		return errors.New("发行规模不能为负数")
	}
	if strings.Count(digitalAssetIssue.CreateOrg, "") <= 1 {
		return errors.New("创建机构不能为空")
	}
	if strings.Count(digitalAssetIssue.ReqSign, "") <= 1 {
		return errors.New("创建机构签名不能为空")
	}
	if strings.Count(digitalAssetIssue.AssetPoolId, "") <= 1 {
		return errors.New("资产库ID不能为空")
	}
	if strings.Count(digitalAssetIssue.EffectSts, "") <= 1 {
		return errors.New("生效状态不能为空")
	}
	return nil
}

//------结束：智能合约-数字资产发行----------
