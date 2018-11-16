package core

type Service interface {
	Setup(i interface{}) bool
	Start() bool
	Stop()
}
