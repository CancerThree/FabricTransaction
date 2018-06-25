package main

import (
	"encoding/json"
	"testing"
)

const (
	requestString = `{"orgId":"zsb-1","publicKey":"MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQCMlvWDfHJ3HXreJ0TsFmCopd6f/cpmpC0iYBLiz4bzgI7mau1i4urhGe6E54U/KjxEcAe0xx47lnIRRkULIP4las6W5Jq2I7H3abz370XHHr1lli1Q+ZHE/aZVHR3QqMwW7rhlNKoP3B156w5k6DOeU/5Zr0cbkKKlcYmq5fldmQIDAQAB","signPublicKey":"MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQC2qv94gINYEtRfSRb7j2ZlLNsR6fwW5JebsBZZspxoaNXQyF1t8H3AAFAJD9F1DH4NXq6nZ+bnpMlQxKffLOgLkSJ+LKzbVEAhaZlbgJhOVx7fBPBBj+53zi6A3ZMVlouPwbwwXdAJFhPj3zmW2lKj6KCXmR7texuITb6dx8z48wIDAQAB","reqOrgId":"zsb-1","reqSign":"Lw8tEf+qJw01PBgyZq7llbEyZlzlNZ8wdL/hnuOSKrekbcUs9dPcwyT9DAQW8YISfGRTcZbpwR04hSgc4T4apgIN96HSq97p/4up3T7S3Zmsi+jPdwKMGjR+aQ71rbRb1NGaMR5QknplBLrA8pJ49lW87SA/WuD3bVmkr/xRQZU="}`
)

func TestSign(t *testing.T) {
	var request TXOrganizationRequest

	err := json.Unmarshal([]byte(requestString), &request)
	if err != nil {
		t.Error(err)
	}
	// 验证字段
	err = request.verifyField()
	if err != nil {
		t.Error(err)
	}

	// 验证签名
	publicKey := request.SignPublicKey
	err = CheckJSONObjectSignature(&request, publicKey)
	if err != nil {
		t.Error(err)
	}
}
