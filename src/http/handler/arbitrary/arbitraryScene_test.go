package arbitrary_test

import "github.com/lillilli/dr_web_exercise/src/http/handler/test"

type ArbitraryHandlerScene struct {
	test.HandlerScene
}

func NewArbitraryHandlerScene() *ArbitraryHandlerScene {
	s := &ArbitraryHandlerScene{}
	s.Init()
	return s
}

func (s *ArbitraryHandlerScene) Init() {
	s.HandlerScene.Init()
}
