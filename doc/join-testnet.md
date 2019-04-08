## Join TestNet

It has two ways to join linkchain testnet.

### Join in testnet at start app 

You can join testnet in start `lcd`

```bash
lcd --bootnodes enode://f6c2aa2d2760a3165578aef39e23ca876d16b8012def36744ff06f9c9f09ad3adbb495bed7ed6ccbad0c6134d26122e1569dc4872e003ab375cef31f7f60f0e9@[::]:40000

```

### Join in testnet during app running

You can join testnet during `lcd` running

```bash
lccli

>peer add enode://f6c2aa2d2760a3165578aef39e23ca876d16b8012def36744ff06f9c9f09ad3adbb495bed7ed6ccbad0c6134d26122e1569dc4872e003ab375cef31f7f60f0e9@[::]:40000
```

### Verify join in success

```bash
lccli

>peer list
```