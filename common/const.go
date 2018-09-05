package common

const (
	ASSET_HAS_SPENT     = "Y"
	ASSET_HAS_NOT_SPENT = "N"
)

const (
	ACCOUNT_LOG_PREFIX = "Log_Account_"
	CHAIN_LOG_PREFIX   = "Log_Chain_"
)

const (
	OBJECT_TYPE_ASSET      = "asset"
	OBJECT_TYPE_ASSET_ADDR = "assetAddr"
	OBJECT_TYPE_ASEETPOOL  = "assetPool"
	// OBJECT_TYPE_ORG       = "org"
	OBJECT_TYPE_LOG_ADDR   = "logAddr"
	OBJECT_TYPE_CHAIN_LOG  = "chainLog"
	OBJECT_TYPE_ORG        = "organization"
	OBJECT_TYPE_ASSET_INFO = "assetInfo"
)

const (
	TX_TYPE_TRANSFER_OUT = "OUTCOME"
	TX_TYPE_TRANSFER_IN  = "INCOME"
	TX_TYPE_ISSUE        = "ISSUE"
	TX_TYPE_TRANSFER     = "TRANSFER"
)
