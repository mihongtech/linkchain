# linkchain

LinkChain is an open source blockchain product launched by Mihong technology team based on its  years of industry practice. Its goal is to provide customers with an intelligent basic blockchain capable of creating business value.


## Contents

1. [Getting start](##Getting-start)
2. [Command line](##Command-Line)
3. [Mainnet](##Mainnet)
4. [Testnet](##Testnet)
5. [Testnet deployment](##Testnet-deployment)

## Getting start
For prerequisites and detailed build instructions please read the
[Installation Docs](https://github.com/mihongtech/linkchain/blob/master/doc/install.md).

Building LinkChain requires a `Go` (version 1.10.1 or later) compiler,`lcd`and `lccli` enrtypoints.

###Install Go
You can install `go` using your favourite package manager or following the `go` [official docs](https://golang.org/doc/install). Then do remember setting your `$GOPATH`, `$GOBIN`, and `$PATH` enviroments.

	mkdir -p $HOME/go/bin
	echo "export GOPATH=$HOME/go" >> ~/.bash_profile
	echo "export GOBIN=\$GOPATH/bin" >> ~/.bash_profile
	echo "export PATH=\$PATH:\$GOBIN" >> ~/.bash_profile
	echo "export GO111MODULE=on" >> ~/.bash_profile
	source ~/.bash_profile

### Install golang-dep
An additional tool [golang-dep](https://github.com/golang/dep) for `go` needs to be installed by using the following command.

	go get -u github.com/golang/dep/cmd/dep


### Install the mainnet

The latest version of linkchain can be found from [github](https://github.com/mihongtech/linkchain). The `master` branch, which contains the latest stable release is recommended to choose. 

    mkdir -p $GOPATH/src/github.com/
    cd $GOPATH/src/github.com/
    git clone https://github.com/mihongtech/linkchain.git
    cd linkchain && git checkout master
    dep ensure -v

### Install the binaries

Then,install `lcd` and `lccli`.
 
- Unix
 
```bash

    cd build/unix
    ./build.sh

````

- Windows

```bash    

    cd build/win
    build.bat

````

### Verify install 
> *NOTE*: If you have issues at this step, please check that you have the latest stable version of GO installed.

That will install the `lcd` and `gaiacli` binaries. Verify that everything is OK:

```bash

	$ which lcd
	$ which lccli

```

### Next

Now you can [join the mainnet](./join-mainnet.md), [the public testnet](./join-testnet.md) or [create your own  testnet](./deploy-testnet.md)


## Command Line
Here is the list of command which will be used.

###block

This is all block command for handling block.

- **hash**: get block body command by blockhash

	*Example* :

	```
	block hash 98acd27a58c79eaab05ea4abd0daa8e63021df3bf2e65fcb38e2474fb706c3fe	
	```

- **height**: get block body command by height

	*Example*:

	````
	block height 0 
	````

###chain

This is all chain command for handling chain.

- **info**: get blockChain info

	*Example*:

	````
	chain info
	````

###contract


This is all contract command for handling contract

- **call**: call contract command which runs in local vm

	`call <from_address> <contract_address> <call_method>` 

	*Example*:

		contract call 8dafd997b6e65e680768076d92821716fd7950ee 91386e326c72b5d7f92431689d3ca921e13de07270a0823100000
		00000000000000000000a35c1bd74497c851265774e7e98027b46c27c41

- **get**: get contract receipt command

	*Example*:
	
		contract get d27a58c79eaab05ea4abd0daa8e63021df3bf2e65fcb38e2474fb706c3fe

- **publish**: create contract command

	`publish <from_address> <amount> <code>`
	
	*Example*:
	
		contract publish 8dafd997b6e65e680768076d92821716fd7950ee 3 6060604052600a8060106000396000f360606040526008565b00

- **transact**: call contract command which only runs on-chain vm

	`transact <from_address> <contract_address> <call_method> <amount>`

	*Example*:
		
		contract transact 8dafd997b6e65e680768076d92821716fd7950ee 91386e326c72b5d7f92431689d3ca921e13de072
		70a082310000000000000000000000000a35c1bd74497c851265774e7e98027b46c27c41 3




###miner


This is all miner command for handling miner

- **info**:get miner info

	*Example*:

		miner info

- **single**:mine single block

	*Example*:

		miner single

- **start**: start auto-miner

	*Example*:

		miner start

- **stop**: stop auto-miner

	*Example*:

		miner stop

###peer

This is all network command for handling network

- **add**:add peer node

	`add <node>`

	*Example*:

		peer add enode://0b049941925387066ddf8c719a1b9126e5c4a6147112cae163ee2ca24d231f33141e74
		93b4bf63f3eebbec19f4bcc7d49f4ab94b9721641afb95047f045b902c@[::]:40000

- **list**: get connected nodes

	*Example*:

		peer list

- **remove**:remove peer node

	`remove <node>`

	*Example*:
		
		peer remove enode://0b049941925387066ddf8c719a1b9126e5c4a6147112cae163ee2ca24d231f33141e7493b4bf63f3ee
		bbec19f4bcc7d49f4ab94b9721641afb95047f045b902c@[::]:40000


- **self**: get self node 

	*Example*:

		peer self

###tx
This is all tx command for handling tx

- **get**：get transaction body

	`get <hash>`

	*Example*:

		tx get 98acd27a58c79eaab05ea4abd0daa8e63021df3bf2e65fcb38e2474fb706c3fe

###quit

command for stopping linkchain client

	quit [flags]

###stop

command for stopping linkchain server

	stop [flags]


###version

get the version information of linkchain 

	version [flags]

###wallet

This is all wallet command for handling wallet

- **account**: get account info

	`account <address>`

	*Example*:

		wallet account 55b55e136cc6671014029dcbefc42a7db8ad9b9d11f62677a47fd2ed77eeef7b

- **export**: export privkey from wallet 

	`export <address>`

	*Example*:

		wallet export 025aa040dddd8f873ac5d02dfd249adc4d2c9d6def472a4405252fa6f6650ee1f0

- **import**: import privkey into wallet

	`import <privkey>`

	*Example*:

		wallet import 55b55e136cc6671014029dcbefc42a7db8ad9b9d11f62677a47fd2ed77eeef7b

- **info**: get wallet info

	*Example*:

		wallet info

- **newaccount**: generate new account

	*Example*:

		wallet newaccount

- **send**: send money to account command(normal tx)
	
	`send <from_address> <target_address> <amount>`

	*Example*:

		wallet send 02ed6749d314c2e725f1d23d250b4a041ea9c6369594b4f55500d7db41746cdf50 55b55e136cc6671014029dcbefc42a7db8ad9b9d11f62677a47fd2ed77eeef7b 10

## Mainnet

## Testnet
There are two ways to join linkchain testnet.

### Join in testnet at start app 

You can join testnet in start `lcd`

```bash
lcd --bootnodes enode://f6c2aa2d2760a3165578aef39e23ca876d16b8012def36744ff06f9c9f09ad3adbb495bed7ed6ccbad0c6134d26122e1569dc4872e003ab375cef31f7f60f0e9@[::]:40000

```

### Join in testnet during app running

You can join testnet during `lcd` running.

```bash
lccli

>peer add enode://f6c2aa2d2760a3165578aef39e23ca876d16b8012def36744ff06f9c9f09ad3adbb495bed7ed6ccbad0c6134d26122e1569dc4872e003ab375cef31f7f60f0e9@[::]:40000
```

### Verify the operation
```bash
lccli

>peer list
```

## Testnet deployment
### Start three node

Node1

```bash
nohup lcd --datadir /data/linkchain >/data/linkchain/debug.log 2>&1 &

```

Node2

```bash
nohup lcd --datadir /data/linkchain >/data/linkchain/debug.log 2>&1 &

```

Node3

```bash
nohup lcd --datadir /data/linkchain >/data/linkchain/debug.log 2>&1 &

```

### Connect Node

Node1

```bash

	lccli
	>peer self
	return:"enode://f6c2aa2d2760a3165578aef39e23ca876d16b8012def36744ff06f9c9f09
	ad3adbb495bed7ed6ccbad0c6134d26122e1569dc4872e003ab375cef31f7f60f0e9@[::]:40000"

```

Node2

```bash

	lccli
	
	>peer add enode://f6c2aa2d2760a3165578aef39e23ca876d16b8012def36744ff06f9c9f09
	ad3adbb495bed7ed6ccbad0c6134d26122e1569dc4872e003ab375cef31f7f60f0e9@[::]:40000
```

Node3

```bash
	
	lccli
	
	>peer add enode://f6c2aa2d2760a3165578aef39e23ca876d16b8012def36744ff06f9c9f09
	ad3adbb495bed7ed6ccbad0c6134d26122e1569dc4872e003ab375cef31f7f60f0e9@[::]:40000

```


### Mine

If you want to mine block, you should have a miner privateKey.

#### Have PrivateKey
If you have a miner privateKey, you can import the key into wallet via waller command in  [command line](##command-line).Next you can mine block.

```bash
lccli

>wallet import <privkey>
```

#### Have not PrivateKey
If you do not have a miner privateKey, you should change POA signer.

- First generate three accounts, 

```bash
	
	lccli
	
	>wallet newaccount
	>wallet newaccount
	>wallet newaccount
```

 The POA signer is in the path `/config/param.go`.

- Then modify `FirstPubMiner/SecondPubMinerThirdPubMiner` to you publicKey (address/accountID).

- Final, the LinkChain project should be rebuilded and restarted.


#### Mine block

Start auto mine

```bash
lccli

>miner start

```

Stop auto mine

```bash
lccli

>miner stop

```

Get auto mine info

```bash
lccli

>miner info

```

Miner single block

```bash
lccli

>miner mine

```




