package core

type Service interface{
	Init(i interface{}) bool
	Start() bool
	Stop()
}