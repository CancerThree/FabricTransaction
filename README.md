# FabricTransaction
Fabric下账户之间的交易模型实现.

## 1. 模块介绍
* __OrgManage组织管理__</br>
    每个Fabric节点都可以对机构信息进行新增（此处机构其实可以对应着现实中每一个使用此系统的用户）。
    每个机构需要将自己的验签密钥传至链上，以便在机构发起交易的时候对所发交易信息进行签名验证；
* __assetPool资产池管理__</br>
    每个机构下可以管理多个资产池，但是为了保证信息的私密性，资产池的地址由各个机构自己保存、管理。在进行交易时，机构选择使用哪个资产池进行交易。即资产池模块对应着其他代币系统的钱包结构。
* __asset资产管理__</br>
    资产对应着代币的结构。为了实现隐藏资产池资产与资产池之间的对应关系,assetPool下所存储的是资产池私钥加密后的资产地址。
## 2. 交易流程

主动转账：
1. 转出方资产池直接调用`transfer(addr, value)`将`value`量的资产转入转入方的资产池。

报价交易：
1. 转出方将待转出资产授权给合约资产池：</br>
   转出方调用`approve(cpontractAssetPool, value)`将`value`值的资产授权给合约资产池使用；
2. 系统展示在本系统中各个挂出出售的资产；
3. 转入方向系统发送想要购买的资产：
    1. 资产转移：合约资产池调用`transferFrom(outPool, inPool, value)`将转入方所要购买的资产转入转入方的资产池；
    2. 代币转移：由转入方主动调用`transfer(out, value)`将对应`value`量的代币资产付给转出方资产库。
   
合约调用费用：
1. 主动转账：主动转账时，调用费用由转出方提供；
2. 报价交易：可以由购买方付出调用费用；
两种调用都可以直接调用`transfer(contractWallet, value)`进行合约调用的付费。

机构支持新增，每次交易都需要对交易对机构签名进行验证，每个Fabric节点上都可以进行机构对
