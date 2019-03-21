package interpreter

import "github.com/linkchain/common/lcdb"

type Interpreter interface {
	Executor
	Validator
	Processor
	CreateOffChain(db lcdb.Database) OffChain
}
