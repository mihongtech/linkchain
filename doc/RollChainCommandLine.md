# RollChain Command Line

## Wallet
1. get wallet info
	
		wallet info
		
2. get wallet account info
		
		wallet getAccount <accountId>
		
3. get new account

		wallet newAccount
		
4. send money to account

		wallet send <to accountId> <money(int)>
		
5. insurance account create 

		wallet createInsurance <insurance accountId> <money(int)> <security accountId> <clearTime(int)>
		
		eg.
		wallet createInsurance 03e1ea8de600f857dab5a3c3261f978d60e4280e0c012e113bd6d8681db485852c 1 03df398d9da7d3b4d8e7e11a4d8114685465a68ff059a2e3e762532172b3790711 10
		
6. insurance account send money to other account

		wallet delayInsurance <insurance accountId(choose account)> <to accountId> <money>
		
		eg.
		wallet delayInsurance 03e1ea8de600f857dab5a3c3261f978d60e4280e0c012e113bd6d8681db485852c 033b2b402318ce82946e2356220c7a875691bf3c44fb4843004a2210bc24178c4a 1
		
		return.
		txid=2ab6eb0277f3588ec1177dea277e5f841301703ca8887276051ce2af06ced716
		
7. insurance withdraw tx

		wallet withdrawInsurance <insurance accountId(choose withdraw account)> <txid((choose withdraw tx))>
		
		eg.
		wallet withdrawInsurance 03e1ea8de600f857dab5a3c3261f978d60e4280e0c012e113bd6d8681db485852c 2ab6eb0277f3588ec1177dea277e5f841301703ca8887276051ce2af06ced716
		
8. insurance account clear
	
		ps:clear tx will be pushed by miner after reaching to cleartime
		
9. insurance account set clear time
	
		wallet setClearTime <insurance accountId(choose withdraw account)> <clearTime(int)>
		
		eg.
		wallet setClearTime 03e1ea8de600f857dab5a3c3261f978d60e4280e0c012e113bd6d8681db485852c 30
	
10. import account by privkey

		wallet import <privkey>
		
11. export account by pubkey

		wallet export <pubkey>
		
## Miner
1. get miner info

		miner info
		
2. start miner

		miner start
		
3. stop miner 

		miner stop
		
## Chain
1. get chain info
	
		chain info
		
## Block
1. get block by height

		block height <height>
		
2. get block by hash

		block hash <blockId>

## Tx
1. get tx by TxId
		
		tx get <txid>
		
## Net
1. get peer list

		peer list
	
2. add peer

		peer add <peer url>
		
3. remove 

		peer remove <peer url>