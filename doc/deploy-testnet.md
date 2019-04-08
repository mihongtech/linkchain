## Deploy TestNet

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
return:"enode://f6c2aa2d2760a3165578aef39e23ca876d16b8012def36744ff06f9c9f09ad3adbb495bed7ed6ccbad0c6134d26122e1569dc4872e003ab375cef31f7f60f0e9@[::]:40000"
```

Node2

```bash
lccli

>peer add enode://f6c2aa2d2760a3165578aef39e23ca876d16b8012def36744ff06f9c9f09ad3adbb495bed7ed6ccbad0c6134d26122e1569dc4872e003ab375cef31f7f60f0e9@[::]:40000
```

Node3

```bash
lccli

>peer add enode://f6c2aa2d2760a3165578aef39e23ca876d16b8012def36744ff06f9c9f09ad3adbb495bed7ed6ccbad0c6134d26122e1569dc4872e003ab375cef31f7f60f0e9@[::]:40000
```


### Mine

If you want to mine block, you must be miner privateKey.

#### Have PrivateKey
If you have miner privateKey, you can import privateKey into wallet.Next you can mine block.

```bash
lccli

>wallet import <privkey>
```

#### Have not PrivateKey
If you have not miner privateKey, you must be change poa signer.

First generate three account

```bash
lccli

>wallet newaccount
>wallet newaccount
>wallet newaccount
```

The poa signer is in config/param.go .Then modify `FirstPubMiner/SecondPubMinerThirdPubMiner` to you publicKey (address/accountID).

After change poa signer,you must rebuild and restart application


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
