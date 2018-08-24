package versioned_test

import (
	test "github.com/lillilli/dr_web_exercise/src/http/handler/test"
)

type VersionedHandlerScene struct {
	test.HandlerScene
}

func NewVersionedHandlerScene() *VersionedHandlerScene {
	s := &VersionedHandlerScene{}
	return s.Init()
}

func (s *VersionedHandlerScene) Init() *VersionedHandlerScene {
	s.HandlerScene.Init()
	return s
}
