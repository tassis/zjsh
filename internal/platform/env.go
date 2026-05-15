package platform

import "os"

type Env interface {
	LookupEnv(key string) (string, bool)
	InZellij() bool
}

type OSEnv struct{}

func (OSEnv) LookupEnv(key string) (string, bool) {
	return os.LookupEnv(key)
}

func (e OSEnv) InZellij() bool {
	_, ok := e.LookupEnv("ZELLIJ")
	return ok
}
