package main

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"hash/crc32"
	"reflect"
	"strings"
)

const (
	QUERY_STR_HDR   = "{\"selector\": {\"ObjTyp\":" //RichSql 拼接字符串头部
	QUERY_STR_TAL   = "},\"fields\":[\"key\"]}"     //RichSql 拼接字符串尾部
	REGEX_CONDITION = "$regex"
	EQ_CONDITION    = "$eq"
)

//非空参数检查
func checkNotEmpty(args []string, notEmptyIndexs []int) ([]string, error) {
	argstmp := args
	for i := 0; i < len(args); i++ {
		argstmp[i] = strings.TrimSpace(args[i])
		for j := 0; j < len(notEmptyIndexs); j++ {
			if i == notEmptyIndexs[j] && "" == argstmp[i] {
				return nil, fmt.Errorf("%s%d%s", "PN001:第 ", i+1, "个 参数不能为空")
			}
		}
	}
	return argstmp, nil
}

func formatStrLen(bseq string, strLen int) string {
	bseqLen := len(bseq)
	var buf bytes.Buffer
	if bseqLen < strLen {
		aLen := strLen - bseqLen
		for i := 0; i < aLen; i++ {
			buf.WriteString("0")
		}
	}
	buf.WriteString(bseq)
	return buf.String()
}

func hashCode(str string) uint32 {
	return crc32.ChecksumIEEE([]byte(str))
}

func checkJsonObjectSignature(obj interface{}, sgnAttrName, encX509Cert string) error {
	refV := reflect.ValueOf(obj).Elem()
	fieldV := refV.FieldByName(sgnAttrName)
	if fieldV.IsValid() {
		bakSgn := fieldV.String()
		fieldV.SetString("")
		data, err := json.Marshal(obj)
		if err != nil {
			return fmt.Errorf("MS100:检查签名时序列化对象出错[%s]", err.Error())
		}
		fieldV.SetString(bakSgn)
		return checkSignature(data, bakSgn, encX509Cert)
	} else {
		return fmt.Errorf("VS001:对象签名属性名%s无效", sgnAttrName)
	}
}

func checkSignature(data []byte, encSgn, encX509Cert string) error {
	x509Cert, err := base64.StdEncoding.DecodeString(encX509Cert)
	if err != nil {
		return fmt.Errorf("PS001:base64解码证书出错 [%s]", err.Error())
	}
	sign, err := base64.StdEncoding.DecodeString(encSgn)
	if err != nil {
		return fmt.Errorf("PS002:base64解码签名出错 [%s]", err.Error())
	}
	cert, err := x509.ParseCertificate(x509Cert)
	if err != nil {
		return fmt.Errorf("PS003:解析x509证书出错[%s]", err.Error())
	}

	pub, ok := cert.PublicKey.(*ecdsa.PublicKey)
	if !ok {
		return fmt.Errorf("PS004:证书公钥类型不是ecdsa椭圆曲线")
	}
	h := sha256.New()
	h.Write(data)
	digest := h.Sum(nil)
	valid, err := verifyECDSA(pub, sign, digest)
	if err != nil {
		return fmt.Errorf("PS005:验证签名出错[%s]", err.Error())
	}
	if !valid {
		return fmt.Errorf("PS006:签名验证不通过")
	}
	return nil
}

/*
	生成查询条件	objectType:表类型 qryStr:查询条件
	"{\"selector\": {\"ObjTyp表类型\":\"ObjTypxxx\",qryStr:查询条件},\"fields\":[\"key\"]}""
*/
func generateQryStr(objTyp string, qryStrs  ...string) (string, error) {
	// 添加数据类型
	if strings.Trim(objTyp, " ") == "" {
		return "", fmt.Errorf("GQ001:非法查询条件objTyp不能为空格")
	}
	var queryStr bytes.Buffer
	queryStr.WriteString(QUERY_STR_HDR)
	queryStr.WriteString("\"")
	queryStr.WriteString(objTyp)
	queryStr.WriteString("\"")

	for _, qryStr := range qryStrs {
		//添加查询条件（查询条件在外部已组装好，此处不进行验证）
		if strings.Trim(qryStr, " ") != "" {
			queryStr.WriteString(",")
			queryStr.WriteString(qryStr)
		}
	}
	queryStr.WriteString(QUERY_STR_TAL)
	return queryStr.String(), nil
}

func addRightMatch(key, value string) string {
	return addOpCondition(REGEX_CONDITION, key, ".*"+value)
}
func addLeftMatch(key, value string) string {
	return addOpCondition(REGEX_CONDITION, key, value+".*")
}
func addContainMatch(key, value string) string {
	return addOpCondition(REGEX_CONDITION, key, ".*"+value+".*")
}

func addRegexCondition(key, value string) string {
	return addOpCondition(REGEX_CONDITION, key, value)
}

func addEqCondition(key, value string) string {
	return addOpCondition(EQ_CONDITION, key, value)
}

func addOpCondition(op, key, value string) string {
	var keyBuffer bytes.Buffer
	keyBuffer.WriteString("\"")
	keyBuffer.WriteString(key)
	keyBuffer.WriteString("\":{\"")
	keyBuffer.WriteString(op)
	keyBuffer.WriteString("\":\"")
	keyBuffer.WriteString(value)
	keyBuffer.WriteString("\"}")
	return keyBuffer.String()
}
