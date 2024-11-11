package config

import (
	"fmt"
	"testing"
)

type Config struct {
	Secret string `yaml:"secret" json:"secret"`
}

var conf = Local[Config]().Watch().InitData(Config{
	Secret: "hello world",
})

func _assert(condition bool, msg string, v ...interface{}) {
	if !condition {
		panic(fmt.Sprintf("assertion failed: "+msg, v...))
	}
}

func TestConfig(t *testing.T) {
	s := conf.Get().Secret
	_assert(s != "", "get config failded")
}
