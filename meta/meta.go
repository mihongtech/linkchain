package meta

type IAmount interface{
	GetInt() int
	GetFloat() float32
	GetString() string

	IsLessThan(otherAmount IAmount) bool

	Subtraction(otherAmount IAmount)
	Addition(otherAmount IAmount)
}
