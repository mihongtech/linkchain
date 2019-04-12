package normal

import "github.com/mihongtech/linkchain/core/meta"

type OffChainState struct {
}

func (o *OffChainState) Setup(i interface{}) bool {

	return true
}

func (o *OffChainState) Start() bool {

	return true
}

func (o *OffChainState) Stop() {
}

func (o *OffChainState) UpdateMainChain(ev meta.ChainEvent) {
}

func (o *OffChainState) UpdateSideChain(ev meta.ChainSideEvent) {
}
