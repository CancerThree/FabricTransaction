package transaction

type address string

type ERC20TokenInterface interface {
	// get name of the token
	// 获取token的名称
	Name() string

	// get short name of the token
	// 获取token的简称
	Symbol() string

	// get the number of decimals the token uses
	// 获取token支持的小数点位数
	Decimals() uint8

	// get the total token supply
	// 获取token的总发行额
	TotalSupply() uint64

	// get the balance within the address
	// 获取某地址下的token余额
	BalanceOf(_owner address) uint64

	// transfer _value amnount of token to adrress _to
	// 将自己的token转账至_to地址
	Transfer(_to address, _value uint64) (bool, error)

	// transfer _value amount token from address _from to address _to
	// 从地址 _from发送数量为 _value的token到地址 _to,必须触发Transfer事件。
	// transferFrom方法用于允许合同代理某人转移token。条件是from账户必须经过了approve。
	TransferFrom(_from address, _to address, _value uint64) (bool, error)

	// Allows _spender to withdraw from your account multiple times,
	// up to the _value amount. If this function is called again it overwrites the current allowance with _value.
	Approve(_spender address, _value uint64) (bool error)

	// Returns the amount which _spender is still allowed to withdraw from _owner.
	Allowance(_owner address, _spender address) uint64

	// MUST trigger when tokens are transferred, including zero value transfers.
	SetTransferEvent(_from address, _to address, _value uint64) error

	// MUST trigger on any successful call to approve(address _spender, uint256 _value).
	SetApprovalEvent(_from address, _to address, _value uint64) error
}
