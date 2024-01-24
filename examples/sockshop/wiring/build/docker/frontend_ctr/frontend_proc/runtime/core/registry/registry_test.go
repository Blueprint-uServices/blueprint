package registry

import (
	"testing"
)

type mystruct struct{}

var reg = NewServiceRegistry[mystruct]("test")

func TestNothing(t *testing.T) {
	reg.SetDefault("hi")
}
