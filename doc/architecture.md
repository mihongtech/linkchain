# LinkChain

## 1. Linkchain架构
- RPC层
		
		提供主链的功能性调用
		
- 功能层
		
		提供钱包管理，挖矿，节点管理
		
- API层 
		
		总领node，consensus，p2p，storage之间的调用
		
- 核心层

		node：程序主程序，各模块，各层之前调度
		consensus：区块链共识模块
		p2p：区块链网络模块
		storage：区块，交易结构，数据存储
		
- 工具层

		为各层提供基础工具库，如数字安全，编码，序列化/反序列化等。
		
![](https://github.com/xixisese/linkchain/blob/master/doc/source/architecture1.png?raw=true)

## 2. 分工
- 李晨
	+ node
	+ storage
	+ API
	+ lib/tool
- 李飞
	+ consensus
	+ wallet
	+ miner
	+ lib/tool
- 王志刚
	+ p2p
	+ manager
	+ lib/tool
- 段方平
	+ rpc

## 3. 开发计划
- 8.20~8.26
		
		框架设计，接口设计

- 8.27~9.2
		
		node主流程实现
		consensus部分实现
		p2p和lib实现

- 9.3~9.9
		
		strorage主流程实现，api注册调用机制实现
		consensus实现，miner实现
		manager实现

- 9.10~9.16

		node storage和api完善
		wallet实现
		rpc实现

- 9.17~9.23

		模块整合

- 9.24~9.30

		测试

## 4. 结构
* 区块头
![](https://github.com/xixisese/linkchain/blob/master/doc/source/architecture6.png?raw=true)

* 交易
![](https://github.com/xixisese/linkchain/blob/master/doc/source/architecture5.png?raw=true)

## 5. 接口
核心层所有模块都必须提供init start stop接口。

- consensus层接口
	+ ProcessBlock
	+ ProcessTX
	+ CheckBlock
	+ CheckTx
	+ GetMainChain
	+ GetBlock
	+ GetBestBlock
	+ GetTx
	+ GetCoin
	+ GetConsensusParams

- consensus层架构
![](https://github.com/xixisese/linkchain/blob/master/doc/source/architecture2.png?raw=true)

- consensus层核心功能
![](https://github.com/xixisese/linkchain/blob/master/doc/source/architecture3.png?raw=true)

![](https://github.com/xixisese/linkchain/blob/master/doc/source/architecture4.png?raw=true)