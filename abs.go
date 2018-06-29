package main

import (
	"fmt"
	"time"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

//------开始：全局结构体定义------

/*Abs智能合约*/
type AbsChaincode struct {
}

//------开始：智能合约-main,init和invoke------

func main() {
	err := shim.Start(new(AbsChaincode))
	if err != nil {
		fmt.Printf("Error starting AbsChaincode: %s", err)
	}
}

/* 初始化：获取版本号、时间、mspId进行记录chaincode的版本更新历史记录
 */
func (t *AbsChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {

}

func (t *AbsChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	function, args := stub.GetFunctionAndParameters()

	switch function {

	case "clearAllData":
		return t.clearAllData(stub, args)
		//TODO: add interface to reset all data
	}

	return shim.Error("该方法" + function + "不存在")
}

func isValidateDate(dateStr string) bool {
	_, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return false
	}
	return true
}
