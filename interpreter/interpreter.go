package interpreter

import "github.com/mihongtech/linkchain/common/lcdb"

type Interpreter interface {
	Executor
	Validator
	Processor
	CreateOffChain(db lcdb.Database) OffChain
}
